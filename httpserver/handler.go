package httpserver

import (
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
)

type Proccessor = func(*gin.Context, LocalHandler)
type LocalHandler = func(string, string) (string, error)

func (e *Engine) InstallHandlers() {
	e.Use(e.HeaderLink, e.StaticLink, e.PrefixLink)

	e.GET("/", e.OK)
	e.POST("/", e.OK)

	e.GET("/health-check", e.OK)
	e.POST("/health-check", e.OK)

	e.GET("/api/*path", e.API)
	e.POST("/api/*path", e.API)

	e.GET("/_/api/*path", e.Debug, e.API)
	e.POST("/_/api/*path", e.Debug, e.API)

	e.NoRoute(e.PageNotFound)
	e.NoMethod(e.MethodNotAllowed)
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

	// request
	if c.Request.Method == http.MethodGet {
		c.Set(RequestContext, e.genGetReq(c))
	} else if c.Request.Method == http.MethodPost {
		c.Set(RequestContext, e.genPostReq(c))
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
	}

	// response
	if c.GetBool(DebugContext) {
		c.String(http.StatusOK, e.formatDebug(c))
		c.Abort()
		return
	} else {
		if v, ok := c.Get(ErrorContext); ok && v != nil {
			c.String(http.StatusOK, v.(error).Error())
			c.Abort()
			return
		} else {
			c.String(http.StatusOK, c.GetString(ResponseContext))
			c.Abort()
			return
		}
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

	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))

	return string(data)
}

func (e *Engine) safeProcessor(c *gin.Context, f LocalHandler) {
	doFunc := func() {
		path := c.GetString(PathContext)
		req := c.GetString(RequestContext)
		rsp, err := f(path, req)
		c.Set(ResponseContext, rsp)
		c.Set(ErrorContext, err)
	}

	c.Set(PanicContext, e.doSafe(doFunc))
}

func (e *Engine) debugProcessor(c *gin.Context, f LocalHandler) {
	doFunc := func() {
		path := c.GetString(PathContext)
		req := c.GetString(RequestContext)
		rsp, err := f(path, req)
		c.Set(ResponseContext, rsp)
		c.Set(ErrorContext, err)
	}

	stdout, stderr, panicErr := e.doDebug(doFunc)
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
