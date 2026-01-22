package dynamic

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigCandidates returns relative paths that will be checked (in order)
// when searching for a default dynamic config.
func DefaultConfigCandidates() []string {
	return []string{
		"dynamic.yaml",
		"dynamic.yml",
		filepath.FromSlash("dynamic/dynamic.yaml"),
		filepath.FromSlash("dynamic/dynamic.yml"),
	}
}

// FindDefaultConfigFile searches for a dynamic config file in a small set of
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

	return "", fmt.Errorf("dynamic: config not found (expected %v)", candidates)
}

// WithDefaultConfigFile finds and loads the default dynamic config file.
// It panics if the file cannot be found or read.
func WithDefaultConfigFile() Option {
	p, err := FindDefaultConfigFile()
	if err != nil {
		return OptionFunc(func(*Options) {
			   panic(fmt.Errorf("dynamic: WithDefaultConfigFile: %w", err))
		})
	}
	return WithConfigFile(p)
}
