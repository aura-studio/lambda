package sqs

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/aura-studio/lambda/dynamic"
	events "github.com/aws/aws-lambda-go/events"
)

type Engine struct {
	*Options
	*dynamic.Dynamic
	r       *router
	running atomic.Int32
}

func NewEngine(opts ...ServeOption) *Engine {
	bag := &serveOptionBag{}
	for _, opt := range opts {
		if opt != nil {
			opt.apply(bag)
		}
	}

	e := &Engine{
		Options: NewOptions(bag.sqs...),
		Dynamic: dynamic.NewDynamic(bag.dynamic...),
	}
	e.InstallHandlers()
	return e
}

type InvokeFunc func(path string, req string) (string, error)

func (e *Engine) Start() {
}

func (e *Engine) Stop() {
}

func (e *Engine) invokeDefault(path string, req string) (string, error) {
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
	_, err := e.handleSQSMessages(ctx, ev)
	return err
}

func (e *Engine) HandleSQSMessagesWithResponse(ctx context.Context, ev events.SQSEvent) (events.SQSEventResponse, error) {
	e.handleSQSMessages()
}

func (e *Engine) handleSQSMessages(ctx context.Context, ev events.SQSEvent, partial bool) (resp events.SQSEventResponse, err error) {
	_ = ctx
	for _, msg := range ev.Records {
		if e.running.Load() == 0 {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		request, err := DecodeRequestBody(msg.Body)
		if err != nil {
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
		b, err := MarshalResponse(rsp)
		if err != nil {
			resp.BatchItemFailures = append(resp.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
			continue
		}

		if rsp.ClientSqsId != "" && rsp.CorrelationId != "" {
			
		}
	}

	return resp, nil
}
