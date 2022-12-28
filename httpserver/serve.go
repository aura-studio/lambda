package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/aura-studio/dynamic"
	"github.com/gin-gonic/gin"
)

var srv *http.Server

func Serve(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/*path", Handlers...)

	r.POST("/*path", Handlers...)

	srv = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func Register(name string, tunnel dynamic.Tunnel) {
	dynamic.RegisterTunnel(name, tunnel)
}
