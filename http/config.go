package http

import (
	"fmt"
	"os"

	"github.com/aura-studio/lambda/dynamic"
	yaml "gopkg.in/yaml.v2"
)

type yamlHTTPConfig struct {
	DebugMode  bool `yaml:"debugMode"`
	Cors       bool `yaml:"cors"`
	StaticLink []struct {
		SrcPath string `yaml:"srcPath"`
		DstPath string `yaml:"dstPath"`
	} `yaml:"staticLink"`
	PrefixLink []struct {
		SrcPrefix string `yaml:"srcPrefix"`
		DstPrefix string `yaml:"dstPrefix"`
	} `yaml:"prefixLink"`
	HeaderLinkKey []struct {
		Key    string `yaml:"key"`
		Prefix string `yaml:"prefix"`
	} `yaml:"headerLinkKey"`
}

type yamlServeConfig struct {
	HTTP    yamlHTTPConfig `yaml:"http"`
	Dynamic any            `yaml:"dynamic"`
}

func optionFromHTTPConfig(cfg yamlHTTPConfig) Option {
	return HttpOption(func(o *Options) {
		o.DebugMode = cfg.DebugMode
		o.CorsMode = cfg.Cors

		if o.StaticLinkMap == nil {
			o.StaticLinkMap = make(map[string]string)
		}
		if o.PrefixLinkMap == nil {
			o.PrefixLinkMap = make(map[string]string)
		}
		if o.HeaderLinkMap == nil {
			o.HeaderLinkMap = make(map[string]string)
		}

		for _, link := range cfg.StaticLink {
			if link.SrcPath == "" || link.DstPath == "" {
				continue
			}
			o.StaticLinkMap[link.SrcPath] = link.DstPath
		}
		for _, link := range cfg.PrefixLink {
			if link.SrcPrefix == "" || link.DstPrefix == "" {
				continue
			}
			o.PrefixLinkMap[link.SrcPrefix] = link.DstPrefix
		}
		for _, link := range cfg.HeaderLinkKey {
			if link.Key == "" || link.Prefix == "" {
				continue
			}
			o.HeaderLinkMap[link.Key] = link.Prefix
		}
	})
}

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlHTTPConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return optionFromHTTPConfig(cfg), nil
}

// WithConfig parses YAML bytes following http.yml structure and applies it to Options.
// It panics if the YAML is invalid.
func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return HttpOption(func(*Options) {
			panic(fmt.Errorf("http.WithConfig: %w", err))
		})
	}
	return opt
}

// WithConfigFile loads a YAML file and applies it to Options.
// It panics if the file cannot be read or YAML is invalid.
func WithConfigFile(path string) Option {
	b, err := os.ReadFile(path)
	if err != nil {
		return HttpOption(func(*Options) {
			panic(fmt.Errorf("http.WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}

type serveConfigOption struct {
	httpOpt Option
	dynOpt  dynamic.Option
	err     error
}

func (o serveConfigOption) apply(b *serveOptionBag) {
	if o.err != nil {
		panic(fmt.Errorf("http.WithServeConfig: %w", o.err))
	}
	if o.httpOpt != nil {
		b.http = append(b.http, o.httpOpt)
	}
	if o.dynOpt != nil {
		b.dynamic = append(b.dynamic, o.dynOpt)
	}
}

// WithServeConfig parses YAML bytes following http.yml structure, and also supports
// embedding dynamic.yml content under top-level `dynamic:`.
// It panics if the YAML is invalid.
func WithServeConfig(yamlBytes []byte) ServeOption {
	var cfg yamlServeConfig
	if err := yaml.Unmarshal(yamlBytes, &cfg); err != nil {
		return serveConfigOption{err: err}
	}

	httpOpt := optionFromHTTPConfig(cfg.HTTP)

	var dynOpt dynamic.Option
	if cfg.Dynamic != nil {
		b, err := yaml.Marshal(cfg.Dynamic)
		if err != nil {
			return serveConfigOption{err: err}
		}
		// cfg.Dynamic is expected to be a dynamic.yml document root (environment/package).
		dynOpt = dynamic.WithConfig(b)
	}

	return serveConfigOption{httpOpt: httpOpt, dynOpt: dynOpt}
}

// WithServeConfigFile loads a YAML file and applies it as ServeOption.
// It panics if the file cannot be read or YAML is invalid.
func WithServeConfigFile(path string) ServeOption {
	b, err := os.ReadFile(path)
	if err != nil {
		return serveConfigOption{err: fmt.Errorf("http.WithServeConfigFile(%s): %w", path, err)}
	}
	return WithServeConfig(b)
}
