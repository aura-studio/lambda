package sqs

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlSQSConfig struct {
	Mode struct {
		Debug bool    `yaml:"debug"`
		Run   RunMode `yaml:"run"`
		Reply bool    `yaml:"reply"`
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

func optionFromSQSConfig(cfg yamlSQSConfig) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = cfg.Mode.Debug
		o.ReplyMode = cfg.Mode.Reply
		if cfg.Mode.Run != "" {
			switch cfg.Mode.Run {
			case RunModeStrict, RunModePartial, RunModeBatch, RunModeReentrant:
				o.RunMode = cfg.Mode.Run
			default:
				panic(fmt.Errorf("sqs: unrecognized run mode: %q", cfg.Mode.Run))
			}
		}

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

func optionFromConfigBytes(b []byte) (Option, error) {
	var cfg yamlSQSConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return optionFromSQSConfig(cfg), nil
}

// WithConfig parses YAML bytes following sqs.yml structure and applies it to Options.
// It panics if the YAML is invalid.
func WithConfig(yamlBytes []byte) Option {
	opt, err := optionFromConfigBytes(yamlBytes)
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("sqs.WithConfig: %w", err))
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
			panic(fmt.Errorf("sqs.WithConfigFile(%s): %w", path, err))
		})
	}
	return WithConfig(b)
}
