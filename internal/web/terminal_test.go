package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/rcarmo/gi/internal/config"
)

func terminalFixture(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	if !terminalSupported() {
		t.Skip("native PTY is Linux/macOS only")
	}
	root := t.TempDir()
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: root, ShellPath: "/bin/bash"})
	srv.terminals.grace = 150 * time.Millisecond
	server := httptest.NewServer(srv.Handler())
	t.Cleanup(func() { srv.CloseTerminals(); server.Close() })
	return srv, server
}
func terminalDial(t *testing.T, server *httptest.Server, client, token string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	target := strings.Replace(server.URL, "http:", "ws:", 1) + "/terminal/ws?client=" + url.QueryEscape(client)
	if token != "" {
		target += "&handoff=" + url.QueryEscape(token)
	}
	conn, _, err := websocket.Dial(ctx, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}
func terminalRead(t *testing.T, c *websocket.Conn) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var p map[string]any
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err, string(data))
	}
	return p
}
func terminalInput(t *testing.T, c *websocket.Conn, text string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	p, _ := json.Marshal(map[string]any{"type": "input", "data": text})
	if err := c.Write(ctx, websocket.MessageText, p); err != nil {
		t.Fatal(err)
	}
}
func terminalUntil(t *testing.T, c *websocket.Conn, text string) string {
	t.Helper()
	var out strings.Builder
	for out.Len() < 100000 {
		p := terminalRead(t, c)
		if p["type"] == "output" {
			out.WriteString(p["data"].(string))
			if strings.Contains(out.String(), text) {
				return out.String()
			}
		}
	}
	t.Fatal("output limit")
	return ""
}

func TestWebTerminalPTYInputResizeIsolationReconnectAndHandoff(t *testing.T) {
	srv, server := terminalFixture(t)
	c := terminalDial(t, server, "client-one-123", "")
	session := terminalRead(t, c)
	if session["type"] != "session" || session["cols"] != float64(120) || session["cwd"] != srv.workspaceRootPath() {
		t.Fatal(session)
	}
	terminalInput(t, c, "stty -echo; printf '\\nNATIVE_PTY_%s\\n' READY; pwd\r")
	out := terminalUntil(t, c, "NATIVE_PTY_READY")
	if !strings.Contains(out, srv.workspaceRootPath()) {
		terminalUntil(t, c, srv.workspaceRootPath())
	}
	p, _ := json.Marshal(map[string]any{"type": "resize", "cols": 88, "rows": 25})
	if err := c.Write(context.Background(), websocket.MessageText, p); err != nil {
		t.Fatal(err)
	}
	terminalInput(t, c, "stty size; printf 'RESIZE_DONE\\n'\r")
	if out := terminalUntil(t, c, "RESIZE_DONE"); !strings.Contains(out, "25 88") {
		t.Fatal(out)
	}
	other := terminalDial(t, server, "client-two-123", "")
	if terminalRead(t, other)["session_id"] == session["session_id"] {
		t.Fatal("owners shared session")
	}
	c.CloseNow()
	// Wait for handler detach, not an arbitrary fixed reconnect delay.
	deadline := time.Now().Add(time.Second)
	for {
		srv.terminals.mu.Lock()
		ts := srv.terminals.sessions["anonymous:client-one-123"]
		detached := false
		if ts != nil {
			ts.mu.Lock()
			detached = len(ts.clients) == 0
			ts.mu.Unlock()
		}
		srv.terminals.mu.Unlock()
		if detached {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("detach not observed")
		}
		time.Sleep(time.Millisecond)
	}
	c = terminalDial(t, server, "client-one-123", "")
	if terminalRead(t, c)["session_id"] != session["session_id"] {
		t.Fatal("reconnect lost PTY")
	}
	if p := terminalRead(t, c); !strings.Contains(p["data"].(string), "NATIVE_PTY_READY") {
		t.Fatal("no replay", p)
	}
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/terminal/handoff", nil)
	req.Header.Set("x-piclaw-terminal-client", "client-one-123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Handoff struct {
			Token string `json:"token"`
		} `json:"handoff"`
	}
	json.NewDecoder(resp.Body).Decode(&body)
	if resp.StatusCode != 200 || body.Handoff.Token == "" {
		t.Fatal(resp.StatusCode, body)
	}
	transferred := terminalDial(t, server, "client-one-123", body.Handoff.Token)
	if terminalRead(t, transferred)["session_id"] != session["session_id"] {
		t.Fatal("handoff reset PTY")
	}
	transferred.CloseNow()
	other.CloseNow()
	srv.CloseTerminals()
	srv.terminals.mu.Lock()
	remaining := len(srv.terminals.sessions)
	srv.terminals.mu.Unlock()
	if remaining != 0 {
		t.Fatal("shutdown leaked sessions")
	}
}

func TestWebTerminalRejectsCrossOriginMissingClientAndUnauthenticatedRequests(t *testing.T) {
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir()})
	for _, tc := range []struct {
		target, origin string
		status         int
	}{
		{"/terminal/session", "", 400}, {"/terminal/session?client=valid-client-123", "https://evil.example", 403}, {"/terminal/ws?client=valid-client-123", "null", 403},
	} {
		r := httptest.NewRequest(http.MethodGet, "http://localhost"+tc.target, nil)
		r.RemoteAddr = "127.0.0.1:12345"
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
	srv = authSessionServer(t)
	for _, path := range []string{"/terminal/session", "/terminal/ws", "/terminal/handoff"} {
		method := http.MethodGet
		if strings.HasSuffix(path, "handoff") {
			method = http.MethodPost
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "http://localhost"+path+"?client=valid-client-123", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		srv.Handler().ServeHTTP(w, req)
		if w.Code != 401 {
			t.Fatal(path, w.Code)
		}
	}
}

func TestWebTerminalReplayBoundedAndDetachKillsProcess(t *testing.T) {
	srv, server := terminalFixture(t)
	c := terminalDial(t, server, "client-expiry-123", "")
	p := terminalRead(t, c)
	srv.terminals.mu.Lock()
	ts := srv.terminals.sessions["anonymous:client-expiry-123"]
	srv.terminals.mu.Unlock()
	c.CloseNow()
	select {
	case <-ts.done:
	case <-time.After(3 * time.Second):
		t.Fatal("detached process leaked", p)
	}
	ts = &terminalSession{clients: map[*terminalClient]struct{}{}}
	ts.output(strings.Repeat("α", terminalReplayLimit))
	ts.mu.Lock()
	n := len(ts.history)
	ts.mu.Unlock()
	if n > terminalReplayLimit {
		t.Fatal(n)
	}
	ts.output("終わり")
	ts.mu.Lock()
	replay := ts.replay()
	ts.mu.Unlock()
	if !utf8.ValidString(replay) || !strings.HasSuffix(replay, "終わり") || len(replay) > terminalReplayLimit {
		t.Fatal("corrupt ring replay")
	}
	// Workspace cwd exists and replay cap does not create disk artifacts.
	if _, err := os.Stat(filepath.Join(srv.workspaceRootPath(), ".terminal")); !os.IsNotExist(err) {
		t.Fatal("unexpected terminal disk state")
	}
}
