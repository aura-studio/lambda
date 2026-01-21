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

func NewEngine(opts ...ServeOption) *Engine {
	bag := &serveOptionBag{}
	bag.apply(opts...)

	e := &Engine{
		Options: NewOptions(bag.http...),
		Engine:  gin.Default(),
		Dynamic: dynamic.NewDynamic(bag.dynamic...),
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
