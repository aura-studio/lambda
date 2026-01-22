package dynamic

import (
	"log"

	"github.com/aura-studio/dynamic"
)

type Package struct {
	Package string
	Version string
	Tunnel  dynamic.Tunnel
}

type Dynamic struct {
	*Options
}

func NewDynamic(opts ...Option) *Dynamic {
	d := &Dynamic{
		Options: NewOptions(opts...),
	}

	d.InstallPackages()

	return d
}

func (e *Dynamic) InstallPackages() {
	if e.Os != "" {
		dynamic.DynamicOS = e.Os
	}
	if e.Arch != "" {
		dynamic.DynamicArch = e.Arch
	}
	if e.Compiler != "" {
		dynamic.DynamicCompiler = e.Compiler
	}
	if e.Variant != "" {
		dynamic.DynamicVariant = e.Variant
	}

	dynamic.UseWarehouse(e.LocalWarehouse, e.RemoteWarehouse)

	if e.PackageNamespace != "" {
		dynamic.UseNamespace(e.PackageNamespace)
	}

	if e.PackageDefaultVersion != "" {
		dynamic.UseDefaultVersion(e.PackageDefaultVersion)
	}

	for _, p := range e.StaticPackages {
		dynamic.RegisterPackage(p.Package, p.Version, p.Tunnel)
	}

	for _, p := range e.PreloadPackages {
		if _, err := dynamic.GetPackage(p.Package, p.Version); err != nil {
			   log.Printf("[lambda] preload package %s_%s_%s failed: dynamic: %v", e.PackageNamespace, p.Package, p.Version, err)
		}
	}
}

func (e *Dynamic) GetPackage(pkg string, version string) (dynamic.Tunnel, error) {
	return dynamic.GetPackage(pkg, version)
}
