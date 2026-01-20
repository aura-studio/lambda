package dynamic

import (
	"github.com/mohae/deepcopy"
)

type Option interface {
	Apply(o *Options)
}

type OptionFunc func(*Options)

func (f OptionFunc) Apply(o *Options) { f(o) }

type Options struct {
	// Dynamic Options
	Os                    string
	Arch                  string
	Compiler              string
	Variant               string
	LocalWarehouse        string
	RemoteWarehouse       string
	PackageNamespace      string
	PackageDefaultVersion string
	StaticPackages        []*Package
	PreloadPackages       []*Package
}

var defaultOptions = &Options{
	Os:                    "",
	Arch:                  "",
	Compiler:              "",
	Variant:               "",
	LocalWarehouse:        "",
	RemoteWarehouse:       "",
	PackageNamespace:      "",
	PackageDefaultVersion: "",
	StaticPackages:        []*Package{},
	PreloadPackages:       []*Package{},
}

func NewOptions(opts ...Option) *Options {
	options := deepcopy.Copy(defaultOptions).(*Options)
	options.init(opts...)
	return options
}

func (o *Options) init(opts ...Option) {
	for _, opt := range opts {
		if opt != nil {
			opt.Apply(o)
		}
	}
}

// -------------- Dynamic Options ----------------
func WithOs(os string) Option {
	return OptionFunc(func(o *Options) {
		o.Os = os
	})
}

func WithArch(arch string) Option {
	return OptionFunc(func(o *Options) {
		o.Arch = arch
	})
}

func WithCompiler(compiler string) Option {
	return OptionFunc(func(o *Options) {
		o.Compiler = compiler
	})
}

func WithVariant(variant string) Option {
	return OptionFunc(func(o *Options) {
		o.Variant = variant
	})
}

func WithLocalWarehouse(localWarehouse string) Option {
	return OptionFunc(func(o *Options) {
		o.LocalWarehouse = localWarehouse
	})
}

func WithRemoteWarehouse(remoteWarehouse string) Option {
	return OptionFunc(func(o *Options) {
		o.RemoteWarehouse = remoteWarehouse
	})
}

func WithPackageNamespace(packageNamespace string) Option {
	return OptionFunc(func(o *Options) {
		o.PackageNamespace = packageNamespace
	})
}

func WithPackageDefaultVersion(packageDefaultVersion string) Option {
	return OptionFunc(func(o *Options) {
		o.PackageDefaultVersion = packageDefaultVersion
	})
}

func WithStaticPackage(p *Package) Option {
	return OptionFunc(func(o *Options) {
		o.StaticPackages = append(o.StaticPackages, p)
	})
}

func WithPreloadPackage(p *Package) Option {
	return OptionFunc(func(o *Options) {
		o.PreloadPackages = append(o.PreloadPackages, p)
	})
}
