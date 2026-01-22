package dynamic

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
	Environment struct {
		Toolchain struct {
			OS       string `yaml:"os"`
			Arch     string `yaml:"arch"`
			Compiler string `yaml:"compiler"`
			Variant  string `yaml:"variant"`
		} `yaml:"toolchain"`
		Warehouse struct {
			Local  string `yaml:"local"`
			Remote string `yaml:"remote"`
		} `yaml:"warehouse"`
	} `yaml:"environment"`

	Package struct {
		Namespace      string `yaml:"namespace"`
		DefaultVersion string `yaml:"defaultVersion"`
		Preload        []struct {
			Package string `yaml:"package"`
			Version string `yaml:"version"`
		} `yaml:"preload"`
	} `yaml:"package"`
}

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return OptionFunc(func(o *Options) {
		o.Os = cfg.Environment.Toolchain.OS
		o.Arch = cfg.Environment.Toolchain.Arch
		o.Compiler = cfg.Environment.Toolchain.Compiler
		o.Variant = cfg.Environment.Toolchain.Variant

		o.LocalWarehouse = cfg.Environment.Warehouse.Local
		o.RemoteWarehouse = cfg.Environment.Warehouse.Remote

		o.PackageNamespace = cfg.Package.Namespace
		o.PackageDefaultVersion = cfg.Package.DefaultVersion

		for _, p := range cfg.Package.Preload {
			if p.Package == "" {
				continue
			}
			o.PreloadPackages = append(o.PreloadPackages, &Package{Package: p.Package, Version: p.Version})
		}
	}), nil
}

// WithConfig parses YAML bytes following dynamic.yml structure and applies it to Options.
// It panics if the YAML is invalid.
func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return OptionFunc(func(*Options) {
			   panic(fmt.Errorf("dynamic: WithConfig: %w", err))
		})
	}
	return opt
}

// WithConfigFile loads a YAML file and applies it to Options.
// It panics if the file cannot be read or YAML is invalid.
func WithConfigFile(path string) Option {
	b, err := os.ReadFile(path)
	if err != nil {
		return OptionFunc(func(*Options) {
			   panic(fmt.Errorf("dynamic: WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}
