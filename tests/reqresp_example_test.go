package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/reqresp"
	"google.golang.org/protobuf/proto"
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
	// 1. 创建业务包
	tunnel := &ReqRespExampleTunnel{name: "user-service"}

	// 2. 注册业务包到 dynamic
	dynamic.RegisterPackage("user-service", "v1", tunnel)

	// 3. 创建 reqresp 引擎
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

	// 4. 构建请求
	request := &reqresp.Request{
		Path:    "/api/user-service/v1/users/123",
		Payload: []byte(`{"action":"get_user","id":123}`),
	}
	payload, _ := proto.Marshal(request)

	// 5. 调用引擎
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// 6. 解析响应
	var response reqresp.Response
	proto.Unmarshal(respBytes, &response)

	// 7. 验证结果
	if response.Error != "" {
		t.Errorf("Unexpected error: %s", response.Error)
	}

	t.Logf("Response: %s", string(response.Payload))
}

// TestReqRespExample_HealthCheck 展示健康检查
func TestReqRespExample_HealthCheck(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	request := &reqresp.Request{
		Path: "/health-check",
	}
	payload, _ := proto.Marshal(request)

	respBytes, _ := engine.Invoke(context.Background(), payload)

	var response reqresp.Response
	proto.Unmarshal(respBytes, &response)

	if string(response.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(response.Payload))
	}

	t.Log("Health check: OK")
}

// TestReqRespExample_EngineLifecycle 展示引擎生命周期控制
func TestReqRespExample_EngineLifecycle(t *testing.T) {
	engine := reqresp.NewEngine(nil, nil)

	if !engine.IsRunning() {
		t.Error("Engine should be running after creation")
	}

	engine.Stop()
	if engine.IsRunning() {
		t.Error("Engine should be stopped after Stop()")
	}

	engine.Start()
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start()")
	}

	t.Log("Engine lifecycle: Create -> Stop -> Start -> OK")
}
