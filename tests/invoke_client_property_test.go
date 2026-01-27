package tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aura-studio/lambda/invoke"
	"github.com/aura-studio/lambda/invoke/invokecli"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/protobuf/proto"
)

// **Feature: invoke-lambda-handler, Property 9: 客户端调用正确性**
// **Validates: Requirements 7.1, 7.2, 7.5, 7.6**
//
// Property 9: 客户端调用正确性
// For any 客户端调用请求，如果 Lambda 调用成功，客户端 SHALL 返回正确的 Response；
// 如果调用超时或失败，客户端 SHALL 返回包含错误信息的结果。

// mockLambdaClient is a mock implementation of LambdaClient interface for testing
type mockLambdaClient struct {
	// Response configuration
	responsePayload []byte
	functionError   *string
	invokeError     error

	// Behavior configuration
	delay time.Duration

	// Tracking
	invokeCalled bool
	lastInput    *lambda.InvokeInput
	mu           sync.Mutex
}

// Invoke implements the LambdaClient interface
func (m *mockLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput,
	optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	m.mu.Lock()
	m.invokeCalled = true
	m.lastInput = params
	m.mu.Unlock()

	// Simulate delay if configured
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
			// Delay completed
		case <-ctx.Done():
			// Context cancelled or deadline exceeded
			return nil, ctx.Err()
		}
	}

	// Return error if configured
	if m.invokeError != nil {
		return nil, m.invokeError
	}

	// Return response
	return &lambda.InvokeOutput{
		Payload:       m.responsePayload,
		FunctionError: m.functionError,
	}, nil
}

// genClientPath generates valid request paths for client testing
func genClientPath() gopter.Gen {
	return gen.OneConstOf("/health-check", "/api/pkg/v1/route", "/test/path")
}

// genClientPayload generates random payloads for client testing
func genClientPayload() gopter.Gen {
	return gen.SliceOfN(100, gen.UInt8())
}

