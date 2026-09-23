package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestStaticManifestAndIcons(t *testing.T) {
	st, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	srv := New(st, turn.New(st), config.RuntimeConfig{AssistantName: "Neo"})
	call := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		return rec
	}
	index := call(http.MethodGet, "/")
	if !strings.Contains(index.Body.String(), `rel="manifest" href="/manifest.json"`) {
		t.Fatalf("missing manifest link: %s", index.Body.String())
	}
	get := call(http.MethodGet, "/manifest.json")
	if get.Code != http.StatusOK || get.Header().Get("Content-Type") != "application/manifest+json; charset=utf-8" || get.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("manifest response: %d %+v", get.Code, get.Header())
	}
	head := call(http.MethodHead, "/manifest.json")
	if head.Code != http.StatusOK || head.Body.Len() != 0 || head.Header().Get("Content-Length") != get.Header().Get("Content-Length") {
		t.Fatalf("HEAD response: %d %+v body=%q", head.Code, head.Header(), head.Body.String())
	}
	var manifest struct {
		Name      string `json:"name"`
		ShortName string `json:"short_name"`
		StartURL  string `json:"start_url"`
		Display   string `json:"display"`
		Icons     []struct {
			Src     string `json:"src"`
			Sizes   string `json:"sizes"`
			Type    string `json:"type"`
			Purpose string `json:"purpose"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "Neo" || manifest.ShortName != "Neo" || manifest.StartURL != "/" || manifest.Display != "standalone" || len(manifest.Icons) != 4 {
		t.Fatalf("manifest: %+v", manifest)
	}
	for _, size := range []string{"192", "512"} {
		for _, purpose := range []string{"any", "maskable"} {
			found := false
			for _, icon := range manifest.Icons {
				if icon.Src == "/static/icon-"+size+".png" && icon.Sizes == size+"x"+size && icon.Type == "image/png" && icon.Purpose == purpose {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing icon %s %s: %+v", size, purpose, manifest.Icons)
			}
		}
		icon := call(http.MethodGet, "/static/icon-"+size+".png")
		if icon.Code != http.StatusOK || !bytes.HasPrefix(icon.Body.Bytes(), []byte("\x89PNG\r\n\x1a\n")) {
			t.Fatalf("icon %s: %d, %q", size, icon.Code, icon.Body.Bytes()[:min(icon.Body.Len(), 16)])
		}
	}
	if wrong := call(http.MethodPost, "/manifest.json"); wrong.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST manifest: %d", wrong.Code)
	}
}
