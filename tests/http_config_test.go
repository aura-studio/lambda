package tests

import (
	"testing"

	lambdahttp "github.com/aura-studio/lambda/http"
)

func TestHTTPWithConfig(t *testing.T) {
	yaml := []byte(`mode:
  debug: true
  cors: true
staticLink:
  - srcPath: /a
    dstPath: /b
prefixLink:
  - srcPrefix: /api
    dstPrefix: /v1
`)

	o := lambdahttp.NewOptions(lambdahttp.WithConfig(yaml))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false")
	}
	if !o.CorsMode {
		t.Fatalf("CorsMode = false")
	}
	if got := o.StaticLinkMap["/a"]; got.Dst != "/b" {
		t.Fatalf("StaticLinkMap['/a'].Dst = %q", got.Dst)
	}
	if got := o.PrefixLinkMap["/api"]; got.Dst != "/v1" {
		t.Fatalf("PrefixLinkMap['/api'].Dst = %q", got.Dst)
	}
}
