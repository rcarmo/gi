package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/rcarmo/gi/internal/config"
)

func TestWebTerminalHandoffRejectsWrongOwnerReplayAndExpiry(t *testing.T) {
	srv, server := terminalFixture(t)
	c := terminalDial(t, server, "handoff-client-123", "")
	session := terminalRead(t, c)
	token := terminalID()
	srv.terminals.mu.Lock()
	srv.terminals.handoffs[token] = terminalHandoff{"anonymous:handoff-client-123", session["session_id"].(string), time.Now().Add(time.Minute)}
	srv.terminals.mu.Unlock()
	fake := &terminalClient{}
	if _, err := srv.attachTerminal("anonymous:wrong-client-123", token, fake); err == nil {
		t.Fatal("handoff crossed owners")
	}
	transferred := terminalDial(t, server, "handoff-client-123", token)
	terminalRead(t, transferred)
	if _, err := srv.attachTerminal("anonymous:handoff-client-123", token, fake); err == nil {
		t.Fatal("handoff replay admitted")
	}
	token = terminalID()
	srv.terminals.mu.Lock()
	srv.terminals.handoffs[token] = terminalHandoff{"anonymous:handoff-client-123", session["session_id"].(string), time.Now().Add(-time.Second)}
	srv.terminals.mu.Unlock()
	if _, err := srv.attachTerminal("anonymous:handoff-client-123", token, fake); err == nil {
		t.Fatal("expired handoff admitted")
	}
}

func TestWebTerminalBrowserCookieRevocationFencesInput(t *testing.T) {
	if !terminalSupported() {
		t.Skip("PTY unavailable")
	}
	srv := authSessionServer(t)
	srv.cfg.ShellPath = "/bin/bash"
	cookie := loginBrowser(t, srv, false)
	server := httptest.NewServer(srv.Handler())
	defer server.Close()
	defer srv.CloseTerminals()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	headers := http.Header{}
	headers.Set("Cookie", cookie.String())
	headers.Set("Origin", server.URL)
	conn, _, err := websocket.Dial(ctx, strings.Replace(server.URL, "http:", "ws:", 1)+"/terminal/ws", &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	terminalRead(t, conn)
	if err := srv.auth.RevokeToken(cookie.Value); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"type": "input", "data": "touch forbidden-after-logout\r"})
	_ = conn.Write(ctx, websocket.MessageText, payload)
	// Initial startup output can precede the auth-close frame.
	for {
		_, _, err = conn.Read(ctx)
		if err != nil {
			break
		}
	}
	if ctx.Err() != nil {
		t.Fatal("revoked connection did not close")
	}
	if _, err = os.Stat(filepath.Join(srv.workspaceRootPath(), "forbidden-after-logout")); !os.IsNotExist(err) {
		t.Fatal("revoked shell input executed", err)
	}
}

func TestWebTerminalInfoDoesNotSpawnAndShutdownRefusesNewSessions(t *testing.T) {
	srv, server := terminalFixture(t)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/terminal/session", nil)
	req.Header.Set("x-piclaw-terminal-client", "info-client-123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var p map[string]any
	json.NewDecoder(resp.Body).Decode(&p)
	if resp.StatusCode != 200 || p["active"] != false || p["transport"] != "websocket" {
		t.Fatal(resp.StatusCode, p)
	}
	srv.terminals.mu.Lock()
	n := len(srv.terminals.sessions)
	srv.terminals.mu.Unlock()
	if n != 0 {
		t.Fatal("metadata spawned a shell")
	}
	srv.CloseTerminals()
	if _, err := srv.attachTerminal("anonymous:info-client-123", "", nil); err == nil {
		t.Fatal("shutdown admitted shell")
	}
	// An unsupported configured executable reports failure without session state.
	srv = New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), ShellPath: "/missing-shell"})
	w := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "http://localhost/terminal/session?client=info-client-123", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}
