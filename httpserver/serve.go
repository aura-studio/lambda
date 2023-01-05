package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aura-studio/dynamic"
	"github.com/gin-gonic/gin"
)

var srv *http.Server

func Serve(addr string, opts ...Option) {
	options.init(opts...)

	r := gin.Default()

	r.GET("/*path", Handlers...)

	r.POST("/*path", Handlers...)

	srv = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	defer cancel()
}

func Register(name string, tunnel dynamic.Tunnel) {
	dynamic.RegisterTunnel(name, tunnel)
}
