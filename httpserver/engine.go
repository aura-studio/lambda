package httpserver

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*Options
	*gin.Engine
	*dynamic.Dynamic
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

	e.InstallHandlers()

	return e
}
