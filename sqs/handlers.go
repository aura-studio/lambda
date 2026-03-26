package sqs

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (e *Engine) InstallHandlers() {
	if e.r == nil {
		e.r = NewRouter()
	}

	e.Handle("/", e.OK)
	e.Handle("/health-check", e.OK)
	e.Handle("/api/*path", e.API)
	e.Handle("/_/api/*path", e.Debug, e.API)
	e.Handle("/meta/*path", e.Meta)
	e.r.NoRoute(e.PageNotFound)
}

func (e *Engine) Use(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.Use(handlers...)
}

func (e *Engine) Handle(pattern string, handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.Handle(pattern, handlers...)
}

func (e *Engine) NoRoute(handlers ...HandlerFunc) {
	if e.r == nil {
		e.r = NewRouter()
	}
	e.r.NoRoute(handlers...)
}

func (e *Engine) OK(c *Context) {
	c.Response = "OK"
}

func (e *Engine) Debug(c *Context) {
	c.DebugMode = true
}

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

func (e *Engine) Meta(c *Context) {
	if c.ParamPath == "" {
		c.Err = fmt.Errorf("missing meta path")
		return
	}
	rsp, err := e.meta(c.ParamPath)
	if err != nil {
		c.Err = err
		return
	}
	c.Response = rsp
}

func (e *Engine) PageNotFound(c *Context) {
	c.Err = fmt.Errorf("404 page not found: %s", c.Path)
}

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
	rsp := tunnel.Invoke(route, req)

	// parse response prefix protocol
	if after, found := strings.CutPrefix(rsp, "error://"); found {
		return "", fmt.Errorf("%s", after)
	}
	if after, found := strings.CutPrefix(rsp, "data://"); found {
		return after, nil
	}
	return rsp, nil
}

func (e *Engine) meta(path string) (string, error) {
	// 获取 tunnel 的 meta 信息（如果路径有效）
	var tunnelMeta string
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		pkg := parts[0]
		version := parts[1]
		if tunnel, err := e.GetPackage(pkg, version); err == nil {
			tunnelMeta = tunnel.Meta()
		}
	}

	// 使用 MetaGenerator 生成完整的 meta 信息
	return e.MetaGenerator.Generate(tunnelMeta), nil
}
