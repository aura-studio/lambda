package tests

import (
	"context"
	"testing"

	"github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/reqresp"
	"google.golang.org/protobuf/proto"
)

// TestEngineCreation tests that NewEngine creates an engine with correct options
func TestEngineCreation(t *testing.T) {
	reqrespOpts := []reqresp.Option{
		reqresp.WithDebugMode(true),
		reqresp.WithStaticLink("/static", "/public"),
		reqresp.WithPrefixLink("/api", "/v1"),
	}
	dynamicOpts := []dynamic.Option{}

	engine := reqresp.NewEngine(reqrespOpts, dynamicOpts)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if !engine.DebugMode {
		t.Error("DebugMode should be true")
	}

	if engine.StaticLinkMap["/static"] != "/public" {
		t.Errorf("StaticLinkMap['/static'] = %q, want '/public'", engine.StaticLinkMap["/static"])
	}

	if engine.PrefixLinkMap["/api"] != "/v1" {
		t.Errorf("PrefixLinkMap['/api'] = %q, want '/v1'", engine.PrefixLinkMap["/api"])
	}

	if !engine.IsRunning() {
		t.Error("Engine should be running after creation")
	}
}

// TestEngineStartStop tests the Start and Stop methods
func TestEngineStartStop(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	if !engine.IsRunning() {
		t.Error("Engine should be running after creation")
	}

	engine.Stop()
	if engine.IsRunning() {
		t.Error("Engine should not be running after Stop")
	}

	engine.Start()
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start")
	}
}

// TestEngineInvokeWhenStopped tests that Invoke returns error when engine is stopped
func TestEngineInvokeWhenStopped(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	engine.Stop()

	req := &reqresp.Request{
		CorrelationId: "test-123",
		Path:          "/health-check",
		Payload:       []byte("test"),
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error in response when engine is stopped")
	}

	if resp.Error != "engine is stopped" {
		t.Errorf("Expected error 'engine is stopped', got %q", resp.Error)
	}
}


// TestEngineInvokeHealthCheck tests the health check endpoint
func TestEngineInvokeHealthCheck(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	req := &reqresp.Request{
		CorrelationId: "test-123",
		Path:          "/health-check",
		Payload:       nil,
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.CorrelationId != "test-123" {
		t.Errorf("CorrelationId = %q, want 'test-123'", resp.CorrelationId)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// TestEngineInvokeRootPath tests the root path endpoint
func TestEngineInvokeRootPath(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	req := &reqresp.Request{
		CorrelationId: "test-456",
		Path:          "/",
		Payload:       nil,
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}
}

// TestEngineInvokePageNotFound tests the 404 handler
func TestEngineInvokePageNotFound(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	req := &reqresp.Request{
		CorrelationId: "test-789",
		Path:          "/nonexistent/path",
		Payload:       nil,
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error for nonexistent path")
	}

	if resp.CorrelationId != "test-789" {
		t.Errorf("CorrelationId = %q, want 'test-789'", resp.CorrelationId)
	}
}

// TestEngineInvokeInvalidPayload tests handling of invalid protobuf payload
func TestEngineInvokeInvalidPayload(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	// Send invalid protobuf data
	respBytes, err := engine.Invoke(context.Background(), []byte("invalid protobuf"))
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error for invalid payload")
	}
}

// TestEngineInvokeStaticLink tests static link path mapping
func TestEngineInvokeStaticLink(t *testing.T) {
	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithStaticLink("/custom-health", "/health-check"),
	}, nil)

	req := &reqresp.Request{
		CorrelationId: "test-static",
		Path:          "/custom-health",
		Payload:       nil,
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK' (static link should map to health-check)", string(resp.Payload))
	}
}

// TestEngineInvokeAPIPathMissingPackage tests API path with missing package
func TestEngineInvokeAPIPathMissingPackage(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	req := &reqresp.Request{
		CorrelationId: "test-api",
		Path:          "/api/nonexistent-pkg/v1/route",
		Payload:       []byte("{}"),
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return error because package doesn't exist
	if resp.Error == "" {
		t.Error("Expected error for nonexistent package")
	}
}

// TestEngineInvokeAPIPathInvalid tests API path with invalid format
func TestEngineInvokeAPIPathInvalid(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	req := &reqresp.Request{
		CorrelationId: "test-api-invalid",
		Path:          "/api/", // Missing package and version
		Payload:       []byte("{}"),
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should return error because path is invalid
	if resp.Error == "" {
		t.Error("Expected error for invalid API path")
	}
}

// TestEngineInvokeDebugMode tests debug mode response format
func TestEngineInvokeDebugMode(t *testing.T) {
	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, nil)

	req := &reqresp.Request{
		CorrelationId: "test-debug",
		Path:          "/_/api/nonexistent/v1/route",
		Payload:       []byte("test-request"),
	}
	payload, _ := proto.Marshal(req)

	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var resp reqresp.Response
	if err := proto.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Debug mode should include debug info in response
	// Even with error, the response payload should contain debug JSON
	if len(resp.Payload) == 0 && resp.Error == "" {
		t.Error("Expected either payload or error in debug mode response")
	}
}
