package web

import (
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
)

// Exercise the real embedded HTTP graph. The stable entry may be requested by
// an old page or a same-version rebuild; it must never be served from cache.
func TestSettingsModuleHTTPGraph(t *testing.T) {
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	srv.version = "settings-assets-test"
	fetch := func(method, url string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, httptest.NewRequest(method, url, nil))
		return w
	}
	for _, url := range []string{"/", "/index.html", "/dist/app.bundle.js", "/dist/app.bundle.js?v=old-version", "/dist/app.bundle.js.map"} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			w := fetch(method, url)
			if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
				t.Fatalf("%s %s: status=%d cache=%q", method, url, w.Code, w.Header().Get("Cache-Control"))
			}
		}
	}
	index := fetch(http.MethodGet, "/").Body.String()
	if !strings.Contains(index, `/dist/app.bundle.js?v=settings-assets-test`) {
		t.Fatal("HTML does not version the bootstrap")
	}
	bootstrap := fetch(http.MethodGet, "/dist/app.bundle.js?v=settings-assets-test").Body.String()
	app := regexp.MustCompile(`import\("(\./chunks/app-[a-z0-9]+\.js)"\)`).FindStringSubmatch(bootstrap)
	if len(app) != 2 {
		t.Fatal("bootstrap does not import a hashed app module")
	}
	// Follow actual import URLs (not just filesystem paths). FileServer must
	// serve JS, never a successful index.html fallback for missing modules.
	seen := map[string]bool{}
	imports := regexp.MustCompile(`(?:from\s*|import\s*\()"(\.[^"\n]+\.js)"`)
	var visit func(string)
	visit = func(url string) {
		t.Helper()
		if seen[url] {
			return
		}
		seen[url] = true
		w := fetch(http.MethodGet, url)
		if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Type"), "javascript") {
			t.Fatalf("module %s: status=%d type=%q", url, w.Code, w.Header().Get("Content-Type"))
		}
		for _, match := range imports.FindAllStringSubmatch(w.Body.String(), -1) {
			dependency := path.Join(path.Dir(url), match[1])
			if !strings.HasPrefix(dependency, "/dist/chunks/") {
				t.Fatalf("module %s imports outside hashed graph: %s", url, dependency)
			}
			visit(dependency)
		}
	}
	visit(path.Join("/dist", app[1]))
	for _, pane := range []string{"models", "appearance", "compaction", "providers", "authentication"} {
		count := 0
		for url := range seen {
			if strings.HasPrefix(url, "/dist/chunks/gi-settings-"+pane+"-") {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%s: expected one reachable module, got %d", pane, count)
		}
	}
	missing := fetch(http.MethodGet, "/dist/chunks/gi-settings-models-removed-build.js")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing old chunk: status %d", missing.Code)
	}
	// No Last-Modified/ETag exists for embedded files. Conditional reloads must
	// still deliver the current entry rather than a misleading old 304.
	req := httptest.NewRequest(http.MethodGet, "/dist/app.bundle.js?v=old-version", nil)
	req.Header.Set("If-None-Match", `"old-build"`)
	req.Header.Set("If-Modified-Since", "Thu, 01 Jan 2099 00:00:00 GMT")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != bootstrap {
		t.Fatalf("conditional bootstrap request: status %d", w.Code)
	}
}
