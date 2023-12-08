package httpserver

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

	"github.com/aura-studio/dynamic"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	HeaderContext       = "header"
	PathContext         = "path"
	RequestContext      = "request"
	ResponseContext     = "response"
	MetaContext         = "meta"
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
	MetaRemoteAddr = "remote_addr"
)

type Proccessor = func(*gin.Context, LocalHandler)
type LocalHandler = func(string, string) (string, error)

var methods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}

func (e *Engine) InstallHandlers() {
	e.Use(e.HeaderLink, e.StaticLink, e.PrefixLink)

	e.HandleAllMethods("/", e.OK)
	e.HandleAllMethods("/health-check", e.OK)
	e.HandleAllMethods("/api/*path", e.API)
	e.HandleAllMethods("/_/api/*path", e.Debug, e.API)
	e.HandleAllMethods("/wapi/*path", e.WAPI)
	e.HandleAllMethods("/_/wapi/*path", e.Debug, e.WAPI)
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
	c.Set(MetaContext, e.genMeta(c))

	// request
	if c.Request.Method == http.MethodGet {
		c.Set(RequestContext, e.genGetReq(c))
	} else if c.Request.Method == http.MethodPost {
		c.Set(RequestContext, e.genPostReq(c))
	} else {
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
		c.String(http.StatusOK, c.GetString(ResponseContext))
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
	if c.Request.Method == http.MethodGet {
		c.Set(RequestContext, e.genGetReq(c))
	} else if c.Request.Method == http.MethodPost {
		c.Set(RequestContext, e.genPostReq(c))
	} else {
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

func (e *Engine) PageNotFound(c *gin.Context) {
	c.String(404, "404 page not found")
	c.Abort()
}

func (e *Engine) MethodNotAllowed(c *gin.Context) {
	c.String(405, "405 method not allowed")
	c.Abort()
}

func (e *Engine) genMeta(c *gin.Context) map[string]interface{} {
	meta := map[string]interface{}{}

	xForwardFor := c.Request.Header.Get("X-Forwarded-For")
	if len(xForwardFor) == 0 {
		meta[MetaRemoteAddr] = c.Request.RemoteAddr
	} else {
		meta[MetaRemoteAddr] = strings.Split(xForwardFor, ",")[0] + ":0"
	}

	return meta
}

func (e *Engine) genGetReq(c *gin.Context) string {
	dataMap := map[string]interface{}{}
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

	return
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
	meta := c.GetStringMap(MetaContext)
	if gjson.ValidBytes([]byte(req)) && !gjson.Get(req, "__meta__").Exists() {
		req, _ = sjson.Set(req, "__meta__", meta)
	}
	rsp, err := f(path, req)
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

func (e *Engine) handle(path string, req string) (string, error) {
	strs := strings.Split(strings.Trim(path, "/"), "/")
	packageName := strs[0]
	commit := strs[1]

	tunnel, err := dynamic.GetPackage(packageName, commit)
	if err != nil {
		return "", err
	}

	route := fmt.Sprintf("/%s", strings.Join(strs[2:], "/"))
	return tunnel.Invoke(route, req), nil
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
	buf.WriteString(`Meta: `)
	metaBytes, _ := json.Marshal(c.GetString(MetaContext))
	buf.WriteString(string(metaBytes))
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
