package tests

import (
	"testing"
	"time"

	lambdahttp "github.com/aura-studio/lambda/http"
)

func TestHTTPServeAndClose(t *testing.T) {
	addr := freeLocalAddr(t)

	errCh := make(chan error, 1)
	go func() {
		errCh <- lambdahttp.Serve(addr)
	}()

	waitHTTPReady(t, "http://"+addr, 3*time.Second)

	if err := lambdahttp.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve() returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Serve() did not return after Close()")
	}
}

func TestHTTPCloseWithoutServe(t *testing.T) {
	if err := lambdahttp.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
}
