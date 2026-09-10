package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aura-studio/dynamic"
	lambdaws "github.com/aura-studio/lambda/websocket"
	events "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	mgmt "github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
)

var (
	errTestUnauthorized = errors.New("unauthorized")
	errTestGone         = &types.GoneException{Message: aws.String("gone")}
)

type mockManagementClient struct {
	posts   []*mgmt.PostToConnectionInput
	deletes []*mgmt.DeleteConnectionInput
	postErr error
}

func (m *mockManagementClient) PostToConnection(ctx context.Context, params *mgmt.PostToConnectionInput, optFns ...func(*mgmt.Options)) (*mgmt.PostToConnectionOutput, error) {
	m.posts = append(m.posts, params)
	return &mgmt.PostToConnectionOutput{}, m.postErr
}

func (m *mockManagementClient) DeleteConnection(ctx context.Context, params *mgmt.DeleteConnectionInput, optFns ...func(*mgmt.Options)) (*mgmt.DeleteConnectionOutput, error) {
	m.deletes = append(m.deletes, params)
	return &mgmt.DeleteConnectionOutput{}, nil
}

func (m *mockManagementClient) GetConnection(ctx context.Context, params *mgmt.GetConnectionInput, optFns ...func(*mgmt.Options)) (*mgmt.GetConnectionOutput, error) {
	return &mgmt.GetConnectionOutput{}, nil
}

func wsEvent(routeKey, connectionID, body string) events.APIGatewayWebsocketProxyRequest {
	return events.APIGatewayWebsocketProxyRequest{
		Body: body,
		RequestContext: events.APIGatewayWebsocketProxyRequestContext{
			RouteKey:     routeKey,
			EventType:    "MESSAGE",
			ConnectionID: connectionID,
			DomainName:   "example.execute-api.us-east-1.amazonaws.com",
			Stage:        "prod",
		},
	}
}

// TestWebSocketEngineCreation tests that NewEngine creates an engine with correct options
func TestWebSocketEngineCreation(t *testing.T) {
	mock := &mockManagementClient{}
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(mock),
		lambdaws.WithDebugMode(true),
		lambdaws.WithReplyMode(false),
	}, nil)

	if e == nil {
		t.Fatal("NewEngine returned nil")
	}
	if !e.DebugMode {
		t.Error("DebugMode should be true")
	}
	if e.ReplyMode {
		t.Error("ReplyMode should be false")
	}
}

// TestWebSocketConnectAccepted tests the default $connect handler accepts the connection
func TestWebSocketConnectAccepted(t *testing.T) {
	e := lambdaws.NewEngine([]lambdaws.Option{lambdaws.WithManagementClient(&mockManagementClient{})}, nil)

	resp, err := e.Invoke(context.Background(), wsEvent("$connect", "c1", ""))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

// TestWebSocketConnectRejected tests that a connect hook error rejects the connection with 403
func TestWebSocketConnectRejected(t *testing.T) {
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(&mockManagementClient{}),
		lambdaws.WithConnectHandler(func(c *lambdaws.Context) {
			c.Set(lambdaws.ContextError, errTestUnauthorized)
		}),
	}, nil)

	resp, err := e.Invoke(context.Background(), wsEvent("$connect", "c1", ""))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", resp.StatusCode)
	}
}

// TestWebSocketDisconnectHook tests that the disconnect hook runs on $disconnect
func TestWebSocketDisconnectHook(t *testing.T) {
	var disconnected string
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(&mockManagementClient{}),
		lambdaws.WithDisconnectHandler(func(c *lambdaws.Context) {
			disconnected = c.GetString(lambdaws.ContextConnectionID)
		}),
	}, nil)

	resp, err := e.Invoke(context.Background(), wsEvent("$disconnect", "c9", ""))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if disconnected != "c9" {
		t.Errorf("disconnect hook got connection %q, want 'c9'", disconnected)
	}
}

