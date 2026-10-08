package web

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

const fileOpenTimeout = 15 * time.Second

type fileOpenPending struct {
	owner, session, path, target string
	result                       chan map[string]any
	deadline                     time.Time
	ctx                          context.Context
}
type fileOpenListener struct {
	owner, session string
	events         chan map[string]any
	valid          func() bool
}
type fileOpenRequests struct {
	mu        sync.Mutex
	closed    bool
	timeout   time.Duration
	pending   map[string]*fileOpenPending
	listeners map[*fileOpenListener]bool
}

func newFileOpenRequests() *fileOpenRequests {
	return &fileOpenRequests{timeout: fileOpenTimeout, pending: map[string]*fileOpenPending{}, listeners: map[*fileOpenListener]bool{}}
}

// UI requests require a browser cookie when enrolled. They never accept bearer
// tokens or cross-origin/unsafe transport even when authentication is disabled.
func (s *Server) fileOpenOwner(r *http.Request) (string, error) {
	if !providerWriteTransport(r) || !browserSameOrigin(r) {
		return "", errors.New("same-origin HTTPS or localhost required")
	}
	if strings.TrimSpace(r.Header.Get("Authorization")) != "" || r.URL.Query().Has("auth_token") {
		return "", errors.New("browser cookie required")
	}
	enrolled, err := s.auth.Enrolled()
	if err != nil {
		return "", err
	}
	if !enrolled {
		return "local-browser", nil
	}
	cookie, err := r.Cookie(browserSessionCookie)
	if err != nil {
		return "", errors.New("browser authentication required")
	}
	valid, err := s.auth.ValidateTokenWithError(cookie.Value)
	if err != nil || !valid {
		return "", errors.New("browser authentication required")
	}
	hash := sha256.Sum256([]byte(cookie.Value))
	return "browser:" + hex.EncodeToString(hash[:]), nil
}
func (s *Server) subscribeFileOpen(r *http.Request, session string) (<-chan map[string]any, func()) {
	owner, err := s.fileOpenOwner(r)
	if err != nil || session == "" {
		return nil, func() {}
	}
	m := s.fileOpens
	l := &fileOpenListener{owner: owner, session: session, events: make(chan map[string]any, 8), valid: func() bool { return true }}
	if strings.HasPrefix(owner, "browser:") {
		cookie, _ := r.Cookie(browserSessionCookie)
		token := cookie.Value
		l.valid = func() bool { ok, err := s.auth.ValidateTokenWithError(token); return err == nil && ok }
	}
	m.mu.Lock()
	if m.closed || len(m.listeners) >= 128 {
		m.mu.Unlock()
		return nil, func() {}
	}
	m.listeners[l] = true
	m.mu.Unlock()
	return l.events, func() { m.mu.Lock(); delete(m.listeners, l); m.mu.Unlock() }
}
func (s *Server) CloseFileOpenRequests() {
	m := s.fileOpens
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	for id, p := range m.pending {
		p.result <- map[string]any{"ok": false, "opened": false, "reason": "shutdown", "path": p.path, "target": p.target}
		delete(m.pending, id)
	}
}
func (s *Server) registerFileOpenTool() {
	if s.turns == nil {
		return
	}
	err := s.turns.RegisterTool(tools.RegisteredTool{Name: "open_workspace_file", Description: "Open an existing workspace file in the current browser editor tab or popout; waits up to 15 seconds for the browser outcome.", Kind: "read-only", Weight: "lightweight", Parameters: json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"target":{"type":"string","enum":["tab","popout"]},"label":{"type":"string"}},"required":["path"],"additionalProperties":false}`), Executor: func(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
		path, _ := call.Arguments["path"].(string)
		target, _ := call.Arguments["target"].(string)
		label, _ := call.Arguments["label"].(string)
		result, err := s.requestFileOpen(ctx, rt.SessionID, path, target, label)
		if err != nil {
			return "", err
		}
		raw, err := json.Marshal(result)
		return string(raw), err
	}})
	if err != nil {
		panic(err)
	}
}
func (s *Server) requestFileOpen(ctx context.Context, session, path, target, label string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if session == "" || s.store == nil {
		return nil, errors.New("current session required")
	}
	if _, err := s.store.GetSession(ctx, session); err != nil {
		return nil, errors.New("unknown session")
	}
	if target == "" {
		target = "tab"
	}
	if target != "tab" && target != "popout" {
		return nil, errors.New("invalid editor target")
	}
	if len(label) > 256 {
		return nil, errors.New("label too long")
	}
	f, _, relative, err := s.openWorkspaceFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open workspace file: %w", err)
	}
	f.Close()
	m := s.fileOpens
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, errors.New("browser requests closed")
	}
	if len(m.pending) >= 32 {
		m.mu.Unlock()
		return nil, errors.New("too many browser requests")
	}
	owner := ""
	var listeners []*fileOpenListener
	for l := range m.listeners {
		if l.session != session || !l.valid() {
			continue
		}
		if owner != "" && owner != l.owner {
			m.mu.Unlock()
			return nil, errors.New("ambiguous browser owner")
		}
		owner = l.owner
		listeners = append(listeners, l)
	}
	if owner == "" {
		m.mu.Unlock()
		return nil, errors.New("no browser connected to current session")
	}
	id := terminalID()
	timeout := m.timeout
	if timeout <= 0 || timeout > fileOpenTimeout {
		timeout = fileOpenTimeout
	}
	p := &fileOpenPending{owner: owner, session: session, path: relative, target: target, result: make(chan map[string]any, 1), deadline: time.Now().Add(timeout), ctx: ctx}
	m.pending[id] = p
	event := map[string]any{"type": "extension_ui_request", "chat_jid": "gi:" + session, "request_id": id, "method": "custom", "kind": "custom", "options": map[string]any{"action": "open_workspace_file", "path": relative, "target": target, "label": label, "timeout": 15000}}
	delivered := false
	for _, l := range listeners {
		select {
		case l.events <- event:
			delivered = true
		default:
		}
	}
	m.mu.Unlock()
	defer func() { m.mu.Lock(); delete(m.pending, id); m.mu.Unlock() }()
	if !delivered {
		return nil, errors.New("browser request queue full")
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case result := <-p.result:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return map[string]any{"ok": false, "opened": false, "reason": "timeout", "path": relative, "target": target}, nil
	}
}
func (s *Server) handleAgentRespond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	owner, err := s.fileOpenOwner(r)
	if err != nil {
		writeJSON(w, 403, map[string]any{"error": err.Error()})
		return
	}
	var req struct {
		RequestID string         `json:"request_id"`
		ChatJID   string         `json:"chat_jid"`
		Outcome   map[string]any `json:"outcome"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, 400, map[string]any{"error": "invalid response"})
		return
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		writeJSON(w, 400, map[string]any{"error": "invalid response"})
		return
	}
	m := s.fileOpens
	m.mu.Lock()
	defer m.mu.Unlock()
	p := m.pending[req.RequestID]
	if p == nil || p.owner != owner || req.ChatJID != "gi:"+p.session || p.ctx.Err() != nil || !time.Now().Before(p.deadline) {
		writeJSON(w, 404, map[string]any{"error": "request not found"})
		return
	}
	ok, okType := req.Outcome["ok"].(bool)
	opened, openedType := req.Outcome["opened"].(bool)
	if !okType || !openedType || req.Outcome["path"] != p.path || req.Outcome["target"] != p.target || opened && !ok {
		writeJSON(w, 400, map[string]any{"error": "outcome scope mismatch"})
		return
	}
	select {
	case p.result <- req.Outcome:
		delete(m.pending, req.RequestID)
		writeJSON(w, 200, map[string]any{"ok": true})
	default:
		writeJSON(w, 409, map[string]any{"error": "response already received"})
	}
}
