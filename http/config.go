package http

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlHTTPConfig struct {
	Address string `yaml:"address"`
	Mode    struct {
		Debug     bool `yaml:"debug"`
		Cors      bool `yaml:"cors"`
	} `yaml:"mode"`
	StaticLink []struct {
		SrcPath string   `yaml:"srcPath"`
		DstPath string   `yaml:"dstPath"`
		Methods []string `yaml:"methods"`
	} `yaml:"staticLink"`
	PrefixLink []struct {
		SrcPrefix string   `yaml:"srcPrefix"`
		DstPrefix string   `yaml:"dstPrefix"`
		Methods   []string `yaml:"methods"`
	} `yaml:"prefixLink"`
	PageNotFound []struct {
		Path    string   `yaml:"path"`
		Methods []string `yaml:"methods"`
	} `yaml:"pageNotFound"`
}

func optionFromHTTPConfig(cfg yamlHTTPConfig) Option {
	return HttpOption(func(o *Options) {
		if cfg.Address != "" {
			o.Address = cfg.Address
		}
		o.DebugMode = cfg.Mode.Debug
		o.CorsMode = cfg.Mode.Cors

		if o.StaticLinkMap == nil {
			o.StaticLinkMap = make(map[string]LinkRule)
		}
		if o.PrefixLinkMap == nil {
			o.PrefixLinkMap = make(map[string]LinkRule)
		}
		for _, link := range cfg.StaticLink {
			if link.SrcPath == "" || link.DstPath == "" {
				continue
			}
			o.StaticLinkMap[normalizePath(link.SrcPath)] = LinkRule{Dst: normalizePath(link.DstPath), Methods: link.Methods}
		}
		for _, link := range cfg.PrefixLink {
			if link.SrcPrefix == "" || link.DstPrefix == "" {
				continue
			}
			o.PrefixLinkMap[normalizePath(link.SrcPrefix)] = LinkRule{Dst: normalizePath(link.DstPrefix), Methods: link.Methods}
		}
		for _, nf := range cfg.PageNotFound {
			if nf.Path != "" {
				o.PageNotFoundRules = append(o.PageNotFoundRules, LinkRule{Dst: normalizePath(nf.Path), Methods: nf.Methods})
			}
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
