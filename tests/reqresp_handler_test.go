package tests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/reqresp"
)

// reqresp_handler_test.go - Integration tests for reqresp handlers

// mockReqRespTunnel implements dynamic.Tunnel for testing
type mockReqRespTunnel struct {
	invokeFunc func(route, req string) string
}

func (m *mockReqRespTunnel) Init() {}

func (m *mockReqRespTunnel) Invoke(route string, req string) string {
	if m.invokeFunc != nil {
		return m.invokeFunc(route, req)
	}
	return "mock-response"
}

func (m *mockReqRespTunnel) Meta() string {
	return "mock-meta"
}

func (m *mockReqRespTunnel) Close() {}

// =============================================================================
// Health Check Route Tests
// =============================================================================

// TestReqRespHandler_HealthCheck_ReturnsOK tests that /health-check returns "OK"
func TestReqRespHandler_HealthCheck_ReturnsOK(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/health-check",
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// TestReqRespHandler_HealthCheck_WithPayload tests health check ignores payload
func TestReqRespHandler_HealthCheck_WithPayload(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/health-check",
		Payload: []byte("ignored-payload"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// TestReqRespHandler_RootPath_ReturnsOK tests that / returns "OK"
func TestReqRespHandler_RootPath_ReturnsOK(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/",
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// =============================================================================
// API Route Tests
// =============================================================================

// TestReqRespHandler_API_CallsDynamic tests that /api/* calls Dynamic and returns response
func TestReqRespHandler_API_CallsDynamic(t *testing.T) {
	var invokedRoute, invokedReq string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			return "api-response-data"
		},
	}

	dynamic.RegisterPackage("testpkg", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "testpkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/testpkg/v1/myroute",
		Payload: []byte("request-data"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	if string(resp.Payload) != "api-response-data" {
		t.Errorf("Payload = %q, want 'api-response-data'", string(resp.Payload))
	}

	if invokedRoute != "/myroute" {
		t.Errorf("invokedRoute = %q, want '/myroute'", invokedRoute)
	}

	if invokedReq != "request-data" {
		t.Errorf("invokedReq = %q, want 'request-data'", invokedReq)
	}
}

// TestReqRespHandler_API_MultipleRouteSegments tests API with multiple route segments
func TestReqRespHandler_API_MultipleRouteSegments(t *testing.T) {
	var invokedRoute string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			return "multi-segment-response"
		},
	}

	dynamic.RegisterPackage("multipkg", "v2", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "multipkg",
			Version: "v2",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/multipkg/v2/users/123/profile",
		Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	if invokedRoute != "/users/123/profile" {
		t.Errorf("invokedRoute = %q, want '/users/123/profile'", invokedRoute)
	}
}

// TestReqRespHandler_API_NoRouteSegment tests API with no route segment
func TestReqRespHandler_API_NoRouteSegment(t *testing.T) {
	var invokedRoute string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			return "no-route-response"
		},
	}

	dynamic.RegisterPackage("noroutepkg", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "noroutepkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/noroutepkg/v1",
		Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	if invokedRoute != "/" {
		t.Errorf("invokedRoute = %q, want '/'", invokedRoute)
	}
}

// TestReqRespHandler_API_PackageNotFound tests non-existent package
func TestReqRespHandler_API_PackageNotFound(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/nonexistent/v1/route",
		Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error for non-existent package")
	}
}


// =============================================================================
// Debug Mode Tests
// =============================================================================

// TestReqRespHandler_DebugMode_ReturnsDebugJSON tests that /_/api/* returns debug JSON format
func TestReqRespHandler_DebugMode_ReturnsDebugJSON(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return "debug-api-response"
		},
	}

	dynamic.RegisterPackage("debugpkg", "v1", tunnel)

	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "debugpkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/_/api/debugpkg/v1/route",
		Payload: []byte("debug-request"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	var debugInfo map[string]interface{}
	if err := json.Unmarshal(resp.Payload, &debugInfo); err != nil {
		t.Fatalf("Failed to unmarshal debug JSON: %v", err)
	}

	if _, ok := debugInfo["mode"]; !ok {
		t.Error("Debug response missing 'mode' field")
	}
	if _, ok := debugInfo["raw_path"]; !ok {
		t.Error("Debug response missing 'raw_path' field")
	}
	if _, ok := debugInfo["path"]; !ok {
		t.Error("Debug response missing 'path' field")
	}
	if _, ok := debugInfo["param"]; !ok {
		t.Error("Debug response missing 'param' field")
	}
	if _, ok := debugInfo["request"]; !ok {
		t.Error("Debug response missing 'request' field")
	}
	if _, ok := debugInfo["response"]; !ok {
		t.Error("Debug response missing 'response' field")
	}

	if mode, ok := debugInfo["mode"].(string); !ok || mode != "api" {
		t.Errorf("Debug mode = %v, want 'api'", debugInfo["mode"])
	}

	if debugInfo["response"] != "debug-api-response" {
		t.Errorf("Debug response = %v, want 'debug-api-response'", debugInfo["response"])
	}
}

