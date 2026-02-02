package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/http"
	"github.com/aura-studio/lambda/sqs"
	yaml "gopkg.in/yaml.v2"
)

type yamlServerConfig struct {
	Lambda  string `yaml:"lambda"`
	HTTP    any    `yaml:"http"`
	SQS     any    `yaml:"sqs"`
	Dynamic any    `yaml:"dynamic"`
}

type Option interface {
	Apply(*Options)
}

type Options struct {
	Lambda  string
	Http    []http.Option
	Sqs     []sqs.Option
	Dynamic []dynamic.Option
}

type serveOptionFunc func(*Options)

func (f serveOptionFunc) Apply(o *Options) { f(o) }

type serveConfigOption struct {
	lambda  string
	httpOpt http.Option
	sqsOpt  sqs.Option
	dynOpt  dynamic.Option
}

func (o serveConfigOption) Apply(opts *Options) {
	if o.lambda != "" {
		opts.Lambda = o.lambda
	}
	if o.httpOpt != nil {
		opts.Http = append(opts.Http, o.httpOpt)
	}
	if o.sqsOpt != nil {
		opts.Sqs = append(opts.Sqs, o.sqsOpt)
	}
	if o.dynOpt != nil {
		opts.Dynamic = append(opts.Dynamic, o.dynOpt)
	}
}

// WithServeConfig parses YAML bytes following server.yml structure.
func WithServeConfig(yamlBytes []byte) Option {
	var cfg yamlServerConfig
	if err := yaml.Unmarshal(yamlBytes, &cfg); err != nil {
		panic(fmt.Errorf("server.WithServeConfig: %w", err))
	}

	var httpOpt http.Option
	if cfg.HTTP != nil {
		b, err := yaml.Marshal(cfg.HTTP)
		if err != nil {
			panic(fmt.Errorf("server.WithServeConfig: %w", err))
		}
		httpOpt = http.WithConfig(b)
	}

	var sqsOpt sqs.Option
	if cfg.SQS != nil {
		b, err := yaml.Marshal(cfg.SQS)
		if err != nil {
			panic(fmt.Errorf("server.WithServeConfig: %w", err))
		}
		sqsOpt = sqs.WithConfig(b)
	}

	var dynOpt dynamic.Option
	if cfg.Dynamic != nil {
		b, err := yaml.Marshal(cfg.Dynamic)
		if err != nil {
			panic(fmt.Errorf("server.WithServeConfig: %w", err))
		}
		dynOpt = dynamic.WithConfig(b)
	}

	return serveConfigOption{
		lambda:  cfg.Lambda,
		httpOpt: httpOpt,
		sqsOpt:  sqsOpt,
		dynOpt:  dynOpt,
	}
}

// WithServeConfigFile loads a YAML file and applies it as ServeOption.
func WithServeConfigFile(path string) Option {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("server.WithServeConfigFile(%s): %w", path, err))
	}
	return WithServeConfig(b)
}

// DefaultServeConfigCandidates returns relative paths that will be checked (in order)
// when searching for a default server config.
func DefaultServeConfigCandidates() []string {
	return []string{
		"lambda.yaml",
		"lambda.yml",
		"server.yaml",
		"server.yml",
		"bootstrap.yaml",
		"bootstrap.yml",
		"app.yaml",
		"app.yml",
		"config.yaml",
		"config.yml",
	}
}

// FindDefaultServeConfigFile searches for a server config file in a small set of
// well-known locations (CWD then executable directory).
func FindDefaultServeConfigFile() (string, error) {
	candidates := DefaultServeConfigCandidates()

	dirs := []string{"."}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}

	for _, dir := range dirs {
		for _, rel := range candidates {
			p := rel
			if dir != "." {
				p = filepath.Join(dir, rel)
			}
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("server config not found (expected %v)", candidates)
}

// WithDefaultServeConfigFile finds and loads the default server config file as a ServeOption.
func WithDefaultServeConfigFile() Option {
	p, err := FindDefaultServeConfigFile()
	if err != nil {
		panic(fmt.Errorf("server.WithDefaultServeConfigFile: %w", err))
	}
	return WithServeConfigFile(p)
}
