package web

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
)

func TestAdaptiveCardSDKPinnedStaticAsset(t *testing.T) {
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, httptest.NewRequest(method, "/static/js/vendor/adaptivecards.min.js", nil))
		if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Type"), "javascript") {
			t.Fatalf("%s status %d type %s", method, w.Code, w.Header().Get("Content-Type"))
		}
		if method == http.MethodHead {
			if w.Body.Len() != 0 {
				t.Fatal("HEAD returned body")
			}
			continue
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(w.Body.Bytes())); got != "c3404b6cba966546c03ab9f81b133a2b8d9185da29806457d0adb92c107f799c" {
			t.Fatal("pinned SDK changed", got)
		}
	}
}
