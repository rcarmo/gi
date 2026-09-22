package web

import (
	"context"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestManualCompactionHTTPGuards(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "manual.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	e := turn.New(db)
	defer e.Close()
	s := New(db, e, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	ctx := context.Background()
	db.CreateSession(ctx, "A", "A", nil)
	call := func(method, path, body string, want int) {
		t.Helper()
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("got %d want %d %s", w.Code, want, w.Body.String())
		}
	}
	call("GET", "/api/sessions/A/compaction", "", 200)
	call("GET", "/api/sessions/missing/compaction", "", 404)
	call("PUT", "/api/sessions/A/compaction", "", 405)
	for _, body := range []string{`{}`, `{"token":"x","extra":1}`, `{"token":"x"} {}`} {
		call("POST", "/api/sessions/A/compaction", body, 400)
	}
	call("POST", "/api/sessions/A/compaction", `{"token":"x"}`, 409)
	db.AddMessage(ctx, "one", "A", "user", "one", nil)
	db.AddMessage(ctx, "two", "A", "assistant", "two", nil)
	call("POST", "/api/sessions/A/compaction", `{"token":"stale"}`, 409)
	turns, err := db.ListTurns(ctx, "A")
	if err != nil || len(turns) != 0 {
		t.Fatal(turns, err)
	}
}
