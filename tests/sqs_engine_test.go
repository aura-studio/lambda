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
		lambdasqs.WithStaticLink("/static", "/public"),
		lambdasqs.WithPrefixLink("/api", "/v1"),
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

	if e.StaticLinkMap["/static"] != "/public" {
		t.Errorf("StaticLinkMap['/static'] = %q, want '/public'", e.StaticLinkMap["/static"])
	}

	if e.PrefixLinkMap["/api"] != "/v1" {
		t.Errorf("PrefixLinkMap['/api'] = %q, want '/v1'", e.PrefixLinkMap["/api"])
	}
}

// TestSQSEngineStartStop tests the Start and Stop methods
func TestSQSEngineStartStop(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	// Engine should be running after creation
	// Test by sending a message
	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/health-check"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures when running, got %d", len(resp.BatchItemFailures))
	}

	// Stop the engine
	e.Stop()

	// Messages should fail when stopped
	resp, err = e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("Expected 1 failure when stopped, got %d", len(resp.BatchItemFailures))
	}

	// Start the engine again
	e.Start()

	// Messages should succeed again
	resp, err = e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures after restart, got %d", len(resp.BatchItemFailures))
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

// TestSQSEngineInvokeStaticLink tests static link path mapping
func TestSQSEngineInvokeStaticLink(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithStaticLink("/custom-health", "/health-check"),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "static", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/custom-health"})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures (static link should map to health-check), got %d", len(resp.BatchItemFailures))
	}
}

// TestSQSEngineInvokePrefixLink tests prefix link path mapping
func TestSQSEngineInvokePrefixLink(t *testing.T) {
	mock := &mockSQSClient{}

	dynamic.RegisterPackage("prefixpkg-sqs", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return "prefix-response"
		},
	})

	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithPrefixLink("/v1", "/api"),
	}, nil)

	// /v1/prefixpkg-sqs/v1/route should be mapped to /api/prefixpkg-sqs/v1/route
	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "prefix", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/v1/prefixpkg-sqs/v1/route", Payload: []byte(`{}`)})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures, got %d", len(resp.BatchItemFailures))
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

// TestSQSEngineInvokeWAPI tests WAPI route
func TestSQSEngineInvokeWAPI(t *testing.T) {
	mock := &mockSQSClient{}

	var invokedRoute string
	dynamic.RegisterPackage("sqswapi", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedRoute = route
			return "wapi-response"
		},
	})

	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "wapi", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/wapi/sqswapi/v1/route", Payload: []byte(`{}`)})},
	}}

	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("HandleSQSMessagesWithResponse error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("Expected 0 failures, got %d", len(resp.BatchItemFailures))
	}

	if invokedRoute != "/route" {
		t.Errorf("invokedRoute = %q, want '/route'", invokedRoute)
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
