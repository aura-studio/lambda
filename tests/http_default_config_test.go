package tests

import (
	"os"
	"path/filepath"
	"testing"

	lambdahttp "github.com/aura-studio/lambda/http"
)

func TestHTTPWithDefaultConfig(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "http.yaml")
	if err := os.WriteFile(p, []byte(`http:
  release: true
  cors: true
  staticLink: []
  prefixLink: []
  headerLinkKey: []
`), 0o644); err != nil {
		t.Fatalf("write http.yaml: %v", err)
	}

	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	o := lambdahttp.NewOptions(lambdahttp.WithDefaultConfig())
	if !o.ReleaseMode {
		t.Fatalf("ReleaseMode = false")
	}
	if !o.CorsMode {
		t.Fatalf("CorsMode = false")
	}
}
