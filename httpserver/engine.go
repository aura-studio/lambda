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

	if e.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	if e.CorsMode {
		e.Use(Cors())
	}

	e.InstallPackages()
	e.InstallHandlers()

	return e
}
