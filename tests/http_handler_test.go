package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	lambdahttp "github.com/aura-studio/lambda/http"
)

// =============================================================================
// Health Check Route Tests
// =============================================================================

// TestHTTPHandler_HealthCheck_GET tests GET /health-check
func TestHTTPHandler_HealthCheck_GET(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health-check", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK'", string(body))
	}
}

// TestHTTPHandler_HealthCheck_POST tests POST /health-check
func TestHTTPHandler_HealthCheck_POST(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/health-check", strings.NewReader("ignored"))
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK'", string(body))
	}
}

// =============================================================================
// API Route Tests
// =============================================================================

// TestHTTPHandler_API_GET tests GET /api/* with query parameters
func TestHTTPHandler_API_GET(t *testing.T) {
	var invokedRoute, invokedReq string
	dynamic.RegisterPackage("handler-get-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			return "get-response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-get-pkg/v1/users?id=123&name=test", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "get-response" {
		t.Errorf("Body = %q, want 'get-response'", string(body))
	}

	if invokedRoute != "/users" {
		t.Errorf("invokedRoute = %q, want '/users'", invokedRoute)
	}

	// Query params should be in request as JSON
	if !strings.Contains(invokedReq, "id") || !strings.Contains(invokedReq, "123") {
		t.Errorf("invokedReq = %q, expected to contain query params", invokedReq)
	}
}

// TestHTTPHandler_API_POST tests POST /api/* with JSON body
func TestHTTPHandler_API_POST(t *testing.T) {
	var invokedRoute, invokedReq string
	dynamic.RegisterPackage("handler-post-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			return "post-response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	reqBody := `{"action":"create","data":{"name":"test"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/handler-post-pkg/v1/create", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "post-response" {
		t.Errorf("Body = %q, want 'post-response'", string(body))
	}

	if invokedRoute != "/create" {
		t.Errorf("invokedRoute = %q, want '/create'", invokedRoute)
	}

	// POST body should be in request (with __meta__ added)
	if !strings.Contains(invokedReq, "action") || !strings.Contains(invokedReq, "create") {
		t.Errorf("invokedReq = %q, expected to contain POST body", invokedReq)
	}
}

// TestHTTPHandler_API_MultipleRouteSegments tests API with multiple route segments
func TestHTTPHandler_API_MultipleRouteSegments(t *testing.T) {
	var invokedRoute string
	dynamic.RegisterPackage("handler-multi-pkg", "v2", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			return "multi-segment-response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-multi-pkg/v2/users/123/profile/settings", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if invokedRoute != "/users/123/profile/settings" {
		t.Errorf("invokedRoute = %q, want '/users/123/profile/settings'", invokedRoute)
	}
}

// TestHTTPHandler_API_PackageNotFound tests API with non-existent package
func TestHTTPHandler_API_PackageNotFound(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent-handler-pkg/v1/route", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d for non-existent package", resp.StatusCode, http.StatusInternalServerError)
	}
}

// =============================================================================
// Debug Mode Tests
// =============================================================================

// TestHTTPHandler_DebugMode_API tests debug mode for API route
func TestHTTPHandler_DebugMode_API(t *testing.T) {
	dynamic.RegisterPackage("handler-debug-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "debug-api-response"
		},
	})

	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithDebugMode(),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/_/api/handler-debug-pkg/v1/route?key=value", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Debug mode should include various debug fields
	bodyStr := string(body)
	expectedFields := []string{"Path:", "Request:", "Response:"}
	for _, field := range expectedFields {
		if !strings.Contains(bodyStr, field) {
			t.Errorf("Body should contain %q, got %q", field, bodyStr)
		}
	}
}

// TestHTTPHandler_DebugMode_WithError tests debug mode includes error info
func TestHTTPHandler_DebugMode_WithError(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithDebugMode(),
	}, nil)

	// Use non-existent package to trigger error
	req := httptest.NewRequest(http.MethodGet, "/_/api/nonexistent-debug-pkg/v1/route", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d (debug mode returns 200 even with errors)", resp.StatusCode, http.StatusOK)
	}

	// Debug mode should include error info
	if !strings.Contains(string(body), "Error:") {
		t.Errorf("Body should contain 'Error:', got %q", string(body))
	}
}

// =============================================================================
// WAPI Route Tests
// =============================================================================

// TestHTTPHandler_WAPI_ValidResponse tests WAPI with valid HTTP response
func TestHTTPHandler_WAPI_ValidResponse(t *testing.T) {
	dynamic.RegisterPackage("handler-wapi-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			// WAPI expects a full HTTP response format
			return "HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"Content-Length: 15\r\n" +
				"\r\n" +
				`{"status":"ok"}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/wapi/handler-wapi-pkg/v1/status", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != `{"status":"ok"}` {
		t.Errorf("Body = %q, want '{\"status\":\"ok\"}'", string(body))
	}
}

