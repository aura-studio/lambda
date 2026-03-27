package reqresp

import (
	"fmt"
	"strings"
)

type HandlerFunc func(*Context)

type Context struct {
	Keys map[string]any

	aborted bool
}

func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (any, bool) {
	if c.Keys == nil {
		return nil, false
	}
	v, ok := c.Keys[key]
	return v, ok
}

func (c *Context) GetString(key string) string {
	if v, ok := c.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Context) GetBool(key string) bool {
	if v, ok := c.Get(key); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (c *Context) GetStringMap(key string) map[string]any {
	if v, ok := c.Get(key); ok {
		if m, ok := v.(map[string]any); ok {
			return m
		}
	}
	return nil
}

func (c *Context) GetError() error {
	if v, ok := c.Get(ContextError); ok {
		if e, ok := v.(error); ok {
			return e
		}
	}
	return nil
}

func (c *Context) Abort() { c.aborted = true }

type route struct {
	pattern  string
	handlers []HandlerFunc
}

type Router struct {
	pre     []HandlerFunc
	routes  []route
	noRoute []HandlerFunc
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Use(handlers ...HandlerFunc) {
	r.pre = append(r.pre, handlers...)
}

func (r *Router) Handle(pattern string, handlers ...HandlerFunc) {
	r.routes = append(r.routes, route{pattern: pattern, handlers: handlers})
}

func (r *Router) NoRoute(handlers ...HandlerFunc) { r.noRoute = handlers }

func (r *Router) Dispatch(ctx *Context) {
	for _, h := range r.pre {
		if h == nil {
			continue
		}
		h(ctx)
		if ctx.aborted {
			return
		}
	}

	matched, handlers := r.match(ctx.GetString(ContextPath))
	if !matched {
		handlers = r.noRoute
	}
	if len(handlers) == 0 {
		ctx.Set(ContextError, fmt.Errorf("no route for path: %q", ctx.GetString(ContextPath)))
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
		if ctx.GetError() != nil {
			return
		}
	}
}

func (r *Router) match(path string) (bool, []HandlerFunc) {
	for _, rt := range r.routes {
		param, ok := MatchPattern(rt.pattern, path)
		if ok {
			return true, withParam(rt.handlers, param)
		}
	}
	return false, nil
}

func withParam(handlers []HandlerFunc, param string) []HandlerFunc {
	if len(handlers) == 0 {
		return handlers
	}
	out := make([]HandlerFunc, 0, len(handlers)+1)
	out = append(out, func(c *Context) {
		c.Set(ContextPath, param)
	})
	out = append(out, handlers...)
	return out
}

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
