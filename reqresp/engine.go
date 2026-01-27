package reqresp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"github.com/aura-studio/lambda/dynamic"
	"google.golang.org/protobuf/proto"
)

// Engine 是 reqresp 模块的核心引擎
type Engine struct {
	*Options
	*dynamic.Dynamic
	r       *Router
	running atomic.Int32
}

// NewEngine 创建新的引擎实例
func NewEngine(reqrespOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(reqrespOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
	}
	e.running.Store(1)
	e.InstallHandlers()
	return e
}

// Start 启动引擎
func (e *Engine) Start() {
	e.running.Store(1)
}

// Stop 停止引擎
func (e *Engine) Stop() {
	e.running.Store(0)
}

// IsRunning 返回引擎是否正在运行
func (e *Engine) IsRunning() bool {
	return e.running.Load() == 1
}

// Invoke 处理 Lambda 调用请求
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
			log.Printf("[ReqResp] Unmarshal request error: %v", err)
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
		log.Printf("[ReqResp] Request: %s %s", c.Path, c.Request)
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
		log.Printf("[ReqResp] Response: %s %s", c.Path, c.Response)
	}

	// 构建响应
	resp := &Response{
		CorrelationId: request.CorrelationId,
		Payload:       []byte(c.Response),
	}
	if c.Err != nil {
		resp.Error = c.Err.Error()
		if e.DebugMode {
			log.Printf("[ReqResp] Error: %v", c.Err)
		}
	}

	return proto.Marshal(resp)
}

// handle 解析路径并调用 Dynamic.GetPackage
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
