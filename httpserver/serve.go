package httpserver

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Serve() {
	r := gin.Default()

	r.GET("/*path", Handlers...)

	r.POST("/*path", Handlers...)

	// listen and serve on 0.0.0.0:8080
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
