package web

import (
	"bytes"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleMediaLookup adapts native globally unique media IDs to the supplied UI.
// Like session media, it uses the instance's single-user authentication boundary.
func (s *Server) handleMediaLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/media/"), "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 || len(parts) > 2 || (len(parts) == 2 && parts[1] != "raw") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if len(parts) == 1 {
		media, err := s.store.GetMedia(r.Context(), id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/json")
			return
		}
		writeJSON(w, http.StatusOK, media)
		return
	}
	media, raw, err := s.store.GetMediaContent(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Do not execute uploaded HTML/SVG as same-origin documents. CSP also applies
	// to raster-labelled bytes; metadata alone is not a content trust boundary.
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	disposition := "attachment"
	switch strings.ToLower(media.ContentType) {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/avif", "image/bmp":
		disposition = "inline"
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": media.Filename}))
	w.Header().Set("Content-Type", media.ContentType)
	http.ServeContent(w, r, media.Filename, time.Time{}, bytes.NewReader(raw))
}
