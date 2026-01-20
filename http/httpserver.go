package http

import (
	"context"
	"net/http"
	"time"
)

var srv *http.Server

func Serve(addr string, opts ...ServeOption) error {
	srv = &http.Server{
		Addr:    addr,
		Handler: NewEngine(opts...),
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if srv == nil {
		return nil
	}
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
