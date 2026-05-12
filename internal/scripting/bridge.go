// Package scripting defines the bridge interface between gi's runtime and
// script engines (Goja/JavaScript and Joker/Clojure).
//
// The bridge exposes gi's internal state to scripts in a controlled way.
// Each engine implements the Runner interface; the bridge provides the
// host functions that scripts can call.
package scripting

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rcarmo/gi/internal/connectivity"
)

// Runner is the interface that script engines must implement.
type Runner interface {
	// Name returns the engine name (e.g. "joker", "js").
	Name() string

	// Execute runs a script string and returns the result.
	Execute(ctx context.Context, script string, bridge *Bridge) (string, error)

	// ExecuteFile runs a script file and returns the result.
	ExecuteFile(ctx context.Context, path string, bridge *Bridge) (string, error)
}

// Bridge provides host functions that scripts can call to interact with
// gi's internals. It is the single point of contact between the
// script engine and gi.
type Bridge struct {
	// SessionID is the active session.
	SessionID string

	// Funcs are the callable host functions exposed to scripts.
	Funcs BridgeFuncs
}

// HTTPCallSpec defines an outbound HTTP request.
// Headers are provided as a map to preserve full caller control.
type HTTPCallSpec struct {
	Method         string              `json:"method"`
	URL            string              `json:"url"`
	Headers        map[string][]string `json:"headers"`
	Body           string              `json:"body"`
	TimeoutMS      int                 `json:"timeout_ms"`
	SkipTLS        bool                `json:"skip_tls"`
	Retry          int                 `json:"retry"`
	AllowRedirects bool                `json:"allow_redirects"`
}

// HTTPResponse holds a low-level HTTP response object.
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	Status     string              `json:"status"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	URL        string              `json:"url"`
}

// EventHookSpec captures a host-level event hook registration intent.
type EventHookSpec struct {
	Name      string            `json:"name"`
	Source    string            `json:"source"`
	Filter    map[string]any    `json:"filter"`
	Arguments map[string]any    `json:"arguments"`
	Engine    string            `json:"engine,omitempty"`
	Script    string            `json:"script,omitempty"`
	Path      string            `json:"path,omitempty"`
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	CWD       string            `json:"cwd,omitempty"`
	Transport string            `json:"transport,omitempty"`
	Protocol  string            `json:"protocol,omitempty"`
}

// ToolSpec captures a script-declared tool registration. A host can either
// execute Script directly or load Path with Engine when the tool is invoked.
type ToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Engine      string          `json:"engine,omitempty"`
	Script      string          `json:"script,omitempty"`
	Path        string          `json:"path,omitempty"`
	Source      string          `json:"source,omitempty"`
}

// RawSocketSpec captures raw socket request metadata.
type RawSocketSpec struct {
	Protocol  string `json:"protocol"` // tcp|udp
	Address   string `json:"address"`
	TimeoutMS int    `json:"timeout_ms"`
	LocalAddr string `json:"local_addr"`
}

// RawSocketPayload contains data for read/write style calls.
type RawSocketPayload struct {
	SocketID  string `json:"socket_id"`
	Data      string `json:"data"`
	MaxBytes  int    `json:"max_bytes"`
	TimeoutMS int    `json:"timeout_ms"`
}

// WebSocketSpec captures websocket connection intent.
type WebSocketSpec struct {
	URL         string              `json:"url"`
	Headers     map[string][]string `json:"headers"`
	Subprotocol string              `json:"subprotocol"`
	TimeoutMS   int                 `json:"timeout_ms"`
}

// BridgeFuncs defines the host functions available to scripts.
// Each function takes and returns generic types that the engine adapter
// marshals to/from the script language's native types.
type BridgeFuncs struct {
	// Session state
	GetSessionState func(ctx context.Context) (map[string]any, error)
	SetSessionState func(ctx context.Context, patch map[string]any) error
	GetSessionInfo  func(ctx context.Context) (map[string]any, error)

	// Messages
	ListMessages func(ctx context.Context, limit int) ([]map[string]any, error)
	AddMessage   func(ctx context.Context, role, content string) error

	// Turns and turn events
	ListTurns      func(ctx context.Context, limit int) ([]map[string]any, error)
	ListTurnEvents func(ctx context.Context, turnID string) ([]map[string]any, error)

	// Runtime config
	GetConfig func(ctx context.Context) (map[string]any, error)

	// Workspace files
	ReadFile  func(ctx context.Context, path string) (string, error)
	WriteFile func(ctx context.Context, path, content string) error
	ListDir   func(ctx context.Context, path string) ([]map[string]any, error)

	// Shell execution
	Exec func(ctx context.Context, command string) (string, error)

	// Event hooks and agentic-loop extension points
	RegisterEventHook func(ctx context.Context, hook EventHookSpec) error
	RegisterTool      func(ctx context.Context, tool ToolSpec) error
	SetActiveTools    func(ctx context.Context, names []string) error
	GetActiveTools    func(ctx context.Context) ([]string, error)
	SetModel          func(ctx context.Context, model string) error
	AppendEntry       func(ctx context.Context, entryType string, data map[string]any) error
	GetEntries        func(ctx context.Context, entryType string) ([]map[string]any, error)
	EmitEvent         func(ctx context.Context, name string, payload map[string]any) error
	ClearEventHooks   func(ctx context.Context) error
	PublishTopic      func(ctx context.Context, envelope map[string]any) error

	// Connectivity and route registration
	RegisterConnectivityRoute   func(ctx context.Context, route connectivity.RouteSpec) (connectivity.RouteInfo, error)
	UnregisterConnectivityRoute func(ctx context.Context, id string) error
	ListConnectivityRoutes      func(ctx context.Context, filter map[string]any) ([]connectivity.RouteInfo, error)
	EmitConnectivityEvent       func(ctx context.Context, topic string, payload map[string]any) error

	// Raw sockets (tcp/udp)
	OpenRawSocket  func(ctx context.Context, spec RawSocketSpec) (string, error)
	WriteRawSocket func(ctx context.Context, payload RawSocketPayload) (int, error)
	ReadRawSocket  func(ctx context.Context, payload RawSocketPayload) (string, error)
	CloseRawSocket func(ctx context.Context, socketID string) error

	// WebSocket interface
	OpenWebSocket  func(ctx context.Context, spec WebSocketSpec) (string, error)
	WriteWebSocket func(ctx context.Context, socketID string, payload string) error
	ReadWebSocket  func(ctx context.Context, socketID string, timeoutMS int) (string, error)
	CloseWebSocket func(ctx context.Context, socketID string) error

	// HTTP with header-first request control
	DoHTTPRequest func(ctx context.Context, req HTTPCallSpec) (HTTPResponse, error)

	// Logging
	Log func(ctx context.Context, level, message string)
}

// NewBridge creates a bridge with the given session ID and functions.
func NewBridge(sessionID string, funcs BridgeFuncs) *Bridge {
	return &Bridge{SessionID: sessionID, Funcs: funcs}
}

// Validate checks that required bridge functions are set.
func (b *Bridge) Validate() error {
	if b.Funcs.GetSessionState == nil {
		return fmt.Errorf("bridge: GetSessionState not set")
	}
	if b.Funcs.ListMessages == nil {
		return fmt.Errorf("bridge: ListMessages not set")
	}
	if b.Funcs.GetConfig == nil {
		return fmt.Errorf("bridge: GetConfig not set")
	}
	if b.Funcs.ReadFile == nil {
		return fmt.Errorf("bridge: ReadFile not set")
	}
	if b.Funcs.Log == nil {
		return fmt.Errorf("bridge: Log not set")
	}
	return nil
}
