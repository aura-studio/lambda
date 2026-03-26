package reqresp

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type yamlReqRespConfig struct {
	Mode struct {
		Debug bool `yaml:"debug"`
	} `yaml:"mode"`
}

func optionFromReqRespConfig(cfg yamlReqRespConfig) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = cfg.Mode.Debug
	})
}

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlReqRespConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return optionFromReqRespConfig(cfg), nil
}

func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("reqresp.WithConfig: %w", err))
		})
	}
	return opt
}

func WithConfigFile(path string) Option {
	b, err := os.ReadFile(path)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("reqresp.WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}
