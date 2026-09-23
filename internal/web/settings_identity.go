package web

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/rcarmo/gi/internal/config"
)

func (s *Server) identityPayload(saved config.IdentitySnapshot) map[string]any {
	active := config.IdentityNames{AssistantName: s.cfg.AssistantName, UserName: s.cfg.UserName}
	return map[string]any{"saved": saved, "active": active, "restart_required": saved.IdentityNames != active}
}

func (s *Server) handleSettingsIdentity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	switch r.Method {
	case http.MethodGet:
		saved, err := config.ReadIdentity(s.cfg.WorkspaceRoot)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "Cannot read saved identity; check .piclaw/config.json without replacing it."})
			return
		}
		writeJSON(w, 200, s.identityPayload(saved))
	case http.MethodPatch:
		// JSON-only and same-origin browser requests; native authenticated callers may
		// omit Origin. This is additional defence, not an authentication replacement.
		if err := http.NewCrossOriginProtection().Check(r); err != nil {
			writeJSON(w, 403, map[string]any{"error": "Cross-origin settings writes are not allowed"})
			return
		}
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if media != "application/json" {
			writeJSON(w, 415, map[string]any{"error": "JSON content type required"})
			return
		}
		var body struct {
			Revision      string `json:"revision"`
			AssistantName string `json:"assistant_name"`
			UserName      string `json:"user_name"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			writeJSON(w, 400, map[string]any{"error": "Invalid identity request"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, 400, map[string]any{"error": "Expected one JSON object"})
			return
		}
		saved, err := config.SaveIdentity(s.cfg.WorkspaceRoot, body.Revision, config.IdentityNames{AssistantName: body.AssistantName, UserName: body.UserName})
		if err != nil {
			status := 500
			message := "Cannot save identity; reload saved names to verify the file before retrying."
			if errors.Is(err, config.ErrIdentityConflict) {
				status = 409
				message = err.Error()
			}
			if errors.Is(err, config.ErrIdentityInvalid) {
				status = 400
				message = err.Error()
			}
			writeJSON(w, status, map[string]any{"error": message})
			return
		}
		writeJSON(w, 200, s.identityPayload(saved))
	default:
		w.Header().Set("Allow", "GET, PATCH")
		writeJSON(w, 405, map[string]any{"error": "Method not allowed"})
	}
}
