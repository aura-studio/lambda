package http

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigCandidates returns relative paths that will be checked (in order)
// when searching for a default http config.
func DefaultConfigCandidates() []string {
	return []string{
		"http.yaml",
		"http.yml",
		filepath.FromSlash("http/http.yaml"),
		filepath.FromSlash("http/http.yml"),
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

	return "", fmt.Errorf("http config not found (expected %v)", candidates)
}

// WithDefaultConfigFile finds and loads the default http config file.
// It panics if the file cannot be found or read.
func WithDefaultConfigFile() Option {
	p, err := FindDefaultConfigFile()
	if err != nil {
		return HttpOption(func(*Options) {
			panic(fmt.Errorf("http.WithDefaultConfigFile: %w", err))
		})
	}
	return WithConfigFile(p)
}
