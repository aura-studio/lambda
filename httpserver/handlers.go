package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aura-studio/dynamic"
	"github.com/gin-gonic/gin"
)

type Proccessor = func(*gin.Context, LocalHandler)
type LocalHandler = func(string, string) (string, error)

var (
	Handlers = []gin.HandlerFunc{
		Path,
		Request,
		Processor,
		Handle,
		Redirect,
		Response,
	}
)

var (
	OK = func(c *gin.Context) {
		if c.Request.URL.Path == "/" {
			c.String(http.StatusOK, "OK")
		} else if c.Request.URL.Path == "/health-check" {
			c.String(http.StatusOK, "OK")
		}
	}

	Path = func(c *gin.Context) {
		path := strings.TrimSuffix(c.Request.URL.Path, "/")
		if strings.HasPrefix(path, "/debug") {
			c.Set(DebugContext, true)
			path = strings.TrimPrefix(path, "/debug")
		} else {
			c.Set(DebugContext, false)
		}
		c.Set(PathContext, path)
	}

	Request = func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Set(RequestContext, genGetReq(c))
		} else if c.Request.Method == http.MethodPost {
			c.Set(RequestContext, genPostReq(c))
		}
	}

	Processor = func(c *gin.Context) {
		if c.GetBool(DebugContext) {
			c.Set(ProcessorContext, genDebugProcessor(c))
		} else {
			c.Set(ProcessorContext, genSafeProcessor(c))
		}
	}

	Handle = func(c *gin.Context) {
		if v, ok := c.Get(ProcessorContext); ok {
			v.(Proccessor)(c, handle)
		} else {
			c.String(http.StatusInternalServerError, "No processor")
		}
	}

	Redirect = func(c *gin.Context) {
		rsp := c.GetString(ResponseContext)
		if strings.HasPrefix(rsp, "http://") || strings.HasPrefix(rsp, "https://") {
			c.Redirect(http.StatusTemporaryRedirect, rsp)
		} else if strings.HasPrefix(rsp, "path://") {
			c.Request.URL.Path = "/" + strings.TrimPrefix(rsp, "path://")
			r.HandleContext(c)
		}
	}

	Response = func(c *gin.Context) {
		if c.GetBool(DebugContext) {
			c.String(http.StatusOK, formatDebug(c))
		} else {
			if v, ok := c.Get(ErrorContext); ok && v != nil {
				c.String(http.StatusOK, v.(error).Error())
			} else {
				c.String(http.StatusOK, c.GetString(ResponseContext))
			}
		}
	}
)

func genGetReq(c *gin.Context) string {
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

func genPostReq(c *gin.Context) string {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}

func genSafeProcessor(c *gin.Context) func(c *gin.Context, f LocalHandler) {
	return func(c *gin.Context, f LocalHandler) {
		doFunc := func() {
			path := c.GetString(PathContext)
			req := c.GetString(RequestContext)
			rsp, err := f(path, req)
			c.Set(ResponseContext, rsp)
			c.Set(ErrorContext, err)
		}

		c.Set(PanicContext, doSafe(doFunc))
	}
}

func genDebugProcessor(c *gin.Context) func(*gin.Context, LocalHandler) {
	return func(c *gin.Context, f LocalHandler) {
		doFunc := func() {
			path := c.GetString(PathContext)
			req := c.GetString(RequestContext)
			rsp, err := f(path, req)
			c.Set(ResponseContext, rsp)
			c.Set(ErrorContext, err)
		}

		stdout, stderr, panicErr := doDebug(doFunc)
		c.Set(StdoutContext, stdout)
		c.Set(StderrContext, stderr)
		c.Set(PanicContext, panicErr)
	}
}

func handle(path string, req string) (string, error) {
	strs := strings.Split(strings.Trim(path, "/"), "/")
	name := strings.Join(strs[:2], "_")
	if len(options.namespace) > 0 {
		name = fmt.Sprintf("%s_%s", options.namespace, name)
	}
	route := fmt.Sprintf("/%s", strings.Join(strs[2:], "/"))
	tunnel, err := dynamic.GetTunnel(name)
	if err != nil {
		return "", err
	}

	return tunnel.Invoke(route, req), nil
}

func formatDebug(c *gin.Context) string {
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