// =============================================================================
// Meta Route Tests
// =============================================================================

// TestHTTPHandler_Meta tests meta route
func TestHTTPHandler_Meta(t *testing.T) {
	dynamic.RegisterPackage("handler-meta-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "meta-response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/meta/handler-meta-pkg/v1/info", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// =============================================================================
// Unmatched Route Tests (404)
// =============================================================================

// TestHTTPHandler_UnmatchedRoute_Returns404 tests that unmatched routes return 404
func TestHTTPHandler_UnmatchedRoute_Returns404(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	testCases := []struct {
		name string
		path string
	}{
		{"random path", "/random/path/here"},
		{"partial api", "/ap/pkg/v1/route"},
		{"typo in api", "/apii/pkg/v1/route"},
		{"special chars", "/path-with-dash"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			e.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("StatusCode = %d, want %d for path %s", resp.StatusCode, http.StatusNotFound, tc.path)
			}
		})
	}
}

// =============================================================================
// Link Mapping Tests
// =============================================================================

// TestHTTPHandler_StaticLink tests static link path mapping
func TestHTTPHandler_StaticLink(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithStaticLink("/status", "/health-check"),
		lambdahttp.WithStaticLink("/ping", "/"),
	}, nil)

	testCases := []struct {
		path     string
		expected string
	}{
		{"/status", "OK"},
		{"/ping", "OK"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			e.ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			if string(body) != tc.expected {
				t.Errorf("Body = %q, want %q", string(body), tc.expected)
			}
		})
	}
}

// TestHTTPHandler_PrefixLink tests prefix link path mapping
func TestHTTPHandler_PrefixLink(t *testing.T) {
	dynamic.RegisterPackage("handler-prefix-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "prefix-mapped"
		},
	})

	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithPrefixLink("/v1", "/api"),
		lambdahttp.WithPrefixLink("/internal", "/api"),
	}, nil)

	testCases := []struct {
		path     string
		expected string
	}{
		{"/v1/handler-prefix-pkg/v1/route", "prefix-mapped"},
		{"/internal/handler-prefix-pkg/v1/route", "prefix-mapped"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			e.ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			if string(body) != tc.expected {
				t.Errorf("Body = %q, want %q", string(body), tc.expected)
			}
		})
	}
}

// TestHTTPHandler_HeaderLink tests header link path mapping
func TestHTTPHandler_HeaderLink(t *testing.T) {
	dynamic.RegisterPackage("handler-header-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "header-mapped"
		},
	})

	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithHeaderLinkKey("X-Api-Route", "/api"),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Api-Route", "handler-header-pkg/v1/route")
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "header-mapped" {
		t.Errorf("Body = %q, want 'header-mapped'", string(body))
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

// TestHTTPHandler_PanicRecovery tests that panics are recovered
func TestHTTPHandler_PanicRecovery(t *testing.T) {
	dynamic.RegisterPackage("handler-panic-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			panic("intentional panic")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-panic-pkg/v1/route", nil)
	w := httptest.NewRecorder()

	// Should not panic
	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d for panic", resp.StatusCode, http.StatusInternalServerError)
	}
}

