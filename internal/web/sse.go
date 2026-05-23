package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/topics"
)

const maxTopicSSEBuffer = 1024

func (s *Server) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	chatJid := r.URL.Query().Get("chat_jid")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send connected event
	writeSSE(w, "connected", map[string]any{
		"app_asset_version": s.version,
		"chat_jid":          chatJid,
	})
	flusher.Flush()

	// Extract session ID from chat_jid (gi:session_xxx -> session_xxx)
	sessionID := ""
	if len(chatJid) > 3 && chatJid[:3] == "gi:" {
		sessionID = chatJid[3:]
	}

	// Subscribe to turn events
	var ch chan map[string]any
	if sessionID != "" {
		ch = s.turns.Subscribe(sessionID)
		defer s.turns.Unsubscribe(sessionID, ch)
	} else {
		ch = make(chan map[string]any)
	}

	ctx := r.Context()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			writeSSE(w, "heartbeat", map[string]any{"ts": time.Now().UnixMilli()})
			flusher.Flush()
		case ev, ok := <-ch:
			if !ok {
				return
			}
			eventType, _ := ev["type"].(string)
			if eventType == "" {
				eventType = "message"
			}
			writeSSE(w, eventType, ev)
			flusher.Flush()
		}
	}
}

func (s *Server) handleTopicSSE(w http.ResponseWriter, r *http.Request) {
	if s.turns == nil || s.turns.Topics() == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "topic bus not available"})
		return
	}
	pattern := strings.TrimSpace(r.URL.Query().Get("topic"))
	if pattern == "" {
		pattern = "*"
	}
	buffer := 64
	if raw := strings.TrimSpace(r.URL.Query().Get("buffer")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > maxTopicSSEBuffer {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid buffer"})
			return
		}
		buffer = parsed
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	opts := topics.SubscribeOptions{Buffer: buffer, SessionID: strings.TrimSpace(r.URL.Query().Get("session_id")), AgentID: strings.TrimSpace(r.URL.Query().Get("agent_id"))}
	ch, unsubscribe := s.turns.Topics().Subscribe(r.Context(), pattern, opts)
	defer unsubscribe()
	lastSequence := s.turns.Topics().LastSequence()
	sinceSequence := topicSSELastEventID(r)
	connected := map[string]any{"topic": pattern, "session_id": opts.SessionID, "agent_id": opts.AgentID, "app_asset_version": s.version, "last_sequence": lastSequence}
	if sinceSequence > 0 {
		connected["last_event_id"] = sinceSequence
		if lastSequence > sinceSequence {
			connected["missed_sequence_count"] = lastSequence - sinceSequence
		}
	}
	if err := writeSSE(w, "connected", connected); err != nil {
		return
	}
	flusher.Flush()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if err := writeSSE(w, "heartbeat", map[string]any{"ts": time.Now().UnixMilli(), "topic": pattern}); err != nil {
				return
			}
			flusher.Flush()
		case env, ok := <-ch:
			if !ok {
				return
			}
			eventName := strings.TrimSpace(env.Topic)
			if env.Sequence > 0 {
				if _, err := fmt.Fprintf(w, "id: %d\n", env.Sequence); err != nil {
					return
				}
			}
			if err := writeSSE(w, eventName, env); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func topicSSELastEventID(r *http.Request) uint64 {
	if r == nil {
		return 0
	}
	value := strings.TrimSpace(r.Header.Get("Last-Event-ID"))
	if value == "" {
		value = strings.TrimSpace(r.URL.Query().Get("last_event_id"))
	}
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func writeSSE(w http.ResponseWriter, eventType string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
	return err
}
