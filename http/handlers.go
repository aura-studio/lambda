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
	GinContextHeader       = "header"
	GinContextPath         = "path"
	GinContextRequest      = "request"
	GinContextResponse     = "response"
	GinContextRequestMeta  = "request_meta"
	GinContextResponseMeta = "response_meta"
	GinContextWireRequest  = "wire_request"
	GinContextWireResponse = "wire_response"
	GinContextError        = "error"
	GinContextPanic        = "panic"
	GinContextDebug        = "debug"
	GinContextStdout       = "stdout"
	GinContextStderr       = "stderr"
	GinContextProcessor    = "processor"
)

const (
	ReqMetaRemoteAddr              = "remote_addr"
	ReqMetaXForwardedFor           = "x_forwarded_for"
	ReqMetaXForwardedPort          = "x_forwarded_port"
	ReqMetaXForwardedProto         = "x_forwarded_proto"
	ReqMetaXForwardedHost          = "x_forwarded_host"
	ReqMetaCloudFrontPolicy        = "cloudfront_policy"
	ReqMetaCloudFrontSignature     = "cloudfront_signature"
	ReqMetaCloudFrontKeyPairId     = "cloudfront_key_pair_id"
	ReqMetaCloudFrontViewerAddress = "cloudfront_viewer_address"
	ReqMetaHost                    = "host"
	ReqMetaRawHost                 = "raw_host"
	ReqMetaPath                    = "path"
	ReqMetaTimestamp               = "timestamp"
	ReqMetaXSign                   = "x_sign"
)

const (
	RspMetaETag        = "etag"
	RspMetaContentType = "content_type"
	RspMetaContent     = "content"
	RspMetaRedirect    = "redirect"
	RspMetaPath        = "path"
	RspMetaError       = "error"
	RspMetaURL         = "url"
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


func (e *Engine) StaticLink(c *gin.Context) {
	path := c.Request.URL.Path
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}
	if dstPath, ok := e.StaticLinkMap[path]; ok {
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
	c.Set(GinContextDebug, true)
}

func (e *Engine) API(c *gin.Context) {
	// path
	c.Set(GinContextPath, c.Param("path"))

	// header
	c.Set(GinContextHeader, c.Request.Header)

	// meta
	c.Set(GinContextRequestMeta, e.genReqMeta(c))

	// request
	switch c.Request.Method {
	case http.MethodGet, "": // empty method treated as GET
		c.Set(GinContextRequest, e.genGetReq(c))
	case http.MethodPost:
		c.Set(GinContextRequest, e.genPostReq(c))
	default:
		c.Set(GinContextRequest, "")
	}

	// processor
	if c.GetBool(GinContextDebug) {
		c.Set(GinContextProcessor, e.debugProcessor)
	} else {
		c.Set(GinContextProcessor, e.safeProcessor)
	}

	// handle
	if v, ok := c.Get(GinContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		c.String(http.StatusInternalServerError, "No processor")
		c.Abort()
		return
	}

	// response
	if c.GetBool(GinContextDebug) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextPanic); ok && v != nil {
		if e.HideErrorMode {
			c.String(http.StatusInternalServerError, "Internal Server Error")
		} else {
			c.String(http.StatusInternalServerError, v.(error).Error())
		}
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextError); ok && v != nil {
		if e.HideErrorMode {
			c.String(http.StatusInternalServerError, "Internal Server Error")
		} else {
			c.String(http.StatusInternalServerError, v.(error).Error())
		}
		c.Abort()
		return
	} else {
		rspMeta := c.GetStringMap(GinContextResponseMeta)
		rspBody := c.GetString(GinContextResponse)
		contentType := "application/json"

		if rspMeta != nil {
			// url meta: parse scheme to determine action
			if u, ok := rspMeta[RspMetaURL]; ok && u != nil && u != "" {
				url := fmt.Sprintf("%v", u)
				if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
					c.Redirect(http.StatusTemporaryRedirect, url)
					c.Abort()
					return
				} else if after, found := strings.CutPrefix(url, "path://"); found {
					c.Request.URL.Path = "/" + strings.TrimLeft(after, "/")
					e.HandleContext(c)
					c.Abort()
					return
				} else if after, found := strings.CutPrefix(url, "error://"); found {
					c.String(http.StatusInternalServerError, after)
					c.Abort()
					return
				}
			}
			if r, ok := rspMeta[RspMetaRedirect]; ok && r != nil && r != "" {
				c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v", r))
				c.Abort()
				return
			}
			if p, ok := rspMeta[RspMetaPath]; ok && p != nil && p != "" {
				c.Request.URL.Path = "/" + strings.TrimLeft(fmt.Sprintf("%v", p), "/")
				e.HandleContext(c)
				c.Abort()
				return
			}
			if e, ok := rspMeta[RspMetaError]; ok && e != nil && e != "" {
				c.String(http.StatusInternalServerError, fmt.Sprintf("%v", e))
				c.Abort()
				return
			}
			if etag, ok := rspMeta[RspMetaETag]; ok && etag != nil && etag != "" {
				c.Header("ETag", fmt.Sprintf("%v", etag))
			}
			if ct, ok := rspMeta[RspMetaContentType]; ok && ct != nil && ct != "" {
				contentType = fmt.Sprintf("%v", ct)
			}
			if content, ok := rspMeta[RspMetaContent]; ok && content != nil && content != "" {
				rspBody = fmt.Sprintf("%v", content)
			}
		}

		c.Data(http.StatusOK, contentType, []byte(rspBody))
		c.Abort()
		return
	}
}

