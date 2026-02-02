package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aura-studio/dynamic"
	lambdahttp "github.com/aura-studio/lambda/http"
)

// mockHTTPTunnel implements dynamic.Tunnel for testing
type mockHTTPTunnel struct {
	invokeFunc func(route, req string) string
}

func (m *mockHTTPTunnel) Init() {}

func (m *mockHTTPTunnel) Invoke(route string, req string) string {
	if m.invokeFunc != nil {
		return m.invokeFunc(route, req)
	}
	return "mock-response"
}

func (m *mockHTTPTunnel) Meta() string {
	return "mock-meta"
}

func (m *mockHTTPTunnel) Close() {}

// TestHTTPEngineCreation tests that NewEngine creates an engine with correct options
func TestHTTPEngineCreation(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithAddress(":9090"),
		lambdahttp.WithDebugMode(),
		lambdahttp.WithCorsMode(),
		lambdahttp.WithStaticLink("/static", "/public"),
		lambdahttp.WithPrefixLink("/api", "/v1"),
		lambdahttp.WithHeaderLinkKey("X-Route", "/route"),
	}, nil)

	if e == nil {
		t.Fatal("NewEngine returned nil")
	}

	if !e.DebugMode {
		t.Error("DebugMode should be true")
	}

	if !e.CorsMode {
		t.Error("CorsMode should be true")
	}

	if e.Address != ":9090" {
		t.Errorf("Address = %q, want ':9090'", e.Address)
	}

	if e.StaticLinkMap["/static"] != "/public" {
		t.Errorf("StaticLinkMap['/static'] = %q, want '/public'", e.StaticLinkMap["/static"])
	}

	if e.PrefixLinkMap["/api"] != "/v1" {
		t.Errorf("PrefixLinkMap['/api'] = %q, want '/v1'", e.PrefixLinkMap["/api"])
	}

	if e.HeaderLinkMap["X-Route"] != "/route" {
		t.Errorf("HeaderLinkMap['X-Route'] = %q, want '/route'", e.HeaderLinkMap["X-Route"])
	}
}

// TestHTTPEngineHealthCheck tests the health check endpoint
func TestHTTPEngineHealthCheck(t *testing.T) {
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

// TestHTTPEngineRootPath tests the root path endpoint
func TestHTTPEngineRootPath(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
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

// TestHTTPEnginePageNotFound tests the 404 handler
func TestHTTPEnginePageNotFound(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent/path", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}

	if !strings.Contains(string(body), "404") {
		t.Errorf("Body = %q, expected to contain '404'", string(body))
	}
}

// TestHTTPEngineStaticLink tests static link path mapping
func TestHTTPEngineStaticLink(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithStaticLink("/custom-health", "/health-check"),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/custom-health", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK' (static link should map to health-check)", string(body))
	}
}

// TestHTTPEnginePrefixLink tests prefix link path mapping
func TestHTTPEnginePrefixLink(t *testing.T) {
	dynamic.RegisterPackage("httpprefixpkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "prefix-response"
		},
	})

	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithPrefixLink("/v1", "/api"),
	}, nil)

	// /v1/httpprefixpkg/v1/route should be mapped to /api/httpprefixpkg/v1/route
	req := httptest.NewRequest(http.MethodGet, "/v1/httpprefixpkg/v1/route", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "prefix-response" {
		t.Errorf("Body = %q, want 'prefix-response'", string(body))
	}
}

// TestHTTPEngineHeaderLink tests header link path mapping
func TestHTTPEngineHeaderLink(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithHeaderLinkKey("X-Route", "/"),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Route", "health-check")
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "OK" {
		t.Errorf("Body = %q, want 'OK' (header link should map to health-check)", string(body))
	}
}

// TestHTTPEngineAPIWithDynamic tests API route calling Dynamic
func TestHTTPEngineAPIWithDynamic(t *testing.T) {
	var invokedRoute, invokedReq string
	dynamic.RegisterPackage("httppkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			return "http-api-response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/httppkg/v1/myroute?key=value", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "http-api-response" {
		t.Errorf("Body = %q, want 'http-api-response'", string(body))
	}

	if invokedRoute != "/myroute" {
		t.Errorf("invokedRoute = %q, want '/myroute'", invokedRoute)
	}

	// GET request should have query params as JSON
	if !strings.Contains(invokedReq, "key") {
		t.Errorf("invokedReq = %q, expected to contain 'key'", invokedReq)
	}
}

// TestHTTPEngineAPIPost tests API route with POST method
func TestHTTPEngineAPIPost(t *testing.T) {
	var invokedReq string
	dynamic.RegisterPackage("httppostpkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			invokedReq = req
			return "post-response"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	reqBody := `{"action":"create","data":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/httppostpkg/v1/create", strings.NewReader(reqBody))
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

	// POST request should have body as request
	if !strings.Contains(invokedReq, "action") {
		t.Errorf("invokedReq = %q, expected to contain 'action'", invokedReq)
	}
}

// TestHTTPEngineWAPI tests WAPI route
func TestHTTPEngineWAPI(t *testing.T) {
	dynamic.RegisterPackage("httpwapipkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			// Return a valid HTTP response
			return "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 12\r\n\r\nwapi-content"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/wapi/httpwapipkg/v1/route", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if string(body) != "wapi-content" {
		t.Errorf("Body = %q, want 'wapi-content'", string(body))
	}
}

// TestHTTPEngineCORS tests CORS middleware
func TestHTTPEngineCORS(t *testing.T) {
	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithCorsMode(),
	}, nil)

	req := httptest.NewRequest(http.MethodOptions, "/health-check", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d for OPTIONS", resp.StatusCode, http.StatusNoContent)
	}

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want '*'", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

// TestHTTPEngineAllMethods tests that all HTTP methods work
func TestHTTPEngineAllMethods(t *testing.T) {
	e := lambdahttp.NewEngine(nil, nil)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health-check", nil)
			w := httptest.NewRecorder()

			e.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d for method %s", resp.StatusCode, http.StatusOK, method)
			}
		})
	}
}

// TestHTTPEngineDebugMode tests debug mode response
func TestHTTPEngineDebugMode(t *testing.T) {
	dynamic.RegisterPackage("httpdebugpkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return "debug-response"
		},
	})

	e := lambdahttp.NewEngine([]lambdahttp.Option{
		lambdahttp.WithDebugMode(),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/_/api/httpdebugpkg/v1/route", nil)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Debug mode should include debug info
	if !strings.Contains(string(body), "Path:") {
		t.Errorf("Body should contain debug info, got %q", string(body))
	}
	if !strings.Contains(string(body), "Response:") {
		t.Errorf("Body should contain Response field, got %q", string(body))
	}
}

// TestHTTPEngineServeAndClose tests the Serve and Close functions
func TestHTTPEngineServeAndClose(t *testing.T) {
	addr := freeLocalAddr(t)

	errCh := make(chan error, 1)
	go func() {
		errCh <- lambdahttp.Serve([]lambdahttp.Option{lambdahttp.WithAddress(addr)}, nil)
	}()

	waitHTTPReady(t, "http://"+addr, 3*time.Second)

	// Make a request to verify server is running
	resp, err := http.Get("http://" + addr + "/health-check")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if err := lambdahttp.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve() returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Serve() did not return after Close()")
	}
}
