package sqs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aura-studio/cast"
)

const (
	RspMetaError = "Error"
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
	rsp, err := e.process(c.ParamPath, c.Request)
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
		"error":    cast.ToString(c.Err),
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
		"error":    cast.ToString(c.Err),
	})
	return string(data)
}

func (e *Engine) process(path string, req string) (string, error) {
	// 封装请求: {"meta":{}, "data":"base64"}
	reqEnvelope := struct {
		Meta map[string]any `json:"meta"`
		Data string         `json:"data"`
	}{
		Meta: map[string]any{},
		Data: base64.StdEncoding.EncodeToString([]byte(req)),
	}
	reqBytes, err := json.Marshal(reqEnvelope)
	if err != nil {
		return "", err
	}

	rsp, err := e.handle(path, string(reqBytes))
	if err != nil {
		return "", err
	}

	// 解封装响应: {"meta":{}, "data":"base64"}
	var rspEnvelope struct {
		Meta map[string]any `json:"meta"`
		Data string         `json:"data"`
	}
	if err := json.Unmarshal([]byte(rsp), &rspEnvelope); err != nil {
		return rsp, nil
	}

	if errMsg := cast.ToString(rspEnvelope.Meta[RspMetaError]); errMsg != "" {
		return "", cast.ToError(errMsg)
	}

	data, err := base64.StdEncoding.DecodeString(rspEnvelope.Data)
	if err != nil {
		return "", err
	}
	return string(data), nil
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

	return rsp, nil
}

func (e *Engine) meta(path string) (string, error) {
	var tunnelMeta string
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		pkg := parts[0]
		version := parts[1]
		if tunnel, err := e.GetPackage(pkg, version); err == nil {
			tunnelMeta = tunnel.Meta()
		}
	}

	return e.MetaGenerator.Generate(tunnelMeta), nil
}
