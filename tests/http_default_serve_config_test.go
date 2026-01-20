package tests

import (
	"os"
	"path/filepath"
	"testing"

	lambdahttp "github.com/aura-studio/lambda/http"
)

func TestHTTPWithDefaultServeConfigFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "serve.yml")
	if err := os.WriteFile(p, []byte(
		"http:\n"+
			"  release: true\n"+
			"  cors: true\n"+
			"\n"+
			"dynamic:\n"+
			"  environment:\n"+
			"    toolchain:\n"+
			"      os: ubuntu24.04\n"+
			"      arch: amd64v1\n"+
			"      compiler: go1.25.5\n"+
			"      variant: generic\n"+
			"    warehouse:\n"+
			"      local:\n"+
			"      remote:\n"+
			"  package:\n"+
			"    namespace: myns\n"+
			"    defaultVersion: v1\n"+
			"    preload: []\n",
	), 0o644); err != nil {
		t.Fatalf("write http.yaml: %v", err)
	}

	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	e := lambdahttp.NewEngine(lambdahttp.WithDefaultServeConfigFile())
	if !e.ReleaseMode {
		t.Logf("Options: %+v", e.Options)
		t.Fatalf("ReleaseMode = false")
	}
	if e.Os != "ubuntu24.04" {
		t.Fatalf("Os = %q", e.Os)
	}
}
