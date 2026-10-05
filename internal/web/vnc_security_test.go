package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/rcarmo/gi/internal/config"
)

func TestVNCBrowserRevocationFencesFramesAndClosesUpstream(t *testing.T) {
	fixture, _ := vncFixture(t)
	srv := authSessionServer(t)
	var targets []config.VNCTarget
	for _, target := range fixture.vnc.targets {
		targets = append(targets, target)
	}
	srv.vnc = newVNCManager(config.RuntimeConfig{VNCTargets: targets})
	server := httptest.NewServer(srv.Handler())
	defer server.Close()
	defer srv.CloseVNC()
	cookie := loginBrowser(t, srv, false)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	headers := http.Header{}
	headers.Set("Cookie", cookie.String())
	headers.Set("Origin", server.URL)
	conn, _, err := websocket.Dial(ctx, strings.Replace(server.URL, "http:", "ws:", 1)+"/vnc/ws?target=test", &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	vncRead(t, conn)
	vncRead(t, conn)
	srv.vnc.mu.Lock()
	var ts *vncSession
	for _, session := range srv.vnc.sessions {
		ts = session
	}
	srv.vnc.mu.Unlock()
	if err := srv.auth.RevokeToken(cookie.Value); err != nil {
		t.Fatal(err)
	}
	_ = conn.Write(ctx, websocket.MessageBinary, []byte("revoked-input"))
	_, _, err = conn.Read(ctx)
	if err == nil || ctx.Err() != nil {
		t.Fatal("revoked socket did not close promptly", err)
	}
	select {
	case <-ts.done:
	case <-time.After(time.Second):
		t.Fatal("revoked upstream leaked")
	}
}

func TestVNCPingIsLocalAndTargetMetadataOmitsHostPort(t *testing.T) {
	_, server := vncFixture(t)
	response, err := http.Get(server.URL + "/vnc/session?target=test")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	target := payload["target"].(map[string]any)
	if target["host"] != nil || target["port"] != nil || target["direct_connect"] != false {
		t.Fatal(target)
	}
	for _, entry := range payload["targets"].([]any) {
		tgt := entry.(map[string]any)
		if tgt["host"] != nil || tgt["port"] != nil {
			t.Fatal(tgt)
		}
	}
}
