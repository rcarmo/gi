package web

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/rcarmo/gi/internal/environment"
	"github.com/rcarmo/gi/internal/shellenv"
	"github.com/rcarmo/gi/internal/tools"
)

const terminalReplayLimit = 2 * 1024 * 1024
const terminalSessionLimit = 16
const terminalFont = "FiraCode Nerd Font Mono"

var terminalClientID = regexp.MustCompile(`^[a-zA-Z0-9._:-]{8,128}$`)

type terminalClient struct {
	conn *websocket.Conn
	out  chan []byte
	stop chan struct{}
	once sync.Once
}

func (c *terminalClient) close() { c.once.Do(func() { close(c.stop); _ = c.conn.CloseNow() }) }
func (c *terminalClient) send(data []byte) {
	select {
	case c.out <- data:
	default:
		c.close()
	}
}

type terminalSession struct {
	id, owner, cwd, created string
	cmd                     *exec.Cmd
	pty                     *os.File
	mu                      sync.Mutex
	cols, rows              int
	clients                 map[*terminalClient]struct{}
	history                 []byte
	historyStart            int
	timer                   *time.Timer
	done                    chan struct{}
	stopped                 bool
}

type terminalHandoff struct {
	owner, session string
	expires        time.Time
}
type terminalManager struct {
	mu         sync.Mutex
	sessions   map[string]*terminalSession
	handoffs   map[string]terminalHandoff
	closed     bool
	workers    sync.WaitGroup
	grace, ttl time.Duration
}

