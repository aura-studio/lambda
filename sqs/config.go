package sqs

import (
	"fmt"
	"os"

	"github.com/aura-studio/lambda/dynamic"
	yaml "gopkg.in/yaml.v2"
)

type yamlSQSConfig struct {
	Debug          bool `yaml:"debug"`
	ResponseSwitch bool `yaml:"responseSwitch"`
	ErrorSuspend   bool `yaml:"errorSuspend"`
	PartialRetry   bool `yaml:"partialRetry"`
	StaticLink     []struct {
		SrcPath string `yaml:"srcPath"`
		DstPath string `yaml:"dstPath"`
	} `yaml:"staticLink"`
	PrefixLink []struct {
		SrcPrefix string `yaml:"srcPrefix"`
		DstPrefix string `yaml:"dstPrefix"`
	} `yaml:"prefixLink"`
}

type yamlServeConfig struct {
	SQS     yamlSQSConfig `yaml:"sqs"`
	Dynamic any           `yaml:"dynamic"`
}

func optionFromSQSConfig(cfg yamlSQSConfig) Option {
	return OptionFunc(func(o *Options) {
		o.DebugMode = cfg.Debug
		o.ResponseSwitch = cfg.ResponseSwitch
		o.ErrorSuspend = cfg.ErrorSuspend
		o.PartialRetry = cfg.PartialRetry

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

type serveConfigOption struct {
	sqsOpt Option
	dynOpt dynamic.Option
	err    error
}

func (o serveConfigOption) apply(b *serveOptionBag) {
	if o.err != nil {
		panic(fmt.Errorf("sqs.WithServeConfig: %w", o.err))
	}
	if o.sqsOpt != nil {
		b.sqs = append(b.sqs, o.sqsOpt)
	}
	if o.dynOpt != nil {
		b.dynamic = append(b.dynamic, o.dynOpt)
	}
}

// WithServeConfig parses YAML bytes following sqs.yml structure, and also supports
// embedding dynamic.yml content under top-level `dynamic:`.
// It panics if the YAML is invalid.
func WithServeConfig(yamlBytes []byte) ServeOption {
	var cfg yamlServeConfig
	if err := yaml.Unmarshal(yamlBytes, &cfg); err != nil {
		return serveConfigOption{err: err}
	}

	sqsOpt := optionFromSQSConfig(cfg.SQS)

	var dynOpt dynamic.Option
	if cfg.Dynamic != nil {
		b, err := yaml.Marshal(cfg.Dynamic)
		if err != nil {
			return serveConfigOption{err: err}
		}
		// cfg.Dynamic is expected to be a dynamic.yml document root (environment/package).
		dynOpt = dynamic.WithConfig(b)
	}

	return serveConfigOption{sqsOpt: sqsOpt, dynOpt: dynOpt}
}

// WithServeConfigFile loads a YAML file and applies it as ServeOption.
// It panics if the file cannot be read or YAML is invalid.
func WithServeConfigFile(path string) ServeOption {
	b, err := os.ReadFile(path)
	if err != nil {
		return serveConfigOption{err: fmt.Errorf("sqs.WithServeConfigFile(%s): %w", path, err)}
	}
	return WithServeConfig(b)
}
