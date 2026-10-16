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
		if got := fmt.Sprintf("%x", sha256.Sum256(w.Body.Bytes())); got != "21be2b5416b886d2404c99ba68abdf420369ded866a6e8b70866d5512b0918f5" {
			t.Fatal("pinned SDK changed", got)
		}
	}
}
