package tests

import (
	"testing"

	lambdahttp "github.com/aura-studio/lambda/http"
	"github.com/aura-studio/lambda/server"
)

func TestHTTPWithServeConfig_EmbeddedDynamic(t *testing.T) {
	yaml := []byte(
		"lambda: http\n" +
			"http:\n" +
			"  address: \":8080\"\n" +
			"  mode:\n" +
			"    debug: true\n" +
			"    cors: true\n" +
			"\n" +
			"dynamic:\n" +
			"  environment:\n" +
			"    toolchain:\n" +
			"      os: ubuntu24.04\n" +
			"      arch: amd64v1\n" +
			"      compiler: go1.25.5\n" +
			"      variant: generic\n" +
			"    warehouse:\n" +
			"      local:\n" +
			"      remote:\n" +
			"  package:\n" +
			"    namespace: myns\n" +
			"    defaultVersion: v1\n" +
			"    preload:\n" +
			"      - package: foo\n" +
			"        version: v2\n",
	)

	opt := server.WithServeConfig(yaml)
	options := &server.Options{}
	opt.Apply(options)
	e := lambdahttp.NewEngine(options.Http, options.Dynamic)
	if !e.DebugMode {
		t.Fatalf("DebugMode = false")
	}
	if !e.CorsMode {
		t.Fatalf("CorsMode = false")
	}
	if e.Os != "ubuntu24.04" {
		t.Fatalf("Os = %q", e.Os)
	}
	if e.PackageNamespace != "myns" {
		t.Fatalf("PackageNamespace = %q", e.PackageNamespace)
	}
	if len(e.PreloadPackages) != 1 {
		t.Fatalf("PreloadPackages len = %d", len(e.PreloadPackages))
	}
}
