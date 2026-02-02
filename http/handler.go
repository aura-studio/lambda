package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	HeaderContext       = "header"
	PathContext         = "path"
	RequestContext      = "request"
	ResponseContext     = "response"
	RequestMetaContext  = "request_meta"
	ResponseMetaContext = "response_meta"
	WireRequestContext  = "wire_request"
	WireResponseContext = "wire_response"
	ErrorContext        = "error"
	PanicContext        = "panic"
	DebugContext        = "debug"
	StdoutContext       = "stdout"
	StderrContext       = "stderr"
	ProcessorContext    = "processor"
)

const (
	ReqMetaRemoteAddr          = "remote_addr"
	ReqMetaXForwardedFor       = "x_forwarded_for"
	ReqMetaXForwardedPort      = "x_forwarded_port"
	ReqMetaXForwardedProto     = "x_forwarded_proto"
	ReqMetaCloudFrontPolicy    = "cloudfront_policy"
	ReqMetaCloudFrontSignature = "cloudfront_signature"
	ReqMetaCloudFrontKeyPairId = "cloudfront_key_pair_id"
	ReqMetaHost                = "host"
)

const (
	RspMetaETag        = "etag"
	RspMetaContentType = "content_type"
	RspMetaContent     = "content"
)

type (
	Proccessor   = func(*gin.Context, LocalHandler)
	LocalHandler = func(string, string) (string, error)
)

var methods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}

func (e *Engine) InstallHandlers() {
	e.Use(e.HeaderLink, e.StaticLink, e.PrefixLink)

	e.HandleAllMethods("/", e.OK)
	e.HandleAllMethods("/health-check", e.OK)
	e.HandleAllMethods("/api/*path", e.API)
	e.HandleAllMethods("/_/api/*path", e.Debug, e.API)
	e.HandleAllMethods("/wapi/*path", e.WAPI)
	e.HandleAllMethods("/_/wapi/*path", e.Debug, e.WAPI)
	e.HandleAllMethods("/meta/*path", e.Meta)
	e.HandleAllMethods("/_/meta/*path", e.Debug, e.Meta)
	e.NoRoute(e.PageNotFound)
	e.NoMethod(e.MethodNotAllowed)
}

func (e *Engine) HandleAllMethods(relativePath string, handlers ...gin.HandlerFunc) {
	for _, method := range methods {
		e.Handle(method, relativePath, handlers...)
	}
}

func (e *Engine) HeaderLink(c *gin.Context) {
	for key, prefix := range e.HeaderLinkMap {
		if headerLink, ok := c.Request.Header[key]; ok && len(headerLink) > 0 {
			strs := []string{strings.TrimRight(prefix, "/"), strings.TrimLeft(headerLink[0], "/")}
			c.Request.URL.Path = strings.Join(strs, "/")
			c.Request.Header.Del(key)
			e.HandleContext(c)
			c.Abort()
			return
		}
	}
}

func (e *Engine) StaticLink(c *gin.Context) {
	if dstPath, ok := e.StaticLinkMap[c.Request.URL.Path]; ok {
		c.Request.URL.Path = dstPath
		e.HandleContext(c)
		c.Abort()
		return
	}
}

func (e *Engine) PrefixLink(c *gin.Context) {
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
	c.String(http.StatusOK, "OK")
	c.Abort()
}

func (e *Engine) Debug(c *gin.Context) {
	c.Set(DebugContext, true)
}

func (e *Engine) API(c *gin.Context) {
	// path
	c.Set(PathContext, c.Param("path"))

	// header
	c.Set(HeaderContext, c.Request.Header)

	// meta
	c.Set(RequestMetaContext, e.genReqMeta(c))

	// request
	switch c.Request.Method {
	case http.MethodGet, "": // empty method treated as GET
		c.Set(RequestContext, e.genGetReq(c))
	case http.MethodPost:
		c.Set(RequestContext, e.genPostReq(c))
	default:
		c.Set(RequestContext, "")
	}

	// processor
	if c.GetBool(DebugContext) {
		c.Set(ProcessorContext, e.debugProcessor)
	} else {
		c.Set(ProcessorContext, e.safeProcessor)
	}

	// handle
	if v, ok := c.Get(ProcessorContext); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		c.String(http.StatusInternalServerError, "No processor")
		c.Abort()
		return
	}

	// redirect
	rsp := c.GetString(ResponseContext)
	if strings.HasPrefix(rsp, "http://") || strings.HasPrefix(rsp, "https://") {
		c.Redirect(http.StatusTemporaryRedirect, rsp)
		c.Abort()
		return
	} else if strings.HasPrefix(rsp, "path://") {
		c.Request.URL.Path = "/" + strings.TrimPrefix(rsp, "path://")
		e.HandleContext(c)
		c.Abort()
		return
	} else if strings.HasPrefix(rsp, "error://") {
		c.String(http.StatusInternalServerError, strings.TrimPrefix(rsp, "error://"))
		c.Abort()
		return
	}

	// response
	if c.GetBool(DebugContext) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(PanicContext); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else if v, ok := c.Get(ErrorContext); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else {
		contentType, rspBody := e.parseRspMeta(c)
		c.Data(http.StatusOK, contentType, []byte(rspBody))
		c.Abort()
		return
	}
}

