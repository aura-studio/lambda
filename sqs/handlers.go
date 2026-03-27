package sqs

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aura-studio/cast"
)

const (
	ContextPath         = "Path"
	ContextRequest      = "Request"
	ContextResponse     = "Response"
	ContextRequestMeta  = "RequestMeta"
	ContextResponseMeta = "ResponseMeta"
	ContextError        = "Error"
	ContextPanic        = "Panic"
	ContextDebug        = "Debug"
	ContextStdout       = "Stdout"
	ContextStderr       = "Stderr"
	ContextProcessor    = "Processor"
)

const (
	RspMetaError = "Error"
)

func (e *Engine) InstallHandlers() {
	e.Handle("/", e.OK)
	e.Handle("/health-check", e.OK)
	e.Handle("/api/*path", e.API)
	e.Handle("/_/api/*path", e.Debug, e.API)
	e.Handle("/meta/*path", e.Meta)
	e.NoRoute(e.PageNotFound)
}

func (e *Engine) OK(c *Context) {
	c.Set(ContextResponse, "OK")
}

func (e *Engine) Debug(c *Context) {
	c.Set(ContextDebug, true)
}

func (e *Engine) API(c *Context) {
	if c.GetString(ContextPath) == "" {
		c.Set(ContextError, fmt.Errorf("missing api path"))
		return
	}

	// processor
	if c.GetBool(ContextDebug) {
		c.Set(ContextProcessor, e.debugProcessor)
	} else {
		c.Set(ContextProcessor, e.safeProcessor)
	}
	if v, ok := c.Get(ContextProcessor); ok {
		v.(func(*Context))(c)
	}

	// response
	if c.GetBool(ContextDebug) {
		c.Set(ContextResponse, e.formatDebug(c))
		return
	}
}

func (e *Engine) Meta(c *Context) {
	if c.GetString(ContextPath) == "" {
		c.Set(ContextError, fmt.Errorf("missing meta path"))
		return
	}

	// processor
	e.safeMetaProcessor(c)

	// response
	if c.GetBool(ContextDebug) {
		c.Set(ContextResponse, e.formatDebug(c))
		return
	}
}

func (e *Engine) PageNotFound(c *Context) {
	c.Set(ContextError, fmt.Errorf("404 page not found: %s", c.GetString(ContextPath)))
}

// ==================== processor ====================

func (e *Engine) doProcessor(c *Context) {
	path := c.GetString(ContextPath)
	req := c.GetString(ContextRequest)
	reqMeta := c.GetStringMap(ContextRequestMeta)
	if reqMeta == nil {
		reqMeta = map[string]any{}
	}

	// 封装请求: {"meta":{...}, "data":"base64"}
	reqEnvelope := struct {
		Meta map[string]any `json:"meta"`
		Data string         `json:"data"`
	}{
		Meta: reqMeta,
		Data: base64.StdEncoding.EncodeToString([]byte(req)),
	}
	reqBytes, err := json.Marshal(reqEnvelope)
	if err != nil {
		c.Set(ContextError, err)
		return
	}

	rsp, err := e.handle(path, string(reqBytes))
	if err != nil {
		c.Set(ContextError, err)
		return
	}

	// 解封装响应: {"meta":{...}, "data":"base64"}
	var rspEnvelope struct {
		Meta map[string]any `json:"meta"`
		Data string         `json:"data"`
	}
	if err := json.Unmarshal([]byte(rsp), &rspEnvelope); err != nil {
		c.Set(ContextResponse, rsp)
		return
	}

	if len(rspEnvelope.Meta) > 0 {
		c.Set(ContextResponseMeta, rspEnvelope.Meta)
	}

	if errMsg := cast.ToString(rspEnvelope.Meta[RspMetaError]); errMsg != "" {
		c.Set(ContextError, cast.ToError(errMsg))
		return
	}

	data, err := base64.StdEncoding.DecodeString(rspEnvelope.Data)
	if err != nil {
		c.Set(ContextError, err)
		return
	}
	c.Set(ContextResponse, string(data))
}

func (e *Engine) safeProcessor(c *Context) {
	c.Set(ContextPanic, e.doSafe(func() {
		e.doProcessor(c)
	}))
}

func (e *Engine) debugProcessor(c *Context) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doProcessor(c)
	})
	c.Set(ContextStdout, stdout)
	c.Set(ContextStderr, stderr)
	c.Set(ContextPanic, panicErr)
}

func (e *Engine) doMetaProcessor(c *Context) {
	path := c.GetString(ContextPath)
	rsp, err := e.meta(path)
	c.Set(ContextResponse, rsp)
	c.Set(ContextError, err)
}

func (e *Engine) safeMetaProcessor(c *Context) {
	c.Set(ContextPanic, e.doSafe(func() {
		e.doMetaProcessor(c)
	}))
}

// ==================== handle / meta ====================

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

// ==================== formatDebug ====================

func (e *Engine) formatDebug(c *Context) string {
	var buf bytes.Buffer
	buf.WriteString(`Path: `)
	buf.WriteString(c.GetString(ContextPath))
	buf.WriteString("\n")
	buf.WriteString(`Request Meta: `)
	reqMetaBytes, _ := json.Marshal(c.GetStringMap(ContextRequestMeta))
	buf.WriteString(string(reqMetaBytes))
	buf.WriteString("\n")
	buf.WriteString(`Response Meta: `)
	rspMetaBytes, _ := json.Marshal(c.GetStringMap(ContextResponseMeta))
	buf.WriteString(string(rspMetaBytes))
	buf.WriteString("\n")
	buf.WriteString(`Stdout: `)
	buf.WriteString(c.GetString(ContextStdout))
	buf.WriteString("\n")
	buf.WriteString(`Stderr: `)
	buf.WriteString(c.GetString(ContextStderr))
	buf.WriteString("\n")
	buf.WriteString(`Error: `)
	if err := c.GetError(); err != nil {
		buf.WriteString(err.Error())
	}
	buf.WriteString("\n")
	buf.WriteString(`Panic: `)
	if v, ok := c.Get(ContextPanic); ok && v != nil {
		buf.WriteString(v.(error).Error())
	}
	buf.WriteString("\n")
	buf.WriteString(`Request: `)
	buf.WriteString(c.GetString(ContextRequest))
	buf.WriteString("\n")
	buf.WriteString(`Response: `)
	buf.WriteString(c.GetString(ContextResponse))
	buf.WriteString("\n")
	return buf.String()
}

// ==================== doSafe / doDebug ====================

func (e *Engine) doSafe(f func()) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()

	f()

	return nil
}

func (e *Engine) doDebug(f func()) (stdout string, stderr string, err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()

	originStdout := os.Stdout
	originStderr := os.Stderr
	defer func() {
		os.Stdout = originStdout
		os.Stderr = originStderr
	}()

	stdoutPipeReader, stdoutPipeWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer stdoutPipeWriter.Close()
	stderrPipeReader, stderrPipeWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer stderrPipeWriter.Close()

	os.Stdout = stdoutPipeWriter
	os.Stderr = stderrPipeWriter

	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	stdoutMultiWriter := io.MultiWriter(&stdoutBuf, originStdout)
	stderrMultiWriter := io.MultiWriter(&stderrBuf, originStderr)

	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(stdoutMultiWriter, stdoutPipeReader)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(stderrMultiWriter, stderrPipeReader)
		errCh <- err
	}()

	f()

	stdoutPipeWriter.Close()
	stderrPipeWriter.Close()
	<-errCh
	<-errCh

	return stdoutBuf.String(), stderrBuf.String(), nil
}
