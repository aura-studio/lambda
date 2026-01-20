package sqs

import (
	"context"
	"encoding/base64"
	"fmt"
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
	for _, opt := range opts {
		if opt != nil {
			opt.apply(bag)
		}
	}

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

func (e *Engine) HandleSQSMessagesWithoutResponse(ctx context.Context, ev events.SQSEvent) error {
	resp, err := e.handleSQSMessages(ctx, ev, false)
	if err != nil {
		return err
	}
	if len(resp.BatchItemFailures) > 0 {
		return fmt.Errorf("batch item failures: %d", len(resp.BatchItemFailures))
	}
	return nil
}

func (e *Engine) HandleSQSMessagesWithResponse(ctx context.Context, ev events.SQSEvent) (events.SQSEventResponse, error) {
	return e.handleSQSMessages(ctx, ev, true)
}

func (e *Engine) handleSQSMessages(ctx context.Context, ev events.SQSEvent, partial bool) (resp events.SQSEventResponse, err error) {
	_ = ctx
	for _, msg := range ev.Records {
		if e.running.Load() == 0 {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		var request Request
		if err := proto.Unmarshal([]byte(msg.Body), &request); err != nil {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		c := &Context{
			Engine:  e,
			RawPath: request.Path,
			Path:    request.Path,
			Request: string(request.Payload),
		}
		e.r.dispatch(c)
		if c.Err != nil {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		// Response is produced only when ServerSqsId is provided.
		if request.ServerSqsId == "" {
			continue
		}
		// When a response is requested, ClientSqsId must be present.
		if request.ClientSqsId == "" {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		rsp := &Response{
			ClientSqsId:   request.ClientSqsId,
			ServerSqsId:   request.ServerSqsId,
			CorrelationId: request.CorrelationId,
			Payload:       []byte(c.Response),
		}
		b, err := proto.Marshal(rsp)
		if err != nil {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		if request.ServerSqsId != "" {
			_, err := e.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: aws.String(base64.StdEncoding.EncodeToString(b)),
				QueueUrl:    &request.ServerSqsId,
			})
			if err != nil {
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			}
		}
	}

	return resp, nil
}
