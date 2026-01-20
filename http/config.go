package http

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
	HTTP struct {
		Release bool `yaml:"release"`
		Cors    bool `yaml:"cors"`
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
	} `yaml:"http"`
}

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return HttpOption(func(o *Options) {
		o.ReleaseMode = cfg.HTTP.Release
		o.CorsMode = cfg.HTTP.Cors

		for _, link := range cfg.HTTP.StaticLink {
			if link.SrcPath == "" || link.DstPath == "" {
				continue
			}
			o.StaticLinkMap[link.SrcPath] = link.DstPath
		}
		for _, link := range cfg.HTTP.PrefixLink {
			if link.SrcPrefix == "" || link.DstPrefix == "" {
				continue
			}
			o.PrefixLinkMap[link.SrcPrefix] = link.DstPrefix
		}
		for _, link := range cfg.HTTP.HeaderLinkKey {
			if link.Key == "" || link.Prefix == "" {
				continue
			}
			o.HeaderLinkMap[link.Key] = link.Prefix
		}
	}), nil
}

// WithConfig parses YAML bytes following http.yaml structure and applies it to Options.
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