// TestReqRespHandler_DebugMode_WithError tests debug mode includes error info
func TestReqRespHandler_DebugMode_WithError(t *testing.T) {
	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/_/api/nonexistent/v1/route",
		Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	var debugInfo map[string]interface{}
	if err := json.Unmarshal(resp.Payload, &debugInfo); err != nil {
		t.Fatalf("Failed to unmarshal debug JSON: %v", err)
	}

	if errField, ok := debugInfo["error"].(string); !ok || errField == "" {
		t.Error("Debug response should contain error for failed request")
	}
}


// =============================================================================
// Unmatched Route Tests (404)
// =============================================================================

// TestReqRespHandler_UnmatchedRoute_Returns404 tests that unmatched routes return 404 error
func TestReqRespHandler_UnmatchedRoute_Returns404(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/nonexistent/path",
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error for unmatched route")
	}

	if !strings.Contains(resp.Error, "404") {
		t.Errorf("Error = %q, expected to contain '404'", resp.Error)
	}
}

// TestReqRespHandler_UnmatchedRoute_VariousPaths tests various unmatched paths
func TestReqRespHandler_UnmatchedRoute_VariousPaths(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	testCases := []struct {
		name string
		path string
	}{
		{"random path", "/random/path/here"},
		{"partial api", "/ap/pkg/v1/route"},
		{"typo in api", "/apii/pkg/v1/route"},
		{"empty segments", "/foo//bar"},
		{"special chars", "/path-with-dash"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := engine.Invoke(context.Background(), &reqresp.Request{
				Path: tc.path,
			})
			if err != nil {
				t.Fatalf("Invoke returned error: %v", err)
			}

			if resp.Error == "" {
				t.Errorf("Expected error for unmatched path: %s", tc.path)
			}
		})
	}
}

// =============================================================================
// Additional Integration Tests
// =============================================================================

// TestReqRespHandler_EngineStopped tests that stopped engine returns error
func TestReqRespHandler_EngineStopped(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	engine.Stop()

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/health-check",
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error == "" {
		t.Error("Expected error when engine is stopped")
	}

	if resp.Error != "engine is stopped" {
		t.Errorf("Error = %q, want 'engine is stopped'", resp.Error)
	}
}

// TestReqRespHandler_StaticLink tests static link path mapping
func TestReqRespHandler_StaticLink(t *testing.T) {
	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithStaticLink("/custom-health", "/health-check"),
	}, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/custom-health",
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK' (static link should map to health-check)", string(resp.Payload))
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// TestReqRespHandler_PrefixLink tests prefix link path mapping
func TestReqRespHandler_PrefixLink(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return "prefix-response"
		},
	}

	dynamic.RegisterPackage("prefixpkg", "v1", tunnel)

	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithPrefixLink("/v1", "/api"),
	}, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "prefixpkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/v1/prefixpkg/v1/route",
		Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	if string(resp.Payload) != "prefix-response" {
		t.Errorf("Payload = %q, want 'prefix-response'", string(resp.Payload))
	}
}
