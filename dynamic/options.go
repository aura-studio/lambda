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
