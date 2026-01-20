package http

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigCandidates returns relative paths that will be checked (in order)
// when searching for a default http config.
func DefaultConfigCandidates() []string {
	return []string{
		"http.yaml",
		filepath.FromSlash("http/http.yaml"),
	}
}

// FindDefaultConfigFile searches for an http config file in a small set of
// well-known locations (CWD then executable directory).
func FindDefaultConfigFile() (string, error) {
	candidates := DefaultConfigCandidates()

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

	return "", errors.New("http config not found (expected http.yaml or http/http.yaml)")
}

// WithDefaultConfig finds and loads the default http config file.
// It panics if the file cannot be found or read.
func WithDefaultConfig() Option {
	p, err := FindDefaultConfigFile()
	if err != nil {
		return HttpOption(func(*Options) {
			panic(fmt.Errorf("http.WithDefaultConfig: %w", err))
		})
	}
	return WithConfigFile(p)
}

// WithDefaultServeConfig finds and loads the default http config file as a ServeOption.
// This supports the optional embedded `dynamic:` section in http.yaml.
func WithDefaultServeConfig() ServeOption {
	p, err := FindDefaultConfigFile()
	if err != nil {
		return serveConfigOption{err: fmt.Errorf("http.WithDefaultServeConfig: %w", err)}
	}
	return WithServeConfigFile(p)
}