func newTerminalManager() *terminalManager {
	return &terminalManager{sessions: map[string]*terminalSession{}, handoffs: map[string]terminalHandoff{}, grace: 3 * time.Second, ttl: 5 * time.Minute}
}
func terminalID() string {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func (s *Server) terminalOwner(w http.ResponseWriter, r *http.Request) (string, bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	// Shell access never accepts cross-origin requests, even when auth is disabled.
	if !providerWriteTransport(r) || !browserSameOrigin(r) {
		writeJSON(w, 403, map[string]any{"error": "Terminal requires same-origin HTTPS or localhost"})
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
	if enrolled {
		cookie, err := r.Cookie(browserSessionCookie)
		if err != nil || strings.TrimSpace(r.Header.Get("Authorization")) != "" || r.URL.Query().Has("auth_token") {
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
	client := strings.TrimSpace(r.URL.Query().Get("client"))
	if client == "" {
		client = strings.TrimSpace(r.Header.Get("x-piclaw-terminal-client"))
	}
	if !terminalClientID.MatchString(client) {
		writeJSON(w, 400, map[string]any{"error": "Terminal client identifier required"})
		return "", false
	}
	return "anonymous:" + client, true
}

func (s *Server) handleTerminalSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	owner, ok := s.terminalOwner(w, r)
	if !ok {
		return
	}
	shell, err := tools.ResolveShell(s.cfg.ShellPath)
	if !terminalSupported() || err != nil {
		writeJSON(w, 503, map[string]any{"error": "Web terminal unavailable"})
		return
	}
	s.terminals.mu.Lock()
	ts := s.terminals.sessions[owner]
	active := ts != nil
	clients := 0
	if ts != nil {
		ts.mu.Lock()
		clients = len(ts.clients)
		ts.mu.Unlock()
	}
	s.terminals.mu.Unlock()
	writeJSON(w, 200, map[string]any{"enabled": true, "transport": "websocket", "ws_path": "/terminal/ws", "cwd": s.workspaceRootPath(), "shell": shell.Shell + " -i", "font_family": terminalFont, "active": active, "connected_clients": clients})
}

func (s *Server) handleTerminalHandoff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	owner, ok := s.terminalOwner(w, r)
	if !ok {
		return
	}
	m := s.terminals
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweep()
	ts := m.sessions[owner]
	if ts == nil {
		writeJSON(w, 409, map[string]any{"error": "No active terminal session"})
		return
	}
	ts.mu.Lock()
	active := len(ts.clients) > 0
	ts.mu.Unlock()
	if !active {
		writeJSON(w, 409, map[string]any{"error": "No active terminal session"})
		return
	}
	// One pending token per owner prevents unbounded handoff records.
	for token, h := range m.handoffs {
		if h.owner == owner {
			delete(m.handoffs, token)
		}
	}
	token := terminalID()
	expires := time.Now().Add(m.ttl)
	m.handoffs[token] = terminalHandoff{owner, ts.id, expires}
	writeJSON(w, 200, map[string]any{"handoff": map[string]any{"token": token, "expires_at": expires.UTC().Format(time.RFC3339Nano)}})
}
func (m *terminalManager) sweep() {
	for token, h := range m.handoffs {
		if !time.Now().Before(h.expires) {
			delete(m.handoffs, token)
		}
	}
}

func (s *Server) handleTerminalWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	owner, ok := s.terminalOwner(w, r)
	if !ok {
		return
	}
	if !terminalSupported() {
		writeJSON(w, 503, map[string]any{"error": "Web terminal unavailable"})
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(64 * 1024)
	client := &terminalClient{conn: conn, out: make(chan []byte, 64), stop: make(chan struct{})}
	ts, err := s.attachTerminal(owner, r.URL.Query().Get("handoff"), client)
	if err != nil {
		_ = conn.Close(websocket.StatusPolicyViolation, err.Error())
		return
	}
	defer s.detachTerminal(ts, client)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer cancel()
		defer client.close()
		for {
			select {
			case <-ctx.Done():
				return
			case <-client.stop:
				return
			case data := <-client.out:
				if strings.HasPrefix(owner, "browser:") {
					cookie, err := r.Cookie(browserSessionCookie)
					if err != nil {
						return
					}
					valid, err := s.auth.ValidateTokenWithError(cookie.Value)
					if err != nil || !valid {
						return
					}
				}
				writeCtx, stop := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Write(writeCtx, websocket.MessageText, data)
				stop()
				if err != nil {
					return
				}
				if strings.Contains(string(data), `"type":"exit"`) {
					return
				}
			}
		}
	}()
	defer func() { cancel(); client.close(); <-writerDone }()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		// Revalidate cookie ownership before each control frame; logout blocks input.
		if strings.HasPrefix(owner, "browser:") {
			cookie, err := r.Cookie(browserSessionCookie)
			if err != nil {
				return
			}
			valid, err := s.auth.ValidateTokenWithError(cookie.Value)
			if err != nil || !valid {
				return
			}
		}
		if err := ts.control(client, data); err != nil {
			_ = conn.Close(websocket.StatusPolicyViolation, err.Error())
			return
		}
	}
}

