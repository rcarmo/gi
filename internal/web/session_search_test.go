package web

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionSearchScopeLiteralAndBounds(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "search.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	e := turn.New(db)
	defer e.Close()
	server := New(db, e, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	ctx := context.Background()
	for _, id := range []string{"root", "child", "sibling", "outside"} {
		if _, err = db.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
		if err = db.AddMessage(ctx, "m-"+id, id, "user", "Needle 100% _ literal "+id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.DB().Exec(`update sessions set parent_session_id='root' where id in ('child','sibling')`); err != nil {
		t.Fatal(err)
	}
	call := func(path string, want int) []store.Message {
		t.Helper()
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
		var out struct {
			Messages []store.Message `json:"messages"`
		}
		if want == 200 {
			if err = json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
		}
		return out.Messages
	}
	for scope, count := range map[string]int{"current": 1, "root": 3, "all": 4} {
		items := call("/api/sessions/child/search?q=NEEDLE&scope="+scope, 200)
		if len(items) != count {
			t.Fatal(scope, items)
		}
	}
	if items := call("/api/sessions/child/search?q="+url.QueryEscape("100% _")+"&scope=root&limit=1&offset=1", 200); len(items) != 1 {
		t.Fatal(items)
	}
	if items := call("/api/sessions/child/search?q="+url.QueryEscape("%' OR 1=1 --")+"&scope=all", 200); len(items) != 0 {
		t.Fatal(items)
	}
	for _, q := range []string{"q=", "q=x&scope=foreign", "q=x&limit=0", "q=x&limit=101", "q=x&offset=-1", "q=x&offset=10001", "q=x&limit=abc", "q=" + strings.Repeat("x", 513)} {
		call("/api/sessions/child/search?"+q, 400)
	}
	call("/api/sessions/missing/search?q=needle", 404)
	// Corrupt ancestry cycles terminate (with no invented root family).
	if _, err = db.DB().Exec(`update sessions set parent_session_id='child' where id='root'`); err != nil {
		t.Fatal(err)
	}
	if items := call("/api/sessions/child/search?q=needle&scope=root", 200); len(items) != 0 {
		t.Fatal(items)
	}
}
