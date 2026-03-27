package http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aura-studio/cast"
	"github.com/gin-gonic/gin"
)

const (
	ContextHeader       = "Header"
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
	ReqMetaHost       = "Host"
	ReqMetaRemoteAddr = "RemoteAddr"
	ReqMetaPath       = "Path"
)

const (
	RspMetaError       = "Error"
	RspMetaContentType = "ContentType"
	RspMetaStatus      = "Status"
)

type (
	Proccessor   = func(*gin.Context, LocalHandler)
	LocalHandler = func(string, string) (string, error)
)

var methods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}

func (e *Engine) InstallHandlers() {
	e.Use(e.StaticLink, e.PrefixLink)

	e.HandleAllMethods("/", e.OK)
	e.HandleAllMethods("/health-check", e.OK)
	e.HandleAllMethods("/api/*path", e.API)
	e.HandleAllMethods("/_/api/*path", e.Debug, e.API)
	e.HandleAllMethods("/wapi/*path", e.WAPI)
	e.HandleAllMethods("/_/wapi/*path", e.Debug, e.WAPI)
	e.HandleAllMethods("/meta/*path", e.Meta)
	e.NoRoute(e.PageNotFound)
	e.NoMethod(e.MethodNotAllowed)
}

func (e *Engine) HandleAllMethods(relativePath string, handlers ...gin.HandlerFunc) {
	for _, method := range methods {
		e.Handle(method, relativePath, handlers...)
	}
}

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	path = strings.ReplaceAll(path, "\\", "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func (e *Engine) StaticLink(c *gin.Context) {
	c.Request.URL.Path = normalizePath(c.Request.URL.Path)
	if dstPath, ok := e.StaticLinkMap[c.Request.URL.Path]; ok {
		c.Request.URL.Path = dstPath
		e.HandleContext(c)
		c.Abort()
		return
	}
}

func (e *Engine) PrefixLink(c *gin.Context) {
	c.Request.URL.Path = normalizePath(c.Request.URL.Path)
	for oldPrefix, newPrefix := range e.PrefixLinkMap {
		if strings.HasPrefix(c.Request.URL.Path, oldPrefix) {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, oldPrefix, newPrefix, 1)
			e.HandleContext(c)
			c.Abort()
			return
		}
	}
}

func (e *Engine) OK(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
	c.Abort()
}

func (e *Engine) Debug(c *gin.Context) {
	c.Set(ContextDebug, true)
}

func (e *Engine) API(c *gin.Context) {
	// path
	c.Set(ContextPath, c.Param("path"))

	// header
	c.Set(ContextHeader, c.Request.Header)

	// meta
	c.Set(ContextRequestMeta, e.genReqMeta(c))

	// request
	switch c.Request.Method {
	case http.MethodGet, "": // empty method treated as GET
		c.Set(ContextRequest, e.genGetReq(c))
	case http.MethodPost:
		c.Set(ContextRequest, e.genPostReq(c))
	default:
		c.Set(ContextRequest, "")
	}

	// processor
	if c.GetBool(ContextDebug) {
		c.Set(ContextProcessor, e.debugProcessor)
	} else {
		c.Set(ContextProcessor, e.safeProcessor)
	}

	// handle
	if v, ok := c.Get(ContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		log.Println("processor not found")
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	}

	// response
	if c.GetBool(ContextDebug) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(ContextPanic); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else if v, ok := c.Get(ContextError); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else {
		rspMeta := c.GetStringMap(ContextResponseMeta)
		rspBody := c.GetString(ContextResponse)
		contentType := "application/json"
		statusCode := http.StatusOK

		if rspMeta != nil {
			if e, ok := rspMeta[RspMetaError]; ok && e != nil && e != "" {
				c.String(http.StatusInternalServerError, cast.ToString(e))
				c.Abort()
				return
			}
			if ct, ok := rspMeta[RspMetaContentType]; ok && ct != nil && ct != "" {
				contentType = cast.ToString(ct)
			}
			if s, ok := rspMeta[RspMetaStatus]; ok && s != nil {
				statusCode = cast.ToInt(s)
			}
		}

		c.Data(statusCode, contentType, []byte(rspBody))
		c.Abort()
		return
	}
}

