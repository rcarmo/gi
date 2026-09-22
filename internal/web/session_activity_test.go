package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionActivityNativeOccurrenceAndCancelGuards(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "activity.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	e := turn.New(db)
	defer e.Close()
	s := New(db, e, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	ctx := context.Background()
	for _, id := range []string{"A", "B"} {
		if _, err := db.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	call := func(method, path, body string, want int) map[string]any {
		t.Helper()
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s got %d %s", path, w.Code, w.Body.String())
		}
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	if call("GET", "/api/sessions/A/activity", "", 200)["status"] != "idle" {
		t.Fatal("not idle")
	}
	call("GET", "/api/sessions/missing/activity", "", 404)
	call("POST", "/api/sessions/A/activity", `{}`, 400)
	call("POST", "/api/sessions/A/activity", `{"turn_id":"run","unknown":1}`, 400)
	db.CreateTurn(ctx, "run", "A", "prompt", nil)
	db.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run")
	seq, err := db.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := call("GET", "/api/sessions/A/activity", "", 200)
	c := snapshot["compaction"].(map[string]any)
	if snapshot["turn_id"] != "run" || c["active"] != true || c["seq"] != float64(seq) {
		t.Fatal(snapshot)
	}
	call("POST", "/api/sessions/B/activity", `{"turn_id":"run"}`, 409)
	call("POST", "/api/sessions/A/activity", `{"turn_id":"stale"}`, 409)
	// A stale persisted claim without a local runner must not cancel queued work.
	call("POST", "/api/sessions/A/activity", `{"turn_id":"run"}`, 409)
	if _, err := db.FinishCompaction(ctx, "A", "run", seq, "suppressed", "", map[string]any{"detail": "hook policy"}); err != nil {
		t.Fatal(err)
	}
	db.UpdateTurnStatusAndPhase(ctx, "run", "completed", "completed")
	db.ReleaseSessionActiveTurn(ctx, "A", "run")
	snapshot = call("GET", "/api/sessions/A/activity", "", 200)
	c = snapshot["compaction"].(map[string]any)
	if snapshot["status"] != "idle" || c["active"] != false || c["detail"] != "hook policy" {
		t.Fatal(snapshot)
	}
	if call("GET", "/api/sessions/B/activity", "", 200)["compaction"] != nil {
		t.Fatal("foreign activity leak")
	}
	call("POST", "/api/sessions/A/activity", `{"turn_id":"run"}`, 409)
	// A newer turn hides prior suppression; unclaimed running rows never imply
	// active work after a restart/lost lease.
	if _, err := db.CreateTurn(ctx, "new", "A", "newer", nil); err != nil {
		t.Fatal(err)
	}
	snapshot = call("GET", "/api/sessions/A/activity", "", 200)
	if snapshot["status"] != "idle" || snapshot["compaction"] != nil {
		t.Fatal(snapshot)
	}
}