func (e *Engine) WAPI(c *gin.Context) {
	// path
	c.Set(PathContext, c.Param("path"))

	// header
	c.Set(HeaderContext, c.Request.Header)

	// request
	switch c.Request.Method {
	case http.MethodGet, "": // empty method treated as GET
		c.Set(RequestContext, e.genGetReq(c))
	case http.MethodPost:
		c.Set(RequestContext, e.genPostReq(c))
	default:
		c.Set(RequestContext, "")
	}

	// processor
	if c.GetBool(DebugContext) {
		c.Set(ProcessorContext, e.debugWireProcessor)
	} else {
		c.Set(ProcessorContext, e.safeWireProcessor)
	}

	// handle
	if v, ok := c.Get(ProcessorContext); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		c.String(http.StatusInternalServerError, "No processor")
		c.Abort()
		return
	}

	// response
	if c.GetBool(DebugContext) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(PanicContext); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else if v, ok := c.Get(ErrorContext); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else {
		response, err := http.ReadResponse(bufio.NewReader(strings.NewReader(c.GetString(WireResponseContext))), c.Request)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
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
	c.Set(PathContext, c.Param("path"))

	// processor
	if c.GetBool(DebugContext) {
		c.Set(ProcessorContext, e.debugMetaProcessor)
	} else {
		c.Set(ProcessorContext, e.safeMetaProcessor)
	}

	// handle
	if v, ok := c.Get(ProcessorContext); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		c.String(http.StatusInternalServerError, "No processor")
		c.Abort()
		return
	}

	// response
	if c.GetBool(DebugContext) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(PanicContext); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else if v, ok := c.Get(ErrorContext); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else {
		c.String(http.StatusOK, c.GetString(ResponseContext))
		c.Abort()
		return
	}
}

func (e *Engine) PageNotFound(c *gin.Context) {
	if e.PageNotFoundPath != "" {
		c.Request.URL.Path = e.PageNotFoundPath
		e.HandleContext(c)
		c.Abort()
		return
	}
	c.String(404, "404 page not found")
	c.Abort()
}

func (e *Engine) MethodNotAllowed(c *gin.Context) {
	c.String(405, "405 method not allowed")
	c.Abort()
}

func (e *Engine) genReqMeta(c *gin.Context) map[string]any {
	meta := map[string]any{}

	meta[ReqMetaXForwardedFor] = c.Request.Header.Get("X-Forwarded-For")
	meta[ReqMetaXForwardedPort] = c.Request.Header.Get("X-Forwarded-Port")
	meta[ReqMetaXForwardedProto] = c.Request.Header.Get("X-Forwarded-Proto")
	meta[ReqMetaRemoteAddr] = c.Request.RemoteAddr
	meta[ReqMetaCloudFrontPolicy] = c.Request.Header.Get("CloudFront-Policy")
	meta[ReqMetaCloudFrontSignature] = c.Request.Header.Get("CloudFront-Signature")
	meta[ReqMetaCloudFrontKeyPairId] = c.Request.Header.Get("CloudFront-Key-Pair-Id")
	meta[ReqMetaHost] = c.Request.Header.Get("Host")

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
		rsp     string
		wireRsp string
		err     error
	)
	defer func() {
		c.Set(WireRequestContext, wireReq)
		c.Set(ResponseContext, rsp)
		c.Set(WireResponseContext, wireRsp)
		c.Set(ErrorContext, err)
	}()

	path := c.GetString(PathContext)

	var buf bytes.Buffer
	err = c.Request.Write(&buf)
	if err != nil {
		return
	}
	wireReq = buf.String()

	wireRsp, err = e.handle(path, wireReq)
	if err != nil {
		return
	}

	response, err := http.ReadResponse(bufio.NewReader(bytes.NewBufferString(wireRsp)), c.Request)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	rsp = string(data)
}

func (e *Engine) safeWireProcessor(c *gin.Context, f LocalHandler) {
	c.Set(PanicContext, e.doSafe(func() {
		e.doWireProcessor(c, f)
	}))
}

func (e *Engine) debugWireProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doWireProcessor(c, f)
	})
	c.Set(StdoutContext, stdout)
	c.Set(StderrContext, stderr)
	c.Set(PanicContext, panicErr)
}

