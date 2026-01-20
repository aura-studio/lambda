package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aura-studio/lambda/dynamic"
)

func TestDynamicWithDefaultConfigFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "dynamic.yml")
	if err := os.WriteFile(p, []byte(`environment:
  toolchain:
    os: ubuntu24.04
    arch: amd64v1
    compiler: go1.25.5
    variant: generic
  warehouse:
    local:
    remote:
package:
  namespace: myns
  defaultVersion: v1
  preload: []
`), 0o644); err != nil {
		t.Fatalf("write dynamic.yaml: %v", err)
	}

	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	o := dynamic.NewOptions(dynamic.WithDefaultConfigFile())
	if o.Os != "ubuntu24.04" {
		t.Fatalf("Os = %q", o.Os)
	}
	if o.PackageNamespace != "myns" {
		t.Fatalf("PackageNamespace = %q", o.PackageNamespace)
	}
}