func (e *Engine) WAPI(c *gin.Context) {
	// path
	c.Set(ContextPath, c.Param("path"))

	// header
	c.Set(ContextHeader, c.Request.Header)

	// request
	switch c.Request.Method {
	case http.MethodGet, "": // empty method treated as GET
		c.Set(ContextRequest, e.genGetReq(c))
	case http.MethodPost:
		c.Set(ContextRequest, e.genPostReq(c))
	default:
		c.Set(ContextRequest, "")
	}

	// processor
	if c.GetBool(ContextDebug) {
		c.Set(ContextProcessor, e.debugWireProcessor)
	} else {
		c.Set(ContextProcessor, e.safeWireProcessor)
	}

	// handle
	if v, ok := c.Get(ContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		log.Println("processor not found")
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	}

	// response
	if c.GetBool(ContextDebug) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(ContextPanic); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else if v, ok := c.Get(ContextError); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else {
		response, err := http.ReadResponse(bufio.NewReader(strings.NewReader(c.GetString(ContextResponse))), c.Request)
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			c.Abort()
			return
		}
		c.Writer.WriteHeader(response.StatusCode)
		for k, v := range response.Header {
			c.Writer.Header().Set(k, v[0])
		}
		io.Copy(c.Writer, response.Body)
		response.Body.Close()
		c.Abort()
		return
	}
}

func (e *Engine) Meta(c *gin.Context) {
	// path
	c.Set(ContextPath, c.Param("path"))

	// processor
	c.Set(ContextProcessor, e.safeMetaProcessor)

	// handle
	if v, ok := c.Get(ContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		log.Println("processor not found")
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	}

	// response
	if v, ok := c.Get(ContextPanic); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else if v, ok := c.Get(ContextError); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else {
		c.String(http.StatusOK, c.GetString(ContextResponse))
		c.Abort()
		return
	}
}

func (e *Engine) PageNotFound(c *gin.Context) {
	e.StaticLink(c)
	if c.IsAborted() {
		return
	}
	e.PrefixLink(c)
	if c.IsAborted() {
		return
	}

	if e.PageNotFoundPath != "" {
		c.Request.URL.Path = e.PageNotFoundPath
		e.HandleContext(c)
		c.Abort()
		return
	}
	c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	c.Abort()
}

func (e *Engine) MethodNotAllowed(c *gin.Context) {
	c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	c.Abort()
}

func (e *Engine) genReqMeta(c *gin.Context) map[string]any {
	meta := map[string]any{}

	// resolve host
	host := c.Request.Host
	if v := c.GetHeader("X-Forwarded-Host"); v != "" {
		host = v
	} else if v := c.GetHeader("Host"); v != "" {
		host = v
	}
	meta[ReqMetaHost] = strings.Split(strings.Split(host, ", ")[0], ":")[0]

	// resolve remote addr
	remoteAddr := c.Request.RemoteAddr
	if v := c.GetHeader("CloudFront-Viewer-Address"); v != "" {
		remoteAddr = v
	} else if v := c.GetHeader("X-Forwarded-For"); v != "" {
		remoteAddr = v
	}
	meta[ReqMetaRemoteAddr] = strings.Split(remoteAddr, ", ")[0]

	// path
	meta[ReqMetaPath] = c.Request.URL.Path

	return meta
}

func (e *Engine) genGetReq(c *gin.Context) string {
	dataMap := map[string]any{}
	for k, v := range c.Request.URL.Query() {
		dataMap[k] = v[0]
	}
	data, err := json.Marshal(dataMap)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}

func (e *Engine) genPostReq(c *gin.Context) string {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Request.Body.Close()

	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))

	return string(data)
}

func (e *Engine) doWireProcessor(c *gin.Context, f LocalHandler) {
	var (
		wireReq string
		wireRsp string
		err     error
	)
	defer func() {
		c.Set(ContextRequest, wireReq)
		c.Set(ContextResponse, wireRsp)
		c.Set(ContextError, err)
	}()

	path := c.GetString(ContextPath)

	var buf bytes.Buffer
	err = c.Request.Write(&buf)
	if err != nil {
		return
	}
	wireReq = buf.String()

	wireRsp, err = f(path, wireReq)
}

func (e *Engine) safeWireProcessor(c *gin.Context, f LocalHandler) {
	c.Set(ContextPanic, e.doSafe(func() {
		e.doWireProcessor(c, f)
	}))
}

func (e *Engine) debugWireProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doWireProcessor(c, f)
	})
	c.Set(ContextStdout, stdout)
	c.Set(ContextStderr, stderr)
	c.Set(ContextPanic, panicErr)
}