func (e *Engine) WAPI(c *gin.Context) {
	// path
	c.Set(GinContextPath, c.Param("path"))

	// header
	c.Set(GinContextHeader, c.Request.Header)

	// request
	switch c.Request.Method {
	case http.MethodGet, "": // empty method treated as GET
		c.Set(GinContextRequest, e.genGetReq(c))
	case http.MethodPost:
		c.Set(GinContextRequest, e.genPostReq(c))
	default:
		c.Set(GinContextRequest, "")
	}

	// processor
	if c.GetBool(GinContextDebug) {
		c.Set(GinContextProcessor, e.debugWireProcessor)
	} else {
		c.Set(GinContextProcessor, e.safeWireProcessor)
	}

	// handle
	if v, ok := c.Get(GinContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		c.String(http.StatusInternalServerError, "No processor")
		c.Abort()
		return
	}

	// response
	if c.GetBool(GinContextDebug) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextPanic); ok && v != nil {
		if e.HideErrorMode {
			c.String(http.StatusInternalServerError, "Internal Server Error")
		} else {
			c.String(http.StatusInternalServerError, v.(error).Error())
		}
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextError); ok && v != nil {
		if e.HideErrorMode {
			c.String(http.StatusInternalServerError, "Internal Server Error")
		} else {
			c.String(http.StatusInternalServerError, v.(error).Error())
		}
		c.Abort()
		return
	} else {
		response, err := http.ReadResponse(bufio.NewReader(strings.NewReader(c.GetString(GinContextWireResponse))), c.Request)
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
	c.Set(GinContextPath, c.Param("path"))

	// processor
	c.Set(GinContextProcessor, e.safeMetaProcessor)

	// handle
	if v, ok := c.Get(GinContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		c.String(http.StatusInternalServerError, "No processor")
		c.Abort()
		return
	}

	// response
	if v, ok := c.Get(GinContextPanic); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextError); ok && v != nil {
		c.String(http.StatusInternalServerError, v.(error).Error())
		c.Abort()
		return
	} else {
		c.String(http.StatusOK, c.GetString(GinContextResponse))
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
	c.String(404, "404 page not found")
	c.Abort()
}

func (e *Engine) MethodNotAllowed(c *gin.Context) {
	c.String(405, "405 method not allowed")
	c.Abort()
}

func (e *Engine) genReqMeta(c *gin.Context) map[string]any {
	meta := map[string]any{}

	set := func(key string, value any) {
		switch v := value.(type) {
		case string:
			if v != "" {
				meta[key] = v
			}
		case nil:
		default:
			meta[key] = v
		}
	}

	set(ReqMetaXForwardedFor, c.Request.Header.Get("X-Forwarded-For"))
	set(ReqMetaXForwardedPort, c.Request.Header.Get("X-Forwarded-Port"))
	set(ReqMetaXForwardedProto, c.Request.Header.Get("X-Forwarded-Proto"))
	set(ReqMetaXForwardedHost, c.Request.Header.Get("X-Forwarded-Host"))
	set(ReqMetaRemoteAddr, c.Request.RemoteAddr)
	set(ReqMetaCloudFrontPolicy, c.Request.Header.Get("CloudFront-Policy"))
	set(ReqMetaCloudFrontSignature, c.Request.Header.Get("CloudFront-Signature"))
	set(ReqMetaCloudFrontKeyPairId, c.Request.Header.Get("CloudFront-Key-Pair-Id"))
	set(ReqMetaCloudFrontViewerAddress, c.Request.Header.Get("CloudFront-Viewer-Address"))
	set(ReqMetaHost, c.Request.Host)
	set(ReqMetaRawHost, c.Request.Header.Get("Host"))
	set(ReqMetaPath, c.Request.URL.Path)
	set(ReqMetaTimestamp, c.Request.Header.Get("Timestamp"))
	set(ReqMetaXSign, c.Request.Header.Get("X-Sign"))

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
		c.Set(GinContextWireRequest, wireReq)
		c.Set(GinContextResponse, rsp)
		c.Set(GinContextWireResponse, wireRsp)
		c.Set(GinContextError, err)
	}()

	path := c.GetString(GinContextPath)

	var buf bytes.Buffer
	err = c.Request.Write(&buf)
	if err != nil {
		return
	}
	wireReq = buf.String()

	wireRsp, err = f(path, wireReq)
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
	c.Set(GinContextPanic, e.doSafe(func() {
		e.doWireProcessor(c, f)
	}))
}

