package httpserver

import (
	"github.com/gin-gonic/gin"
)

const (
	HeaderContext       = "header"
	PathContext         = "path"
	RequestContext      = "request"
	ResponseContext     = "response"
	WireRequestContext  = "wire_request"
	WireResponseContext = "wire_response"
	ErrorContext        = "error"
	PanicContext        = "panic"
	DebugContext        = "debug"
	StdoutContext       = "stdout"
	StderrContext       = "stderr"
	ProcessorContext    = "processor"
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

	e.InstallPackages()
	e.InstallHandlers()

	return e
}
