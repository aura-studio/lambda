package httpserver

import "github.com/aura-studio/dynamic"

func ActivePackage(packageName string, commit string) {
	dynamic.GetPackage(packageName, commit)
}
