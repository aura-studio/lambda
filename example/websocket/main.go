// Command example is a minimal WebSocket Lambda built on the websocket engine.
//
// It wires one dynamic package tunnel ("echo/v1") plus a lightweight "ping"
// route, and demonstrates the connect/disconnect hooks. The tunnel is built
// with tunnelkit, so handlers never touch the wire envelope. Build and
// deployment are documented in README.md (Chinese) in this directory.
//
// Message contract (client -> server, route selection $request.body.action):
//
//	{"path": "/echo/v1/echo", "payload": {...}}   -> $default -> echo tunnel
//	{"path": "/echo/v1/time"}                     -> $default -> echo tunnel
//	{"action": "ping"}                            -> ping route (cheap heartbeat)
//
// Connect with an optional query token, e.g.:
//
//	wss://{api-id}.execute-api.{region}.amazonaws.com/dev?token=demo
//
// token=deny is rejected on $connect to demonstrate authorization.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aura-studio/lambda/dynamic"
	"github.com/aura-studio/lambda/server"
	"github.com/aura-studio/lambda/tunnelkit"
	lambdaws "github.com/aura-studio/lambda/websocket"
)

func main() {
	echo := tunnelkit.New("") // Meta() is generated from the registered routes
	echo.Handle("/echo", onEcho)
	echo.Handle("/time", onTime)

	err := server.Serve(
		server.WithLambdaType("websocket"),
		server.WithWebsocketOptions(
			lambdaws.WithConnectHandler(onConnect),
			lambdaws.WithDisconnectHandler(onDisconnect),
			// "ping" is a dedicated cheap route: it never touches tunnels or
			// storage, keeping per-heartbeat Lambda duration at a minimum.
			lambdaws.WithRoute("ping", onPing),
		),
		server.WithDynamicOptions(
			dynamic.WithStaticPackage(&dynamic.Package{
				Package: "echo",
				Version: "v1",
				Tunnel:  echo,
			}),
		),
	)
	if err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func onConnect(c *lambdaws.Context) {
	token := c.GetStringMapString(lambdaws.ContextQuery)["token"]
	log.Printf("[example] connect id=%s token=%q", c.GetString(lambdaws.ContextConnectionID), token)

	// Demonstrate $connect authorization: setting an error rejects the
	// connection with HTTP 403. Replace with real session validation.
	if token == "deny" {
		c.Set(lambdaws.ContextError, fmt.Errorf("invalid token"))
		return
	}

	// Persist c.GetString(lambdaws.ContextConnectionID) -> user mapping here
	// (e.g. DynamoDB with TTL) to enable server-initiated pushes later.
}

func onDisconnect(c *lambdaws.Context) {
	log.Printf("[example] disconnect id=%s", c.GetString(lambdaws.ContextConnectionID))
	// Delete the connection mapping here. Note $disconnect is best-effort;
	// also clean up on IsGone(err) when pushing.
}

func onPing(c *lambdaws.Context) {
	body, err := json.Marshal(map[string]any{"pong": time.Now().UnixMilli()})
	if err != nil {
		c.Set(lambdaws.ContextError, err)
		return
	}
	if err := c.Post(body); err != nil {
		c.Set(lambdaws.ContextError, err)
	}
}

func onEcho(c *tunnelkit.Context) error {
	return c.JSON(map[string]any{
		"route": c.Route,
		"echo":  rawOrString(c.Payload),
	})
}

func onTime(c *tunnelkit.Context) error {
	return c.JSON(map[string]any{
		"now": time.Now().UTC().Format(time.RFC3339),
	})
}

// rawOrString embeds payload as raw JSON when valid, otherwise as a string.
func rawOrString(data []byte) any {
	if len(data) > 0 && json.Valid(data) {
		return json.RawMessage(data)
	}
	return string(data)
}