// TestClientCallSuccessReturnsCorrectResponse tests that successful Lambda calls return correct Response
// **Validates: Requirements 7.1**
func TestClientCallSuccessReturnsCorrectResponse(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client Call: successful Lambda call returns correct Response", prop.ForAll(
		func(path string, payload []byte, responsePayload []byte) bool {
			// Create expected response
			expectedResp := &invoke.Response{
				CorrelationId: "test-correlation-id",
				Payload:       responsePayload,
				Error:         "",
			}
			expectedRespBytes, err := proto.Marshal(expectedResp)
			if err != nil {
				t.Logf("Failed to marshal expected response: %v", err)
				return false
			}

			// Create mock client that returns success
			mockClient := &mockLambdaClient{
				responsePayload: expectedRespBytes,
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Call the client
			resp, err := client.Call(context.Background(), path, payload)
			if err != nil {
				t.Logf("Call returned unexpected error: %v", err)
				return false
			}

			// Verify response is not nil
			if resp == nil {
				t.Logf("Expected non-nil response")
				return false
			}

			// Verify response payload matches expected
			if string(resp.Payload) != string(responsePayload) {
				t.Logf("Response payload mismatch: expected %q, got %q", string(responsePayload), string(resp.Payload))
				return false
			}

			// Verify no error in response
			if resp.Error != "" {
				t.Logf("Expected no error in response, got %q", resp.Error)
				return false
			}

			// Verify mock was called
			if !mockClient.invokeCalled {
				t.Logf("Expected mock Invoke to be called")
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}

// TestClientCallTimeoutReturnsError tests that timeout errors are properly returned
// **Validates: Requirements 7.5**
func TestClientCallTimeoutReturnsError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client Call: timeout returns error", prop.ForAll(
		func(path string, payload []byte) bool {
			// Create mock client with delay longer than timeout
			mockClient := &mockLambdaClient{
				delay: 2 * time.Second, // Delay longer than timeout
			}

			// Create client with short timeout
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(50*time.Millisecond), // Very short timeout
			)

			// Call the client
			resp, err := client.Call(context.Background(), path, payload)

			// Should return error for timeout
			if err == nil {
				t.Logf("Expected error for timeout, got nil")
				return false
			}

			// Response should be nil on error
			if resp != nil {
				t.Logf("Expected nil response on timeout error")
				return false
			}

			// Error message should indicate timeout
			if err.Error() != "request timeout" {
				t.Logf("Expected 'request timeout' error, got %q", err.Error())
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}

// TestClientCallLambdaErrorReturnsErrorResponse tests that Lambda function errors are properly returned
// **Validates: Requirements 7.6**
func TestClientCallLambdaErrorReturnsErrorResponse(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client Call: Lambda function error returns Response with error field", prop.ForAll(
		func(path string, payload []byte, errorMsg string) bool {
			// Create mock client that returns function error
			mockClient := &mockLambdaClient{
				functionError: aws.String(errorMsg),
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Call the client
			resp, err := client.Call(context.Background(), path, payload)

			// Should not return Go error for function error
			if err != nil {
				t.Logf("Expected no Go error for function error, got %v", err)
				return false
			}

			// Response should not be nil
			if resp == nil {
				t.Logf("Expected non-nil response for function error")
				return false
			}

			// Response should contain the function error
			if resp.Error != errorMsg {
				t.Logf("Expected error %q in response, got %q", errorMsg, resp.Error)
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
		gen.OneConstOf("Unhandled", "Runtime.ExitError", "Function.ResponseSizeTooLarge"),
	))

	properties.TestingRun(t)
}

// TestClientCallInvokeErrorReturnsError tests that Lambda invoke errors are properly returned
// **Validates: Requirements 7.6**
func TestClientCallInvokeErrorReturnsError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client Call: Lambda invoke error returns Go error", prop.ForAll(
		func(path string, payload []byte, errorMsg string) bool {
			// Create mock client that returns invoke error
			mockClient := &mockLambdaClient{
				invokeError: errors.New(errorMsg),
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Call the client
			resp, err := client.Call(context.Background(), path, payload)

			// Should return Go error
			if err == nil {
				t.Logf("Expected Go error for invoke error, got nil")
				return false
			}

			// Response should be nil on error
			if resp != nil {
				t.Logf("Expected nil response on invoke error")
				return false
			}

			// Error should contain the original error message
			if !errors.Is(err, mockClient.invokeError) && err.Error() != "lambda invoke failed: "+errorMsg {
				t.Logf("Expected error to contain %q, got %q", errorMsg, err.Error())
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
		gen.OneConstOf("connection refused", "service unavailable", "access denied"),
	))

	properties.TestingRun(t)
}

// TestClientCallAsyncInvokesCallback tests that CallAsync invokes callback with results
// **Validates: Requirements 7.2**
func TestClientCallAsyncInvokesCallback(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client CallAsync: callback is invoked with results", prop.ForAll(
		func(path string, payload []byte, responsePayload []byte) bool {
			// Create expected response
			expectedResp := &invoke.Response{
				CorrelationId: "test-correlation-id",
				Payload:       responsePayload,
				Error:         "",
			}
			expectedRespBytes, err := proto.Marshal(expectedResp)
			if err != nil {
				t.Logf("Failed to marshal expected response: %v", err)
				return false
			}

			// Create mock client that returns success
			mockClient := &mockLambdaClient{
				responsePayload: expectedRespBytes,
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Track callback invocation
			var callbackResp *invoke.Response
			var callbackErr error
			callbackCalled := false
			var wg sync.WaitGroup
			wg.Add(1)

			// Call async
			client.CallAsync(context.Background(), path, payload, func(resp *invoke.Response, err error) {
				callbackResp = resp
				callbackErr = err
				callbackCalled = true
				wg.Done()
			})

			// Wait for callback with timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Callback completed
			case <-time.After(5 * time.Second):
				t.Logf("Callback was not invoked within timeout")
				return false
			}

			// Verify callback was called
			if !callbackCalled {
				t.Logf("Expected callback to be called")
				return false
			}

			// Verify no error in callback
			if callbackErr != nil {
				t.Logf("Expected no error in callback, got %v", callbackErr)
				return false
			}

			// Verify response in callback
			if callbackResp == nil {
				t.Logf("Expected non-nil response in callback")
				return false
			}

			// Verify response payload matches expected
			if string(callbackResp.Payload) != string(responsePayload) {
				t.Logf("Callback response payload mismatch: expected %q, got %q",
					string(responsePayload), string(callbackResp.Payload))
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}

// TestClientCallAsyncErrorInvokesCallback tests that CallAsync invokes callback with error
// **Validates: Requirements 7.2, 7.6**
func TestClientCallAsyncErrorInvokesCallback(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client CallAsync: callback is invoked with error on failure", prop.ForAll(
		func(path string, payload []byte, errorMsg string) bool {
			// Create mock client that returns invoke error
			mockClient := &mockLambdaClient{
				invokeError: errors.New(errorMsg),
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Track callback invocation
			var callbackResp *invoke.Response
			var callbackErr error
			callbackCalled := false
			var wg sync.WaitGroup
			wg.Add(1)

			// Call async
			client.CallAsync(context.Background(), path, payload, func(resp *invoke.Response, err error) {
				callbackResp = resp
				callbackErr = err
				callbackCalled = true
				wg.Done()
			})

			// Wait for callback with timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Callback completed
			case <-time.After(5 * time.Second):
				t.Logf("Callback was not invoked within timeout")
				return false
			}

			// Verify callback was called
			if !callbackCalled {
				t.Logf("Expected callback to be called")
				return false
			}

			// Verify error in callback
			if callbackErr == nil {
				t.Logf("Expected error in callback, got nil")
				return false
			}

			// Verify response is nil on error
			if callbackResp != nil {
				t.Logf("Expected nil response in callback on error")
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
		gen.OneConstOf("connection refused", "service unavailable", "access denied"),
	))

	properties.TestingRun(t)
}

// TestClientCallAsyncTimeoutInvokesCallback tests that CallAsync invokes callback with timeout error
// **Validates: Requirements 7.2, 7.5**
func TestClientCallAsyncTimeoutInvokesCallback(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client CallAsync: callback is invoked with timeout error", prop.ForAll(
		func(path string, payload []byte) bool {
			// Create mock client with delay longer than timeout
			mockClient := &mockLambdaClient{
				delay: 2 * time.Second, // Delay longer than timeout
			}

			// Create client with short timeout
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(50*time.Millisecond), // Very short timeout
			)

			// Track callback invocation
			var callbackResp *invoke.Response
			var callbackErr error
			callbackCalled := false
			var wg sync.WaitGroup
			wg.Add(1)

			// Call async
			client.CallAsync(context.Background(), path, payload, func(resp *invoke.Response, err error) {
				callbackResp = resp
				callbackErr = err
				callbackCalled = true
				wg.Done()
			})

			// Wait for callback with timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Callback completed
			case <-time.After(5 * time.Second):
				t.Logf("Callback was not invoked within timeout")
				return false
			}

			// Verify callback was called
			if !callbackCalled {
				t.Logf("Expected callback to be called")
				return false
			}

			// Verify error in callback
			if callbackErr == nil {
				t.Logf("Expected error in callback for timeout, got nil")
				return false
			}

			// Verify response is nil on error
			if callbackResp != nil {
				t.Logf("Expected nil response in callback on timeout error")
				return false
			}

			// Verify error message indicates timeout
			if callbackErr.Error() != "request timeout" {
				t.Logf("Expected 'request timeout' error, got %q", callbackErr.Error())
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}

// capturingMockLambdaClient is a mock that captures the request for inspection
type capturingMockLambdaClient struct {
	responsePayload  []byte
	capturedRequest  *invoke.Request
	mu               sync.Mutex
}

// Invoke implements the LambdaClient interface and captures the request
func (m *capturingMockLambdaClient) Invoke(ctx context.Context, params *lambda.InvokeInput,
	optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Capture the request
	if params.Payload != nil {
		m.capturedRequest = &invoke.Request{}
		proto.Unmarshal(params.Payload, m.capturedRequest)
	}

	return &lambda.InvokeOutput{
		Payload: m.responsePayload,
	}, nil
}

// TestClientCallPreservesCorrelationId tests that correlation ID is generated and preserved
// **Validates: Requirements 7.1**
func TestClientCallPreservesCorrelationId(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client Call: correlation ID is generated and sent in request", prop.ForAll(
		func(path string, payload []byte) bool {
			// Create response
			resp := &invoke.Response{
				CorrelationId: "response-correlation-id",
				Payload:       []byte("response"),
			}
			respBytes, _ := proto.Marshal(resp)

			// Create capturing mock client
			mockClient := &capturingMockLambdaClient{
				responsePayload: respBytes,
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Call the client
			_, err := client.Call(context.Background(), path, payload)
			if err != nil {
				t.Logf("Call returned unexpected error: %v", err)
				return false
			}

			// Verify request was captured
			mockClient.mu.Lock()
			capturedRequest := mockClient.capturedRequest
			mockClient.mu.Unlock()

			if capturedRequest == nil {
				t.Logf("Expected request to be captured")
				return false
			}

			// Verify correlation ID was generated (non-empty)
			if capturedRequest.CorrelationId == "" {
				t.Logf("Expected non-empty correlation ID in request")
				return false
			}

			// Verify path was set correctly
			if capturedRequest.Path != path {
				t.Logf("Expected path %q, got %q", path, capturedRequest.Path)
				return false
			}

			// Verify payload was set correctly
			if string(capturedRequest.Payload) != string(payload) {
				t.Logf("Expected payload %q, got %q", string(payload), string(capturedRequest.Payload))
				return false
			}

			return true
		},
		genClientPath(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}

// TestClientCallFunctionNameUsed tests that the configured function name is used
// **Validates: Requirements 7.4**
func TestClientCallFunctionNameUsed(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client Call: configured function name is used in Lambda invoke", prop.ForAll(
		func(functionName string, path string) bool {
			// Create mock client
			mockClient := &mockLambdaClient{
				responsePayload: func() []byte {
					resp := &invoke.Response{Payload: []byte("ok")}
					b, _ := proto.Marshal(resp)
					return b
				}(),
			}

			// Create client with mock and specific function name
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName(functionName),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Call the client
			_, err := client.Call(context.Background(), path, []byte("test"))
			if err != nil {
				t.Logf("Call returned unexpected error: %v", err)
				return false
			}

			// Verify function name was used
			mockClient.mu.Lock()
			lastInput := mockClient.lastInput
			mockClient.mu.Unlock()

			if lastInput == nil {
				t.Logf("Expected lastInput to be set")
				return false
			}

			if lastInput.FunctionName == nil || *lastInput.FunctionName != functionName {
				t.Logf("Expected function name %q, got %v", functionName, lastInput.FunctionName)
				return false
			}

			return true
		},
		gen.OneConstOf("my-lambda-function", "test-function", "prod-api-handler"),
		genClientPath(),
	))

	properties.TestingRun(t)
}

// TestClientCallNilCallbackHandled tests that nil callback is handled gracefully in CallAsync
// **Validates: Requirements 7.2**
func TestClientCallNilCallbackHandled(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42)

	properties := gopter.NewProperties(parameters)

	properties.Property("Client CallAsync: nil callback is handled gracefully", prop.ForAll(
		func(path string, payload []byte) bool {
			// Create mock client
			mockClient := &mockLambdaClient{
				responsePayload: func() []byte {
					resp := &invoke.Response{Payload: []byte("ok")}
					b, _ := proto.Marshal(resp)
					return b
				}(),
			}

			// Create client with mock
			client := invokecli.NewClient(
				invokecli.WithLambdaClient(mockClient),
				invokecli.WithFunctionName("test-function"),
				invokecli.WithDefaultTimeout(5*time.Second),
			)

			// Call async with nil callback - should not panic
			client.CallAsync(context.Background(), path, payload, nil)

			// Wait a short time for the goroutine to complete
			time.Sleep(10 * time.Millisecond)

			// If we get here without panic, the test passes
			return true
		},
		genClientPath(),
		genClientPayload(),
	))

	properties.TestingRun(t)
}
