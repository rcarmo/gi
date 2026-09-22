package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"path/filepath"
)

func TestQueueSteerAPIRequiresMatchingActiveRun(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "steer.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	engine := turn.New(db)
	defer engine.Close()
	s := New(db, engine, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	ctx := context.Background()
	if _, err := db.CreateSession(ctx, "A", "test", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateSession(ctx, "B", "test", nil); err != nil {
		t.Fatal(err)
	}
	for _, v := range []struct{ id, session, status string }{{"run", "A", "running"}, {"q", "A", "queued"}, {"other", "B", "queued"}} {
		if _, err := db.CreateTurnWithStatus(ctx, v.id, v.session, v.status, v.id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if ok, err := db.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	call := func(method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	for _, body := range []string{`{}`, `{"active_turn_id":"run","extra":1}`, `{"active_turn_id":"run"} {}`} {
		call(http.MethodPost, "/api/sessions/A/queue/q/steer", body, 400)
	}
	call(http.MethodPost, "/api/sessions/A/queue/q/steer", `{"active_turn_id":"stale"}`, 409)
	call(http.MethodPost, "/api/sessions/B/queue/q/steer", `{"active_turn_id":"run"}`, 409)
	call(http.MethodPost, "/api/sessions/A/queue/other/steer", `{"active_turn_id":"run"}`, 409)
	call(http.MethodPost, "/api/sessions/A/queue/q/steer", `{"active_turn_id":"run"}`, 200)
	call(http.MethodPost, "/api/sessions/A/queue/q/steer", `{"active_turn_id":"run"}`, 409)
	call(http.MethodDelete, "/api/sessions/A/queue/q", "", 409)
	// Simulate interrupted runner restart against the same durable database.
	if err := db.UpdateTurnStatusAndPhase(ctx, "run", "completed", "completed"); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseSessionActiveTurn(ctx, "A", "run"); err != nil {
		t.Fatal(err)
	}
	call(http.MethodPost, "/api/sessions/A/queue/q/steer", `{"active_turn_id":"run"}`, 409)
	var got struct {
		Items  []store.Turn `json:"items"`
		Active string       `json:"active_turn_id"`
	}
	if err := json.Unmarshal(call(http.MethodGet, "/api/sessions/A/queue", "", 200).Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Active != "" || len(got.Items) != 1 || got.Items[0].Phase != "steer_returned" {
		t.Fatal(fmt.Sprintf("%+v", got))
	}
	if started, err := engine.ContinueSession(ctx, "A"); err != nil || started {
		t.Fatal("unconsumed Steer auto-sent", started, err)
	}
	call(http.MethodDelete, "/api/sessions/A/queue/q", "", 200)
}
