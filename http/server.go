package http

import (
	"context"
	"net/http"
	"time"

	"github.com/aura-studio/lambda/dynamic"
)

var srv *http.Server

func Serve(httpOpts []Option, dynamicOpts []dynamic.Option) error {
	opts := NewOptions(httpOpts...)
	srv = &http.Server{
		Addr:    opts.Address,
		Handler: NewEngine(httpOpts, dynamicOpts),
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
