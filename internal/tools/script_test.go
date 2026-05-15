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
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/scripting"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
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

func TestScriptToolJSCanPublishTopics(t *testing.T) {
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
	var published map[string]any
	tool.SetConnectivityCallbacks(nil, nil, nil, nil, func(ctx context.Context, sessionID string, envelope map[string]any) error {
		published = envelope
		if sessionID != session.ID {
			t.Fatalf("unexpected session id for published topic: %q", sessionID)
		}
		return nil
	}, nil)
	out := tool.Execute(context.Background(), ScriptInput{SessionID: session.ID, Script: `gi.topics.publish({topic: "runtime.test", payload: {ok: true}, type: "notice"}); "ok";`})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "ok" {
		t.Fatalf("unexpected result: %q", out.Result)
	}
	if published["topic"] != "runtime.test" {
		t.Fatalf("unexpected published topic envelope: %#v", published)
	}
}

func TestScriptToolReadTopicSubscriptionRemovesClosedHandle(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	session, err := s.CreateSession(context.Background(), "session_topic_closed", "ClosedTopic", map[string]any{"model": "bootstrap"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), DefaultModel: "test-model", DefaultProvider: "test", DefaultThinkingLevel: "low"})
	tool.SetConnectivityCallbacks(nil, nil, nil, nil, nil, func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error) {
		ch := make(chan topics.Envelope, 1)
		close(ch)
		return ch, func() {}, nil
	})
	id, err := tool.subscribeTopic(context.Background(), session.ID, "runtime.test", scripting.TopicSubscribeOptions{})
	if err != nil {
		t.Fatalf("subscribe topic: %v", err)
	}
	messages, err := tool.readTopicSubscription(context.Background(), session.ID, id, 10)
	if err != nil {
		t.Fatalf("read topic subscription: %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("expected no messages from closed subscription, got %#v", messages)
	}
	if _, err := tool.readTopicSubscription(context.Background(), session.ID, id, 10); err == nil {
		t.Fatal("expected closed subscription handle to be removed after read")
	}
}

func TestScriptToolJSCanSubscribeReadAndUnsubscribeTopics(t *testing.T) {
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
	tool.SetConnectivityCallbacks(nil, nil, nil, nil, nil, func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error) {
		ch := make(chan topics.Envelope, 2)
		ch <- topics.Envelope{Topic: "runtime.test", SessionID: sessionID, Source: "script", Type: "notice", Payload: map[string]any{"ok": true}}
		close(ch)
		return ch, func() {}, nil
	})
	out := tool.Execute(context.Background(), ScriptInput{SessionID: session.ID, Script: `var sub = gi.topics.subscribe("runtime.*"); var ev = gi.topics.read(sub, 5); gi.topics.unsubscribe(sub); ev[0].topic + ":" + ev[0].payload.ok;`})
	if out.Error != "" {
		t.Fatalf("script error: %v", out.Error)
	}
	if out.Result != "runtime.test:true" {
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

func TestScriptToolClearEventHooksOnlyRemovesCurrentSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	if err := tool.registerEventHook(context.Background(), "session_a", scripting.EventHookSpec{Name: "evt", Source: "script"}); err != nil {
		t.Fatalf("register event hook session_a: %v", err)
	}
	if err := tool.registerEventHook(context.Background(), "session_b", scripting.EventHookSpec{Name: "evt", Source: "script"}); err != nil {
		t.Fatalf("register event hook session_b: %v", err)
	}
	if len(tool.eventHooks) != 2 {
		t.Fatalf("expected two session-scoped event hooks, got %d", len(tool.eventHooks))
	}
	if err := tool.clearEventHooks(context.Background(), "session_a"); err != nil {
		t.Fatalf("clear event hooks: %v", err)
	}
	if len(tool.eventHooks) != 1 || tool.eventHooks[0].sessionID != "session_b" {
		t.Fatalf("expected only session_b hook to remain, got %#v", tool.eventHooks)
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

func TestScriptToolClearEventHooksAlsoClearsHostRegisteredHooks(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	clearedA := 0
	clearedB := 0
	tool.SetAgenticCallbacks(
		func(ctx context.Context, sessionID string, hook scripting.EventHookSpec) (func(), error) {
			switch sessionID {
			case "session_a":
				return func() { clearedA++ }, nil
			case "session_b":
				return func() { clearedB++ }, nil
			default:
				return nil, nil
			}
		},
		nil, nil, nil,
	)
	bridgeA := tool.buildBridge("session_a")
	if err := bridgeA.Funcs.RegisterEventHook(context.Background(), scripting.EventHookSpec{Name: "evt"}); err != nil {
		t.Fatalf("register host hook session_a: %v", err)
	}
	bridgeB := tool.buildBridge("session_b")
	if err := bridgeB.Funcs.RegisterEventHook(context.Background(), scripting.EventHookSpec{Name: "evt"}); err != nil {
		t.Fatalf("register host hook session_b: %v", err)
	}
	if err := bridgeA.Funcs.ClearEventHooks(context.Background()); err != nil {
		t.Fatalf("clear event hooks session_a: %v", err)
	}
	if clearedA != 1 || clearedB != 0 {
		t.Fatalf("expected only session_a host hook to be cleared, got clearedA=%d clearedB=%d", clearedA, clearedB)
	}
	if err := bridgeB.Funcs.ClearEventHooks(context.Background()); err != nil {
		t.Fatalf("clear event hooks session_b: %v", err)
	}
	if clearedA != 1 || clearedB != 1 {
		t.Fatalf("expected both host hooks cleared by their own sessions, got clearedA=%d clearedB=%d", clearedA, clearedB)
	}
}

func TestScriptToolEmitEventOnlyMatchesCurrentSessionHooks(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	if err := tool.registerEventHook(context.Background(), "session_a", scripting.EventHookSpec{Name: "evt", Source: "script"}); err != nil {
		t.Fatalf("register event hook session_a: %v", err)
	}
	if err := tool.registerEventHook(context.Background(), "session_b", scripting.EventHookSpec{Name: "evt", Source: "script"}); err != nil {
		t.Fatalf("register event hook session_b: %v", err)
	}
	if err := tool.emitEvent(context.Background(), "session_a", "evt", map[string]any{"source": "script"}); err != nil {
		t.Fatalf("emit event: %v", err)
	}
	matched := 0
	for _, hook := range tool.eventHooks {
		if hook.sessionID == "session_a" {
			matched++
		}
	}
	if matched != 1 {
		t.Fatalf("expected session_a hook to remain intact, got %#v", tool.eventHooks)
	}
}

func TestScriptToolTopicSubscriptionRejectsCrossSessionReadAndUnsubscribe(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	ch := make(chan topics.Envelope, 1)
	tool.topicSubs["topic_1"] = topicSubscription{sessionID: "session_a", ch: ch, unsubscribe: func() {}}

	if _, err := tool.readTopicSubscription(context.Background(), "session_b", "topic_1", 1); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session topic read rejection, got %v", err)
	}
	if err := tool.unsubscribeTopic(context.Background(), "session_b", "topic_1"); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session topic unsubscribe rejection, got %v", err)
	}
}

