package web

import (
	"net/http"
	"strings"
)

// Policies match the extracted Piclaw v3.3.0 viewers in the pinned UI. Tests
// compare these values with the upstream extraction manifest.
var staticViewerPolicies = map[string]string{
	"/html-viewer/":  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: https: http:; connect-src 'self'; frame-src 'self' blob:; frame-ancestors 'self'; base-uri 'self'; form-action 'self'",
	"/image-viewer/": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; frame-src 'self' blob:; frame-ancestors 'self'; base-uri 'self'; form-action 'self'",
	"/video-viewer/": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; media-src 'self' data: blob:; connect-src 'self'; frame-src 'self' blob:; frame-ancestors 'self'; base-uri 'self'; form-action 'self'",
	"/pdf-viewer/":   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; frame-src 'self' blob:; frame-ancestors 'self'; base-uri 'self'; form-action 'self'",
	"/data-viewer/":  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; frame-src 'self'; frame-ancestors 'self'; base-uri 'self'; form-action 'self'",
}

func withStaticViewerPolicies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "index.html")
		if policy, ok := staticViewerPolicies[path]; ok {
			w.Header().Set("Content-Security-Policy", policy)
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
