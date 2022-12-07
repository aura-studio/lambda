package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"strings"

	"github.com/aura-studio/dynamic"
	"github.com/gin-gonic/gin"
)

func ServeHTTP() {
	r := gin.Default()

	r.GET("/*path", func(c *gin.Context) {
		var dataMap = map[string]interface{}{}
		for k, v := range c.Request.URL.Query() {
			dataMap[k] = v[0]
		}
		data, err := json.Marshal(dataMap)
		if err != nil {
			log.Printf("Parse queries error: %v", err)
			c.String(200, err.Error())
			return
		}

		process(c, string(data))
	})

	r.POST("/*path", func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
		}

		process(c, string(data))
	})

	// listen and serve on 0.0.0.0:8080
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

func process(c *gin.Context, req string) {
	defer func() {
		if v := recover(); v != nil {
			log.Printf("Recovered from panic: %s\n%s", v.(error).Error(), debug.Stack())
			c.String(200, "Internal Server Error")
		}
	}()

	if c.Request.URL.Path == "/" {
		c.String(200, "OK")
		return
	}

	var (
		rsp string
		err error
	)
	if strings.HasPrefix(c.Request.URL.Path, "/__debug__") {
		stdout, stderr, debugErr := doDebug(func() {
			rsp, err = handler(strings.TrimPrefix(c.Request.URL.Path, "/__debug__"), req)
		})
		if strings.HasPrefix(rsp, "http") {
			c.Redirect(302, rsp)
			return
		}
		rsp = fmt.Sprintf("Stdout: %s\nStderr: %s\nDebug Error: %s\nHandler Error: %s\nResponse Body: %s", stdout, stderr, debugErr.Error(), err.Error(), rsp)
		c.String(200, rsp)
		return
	} else {
		rsp, err = handler(c.Request.URL.Path, req)
		if err != nil {
			c.String(200, err.Error())
			return
		}
		if strings.HasPrefix(rsp, "http") {
			c.Redirect(302, rsp)
			return
		}
		c.String(200, rsp)
		return
	}
}

func handler(path string, req string) (string, error) {
	strs := strings.Split(strings.Trim(path, "/"), "/")
	name := strings.Join(strs[:2], "_")
	route := fmt.Sprintf("/%s", strings.Join(strs[2:], "/"))
	tunnel, err := dynamic.GetTunnel(name)
	if err != nil {
		return "", err
	}

	return tunnel.Invoke(route, req), nil
}
