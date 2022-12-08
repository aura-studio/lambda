package httpserver

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Serve(addrs ...string) {
	r := gin.Default()

	r.GET("/*path", Handlers...)

	r.POST("/*path", Handlers...)

	if err := r.Run(addrs...); err != nil {
		log.Fatal(err)
	}
}
