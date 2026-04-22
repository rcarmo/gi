package web

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) handleFrontendLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Entries []struct {
			TS      string `json:"ts"`
			Level   string `json:"level"`
			Message string `json:"message"`
			Detail  any    `json:"detail"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	for _, entry := range req.Entries {
		payload, _ := json.Marshal(entry.Detail)
		if entry.Level == "" {
			entry.Level = "info"
		}
		log.Printf("frontend[%s] %s detail=%s ts=%s", entry.Level, entry.Message, string(payload), entry.TS)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