func (e *Engine) doProcessor(c *gin.Context, f LocalHandler) {
	path := c.GetString(PathContext)
	req := c.GetString(RequestContext)
	reqMeta := c.GetStringMap(RequestMetaContext)
	if gjson.Valid(req) && !gjson.Get(req, "__meta__").Exists() {
		req, _ = sjson.Set(req, "__meta__", reqMeta)
	}
	rsp, err := f(path, req)
	if gjson.Valid(rsp) && gjson.Get(rsp, "__meta__").Exists() {
		rspMeta := make(map[string]any)
		gjson.Get(rsp, "__meta__").ForEach(func(key, value gjson.Result) bool {
			rspMeta[key.String()] = value.Value()
			return true
		})
		c.Set(ResponseMetaContext, rspMeta)
		rsp, _ = sjson.Delete(rsp, "__meta__")
	}
	c.Set(ResponseContext, rsp)
	c.Set(ErrorContext, err)
}

func (e *Engine) safeProcessor(c *gin.Context, f LocalHandler) {
	c.Set(PanicContext, e.doSafe(func() {
		e.doProcessor(c, f)
	}))
}

func (e *Engine) debugProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doProcessor(c, f)
	})
	c.Set(StdoutContext, stdout)
	c.Set(StderrContext, stderr)
	c.Set(PanicContext, panicErr)
}

func (e *Engine) doMetaProcessor(c *gin.Context) {
	path := c.GetString(PathContext)
	rsp, err := e.meta(path)
	c.Set(ResponseContext, rsp)
	c.Set(ErrorContext, err)
}

func (e *Engine) safeMetaProcessor(c *gin.Context, f LocalHandler) {
	c.Set(PanicContext, e.doSafe(func() {
		e.doMetaProcessor(c)
	}))
}

func (e *Engine) debugMetaProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doMetaProcessor(c)
	})
	c.Set(StdoutContext, stdout)
	c.Set(StderrContext, stderr)
	c.Set(PanicContext, panicErr)
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
	return tunnel.Invoke(route, req), nil
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
	buf.WriteString(c.GetString(PathContext))
	buf.WriteString("\n")
	buf.WriteString(`Header: `)
	headerBytes, _ := json.Marshal(c.GetString(HeaderContext))
	buf.WriteString(string(headerBytes))
	buf.WriteString("\n")
	buf.WriteString(`Request Meta: `)
	reqMetaBytes, _ := json.Marshal(c.GetString(RequestMetaContext))
	buf.WriteString(string(reqMetaBytes))
	buf.WriteString("\n")
	buf.WriteString(`Response Meta: `)
	rspMetaBytes, _ := json.Marshal(c.GetStringMap(ResponseMetaContext))
	buf.WriteString(string(rspMetaBytes))
	buf.WriteString("\n")
	buf.WriteString(`Stdout: `)
	buf.WriteString(c.GetString(StdoutContext))
	buf.WriteString("\n")
	buf.WriteString(`Stderr: `)
	buf.WriteString(c.GetString(StderrContext))
	buf.WriteString("\n")
	buf.WriteString(`Error: `)
	if v, ok := c.Get(ErrorContext); ok && v != nil {
		buf.WriteString(v.(error).Error())
	}
	buf.WriteString("\n")
	buf.WriteString(`Panic: `)
	if v, ok := c.Get(PanicContext); ok && v != nil {
		buf.WriteString(v.(error).Error())
	}
	buf.WriteString("\n")
	buf.WriteString(`Request: `)
	buf.WriteString(c.GetString(RequestContext))
	buf.WriteString("\n")
	buf.WriteString(`Response: `)
	buf.WriteString(c.GetString(ResponseContext))
	buf.WriteString("\n")
	buf.WriteString(`Wire Request: `)
	buf.WriteString(c.GetString(WireRequestContext))
	buf.WriteString("\n")
	buf.WriteString(`Wire Response: `)
	buf.WriteString(c.GetString(WireResponseContext))
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

// parseRspMeta applies response meta and sends the response
// Rules:
// - If content_type is empty: default to application/json
// - If content_type has value but content is empty: only override Content-Type header
// - If content has value: override the response body with content
func (e *Engine) parseRspMeta(c *gin.Context) (string, string) {
	respMeta := c.GetStringMap(ResponseMetaContext)
	rspBody := c.GetString(ResponseContext)

	// Default content type
	contentType := "application/json"

	if respMeta != nil {
		// Apply ETag header
		if etag, ok := respMeta[RspMetaETag]; ok && etag != nil && etag != "" {
			c.Header("ETag", fmt.Sprintf("%v", etag))
		}

		// Check content_type
		if ct, ok := respMeta[RspMetaContentType]; ok && ct != nil && ct != "" {
			contentType = fmt.Sprintf("%v", ct)
		}

		// Check content - if has value, override response body
		if content, ok := respMeta[RspMetaContent]; ok && content != nil && content != "" {
			rspBody = fmt.Sprintf("%v", content)
		}
	}

	return contentType, rspBody
}
