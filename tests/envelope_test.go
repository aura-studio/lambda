package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	lambdahttp "github.com/aura-studio/lambda/http"
	"github.com/aura-studio/lambda/reqresp"
)

// envelope is the shared request/response format: {"meta":{...}, "data":"base64"}
type envelope struct {
	Meta map[string]any `json:"meta"`
	Data string         `json:"data"`
}

func encodeEnvelope(meta map[string]any, data string) string {
	env := envelope{
		Meta: meta,
		Data: base64.StdEncoding.EncodeToString([]byte(data)),
	}
	b, _ := json.Marshal(env)
	return string(b)
}

func decodeEnvelope(t *testing.T, raw string) envelope {
	t.Helper()
	var env envelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("failed to decode envelope: %v", err)
	}
	return env
}

func decodeEnvelopeData(t *testing.T, raw string) string {
	t.Helper()
	env := decodeEnvelope(t, raw)
	data, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		t.Fatalf("failed to decode base64 data: %v", err)
	}
	return string(data)
}

// =============================================================================
// HTTP: doProcessor envelope wrap/unwrap
// =============================================================================

// TestHTTP_Envelope_RequestFormat verifies tunnel receives envelope with meta and base64 data
func TestHTTP_Envelope_RequestFormat(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("env-req-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, `{"result":"ok"}`)
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/env-req-pkg/v1/test",
		strings.NewReader(`{"action":"create"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	// Verify tunnel received envelope format
	env := decodeEnvelope(t, receivedReq)
	if env.Meta == nil {
		t.Error("envelope meta should not be nil")
	}
	data, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		t.Fatalf("failed to decode data: %v", err)
	}
	if string(data) != `{"action":"create"}` {
		t.Errorf("data = %q, want '{\"action\":\"create\"}'", string(data))
	}
}

// TestHTTP_Envelope_ResponseDecode verifies response data is base64-decoded
func TestHTTP_Envelope_ResponseDecode(t *testing.T) {
	dynamic.RegisterPackage("env-rsp-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(nil, `{"status":"success"}`)
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-rsp-pkg/v1/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if string(body) != `{"status":"success"}` {
		t.Errorf("Body = %q, want '{\"status\":\"success\"}'", string(body))
	}
}

// =============================================================================
// HTTP: reqMeta in envelope
// =============================================================================

// TestHTTP_Envelope_ReqMeta_Host verifies host resolution via X-Forwarded-Host
func TestHTTP_Envelope_ReqMeta_Host(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("env-meta-host-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "ok")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-meta-host-pkg/v1/test", nil)
	req.Header.Set("X-Forwarded-Host", "example.com:8080, proxy.com")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	env := decodeEnvelope(t, receivedReq)
	host, _ := env.Meta["Host"].(string)
	if host != "example.com" {
		t.Errorf("meta.host = %q, want 'example.com'", host)
	}
}

// TestHTTP_Envelope_ReqMeta_RemoteAddr verifies remote addr resolution via CloudFront header
func TestHTTP_Envelope_ReqMeta_RemoteAddr(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("env-meta-addr-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "ok")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-meta-addr-pkg/v1/test", nil)
	req.Header.Set("CloudFront-Viewer-Address", "1.2.3.4, 5.6.7.8")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	env := decodeEnvelope(t, receivedReq)
	addr, _ := env.Meta["RemoteAddr"].(string)
	if addr != "1.2.3.4" {
		t.Errorf("meta.remote_addr = %q, want '1.2.3.4'", addr)
	}
}

// TestHTTP_Envelope_ReqMeta_RemoteAddr_XForwardedFor verifies fallback to X-Forwarded-For
func TestHTTP_Envelope_ReqMeta_RemoteAddr_XForwardedFor(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("env-meta-xff-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "ok")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-meta-xff-pkg/v1/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	env := decodeEnvelope(t, receivedReq)
	addr, _ := env.Meta["RemoteAddr"].(string)
	if addr != "10.0.0.1" {
		t.Errorf("meta.remote_addr = %q, want '10.0.0.1'", addr)
	}
}

// TestHTTP_Envelope_ReqMeta_Path verifies path is included in meta
func TestHTTP_Envelope_ReqMeta_Path(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("env-meta-path-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "ok")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-meta-path-pkg/v1/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	env := decodeEnvelope(t, receivedReq)
	path, _ := env.Meta["Path"].(string)
	if path != "/api/env-meta-path-pkg/v1/test" {
		t.Errorf("meta.path = %q, want '/api/env-meta-path-pkg/v1/test'", path)
	}
}

// =============================================================================
// HTTP: rspMeta handling
// =============================================================================

// TestHTTP_Envelope_RspMeta_Error verifies rspMeta error triggers 500
func TestHTTP_Envelope_RspMeta_Error(t *testing.T) {
	dynamic.RegisterPackage("env-rsperr-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(map[string]any{"Error": "something went wrong"}, "")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-rsperr-pkg/v1/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// TestHTTP_Envelope_RspMeta_ContentType verifies rspMeta content_type is applied
func TestHTTP_Envelope_RspMeta_ContentType(t *testing.T) {
	dynamic.RegisterPackage("env-ct-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(map[string]any{"ContentType": "text/html"}, "<h1>hello</h1>")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-ct-pkg/v1/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/html" {
		t.Errorf("Content-Type = %q, want 'text/html'", ct)
	}
	if string(body) != "<h1>hello</h1>" {
		t.Errorf("Body = %q, want '<h1>hello</h1>'", string(body))
	}
}

// TestHTTP_Envelope_RspMeta_Status verifies rspMeta status overrides HTTP status code
func TestHTTP_Envelope_RspMeta_Status(t *testing.T) {
	dynamic.RegisterPackage("env-status-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(map[string]any{"Status": 201}, `{"id":"new"}`)
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/env-status-pkg/v1/create",
		strings.NewReader(`{"name":"test"}`))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	if string(body) != `{"id":"new"}` {
		t.Errorf("Body = %q, want '{\"id\":\"new\"}'", string(body))
	}
}

// TestHTTP_Envelope_RspMeta_StatusDefault verifies default status is 200
func TestHTTP_Envelope_RspMeta_StatusDefault(t *testing.T) {
	dynamic.RegisterPackage("env-status-default-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(nil, "ok")
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/env-status-default-pkg/v1/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// =============================================================================
// HTTP: WAPI does NOT use envelope
// =============================================================================

// TestHTTP_WAPI_NoEnvelope verifies WAPI passes raw request, no envelope wrapping
func TestHTTP_WAPI_NoEnvelope(t *testing.T) {
	var receivedReq string
	dynamic.RegisterPackage("env-wapi-pkg", "v1", &mockHTTPTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 4\r\n\r\nwapi"
		},
	})

	e := lambdahttp.NewEngine(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/wapi/env-wapi-pkg/v1/test", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if string(body) != "wapi" {
		t.Errorf("Body = %q, want 'wapi'", string(body))
	}

	// WAPI should NOT wrap in envelope — receivedReq should be raw HTTP, not JSON envelope
	var env envelope
	if err := json.Unmarshal([]byte(receivedReq), &env); err == nil && env.Data != "" {
		t.Error("WAPI should not wrap request in envelope format")
	}
}

// =============================================================================
// ReqResp: process envelope wrap/unwrap
// =============================================================================

// TestReqResp_Envelope_RequestFormat verifies tunnel receives envelope format
func TestReqResp_Envelope_RequestFormat(t *testing.T) {
	var receivedReq string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "response-data")
		},
	}
	dynamic.RegisterPackage("env-rr-req-pkg", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "env-rr-req-pkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/env-rr-req-pkg/v1/test",
		Payload: []byte("hello"),
	})
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}

	// Verify tunnel received envelope
	env := decodeEnvelope(t, receivedReq)
	data, _ := base64.StdEncoding.DecodeString(env.Data)
	if string(data) != "hello" {
		t.Errorf("data = %q, want 'hello'", string(data))
	}

	// Verify response is decoded
	if string(resp.Payload) != "response-data" {
		t.Errorf("Payload = %q, want 'response-data'", string(resp.Payload))
	}
}

// TestReqResp_Envelope_RspMeta_Error verifies error in rspMeta is returned as error
func TestReqResp_Envelope_RspMeta_Error(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(map[string]any{"Error": "bad request"}, "")
		},
	}
	dynamic.RegisterPackage("env-rr-err-pkg", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
			Package: "env-rr-err-pkg",
			Version: "v1",
			Tunnel:  tunnel,
		}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/env-rr-err-pkg/v1/test",
		Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}

	if resp.Error == "" {
		t.Error("expected error from rspMeta")
	}
	if !strings.Contains(resp.Error, "bad request") {
		t.Errorf("Error = %q, want to contain 'bad request'", resp.Error)
	}
}
