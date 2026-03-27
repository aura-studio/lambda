package tests

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aura-studio/dynamic"
	lambdasqs "github.com/aura-studio/lambda/sqs"
	events "github.com/aws/aws-lambda-go/events"
)

// TestSQSEngineCreation tests that NewEngine creates an engine with correct options
func TestSQSEngineCreation(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithDebugMode(true),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
		lambdasqs.WithReplyMode(true),
	}, nil)

	if e == nil {
		t.Fatal("NewEngine returned nil")
	}

	if !e.DebugMode {
		t.Error("DebugMode should be true")
	}

	if e.RunMode != lambdasqs.RunModePartial {
		t.Errorf("RunMode = %q, want 'partial'", e.RunMode)
	}

	if !e.ReplyMode {
		t.Error("ReplyMode should be true")
	}
}

// TestSQSEngineInvokeHealthCheck tests the health check endpoint
func TestSQSEngineInvokeHealthCheck(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "h1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/health-check"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures, got %d", len(resp.BatchItemFailures))
	}
}

// TestSQSEngineInvokeRootPath tests the root path endpoint
func TestSQSEngineInvokeRootPath(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "r1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures, got %d", len(resp.BatchItemFailures))
	}
}

// TestSQSEngineInvokePageNotFound tests the 404 handler
func TestSQSEngineInvokePageNotFound(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "404", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/nonexistent/path"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("Expected 1 failure for 404, got %d", len(resp.BatchItemFailures))
	}
}

// TestSQSEngineInvokeInvalidBase64 tests handling of invalid base64 payload
func TestSQSEngineInvokeInvalidBase64(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "invalid", Body: "not-valid-base64!!!"},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("Expected 1 failure for invalid base64, got %d", len(resp.BatchItemFailures))
	}
}

// TestSQSEngineInvokeInvalidProtobuf tests handling of invalid protobuf payload
func TestSQSEngineInvokeInvalidProtobuf(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	// Valid base64 but invalid protobuf
	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "invalid-pb", Body: base64.StdEncoding.EncodeToString([]byte("not-protobuf"))},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("Expected 1 failure for invalid protobuf, got %d", len(resp.BatchItemFailures))
	}
}

// TestSQSEngineInvokeAPIWithDynamic tests API route calling Dynamic
func TestSQSEngineInvokeAPIWithDynamic(t *testing.T) {
	mock := &mockSQSClient{}

	var invokedRoute, invokedReq string
	dynamic.RegisterPackage("sqspkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			return "sqs-api-response"
		},
	})

	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "api", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/sqspkg/v1/myroute", Payload: []byte(`{"key":"value"}`)})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures, got %d", len(resp.BatchItemFailures))
	}

	if invokedRoute != "/myroute" {
		t.Errorf("invokedRoute = %q, want '/myroute'", invokedRoute)
	}

	if invokedReq != `{"key":"value"}` {
		t.Errorf("invokedReq = %q, want '{\"key\":\"value\"}'", invokedReq)
	}
}


// TestSQSEngineMultipleMessages tests processing multiple messages
func TestSQSEngineMultipleMessages(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "m1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/health-check"})},
		{MessageId: "m2", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/"})},
		{MessageId: "m3", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/health-check"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures, got %d", len(resp.BatchItemFailures))
	}
}
