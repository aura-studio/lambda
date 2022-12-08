package httpserver

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

const (
	debugFormat = `Stdout: %s
Stderr: %s
Error: %v
Response: %s`
)
