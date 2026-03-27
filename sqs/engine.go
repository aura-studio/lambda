package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type Engine struct {
	*Options
	*Router
	*dynamic.Dynamic
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
		Router:  NewRouter(),
	}
	if e.Options.SQSClient != nil {
		e.sqsClient = e.Options.SQSClient
	} else {
		e.sqsClient = sqs.NewFromConfig(cfg)
	}
	e.InstallHandlers()
	return e
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
		if e.DebugMode {
			log.Printf("[SQS] Message %s body: %s", msg.MessageId, msg.Body)
		}

		var request Request
		if unmarshalErr := json.Unmarshal([]byte(msg.Body), &request); unmarshalErr != nil {
			log.Printf("[SQS] Unmarshal message %s body error: %v", msg.MessageId, unmarshalErr)
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
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				err = unmarshalErr
				continue
			}
		}

		c := &Context{}
		c.Set(ContextPath, request.Path)
		c.Set(ContextRequest, string(request.Payload))

		if e.DebugMode {
			log.Printf("[SQS] Request: %s %s", c.GetString(ContextPath), c.GetString(ContextRequest))
		}

		e.Router.Dispatch(c)

		if e.DebugMode {
			log.Printf("[SQS] Response: %s %s", c.GetString(ContextPath), c.GetString(ContextResponse))
		}

		// check panic first, then error
		var cErr error
		if v, ok := c.Get(ContextPanic); ok && v != nil {
			cErr = v.(error)
		} else {
			cErr = c.GetError()
		}
		if cErr != nil {
			log.Printf("[SQS] Dispatch message %s error: %v", msg.MessageId, cErr)

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
				return resp, cErr
			case RunModeReentrant:
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				err = cErr
				continue
			}
		}

		// Response is produced only when ResponseSqsId is provided.
		if request.ResponseSqsId == "" {
			continue
		}
		if request.RequestSqsId == "" {
			log.Printf("[SQS] RequestSqsId is empty for message %s", msg.MessageId)
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		rsp := &Response{
			RequestSqsId:  request.RequestSqsId,
			ResponseSqsId: request.ResponseSqsId,
			CorrelationId: request.CorrelationId,
			Payload:       []byte(c.GetString(ContextResponse)),
		}
		b, marshalErr := json.Marshal(rsp)
		if marshalErr != nil {
			log.Printf("[SQS] Marshal response for message %s error: %v", msg.MessageId, marshalErr)
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		if e.ReplyMode && request.ResponseSqsId != "" {
			_, sendErr := e.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: aws.String(string(b)),
				QueueUrl:    &request.ResponseSqsId,
			})
			if sendErr != nil {
				log.Printf("[SQS] Send response for message %s error: %v", msg.MessageId, sendErr)
				resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
				continue
			}
		}
	}

	return resp, err
}
