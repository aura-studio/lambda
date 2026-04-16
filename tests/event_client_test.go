package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/aura-studio/lambda/event"
	"github.com/aura-studio/lambda/event/client"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// mockLambdaClientForEvent captures the InvokeInput for verification
type mockLambdaClientForEvent struct {
	mu            sync.Mutex
	capturedInput *lambda.InvokeInput
}

func (m *mockLambdaClientForEvent) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capturedInput = params
	return &lambda.InvokeOutput{
		StatusCode: 202,
	}, nil
}

// TestProperty10_ClientSendConstructsCorrectInvokeInput verifies that Client.Send
// constructs the correct InvokeInput with InvocationType=Event, correct FunctionName,
// and Payload that deserializes to a Request with matching Path and Payload.
//
// **Validates: Requirements 9.4, 9.5, 9.6**
func TestProperty10_ClientSendConstructsCorrectInvokeInput(t *testing.T) {
	functionName := "test-event-function"

	f := func(path string, payload []byte) bool {
		mock := &mockLambdaClientForEvent{}
		cli := client.NewClient(
			client.WithLambdaClient(mock),
			client.WithFunctionName(functionName),
			client.WithDefaultTimeout(5*time.Second),
		)

		err := cli.Send(context.Background(), path, payload)
		if err != nil {
			t.Logf("Send returned error: %v", err)
			return false
		}

		mock.mu.Lock()
		captured := mock.capturedInput
		mock.mu.Unlock()

		if captured == nil {
			t.Log("InvokeInput was not captured")
			return false
		}

		// (1) InvocationType must be Event
		if captured.InvocationType != types.InvocationTypeEvent {
			t.Logf("Expected InvocationType %q, got %q", types.InvocationTypeEvent, captured.InvocationType)
			return false
		}

		// (2) FunctionName must match the configured function name
		if captured.FunctionName == nil || *captured.FunctionName != functionName {
			t.Logf("Expected FunctionName %q, got %v", functionName, captured.FunctionName)
			return false
		}

		// (3) Payload deserializes to event.Request with matching Path and Payload
		var req event.Request
		if err := json.Unmarshal(captured.Payload, &req); err != nil {
			t.Logf("Failed to unmarshal payload: %v", err)
			return false
		}

		if req.Path != path {
			t.Logf("Expected Path %q, got %q", path, req.Path)
			return false
		}

		// Handle nil vs empty slice comparison for payload
		if len(payload) == 0 && len(req.Payload) == 0 {
			return true
		}
		if string(req.Payload) != string(payload) {
			t.Logf("Expected Payload %v, got %v", payload, req.Payload)
			return false
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property 10 failed: %v", err)
	}
}

// --- Additional mock types for unit tests ---

// mockLambdaClientError always returns an error from Invoke
type mockLambdaClientError struct {
	err error
}

func (m *mockLambdaClientError) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	return nil, m.err
}

// mockLambdaClientFunctionError returns an InvokeOutput with FunctionError set
type mockLambdaClientFunctionError struct {
	functionError string
	payload       []byte
}

func (m *mockLambdaClientFunctionError) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	return &lambda.InvokeOutput{
		StatusCode:    200,
		FunctionError: &m.functionError,
		Payload:       m.payload,
	}, nil
}

// mockLambdaClientSlow simulates a slow Lambda invocation that respects context cancellation
type mockLambdaClientSlow struct {
	delay time.Duration
}

func (m *mockLambdaClientSlow) Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	select {
	case <-time.After(m.delay):
		return &lambda.InvokeOutput{StatusCode: 202}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// --- Unit tests for Client error handling, SendAsync, and timeout ---

// TestEventClient_SendInvokeFailure verifies that when the Lambda Invoke call
// returns an error, Send returns an error containing "lambda invoke failed".
//
// **Validates: Requirements 9.7**
func TestEventClient_SendInvokeFailure(t *testing.T) {
	mock := &mockLambdaClientError{
		err: fmt.Errorf("connection refused"),
	}
	cli := client.NewClient(
		client.WithLambdaClient(mock),
		client.WithFunctionName("test-function"),
		client.WithDefaultTimeout(5*time.Second),
	)

	err := cli.Send(context.Background(), "/test", []byte("hello"))
	if err == nil {
		t.Fatal("Expected error from Send, got nil")
	}
	if !strings.Contains(err.Error(), "lambda invoke failed") {
		t.Fatalf("Expected error to contain 'lambda invoke failed', got: %v", err)
	}
}

// TestEventClient_SendFunctionError verifies that when the Lambda Invoke returns
// an InvokeOutput with FunctionError set, Send returns an error containing the function error.
//
// **Validates: Requirements 9.8**
func TestEventClient_SendFunctionError(t *testing.T) {
	mock := &mockLambdaClientFunctionError{
		functionError: "Unhandled",
		payload:       []byte("some error details"),
	}
	cli := client.NewClient(
		client.WithLambdaClient(mock),
		client.WithFunctionName("test-function"),
		client.WithDefaultTimeout(5*time.Second),
	)

	err := cli.Send(context.Background(), "/test", []byte("hello"))
	if err == nil {
		t.Fatal("Expected error from Send, got nil")
	}
	if !strings.Contains(err.Error(), "Unhandled") {
		t.Fatalf("Expected error to contain 'Unhandled', got: %v", err)
	}
	if !strings.Contains(err.Error(), "some error details") {
		t.Fatalf("Expected error to contain 'some error details', got: %v", err)
	}
}

// TestEventClient_SendAsyncCallback verifies that SendAsync calls the callback
// with the result of Send.
//
// **Validates: Requirements 9.9**
func TestEventClient_SendAsyncCallback(t *testing.T) {
	mock := &mockLambdaClientForEvent{}
	cli := client.NewClient(
		client.WithLambdaClient(mock),
		client.WithFunctionName("test-function"),
		client.WithDefaultTimeout(5*time.Second),
	)

	var wg sync.WaitGroup
	wg.Add(1)

	var callbackErr error
	var callbackCalled bool

	cli.SendAsync(context.Background(), "/test", []byte("hello"), func(err error) {
		defer wg.Done()
		callbackCalled = true
		callbackErr = err
	})

	wg.Wait()

	if !callbackCalled {
		t.Fatal("Expected callback to be called")
	}
	if callbackErr != nil {
		t.Fatalf("Expected nil error in callback, got: %v", callbackErr)
	}
}

// TestEventClient_SendTimeout verifies that when the context deadline is exceeded,
// Send returns a "request timeout" error.
//
// **Validates: Requirements 9.7**
func TestEventClient_SendTimeout(t *testing.T) {
	mock := &mockLambdaClientSlow{
		delay: 5 * time.Second,
	}
	cli := client.NewClient(
		client.WithLambdaClient(mock),
		client.WithFunctionName("test-function"),
		client.WithDefaultTimeout(5*time.Second),
	)

	// Use a context with a very short deadline
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := cli.Send(ctx, "/test", []byte("hello"))
	if err == nil {
		t.Fatal("Expected timeout error from Send, got nil")
	}
	if !strings.Contains(err.Error(), "request timeout") {
		t.Fatalf("Expected error to contain 'request timeout', got: %v", err)
	}
}
