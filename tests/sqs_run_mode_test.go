package tests

import (
	"context"
	"testing"

	"github.com/aura-studio/dynamic"
	lambdasqs "github.com/aura-studio/lambda/sqs"
	events "github.com/aws/aws-lambda-go/events"
)

func TestSQSRunMode(t *testing.T) {
	// Register a package that returns error for specific routes
	dynamic.RegisterPackage("mode-test", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			if route == "/fail" {
				panic("intentional failure")
			}
			return "OK"
		},
	})

	ev := events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "msg-1", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/mode-test/v1/success", Payload: []byte(`{}`)})},
		{MessageId: "msg-2", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/mode-test/v1/fail", Payload: []byte(`{}`)})},
		{MessageId: "msg-3", Body: mustPBRequest(t, &lambdasqs.Request{Path: "/api/mode-test/v1/success", Payload: []byte(`{}`)})},
	}}

	t.Run("Strict", func(t *testing.T) {
		e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithRunMode(lambdasqs.RunModeStrict)}, nil)
		resp, err := e.Invoke(context.Background(), ev)
		if err != nil {
			t.Fatalf("Strict mode should not return error, got %v", err)
		}
		// msg-2 fails, so msg-2 and msg-3 should be in BatchItemFailures
		if len(resp.BatchItemFailures) != 2 {
			t.Errorf("Expected 2 failures, got %d", len(resp.BatchItemFailures))
		}
		if resp.BatchItemFailures[0].ItemIdentifier != "msg-2" || resp.BatchItemFailures[1].ItemIdentifier != "msg-3" {
			t.Errorf("Unexpected failures: %+v", resp.BatchItemFailures)
		}
	})

	t.Run("Partial", func(t *testing.T) {
		e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithRunMode(lambdasqs.RunModePartial)}, nil)
		resp, err := e.Invoke(context.Background(), ev)
		if err != nil {
			t.Fatalf("Partial mode should not return error, got %v", err)
		}
		// Only msg-2 fails
		if len(resp.BatchItemFailures) != 1 {
			t.Errorf("Expected 1 failure, got %d", len(resp.BatchItemFailures))
		}
		if resp.BatchItemFailures[0].ItemIdentifier != "msg-2" {
			t.Errorf("Unexpected failure: %+v", resp.BatchItemFailures)
		}
	})

	t.Run("Batch", func(t *testing.T) {
		e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithRunMode(lambdasqs.RunModeBatch)}, nil)
		_, err := e.Invoke(context.Background(), ev)
		if err == nil {
			t.Fatal("Batch mode should return error on first failure")
		}
		if err.Error() != "panic: intentional failure" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("Reentrant", func(t *testing.T) {
		e := lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithRunMode(lambdasqs.RunModeReentrant)}, nil)
		_, err := e.Invoke(context.Background(), ev)
		if err == nil {
			t.Fatal("Reentrant mode should return error at the end")
		}
		if err.Error() != "panic: intentional failure" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("InvalidMode", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic for invalid run mode")
			}
		}()
		lambdasqs.NewEngine([]lambdasqs.Option{lambdasqs.WithRunMode(lambdasqs.RunMode("invalid"))}, nil)
	})
}
