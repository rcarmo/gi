package tools

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	xwebsocket "golang.org/x/net/websocket"
)

func TestScriptToolJSCanIntrospectAndMutateSessionState(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model", DefaultProvider: "test", DefaultThinkingLevel: "low"})

	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Script:    `const info = gi.getSessionInfo(); const turns = gi.listTurns(); const cfg = gi.getRuntimeConfig(); gi.setSessionState({markdown_ready: true}); "# " + (info.session.title || info.session.id) + " / " + cfg.default_model + " / turns=" + turns.length;`,
	})
	if out.Error != "" {
		t.Fatalf("script returned error: %s", out.Error)
	}
	if !strings.Contains(out.Result, "# demo") || !strings.Contains(out.Result, "test-model") || !strings.Contains(out.Result, "turns=0") {
		t.Fatalf("unexpected result: %q", out.Result)
	}
	updated, err := s.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if updated.State["markdown_ready"] != true {
		t.Fatalf("expected markdown_ready=true, got %#v", updated.State["markdown_ready"])
	}
}

func TestScriptToolJokerCanIntrospectAndMutateSessionState(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model", DefaultProvider: "test", DefaultThinkingLevel: "low"})

	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Engine:    "joker",
		Script:    `(do (gi-set-session-state! {:markdown_ready true}) (str "# " (or (:title (:session (gi-get-session-info))) (:id (:session (gi-get-session-info)))) " / " (:default_model (gi-get-runtime-config)) " / turns=" (count (gi-list-turns))))`,
	})
	if out.Error != "" {
		t.Fatalf("script returned error: %s", out.Error)
	}
	if !strings.Contains(out.Result, "# demo") || !strings.Contains(out.Result, "test-model") || !strings.Contains(out.Result, "turns=0") {
		t.Fatalf("unexpected result: %q", out.Result)
	}
	updated, err := s.GetSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if updated.State["markdown_ready"] != true {
		t.Fatalf("expected markdown_ready=true, got %#v", updated.State["markdown_ready"])
	}
}

func TestScriptToolJSCanListMessages(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(context.Background(), "m1", session.ID, "user", "first", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if err := s.AddMessage(context.Background(), "m2", session.ID, "assistant", "second", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if err := s.AddMessage(context.Background(), "m3", session.ID, "assistant", "third", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model", DefaultProvider: "test", DefaultThinkingLevel: "low"})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Script:    `gi.listMessages().length + ":" + gi.listMessages(1).length;`,
	})
	if out.Error != "" {
		t.Fatalf("script error: %s", out.Error)
	}
	if out.Result != "3:1" {
		t.Fatalf("unexpected message count: %q", out.Result)
	}
}

func TestScriptToolJokerCanListMessages(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := s.AddMessage(context.Background(), "m1", session.ID, "user", "first", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if err := s.AddMessage(context.Background(), "m2", session.ID, "assistant", "second", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}
	if err := s.AddMessage(context.Background(), "m3", session.ID, "assistant", "third", nil); err != nil {
		t.Fatalf("add message: %v", err)
	}

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model", DefaultProvider: "test", DefaultThinkingLevel: "low"})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Engine:    "joker",
		Script:    `(str (count (gi-list-messages)) ":" (count (gi-list-messages {:limit 1})))`,
	})
	if out.Error != "" {
		t.Fatalf("script error: %s", out.Error)
	}
	if out.Result != "3:1" {
		t.Fatalf("unexpected message count: %q", out.Result)
	}
}

func TestScriptToolCanLoadJSFromVFSPath(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.SaveVFSFile(context.Background(), "scripts", "sample.js", "application/javascript", []byte("'ok from vfs'"), map[string]any{}); err != nil {
		t.Fatalf("seed vfs script: %v", err)
	}

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model", DefaultProvider: "test", DefaultThinkingLevel: "low"})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Path:      "vfs://scripts/sample.js",
	})
	if out.Error != "" {
		t.Fatalf("script error: %s", out.Error)
	}
	if out.Result != "ok from vfs" {
		t.Fatalf("unexpected result: %q", out.Result)
	}
}

func TestScriptToolJokerSupportsEventHooksAndEmit(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Engine:    "joker",
		Script:    `(do (gi-register-event-hook {:name "evt" :source "script" :filter {:kind "test"}}) (gi-emit-event "evt" {:source "script" :payload {:ok true}}) (gi-clear-event-hooks) "ok")`,
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "ok" {
		t.Fatalf("unexpected result: %q", out.Result)
	}
	if len(tool.eventHooks) != 0 {
		t.Fatalf("expected event hooks cleared, got %d", len(tool.eventHooks))
	}
}

