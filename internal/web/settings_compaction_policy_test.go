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

func TestSettingsCompactionPolicyPersistenceAndGuards(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".pi"), 0700)
	path := filepath.Join(root, ".pi", "settings.json")
	os.WriteFile(path, []byte(`{"defaultModel":"test-model","private":"never-return","compaction":{"enabled":true}}`), 0600)
	cfg := config.Load(root)
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	srv := New(st, turn.NewWithRuntimeConfig(st, cfg, "test-model"), cfg)
	call := func(method, body, origin, ct string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/settings/compaction", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", ct)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res := httptest.NewRecorder()
		srv.Handler().ServeHTTP(res, req)
		return res
	}
	get := call("GET", "", "", "")
	if get.Code != 200 || get.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(get.Code)
	}
	if bytes.Contains(get.Body.Bytes(), []byte("never-return")) {
		t.Fatal("unrelated config leak")
	}
	var state struct {
		Saved   config.CompactionPolicySnapshot `json:"saved"`
		Active  config.CompactionSettings       `json:"active"`
		Restart bool                            `json:"restart_required"`
	}
	json.Unmarshal(get.Body.Bytes(), &state)
	body, _ := json.Marshal(map[string]any{"revision": state.Saved.Revision, "enabled": false, "context_window": 64000, "reserve_tokens": 4000, "keep_recent_tokens": 8000, "threshold_tokens": 50000})
	save := call("PATCH", string(body), "http://example.com", "application/json")
	if save.Code != 200 {
		t.Fatalf("save %d %s", save.Code, save.Body.String())
	}
	json.Unmarshal(save.Body.Bytes(), &state)
	if !state.Restart || state.Active != cfg.Compaction || state.Saved.Policy.Enabled || config.Load(root).Compaction != state.Saved.Policy {
		t.Fatalf("state %+v", state)
	}
	if stale := call("PATCH", string(body), "", "application/json"); stale.Code != 409 {
		t.Fatal(stale.Code)
	}
	for _, tc := range []struct {
		body, origin, ct string
		code             int
	}{
		{"bad", "http://elsewhere.invalid", "application/json", 403}, {"{}", "", "text/plain", 415}, {"{}", "", "application/json", 400},
		{`{"enabled":true,"strategy":"injected"}`, "", "application/json", 400},
		{`{"enabled":true,"context_window":1.5}`, "", "application/json", 400}, {`{} {}`, "", "application/json", 400},
		{`{"revision":"x","enabled":true,"context_window":5,"reserve_tokens":5,"keep_recent_tokens":1,"threshold_tokens":1}`, "", "application/json", 400},
	} {
		res := call("PATCH", tc.body, tc.origin, tc.ct)
		if res.Code != tc.code {
			t.Fatalf("%+v => %d %s", tc, res.Code, res.Body.String())
		}
	}
	if call("PUT", "", "", "").Code != 405 {
		t.Fatal("method accepted")
	}
	before, _ := os.ReadFile(path)
	os.Remove(filepath.Join(root, ".pi", ".gi-settings.lock"))
	os.Mkdir(filepath.Join(root, ".pi", ".gi-settings.lock"), 0700)
	var payload map[string]any
	json.Unmarshal(body, &payload)
	payload["revision"] = state.Saved.Revision
	body, _ = json.Marshal(payload)
	if res := call("PATCH", string(body), "", "application/json"); res.Code != 500 {
		t.Fatal(res.Code)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("failed save changed settings")
	}
	os.WriteFile(path, []byte("bad json"), 0600)
	if call("GET", "", "", "").Code != 500 {
		t.Fatal("corrupt config accepted")
	}
}
