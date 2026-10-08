package web

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/plan"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestSessionPlanHTTPValidationAuthIsolationAndSSE(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "plans.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.CreateSession(t.Context(), "s", "s", nil)
	db.CreateSession(t.Context(), "other", "other", nil)
	engine := turn.New(db)
	defer engine.Close()
	events := engine.Subscribe("s")
	defer engine.Unsubscribe("s", events)
	server := New(db, engine, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	call := func(method, path, body string) *httptest.ResponseRecorder {
		if method == "POST" {
			var input map[string]any
			if json.Unmarshal([]byte(body), &input) == nil {
				loaded := httptest.NewRecorder()
				server.Handler().ServeHTTP(loaded, httptest.NewRequest("GET", "http://127.0.0.1"+path, nil))
				var response struct {
					Plan store.SessionPlan `json:"plan"`
				}
				json.Unmarshal(loaded.Body.Bytes(), &response)
				input["expected_revision"] = response.Plan.Revision
				raw, _ := json.Marshal(input)
				body = string(raw)
			}
		}
		r := httptest.NewRequest(method, "http://127.0.0.1"+path, strings.NewReader(body))
		r.RemoteAddr = "127.0.0.1:12345"
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		return w
	}
	r := call("GET", "/api/sessions/s/plan", "")
	if r.Code != 200 || !strings.Contains(r.Body.String(), "Clarify the current objective") {
		t.Fatal(r.Code, r.Body.String())
	}
	r = call("POST", "/api/sessions/s/plan", `{"markdown":"## Heading\n- [ ] pending\n- [-] doing\n- [x] done"}`)
	if r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	var response struct {
		OK   bool              `json:"ok"`
		Plan store.SessionPlan `json:"plan"`
	}
	json.Unmarshal(r.Body.Bytes(), &response)
	if !response.OK || response.Plan.ChatJID != "gi:s" || len(response.Plan.Plan) != 3 {
		t.Fatal(response)
	}
	select {
	case event := <-events:
		if event["type"] != "extension_ui_status" || event["key"] != "plan.changes" || event["chat_jid"] != "gi:s" || event["source"] != "api" {
			t.Fatal(event)
		}
	case <-time.After(time.Second):
		t.Fatal("missing event")
	}
	for _, tc := range []struct {
		method, path, body string
		code               int
	}{
		{"POST", "/api/sessions/s/plan", `{"markdown":"- [-] one\n- [-] two"}`, 400},
		{"POST", "/api/sessions/s/plan", `{"markdown":null}`, 400},
		{"POST", "/api/sessions/s/plan", `{"markdown":""} {}`, 400},
		{"POST", "/api/sessions/s/plan", `{"action":"patch"}`, 400},
		{"POST", "/api/sessions/s/plan", `{"chat_jid":"gi:other","markdown":"oops"}`, 400},
		{"POST", "/api/sessions/missing/plan", `{"action":"reset"}`, 404},
		{"GET", "/api/sessions/missing/plan", "", 404},
		{"DELETE", "/api/sessions/s/plan", "", 405},
		{"GET", "/api/sessions/s/plan/extra", "", 404},
	} {
		if r = call(tc.method, tc.path, tc.body); r.Code != tc.code {
			t.Fatal(tc, r.Code, r.Body.String())
		}
	}
	unchanged, _ := db.SessionPlan(t.Context(), "s")
	if unchanged.Markdown != response.Plan.Markdown {
		t.Fatal("invalid changed saved plan", unchanged)
	}
	other, _ := db.SessionPlan(t.Context(), "other")
	if other.Markdown != plan.DefaultMarkdown {
		t.Fatal("cross-session plan changed")
	}
	select {
	case e := <-events:
		t.Fatal("invalid plan published", e)
	default:
	}
	r = call("POST", "/api/sessions/s/plan", `{"action":"reset"}`)
	if r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	pending, err := server.auth.StartEnrollment("rui")
	if err != nil {
		t.Fatal(err)
	}
	server.auth.VerifyEnrollment("rui", webTestTOTPCode(pending.Secret))
	if r = call("GET", "/api/sessions/s/plan", ""); r.Code != 401 {
		t.Fatal("public plan", r.Code)
	}
	if r = call("POST", "/api/sessions/s/plan", `{"action":"reset"}`); r.Code != 401 {
		t.Fatal("public write", r.Code)
	}
}

func TestSessionPlanHTTPRequiresLoadedRevision(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "revision.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.CreateSession(t.Context(), "s", "s", nil)
	s := New(db, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	call := func(body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "http://localhost/api/sessions/s/plan", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	loaded, _ := db.SessionPlan(t.Context(), "s")
	if w := call(`{"action":"reset"}`); w.Code != 428 || !strings.Contains(w.Body.String(), "revision_required") {
		t.Fatal(w.Code, w.Body.String())
	}
	text := "- [ ] agent"
	db.MutateSessionPlan(t.Context(), "s", plan.Mutation{Action: "write", Markdown: &text})
	body, _ := json.Marshal(map[string]any{"action": "reset", "expected_revision": loaded.Revision})
	if w := call(string(body)); w.Code != 409 || !strings.Contains(w.Body.String(), "revision_conflict") {
		t.Fatal(w.Code, w.Body.String())
	}
	current, _ := db.SessionPlan(t.Context(), "s")
	if current.Markdown != text {
		t.Fatal("stale reset overwrote plan")
	}
	body, _ = json.Marshal(map[string]any{"action": "reset", "expected_revision": current.Revision})
	if w := call(string(body)); w.Code != 200 || !strings.Contains(w.Body.String(), `"revision":"plan-v1-`) {
		t.Fatal(w.Code, w.Body.String())
	}
}
