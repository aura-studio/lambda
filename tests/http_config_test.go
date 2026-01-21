package tests

import (
	"testing"

	lambdahttp "github.com/aura-studio/lambda/http"
)

func TestHTTPWithConfig(t *testing.T) {
	yaml := []byte(`debug: true
cors: true
staticLink:
  - srcPath: /a
    dstPath: /b
prefixLink:
  - srcPrefix: /api
    dstPrefix: /v1
headerLinkKey:
  - key: X-Rewrite
    prefix: /p
`)

	o := lambdahttp.NewOptions(lambdahttp.WithConfig(yaml))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false")
	}
	if !o.CorsMode {
		t.Fatalf("CorsMode = false")
	}
	if got := o.StaticLinkMap["/a"]; got != "/b" {
		t.Fatalf("StaticLinkMap['/a'] = %q", got)
	}
	if got := o.PrefixLinkMap["/api"]; got != "/v1" {
		t.Fatalf("PrefixLinkMap['/api'] = %q", got)
	}
	if got := o.HeaderLinkMap["X-Rewrite"]; got != "/p" {
		t.Fatalf("HeaderLinkMap['X-Rewrite'] = %q", got)
	}
}
