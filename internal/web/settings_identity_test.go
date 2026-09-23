package web

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSettingsIdentityNativePersistenceAndFailures(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".piclaw"), 0700)
	path := filepath.Join(root, ".piclaw", "config.json")
	raw := []byte(`{"assistant":{"assistantName":"Active","assistantAvatar":"keep.png"},"user":{"userName":"Reader"},"private":"do not disclose"}`)
	os.WriteFile(path, raw, 0600)
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	cfg := config.RuntimeConfig{WorkspaceRoot: root, AssistantName: "Active", UserName: "Reader"}
	srv := New(st, turn.New(st), cfg)
	call := func(method, body, origin, contentType string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, "/api/settings/identity", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", contentType)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, req)
		return res
	}
	get := call("GET", "", "", "")
	if get.Code != 200 {
		t.Fatalf("get %d %s", get.Code, get.Body.String())
	}
	if bytes.Contains(get.Body.Bytes(), []byte("do not disclose")) || bytes.Contains(get.Body.Bytes(), []byte("keep.png")) {
		t.Fatal("config data leaked")
	}
	var snap struct {
		Saved   config.IdentitySnapshot `json:"saved"`
		Active  config.IdentityNames    `json:"active"`
		Restart bool                    `json:"restart_required"`
	}
	json.Unmarshal(get.Body.Bytes(), &snap)
	if snap.Restart || snap.Active.AssistantName != "Active" {
		t.Fatalf("snapshot %+v", snap)
	}
	body, _ := json.Marshal(map[string]any{"revision": snap.Saved.Revision, "assistant_name": "Saved", "user_name": "User Two"})
	saved := call("PATCH", string(body), "http://example.com", "application/json")
	if saved.Code != 200 {
		t.Fatalf("save %d %s", saved.Code, saved.Body.String())
	}
	json.Unmarshal(saved.Body.Bytes(), &snap)
	if !snap.Restart || snap.Saved.AssistantName != "Saved" || snap.Active.AssistantName != "Active" {
		t.Fatalf("saved %+v", snap)
	}
	if cfg.AssistantName != "Active" || config.Load(root).AssistantName != "Saved" {
		t.Fatal("activation semantics incorrect")
	}
	if stale := call("PATCH", string(body), "", "application/json"); stale.Code != 409 {
		t.Fatalf("stale %d", stale.Code)
	}
	for _, tc := range []struct {
		body, origin, ct string
		code             int
	}{
		{`bad`, `http://elsewhere.invalid`, `application/json`, 403},
		{`{}`, ``, `text/plain`, 415},
		{`{"surprise":true}`, ``, `application/json`, 400},
		{`{} {}`, ``, `application/json`, 400},
		{`{"assistant_name":"","user_name":"x","revision":"anything"}`, ``, `application/json`, 400},
	} {
		res := call("PATCH", tc.body, tc.origin, tc.ct)
		if res.Code != tc.code {
			t.Fatalf("request %+v: %d %s", tc, res.Code, res.Body.String())
		}
	}
	before, _ := os.ReadFile(path)
	os.Mkdir(filepath.Join(root, ".piclaw", "blocked"), 0700)
	os.Remove(filepath.Join(root, ".piclaw", ".gi-identity.lock"))
	os.Mkdir(filepath.Join(root, ".piclaw", ".gi-identity.lock"), 0700)
	body, _ = json.Marshal(map[string]any{"revision": snap.Saved.Revision, "assistant_name": "Never", "user_name": "Written"})
	failure := call("PATCH", string(body), "", "application/json")
	if failure.Code != 500 {
		t.Fatalf("failure %d", failure.Code)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed write changed file")
	}
	if method := call("POST", "", "", ""); method.Code != 405 {
		t.Fatal("wrong method allowed")
	}
	os.WriteFile(path, []byte("bad config"), 0600)
	if bad := call("GET", "", "", ""); bad.Code != 500 {
		t.Fatal("bad config silently defaulted")
	}
}
