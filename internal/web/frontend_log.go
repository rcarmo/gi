package web

import (
	"encoding/json"
	"fmt"
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
		payload := marshalFrontendLogDetail(entry.Detail)
		if entry.Level == "" {
			entry.Level = "info"
		}
		log.Printf("frontend[%s] %s detail=%s ts=%s", entry.Level, entry.Message, payload, entry.TS)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func marshalFrontendLogDetail(detail any) string {
	payload, err := json.Marshal(detail)
	if err == nil {
		return string(payload)
	}
	return fmt.Sprintf(`{"marshal_error":%q,"detail_type":%q}`, err.Error(), fmt.Sprintf("%T", detail))
}
