package sqs

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (e *Engine) InstallHandlers() {
	if e.r == nil {
		e.r = newRouter()
	}

	e.r.Use(e.HeaderLink, e.StaticLink, e.PrefixLink)

	e.r.Handle("/", e.OK)
	e.r.Handle("/health-check", e.OK)
	e.r.Handle("/api/*path", e.API)
	e.r.Handle("/_/api/*path", e.Debug, e.API)
	e.r.Handle("/wapi/*path", e.WAPI)
	e.r.Handle("/_/wapi/*path", e.Debug, e.WAPI)
}

func (e *Engine) Use(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = newRouter()
	}
	e.r.Use(handlers...)
}

func (e *Engine) Handle(pattern string, handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = newRouter()
	}
	e.r.Handle(pattern, handlers...)
}

func (e *Engine) NoRoute(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = newRouter()
	}
	e.r.NoRoute(handlers...)
}

func (e *Engine) HeaderLink(c *Context) {
	// SQS request currently has no headers. Reserved for future.
}

func (e *Engine) StaticLink(c *Context) {
	if e.StaticLinkMap == nil {
		return
	}
	if dst, ok := e.StaticLinkMap[c.Path]; ok {
		c.Path = dst
	}
}

func (e *Engine) PrefixLink(c *Context) {
	if e.PrefixLinkMap == nil {
		return
	}
	for oldPrefix, newPrefix := range e.PrefixLinkMap {
		if strings.HasPrefix(c.Path, oldPrefix) {
			c.Path = strings.Replace(c.Path, oldPrefix, newPrefix, 1)
			return
		}
	}
}

func (e *Engine) OK(c *Context) {
	c.Response = "OK"
}

func (e *Engine) Debug(c *Context) {
	c.Debug = true
}

func (e *Engine) API(c *Context) {
	if c.ParamPath == "" {
		c.Err = fmt.Errorf("missing api path")
		return
	}
	rsp, err := e.invokeFunc(c.ParamPath, c.Request)
	if err != nil {
		c.Err = err
		if c.Debug {
			c.Response = e.formatDebug(c, "api")
		}
		return
	}
	if c.Debug {
		c.Response = e.formatDebugWithResponse(c, "api", rsp)
		return
	}
	c.Response = rsp
}

func (e *Engine) WAPI(c *Context) {
	// For now, WAPI behaves like API in SQS.
	// It exists to mirror http handler layout and can be evolved later.
	e.API(c)
}

func (e *Engine) PageNotFound(c *Context) {
	c.Err = fmt.Errorf("404 page not found: %s", c.Path)
}

func (e *Engine) MethodNotAllowed(c *Context) {
	c.Err = fmt.Errorf("405 method not allowed")
}

func (e *Engine) formatDebug(c *Context, mode string) string {
	data, _ := json.Marshal(map[string]any{
		"mode":     mode,
		"raw_path": c.RawPath,
		"path":     c.Path,
		"param":    c.ParamPath,
		"request":  c.Request,
		"error":    errString(c.Err),
	})
	return string(data)
}

func (e *Engine) formatDebugWithResponse(c *Context, mode string, rsp string) string {
	data, _ := json.Marshal(map[string]any{
		"mode":     mode,
		"raw_path": c.RawPath,
		"path":     c.Path,
		"param":    c.ParamPath,
		"request":  c.Request,
		"response": rsp,
		"error":    errString(c.Err),
	})
	return string(data)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
