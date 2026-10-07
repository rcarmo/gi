package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

func TestStaticViewerPagesMatchPinnedPolicies(t *testing.T) {
	raw, err := os.ReadFile("../../references/fixtures-vibes/ui/classic/piclaw/viewers-3.3.0/csp.json")
	if err != nil {
		t.Fatal(err)
	}
	var policies map[string]string
	if err := json.Unmarshal(raw, &policies); err != nil {
		t.Fatal(err)
	}
	if len(policies) != len(staticViewerPolicies) {
		t.Fatal("viewer policy count changed")
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	srv := New(s, nil, config.RuntimeConfig{})
	for path, policy := range policies {
		for _, prefix := range []string{"", "/static"} {
			for _, suffix := range []string{"", "index.html"} {
				for _, method := range []string{http.MethodGet, http.MethodHead} {
					res := httptest.NewRecorder()
					req := httptest.NewRequest(method, prefix+path+suffix, nil)
					req.Header.Set("Accept-Encoding", "br, gzip")
					srv.Handler().ServeHTTP(res, req)
					// The precompressed handler serves explicit index files directly.
					if res.Code != http.StatusOK || res.Header().Get("Content-Security-Policy") != policy || res.Header().Get("X-Frame-Options") != "SAMEORIGIN" || res.Header().Get("Cache-Control") != "no-cache" {
						t.Fatalf("%s %s: status=%d headers=%v", method, req.URL, res.Code, res.Header())
					}
					if method == http.MethodHead && res.Body.Len() != 0 {
						t.Fatal("HEAD returned body")
					}
					if suffix == "" && !strings.Contains(res.Header().Get("Content-Type"), "text/html") {
						t.Fatal("viewer did not serve HTML", res.Header())
					}
				}
			}
		}
	}
	for _, path := range []string{"/dist/app.bundle.js", "/image-viewer/missing", "/image-viewer-extra/"} {
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Header().Get("Content-Security-Policy") != "" {
			t.Fatal("viewer policy escaped exact paths", path)
		}
		if path != "/dist/app.bundle.js" && res.Code != http.StatusNotFound {
			t.Fatal("missing asset used app fallback", path, res.Code)
		}
	}
}
