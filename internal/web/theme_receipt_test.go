package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestThemeCommandCommitsReceiptOnlyAfterPersistence(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_, err = s.CreateSession(context.Background(), "theme-receipt", "theme", nil)
	if err != nil {
		t.Fatal(err)
	}
	e := turn.New(s)
	defer e.Close()
	server := New(s, e, config.RuntimeConfig{})
	for _, command := range []string{"/theme dark", "/tint", "/tint invalid", "/theme default"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/sessions/theme-receipt/prompt", nil)
		if !server.handleThemeCommand(w, r, "theme-receipt", command) {
			t.Fatal("unhandled")
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || body["source_committed"] != true || body["ui_only"] != true {
			t.Fatalf("%s: %d %s", command, w.Code, w.Body.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/sessions/theme-receipt/prompt", nil).WithContext(ctx)
	server.handleThemeCommand(w, r, "theme-receipt", "/theme default")
	if w.Code == 200 {
		t.Fatalf("cancelled storage acknowledged: %s", w.Body.String())
	}
}
