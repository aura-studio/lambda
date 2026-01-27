package reqresp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InstallHandlers 注册默认路由处理器
func (e *Engine) InstallHandlers() {
	if e.r == nil {
		e.r = NewRouter()
	}

	e.r.Use(e.HeaderLink, e.StaticLink, e.PrefixLink)

	e.HandleAllMethods("/", e.OK)
	e.HandleAllMethods("/health-check", e.OK)
	e.HandleAllMethods("/api/*path", e.API)
	e.HandleAllMethods("/_/api/*path", e.Debug, e.API)
	e.HandleAllMethods("/wapi/*path", e.WAPI)
	e.HandleAllMethods("/_/wapi/*path", e.Debug, e.WAPI)
	e.r.NoRoute(e.PageNotFound)
	e.r.NoMethod(e.MethodNotAllowed)
}

// Use 注册前置中间件
func (e *Engine) Use(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.Use(handlers...)
}

// Handle 注册路由处理器
func (e *Engine) Handle(pattern string, handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.Handle(pattern, handlers...)
}

// HandleAllMethods 注册所有方法的路由处理器
func (e *Engine) HandleAllMethods(pattern string, handlers ...HandlerFunc) {
	e.Handle(pattern, handlers...)
}

// NoRoute 设置未匹配路由处理器
func (e *Engine) NoRoute(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.NoRoute(handlers...)
}

// NoMethod 设置方法不允许处理器
func (e *Engine) NoMethod(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.NoMethod(handlers...)
}

// HeaderLink 头部链接中间件（预留）
func (e *Engine) HeaderLink(c *Context) {
	// ReqResp request currently has no headers. Reserved for future.
}

// StaticLink 静态路径映射中间件
func (e *Engine) StaticLink(c *Context) {
	if e.StaticLinkMap == nil {
		return
	}
	if dst, ok := e.StaticLinkMap[c.Path]; ok {
		c.Path = dst
	}
}

// PrefixLink 前缀路径映射中间件
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

// OK 健康检查处理器
func (e *Engine) OK(c *Context) {
	c.Response = "OK"
}

// Debug 调试模式处理器
func (e *Engine) Debug(c *Context) {
	c.DebugMode = true
}

// API 处理器用于调用 Dynamic 业务包
func (e *Engine) API(c *Context) {
	if c.ParamPath == "" {
		c.Err = fmt.Errorf("missing api path")
		return
	}
	rsp, err := e.handle(c.ParamPath, c.Request)
	if err != nil {
		c.Err = err
		if c.DebugMode {
			c.Response = e.FormatDebug(c, "api")
		}
		return
	}
	if c.DebugMode {
		c.Response = e.FormatDebugWithResponse(c, "api", rsp)
		return
	}
	c.Response = rsp
}

// WAPI 处理器作为 API 的别名
func (e *Engine) WAPI(c *Context) {
	e.API(c)
}

// PageNotFound 处理器用于处理未匹配路由
func (e *Engine) PageNotFound(c *Context) {
	c.Err = fmt.Errorf("404 page not found: %s", c.Path)
}

// MethodNotAllowed 方法不允许处理器
func (e *Engine) MethodNotAllowed(c *Context) {
	c.Err = fmt.Errorf("405 method not allowed")
}

// FormatDebug 格式化调试信息
func (e *Engine) FormatDebug(c *Context, mode string) string {
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

// FormatDebugWithResponse 格式化带响应的调试信息
func (e *Engine) FormatDebugWithResponse(c *Context, mode string, rsp string) string {
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