func (s *Server) attachTerminal(owner, handoff string, client *terminalClient) (*terminalSession, error) {
	m := s.terminals
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweep()
	if m.closed {
		return nil, errors.New("Terminal shutting down")
	}
	ts := m.sessions[owner]
	transfer := false
	if handoff != "" {
		h, ok := m.handoffs[handoff]
		if !ok || h.owner != owner || ts == nil || h.session != ts.id {
			return nil, errors.New("Invalid terminal handoff")
		}
		delete(m.handoffs, handoff)
		transfer = true
	}
	if ts != nil {
		ts.mu.Lock()
		resume := len(ts.clients) == 0 && ts.timer != nil && !ts.stopped
		if transfer && ts.stopped {
			ts.mu.Unlock()
			return nil, errors.New("Terminal session closed")
		}
		if transfer || resume {
			if ts.timer != nil {
				ts.timer.Stop()
				ts.timer = nil
			}
			if transfer {
				for c := range ts.clients {
					c.close()
					delete(ts.clients, c)
				}
			}
			ts.mu.Unlock()
		} else {
			ts.mu.Unlock()
			ts.stop()
			delete(m.sessions, owner)
			ts = nil
		}
	}
	if ts == nil {
		if len(m.sessions) >= terminalSessionLimit {
			return nil, errors.New("Terminal session limit reached")
		}
		shell, err := tools.ResolveShell(s.cfg.ShellPath)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(shell.Shell, "-i")
		cmd.Dir = s.workspaceRootPath()
		env := os.Environ()
		if s.store != nil {
			overrides, loadErr := shellenv.Environment(s.store.DB()).Overrides(context.Background())
			if loadErr != nil {
				return nil, loadErr
			}
			env = environment.Apply(env, overrides)
		}
		// No implicit keychain injection; a browser shell inherits configured env only.
		env = append(env, "TERM=xterm-256color", "COLORTERM=truecolor", "COLUMNS=120", "LINES=30")
		cmd.Env = env
		file, err := startTerminalProcess(cmd, 120, 30)
		if err != nil {
			return nil, fmt.Errorf("Cannot start terminal: %w", err)
		}
		ts = &terminalSession{id: terminalID(), owner: owner, cwd: cmd.Dir, created: time.Now().UTC().Format(time.RFC3339Nano), cmd: cmd, pty: file, cols: 120, rows: 30, clients: map[*terminalClient]struct{}{}, done: make(chan struct{})}
		m.sessions[owner] = ts
		m.workers.Add(1)
		go s.runTerminal(ts)
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	metadata, _ := json.Marshal(map[string]any{"type": "session", "session_id": ts.id, "created_at": ts.created, "process_pid": ts.cmd.Process.Pid, "cwd": ts.cwd, "cols": ts.cols, "rows": ts.rows, "font_family": terminalFont})
	client.send(metadata)
	if len(ts.history) > 0 {
		data, _ := json.Marshal(map[string]any{"type": "output", "data": ts.replay()})
		client.send(data)
	}
	ts.clients[client] = struct{}{}
	return ts, nil
}

func (s *Server) runTerminal(ts *terminalSession) {
	defer s.terminals.workers.Done()
	defer close(ts.done)
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		buffer := make([]byte, 8192)
		var tail []byte
		for {
			n, err := ts.pty.Read(buffer)
			if n > 0 {
				text := append(tail, buffer[:n]...)
				// Decode complete runes; keep an incomplete UTF-8 suffix across PTY reads.
				end := 0
				for end < len(text) {
					if !utf8.FullRune(text[end:]) {
						break
					}
					_, size := utf8.DecodeRune(text[end:])
					end += size
				}
				tail = append([]byte(nil), text[end:]...)
				if end > 0 {
					ts.output(string(text[:end]))
				}
			}
			if err != nil {
				if len(tail) > 0 {
					ts.output(string(tail))
				}
				return
			}
		}
	}()
	err := ts.cmd.Wait()
	// Drain final output, but inherited slave descriptors must not keep a
	// terminated shell alive indefinitely.
	select {
	case <-readDone:
	case <-time.After(100 * time.Millisecond):
	}
	_ = ts.pty.Close()
	<-readDone
	code := 0
	if err != nil {
		code = ts.cmd.ProcessState.ExitCode()
	}
	payload, _ := json.Marshal(map[string]any{"type": "exit", "code": code, "signal": nil})
	m := s.terminals
	m.mu.Lock()
	ts.mu.Lock()
	ts.stopped = true
	if ts.timer != nil {
		ts.timer.Stop()
		ts.timer = nil
	}
	for c := range ts.clients {
		c.send(payload)
	}
	if m.sessions[ts.owner] == ts {
		delete(m.sessions, ts.owner)
	}
	for token, h := range m.handoffs {
		if h.session == ts.id {
			delete(m.handoffs, token)
		}
	}
	ts.mu.Unlock()
	m.mu.Unlock()
}
func (ts *terminalSession) output(text string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	original := text
	if len(text) >= terminalReplayLimit {
		ts.history = append(ts.history[:0], text[len(text)-terminalReplayLimit:]...)
		ts.historyStart = 0
	} else {
		// After growth, overwrite a ring instead of copying the complete
		// 2MiB replay window for every small output frame.
		n := min(terminalReplayLimit-len(ts.history), len(text))
		ts.history = append(ts.history, text[:n]...)
		text = text[n:]
		if len(text) > 0 {
			n = copy(ts.history[ts.historyStart:], text)
			copy(ts.history, text[n:])
			ts.historyStart = (ts.historyStart + len(text)) % terminalReplayLimit
		}
	}
	text = original
	if len(ts.clients) > 0 {
		data, _ := json.Marshal(struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}{"output", text})
		for c := range ts.clients {
			c.send(data)
		}
	}
}

