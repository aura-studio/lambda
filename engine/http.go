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
	var handler = func(path string, req string) string {
		strs := strings.Split(path, "/")[1:]
		name := strings.Join(strs[:2], "_")
		route := fmt.Sprintf("/%s", strings.Join(strs[2:], "/"))

		tunnel, err := dynamic.GetTunnel(name)
		if err != nil {
			panic(err)
		}

		return tunnel.Invoke(route, req)
	}

	r := gin.Default()
	r.GET("/*path", func(c *gin.Context) {
		var dataMap = map[string]interface{}{}
		for k, v := range c.Request.URL.Query() {
			dataMap[k] = v[0]
		}
		data, err := json.Marshal(dataMap)
		if err != nil {
			log.Printf("Parse queries error: %v", err)
			c.String(500, err.Error())
			return
		}

		var rsp string
		defer func() {
			if v := recover(); v != nil {
				log.Printf("Recovered from panic: %s\n%s", v.(error).Error(), debug.Stack())
				c.String(500, "Internal Server Error")
			} else {
				c.String(200, rsp)
			}
		}()

		rsp = handler(c.Request.URL.Path, string(data))
	})

	r.POST("/*path", func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Fatal(err)
		}

		var rsp string
		defer func() {
			if v := recover(); v != nil {
				log.Printf("Recovered from panic: %s\n%s", v.(error).Error(), debug.Stack())
				c.String(500, "Internal Server Error")
			} else {
				c.String(200, rsp)
			}
		}()

		rsp = handler(c.Request.URL.Path, string(data))
	})

	// listen and serve on 0.0.0.0:8080
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