func TestScriptToolJokerRawSocketsRoundTrip(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		_, _ = conn.Write(buf[:n])
	}()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	addr := ln.Addr().String()
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Engine:    "joker",
		Script: fmt.Sprintf(`(let [sid (gi-open-raw-socket {:protocol "tcp" :address "%s"})]
			(gi-write-raw-socket {:socket_id sid :data "ping" :max_bytes 32})
			(let [msg (gi-read-raw-socket {:socket_id sid :max_bytes 32})]
				(gi-close-raw-socket sid)
				msg))`, addr),
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "ping" {
		t.Fatalf("expected ping, got %q", out.Result)
	}
}

func TestScriptToolJokerWebSocketRoundTrip(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xwebsocket.Handler(func(c *xwebsocket.Conn) {
			_, _ = io.Copy(c, c)
		}).ServeHTTP(w, r)
	}))
	defer server.Close()
	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://") + "/"

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Engine:    "joker",
		Script: fmt.Sprintf(`(let [sid (gi-open-websocket {:url "%s" :timeout_ms 5000})]
			(gi-write-websocket sid "hi")
			(let [msg (gi-read-websocket sid 5000)]
				(gi-close-websocket sid)
				msg))`, wsURL),
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "hi" {
		t.Fatalf("expected hi, got %q", out.Result)
	}
}

func TestScriptToolJokerHTTPRequest(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("x-test"); got != "abc" {
			t.Errorf("unexpected header: %q", got)
		}
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		w.Header().Set("content-type", "text/plain")
		_, _ = w.Write(append([]byte("echo:"), body...))
	}))
	defer h.Close()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Engine:    "joker",
		Script: fmt.Sprintf(`(let [r (gi-http-request {:method "POST" :url "%s" :headers {"x-test" ["abc"]} :body "hello" :timeout_ms 5000})]
			(str (:status_code r) ":" (:body r)))`, h.URL),
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "200:echo:hello" {
		t.Fatalf("unexpected response payload: %q", out.Result)
	}
}

func TestScriptToolSupportsEventHooksAndEmit(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Script:    `gi.registerEventHook({name: "evt", source: "script", filter: {kind: "test"}}); gi.emitEvent("evt", {source: "script", payload: {ok: true}}); gi.clearEventHooks(); "ok";`,
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "ok" {
		t.Fatalf("unexpected result: %q", out.Result)
	}
	if len(tool.eventHooks) != 0 {
		t.Fatalf("expected event hooks cleared, got %d", len(tool.eventHooks))
	}
}

func TestScriptToolRawSocketsRoundTrip(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		_, _ = conn.Write(buf[:n])
	}()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	addr := ln.Addr().String()
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Script:    fmt.Sprintf(`var sid = gi.net.openRawSocket({protocol: "tcp", address: "%s"}); gi.net.writeRawSocket({socket_id: sid, data: "ping"}); var out = gi.net.readRawSocket({socket_id: sid, max_bytes: 32}); gi.net.closeRawSocket(sid); out;`, addr),
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "ping" {
		t.Fatalf("expected ping, got %q", out.Result)
	}
}

func TestScriptToolWebSocketRoundTrip(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xwebsocket.Handler(func(c *xwebsocket.Conn) {
			_, _ = io.Copy(c, c)
		}).ServeHTTP(w, r)
	}))
	defer server.Close()
	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://") + "/"

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Script:    fmt.Sprintf(`var sid = gi.websocket.open({url: "%s", timeout_ms: 5000}); gi.websocket.write(sid, "hi"); var out = gi.websocket.read(sid, 5000); gi.websocket.close(sid); out;`, wsURL),
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "hi" {
		t.Fatalf("expected hi, got %q", out.Result)
	}
}

func TestScriptToolHTTPRequest(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	session, err := s.CreateSession(context.Background(), store.NowID("session"), "demo", map[string]any{"model": "test-model", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("x-test"); got != "abc" {
			t.Errorf("unexpected header: %q", got)
		}
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		w.Header().Set("content-type", "text/plain")
		_, _ = w.Write(append([]byte("echo:"), body...))
	}))
	defer h.Close()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	out := tool.Execute(context.Background(), ScriptInput{
		SessionID: session.ID,
		Script:    fmt.Sprintf(`var r = gi.http.request({method: "POST", url: "%s", headers: {"x-test": ["abc"]}, body: "hello", timeout_ms: 5000}); JSON.stringify({code: r.status_code, body: r.body});`, h.URL),
	})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if !strings.Contains(out.Result, `"code":200`) || !strings.Contains(out.Result, `"body":"echo:hello"`) {
		t.Fatalf("unexpected response payload: %q", out.Result)
	}
}
