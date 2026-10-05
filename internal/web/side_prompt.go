package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleSidePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	if !providerWriteTransport(r) || !browserSameOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Side prompt requires same-origin HTTPS or localhost"})
		return
	}
	var req struct {
		Prompt string `json:"prompt"`
		System string `json:"system_prompt"`
		Chat   string `json:"chat_jid"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 96*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "Invalid side prompt JSON"})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "Expected one JSON object"})
		return
	}
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" || len(req.Prompt) > 64*1024 || len(req.System) > 16*1024 || !strings.HasPrefix(req.Chat, "gi:") || len(req.Chat) > 256 {
		writeJSON(w, 400, map[string]any{"error": "Prompt and session are required"})
		return
	}
	sessionID := strings.TrimPrefix(req.Chat, "gi:")
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": "Session not found"})
		return
	}
	// Bound one-off provider concurrency independently of the main queue.
	select {
	case s.sidePromptSlots <- struct{}{}:
		defer func() { <-s.sidePromptSlots }()
	default:
		writeJSON(w, 429, map[string]any{"error": "Side prompt concurrency limit reached"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	streaming := strings.HasSuffix(r.URL.Path, "/stream")
	flusher, ok := w.(http.Flusher)
	if streaming && !ok {
		writeJSON(w, 500, map[string]any{"error": "Streaming unavailable"})
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	send := func(event string, payload any) bool {
		_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := writeSSE(w, event, payload); err != nil {
			cancel()
			return false
		}
		flusher.Flush()
		return true
	}
	if streaming {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("X-Accel-Buffering", "no")
		if !send("side_prompt_start", map[string]any{"chat_jid": req.Chat}) {
			return
		}
	}
	result, err := s.turns.RunSidePrompt(ctx, sessionID, req.Prompt, req.System, func(kind, delta string) {
		if !streaming {
			return
		}
		event := "side_prompt_text_delta"
		if kind == "thinking_delta" {
			event = "side_prompt_thinking_delta"
		}
		send(event, map[string]any{"delta": delta})
	})
	if err != nil {
		payload := map[string]any{"status": "error", "result": nil, "thinking": nil, "model": nil, "error": err.Error()}
		if streaming {
			if r.Context().Err() == nil {
				send("side_prompt_error", payload)
			}
		} else {
			writeJSON(w, 502, payload)
		}
		return
	}
	if streaming {
		send("side_prompt_done", result)
	} else {
		writeJSON(w, 200, result)
	}
}
