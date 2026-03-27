package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	lambdasqs "github.com/aura-studio/lambda/sqs"
	events "github.com/aws/aws-lambda-go/events"
)

func mustJSONSQSRequest(path string, payload string) string {
	b, _ := json.Marshal(map[string]any{
		"path":    path,
		"payload": base64.StdEncoding.EncodeToString([]byte(payload)),
	})
	return string(b)
}

// =============================================================================
// SQS: Envelope protocol
// =============================================================================

func TestSQS_Envelope_TunnelReceivesEnvelope(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("sqs-env-recv", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-env-recv/v1/test", "hello-sqs")},
	}}
	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("failures: %d", len(resp.BatchItemFailures))
	}

	env := decodeEnvelope(t, receivedReq)
	data, _ := base64.StdEncoding.DecodeString(env.Data)
	if string(data) != "hello-sqs" {
		t.Errorf("data = %q, want 'hello-sqs'", string(data))
	}
}

func TestSQS_Envelope_ResponseDecoded(t *testing.T) {
	dynamic.RegisterPackage("sqs-env-rsp", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return encodeEnvelope(nil, `{"status":"done"}`)
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithReplyMode(true),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-env-rsp/v1/test", "{}")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("failures: %d", len(resp.BatchItemFailures))
	}
}

func TestSQS_Envelope_RspMetaError(t *testing.T) {
	dynamic.RegisterPackage("sqs-env-err", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return encodeEnvelope(map[string]any{"Error": "sqs error"}, "")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-env-err/v1/test", "{}")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("failures = %d, want 1", len(resp.BatchItemFailures))
	}
}

// =============================================================================
// SQS: Routing
// =============================================================================

func TestSQS_Route_HealthCheck(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/health-check", "")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("failures = %d, want 0", len(resp.BatchItemFailures))
	}
}

func TestSQS_Route_Root(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/", "")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("failures = %d, want 0", len(resp.BatchItemFailures))
	}
}

func TestSQS_Route_NotFound(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/nonexistent", "")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("failures = %d, want 1", len(resp.BatchItemFailures))
	}
}

func TestSQS_Route_APIMultiSegment(t *testing.T) {
	var invokedRoute string
	dynamic.RegisterPackage("sqs-multi", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedRoute = route
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-multi/v1/users/456/settings", "{}")},
	}}
	e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if invokedRoute != "/users/456/settings" {
		t.Errorf("route = %q, want '/users/456/settings'", invokedRoute)
	}
}

func TestSQS_Route_PackageNotFound(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/nonexistent-sqs/v1/route", "{}")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("failures = %d, want 1", len(resp.BatchItemFailures))
	}
}

// =============================================================================
// SQS: Panic recovery
// =============================================================================

func TestSQS_PanicRecovery(t *testing.T) {
	dynamic.RegisterPackage("sqs-panic", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			panic("sqs-boom")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-panic/v1/test", "{}")},
	}}
	resp, err := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if err != nil {
		t.Fatalf("should not return Go error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("failures = %d, want 1 (panic should be caught)", len(resp.BatchItemFailures))
	}
}

// =============================================================================
// SQS: Run modes
// =============================================================================

func TestSQS_RunMode_Strict(t *testing.T) {
	dynamic.RegisterPackage("sqs-strict", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			if route == "/fail" {
				panic("fail")
			}
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModeStrict),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-strict/v1/ok", "{}")},
		{MessageId: "2", Body: mustJSONSQSRequest("/api/sqs-strict/v1/fail", "{}")},
		{MessageId: "3", Body: mustJSONSQSRequest("/api/sqs-strict/v1/ok", "{}")},
	}}
	resp, _ := e.Invoke(context.Background(), ev)
	// Strict: msg-2 fails, msg-2 and msg-3 should be in failures
	if len(resp.BatchItemFailures) != 2 {
		t.Errorf("failures = %d, want 2", len(resp.BatchItemFailures))
	}
}

