package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/reqresp"
)

// =============================================================================
// reqresp 模块使用示例
// =============================================================================

// ReqRespExampleTunnel 是一个示例业务包实现
type ReqRespExampleTunnel struct {
	name string
}

func (t *ReqRespExampleTunnel) Init() {}

func (t *ReqRespExampleTunnel) Invoke(route string, req string) string {
	var request map[string]interface{}
	json.Unmarshal([]byte(req), &request)

	response := map[string]interface{}{
		"package": t.name,
		"route":   route,
		"message": "Hello from " + t.name,
		"input":   request,
	}

	data, _ := json.Marshal(response)
	return string(data)
}

func (t *ReqRespExampleTunnel) Meta() string {
	return `{"name":"` + t.name + `","version":"v1"}`
}

func (t *ReqRespExampleTunnel) Close() {}

// TestReqRespExample_BasicUsage 展示 reqresp 模块的基本使用
func TestReqRespExample_BasicUsage(t *testing.T) {
	tunnel := &ReqRespExampleTunnel{name: "user-service"}

	dynamic.RegisterPackage("user-service", "v1", tunnel)

	engine := reqresp.NewEngine(
		[]reqresp.Option{
			reqresp.WithDebugMode(false),
		},
		[]dynamicpkg.Option{
			dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
				Package: "user-service",
				Version: "v1",
				Tunnel:  tunnel,
			}),
		},
	)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path:    "/api/user-service/v1/users/123",
		Payload: []byte(`{"action":"get_user","id":123}`),
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if resp.Error != "" {
		t.Errorf("Unexpected error: %s", resp.Error)
	}

	t.Logf("Response: %s", string(resp.Payload))
}

// TestReqRespExample_HealthCheck 展示健康检查
func TestReqRespExample_HealthCheck(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	resp, err := engine.Invoke(context.Background(), &reqresp.Request{
		Path: "/health-check",
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if string(resp.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(resp.Payload))
	}

	t.Log("Health check: OK")
}