func (e *Engine) doProcessor(c *gin.Context, f LocalHandler) {
	path := c.GetString(ContextPath)
	req := c.GetString(ContextRequest)
	reqMeta := c.GetStringMap(ContextRequestMeta)

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
		c.Set(ContextResponse, "")
		c.Set(ContextError, err)
		return
	}

	rsp, err := f(path, string(reqBytes))
	if err != nil {
		c.Set(ContextResponse, "")
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
		c.Set(ContextError, nil)
		return
	}

	if len(rspEnvelope.Meta) > 0 {
		c.Set(ContextResponseMeta, rspEnvelope.Meta)
	}

	if errMsg := cast.ToString(rspEnvelope.Meta[RspMetaError]); errMsg != "" {
		c.Set(ContextResponse, "")
		c.Set(ContextError, cast.ToError(errMsg))
		return
	}

	data, decErr := base64.StdEncoding.DecodeString(rspEnvelope.Data)
	if decErr != nil {
		c.Set(ContextResponse, "")
		c.Set(ContextError, decErr)
		return
	}
	c.Set(ContextResponse, string(data))
	c.Set(ContextError, nil)
}

func (e *Engine) safeProcessor(c *gin.Context, f LocalHandler) {
	c.Set(ContextPanic, e.doSafe(func() {
		e.doProcessor(c, f)
	}))
}

func (e *Engine) debugProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doProcessor(c, f)
	})
	c.Set(ContextStdout, stdout)
	c.Set(ContextStderr, stderr)
	c.Set(ContextPanic, panicErr)
}

func (e *Engine) doMetaProcessor(c *gin.Context) {
	path := c.GetString(ContextPath)
	rsp, err := e.meta(path)
	c.Set(ContextResponse, rsp)
	c.Set(ContextError, err)
}

func (e *Engine) safeMetaProcessor(c *gin.Context, f LocalHandler) {
	c.Set(ContextPanic, e.doSafe(func() {
		e.doMetaProcessor(c)
	}))
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

func (e *Engine) formatDebug(c *gin.Context) string {
	var buf bytes.Buffer
	buf.WriteString(`Schema: `)
	buf.WriteString(c.Request.URL.Scheme)
	buf.WriteString("\n")
	buf.WriteString(`Method: `)
	buf.WriteString(c.Request.Method)
	buf.WriteString("\n")
	buf.WriteString(`Host: `)
	buf.WriteString(c.Request.URL.Host)
	buf.WriteString("\n")
	buf.WriteString(`Path: `)
	buf.WriteString(c.GetString(ContextPath))
	buf.WriteString("\n")
	buf.WriteString(`Header: `)
	headerBytes, _ := json.Marshal(c.GetString(ContextHeader))
	buf.WriteString(string(headerBytes))
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
	if v, ok := c.Get(ContextError); ok && v != nil {
		buf.WriteString(v.(error).Error())
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

	// keep backup of the real file
	originStdout := os.Stdout
	originStderr := os.Stderr

	// Restore original file
	defer func() {
		os.Stdout = originStdout
		os.Stderr = originStderr
	}()

	// Create pipe to create reader & writer
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

	// Connect file to writer side of pipe
	os.Stdout = stdoutPipeWriter
	os.Stderr = stderrPipeWriter

	// Create MultiWriter to write to buffer and file at the same time
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	stdoutMultiWriter := io.MultiWriter(&stdoutBuf, originStdout)
	stderrMultiWriter := io.MultiWriter(&stderrBuf, originStderr)

	// copy the output in a separate goroutine so printing can't block indefinitely
	stdoutErrCh := make(chan error, 1)
	go func() {
		if _, err := io.Copy(stdoutMultiWriter, stdoutPipeReader); err != nil {
			stdoutErrCh <- err
		}
		stdoutErrCh <- nil
	}()
	go func() {
		if _, err := io.Copy(stderrMultiWriter, stderrPipeReader); err != nil {
			stdoutErrCh <- err
		}
		stdoutErrCh <- nil
	}()

	f()

	if err := stdoutPipeWriter.Close(); err != nil {
		panic(err)
	}

	if err := stderrPipeWriter.Close(); err != nil {
		panic(err)
	}

	if err := <-stdoutErrCh; err != nil {
		panic(err)
	}

	if err := <-stdoutErrCh; err != nil {
		panic(err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}
