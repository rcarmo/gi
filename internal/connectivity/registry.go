package connectivity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

type routeRecord struct {
	spec      RouteSpec
	handler   RouteHandler
	createdAt time.Time
}

// Registry owns route registrations and the shared event bus.
type Registry struct {
	mu     sync.RWMutex
	routes map[string]routeRecord
	bus    *EventBus
}

func NewRegistry() *Registry {
	return &Registry{routes: make(map[string]routeRecord), bus: NewEventBus()}
}

func (r *Registry) Bus() *EventBus { return r.bus }

func (r *Registry) Register(ctx context.Context, spec RouteSpec, handler RouteHandler) (RouteInfo, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return RouteInfo{}, fmt.Errorf("route name is required")
	}
	if strings.TrimSpace(spec.Transport) == "" {
		return RouteInfo{}, fmt.Errorf("route transport is required")
	}
	if handler == nil {
		handler = func(ctx context.Context, event EventEnvelope) (RouteResponse, error) {
			return RouteResponse{Status: 204}, nil
		}
	}
	if spec.ID == "" {
		spec.ID = newID("route")
	}
	if spec.Direction == "" {
		spec.Direction = "inbound"
	}
	if spec.Lifetime == "" {
		spec.Lifetime = "process"
	}
	if spec.Mode == "" {
		spec.Mode = "respond"
	}
	rec := routeRecord{spec: spec, handler: handler, createdAt: time.Now().UTC()}
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case <-ctx.Done():
		return RouteInfo{}, ctx.Err()
	default:
	}
	r.routes[spec.ID] = rec
	return rec.info(), nil
}

func (r *Registry) Unregister(_ context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("route id is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.routes[id]; !ok {
		return nil
	}
	delete(r.routes, id)
	return nil
}

func (r *Registry) List(_ context.Context, filter map[string]any) ([]RouteInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	infos := make([]RouteInfo, 0, len(r.routes))
	for _, rec := range r.routes {
		if !rec.matches(filter) {
			continue
		}
		infos = append(infos, rec.info())
	}
	return infos, nil
}

func (r *Registry) Get(id string) (RouteInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.routes[id]
	if !ok {
		return RouteInfo{}, false
	}
	return rec.info(), true
}

func (r *Registry) GetSpec(id string) (RouteSpec, RouteInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.routes[id]
	if !ok {
		return RouteSpec{}, RouteInfo{}, false
	}
	return rec.spec, rec.info(), true
}

func (r *Registry) Deliver(ctx context.Context, routeID string, event EventEnvelope) (RouteResponse, error) {
	r.mu.RLock()
	rec, ok := r.routes[routeID]
	r.mu.RUnlock()
	if !ok {
		return RouteResponse{}, fmt.Errorf("route not found: %s", routeID)
	}
	if event.ID == "" {
		event.ID = newID("evt")
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.RouteID == "" {
		event.RouteID = routeID
	}
	if event.Transport == "" {
		event.Transport = rec.spec.Transport
	}
	if event.Source == "" {
		event.Source = firstNonEmpty(rec.spec.Source, rec.spec.Name)
	}
	if event.SessionID == "" {
		event.SessionID = rec.spec.SessionID
	}
	if event.AgentID == "" {
		event.AgentID = rec.spec.AgentID
	}
	if event.Topic == "" {
		event.Topic = routeTopic(rec.spec)
	}
	if err := r.bus.Emit(ctx, event); err != nil {
		return RouteResponse{}, err
	}
	resp, err := rec.handler(ctx, event)
	for _, emitted := range resp.Events {
		if emitErr := r.bus.Emit(ctx, emitted); emitErr != nil {
			return resp, emitErr
		}
	}
	return resp, err
}

func (r *Registry) Emit(ctx context.Context, topic string, payload map[string]any) error {
	return r.bus.Emit(ctx, EventEnvelope{ID: newID("evt"), Topic: topic, Transport: "event", Timestamp: time.Now().UTC(), Payload: payload})
}

func (rec routeRecord) info() RouteInfo {
	return RouteInfo{ID: rec.spec.ID, Name: rec.spec.Name, Transport: rec.spec.Transport, Direction: rec.spec.Direction, Source: rec.spec.Source, Match: rec.spec.Match, SessionID: rec.spec.SessionID, AgentID: rec.spec.AgentID, Mode: rec.spec.Mode, Lifetime: rec.spec.Lifetime, CreatedAt: rec.createdAt}
}

func (rec routeRecord) matches(filter map[string]any) bool {
	if len(filter) == 0 {
		return true
	}
	info := rec.info()
	if v, ok := filter["transport"].(string); ok && v != "" && info.Transport != v {
		return false
	}
	if v, ok := filter["session_id"].(string); ok && v != "" && info.SessionID != v {
		return false
	}
	if v, ok := filter["name"].(string); ok && v != "" && info.Name != v {
		return false
	}
	return true
}

func routeTopic(spec RouteSpec) string {
	if topic, ok := spec.Match["topic"].(string); ok && topic != "" {
		return topic
	}
	return strings.Join([]string{"route", spec.Transport, spec.Name}, ".")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}

func storeID(prefix string, n uint64) string { return fmt.Sprintf("%s_%d", prefix, n) }
