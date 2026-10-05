package web

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/rcarmo/gi/internal/config"
)

// Target addresses are operator configuration; untrusted target references
// never become network destinations unless direct access was opted into.
type vncFrame struct {
	kind websocket.MessageType
	data []byte
}
type vncClient struct {
	conn *websocket.Conn
	out  chan vncFrame
	stop chan struct{}
	once sync.Once
}

func (c *vncClient) close() { c.once.Do(func() { close(c.stop); _ = c.conn.CloseNow() }) }
func (c *vncClient) send(f vncFrame) {
	select {
	case c.out <- f:
	default:
		c.close()
	}
}

type vncSession struct {
	owner, key string
	target     config.VNCTarget
	tcp        net.Conn
	client     *vncClient
	token      string
	expires    time.Time
	timer      *time.Timer
	done       chan struct{}
}
type vncManager struct {
	mu              sync.Mutex
	targets         map[string]config.VNCTarget
	direct          bool
	sessions        map[string]*vncSession
	handoffs        map[string]*vncSession
	closed          bool
	dialing         int
	workers         sync.WaitGroup
	ttl             time.Duration
	greetingTimeout time.Duration
}

func newVNCManager(cfg config.RuntimeConfig) *vncManager {
	m := &vncManager{targets: map[string]config.VNCTarget{}, direct: cfg.VNCAllowDirect, sessions: map[string]*vncSession{}, handoffs: map[string]*vncSession{}, ttl: 15 * time.Second, greetingTimeout: 10 * time.Second}
	for _, t := range cfg.VNCTargets {
		if len(m.targets) >= 128 {
			break
		}
		if validVNCTarget(t) {
			if t.Label == "" {
				t.Label = t.ID
			}
			m.targets[t.ID] = t
		}
	}
	return m
}
func validVNCTarget(t config.VNCTarget) bool {
	return t.ID != "" && len(t.ID) <= 128 && !strings.ContainsAny(t.ID, "\x00\r\n") && len(t.Host) <= 255 && t.Host != "" && !strings.ContainsAny(t.Host, " /?#\\\x00\t\r\n") && t.Port > 0 && t.Port <= 65535
}
func (m *vncManager) target(ref string) (config.VNCTarget, bool) {
	if t, ok := m.targets[ref]; ok {
		return t, true
	}
	if !m.direct {
		return config.VNCTarget{}, false
	}
	host, port, err := net.SplitHostPort(ref)
	if err != nil {
		return config.VNCTarget{}, false
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		return config.VNCTarget{}, false
	}
	t := config.VNCTarget{ID: ref, Label: ref, Host: host, Port: n}
	return t, validVNCTarget(t)
}
func (s *Server) vncOwner(w http.ResponseWriter, r *http.Request) (string, bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	if !providerWriteTransport(r) || !browserSameOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "VNC requires same-origin HTTPS or localhost"})
		return "", false
	}
	if !s.requireAuthenticatedRequest(w, r) {
		return "", false
	}
	enrolled, err := s.auth.Enrolled()
	if err != nil {
		writeJSON(w, 500, map[string]any{"error": "Authentication unavailable"})
		return "", false
	}
	if !enrolled {
		return "vnc-local", true
	}
	cookie, err := r.Cookie(browserSessionCookie)
	if err != nil || r.Header.Get("Authorization") != "" || r.URL.Query().Has("auth_token") {
		writeJSON(w, 401, map[string]any{"error": "Browser authentication required"})
		return "", false
	}
	valid, err := s.auth.ValidateTokenWithError(cookie.Value)
	if err != nil || !valid {
		writeJSON(w, 401, map[string]any{"error": "Browser authentication required"})
		return "", false
	}
	hash := sha256.Sum256([]byte(cookie.Value))
	return "browser:" + hex.EncodeToString(hash[:]), true
}
func (s *Server) vncAuthorized(r *http.Request, owner string) bool {
	if !strings.HasPrefix(owner, "browser:") {
		return true
	}
	cookie, err := r.Cookie(browserSessionCookie)
	if err != nil {
		return false
	}
	valid, err := s.auth.ValidateTokenWithError(cookie.Value)
	return err == nil && valid
}
func (s *Server) handleVNCSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	if _, ok := s.vncOwner(w, r); !ok {
		return
	}
	m := s.vnc
	m.mu.Lock()
	defer m.mu.Unlock()
	targets := make([]any, 0, len(m.targets))
	for _, t := range m.targets {
		targets = append(targets, map[string]any{"id": t.ID, "label": t.Label, "readOnly": t.ReadOnly})
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].(map[string]any)["id"].(string) < targets[j].(map[string]any)["id"].(string)
	})
	policy := "allowlist"
	if m.direct {
		policy = "allowlist+direct-opt-in"
	}
	p := map[string]any{"enabled": len(m.targets) > 0 || m.direct, "transport": "websocket", "ws_path": "/vnc/ws", "renderer": "placeholder", "host_policy": policy, "direct_connect_enabled": m.direct, "targets": targets}
	if ref := r.URL.Query().Get("target"); ref != "" {
		if t, ok := m.target(ref); ok {
			_, listed := m.targets[t.ID]
			p["target"] = map[string]any{"id": t.ID, "label": t.Label, "read_only": t.ReadOnly, "direct_connect": !listed}
		} else {
			writeJSON(w, 404, map[string]any{"error": "Unknown VNC target"})
			return
		}
	}
	writeJSON(w, 200, p)
}
func (s *Server) handleVNCHandoff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	owner, ok := s.vncOwner(w, r)
	if !ok {
		return
	}
	ref := r.URL.Query().Get("target")
	if ref == "" {
		var req struct {
			Target string `json:"target"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req); err != nil {
			writeJSON(w, 400, map[string]any{"error": "Target required"})
			return
		}
		ref = req.Target
	}
	m := s.vnc
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.target(ref)
	if !ok {
		writeJSON(w, 404, map[string]any{"error": "Unknown VNC target"})
		return
	}
	ts := m.sessions[owner+"\x00"+t.ID]
	if ts == nil || ts.client == nil {
		writeJSON(w, 409, map[string]any{"error": "No active VNC connection"})
		return
	}
	m.clearToken(ts)
	token := terminalID()
	ts.token = token
	ts.expires = time.Now().Add(m.ttl)
	m.handoffs[token] = ts
	ts.timer = time.AfterFunc(m.ttl, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if ts.token != token {
			return
		}
		m.clearToken(ts)
		if ts.client == nil {
			m.closeSession(ts)
		}
	})
	writeJSON(w, 200, map[string]any{"handoff": map[string]any{"token": token, "expires_at": ts.expires.UTC().Format(time.RFC3339Nano)}})
}
func (m *vncManager) clearToken(ts *vncSession) {
	if ts.token != "" {
		delete(m.handoffs, ts.token)
	}
	ts.token = ""
	if ts.timer != nil {
		ts.timer.Stop()
		ts.timer = nil
	}
}
func (m *vncManager) closeSession(ts *vncSession) {
	if m.sessions[ts.key] != ts {
		return
	}
	delete(m.sessions, ts.key)
	m.clearToken(ts)
	if ts.client != nil {
		ts.client.close()
		ts.client = nil
	}
	_ = ts.tcp.Close()
}
func (s *Server) attachVNC(ctx context.Context, owner, ref, token string, c *vncClient) (*vncSession, error) {
	m := s.vnc
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, errors.New("VNC shutting down")
	}
	target, ok := m.target(ref)
	if !ok {
		return nil, errors.New("Unknown VNC target")
	}
	var ts *vncSession
	if token != "" {
		ts = m.handoffs[token]
		if ts == nil || ts.owner != owner || ts.target.ID != target.ID || !time.Now().Before(ts.expires) {
			return nil, errors.New("Invalid or expired VNC handoff")
		}
		m.clearToken(ts)
		if ts.client != nil {
			ts.client.close()
		}
	}
	if ts == nil {
		key := owner + "\x00" + target.ID
		if old := m.sessions[key]; old != nil {
			m.closeSession(old)
		}
		if len(m.sessions)+m.dialing >= 16 {
			return nil, errors.New("VNC connection limit reached")
		}
		m.dialing++
		m.workers.Add(1) // Shutdown also waits for pending dials.
		m.mu.Unlock()
		dialer := net.Dialer{Timeout: 5 * time.Second}
		tcp, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(target.Host, strconv.Itoa(target.Port)))
		m.mu.Lock()
		m.dialing--
		m.workers.Done()
		if err != nil {
			return nil, errors.New("Cannot connect to VNC target")
		}
		if m.closed {
			tcp.Close()
			return nil, errors.New("VNC shutting down")
		}
		// Another same-owner attach may have completed while this dial ran.
		if old := m.sessions[key]; old != nil {
			m.closeSession(old)
		}
		ts = &vncSession{owner: owner, key: key, target: target, tcp: tcp, done: make(chan struct{})}
		m.sessions[key] = ts
		// Abort servers which accept TCP but never send their initial RFB greeting.
		_ = tcp.SetReadDeadline(time.Now().Add(m.greetingTimeout))
		m.workers.Add(1)
		go s.readVNC(ts)
	}
	ts.client = c
	payload, _ := json.Marshal(map[string]any{"type": "vnc.connected", "target": map[string]any{"id": target.ID, "label": target.Label}})
	c.send(vncFrame{websocket.MessageText, payload})
	return ts, nil
}
func (s *Server) readVNC(ts *vncSession) {
	defer s.vnc.workers.Done()
	defer close(ts.done)
	buffer := make([]byte, 32*1024)
	first := true
	for {
		n, err := ts.tcp.Read(buffer)
		if n > 0 {
			if first {
				_ = ts.tcp.SetReadDeadline(time.Time{})
				first = false
			}
			s.vnc.mu.Lock()
			if c := ts.client; c != nil {
				data := append([]byte(nil), buffer[:n]...)
				c.send(vncFrame{websocket.MessageBinary, data})
			}
			s.vnc.mu.Unlock()
		}
		if err != nil {
			s.vnc.mu.Lock()
			s.vnc.closeSession(ts)
			s.vnc.mu.Unlock()
			return
		}
	}
}
func (s *Server) handleVNCWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	owner, ok := s.vncOwner(w, r)
	if !ok {
		return
	}
	ref := r.URL.Query().Get("target")
	s.vnc.mu.Lock()
	_, ok = s.vnc.target(ref)
	s.vnc.mu.Unlock()
	if !ok {
		writeJSON(w, 404, map[string]any{"error": "Unknown VNC target"})
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024 * 1024)
	c := &vncClient{conn: conn, out: make(chan vncFrame, 64), stop: make(chan struct{})}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	ts, err := s.attachVNC(ctx, owner, ref, r.URL.Query().Get("handoff"), c)
	if err != nil {
		payload, _ := json.Marshal(map[string]any{"type": "vnc.error", "error": err.Error()})
		writeCtx, stop := context.WithTimeout(ctx, time.Second)
		_ = conn.Write(writeCtx, websocket.MessageText, payload)
		stop()
		_ = conn.Close(websocket.StatusPolicyViolation, err.Error())
		return
	}
	defer func() {
		c.close()
		s.vnc.mu.Lock()
		defer s.vnc.mu.Unlock()
		if ts.client != c {
			return
		}
		ts.client = nil
		if ts.token == "" {
			s.vnc.closeSession(ts)
		}
	}()
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer cancel()
		defer c.close()
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			case frame := <-c.out:
				if !s.vncAuthorized(r, owner) {
					return
				}
				writeCtx, stop := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Write(writeCtx, frame.kind, frame.data)
				stop()
				if err != nil {
					return
				}
			}
		}
	}()
	defer func() { cancel(); c.close(); <-writerDone }()
	for {
		kind, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if !s.vncAuthorized(r, owner) {
			return
		}
		if kind == websocket.MessageText {
			var p struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &p) == nil && p.Type == "ping" {
				c.send(vncFrame{websocket.MessageText, []byte(`{"type":"pong"}`)})
				continue
			}
		}
		s.vnc.mu.Lock()
		active := ts.client == c && s.vnc.sessions[ts.key] == ts
		s.vnc.mu.Unlock()
		if !active {
			return
		}
		_ = ts.tcp.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if _, err := writeVNCBytes(ts.tcp, data); err != nil {
			return
		}
	}
}
func writeVNCBytes(w io.Writer, data []byte) (int, error) {
	total := 0
	for len(data) > 0 {
		n, err := w.Write(data)
		total += n
		data = data[n:]
		if err != nil {
			return total, err
		}
		if n == 0 {
			return total, io.ErrShortWrite
		}
	}
	return total, nil
}
func (s *Server) CloseVNC() {
	m := s.vnc
	m.mu.Lock()
	m.closed = true
	for _, ts := range m.sessions {
		m.closeSession(ts)
	}
	m.mu.Unlock()
	m.workers.Wait()
}
