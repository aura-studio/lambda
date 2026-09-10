package websocket

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlWebSocketConfig struct {
	Mode struct {
		Debug bool  `yaml:"debug"`
		Reply *bool `yaml:"reply"`
	} `yaml:"mode"`
}

func optionFromWebSocketConfig(cfg yamlWebSocketConfig) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = cfg.Mode.Debug
		if cfg.Mode.Reply != nil {
			o.ReplyMode = *cfg.Mode.Reply
		}
	})
}

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlWebSocketConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return optionFromWebSocketConfig(cfg), nil
}

// WithConfig parses YAML bytes following websocket.yml structure and applies it
// to Options. It panics if the YAML is invalid.
func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("websocket.WithConfig: %w", err))
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
			panic(fmt.Errorf("websocket.WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}