// =============================================================================
// NotFoundPath Tests
// =============================================================================

// TestHTTPHandler_NotFoundPath_Redirect tests that NotFoundPath redirects to specified path
func TestHTTPHandler_NotFoundPath_Redirect(t *testing.T) {
	dynamic.RegisterPackage("handler-notfound-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "notfound-handler-response"
		},
	})

	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithPageNotFoundPath("/api/handler-notfound-pkg/v1/fallback"),
	}, nil)

	// Request a non-existent path
	req := httptest.NewRequest(http.MethodGet, "/some/random/path", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "notfound-handler-response" {
		t.Errorf("Body = %q, want 'notfound-handler-response'", string(body))
	}
}

// TestHTTPHandler_NotFoundPath_ToHealthCheck tests NotFoundPath redirecting to health-check
func TestHTTPHandler_NotFoundPath_ToHealthCheck(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithPageNotFoundPath("/health-check"),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/path", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK'", string(body))
	}
}

// TestHTTPHandler_NotFoundPath_Empty tests default 404 behavior when NotFoundPath is empty
func TestHTTPHandler_NotFoundPath_Empty(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/path", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	if string(body) != "404 page not found" {
		t.Errorf("Body = %q, want '404 page not found'", string(body))
	}
}

// =============================================================================
// Response Meta Tests
// =============================================================================

// TestHTTPHandler_ResponseMeta_ETag tests that ETag from response __meta__ is set as header
func TestHTTPHandler_ResponseMeta_ETag(t *testing.T) {
	dynamic.RegisterPackage("handler-respmeta-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"data":"test","__meta__":{"etag":"\"abc123\""}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-respmeta-pkg/v1/resource", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Check ETag header is set
	etag := resp.Header.Get("ETag")
	if etag != `"abc123"` {
		t.Errorf("ETag header = %q, want %q", etag, `"abc123"`)
	}

	// Check __meta__ is stripped from response body
	bodyStr := string(body)
	if strings.Contains(bodyStr, "__meta__") {
		t.Errorf("Response body should not contain __meta__, got %q", bodyStr)
	}

	if !strings.Contains(bodyStr, "data") {
		t.Errorf("Response body should contain data field, got %q", bodyStr)
	}
}

