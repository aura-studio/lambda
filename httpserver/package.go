package httpserver

import (
	"log"

	"github.com/aura-studio/dynamic"
)

func (e *Engine) InstallPackages() {
	dynamic.UseNamespace(e.Namespace)

	for _, p := range e.StaticPackages {
		dynamic.RegisterPackage(p.Name, p.Commit, p.Tunnel)
	}

	for _, p := range e.PreloadPackages {
		if _, err := dynamic.GetPackage(p.Name, p.Commit); err != nil {
			log.Printf("preload package %s@%s failed: %v", p.Name, p.Commit, err)
		}
	}
}
