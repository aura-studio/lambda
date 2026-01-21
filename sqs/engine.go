package sqs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/aura-studio/lambda/dynamic"
	events "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"google.golang.org/protobuf/proto"
)

type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type Engine struct {
	*Options
	*dynamic.Dynamic
	r         *router
	running   atomic.Int32
	sqsClient SQSClient
}

func NewEngine(opts ...ServeOption) *Engine {
	bag := &serveOptionBag{}
	bag.apply(opts...)

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	e := &Engine{
		Options: NewOptions(bag.sqs...),
		Dynamic: dynamic.NewDynamic(bag.dynamic...),
	}
	if e.Options.SQSClient != nil {
		e.sqsClient = e.Options.SQSClient
	} else {
		e.sqsClient = sqs.NewFromConfig(cfg)
	}
	e.running.Store(1)
	e.InstallHandlers()
	return e
}

func (e *Engine) Start() {
	e.running.Store(1)
}

func (e *Engine) Stop() {
	e.running.Store(0)
}

func (e *Engine) handle(path string, req string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid path: %q", path)
	}
	pkg := parts[0]
	version := parts[1]

	tunnel, err := e.GetPackage(pkg, version)
	if err != nil {
		return "", err
	}

	route := fmt.Sprintf("/%s", strings.Join(parts[2:], "/"))
	return tunnel.Invoke(route, req), nil
}

// HandleSQSMessagesWithoutResponse 重试全部数据
func (e *Engine) HandleSQSMessagesWithoutResponse(ctx context.Context, ev events.SQSEvent) error {
	resp, err := e.handleSQSMessages(ctx, ev)
	if err != nil {
		return err
	}
	if len(resp.BatchItemFailures) > 0 {
		return fmt.Errorf("batch item failures: %d", len(resp.BatchItemFailures))
	}
	return nil
}

// HandleSQSMessagesWithResponse 部分重试
func (e *Engine) HandleSQSMessagesWithResponse(ctx context.Context, ev events.SQSEvent) (events.SQSEventResponse, error) {
	return e.handleSQSMessages(ctx, ev)
}

func (e *Engine) Invoke(ctx context.Context, ev events.SQSEvent) (events.SQSEventResponse, error) {
	if e.PartialRetry {
		return e.HandleSQSMessagesWithResponse(ctx, ev)
	}
	return events.SQSEventResponse{}, e.HandleSQSMessagesWithoutResponse(ctx, ev)
}

func (e *Engine) handleSQSMessages(ctx context.Context, ev events.SQSEvent) (resp events.SQSEventResponse, err error) {
	_ = ctx
	for _, msg := range ev.Records {
		if e.running.Load() == 0 {
			if e.DebugMode {
				log.Printf("[SQS] Engine stopped, message %s failed", msg.MessageId)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		b, err := base64.StdEncoding.DecodeString(msg.Body)
		if err != nil {
			if e.DebugMode {
				log.Printf("[SQS] Decode message %s body error: %v", msg.MessageId, err)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		var request Request
		if err := proto.Unmarshal(b, &request); err != nil {
			if e.DebugMode {
				log.Printf("[SQS] Unmarshal message %s body error: %v", msg.MessageId, err)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		c := &Context{
			Engine:  e,
			RawPath: request.Path,
			Path:    request.Path,
			Request: string(request.Payload),
		}

		if e.DebugMode {
			log.Printf("[SQS] Request: %s %s", c.Path, c.Request)
		}

		e.r.dispatch(c)

		if e.DebugMode {
			log.Printf("[SQS] Response: %s %s", c.Path, c.Response)
		}
		if c.Err != nil {
			if e.ErrorSuspend {
				return resp, c.Err
			}
			if e.DebugMode {
				log.Printf("[SQS] Dispatch message %s error: %v", msg.MessageId, c.Err)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		// Response is produced only when ResponseSqsId is provided and ResponseSwitch is on.
		if !e.ResponseSwitch || request.ResponseSqsId == "" {
			continue
		}
		// When a response is requested, RequestSqsId must be present.
		if request.RequestSqsId == "" {
			if e.DebugMode {
				log.Printf("[SQS] RequestSqsId is empty for message %s", msg.MessageId)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		rsp := &Response{
			RequestSqsId:  request.RequestSqsId,
			ResponseSqsId: request.ResponseSqsId,
			CorrelationId: request.CorrelationId,
			Payload:       []byte(c.Response),
		}
		b, err = proto.Marshal(rsp)
		if err != nil {
			if e.DebugMode {
				log.Printf("[SQS] Marshal response for message %s error: %v", msg.MessageId, err)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		if request.ResponseSqsId != "" {
			_, err := e.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: aws.String(base64.StdEncoding.EncodeToString(b)),
				QueueUrl:    &request.ResponseSqsId,
			})
			if err != nil {
				if e.DebugMode {
					log.Printf("[SQS] Send response for message %s error: %v", msg.MessageId, err)
				}
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			}
		}
	}

	return resp, nil
}
