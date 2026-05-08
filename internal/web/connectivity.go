package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/tools"
)

const connectivityRoutePrefix = "/api/connect/routes/"
const connectivitySSEPrefix = "/api/connect/sse/"

func (s *Server) configureScriptConnectivity() {
	if s.scriptTool == nil || s.turns == nil || s.turns.Connectivity() == nil {
		return
	}
	s.scriptTool.SetConnectivityCallbacks(
		func(ctx context.Context, sessionID string, spec connectivity.RouteSpec) (connectivity.RouteInfo, error) {
			if spec.SessionID == "" {
				spec.SessionID = sessionID
			}
			return s.turns.Connectivity().Register(ctx, spec, func(ctx context.Context, event connectivity.EventEnvelope) (connectivity.RouteResponse, error) {
				payload := map[string]any{"event": event, "route": spec, "payload": event.Payload}
				input := tools.ScriptInput{Engine: spec.Engine, Path: spec.Path, SessionID: event.SessionID, Script: scriptWithPayload(spec.Engine, "event", payload, spec.Script)}
				out := s.scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return connectivity.RouteResponse{Status: http.StatusInternalServerError, Body: out.Result}, errors.New(out.Error)
				}
				resp := connectivity.RouteResponse{Status: http.StatusOK, Body: out.Result}
				if strings.TrimSpace(out.Result) != "" {
					if err := json.Unmarshal([]byte(out.Result), &resp); err != nil {
						resp = connectivity.RouteResponse{Status: http.StatusOK, Body: out.Result}
					} else if resp.Status == 0 && resp.Body == "" {
						resp.Body = out.Result
					}
				}
				if resp.Status == 0 {
					resp.Status = http.StatusOK
				}
				return resp, nil
			})
		},
		func(ctx context.Context, sessionID, id string) error {
			return s.turns.Connectivity().Unregister(ctx, id)
		},
		func(ctx context.Context, sessionID string, filter map[string]any) ([]connectivity.RouteInfo, error) {
			return s.turns.Connectivity().List(ctx, filter)
		},
		func(ctx context.Context, sessionID, topic string, payload map[string]any) error {
			if payload == nil {
				payload = map[string]any{}
			}
			payload["session_id"] = sessionID
			return s.turns.Connectivity().Emit(ctx, topic, payload)
		},
	)
}

func scriptWithPayload(engine, name string, payload map[string]any, script string) string {
	if strings.TrimSpace(script) == "" {
		return script
	}
	b, err := json.Marshal(payload)
	if err != nil {
		b = []byte("{}")
	}
	if engine == "joker" || engine == "" && strings.HasPrefix(strings.TrimSpace(script), "(") {
		return fmt.Sprintf("(def *gi-%s* (json/read-string %q))\n%s", name, string(b), script)
	}
	return fmt.Sprintf("gi.%s = %s; gi.%sPayload = gi.%s; gi.toolArgs = (gi.tool && gi.tool.arguments) || {};\n%s", name, string(b), name, name, script)
}

func (s *Server) handleConnectivityRoutes(w http.ResponseWriter, r *http.Request) {
	if s.turns == nil || s.turns.Connectivity() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "connectivity registry not available"})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, connectivityRoutePrefix)
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}
	routeID := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	headers := map[string][]string{}
	for k, v := range r.Header {
		headers[k] = append([]string(nil), v...)
	}
	spec, _, ok := s.turns.Connectivity().GetSpec(routeID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "route not found: " + routeID})
		return
	}
	if err := s.authorizeConnectivityRequest(spec, r, body); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}
	payload := map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"route_path":  rest,
		"query":       r.URL.Query(),
		"headers":     headers,
		"body":        string(body),
		"remote_addr": r.RemoteAddr,
	}
	event := connectivity.EventEnvelope{Transport: "http", Payload: payload}
	resp, err := s.turns.Connectivity().Deliver(r.Context(), routeID, event)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error(), "body": resp.Body})
		return
	}
	for k, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(k, value)
		}
	}
	status := resp.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	if _, err := w.Write([]byte(resp.Body)); err != nil {
		log.Printf("connectivity route write response: %v", err)
	}
}

func (s *Server) handleConnectivitySSE(w http.ResponseWriter, r *http.Request) {
	if s.turns == nil || s.turns.Connectivity() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "connectivity registry not available"})
		return
	}
	pattern := strings.TrimPrefix(r.URL.Path, connectivitySSEPrefix)
	pattern = strings.Trim(pattern, "/")
	if pattern == "" {
		pattern = r.URL.Query().Get("topic")
	}
	if pattern == "" {
		pattern = "*"
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ch, unsubscribe := s.turns.Connectivity().Bus().Subscribe(r.Context(), pattern, 64)
	defer unsubscribe()
	if _, err := fmt.Fprintf(w, ": connected %s\n\n", time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return
	}
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			b, err := json.Marshal(ev)
			if err != nil {
				log.Printf("connectivity sse marshal event: %v", err)
				continue
			}
			if _, err := fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", ev.ID, ev.Topic, string(b)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
