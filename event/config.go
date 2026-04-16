package event

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type yamlEventConfig struct {
	Mode struct {
		Debug bool `yaml:"debug"`
	} `yaml:"mode"`
}

func optionFromEventConfig(cfg yamlEventConfig) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = cfg.Mode.Debug
	})
}

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlEventConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return optionFromEventConfig(cfg), nil
}

func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("event.WithConfig: %w", err))
		})
	}
	return opt
}

func WithConfigFile(path string) Option {
	b, err := os.ReadFile(path)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("event.WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}
