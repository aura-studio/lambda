package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aura-studio/dynamic"
	"github.com/gin-gonic/gin"
)

type Proccessor func(*gin.Context, LocalHandler)
type LocalHandler func(string, string) (string, error)

var (
	Handlers = []gin.HandlerFunc{
		OK,
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
		}
	}

	Path = func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/__debug__") {
			c.Set(DebugContext, true)
			c.Set(PathContext, strings.TrimPrefix(c.Request.URL.Path, "/__debug__"))
		} else {
			c.Set(DebugContext, false)
			c.Set(PathContext, c.Request.URL.Path)
		}
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
			v.(Proccessor)(c, do)
		} else {
			c.String(http.StatusInternalServerError, "No processor")
		}
	}

	Redirect = func(c *gin.Context) {
		rsp := c.GetString(ResponseContext)
		if strings.HasPrefix(rsp, "http") {
			c.Redirect(http.StatusTemporaryRedirect, rsp)
		}
	}

	Response = func(c *gin.Context) {
		if c.GetBool(DebugContext) {
			c.String(http.StatusOK, debugFormat,
				c.GetString(StdoutContext),
				c.GetString(StderrContext),
				c.GetString(ErrorContext),
				c.GetString(ResponseContext),
			)
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

		stdout, stderr, p := doDebug(doFunc)
		c.Set(StdoutContext, stdout)
		c.Set(StderrContext, stderr)
		c.Set(PanicContext, p)
	}
}

func do(path string, req string) (string, error) {
	strs := strings.Split(strings.Trim(path, "/"), "/")
	name := strings.Join(strs[:2], "_")
	route := fmt.Sprintf("/%s", strings.Join(strs[2:], "/"))
	tunnel, err := dynamic.GetTunnel(name)
	if err != nil {
		return "", err
	}

	return tunnel.Invoke(route, req), nil
}
