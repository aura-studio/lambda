package websocket

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigCandidates returns relative paths that will be checked (in order)
// when searching for a default websocket config.
func DefaultConfigCandidates() []string {
	return []string{
		"websocket.yaml",
		"websocket.yml",
		filepath.FromSlash("websocket/websocket.yaml"),
		filepath.FromSlash("websocket/websocket.yml"),
	}
}

// FindDefaultConfigFile searches for a websocket config file in a small set of
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

	return "", fmt.Errorf("websocket config not found (expected %v)", candidates)
}

// WithDefaultConfigFile finds and loads the default websocket config file.
// It panics if the file cannot be found or read.
func WithDefaultConfigFile() Option {
	p, err := FindDefaultConfigFile()
	if err != nil {
		return OptionFunc(func(*Options) {
			panic(fmt.Errorf("websocket.WithDefaultConfigFile: %w", err))
		})
	}
	return WithConfigFile(p)
}
