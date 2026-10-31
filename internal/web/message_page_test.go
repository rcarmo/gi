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
	"strings"
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

func TestConversationViewOptInPreservesRawMessageAPIs(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "conversation.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "conversation", "conversation", nil); err != nil {
		t.Fatal(err)
	}
	for i, m := range []struct {
		role, text string
		payload    map[string]any
	}{{"user", "question", nil}, {"tool_result", "private needle", nil}, {"assistant", "Visible reply\n[tool_call: shell]", map[string]any{"kind": "tool_calls"}}, {"system", "notice", nil}, {"assistant", "internal summary", map[string]any{"kind": "compaction"}}} {
		if err := s.AddMessage(ctx, fmt.Sprint(i), "conversation", m.role, m.text, m.payload); err != nil {
			t.Fatal(err)
		}
	}
	engine := turn.New(s)
	defer engine.Close()
	srv := New(s, engine, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	h := srv.Handler()
	for _, tc := range []struct {
		url           string
		count, status int
	}{{"/api/sessions/conversation/messages", 5, 200}, {"/api/sessions/conversation/messages?limit=50", 5, 200}, {"/api/sessions/conversation/messages?view=conversation", 3, 200}, {"/api/sessions/conversation/search?q=internal", 1, 200}, {"/api/sessions/conversation/search?q=internal&view=conversation", 0, 200}, {"/api/sessions/conversation/messages?view=conversation&limit=2", 2, 200}, {"/api/sessions/conversation/search?q=private", 1, 200}, {"/api/sessions/conversation/search?q=private&view=conversation", 0, 200}, {"/api/sessions/conversation/search?q=Visible&view=conversation", 1, 200}, {"/api/sessions/conversation/messages?view=unknown", 0, 400}, {"/api/sessions/conversation/search?q=Visible&view=unknown", 0, 400}} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest("GET", tc.url, nil))
		if rr.Code != tc.status {
			t.Fatal(tc.url, rr.Code, rr.Body.String())
		}
		if tc.status == 200 {
			var body struct {
				Messages []store.Message `json:"messages"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil || len(body.Messages) != tc.count {
				t.Fatal(tc.url, body, err)
			}
			if strings.Contains(tc.url, "view=conversation") {
				for _, m := range body.Messages {
					if m.Role == "tool_result" || m.Payload["kind"] == "compaction" || strings.Contains(m.Content, "[tool_call:") {
						t.Fatal("internal record leaked", tc.url, m)
					}
				}
			}
		}
	}
}
