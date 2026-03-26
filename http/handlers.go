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
)

const (
	GinContextHeader    = "header"
	GinContextPath      = "path"
	GinContextRequest   = "request"
	GinContextResponse  = "response"
	GinContextError     = "error"
	GinContextPanic     = "panic"
	GinContextDebug     = "debug"
	GinContextStdout    = "stdout"
	GinContextStderr    = "stderr"
	GinContextProcessor = "processor"
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
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
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
		log.Println("processor not found")
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	}

	// response
	if c.GetBool(GinContextDebug) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextPanic); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextError); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else {
		c.Data(http.StatusOK, "application/json", []byte(c.GetString(GinContextResponse)))
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
		log.Println("processor not found")
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	}

	// response
	if c.GetBool(GinContextDebug) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextPanic); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextError); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else {
		response, err := http.ReadResponse(bufio.NewReader(strings.NewReader(c.GetString(GinContextResponse))), c.Request)
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
	c.Set(GinContextPath, c.Param("path"))

	// processor
	c.Set(GinContextProcessor, e.safeMetaProcessor)

	// handle
	if v, ok := c.Get(GinContextProcessor); ok {
		v.(Proccessor)(c, e.handle)
	} else {
		log.Println("processor not found")
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	}

	// response
	if v, ok := c.Get(GinContextPanic); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Abort()
		return
	} else if v, ok := c.Get(GinContextError); ok && v != nil {
		log.Println(v.(error).Error())
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
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
	c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	c.Abort()
}

func (e *Engine) MethodNotAllowed(c *gin.Context) {
	c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	c.Abort()
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
		c.Set(GinContextRequest, wireReq)
		c.Set(GinContextResponse, wireRsp)
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
	rsp, err := f(path, req)
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