func TestScriptToolCloseRawSocketIsIdempotent(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	socketID, err := tool.openRawSocket("session_a", context.Background(), scripting.RawSocketSpec{Protocol: "tcp", Address: ln.Addr().String()})
	if err != nil {
		t.Fatalf("open raw socket: %v", err)
	}
	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for raw socket accept")
	}
	if err := tool.closeRawSocket("session_a", context.Background(), socketID); err != nil {
		t.Fatalf("first close raw socket: %v", err)
	}
	if err := tool.closeRawSocket("session_a", context.Background(), socketID); err != nil {
		t.Fatalf("second close raw socket should be idempotent: %v", err)
	}
}

func TestScriptToolRawSocketRejectsCrossSessionAccess(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()
	tool.rawSockets["raw_1"] = ownedRawSocket{sessionID: "session_a", conn: clientConn}

	if _, err := tool.writeRawSocket("session_b", context.Background(), scripting.RawSocketPayload{SocketID: "raw_1", Data: "ping"}); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session raw socket write rejection, got %v", err)
	}
	if _, err := tool.readRawSocket("session_b", context.Background(), scripting.RawSocketPayload{SocketID: "raw_1", MaxBytes: 4}); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session raw socket read rejection, got %v", err)
	}
	if err := tool.closeRawSocket("session_b", context.Background(), "raw_1"); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session raw socket close rejection, got %v", err)
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

func TestScriptToolWebSocketRejectsCrossSessionAccess(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	tool.webSockets["ws_1"] = ownedWebSocket{sessionID: "session_a", conn: &xwebsocket.Conn{}}

	if err := tool.writeWebSocket("session_b", context.Background(), "ws_1", "hi"); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session websocket write rejection, got %v", err)
	}
	if _, err := tool.readWebSocket("session_b", context.Background(), "ws_1", 1); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session websocket read rejection, got %v", err)
	}
	if err := tool.closeWebSocket("session_b", context.Background(), "ws_1"); err == nil || !strings.Contains(err.Error(), "does not belong to session") {
		t.Fatalf("expected cross-session websocket close rejection, got %v", err)
	}
}

func TestScriptToolCloseWebSocketIsIdempotent(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xwebsocket.Handler(func(c *xwebsocket.Conn) {
			select {}
		}).ServeHTTP(w, r)
	}))
	defer server.Close()
	wsURL := "ws://" + strings.TrimPrefix(server.URL, "http://") + "/"

	tool := NewScriptTool(s, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	socketID, err := tool.openWebSocket("session_a", context.Background(), scripting.WebSocketSpec{URL: wsURL, TimeoutMS: 5000})
	if err != nil {
		t.Fatalf("open websocket: %v", err)
	}
	if err := tool.closeWebSocket("session_a", context.Background(), socketID); err != nil {
		t.Fatalf("first close websocket: %v", err)
	}
	if err := tool.closeWebSocket("session_a", context.Background(), socketID); err != nil {
		t.Fatalf("second close websocket should be idempotent: %v", err)
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
