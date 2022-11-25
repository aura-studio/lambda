package engine

import (
	"encoding/json"
	"io"
	"log"

	"github.com/aura-studio/lambda/boost/cast"
	"github.com/gin-gonic/gin"
)

type Engine struct{}

func (engine *Engine) ServeHTTP(handler func(string, string) string) {
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
				log.Printf("Recovered from panic: %s", cast.ToError(v).Error())
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
				log.Printf("Recovered from panic: %s", cast.ToError(v).Error())
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
