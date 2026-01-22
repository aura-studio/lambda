package http

import (
	"github.com/aura-studio/lambda/dynamic"
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*Options
	*gin.Engine
	*dynamic.Dynamic
}

func NewEngine(httpOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(httpOpts...),
		Engine:  gin.Default(),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
	}

	if !e.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	if e.CorsMode {
		e.Use(Cors())
	}

	e.InstallHandlers()

	return e
}
