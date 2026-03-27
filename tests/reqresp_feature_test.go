package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/reqresp"
)

// =============================================================================
// ReqResp: Envelope protocol
// =============================================================================

func TestReqResp_Envelope_TunnelReceivesEnvelope(t *testing.T) {
	var receivedReq string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			receivedReq = req
			return encodeEnvelope(nil, "ok")
		},
	}
	dynamic.RegisterPackage("rr-env-recv", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-env-recv", Version: "v1", Tunnel: tunnel}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/rr-env-recv/v1/test",
		Payload: []byte("hello-world"),
	})
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("resp.Error = %q", resp.Error)
	}

	env := decodeEnvelope(t, receivedReq)
	data, _ := base64.StdEncoding.DecodeString(env.Data)
	if string(data) != "hello-world" {
		t.Errorf("data = %q, want 'hello-world'", string(data))
	}
}

func TestReqResp_Envelope_ResponseDecoded(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(nil, `{"result":"success"}`)
		},
	}
	dynamic.RegisterPackage("rr-env-rsp", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-env-rsp", Version: "v1", Tunnel: tunnel}),
	})

	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/api/rr-env-rsp/v1/test", Payload: []byte("{}"),
	})
	if string(resp.Payload) != `{"result":"success"}` {
		t.Errorf("Payload = %q, want '{\"result\":\"success\"}'", string(resp.Payload))
	}
}

func TestReqResp_Envelope_RspMetaError(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(map[string]any{"Error": "bad request"}, "")
		},
	}
	dynamic.RegisterPackage("rr-env-err", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-env-err", Version: "v1", Tunnel: tunnel}),
	})

	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/api/rr-env-err/v1/test", Payload: []byte("{}"),
	})
	if resp.Error == "" || !strings.Contains(resp.Error, "bad request") {
		t.Errorf("Error = %q, want to contain 'bad request'", resp.Error)
	}
}

// =============================================================================
// ReqResp: Routing
// =============================================================================

func TestReqResp_Route_HealthCheck(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{Path: "/health-check"})
	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}
}

func TestReqResp_Route_Root(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{Path: "/"})
	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}
}

func TestReqResp_Route_NotFound(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{Path: "/nonexistent"})
	if resp.Error == "" || !strings.Contains(resp.Error, "404") {
		t.Errorf("Error = %q, want to contain '404'", resp.Error)
	}
}

func TestReqResp_Route_APIMultiSegment(t *testing.T) {
	var invokedRoute string
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			invokedRoute = route
			return encodeEnvelope(nil, "ok")
		},
	}
	dynamic.RegisterPackage("rr-multi", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-multi", Version: "v1", Tunnel: tunnel}),
	})

	engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/api/rr-multi/v1/users/123/profile", Payload: []byte("{}"),
	})
	if invokedRoute != "/users/123/profile" {
		t.Errorf("route = %q, want '/users/123/profile'", invokedRoute)
	}
}

func TestReqResp_Route_PackageNotFound(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/api/nonexistent-pkg/v1/route", Payload: []byte("{}"),
	})
	if resp.Error == "" {
		t.Error("expected error for nonexistent package")
	}
}

// =============================================================================
// ReqResp: Panic recovery
// =============================================================================

func TestReqResp_PanicRecovery(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			panic("boom")
		},
	}
	dynamic.RegisterPackage("rr-panic", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-panic", Version: "v1", Tunnel: tunnel}),
	})

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/api/rr-panic/v1/test", Payload: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("Invoke should not return Go error, got: %v", err)
	}
	if resp.Error == "" || !strings.Contains(resp.Error, "boom") {
		t.Errorf("Error = %q, want to contain 'boom'", resp.Error)
	}
}

// =============================================================================
// ReqResp: Debug mode
// =============================================================================

