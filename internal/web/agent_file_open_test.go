package web

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func fileOpenTestServer(t *testing.T) *Server {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file.md"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(root, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B"} {
		if _, err := db.CreateSession(t.Context(), id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	engine := turn.New(db)
	s := New(db, engine, config.RuntimeConfig{WorkspaceRoot: root})
	t.Cleanup(func() { s.CloseFileOpenRequests(); engine.Close(); db.Close() })
	return s
}
func fileOpenListenerFor(t *testing.T, s *Server, session string) <-chan map[string]any {
	t.Helper()
	r := httptest.NewRequest("GET", "http://localhost/sse/stream?chat_jid=gi:"+session, nil)
	r.RemoteAddr = "127.0.0.1:1234"
	ch, stop := s.subscribeFileOpen(r, session)
	t.Cleanup(stop)
	return ch
}
func fileOpenReply(s *Server, event map[string]any, chat, target, path string) *httptest.ResponseRecorder {
	options := event["options"].(map[string]any)
	if target == "" {
		target = options["target"].(string)
	}
	if path == "" {
		path = options["path"].(string)
	}
	if chat == "" {
		chat = event["chat_jid"].(string)
	}
	body, _ := json.Marshal(map[string]any{"request_id": event["request_id"], "chat_jid": chat, "outcome": map[string]any{"ok": true, "opened": true, "target": target, "path": path}})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "http://localhost/agent/respond", strings.NewReader(string(body)))
	r.RemoteAddr = "127.0.0.1:1234"
	s.Handler().ServeHTTP(w, r)
	return w
}
func TestAgentFileOpenToolOwnerSessionResponseAndReplay(t *testing.T) {
	s := fileOpenTestServer(t)
	ch := fileOpenListenerFor(t, s, "A")
	other := fileOpenListenerFor(t, s, "B")
	result := make(chan string, 1)
	errs := make(chan error, 1)
	go func() {
		out, err := s.turns.ExecuteToolByName(t.Context(), "open_workspace_file", "A", map[string]any{"path": filepath.Join(s.cfg.WorkspaceRoot, "file.md"), "target": "popout", "label": "File"})
		result <- out
		errs <- err
	}()
	var event map[string]any
	select {
	case event = <-ch:
	case <-time.After(time.Second):
		t.Fatal("no request")
	}
	if event["method"] != "custom" || event["type"] != "extension_ui_request" || event["options"].(map[string]any)["action"] != "open_workspace_file" {
		t.Fatal(event)
	}
	select {
	case <-other:
		t.Fatal("request leaked to other session")
	default:
	}
	if w := fileOpenReply(s, event, "gi:B", "", ""); w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := fileOpenReply(s, event, "", "tab", ""); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := fileOpenReply(s, event, "", "", "outside.md"); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := fileOpenReply(s, event, "", "", ""); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	select {
	case out := <-result:
		if !strings.Contains(out, `"opened":true`) {
			t.Fatal(out)
		}
	case <-time.After(time.Second):
		t.Fatal("tool blocked")
	}
	if err := <-errs; err != nil {
		t.Fatal(err)
	}
	if w := fileOpenReply(s, event, "", "", ""); w.Code != 404 {
		t.Fatal("replay accepted", w.Code)
	}
}
func TestAgentFileOpenRejectsUnsafeTargetsBeforeRequest(t *testing.T) {
	s := fileOpenTestServer(t)
	ch := fileOpenListenerFor(t, s, "A")
	outside := filepath.Join(t.TempDir(), "outside.md")
	os.WriteFile(outside, []byte("secret"), 0600)
	os.Symlink(outside, filepath.Join(s.cfg.WorkspaceRoot, "escape.md"))
	for _, path := range []string{"", "missing.md", ".", outside, "../outside.md", "escape.md"} {
		if _, err := s.requestFileOpen(t.Context(), "A", path, "tab", ""); err == nil {
			t.Fatal(path)
		}
	}
	if _, err := s.requestFileOpen(t.Context(), "A", "file.md", "outside", ""); err == nil {
		t.Fatal("bad target")
	}
	select {
	case event := <-ch:
		t.Fatal("unsafe request emitted", event)
	default:
	}
	if _, err := s.requestFileOpen(t.Context(), "B", "file.md", "tab", ""); err == nil {
		t.Fatal("missing browser accepted")
	}
}
func TestAgentFileOpenCancellationTimeoutShutdownAndLateReply(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout", "shutdown"} {
		t.Run(mode, func(t *testing.T) {
			s := fileOpenTestServer(t)
			s.fileOpens.timeout = 30 * time.Millisecond
			ch := fileOpenListenerFor(t, s, "A")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				out, err := s.requestFileOpen(ctx, "A", "file.md", "tab", "")
				if mode != "cancel" && err == nil && out["reason"] != mode {
					err = errors.New("wrong outcome")
				}
				done <- err
			}()
			var event map[string]any
			select {
			case event = <-ch:
			case err := <-done:
				t.Fatal("request rejected", err)
			case <-time.After(time.Second):
				t.Fatal("no request")
			}
			if mode == "cancel" {
				cancel()
			}
			if mode == "shutdown" {
				s.CloseFileOpenRequests()
			}
			select {
			case err := <-done:
				if mode == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if mode != "cancel" && err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("wait leaked")
			}
			if w := fileOpenReply(s, event, "", "", ""); w.Code != 404 {
				t.Fatal("late reply accepted", w.Code)
			}
			s.fileOpens.mu.Lock()
			count := len(s.fileOpens.pending)
			s.fileOpens.mu.Unlock()
			if count != 0 {
				t.Fatal("pending request leaked")
			}
		})
	}
}
func TestAgentFileOpenRespondAuthOriginAndStrictBody(t *testing.T) {
	s := fileOpenTestServer(t)
	for _, tc := range []struct {
		method, url, body, origin, auth string
		status                          int
	}{
		{"GET", "http://localhost/agent/respond", "", "", "", 405},
		{"POST", "http://localhost/agent/respond", `{}`, "https://evil.example", "", 403},
		{"POST", "http://example.com/agent/respond", `{}`, "", "", 403},
		{"POST", "http://localhost/agent/respond", `{}`, "", "Bearer fixture", 403},
		{"POST", "http://localhost/agent/respond", `{} {}`, "", "", 400},
		{"POST", "http://localhost/agent/respond", `{"unknown":true}`, "", "", 400},
		{"POST", "http://localhost/agent/respond", `{"request_id":"unknown","chat_jid":"gi:A","outcome":{}}`, "", "", 404},
	} {
		r := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
		r.RemoteAddr = "127.0.0.1:1234"
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("Authorization", tc.auth)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
}

func TestAgentFileOpenBrowserOwnerRevocationAndAmbiguity(t *testing.T) {
	s := fileOpenTestServer(t)
	// Enrol after creating the engine fixture so request delivery uses cookie owners.
	loginServer := authSessionServer(t)
	s.auth = loginServer.auth
	c1 := loginBrowser(t, loginServer, false)
	c2 := loginBrowser(t, loginServer, false)
	listen := func(cookie *http.Cookie) (<-chan map[string]any, func()) {
		r := httptest.NewRequest("GET", "http://localhost/sse/stream?chat_jid=gi:A", nil)
		r.RemoteAddr = "127.0.0.1:1234"
		r.AddCookie(cookie)
		return s.subscribeFileOpen(r, "A")
	}
	ch, stop := listen(c1)
	defer stop()
	_, stop2 := listen(c2)
	if _, err := s.requestFileOpen(t.Context(), "A", "file.md", "tab", ""); err == nil {
		t.Fatal("ambiguous browser owners accepted")
	}
	stop2()
	done := make(chan error, 1)
	go func() { _, err := s.requestFileOpen(t.Context(), "A", "file.md", "tab", ""); done <- err }()
	event := <-ch
	options := event["options"].(map[string]any)
	body, _ := json.Marshal(map[string]any{"request_id": event["request_id"], "chat_jid": "gi:A", "outcome": map[string]any{"ok": true, "opened": true, "path": options["path"], "target": options["target"]}})
	reply := func(cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "http://localhost/agent/respond", strings.NewReader(string(body)))
		r.RemoteAddr = "127.0.0.1:1234"
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	if w := reply(c2); w.Code != 404 {
		t.Fatal("foreign cookie consumed request", w.Code)
	}
	if w := reply(c1); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	s.auth.RevokeToken(c1.Value)
	if _, err := s.requestFileOpen(t.Context(), "A", "file.md", "tab", ""); err == nil {
		t.Fatal("revoked listener received request")
	}
	if w := reply(c1); w.Code != 401 {
		t.Fatal("revoked cookie accepted", w.Code)
	}
}

func TestAgentFileOpenActualSSEDelivery(t *testing.T) {
	s := fileOpenTestServer(t)
	server := httptest.NewServer(s.Handler())
	defer server.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/sse/stream?chat_jid=gi:A", nil)
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	scan := bufio.NewScanner(response.Body)
	for scan.Scan() {
		if scan.Text() == "" {
			break
		}
	}
	done := make(chan error, 1)
	go func() {
		_, err := s.turns.ExecuteToolByName(t.Context(), "open_workspace_file", "A", map[string]any{"path": "file.md", "target": "tab"})
		done <- err
	}()
	var event map[string]any
	for scan.Scan() {
		line := scan.Text()
		if strings.HasPrefix(line, "data: ") {
			json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event)
			if event["type"] == "extension_ui_request" {
				break
			}
		}
	}
	if event["kind"] != "custom" || event["chat_jid"] != "gi:A" {
		t.Fatal("wire shape", event)
	}
	body, _ := json.Marshal(map[string]any{"request_id": event["request_id"], "chat_jid": "gi:A", "outcome": map[string]any{"ok": true, "opened": true, "path": "file.md", "target": "tab"}})
	reply, err := server.Client().Post(server.URL+"/agent/respond", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	reply.Body.Close()
	if reply.StatusCode != 200 {
		t.Fatal(reply.StatusCode)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("SSE request still waiting")
	}
	cancel()
	response.Body.Close()
}
