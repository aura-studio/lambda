package sqs

import (
	"fmt"
	"strings"
)

type HandlerFunc func(*Context)

type Context struct {
	Engine *Engine

	// RawPath is the path as provided by the request (before link rewriting).
	RawPath string
	// Path is the current effective path (after link rewriting).
	Path string
	// ParamPath is the wildcard parameter value for routes like /api/*path.
	// It includes a leading slash, e.g. /pkg/commit/route.
	ParamPath string

	Request string
	Response string
	Err      error

	Debug bool

	aborted bool
}

func (c *Context) Abort() { c.aborted = true }

type route struct {
	pattern  string
	handlers []HandlerFunc
}

type router struct {
	pre []HandlerFunc

	routes []route
	noRoute []HandlerFunc
	noMethod []HandlerFunc
}

func newRouter() *router {
	return &router{}
}

func (r *router) Use(handlers ...HandlerFunc) {
	r.pre = append(r.pre, handlers...)
}

func (r *router) HandleAllMethods(pattern string, handlers ...HandlerFunc) {
	r.routes = append(r.routes, route{pattern: pattern, handlers: handlers})
}

func (r *router) NoRoute(handlers ...HandlerFunc) { r.noRoute = handlers }
func (r *router) NoMethod(handlers ...HandlerFunc) { r.noMethod = handlers }

func (r *router) dispatch(ctx *Context) {
	for _, h := range r.pre {
		if h == nil {
			continue
		}
		h(ctx)
		if ctx.aborted {
			return
		}
	}

	matched, handlers := r.match(ctx.Path)
	if !matched {
		handlers = r.noRoute
	}
	if len(handlers) == 0 {
		ctx.Err = fmt.Errorf("no route for path: %q", ctx.Path)
		return
	}

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

func (r *router) match(path string) (bool, []HandlerFunc) {
	for _, rt := range r.routes {
		param, ok := matchPattern(rt.pattern, path)
		if ok {
			return true, withParam(rt.handlers, param)
		}
	}
	return false, nil
}

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
