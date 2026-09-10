package websocket

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
	ContextRouteKey       = "RouteKey"
	ContextEventType      = "EventType"
	ContextConnectionID   = "ConnectionID"
	ContextDomainName     = "DomainName"
	ContextStage          = "Stage"
	ContextRequestContext = "RequestContext"
	ContextHeaders        = "Headers"
	ContextQuery          = "Query"
	ContextPath           = "Path"
	ContextRequest        = "Request"
	ContextResponse       = "Response"
	ContextRequestMeta    = "RequestMeta"
	ContextResponseMeta   = "ResponseMeta"
	ContextError          = "Error"
	ContextPanic          = "Panic"
	ContextDebug          = "Debug"
	ContextStdout         = "Stdout"
	ContextStderr         = "Stderr"
	ContextProcessor      = "Processor"
)

const (
	RspMetaError = "Error"
)

func (e *Engine) InstallHandlers() {
	e.Handle("$connect", e.Connect)
	e.Handle("$disconnect", e.Disconnect)
	if e.Options.DefaultHandler != nil {
		e.Handle("$default", e.Options.DefaultHandler)
	} else {
		e.Handle("$default", e.Message)
	}
	e.NoRoute(e.RouteNotFound)
}

// Connect handles $connect. The default behavior accepts every connection; a
// ConnectHandler option may inspect the request and set ContextError to reject.
func (e *Engine) Connect(c *Context) {
	if e.Options.ConnectHandler != nil {
		e.Options.ConnectHandler(c)
	}
}

// Disconnect handles $disconnect. The gateway ignores the integration response
// on this route; the handler exists for cleanup hooks only.
func (e *Engine) Disconnect(c *Context) {
	if e.Options.DisconnectHandler != nil {
		e.Options.DisconnectHandler(c)
	}
}

// Message handles $default. The body must be a JSON envelope:
//
//	{"path": "/{package}/{version}/{route}", "payload": <any>, "meta": {...}}
//
// The payload is dispatched to the package tunnel exactly like the sqs engine's
// /api route. With ReplyMode on, the response is pushed back to the requesting
// connection as {"path": ..., "payload": ..., "meta": {...}}, or
// {"path": ..., "error": ...} when the dispatch failed.
func (e *Engine) Message(c *Context) {
	var envelope struct {
		Path    string          `json:"path"`
		Payload json.RawMessage `json:"payload"`
		Meta    map[string]any  `json:"meta"`
	}
	if err := json.Unmarshal([]byte(c.GetString(ContextRequest)), &envelope); err != nil {
		c.Set(ContextError, fmt.Errorf("invalid message envelope: %w", err))
		e.replyError(c)
		return
	}
	if envelope.Path == "" {
		c.Set(ContextError, fmt.Errorf("missing message path"))
		e.replyError(c)
		return
	}

	c.Set(ContextPath, envelope.Path)
	c.Set(ContextRequest, rawMessageString(envelope.Payload))
	if len(envelope.Meta) > 0 {
		c.Set(ContextRequestMeta, envelope.Meta)
	}

	if e.DebugMode {
		c.Set(ContextDebug, true)
		c.Set(ContextProcessor, e.debugProcessor)
	} else {
		c.Set(ContextProcessor, e.safeProcessor)
	}
	if v, ok := c.Get(ContextProcessor); ok {
		v.(func(*Context))(c)
	}

	if c.GetBool(ContextDebug) {
		c.Set(ContextResponse, e.formatDebug(c))
	}

	e.reply(c)
}

func (e *Engine) RouteNotFound(c *Context) {
	c.Set(ContextError, fmt.Errorf("no route for key: %q", c.GetString(ContextRouteKey)))
}

// ==================== reply ====================

func (e *Engine) reply(c *Context) {
	if !e.ReplyMode {
		return
	}
	if err := c.GetError(); err != nil {
		e.replyError(c)
		return
	}

	body, err := marshalReply(c.GetString(ContextPath), c.GetString(ContextResponse), c.GetStringMap(ContextResponseMeta))
	if err != nil {
		c.Set(ContextError, err)
		e.replyError(c)
		return
	}

	if err := c.Post(body); err != nil {
		c.Set(ContextError, fmt.Errorf("post to connection %s: %w", c.GetString(ContextConnectionID), err))
	}
}

func (e *Engine) replyError(c *Context) {
	if !e.ReplyMode {
		return
	}
	err := c.GetError()
	if err == nil {
		return
	}
	body, marshalErr := json.Marshal(map[string]any{
		"path":  c.GetString(ContextPath),
		"error": err.Error(),
	})
	if marshalErr != nil {
		return
	}
	// Best effort: the original dispatch error stays on the context; a failed
	// push only logs, so a dead connection does not mask the real failure.
	if postErr := c.Post(body); postErr != nil {
		e.logf("post error reply to connection %s: %v", c.GetString(ContextConnectionID), postErr)
	}
}

// marshalReply embeds payload as raw JSON when it parses as such, otherwise as
// a plain string.
func marshalReply(path string, payload string, meta map[string]any) ([]byte, error) {
	reply := map[string]any{"path": path}
	if json.Valid([]byte(payload)) {
		reply["payload"] = json.RawMessage(payload)
	} else {
		reply["payload"] = payload
	}
	if len(meta) > 0 {
		reply["meta"] = meta
	}
	return json.Marshal(reply)
}

// rawMessageString unquotes a JSON string payload; non-string payloads are
// returned as their raw JSON text.
func rawMessageString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
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

// ==================== handle ====================

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

// ==================== formatDebug ====================

func (e *Engine) formatDebug(c *Context) string {
	var buf bytes.Buffer
	buf.WriteString(`RouteKey: `)
	buf.WriteString(c.GetString(ContextRouteKey))
	buf.WriteString("\n")
	buf.WriteString(`ConnectionID: `)
	buf.WriteString(c.GetString(ContextConnectionID))
	buf.WriteString("\n")
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
