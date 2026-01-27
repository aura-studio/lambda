package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aura-studio/dynamic"
	dynamicpkg "github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/invoke"
	"google.golang.org/protobuf/proto"
)

// =============================================================================
// invoke 模块使用示例
// =============================================================================

// ExampleTunnel 是一个示例业务包实现
type ExampleTunnel struct {
	name string
}

func (t *ExampleTunnel) Init() {}

func (t *ExampleTunnel) Invoke(route string, req string) string {
	// 解析请求
	var request map[string]interface{}
	json.Unmarshal([]byte(req), &request)

	// 构建响应
	response := map[string]interface{}{
		"package": t.name,
		"route":   route,
		"message": "Hello from " + t.name,
		"input":   request,
	}

	data, _ := json.Marshal(response)
	return string(data)
}

func (t *ExampleTunnel) Meta() string {
	return `{"name":"` + t.name + `","version":"v1"}`
}

func (t *ExampleTunnel) Close() {}

// TestInvokeExample_BasicUsage 展示 invoke 模块的基本使用
func TestInvokeExample_BasicUsage(t *testing.T) {
	// 1. 创建业务包
	tunnel := &ExampleTunnel{name: "user-service"}

	// 2. 注册业务包到 dynamic
	dynamic.RegisterPackage("user-service", "v1", tunnel)

	// 3. 创建 invoke 引擎
	engine := invoke.NewEngine(
		// invoke 选项
		[]invoke.Option{
			invoke.WithDebugMode(false),
		},
		// dynamic 选项
		[]dynamicpkg.Option{
			dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
				Package: "user-service",
				Version: "v1",
				Tunnel:  tunnel,
			}),
		},
	)

	// 4. 构建请求
	request := &invoke.Request{
		CorrelationId: "req-001",
		Path:          "/api/user-service/v1/users/123",
		Payload:       []byte(`{"action":"get_user","id":123}`),
	}
	payload, _ := proto.Marshal(request)

	// 5. 调用引擎
	respBytes, err := engine.Invoke(context.Background(), payload)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// 6. 解析响应
	var response invoke.Response
	proto.Unmarshal(respBytes, &response)

	// 7. 验证结果
	if response.CorrelationId != "req-001" {
		t.Errorf("CorrelationId = %q, want 'req-001'", response.CorrelationId)
	}
	if response.Error != "" {
		t.Errorf("Unexpected error: %s", response.Error)
	}

	// 解析业务响应
	var bizResp map[string]interface{}
	json.Unmarshal(response.Payload, &bizResp)

	t.Logf("Response: %s", string(response.Payload))

	if bizResp["package"] != "user-service" {
		t.Errorf("package = %v, want 'user-service'", bizResp["package"])
	}
	if bizResp["route"] != "/users/123" {
		t.Errorf("route = %v, want '/users/123'", bizResp["route"])
	}
}

// TestInvokeExample_WithStaticLink 展示静态路径映射
func TestInvokeExample_WithStaticLink(t *testing.T) {
	engine := invoke.NewEngine(
		[]invoke.Option{
			// 将 /ping 映射到 /health-check
			invoke.WithStaticLink("/ping", "/health-check"),
			// 将 /status 也映射到 /health-check
			invoke.WithStaticLink("/status", "/health-check"),
		},
		nil,
	)

	// 请求 /ping，实际会路由到 /health-check
	request := &invoke.Request{
		CorrelationId: "ping-001",
		Path:          "/ping",
	}
	payload, _ := proto.Marshal(request)

	respBytes, _ := engine.Invoke(context.Background(), payload)

	var response invoke.Response
	proto.Unmarshal(respBytes, &response)

	if string(response.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(response.Payload))
	}

	t.Logf("/ping -> /health-check -> OK")
}

// TestInvokeExample_WithPrefixLink 展示前缀路径映射
func TestInvokeExample_WithPrefixLink(t *testing.T) {
	tunnel := &ExampleTunnel{name: "order-service"}
	dynamic.RegisterPackage("order-service", "v1", tunnel)

	engine := invoke.NewEngine(
		[]invoke.Option{
			// 将 /v1/* 映射到 /api/*
			invoke.WithPrefixLink("/v1", "/api"),
		},
		[]dynamicpkg.Option{
			dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
				Package: "order-service",
				Version: "v1",
				Tunnel:  tunnel,
			}),
		},
	)

	// 请求 /v1/order-service/v1/orders，实际会路由到 /api/order-service/v1/orders
	request := &invoke.Request{
		CorrelationId: "order-001",
		Path:          "/v1/order-service/v1/orders",
		Payload:       []byte(`{"action":"list"}`),
	}
	payload, _ := proto.Marshal(request)

	respBytes, _ := engine.Invoke(context.Background(), payload)

	var response invoke.Response
	proto.Unmarshal(respBytes, &response)

	if response.Error != "" {
		t.Errorf("Unexpected error: %s", response.Error)
	}

	t.Logf("/v1/order-service/v1/orders -> /api/order-service/v1/orders")
	t.Logf("Response: %s", string(response.Payload))
}

