package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
)

func TestMessagePageAPICompatibilityAndValidation(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "pages.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	e := turn.New(db)
	defer e.Close()
	s := New(db, e, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	ctx := context.Background()
	for _, id := range []string{"A", "B"} {
		if _, err = db.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 56; i++ {
		if err = db.AddMessage(ctx, fmt.Sprintf("m%03d", i), "A", "user", fmt.Sprint(i), nil); err != nil {
			t.Fatal(err)
		}
	}
	call := func(path string, want int) store.MessagePage {
		t.Helper()
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		var page store.MessagePage
		if want == 200 {
			if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
				t.Fatal(err)
			}
		}
		return page
	}
	if all := call("/api/sessions/A/messages", 200); len(all.Messages) != 56 {
		t.Fatal(len(all.Messages))
	}
	latest := call("/api/sessions/A/messages?limit=50", 200)
	if len(latest.Messages) != 50 || !latest.HasMore {
		t.Fatal(latest)
	}
	old := call("/api/sessions/A/messages?limit=50&before="+url.QueryEscape(latest.Before), 200)
	if len(old.Messages) != 6 || old.HasMore {
		t.Fatal(old)
	}
	next := call("/api/sessions/A/messages?limit=10&after="+url.QueryEscape(old.After), 200)
	if len(next.Messages) != 10 || !next.HasMore || next.Messages[0].ID != latest.Messages[0].ID {
		t.Fatal(next)
	}
	call("/api/sessions/B/messages?before="+url.QueryEscape(latest.Before), 400)
	call("/api/sessions/missing/messages?limit=50", 404)
	for _, q := range []string{"limit=0", "limit=101", "limit=abc", "before=not-a-cursor", "after=e30", "before=" + latest.Before + "&after=" + latest.After} {
		call("/api/sessions/A/messages?"+q, 400)
	}
	empty := call("/api/sessions/B/messages?limit=1", 200)
	if len(empty.Messages) != 0 || empty.HasMore || empty.After != "" {
		t.Fatal(empty)
	}
}