// replay is called under ts.mu; the ring boundary may split a UTF-8 rune.
func (ts *terminalSession) replay() string {
	if len(ts.history) == 0 {
		return ""
	}
	var text string
	if ts.historyStart == 0 {
		text = string(ts.history)
	} else {
		var replay strings.Builder
		replay.Grow(len(ts.history))
		replay.Write(ts.history[ts.historyStart:])
		replay.Write(ts.history[:ts.historyStart])
		text = replay.String()
	}
	for len(text) > 0 && !utf8.RuneStart(text[0]) {
		text = text[1:]
	}
	return text
}

func (ts *terminalSession) control(client *terminalClient, data []byte) error {
	var p struct {
		Type string   `json:"type"`
		Data string   `json:"data"`
		Cols *float64 `json:"cols"`
		Rows *float64 `json:"rows"`
	}
	if json.Unmarshal(data, &p) != nil {
		p.Type = "input"
		p.Data = string(data)
	}
	ts.mu.Lock()
	if _, ok := ts.clients[client]; !ok || ts.stopped {
		ts.mu.Unlock()
		return errors.New("Terminal session closed")
	}
	switch p.Type {
	case "input":
		ts.mu.Unlock()
		if len(p.Data) > 0 {
			if err := ts.pty.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return err
			}
			_, err := io.WriteString(ts.pty, p.Data)
			return err
		}
		return nil
	case "resize":
		cols, rows := ts.cols, ts.rows
		if p.Cols != nil {
			cols = max(20, min(400, int(*p.Cols)))
		}
		if p.Rows != nil {
			rows = max(5, min(200, int(*p.Rows)))
		}
		if err := resizeTerminalProcess(ts.pty, cols, rows); err != nil {
			ts.mu.Unlock()
			return err
		}
		ts.cols, ts.rows = cols, rows
	}
	ts.mu.Unlock()
	return nil
}
func (ts *terminalSession) stop() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.stopped {
		return
	}
	ts.stopped = true
	if ts.timer != nil {
		ts.timer.Stop()
		ts.timer = nil
	}
	for c := range ts.clients {
		c.close()
	}
	stopTerminalProcess(ts.cmd)
	_ = ts.pty.Close()
}
func (s *Server) detachTerminal(ts *terminalSession, c *terminalClient) {
	c.close()
	m := s.terminals
	m.mu.Lock()
	defer m.mu.Unlock()
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.clients, c)
	if m.sessions[ts.owner] != ts || ts.stopped || len(ts.clients) > 0 {
		return
	}
	wait := m.grace
	m.sweep()
	for _, h := range m.handoffs {
		if h.session == ts.id {
			wait = max(wait, time.Until(h.expires))
		}
	}
	var timer *time.Timer
	timer = time.AfterFunc(wait, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.sessions[ts.owner] != ts {
			return
		}
		ts.mu.Lock()
		empty := len(ts.clients) == 0 && ts.timer == timer
		ts.mu.Unlock()
		if empty {
			ts.stop()
			delete(m.sessions, ts.owner)
		}
	})
	ts.timer = timer
}

// CloseTerminals joins all PTY processes; called during web-server shutdown.
func (s *Server) CloseTerminals() {
	m := s.terminals
	m.mu.Lock()
	m.closed = true
	sessions := make([]*terminalSession, 0, len(m.sessions))
	for _, ts := range m.sessions {
		sessions = append(sessions, ts)
		ts.stop()
	}
	clear(m.sessions)
	clear(m.handoffs)
	m.mu.Unlock()
	for _, ts := range sessions {
		<-ts.done
	}
	m.workers.Wait() // Includes replaced and grace-expired sessions.
}
