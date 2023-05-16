package httpserver

import (
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*Options
	*gin.Engine
}

func NewEngine(opts ...Option) *Engine {
	e := &Engine{
		Options: NewOptions(opts...),
		Engine:  gin.Default(),
	}

	e.Engine.SetTrustedProxies(nil)
	e.Engine.TrustedPlatform = "X-Forwarded-For"

	if e.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	e.InstallPackages()
	e.InstallHandlers()

	return e
}
