package websocket

import (
	"context"
	"fmt"
)

type HandlerFunc func(*Context)

type Context struct {
	Keys map[string]any

	ctx     context.Context
	engine  *Engine
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

// GetStringMapString reads a map[string]string value (e.g. ContextHeaders,
// ContextQuery) from the context.
func (c *Context) GetStringMapString(key string) map[string]string {
	if v, ok := c.Get(key); ok {
		if m, ok := v.(map[string]string); ok {
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

// StdContext returns the standard request-scoped context of the current invocation.
func (c *Context) StdContext() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// Post pushes data back to the connection that triggered this context, via the
// API Gateway management API. It is a shorthand for Engine.PostToConnection with
// the domain/stage/connection of the current event.
func (c *Context) Post(data []byte) error {
	if c.engine == nil {
		return fmt.Errorf("websocket: context is not bound to an engine")
	}
	return c.engine.PostToConnection(
		c.StdContext(),
		c.GetString(ContextDomainName),
		c.GetString(ContextStage),
		c.GetString(ContextConnectionID),
		data,
	)
}

func (c *Context) Abort() { c.aborted = true }

type route struct {
	key      string
	handlers []HandlerFunc
}

type Router struct {
	pre []HandlerFunc

	routes  []route
	noRoute []HandlerFunc
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Use(handlers ...HandlerFunc) {
	r.pre = append(r.pre, handlers...)
}

func (r *Router) Handle(key string, handlers ...HandlerFunc) {
	r.routes = append(r.routes, route{key: key, handlers: handlers})
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

	handlers := r.match(ctx.GetString(ContextRouteKey))
	if handlers == nil {
		handlers = r.noRoute
	}
	if len(handlers) == 0 {
		ctx.Set(ContextError, fmt.Errorf("no route for key: %q", ctx.GetString(ContextRouteKey)))
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

func (r *Router) match(key string) []HandlerFunc {
	for _, rt := range r.routes {
		if rt.key == key {
			return rt.handlers
		}
	}
	return nil
}
