package tests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/reqresp"
	"google.golang.org/protobuf/proto"
)

// reqresp_handler_test.go - Integration tests for reqresp handlers
// **Validates: Requirements 5.1, 5.2, 5.4, 5.5**

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

// Helper function to create and marshal a request
func createReqRespRequest(t *testing.T, correlationId, path string, payload []byte) []byte {
	t.Helper()
	req := &reqresp.Request{
		CorrelationId: correlationId,
		Path:          path,
		Payload:       payload,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	return data
}

// Helper function to unmarshal response
func unmarshalReqRespResponse(t *testing.T, data []byte) *reqresp.Response {
	t.Helper()
	var resp reqresp.Response
	if err := proto.Unmarshal(data, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	return &resp
}


// =============================================================================
// Health Check Route Tests
// **Validates: Requirements 5.1**
// =============================================================================

// TestReqRespHandler_HealthCheck_ReturnsOK tests that /health-check returns "OK"
func TestReqRespHandler_HealthCheck_ReturnsOK(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	payload := createReqRespRequest(t, "health-1", "/health-check", nil)
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.CorrelationId != "health-1" {
		t.Errorf("CorrelationId = %q, want 'health-1'", resp.CorrelationId)
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

	payload := createReqRespRequest(t, "health-2", "/health-check", []byte("ignored-payload"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

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

	payload := createReqRespRequest(t, "root-1", "/", nil)
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// =============================================================================
// API Route Tests
// **Validates: Requirements 5.2**
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

	// Register the mock package
	dynamic.RegisterPackage("testpkg", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "testpkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	payload := createReqRespRequest(t, "api-1", "/api/testpkg/v1/myroute", []byte("request-data"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.CorrelationId != "api-1" {
		t.Errorf("CorrelationId = %q, want 'api-1'", resp.CorrelationId)
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

	payload := createReqRespRequest(t, "api-2", "/api/multipkg/v2/users/123/profile", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

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

	payload := createReqRespRequest(t, "api-3", "/api/noroutepkg/v1", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	// When no route segment, the route should be "/"
	if invokedRoute != "/" {
		t.Errorf("invokedRoute = %q, want '/'", invokedRoute)
	}
}

// TestReqRespHandler_API_PackageNotFound tests API with non-existent package
func TestReqRespHandler_API_PackageNotFound(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	payload := createReqRespRequest(t, "api-notfound", "/api/nonexistent/v1/route", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error == "" {
		t.Error("Expected error for non-existent package")
	}
}

// TestReqRespHandler_WAPI_CallsDynamic tests that /wapi/* works as API alias
func TestReqRespHandler_WAPI_CallsDynamic(t *testing.T) {
	var invokedRoute string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			return "wapi-response"
		},
	}

	dynamic.RegisterPackage("wapipkg", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "wapipkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	payload := createReqRespRequest(t, "wapi-1", "/wapi/wapipkg/v1/myroute", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	if string(resp.Payload) != "wapi-response" {
		t.Errorf("Payload = %q, want 'wapi-response'", string(resp.Payload))
	}

	if invokedRoute != "/myroute" {
		t.Errorf("invokedRoute = %q, want '/myroute'", invokedRoute)
	}
}


// =============================================================================
// Debug Mode Tests
// **Validates: Requirements 5.4**
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

	payload := createReqRespRequest(t, "debug-1", "/_/api/debugpkg/v1/route", []byte("debug-request"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	// Debug mode should return JSON with debug fields
	var debugInfo map[string]interface{}
	if err := json.Unmarshal(resp.Payload, &debugInfo); err != nil {
		t.Fatalf("Failed to unmarshal debug JSON: %v", err)
	}

	// Check required debug fields
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

	// Verify mode is "api"
	if mode, ok := debugInfo["mode"].(string); !ok || mode != "api" {
		t.Errorf("Debug mode = %v, want 'api'", debugInfo["mode"])
	}

	// Verify response contains the actual response
	if response, ok := debugInfo["response"].(string); !ok || response != "debug-api-response" {
		t.Errorf("Debug response = %v, want 'debug-api-response'", debugInfo["response"])
	}
}

// TestReqRespHandler_DebugMode_WithError tests debug mode includes error info
func TestReqRespHandler_DebugMode_WithError(t *testing.T) {
	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, nil)

	// Use non-existent package to trigger error
	payload := createReqRespRequest(t, "debug-err", "/_/api/nonexistent/v1/route", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	// Debug mode should still return JSON even with error
	var debugInfo map[string]interface{}
	if err := json.Unmarshal(resp.Payload, &debugInfo); err != nil {
		t.Fatalf("Failed to unmarshal debug JSON: %v", err)
	}

	// Check error field is present and non-empty
	if errField, ok := debugInfo["error"].(string); !ok || errField == "" {
		t.Error("Debug response should have non-empty 'error' field for failed request")
	}
}

// TestReqRespHandler_DebugMode_WAPI tests debug mode with WAPI route
func TestReqRespHandler_DebugMode_WAPI(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return "debug-wapi-response"
		},
	}

	dynamic.RegisterPackage("debugwapi", "v1", tunnel)

	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "debugwapi",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	payload := createReqRespRequest(t, "debug-wapi", "/_/wapi/debugwapi/v1/route", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	// Debug mode should return JSON
	var debugInfo map[string]interface{}
	if err := json.Unmarshal(resp.Payload, &debugInfo); err != nil {
		t.Fatalf("Failed to unmarshal debug JSON: %v", err)
	}

	// Verify mode is "api" (WAPI is alias for API)
	if mode, ok := debugInfo["mode"].(string); !ok || mode != "api" {
		t.Errorf("Debug mode = %v, want 'api'", debugInfo["mode"])
	}
}

// =============================================================================
// Unmatched Route Tests (404)
// **Validates: Requirements 5.5**
// =============================================================================

// TestReqRespHandler_UnmatchedRoute_Returns404 tests that unmatched routes return 404 error
func TestReqRespHandler_UnmatchedRoute_Returns404(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	payload := createReqRespRequest(t, "404-1", "/nonexistent/path", nil)
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error == "" {
		t.Error("Expected error for unmatched route")
	}

	if !strings.Contains(resp.Error, "404") {
		t.Errorf("Error = %q, expected to contain '404'", resp.Error)
	}
}

// TestReqRespHandler_UnmatchedRoute_PreservesCorrelationId tests correlation ID is preserved
func TestReqRespHandler_UnmatchedRoute_PreservesCorrelationId(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	payload := createReqRespRequest(t, "corr-404", "/unknown/route", nil)
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.CorrelationId != "corr-404" {
		t.Errorf("CorrelationId = %q, want 'corr-404'", resp.CorrelationId)
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
			payload := createReqRespRequest(t, "test-"+tc.name, tc.path, nil)
			respBytes, err := engine.Invoke(context.Background(), payload)
			if err != nil {
				t.Fatalf("Invoke returned error: %v", err)
			}

			resp := unmarshalReqRespResponse(t, respBytes)

			if resp.Error == "" {
				t.Errorf("Expected error for unmatched path: %s", tc.path)
			}
		})
	}
}

// =============================================================================
// Additional Integration Tests
// =============================================================================

// TestReqRespHandler_InvalidProtobuf tests handling of invalid protobuf payload
func TestReqRespHandler_InvalidProtobuf(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	respBytes, err := engine.Invoke(context.Background(), []byte("not-valid-protobuf"))
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error == "" {
		t.Error("Expected error for invalid protobuf")
	}

	if !strings.Contains(resp.Error, "unmarshal") {
		t.Errorf("Error = %q, expected to contain 'unmarshal'", resp.Error)
	}
}

// TestReqRespHandler_EngineStopped tests that stopped engine returns error
func TestReqRespHandler_EngineStopped(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	engine.Stop()

	payload := createReqRespRequest(t, "stopped-1", "/health-check", nil)
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

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

	payload := createReqRespRequest(t, "static-1", "/custom-health", nil)
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

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

	// /v1/prefixpkg/v1/route should be mapped to /api/prefixpkg/v1/route
	payload := createReqRespRequest(t, "prefix-1", "/v1/prefixpkg/v1/route", []byte("{}"))
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke returned error: %v", err)
	}

	resp := unmarshalReqRespResponse(t, respBytes)

	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}

	if string(resp.Payload) != "prefix-response" {
		t.Errorf("Payload = %q, want 'prefix-response'", string(resp.Payload))
	}
}
