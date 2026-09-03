package websocket

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/aura-studio/lambda/dynamic"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

type Engine struct {
	*Options
	*Router
	*dynamic.Dynamic
}

func NewEngine(wsOpts []Option, dynamicOpts []dynamic.Option) *Engine {
	e := &Engine{
		Options: NewOptions(wsOpts...),
		Dynamic: dynamic.NewDynamic(dynamicOpts...),
		Router:  NewRouter(),
	}
	e.InstallHandlers()
	for key, handlers := range e.Options.Routes {
		e.Handle(key, handlers...)
	}
	return e
}

// Invoke handles one API Gateway WebSocket proxy event. Route selection is
// driven by RequestContext.RouteKey: the built-in $connect/$disconnect/$default
// routes run first, custom route keys registered via WithRoute take precedence
// over $default fallback handled by the gateway itself.
func (e *Engine) Invoke(ctx context.Context, ev events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	rc := ev.RequestContext
	if e.DebugMode {
		log.Printf("[WebSocket] event %s connection %s route %s", rc.EventType, rc.ConnectionID, rc.RouteKey)
	}

	body := ev.Body
	if ev.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(ev.Body)
		if err != nil {
			log.Printf("[WebSocket] decode base64 body for connection %s error: %v", rc.ConnectionID, err)
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "invalid base64 body"}, nil
		}
		body = string(decoded)
	}

	c := &Context{ctx: ctx, engine: e}
	c.Set(ContextRouteKey, rc.RouteKey)
	c.Set(ContextEventType, rc.EventType)
	c.Set(ContextConnectionID, rc.ConnectionID)
	c.Set(ContextDomainName, rc.DomainName)
	c.Set(ContextStage, rc.Stage)
	c.Set(ContextRequestContext, rc)
	c.Set(ContextHeaders, ev.Headers)
	c.Set(ContextQuery, ev.QueryStringParameters)
	c.Set(ContextRequest, body)

	e.Router.Dispatch(c)

	if v, ok := c.Get(ContextPanic); ok && v != nil {
		log.Printf("[WebSocket] connection %s route %s panic: %v", rc.ConnectionID, rc.RouteKey, v)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "internal error"}, nil
	}
	if err := c.GetError(); err != nil {
		log.Printf("[WebSocket] connection %s route %s error: %v", rc.ConnectionID, rc.RouteKey, err)
		if rc.RouteKey == "$connect" {
			// A non-200 status on $connect rejects the connection.
			return events.APIGatewayProxyResponse{StatusCode: 403, Body: err.Error()}, nil
		}
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: err.Error()}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

// PostToConnection pushes data to an established connection via the API Gateway
// management API. It is exported so event producers outside the request flow
// (broadcasts, jackpot pushes) can reuse the engine's client resolution; check
// IsGone on the returned error to detect dead connections.
func (e *Engine) PostToConnection(ctx context.Context, domainName, stage, connectionID string, data []byte) error {
	client, err := e.managementClient(ctx, domainName, stage)
	if err != nil {
		return err
	}

	_, err = client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         data,
	})
	return err
}

// DropConnection force-closes an established connection (e.g. kick on logout).
func (e *Engine) DropConnection(ctx context.Context, domainName, stage, connectionID string) error {
	client, err := e.managementClient(ctx, domainName, stage)
	if err != nil {
		return err
	}

	_, err = client.DeleteConnection(ctx, &apigatewaymanagementapi.DeleteConnectionInput{
		ConnectionId: aws.String(connectionID),
	})
	return err
}

func (e *Engine) managementClient(ctx context.Context, domainName, stage string) (ManagementClient, error) {
	if e.Options.ManagementClient != nil {
		return e.Options.ManagementClient, nil
	}
	factory := e.Options.ManagementClientFactory
	if factory == nil {
		factory = NewManagementClient
	}
	return factory(ctx, domainName, stage)
}

func (e *Engine) logf(format string, args ...any) {
	log.Printf("[WebSocket] "+format, args...)
}
