package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

func writeSSE(w http.ResponseWriter, eventType string, data any) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
}