// TestInvokeExample_DebugMode 展示调试模式
func TestInvokeExample_DebugMode(t *testing.T) {
	tunnel := &ExampleTunnel{name: "debug-service"}
	dynamic.RegisterPackage("debug-service", "v1", tunnel)

	engine := invoke.NewEngine(
		[]invoke.Option{
			invoke.WithDebugMode(true),
		},
		[]dynamicpkg.Option{
			dynamicpkg.WithStaticPackage(&dynamicpkg.Package{
				Package: "debug-service",
				Version: "v1",
				Tunnel:  tunnel,
			}),
		},
	)

	// 使用 /_/api/* 路径触发调试模式
	request := &invoke.Request{
		CorrelationId: "debug-001",
		Path:          "/_/api/debug-service/v1/test",
		Payload:       []byte(`{"debug":true}`),
	}
	payload, _ := proto.Marshal(request)

	respBytes, _ := engine.Invoke(context.Background(), payload)

	var response invoke.Response
	proto.Unmarshal(respBytes, &response)

	// 调试模式返回 JSON 格式的调试信息
	var debugInfo map[string]interface{}
	if err := json.Unmarshal(response.Payload, &debugInfo); err != nil {
		t.Fatalf("Debug response is not JSON: %v", err)
	}

	t.Logf("Debug response fields:")
	t.Logf("  mode: %v", debugInfo["mode"])
	t.Logf("  raw_path: %v", debugInfo["raw_path"])
	t.Logf("  path: %v", debugInfo["path"])
	t.Logf("  param: %v", debugInfo["param"])
	t.Logf("  request: %v", debugInfo["request"])
	t.Logf("  response: %v", debugInfo["response"])

	if debugInfo["mode"] != "api" {
		t.Errorf("mode = %v, want 'api'", debugInfo["mode"])
	}
}

// TestInvokeExample_EngineLifecycle 展示引擎生命周期控制
func TestInvokeExample_EngineLifecycle(t *testing.T) {
	engine := invoke.NewEngine(nil, nil)

	// 引擎创建后默认是运行状态
	if !engine.IsRunning() {
		t.Error("Engine should be running after creation")
	}

	// 停止引擎
	engine.Stop()
	if engine.IsRunning() {
		t.Error("Engine should be stopped after Stop()")
	}

	// 停止后的请求会返回错误
	request := &invoke.Request{Path: "/health-check"}
	payload, _ := proto.Marshal(request)
	respBytes, _ := engine.Invoke(context.Background(), payload)

	var response invoke.Response
	proto.Unmarshal(respBytes, &response)

	if response.Error != "engine is stopped" {
		t.Errorf("Error = %q, want 'engine is stopped'", response.Error)
	}

	// 重新启动引擎
	engine.Start()
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start()")
	}

	// 启动后可以正常处理请求
	respBytes, _ = engine.Invoke(context.Background(), payload)
	proto.Unmarshal(respBytes, &response)

	if response.Error != "" {
		t.Errorf("Unexpected error after Start(): %s", response.Error)
	}
	if string(response.Payload) != "OK" {
		t.Errorf("Payload = %q, want 'OK'", string(response.Payload))
	}

	t.Log("Engine lifecycle: Create -> Stop -> Start -> OK")
}

// TestInvokeExample_ErrorHandling 展示错误处理
func TestInvokeExample_ErrorHandling(t *testing.T) {
	engine := invoke.NewEngine(nil, nil)

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"健康检查", "/health-check", false},
		{"根路径", "/", false},
		{"不存在的路由", "/unknown/path", true},
		{"不存在的包", "/api/nonexistent/v1/route", true},
		{"无效的 API 路径", "/api/onlypkg", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &invoke.Request{
				CorrelationId: "test",
				Path:          tc.path,
				Payload:       []byte("{}"),
			}
			payload, _ := proto.Marshal(request)

			respBytes, _ := engine.Invoke(context.Background(), payload)

			var response invoke.Response
			proto.Unmarshal(respBytes, &response)

			hasError := response.Error != ""
			if hasError != tc.expectError {
				t.Errorf("path=%q: hasError=%v, want %v (error=%q)",
					tc.path, hasError, tc.expectError, response.Error)
			}
		})
	}
}
