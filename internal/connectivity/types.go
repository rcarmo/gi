package connectivity

import (
	"context"
	"time"
)

// RouteSpec describes a managed connectivity route. The common shape is used
// by transport-specific helpers (HTTP, SSE, WebSocket, MQTT, raw sockets, pipes)
// and by scripts through gi.connect.registerRoute().
type RouteSpec struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Transport string         `json:"transport"` // http|sse|websocket|mqtt|tcp|udp|unix|pipe|event
	Direction string         `json:"direction,omitempty"`
	Source    string         `json:"source,omitempty"`
	Match     map[string]any `json:"match,omitempty"`
	Auth      map[string]any `json:"auth,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
	Engine    string         `json:"engine,omitempty"`
	Script    string         `json:"script,omitempty"`
	Path      string         `json:"path,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	Mode      string         `json:"mode,omitempty"`     // observe|message|turn|tool|respond
	Lifetime  string         `json:"lifetime,omitempty"` // turn|session|workspace|process
}

// RouteInfo is a safe listing representation.
type RouteInfo struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Transport string         `json:"transport"`
	Direction string         `json:"direction,omitempty"`
	Source    string         `json:"source,omitempty"`
	Match     map[string]any `json:"match,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	Mode      string         `json:"mode,omitempty"`
	Lifetime  string         `json:"lifetime,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// EventEnvelope is the transport-neutral event delivered by connectivity
// routes and emitted on the internal event bus.
type EventEnvelope struct {
	ID        string         `json:"id"`
	Topic     string         `json:"topic"`
	Source    string         `json:"source,omitempty"`
	Transport string         `json:"transport"`
	Timestamp time.Time      `json:"timestamp"`
	SessionID string         `json:"session_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	RouteID   string         `json:"route_id,omitempty"`
	Payload   map[string]any `json:"payload"`
}

// RouteResponse is a generic route handler result. Request/response transports
// map it to native responses; event transports may ignore it.
type RouteResponse struct {
	Status  int                 `json:"status,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
	Events  []EventEnvelope     `json:"events,omitempty"`
}

// RouteHandler handles a normalized connectivity event.
type RouteHandler func(ctx context.Context, event EventEnvelope) (RouteResponse, error)
