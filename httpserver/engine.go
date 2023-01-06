package httpserver

import (
	"github.com/gin-gonic/gin"
)

const (
	PathContext      = "path"
	RequestContext   = "request"
	ResponseContext  = "response"
	ErrorContext     = "error"
	PanicContext     = "panic"
	DebugContext     = "debug"
	StdoutContext    = "stdout"
	StderrContext    = "stderr"
	ProcessorContext = "processor"
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

	e.InstallPackages()
	e.InstallHandlers()

	return e
}
