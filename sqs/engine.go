package sqs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
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

func NewEngine(sqsOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	e := &Engine{
		Options: NewOptions(sqsOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
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
	switch e.RunMode {
	case RunModeStrict, RunModePartial:
		return e.HandleSQSMessagesWithResponse(ctx, ev)
	case RunModeBatch, RunModeReentrant:
		return events.SQSEventResponse{}, e.HandleSQSMessagesWithoutResponse(ctx, ev)
	default:
		return events.SQSEventResponse{}, e.HandleSQSMessagesWithoutResponse(ctx, ev)
	}
}

func (e *Engine) handleSQSMessages(ctx context.Context, ev events.SQSEvent) (resp events.SQSEventResponse, err error) {
	_ = ctx
	for i, msg := range ev.Records {
		if e.running.Load() == 0 {
			if e.DebugMode {
				log.Printf("[SQS] Engine stopped, message %s failed", msg.MessageId)
			}
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		b, decodeErr := base64.StdEncoding.DecodeString(msg.Body)
		if decodeErr != nil {
			if e.DebugMode {
				log.Printf("[SQS] Decode message %s body error: %v", msg.MessageId, decodeErr)
			}
			switch e.RunMode {
			case RunModeStrict:
				for j := i; j < len(ev.Records); j++ {
					resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: ev.Records[j].MessageId})
				}
				return resp, nil
			case RunModePartial:
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			case RunModeBatch:
				return resp, decodeErr
			case RunModeReentrant:
				err = decodeErr
				continue
			}
		}

		var request Request
		if unmarshalErr := proto.Unmarshal(b, &request); unmarshalErr != nil {
			if e.DebugMode {
				log.Printf("[SQS] Unmarshal message %s body error: %v", msg.MessageId, unmarshalErr)
			}
			switch e.RunMode {
			case RunModeStrict:
				for j := i; j < len(ev.Records); j++ {
					resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: ev.Records[j].MessageId})
				}
				return resp, nil
			case RunModePartial:
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			case RunModeBatch:
				return resp, unmarshalErr
			case RunModeReentrant:
				err = unmarshalErr
				continue
			}
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

		func() {
			defer func() {
				if r := recover(); r != nil {
					c.Err = fmt.Errorf("panic: %v", r)
				}
			}()
			e.r.dispatch(c)
		}()

		if e.DebugMode {
			log.Printf("[SQS] Response: %s %s", c.Path, c.Response)
		}
		if c.Err != nil {
			if e.DebugMode {
				log.Printf("[SQS] Dispatch message %s error: %v", msg.MessageId, c.Err)
			}

			switch e.RunMode {
			case RunModeStrict:
				for j := i; j < len(ev.Records); j++ {
					resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: ev.Records[j].MessageId})
				}
				return resp, nil
			case RunModePartial:
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			case RunModeBatch:
				return resp, c.Err
			case RunModeReentrant:
				err = c.Err
				continue
			}
		}

		// Response is produced only when ResponseSqsId is provided.
		if request.ResponseSqsId == "" {
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

		if e.ReplyMode && request.ResponseSqsId != "" {
			_, sendErr := e.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: aws.String(base64.StdEncoding.EncodeToString(b)),
				QueueUrl:    &request.ResponseSqsId,
			})
			if sendErr != nil {
				if e.DebugMode {
					log.Printf("[SQS] Send response for message %s error: %v", msg.MessageId, sendErr)
				}
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			}
		}
	}

	return resp, err
}