func TestReqResp_DebugMode(t *testing.T) {
	tunnel := &mockReqRespTunnel{
		invokeFunc: func(route, req string) string {
			return encodeEnvelope(nil, "debug-data")
		},
	}
	dynamic.RegisterPackage("rr-debug", "v1", tunnel)

	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-debug", Version: "v1", Tunnel: tunnel}),
	})

	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/_/api/rr-debug/v1/test", Payload: []byte("test-input"),
	})

	body := string(resp.Payload)
	for _, field := range []string{"Path:", "Request:", "Response:"} {
		if !strings.Contains(body, field) {
			t.Errorf("debug output missing %q, got: %s", field, body)
		}
	}
}

func TestReqResp_DebugMode_WithError(t *testing.T) {
	engine := reqresp.NewEngine([]reqresp.Option{
		reqresp.WithDebugMode(true),
	}, nil)

	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/_/api/nonexistent/v1/route", Payload: []byte("{}"),
	})

	body := string(resp.Payload)
	if !strings.Contains(body, "Error:") {
		t.Errorf("debug output missing 'Error:', got: %s", body)
	}
}

// =============================================================================
// ReqResp: Meta endpoint
// =============================================================================

func TestReqResp_Meta(t *testing.T) {
	tunnel := &mockReqRespTunnel{}
	dynamic.RegisterPackage("rr-meta", "v1", tunnel)

	engine := reqresp.NewEngine(nil, []dynamicpkg.Option{
		dynamicpkg.WithStaticPackage(&dynamicpkg.Package{Package: "rr-meta", Version: "v1", Tunnel: tunnel}),
	})

	resp, _ := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/meta/rr-meta/v1",
	})
	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
	if len(resp.Payload) == 0 {
		t.Error("expected non-empty meta response")
	}
}

func TestReqResp_Meta_PanicRecovery(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)
	// meta with nonexistent package should not panic
	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/meta/nonexistent/v1",
	})
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	// should succeed (meta generates info even without tunnel)
	if resp.Error != "" {
		t.Errorf("Error = %q, want empty", resp.Error)
	}
}

// =============================================================================
// ReqResp: Context Set/Get
// =============================================================================

func TestReqResp_Context_SetGet(t *testing.T) {
	c := &reqresp.Context{}
	c.Set("key", "value")
	v, ok := c.Get("key")
	if !ok || v != "value" {
		t.Errorf("Get('key') = %v, %v, want 'value', true", v, ok)
	}
}

func TestReqResp_Context_GetString(t *testing.T) {
	c := &reqresp.Context{}
	c.Set("k", "hello")
	if c.GetString("k") != "hello" {
		t.Errorf("GetString = %q, want 'hello'", c.GetString("k"))
	}
	if c.GetString("missing") != "" {
		t.Errorf("GetString(missing) = %q, want ''", c.GetString("missing"))
	}
}

func TestReqResp_Context_GetBool(t *testing.T) {
	c := &reqresp.Context{}
	c.Set("flag", true)
	if !c.GetBool("flag") {
		t.Error("GetBool = false, want true")
	}
	if c.GetBool("missing") {
		t.Error("GetBool(missing) = true, want false")
	}
}

func TestReqResp_Context_GetStringMap(t *testing.T) {
	c := &reqresp.Context{}
	m := map[string]any{"a": "b"}
	c.Set("meta", m)
	got := c.GetStringMap("meta")
	if got["a"] != "b" {
		t.Errorf("GetStringMap = %v, want map with a=b", got)
	}
	if c.GetStringMap("missing") != nil {
		t.Error("GetStringMap(missing) should be nil")
	}
}

func TestReqResp_Context_GetError(t *testing.T) {
	c := &reqresp.Context{}
	if c.GetError() != nil {
		t.Error("GetError should be nil initially")
	}
	c.Set(reqresp.ContextError, fmt.Errorf("test error"))
	if c.GetError() == nil || c.GetError().Error() != "test error" {
		t.Errorf("GetError = %v, want 'test error'", c.GetError())
	}
}
