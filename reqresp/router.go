package reqresp

import (
	"fmt"
	"strings"
)

// HandlerFunc 处理器函数类型
type HandlerFunc func(*Context)

// Context 请求上下文
type Context struct {
	Engine *Engine

	// RawPath is the path as provided by the request (before link rewriting).
	RawPath string
	// Path is the current effective path (after link rewriting).
	Path string
	// ParamPath is the wildcard parameter value for routes like /api/*path.
	// It includes a leading slash, e.g. /pkg/commit/route.
	ParamPath string

	Request  string
	Response string
	Err      error

	DebugMode bool

	aborted bool
}

// Abort 中止当前请求处理链
func (c *Context) Abort() { c.aborted = true }

// route 路由定义
type route struct {
	pattern  string
	handlers []HandlerFunc
}

// Router 路由器（导出用于测试）
type Router struct {
	pre []HandlerFunc // 前置中间件

	routes   []route       // 路由表
	noRoute  []HandlerFunc // 未匹配路由处理器
	noMethod []HandlerFunc // 方法不允许处理器
}

// NewRouter 创建新的路由器实例（导出用于测试）
func NewRouter() *Router {
	return &Router{}
}

// Use 注册前置中间件
func (r *Router) Use(handlers ...HandlerFunc) {
	r.pre = append(r.pre, handlers...)
}

// Handle 注册路由处理器
func (r *Router) Handle(pattern string, handlers ...HandlerFunc) {
	r.routes = append(r.routes, route{pattern: pattern, handlers: handlers})
}

// NoRoute 设置未匹配路由处理器
func (r *Router) NoRoute(handlers ...HandlerFunc) { r.noRoute = handlers }

// NoMethod 设置方法不允许处理器
func (r *Router) NoMethod(handlers ...HandlerFunc) { r.noMethod = handlers }

// Dispatch 分发请求到匹配的处理器（导出用于测试）
func (r *Router) Dispatch(ctx *Context) {
	// 执行前置中间件
	for _, h := range r.pre {
		if h == nil {
			continue
		}
		h(ctx)
		if ctx.aborted {
			return
		}
	}

	// 匹配路由
	matched, handlers := r.match(ctx.Path)
	if !matched {
		handlers = r.noRoute
	}
	if len(handlers) == 0 {
		ctx.Err = fmt.Errorf("no route for path: %q", ctx.Path)
		return
	}

	// 执行路由处理器
	for _, h := range handlers {
		if h == nil {
			continue
		}
		h(ctx)
		if ctx.aborted {
			return
		}
		if ctx.Err != nil {
			return
		}
	}
}

// match 匹配路径到路由
func (r *Router) match(path string) (bool, []HandlerFunc) {
	for _, rt := range r.routes {
		param, ok := MatchPattern(rt.pattern, path)
		if ok {
			return true, withParam(rt.handlers, param)
		}
	}
	return false, nil
}

// withParam 包装处理器链，在第一个处理器前设置 ParamPath
func withParam(handlers []HandlerFunc, param string) []HandlerFunc {
	if len(handlers) == 0 {
		return handlers
	}
	out := make([]HandlerFunc, 0, len(handlers)+1)
	out = append(out, func(c *Context) {
		c.ParamPath = param
	})
	out = append(out, handlers...)
	return out
}

// MatchPattern 匹配路径模式
// 支持精确匹配和通配符匹配（如 /api/*path）
// 导出用于测试
func MatchPattern(pattern, path string) (param string, ok bool) {
	if strings.Contains(pattern, "*path") {
		prefix := strings.TrimSuffix(pattern, "*path")
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if !strings.HasPrefix(path, prefix) {
			return "", false
		}
		rest := strings.TrimPrefix(path, prefix)
		if rest == "" {
			return "/", true
		}
		return "/" + rest, true
	}
	return "", pattern == path
}