// TestWebSocketMessageToDynamicTunnel tests $default dispatching to a dynamic package
// tunnel and pushing the reply back via the management API
func TestWebSocketMessageToDynamicTunnel(t *testing.T) {
	mock := &mockManagementClient{}

	var invokedRoute, invokedReq string
	dynamic.RegisterPackage("wspkg", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedRoute = route
			invokedReq = req
			data := base64.StdEncoding.EncodeToString([]byte(`{"balance":100}`))
			return `{"meta":{},"data":"` + data + `"}`
		},
	})

	e := lambdaws.NewEngine([]lambdaws.Option{lambdaws.WithManagementClient(mock)}, nil)

	body := `{"path":"/wspkg/v1/spin","payload":{"amount":1}}`
	resp, err := e.Invoke(context.Background(), wsEvent("$default", "c2", body))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if invokedRoute != "/spin" {
		t.Errorf("invokedRoute = %q, want '/spin'", invokedRoute)
	}

	var reqEnvelope struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(invokedReq), &reqEnvelope); err != nil {
		t.Fatalf("tunnel request is not an envelope: %v", err)
	}
	payload, err := base64.StdEncoding.DecodeString(reqEnvelope.Data)
	if err != nil {
		t.Fatalf("decode tunnel payload: %v", err)
	}
	if string(payload) != `{"amount":1}` {
		t.Errorf("tunnel payload = %q, want '{\"amount\":1}'", payload)
	}

	if len(mock.posts) != 1 {
		t.Fatalf("Expected 1 post to connection, got %d", len(mock.posts))
	}
	if *mock.posts[0].ConnectionId != "c2" {
		t.Errorf("post connection id = %q, want 'c2'", *mock.posts[0].ConnectionId)
	}
	var reply struct {
		Path    string          `json:"path"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(mock.posts[0].Data, &reply); err != nil {
		t.Fatalf("reply is not JSON: %v", err)
	}
	if reply.Path != "/wspkg/v1/spin" {
		t.Errorf("reply path = %q, want '/wspkg/v1/spin'", reply.Path)
	}
	if string(reply.Payload) != `{"balance":100}` {
		t.Errorf("reply payload = %q, want '{\"balance\":100}'", reply.Payload)
	}
}

// TestWebSocketMessageInvalidEnvelope tests that a malformed body posts an error reply
func TestWebSocketMessageInvalidEnvelope(t *testing.T) {
	mock := &mockManagementClient{}
	e := lambdaws.NewEngine([]lambdaws.Option{lambdaws.WithManagementClient(mock)}, nil)

	resp, err := e.Invoke(context.Background(), wsEvent("$default", "c3", "not-json"))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", resp.StatusCode)
	}

	if len(mock.posts) != 1 {
		t.Fatalf("Expected 1 error post, got %d", len(mock.posts))
	}
	var reply struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(mock.posts[0].Data, &reply); err != nil {
		t.Fatalf("error reply is not JSON: %v", err)
	}
	if reply.Error == "" {
		t.Error("error reply should carry an error message")
	}
}

// TestWebSocketMessageNoReplyMode tests that ReplyMode=false suppresses pushes
func TestWebSocketMessageNoReplyMode(t *testing.T) {
	mock := &mockManagementClient{}
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(mock),
		lambdaws.WithReplyMode(false),
	}, nil)

	body := `{"path":"/wspkg/v1/spin","payload":{}}`
	resp, err := e.Invoke(context.Background(), wsEvent("$default", "c4", body))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if len(mock.posts) != 0 {
		t.Errorf("Expected no posts with ReplyMode off, got %d", len(mock.posts))
	}
}

// TestWebSocketMessageBase64Body tests base64 frame decoding
func TestWebSocketMessageBase64Body(t *testing.T) {
	mock := &mockManagementClient{}

	var invokedReq string
	dynamic.RegisterPackage("wsb64", "v1", &mockTunnel{
		invoke: func(route, req string) string {
			invokedReq = req
			return "ok"
		},
	})

	e := lambdaws.NewEngine([]lambdaws.Option{lambdaws.WithManagementClient(mock)}, nil)

	raw := `{"path":"/wsb64/v1/ping","payload":"hello"}`
	ev := wsEvent("$default", "c5", base64.StdEncoding.EncodeToString([]byte(raw)))
	ev.IsBase64Encoded = true

	resp, err := e.Invoke(context.Background(), ev)
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	var reqEnvelope struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(invokedReq), &reqEnvelope); err != nil {
		t.Fatalf("tunnel request is not an envelope: %v", err)
	}
	payload, _ := base64.StdEncoding.DecodeString(reqEnvelope.Data)
	if string(payload) != "hello" {
		t.Errorf("tunnel payload = %q, want 'hello'", payload)
	}
}

// TestWebSocketCustomRoute tests registering a custom route key via WithRoute
func TestWebSocketCustomRoute(t *testing.T) {
	var gotConn, gotBody string
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(&mockManagementClient{}),
		lambdaws.WithRoute("subscribe", func(c *lambdaws.Context) {
			gotConn = c.GetString(lambdaws.ContextConnectionID)
			gotBody = c.GetString(lambdaws.ContextRequest)
		}),
	}, nil)

	resp, err := e.Invoke(context.Background(), wsEvent("subscribe", "c6", `{"topic":"jackpot"}`))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if gotConn != "c6" {
		t.Errorf("handler got connection %q, want 'c6'", gotConn)
	}
	if gotBody != `{"topic":"jackpot"}` {
		t.Errorf("handler got body %q, want '{\"topic\":\"jackpot\"}'", gotBody)
	}
}

// TestWebSocketUnknownRouteKey tests that an unmatched route key hits NoRoute
func TestWebSocketUnknownRouteKey(t *testing.T) {
	e := lambdaws.NewEngine([]lambdaws.Option{lambdaws.WithManagementClient(&mockManagementClient{})}, nil)

	resp, err := e.Invoke(context.Background(), wsEvent("nosuchroute", "c7", ""))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", resp.StatusCode)
	}
}

// TestWebSocketPostToConnectionGone tests that PostToConnection surfaces GoneException
func TestWebSocketPostToConnectionGone(t *testing.T) {
	mock := &mockManagementClient{postErr: errTestGone}
	e := lambdaws.NewEngine([]lambdaws.Option{lambdaws.WithManagementClient(mock)}, nil)

	err := e.PostToConnection(context.Background(), "domain", "stage", "dead-conn", []byte("x"))
	if err == nil {
		t.Fatal("PostToConnection should return the gone error")
	}
	if !lambdaws.IsGone(err) {
		t.Errorf("IsGone(%v) = false, want true", err)
	}
}

// TestWebSocketConnectQueryParams tests that headers and query string are exposed on the context
func TestWebSocketConnectQueryParams(t *testing.T) {
	var token, ua string
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(&mockManagementClient{}),
		lambdaws.WithConnectHandler(func(c *lambdaws.Context) {
			token = c.GetStringMapString(lambdaws.ContextQuery)["token"]
			ua = c.GetStringMapString(lambdaws.ContextHeaders)["User-Agent"]
		}),
	}, nil)

	ev := wsEvent("$connect", "c8", "")
	ev.QueryStringParameters = map[string]string{"token": "t-123"}
	ev.Headers = map[string]string{"User-Agent": "tester"}

	resp, err := e.Invoke(context.Background(), ev)
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if token != "t-123" {
		t.Errorf("token = %q, want 't-123'", token)
	}
	if ua != "tester" {
		t.Errorf("ua = %q, want 'tester'", ua)
	}
}

// TestWebSocketDefaultHandlerOverride tests that WithDefaultHandler replaces the
// built-in envelope Message handler, so non-envelope bodies reach it raw
func TestWebSocketDefaultHandlerOverride(t *testing.T) {
	var gotBody, gotConn string
	e := lambdaws.NewEngine([]lambdaws.Option{
		lambdaws.WithManagementClient(&mockManagementClient{}),
		lambdaws.WithDefaultHandler(func(c *lambdaws.Context) {
			gotBody = c.GetString(lambdaws.ContextRequest)
			gotConn = c.GetString(lambdaws.ContextConnectionID)
		}),
	}, nil)

	// Not a valid envelope — the built-in Message handler would reject this.
	resp, err := e.Invoke(context.Background(), wsEvent("$default", "c10", "free-form text"))
	if err != nil {
		t.Fatalf("Invoke error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if gotBody != "free-form text" {
		t.Errorf("handler got body %q, want 'free-form text'", gotBody)
	}
	if gotConn != "c10" {
		t.Errorf("handler got connection %q, want 'c10'", gotConn)
	}
}
