package reqresp

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// yamlReqRespConfig represents the YAML configuration structure for reqresp module
type yamlReqRespConfig struct {
	Mode struct {
		Debug bool `yaml:"debug"`
	} `yaml:"mode"`
	StaticLink []struct {
		SrcPath string `yaml:"srcPath"`
		DstPath string `yaml:"dstPath"`
	} `yaml:"staticLink"`
	PrefixLink []struct {
		SrcPrefix string `yaml:"srcPrefix"`
		DstPrefix string `yaml:"dstPrefix"`
	} `yaml:"prefixLink"`
}

func optionFromReqRespConfig(cfg yamlReqRespConfig) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = cfg.Mode.Debug

		if o.StaticLinkMap == nil {
			o.StaticLinkMap = make(map[string]string)
		}
		for _, link := range cfg.StaticLink {
			if link.SrcPath == "" || link.DstPath == "" {
				continue
			}
			o.StaticLinkMap[link.SrcPath] = link.DstPath
		}

		if o.PrefixLinkMap == nil {
			o.PrefixLinkMap = make(map[string]string)
		}
		for _, link := range cfg.PrefixLink {
			if link.SrcPrefix == "" || link.DstPrefix == "" {
				continue
			}
			o.PrefixLinkMap[link.SrcPrefix] = link.DstPrefix
		}
	})
}

// optionFromConfigBytes parses YAML bytes and returns an Option.
// Returns an error if the YAML is invalid.
func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlReqRespConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return optionFromReqRespConfig(cfg), nil
}

// WithConfig parses YAML bytes following reqresp.yml structure and applies it to Options.
// It panics if the YAML is invalid.
func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("reqresp.WithConfig: %w", err))
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
			panic(fmt.Errorf("reqresp.WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}
