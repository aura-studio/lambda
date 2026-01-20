package tests

import (
	"context"
	"strings"
	"testing"

	lambdasqs "github.com/aura-studio/lambda/sqs"
	events "github.com/aws/aws-lambda-go/events"
)

func mustPBRequest(t *testing.T, r *lambdasqs.Request) string {
	t.Helper()
	b, err := lambdasqs.MarshalRequest(r)
	if err != nil {
		t.Fatalf("MarshalRequest: %v", err)
	}
	return string(b)
}

func TestSQSHandler_PartialFailures(t *testing.T) {
	e := lambdasqs.NewEngine()

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: "not-proto"}, // invalid protobuf -> fail
		{MessageId: "2", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`), ClientSqsId: "c2"})}, // GetPackage likely fails -> fail
	}}

	resp, err := e.Handle(context.Background(), ev)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if len(resp.BatchItemFailures) != 2 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
}

func TestSQSHandler_ResponseRouting(t *testing.T) {
	e := lambdasqs.NewEngine()
	e.SetInvokeFunc(func(path string, req string) (string, error) {
		return "OK", nil
	})

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "ignored-1", Body: mustPBRequest(t, &lambdasqs.Request{ClientSqsId: "client-1", ServerSqsId: "server-9", Path: "/api/pkg/version/route", Payload: []byte(`{}`)})},
		{MessageId: "ignored-2", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`), ClientSqsId: "client-2"})},
	}}

	_, out, err := e.HandleWithResponses(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleWithResponses error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("out len = %d", len(out))
	}
	if out[0].QueueID != "server-9" {
		t.Fatalf("QueueID = %q", out[0].QueueID)
	}
	rsp, err := lambdasqs.UnmarshalResponse([]byte(out[0].Body))
	if err != nil {
		t.Fatalf("UnmarshalResponse: %v", err)
	}
	if rsp.ClientSqsId != "client-1" {
		t.Fatalf("ClientSqsId = %q", rsp.ClientSqsId)
	}
	if rsp.ServerSqsId != "server-9" {
		t.Fatalf("ServerSqsId = %q", rsp.ServerSqsId)
	}
	if strings.TrimSpace(string(rsp.Payload)) != "OK" {
		t.Fatalf("Payload = %q", string(rsp.Payload))
	}
}

func TestSQSHandler_NoResponse_AllowsEmptyClientSqsId(t *testing.T) {
	e := lambdasqs.NewEngine()
	e.SetInvokeFunc(func(path string, req string) (string, error) {
		return "OK", nil
	})

	// No ServerSqsId and no `?rsp=` in path => no response needed.
	// ClientSqsId is allowed to be empty in this case.
	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "m1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`)})},
	}}

	resp, out, err := e.HandleWithResponses(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleWithResponses error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
	if len(out) != 0 {
		t.Fatalf("out len = %d", len(out))
	}
}

func TestSQSHandler_HealthCheck_OK(t *testing.T) {
	e := lambdasqs.NewEngine()

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "h1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/health-check"})},
	}}

	resp, out, err := e.HandleWithResponses(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleWithResponses error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
	if len(out) != 0 {
		t.Fatalf("out len = %d", len(out))
	}
}

func TestSQSHandler_APIPrefix_StripsToWildcardPath(t *testing.T) {
	e := lambdasqs.NewEngine()
	var gotPath string
	e.SetInvokeFunc(func(path string, req string) (string, error) {
		gotPath = path
		return "OK", nil
	})

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "a1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/pkg/version/route", Payload: []byte(`{}`)})},
	}}

	resp, out, err := e.HandleWithResponses(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleWithResponses error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
	if len(out) != 0 {
		t.Fatalf("out len = %d", len(out))
	}
	if gotPath != "/pkg/version/route" {
		t.Fatalf("gotPath = %q", gotPath)
	}
}

func TestSQSHandler_APIPath_RequiresPrefix(t *testing.T) {
	e := lambdasqs.NewEngine()
	e.SetInvokeFunc(func(path string, req string) (string, error) {
		return "OK", nil
	})

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "p1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/pkg/version/route", Payload: []byte(`{}`)})},
	}}

	resp, out, err := e.HandleWithResponses(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleWithResponses error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("out len = %d", len(out))
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("BatchItemFailures len = %d", len(resp.BatchItemFailures))
	}
}
