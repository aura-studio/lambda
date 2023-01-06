package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"
)

var srv *http.Server

func Serve(addr string, opts ...Option) {
	srv = &http.Server{
		Addr:    addr,
		Handler: NewEngine(opts...),
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