func TestSQS_RunMode_Partial(t *testing.T) {
	dynamic.RegisterPackage("sqs-partial", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			if route == "/fail" {
				panic("fail")
			}
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-partial/v1/ok", "{}")},
		{MessageId: "2", Body: mustJSONSQSRequest("/api/sqs-partial/v1/fail", "{}")},
		{MessageId: "3", Body: mustJSONSQSRequest("/api/sqs-partial/v1/ok", "{}")},
	}}
	resp, _ := e.Invoke(context.Background(), ev)
	// Partial: only msg-2 fails
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("failures = %d, want 1", len(resp.BatchItemFailures))
	}
	if resp.BatchItemFailures[0].ItemIdentifier != "2" {
		t.Errorf("failed item = %q, want '2'", resp.BatchItemFailures[0].ItemIdentifier)
	}
}

func TestSQS_RunMode_Batch(t *testing.T) {
	dynamic.RegisterPackage("sqs-batch", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			if route == "/fail" {
				panic("batch-fail")
			}
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModeBatch),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-batch/v1/ok", "{}")},
		{MessageId: "2", Body: mustJSONSQSRequest("/api/sqs-batch/v1/fail", "{}")},
	}}
	_, err := e.Invoke(context.Background(), ev)
	if err == nil {
		t.Fatal("Batch mode should return error")
	}
	if !strings.Contains(err.Error(), "batch-fail") {
		t.Errorf("error = %q, want to contain 'batch-fail'", err.Error())
	}
}

func TestSQS_RunMode_Reentrant(t *testing.T) {
	dynamic.RegisterPackage("sqs-reentrant", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			if route == "/fail" {
				panic("reentrant-fail")
			}
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModeReentrant),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-reentrant/v1/ok", "{}")},
		{MessageId: "2", Body: mustJSONSQSRequest("/api/sqs-reentrant/v1/fail", "{}")},
		{MessageId: "3", Body: mustJSONSQSRequest("/api/sqs-reentrant/v1/ok", "{}")},
	}}
	_, err := e.Invoke(context.Background(), ev)
	if err == nil {
		t.Fatal("Reentrant mode should return error")
	}
	if !strings.Contains(err.Error(), "reentrant-fail") {
		t.Errorf("error = %q, want to contain 'reentrant-fail'", err.Error())
	}
}

// =============================================================================
// SQS: Invalid message body
// =============================================================================

func TestSQS_InvalidMessageBody(t *testing.T) {
	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{
		lambdasqs.WithSQSClient(mock),
		lambdasqs.WithRunMode(lambdasqs.RunModePartial),
	}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "bad", Body: "not-json"},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 1 {
		t.Errorf("failures = %d, want 1", len(resp.BatchItemFailures))
	}
}

// =============================================================================
// SQS: Meta endpoint
// =============================================================================

func TestSQS_Meta(t *testing.T) {
	dynamic.RegisterPackage("sqs-meta", "v1", &mockTunnel{
		invoke: func(route, req string) string { return "unused" },
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/meta/sqs-meta/v1", "")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("failures = %d, want 0", len(resp.BatchItemFailures))
	}
}

// =============================================================================
// SQS: Multiple messages
// =============================================================================

func TestSQS_MultipleMessages_AllSuccess(t *testing.T) {
	dynamic.RegisterPackage("sqs-multi-ok", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			return encodeEnvelope(nil, "ok")
		},
	})

	mock := &mockSQSClient{}
	e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithSQSClient(mock)}, nil)

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "1", Body: mustJSONSQSRequest("/api/sqs-multi-ok/v1/a", "{}")},
		{MessageId: "2", Body: mustJSONSQSRequest("/api/sqs-multi-ok/v1/b", "{}")},
		{MessageId: "3", Body: mustJSONSQSRequest("/health-check", "")},
	}}
	resp, _ := e.HandleSQSMessagesWithResponse(context.Background(), ev)
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("failures = %d, want 0", len(resp.BatchItemFailures))
	}
}
