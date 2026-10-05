package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/rcarmo/gi/internal/config"
)

func vncFixture(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var conns []net.Conn
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			c, err := listener.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				_, _ = c.Write([]byte("RFB 003.008\n"))
				_, _ = io.Copy(c, c)
			}(c)
		}
	}()
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	n, _ := strconv.Atoi(port)
	srv := New(nil, nil, config.RuntimeConfig{WorkspaceRoot: t.TempDir(), VNCTargets: []config.VNCTarget{{ID: "test", Label: "Local test", Host: host, Port: n}}})
	server := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		srv.CloseVNC()
		server.Close()
		listener.Close()
		mu.Lock()
		for _, c := range conns {
			c.Close()
		}
		mu.Unlock()
		wg.Wait()
	})
	return srv, server
}
func vncDial(t *testing.T, server *httptest.Server, target, token string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	u := strings.Replace(server.URL, "http:", "ws:", 1) + "/vnc/ws?target=" + url.QueryEscape(target)
	if token != "" {
		u += "&handoff=" + url.QueryEscape(token)
	}
	c, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.CloseNow() })
	return c
}
func vncRead(t *testing.T, c *websocket.Conn) (websocket.MessageType, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	kind, data, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return kind, data
}

func TestVNCBridgeBinaryPingHandoffAndShutdown(t *testing.T) {
	srv, server := vncFixture(t)
	c := vncDial(t, server, "test", "")
	kind, data := vncRead(t, c)
	if kind != websocket.MessageText || !bytes.Contains(data, []byte("vnc.connected")) {
		t.Fatal(kind, string(data))
	}
	kind, data = vncRead(t, c)
	if kind != websocket.MessageBinary || string(data) != "RFB 003.008\n" {
		t.Fatal(kind, string(data))
	}
	original := []byte{0, 1, 2, 3, 255}
	if err := c.Write(context.Background(), websocket.MessageBinary, original); err != nil {
		t.Fatal(err)
	}
	_, data = vncRead(t, c)
	if !bytes.Equal(data, original) {
		t.Fatal(data)
	}
	if err := c.Write(context.Background(), websocket.MessageText, []byte(`{"type":"ping"}`)); err != nil {
		t.Fatal(err)
	}
	kind, data = vncRead(t, c)
	if kind != websocket.MessageText || string(data) != `{"type":"pong"}` {
		t.Fatal(kind, string(data))
	}
	response, err := http.Post(server.URL+"/vnc/handoff?target=test", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var p struct {
		Handoff struct {
			Token string `json:"token"`
		} `json:"handoff"`
	}
	json.NewDecoder(response.Body).Decode(&p)
	if response.StatusCode != 200 || p.Handoff.Token == "" {
		t.Fatal(response.StatusCode)
	}
	srv.vnc.mu.Lock()
	old := srv.vnc.sessions["vnc-local\x00test"]
	srv.vnc.mu.Unlock()
	transferred := vncDial(t, server, "test", p.Handoff.Token)
	_, data = vncRead(t, transferred)
	if !bytes.Contains(data, []byte("vnc.connected")) {
		t.Fatal(string(data))
	}
	srv.vnc.mu.Lock()
	same := srv.vnc.sessions[old.key] == old
	srv.vnc.mu.Unlock()
	if !same {
		t.Fatal("handoff reconnected TCP")
	}
	if err := transferred.Write(context.Background(), websocket.MessageBinary, original); err != nil {
		t.Fatal(err)
	}
	_, data = vncRead(t, transferred)
	if !bytes.Equal(data, original) {
		t.Fatal(data)
	}
	if _, err := srv.attachVNC(context.Background(), "vnc-local", "test", p.Handoff.Token, nil); err == nil {
		t.Fatal("handoff replay accepted")
	}
	srv.CloseVNC()
	select {
	case <-old.done:
	case <-time.After(time.Second):
		t.Fatal("upstream reader leaked")
	}
}

func TestVNCTargetPolicyAndAuthDoNotDialArbitraryHosts(t *testing.T) {
	srv, server := vncFixture(t)
	response, err := http.Get(server.URL + "/vnc/session")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var p map[string]any
	json.NewDecoder(response.Body).Decode(&p)
	if p["direct_connect_enabled"] != false || p["host_policy"] != "allowlist" {
		t.Fatal(p)
	}
	for _, ref := range []string{"127.0.0.1:22", "[::1]:5900", "unknown", "http://169.254.169.254"} {
		r := httptest.NewRequest(http.MethodGet, "http://localhost/vnc/ws?target="+url.QueryEscape(ref), nil)
		r.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if w.Code != 404 {
			t.Fatal(ref, w.Code)
		}
	}
	srv = authSessionServer(t)
	srv.vnc = newVNCManager(config.RuntimeConfig{VNCTargets: []config.VNCTarget{{ID: "test", Host: "localhost", Port: 5900}}})
	for _, path := range []string{"/vnc/session", "/vnc/ws?target=test", "/vnc/handoff?target=test"} {
		method := http.MethodGet
		if strings.Contains(path, "handoff") {
			method = http.MethodPost
		}
		r := httptest.NewRequest(method, "http://localhost"+path, nil)
		r.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatal(path, w.Code)
		}
	}
	r := httptest.NewRequest(http.MethodGet, "http://localhost/vnc/session", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	m := newVNCManager(config.RuntimeConfig{VNCAllowDirect: true})
	for _, ref := range []string{"127.0.0.1:5900", "[::1]:5900"} {
		if _, ok := m.target(ref); !ok {
			t.Fatal(ref)
		}
	}
	for _, ref := range []string{"x:0", "x:65536", "x:bad", "x/y:5900"} {
		if _, ok := m.target(ref); ok {
			t.Fatal("invalid direct target", ref)
		}
	}
}

func TestVNCHandoffExpiryWrongOwnerAndDetachCleanup(t *testing.T) {
	srv, server := vncFixture(t)
	srv.vnc.ttl = 80 * time.Millisecond
	c := vncDial(t, server, "test", "")
	vncRead(t, c)
	vncRead(t, c)
	response, err := http.Post(server.URL+"/vnc/handoff?target=test", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		Handoff struct {
			Token string `json:"token"`
		} `json:"handoff"`
	}
	json.NewDecoder(response.Body).Decode(&p)
	response.Body.Close()
	if _, err := srv.attachVNC(context.Background(), "another-owner", "test", p.Handoff.Token, nil); err == nil {
		t.Fatal("handoff crossed owner")
	}
	srv.vnc.mu.Lock()
	ts := srv.vnc.sessions["vnc-local\x00test"]
	srv.vnc.mu.Unlock()
	c.CloseNow()
	select {
	case <-ts.done:
	case <-time.After(2 * time.Second):
		t.Fatal("expired handoff leaked TCP")
	}
}
