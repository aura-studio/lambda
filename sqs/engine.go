package sqs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aura-studio/lambda/dynamic"
	events "github.com/aws/aws-lambda-go/events"
)

type Engine struct {
	*Options
	*dynamic.Dynamic
	r *router

	invokeFunc InvokeFunc
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

type OutgoingMessage struct {
	QueueID string
	Body    string
}

type InvokeFunc func(path string, req string) (string, error)

func (e *Engine) Start() {
	e.invokeFunc = fn
}

func (e *Engine) Stop() {
	e.invokeFunc = nil
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

// Handle implements an AWS Lambda SQS event handler.
// It returns partial batch failures via SQSEventResponse.
func (e *Engine) Handle(ctx context.Context, ev events.SQSEvent) (events.SQSEventResponse, error) {
	resp, _, err := e.HandleWithResponses(ctx, ev)
	return resp, err
}

// HandleWithResponses is like Handle, but additionally returns response messages
// when the request path includes a response SQS id via query param `rsp`.
//
//   - If `rsp` is absent: no response message is produced for that record.
//   - If `rsp` is present: response is serialized and returned as OutgoingMessage.
//     The body includes both client and server SQS ids.
func (e *Engine) HandleWithResponses(ctx context.Context, ev events.SQSEvent) (events.SQSEventResponse, []OutgoingMessage, error) {
	_ = ctx
	resp := events.SQSEventResponse{}
	var out []OutgoingMessage
	for _, msg := range ev.Records {
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
		out = append(out, OutgoingMessage{QueueID: request.ServerSqsId, Body: string(b)})
	}
	return resp, out, nil
}
