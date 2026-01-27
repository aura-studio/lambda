package event

import (
	"fmt"
	"strings"
)

// HandlerFunc defines the handler function signature for event processing.
type HandlerFunc func(*Context)

// Context represents the request context for event processing.
type Context struct {
	Engine *Engine

	// RawPath is the path as provided by the request (before link rewriting).
	RawPath string
	// Path is the current effective path (after link rewriting).
	Path string
	// ParamPath is the wildcard parameter value for routes like /api/*path.
	// It includes a leading slash, e.g. /pkg/commit/route.
	ParamPath string

	Request string // 请求负载
	Err     error  // 错误信息

	DebugMode bool

	aborted bool
}

// Abort stops the handler chain execution.
// Validates: Requirement 2.5
func (c *Context) Abort() { c.aborted = true }

// route represents a single route with its pattern and handlers.
type route struct {
	pattern  string
	handlers []HandlerFunc
}

// Router handles request routing and middleware execution.
type Router struct {
	pre      []HandlerFunc // 前置中间件
	routes   []route       // 路由表
	noRoute  []HandlerFunc // 未匹配路由处理器
	noMethod []HandlerFunc // 方法不允许处理器（预留）
}

// NewRouter creates a new Router instance.
func NewRouter() *Router {
	return &Router{}
}

// Use registers middleware handlers that run before route handlers.
// Validates: Requirement 2.4
func (r *Router) Use(handlers ...HandlerFunc) {
	r.pre = append(r.pre, handlers...)
}

// Handle registers a route pattern with its handlers.
// Validates: Requirement 2.1
func (r *Router) Handle(pattern string, handlers ...HandlerFunc) {
	r.routes = append(r.routes, route{pattern: pattern, handlers: handlers})
}

// NoRoute sets the handlers for unmatched routes.
// Validates: Requirement 2.2
func (r *Router) NoRoute(handlers ...HandlerFunc) { r.noRoute = handlers }

// NoMethod sets the handlers for method not allowed (reserved for future use).
func (r *Router) NoMethod(handlers ...HandlerFunc) { r.noMethod = handlers }

// Dispatch routes the request to the appropriate handler.
// Validates: Requirements 2.1, 2.2, 2.4, 2.5
func (r *Router) Dispatch(ctx *Context) {
	// Execute pre-middleware handlers first
	for _, h := range r.pre {
		if h == nil {
			continue
		}
		h(ctx)
		if ctx.aborted {
			return
		}
	}

	// Match route and get handlers
	matched, handlers := r.match(ctx.Path)
	if !matched {
		handlers = r.noRoute
	}
	if len(handlers) == 0 {
		ctx.Err = fmt.Errorf("no route for path: %q", ctx.Path)
		return
	}

	// Execute route handlers
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

// match finds a matching route for the given path.
func (r *Router) match(path string) (bool, []HandlerFunc) {
	for _, rt := range r.routes {
		param, ok := matchPattern(rt.pattern, path)
		if ok {
			return true, withParam(rt.handlers, param)
		}
	}
	return false, nil
}

// withParam wraps handlers to set the ParamPath on the context.
func withParam(handlers []HandlerFunc, param string) []HandlerFunc {
	// Param is set on ctx by the first handler invocation in this chain.
	// We achieve this by wrapping the first handler.
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

// matchPattern matches a path against a pattern.
// Supports wildcard patterns like /api/*path.
// Validates: Requirement 2.3
func matchPattern(pattern, path string) (param string, ok bool) {
	if strings.Contains(pattern, "*path") {
		prefix := strings.TrimSuffix(pattern, "*path")
		// pattern like "/api/*path" => prefix "/api/"
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
