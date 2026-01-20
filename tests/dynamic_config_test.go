package tests

import (
	"testing"

	"github.com/aura-studio/lambda/dynamic"
)

func TestDynamicWithConfig(t *testing.T) {
	yaml := []byte(`environment:
  toolchain:
    os: ubuntu24.04
    arch: amd64v1
    compiler: go1.25.5
    variant: generic
  warehouse:
    local: /tmp/local
    remote: s3://bucket

package:
  namespace: myns
  defaultVersion: v1
  preload:
    - package: foo
      version: v2
    - package: bar
      version: ''
`)

	o := dynamic.NewOptions(dynamic.WithConfig(yaml))
	if o.Os != "ubuntu24.04" {
		t.Fatalf("Os = %q", o.Os)
	}
	if o.Arch != "amd64v1" {
		t.Fatalf("Arch = %q", o.Arch)
	}
	if o.Compiler != "go1.25.5" {
		t.Fatalf("Compiler = %q", o.Compiler)
	}
	if o.Variant != "generic" {
		t.Fatalf("Variant = %q", o.Variant)
	}
	if o.LocalWarehouse != "/tmp/local" {
		t.Fatalf("LocalWarehouse = %q", o.LocalWarehouse)
	}
	if o.RemoteWarehouse != "s3://bucket" {
		t.Fatalf("RemoteWarehouse = %q", o.RemoteWarehouse)
	}
	if o.PackageNamespace != "myns" {
		t.Fatalf("PackageNamespace = %q", o.PackageNamespace)
	}
	if o.PackageDefaultVersion != "v1" {
		t.Fatalf("PackageDefaultVersion = %q", o.PackageDefaultVersion)
	}
	if len(o.PreloadPackages) != 2 {
		t.Fatalf("PreloadPackages len = %d", len(o.PreloadPackages))
	}
	if o.PreloadPackages[0].Package != "foo" || o.PreloadPackages[0].Version != "v2" {
		t.Fatalf("preload[0] = %+v", o.PreloadPackages[0])
	}
	if o.PreloadPackages[1].Package != "bar" {
		t.Fatalf("preload[1] = %+v", o.PreloadPackages[1])
	}
}
