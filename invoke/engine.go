package invoke

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/aura-studio/lambda/dynamic"
	"google.golang.org/protobuf/proto"
)

// Engine 是 invoke 模块的核心引擎
// Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 8.1, 8.2, 8.3, 8.4
type Engine struct {
	*Options
	*dynamic.Dynamic
	r       *Router
	running atomic.Int32
}

// NewEngine 创建新的引擎实例
// Requirements: 1.1 - 接受 invokeOpts 和 dynamicOpts 两组配置选项
// Requirements: 8.1 - 初始化 Dynamic 组件
func NewEngine(invokeOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(invokeOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
	}
	e.running.Store(1)
	e.InstallHandlers()
	return e
}

// Start 启动引擎
// Requirements: 1.6 - 提供 Start 方法来控制引擎运行状态
func (e *Engine) Start() {
	e.running.Store(1)
}

// Stop 停止引擎
// Requirements: 1.6 - 提供 Stop 方法来控制引擎运行状态
func (e *Engine) Stop() {
	e.running.Store(0)
}

// IsRunning 返回引擎是否正在运行
func (e *Engine) IsRunning() bool {
	return e.running.Load() == 1
}

// Invoke 处理 Lambda 调用请求
// Requirements: 1.2 - 解析请求并路由到对应的处理器
// Requirements: 1.4 - 返回包含响应数据的 Response 结构
// Requirements: 1.5 - 如果请求处理过程中发生错误，在 Response 中包含错误信息
func (e *Engine) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	_ = ctx // reserved for future use

	// 检查引擎是否正在运行
	if e.running.Load() == 0 {
		resp := &Response{
			Error: "engine is stopped",
		}
		return proto.Marshal(resp)
	}

	// 解析请求
	var request Request
	if err := proto.Unmarshal(payload, &request); err != nil {
		if e.DebugMode {
			log.Printf("[Invoke] Unmarshal request error: %v", err)
		}
		resp := &Response{
			Error: fmt.Sprintf("unmarshal request error: %v", err),
		}
		return proto.Marshal(resp)
	}

	// 创建上下文
	c := &Context{
		Engine:  e,
		RawPath: request.Path,
		Path:    request.Path,
		Request: string(request.Payload),
	}

	if e.DebugMode {
		log.Printf("[Invoke] Request: %s %s", c.Path, c.Request)
	}

	// 分发请求到路由器，捕获 panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				c.Err = fmt.Errorf("panic: %v", r)
			}
		}()
		e.r.Dispatch(c)
	}()

	if e.DebugMode {
		log.Printf("[Invoke] Response: %s %s", c.Path, c.Response)
	}

	// 构建响应
	resp := &Response{
		CorrelationId: request.CorrelationId,
		Payload:       []byte(c.Response),
	}
	if c.Err != nil {
		resp.Error = c.Err.Error()
		if e.DebugMode {
			log.Printf("[Invoke] Error: %v", c.Err)
		}
	}

	return proto.Marshal(resp)
}

// handle 解析路径并调用 Dynamic.GetPackage
// Requirements: 8.2 - 从路径中解析包名和版本号
// Requirements: 8.3 - 通过 Dynamic.GetPackage 获取对应的 Tunnel
// Requirements: 8.4 - 如果业务包不存在，返回包含错误信息的响应
func (e *Engine) handle(path string, req string) (string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid path: %q", path)
	}
	pkg := parts[0]
	version := parts[1]

	tunnel, err := e.GetPackage(pkg, version)
	if err != nil {
		return "", err
	}

	route := fmt.Sprintf("/%s", strings.Join(parts[2:], "/"))
	return tunnel.Invoke(route, req), nil
}