// TestHTTPHandler_ResponseMeta_NoMeta tests response without __meta__ works normally
func TestHTTPHandler_ResponseMeta_NoMeta(t *testing.T) {
	dynamic.RegisterPackage("handler-nometa-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"data":"test"}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-nometa-pkg/v1/resource", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// No ETag header should be set
	etag := resp.Header.Get("ETag")
	if etag != "" {
		t.Errorf("ETag header = %q, want empty", etag)
	}

	// Response body should be unchanged
	if string(body) != `{"data":"test"}` {
		t.Errorf("Body = %q, want '{\"data\":\"test\"}'", string(body))
	}
}

// TestHTTPHandler_ResponseMeta_NonJSON tests non-JSON response works normally
func TestHTTPHandler_ResponseMeta_NonJSON(t *testing.T) {
	dynamic.RegisterPackage("handler-nonjson-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "plain text response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-nonjson-pkg/v1/resource", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Response body should be unchanged
	if string(body) != "plain text response" {
		t.Errorf("Body = %q, want 'plain text response'", string(body))
	}
}

// =============================================================================
// Response Meta URL Tests
// =============================================================================

// TestHTTPHandler_ResponseMeta_URL_Redirect tests url meta with http/https scheme triggers 307 redirect
func TestHTTPHandler_ResponseMeta_URL_Redirect(t *testing.T) {
	dynamic.RegisterPackage("handler-url-redirect-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"__meta__":{"url":"https://example.com/target"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-url-redirect-pkg/v1/redirect", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusTemporaryRedirect)
	}

	location := resp.Header.Get("Location")
	if location != "https://example.com/target" {
		t.Errorf("Location = %q, want 'https://example.com/target'", location)
	}
}

// TestHTTPHandler_ResponseMeta_URL_Path tests url meta with path:// scheme triggers internal route rewrite
func TestHTTPHandler_ResponseMeta_URL_Path(t *testing.T) {
	dynamic.RegisterPackage("handler-url-path-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"__meta__":{"url":"path://health-check"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-url-path-pkg/v1/rewrite", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK'", string(body))
	}
}

// TestHTTPHandler_ResponseMeta_URL_Error tests url meta with error:// scheme triggers 500 error
func TestHTTPHandler_ResponseMeta_URL_Error(t *testing.T) {
	dynamic.RegisterPackage("handler-url-error-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"__meta__":{"url":"error://something went wrong"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-url-error-pkg/v1/fail", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	if string(body) != "something went wrong" {
		t.Errorf("Body = %q, want 'something went wrong'", string(body))
	}
}

// TestHTTPHandler_ResponseMeta_Redirect tests redirect meta triggers 307 redirect
func TestHTTPHandler_ResponseMeta_Redirect(t *testing.T) {
	dynamic.RegisterPackage("handler-redirect-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"__meta__":{"redirect":"https://example.com/redirect-target"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-redirect-pkg/v1/redirect", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusTemporaryRedirect)
	}

	location := resp.Header.Get("Location")
	if location != "https://example.com/redirect-target" {
		t.Errorf("Location = %q, want 'https://example.com/redirect-target'", location)
	}
}

// TestHTTPHandler_ResponseMeta_Path tests path meta triggers internal route rewrite
func TestHTTPHandler_ResponseMeta_Path(t *testing.T) {
	dynamic.RegisterPackage("handler-path-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"__meta__":{"path":"health-check"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-path-pkg/v1/rewrite", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK'", string(body))
	}
}

// TestHTTPHandler_ResponseMeta_Error tests error meta triggers 500 error
func TestHTTPHandler_ResponseMeta_Error(t *testing.T) {
	dynamic.RegisterPackage("handler-error-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"__meta__":{"error":"custom error message"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-error-pkg/v1/fail", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	if string(body) != "custom error message" {
		t.Errorf("Body = %q, want 'custom error message'", string(body))
	}
}

// TestHTTPHandler_ResponseMeta_ContentType tests content_type meta overrides Content-Type header
func TestHTTPHandler_ResponseMeta_ContentType(t *testing.T) {
	dynamic.RegisterPackage("handler-contenttype-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"data":"test","__meta__":{"content_type":"text/plain"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-contenttype-pkg/v1/resource", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Content-Type = %q, want 'text/plain'", contentType)
	}
}

// TestHTTPHandler_ResponseMeta_Content tests content meta overrides response body
func TestHTTPHandler_ResponseMeta_Content(t *testing.T) {
	dynamic.RegisterPackage("handler-content-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return `{"data":"original","__meta__":{"content":"overridden content"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-content-pkg/v1/resource", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "overridden content" {
		t.Errorf("Body = %q, want 'overridden content'", string(body))
	}
}

// TestHTTPHandler_ResponseMeta_URL_Priority tests url meta has highest priority
func TestHTTPHandler_ResponseMeta_URL_Priority(t *testing.T) {
	dynamic.RegisterPackage("handler-url-priority-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			// url should take priority over redirect
			return `{"__meta__":{"url":"https://url-target.com","redirect":"https://redirect-target.com"}}`
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/handler-url-priority-pkg/v1/test", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusTemporaryRedirect)
	}

	location := resp.Header.Get("Location")
	if location != "https://url-target.com" {
		t.Errorf("Location = %q, want 'https://url-target.com' (url should have priority)", location)
	}
}
