package tests

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func freeLocalAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for free port: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func waitHTTPReady(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	client := &http.Client{Timeout: 250 * time.Millisecond}
	var lastErr error
	for {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health-check", nil)
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return
			}
			lastErr = errors.New("unexpected status")
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			t.Fatalf("server not ready within %s (last err: %v)", timeout, lastErr)
		case <-ticker.C:
		}
	}
}