func (e *Engine) debugWireProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doWireProcessor(c, f)
	})
	c.Set(GinContextStdout, stdout)
	c.Set(GinContextStderr, stderr)
	c.Set(GinContextPanic, panicErr)
}

func (e *Engine) doProcessor(c *gin.Context, f LocalHandler) {
	path := c.GetString(GinContextPath)
	req := c.GetString(GinContextRequest)
	reqMeta := c.GetStringMap(GinContextRequestMeta)
	if gjson.Valid(req) {
		if !gjson.Get(req, "__meta__").Exists() {
			req, _ = sjson.Set(req, "__meta__", reqMeta)
		}
	}
	rsp, err := f(path, req)
	if gjson.Valid(rsp) && gjson.Get(rsp, "__meta__").Exists() {
		rspMeta := make(map[string]any)
		gjson.Get(rsp, "__meta__").ForEach(func(key, value gjson.Result) bool {
			rspMeta[key.String()] = value.Value()
			return true
		})
		c.Set(GinContextResponseMeta, rspMeta)
		rsp, _ = sjson.Delete(rsp, "__meta__")
	}
	c.Set(GinContextResponse, rsp)
	c.Set(GinContextError, err)
}

func (e *Engine) safeProcessor(c *gin.Context, f LocalHandler) {
	c.Set(GinContextPanic, e.doSafe(func() {
		e.doProcessor(c, f)
	}))
}

func (e *Engine) debugProcessor(c *gin.Context, f LocalHandler) {
	stdout, stderr, panicErr := e.doDebug(func() {
		e.doProcessor(c, f)
	})
	c.Set(GinContextStdout, stdout)
	c.Set(GinContextStderr, stderr)
	c.Set(GinContextPanic, panicErr)
}

func (e *Engine) doMetaProcessor(c *gin.Context) {
	path := c.GetString(GinContextPath)
	rsp, err := e.meta(path)
	c.Set(GinContextResponse, rsp)
	c.Set(GinContextError, err)
}

func (e *Engine) safeMetaProcessor(c *gin.Context, f LocalHandler) {
	c.Set(GinContextPanic, e.doSafe(func() {
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
	buf.WriteString(c.GetString(GinContextPath))
	buf.WriteString("\n")
	buf.WriteString(`Header: `)
	headerBytes, _ := json.Marshal(c.GetString(GinContextHeader))
	buf.WriteString(string(headerBytes))
	buf.WriteString("\n")
	buf.WriteString(`Request Meta: `)
	reqMetaBytes, _ := json.Marshal(c.GetString(GinContextRequestMeta))
	buf.WriteString(string(reqMetaBytes))
	buf.WriteString("\n")
	buf.WriteString(`Response Meta: `)
	rspMetaBytes, _ := json.Marshal(c.GetStringMap(GinContextResponseMeta))
	buf.WriteString(string(rspMetaBytes))
	buf.WriteString("\n")
	buf.WriteString(`Stdout: `)
	buf.WriteString(c.GetString(GinContextStdout))
	buf.WriteString("\n")
	buf.WriteString(`Stderr: `)
	buf.WriteString(c.GetString(GinContextStderr))
	buf.WriteString("\n")
	buf.WriteString(`Error: `)
	if v, ok := c.Get(GinContextError); ok && v != nil {
		buf.WriteString(v.(error).Error())
	}
	buf.WriteString("\n")
	buf.WriteString(`Panic: `)
	if v, ok := c.Get(GinContextPanic); ok && v != nil {
		buf.WriteString(v.(error).Error())
	}
	buf.WriteString("\n")
	buf.WriteString(`Request: `)
	buf.WriteString(c.GetString(GinContextRequest))
	buf.WriteString("\n")
	buf.WriteString(`Response: `)
	buf.WriteString(c.GetString(GinContextResponse))
	buf.WriteString("\n")
	buf.WriteString(`Wire Request: `)
	buf.WriteString(c.GetString(GinContextWireRequest))
	buf.WriteString("\n")
	buf.WriteString(`Wire Response: `)
	buf.WriteString(c.GetString(GinContextWireResponse))
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
