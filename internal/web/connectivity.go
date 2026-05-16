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
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/topics"
)

const connectivityRoutePrefix = "/api/connect/routes/"
const connectivitySSEPrefix = "/api/connect/sse/"

func (s *Server) configureScriptConnectivity() {
	if s.scriptTool == nil || s.turns == nil || s.turns.Connectivity() == nil {
		return
	}
	s.scriptTool.SetConnectivityCallbacks(
		func(ctx context.Context, sessionID string, spec connectivity.RouteSpec) (connectivity.RouteInfo, error) {
			specSessionID := strings.TrimSpace(spec.SessionID)
			if specSessionID == "" {
				spec.SessionID = sessionID
			} else if strings.TrimSpace(sessionID) != "" && specSessionID != sessionID {
				return connectivity.RouteInfo{}, fmt.Errorf("register route: session %q does not match current session %q", specSessionID, sessionID)
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
			_, info, ok := s.turns.Connectivity().GetSpec(id)
			if !ok {
				return nil
			}
			if strings.TrimSpace(info.SessionID) != "" && strings.TrimSpace(sessionID) != "" && info.SessionID != sessionID {
				return fmt.Errorf("unregister route: route %q does not belong to session %q", id, sessionID)
			}
			return s.turns.Connectivity().Unregister(ctx, id)
		},
		func(ctx context.Context, sessionID string, filter map[string]any) ([]connectivity.RouteInfo, error) {
			normalizedFilter, err := normalizeSessionRouteFilter(sessionID, filter)
			if err != nil {
				return nil, err
			}
			return s.turns.Connectivity().List(ctx, normalizedFilter)
		},
		func(ctx context.Context, sessionID, topic string, payload map[string]any) error {
			return s.turns.Connectivity().Emit(ctx, topic, withSessionID(payload, sessionID))
		},
		func(ctx context.Context, sessionID string, envelope map[string]any) error {
			if s.turns == nil || s.turns.Topics() == nil {
				return fmt.Errorf("publish topic: topic bus not available")
			}
			topicName, _ := envelope["topic"].(string)
			topicName = strings.TrimSpace(topicName)
			if topicName == "" {
				return fmt.Errorf("publish topic: topic is required")
			}
			payload, _ := envelope["payload"].(map[string]any)
			if payload == nil {
				payload = map[string]any{}
			}
			agentID, _ := envelope["agent_id"].(string)
			source, _ := envelope["source"].(string)
			if strings.TrimSpace(source) == "" {
				source = "script"
			}
			typ, _ := envelope["type"].(string)
			sessionValue, _ := envelope["session_id"].(string)
			sessionValue = strings.TrimSpace(sessionValue)
			if sessionValue == "" {
				sessionValue = sessionID
			} else if strings.TrimSpace(sessionID) != "" && sessionValue != sessionID {
				return fmt.Errorf("publish topic: session %q does not match current session %q", sessionValue, sessionID)
			}
			s.turns.PublishTopicEvent(topics.Envelope{Topic: topicName, SessionID: sessionValue, AgentID: strings.TrimSpace(agentID), Source: source, Type: strings.TrimSpace(typ), Payload: payload})
			return nil
		},
		func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error) {
			if s.turns == nil || s.turns.Topics() == nil {
				return nil, nil, fmt.Errorf("subscribe topic: topic bus not available")
			}
			subSessionID := strings.TrimSpace(opts.SessionID)
			if subSessionID == "" {
				subSessionID = sessionID
			} else if strings.TrimSpace(sessionID) != "" && subSessionID != sessionID {
				return nil, nil, fmt.Errorf("subscribe topic: session %q does not match current session %q", subSessionID, sessionID)
			}
			subOpts := topics.SubscribeOptions{Buffer: opts.Buffer, SessionID: subSessionID, AgentID: strings.TrimSpace(opts.AgentID)}
			ch, unsubscribe := s.turns.Topics().Subscribe(ctx, pattern, subOpts)
			return ch, unsubscribe, nil
		},
	)
}

func cloneStringAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeSessionRouteFilter(sessionID string, filter map[string]any) (map[string]any, error) {
	normalized := cloneStringAnyMap(filter)
	if filterSessionID, _ := normalized["session_id"].(string); strings.TrimSpace(filterSessionID) == "" {
		normalized["session_id"] = sessionID
	} else if strings.TrimSpace(sessionID) != "" && filterSessionID != sessionID {
		return nil, fmt.Errorf("list routes: session filter %q does not match current session %q", filterSessionID, sessionID)
	}
	return normalized, nil
}

func withSessionID(payload map[string]any, sessionID string) map[string]any {
	normalized := cloneStringAnyMap(payload)
	normalized["session_id"] = sessionID
	return normalized
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
		writeJSON(w, connectivityAuthHTTPStatus(err), map[string]any{"error": err.Error()})
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
				return
			}
			if _, err := fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", ev.ID, ev.Topic, string(b)); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
