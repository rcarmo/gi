package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rcarmo/gi/internal/inference"
)

// Server connection states, as shown by /mcp.
const (
	StateDisabled     = "disabled"
	StateDisconnected = "disconnected"
	StateConnecting   = "connecting"
	StateConnected    = "connected"
	StateFailed       = "failed"
	StateNeedsAuth    = "needs-auth"
)

const (
	clientName       = "gi"
	maxFrameBytes    = 16 << 20 // largest inbound JSON-RPC frame from a stdio server
	maxToolPages     = 64
	stderrTailBytes  = 8 << 10
	terminateGrace   = 2 * time.Second
	httpConnectTries = 3 // first attempt + Pi's two retries on transient errors
)

// Status is a point-in-time view of one server, for /mcp and diagnostics.
type Status struct {
	Name         string
	Transport    string
	State        string
	Error        string
	Tools        int
	Exposure     string
	Source       string
	Scope        string // "global" or "project"
	Override     string // the project mcp.json overriding a global server
	Endpoint     string // Pi's describeTransport: the URL, or the command and arguments
	Instructions string
	StderrTail   string
}

// Manager owns the configured MCP servers for one gi process. Connections are
// lazy (or started in the background by ConnectAll), shared by concurrent
// callers, and re-established after a dropped connection on the next call.
type Manager struct {
	cfg       Config
	workspace string
	log       *rotatingLog
	lookupEnv func(string) (string, bool)

	mu             sync.Mutex
	servers        map[string]*server
	closed         bool
	onToolsChanged func(server string)
	credentials    *CredentialStore // OAuth (mcp-auth.json); nil: no OAuth
}

// SetCredentials enables OAuth for HTTP servers without an Authorization
// header, with credentials in store (Pi's mcp-auth.json).
func (m *Manager) SetCredentials(store *CredentialStore) { m.credentials = store }

// Credentials returns the OAuth credential store (nil without OAuth).
func (m *Manager) Credentials() *CredentialStore { return m.credentials }

// SetOnToolsChanged registers a callback run (in its own goroutine) when a
// server announces a changed tool list, so callers can re-register tools.
func (m *Manager) SetOnToolsChanged(fn func(server string)) {
	m.mu.Lock()
	m.onToolsChanged = fn
	m.mu.Unlock()
}

type server struct {
	cfg ServerConfig

	mu           sync.Mutex // serializes connect/disconnect
	session      *mcp.ClientSession
	closeConn    func()
	state        string
	err          error
	tools        []*mcp.Tool
	toolsValid   bool
	instructions string
	stderr       *tailBuffer
	auth         *connectionAuth // OAuth servers
	needsAuth    atomic.Bool     // the last connection needed a sign-in
}

// NewManager prepares servers without connecting. logPath receives server
// logging notifications (Pi: ~/.pi/agent/mcp.log); empty disables logging.
func NewManager(cfg Config, workspace, logPath string) *Manager {
	m := &Manager{cfg: cfg, workspace: workspace, lookupEnv: os.LookupEnv, servers: map[string]*server{}}
	if logPath != "" {
		m.log = newRotatingLog(logPath, 5<<20)
	}
	for name, sc := range cfg.Servers {
		state := StateDisconnected
		if !sc.Enabled {
			state = StateDisabled
		}
		m.servers[name] = &server{cfg: sc, state: state, stderr: &tailBuffer{max: stderrTailBytes}}
	}
	return m
}

// Config returns the configuration the manager was built from.
func (m *Manager) Config() Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

// ServerPatch is a change to one server's settings (Pi's saveConfig patch):
// Enabled when non-nil, Exposure when non-empty. InProject saves it as a
// project override of a global server (Pi's inProject).
type ServerPatch struct {
	Enabled   *bool
	Exposure  string
	InProject bool
}

// UpdateServer ports Pi's saveConfig: the change is written to the mcp.json
// that defines the server, or to the project mcp.json that overrides it
// (other content kept, see UpdateServerConfig), then applied here. Disabling
// closes the server's connection; enabling leaves it to connect on next use.
func (m *Manager) UpdateServer(name string, patch ServerPatch) error {
	s, err := m.server(name)
	if err != nil {
		return err
	}
	m.mu.Lock()
	sc := m.cfg.Servers[name]
	if patch.InProject {
		sc.Override = m.cfg.ProjectConfig
	}
	m.mu.Unlock()
	path := sc.Source
	if sc.Override != "" {
		path = sc.Override
	}
	if err := UpdateServerConfig(path, name, patch, sc.Override != ""); err != nil {
		return fmt.Errorf("Could not update %s: %w", path, err)
	}
	m.mu.Lock()
	if patch.Enabled != nil {
		sc.Enabled = *patch.Enabled
	}
	if patch.Exposure != "" {
		sc.Exposure = patch.Exposure
	}
	// Copy on write: callers may be reading the previous map.
	servers := make(map[string]ServerConfig, len(m.cfg.Servers))
	for k, v := range m.cfg.Servers {
		servers[k] = v
	}
	servers[name] = sc
	m.cfg.Servers = servers
	m.mu.Unlock()
	s.mu.Lock()
	s.cfg = sc
	closeConn := s.closeConn
	if !sc.Enabled {
		s.session, s.closeConn, s.toolsValid, s.tools = nil, nil, false, nil
		s.state, s.err = StateDisabled, nil
	} else if s.state == StateDisabled {
		s.state, closeConn = StateDisconnected, nil
	} else {
		closeConn = nil
	}
	s.mu.Unlock()
	if closeConn != nil {
		closeConn()
	}
	return nil
}

func (m *Manager) server(name string) (*server, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, errors.New("mcp: manager closed")
	}
	s, ok := m.servers[name]
	if !ok {
		return nil, fmt.Errorf("mcp: unknown server %q", name)
	}
	return s, nil
}

// Status reports every configured server, sorted by name.
func (m *Manager) Status() []Status {
	var out []Status
	for _, name := range m.Config().Names() {
		s, err := m.server(name)
		if err != nil {
			continue
		}
		s.mu.Lock()
		st := Status{Name: name, Transport: s.cfg.Transport, State: s.state, Tools: len(s.tools), Exposure: s.cfg.Exposure,
			Source: s.cfg.Source, Scope: s.cfg.Scope, Override: s.cfg.Override, Endpoint: describeTransport(s.cfg), Instructions: s.instructions, StderrTail: s.stderr.String()}
		if s.err != nil {
			st.Error = s.err.Error()
		}
		s.mu.Unlock()
		out = append(out, st)
	}
	return out
}

// ConnectAll starts connecting every enabled server in the background and
// returns immediately (Pi connects servers when a session starts).
func (m *Manager) ConnectAll(ctx context.Context) {
	for _, name := range m.Config().Names() {
		if s, err := m.server(name); err == nil && s.cfg.Enabled {
			go func(s *server) { _, _ = m.ensureSession(ctx, s) }(s)
		}
	}
}

// Instructions returns the server's initialize instructions (empty until it
// has connected).
func (m *Manager) Instructions(name string) string {
	s, err := m.server(name)
	if err != nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.instructions
}

// Tools lists the server's tools, connecting if needed. The list is cached
// until the server announces a change or the connection drops.
func (m *Manager) Tools(ctx context.Context, name string) ([]*mcp.Tool, error) {
	s, err := m.server(name)
	if err != nil {
		return nil, err
	}
	session, err := m.ensureSession(ctx, s)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	if s.toolsValid && s.session == session {
		tools := s.tools
		s.mu.Unlock()
		return tools, nil
	}
	s.mu.Unlock()
	tools, err := listAllTools(ctx, session, s.cfg.Timeout)
	if err != nil {
		m.noteFailure(s, session, err)
		return nil, fmt.Errorf("mcp %s: tools/list: %w", name, err)
	}
	s.mu.Lock()
	if s.session == session {
		s.tools, s.toolsValid = tools, true
	}
	s.mu.Unlock()
	return tools, nil
}

// CallTool calls one server tool. Tool calls are never retried: the server
// may already have performed them (Pi). A result with IsError is returned as
// a result, not an error.
func (m *Manager) CallTool(ctx context.Context, name, tool string, args any) (*mcp.CallToolResult, error) {
	s, err := m.server(name)
	if err != nil {
		return nil, err
	}
	session, err := m.ensureSession(ctx, s)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()
	params := &mcp.CallToolParams{Name: tool, Arguments: args}
	if m.log != nil {
		// Protocol 2026-07-28 carries the logging level per request; older
		// servers use the session level set after connecting.
		params.Meta = mcp.Meta{mcp.MetaKeyLogLevel: "info"}
	}
	result, err := session.CallTool(callCtx, params)
	if err != nil {
		if ctx.Err() == nil && callCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("mcp %s: %s timed out after %s", name, tool, s.cfg.Timeout)
		}
		m.noteFailure(s, session, err)
		return nil, fmt.Errorf("mcp %s: %s: %w", name, tool, err)
	}
	return result, nil
}

// Close disconnects every server (stdio: close stdin, SIGTERM, then SIGKILL
// to the process group).
func (m *Manager) Close() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	servers := make([]*server, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.Unlock()
	var wg sync.WaitGroup
	for _, s := range servers {
		wg.Add(1)
		go func(s *server) {
			defer wg.Done()
			s.mu.Lock()
			closeConn := s.closeConn
			s.session, s.closeConn, s.toolsValid = nil, nil, false
			if s.state != StateDisabled {
				s.state = StateDisconnected
			}
			s.mu.Unlock()
			if closeConn != nil {
				closeConn()
			}
		}(s)
	}
	wg.Wait()
	if m.log != nil {
		m.log.Close()
	}
}

// noteFailure drops a session after a transport failure so the next call
// reconnects. Errors the server returned (JSON-RPC errors) and our own
// cancellations keep the session.
func (m *Manager) noteFailure(s *server, session *mcp.ClientSession, err error) {
	var rpcErr *jsonrpc.Error
	if errors.As(err, &rpcErr) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	s.mu.Lock()
	if s.session != session {
		s.mu.Unlock()
		return
	}
	closeConn := s.closeConn
	s.session, s.closeConn, s.toolsValid = nil, nil, false
	s.state, s.err = StateDisconnected, err
	s.mu.Unlock()
	if closeConn != nil {
		go closeConn()
	}
}

func (m *Manager) ensureSession(ctx context.Context, s *server) (*mcp.ClientSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.cfg.Enabled {
		return nil, fmt.Errorf("mcp %s: server is disabled", s.cfg.Name)
	}
	if s.session != nil {
		return s.session, nil
	}
	s.state, s.err = StateConnecting, nil
	s.needsAuth.Store(false)
	session, closeConn, err := m.connect(ctx, s)
	if err != nil {
		if s.needsAuth.Load() || errors.Is(err, ErrAuthorizationRequired) {
			s.state, s.err = StateNeedsAuth, nil
			return nil, fmt.Errorf("mcp %s: %w", s.cfg.Name, ErrAuthorizationRequired)
		}
		s.state, s.err = StateFailed, err
		return nil, fmt.Errorf("mcp %s: %w", s.cfg.Name, err)
	}
	s.session, s.closeConn, s.state = session, closeConn, StateConnected
	if res := session.InitializeResult(); res != nil {
		s.instructions = strings.TrimSpace(res.Instructions)
		// Servers send logging notifications only after the client sets a
		// level; Pi records them in mcp.log.
		if m.log != nil && res.Capabilities != nil && res.Capabilities.Logging != nil {
			levelCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
			_ = session.SetLoggingLevel(levelCtx, &mcp.SetLoggingLevelParams{Level: "info"})
			cancel()
		}
	}
	return session, nil
}

func (m *Manager) clientFor(s *server) *mcp.Client {
	name := s.cfg.Name
	return mcp.NewClient(&mcp.Implementation{Name: clientName, Version: "1"}, &mcp.ClientOptions{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		ToolListChangedHandler: func(context.Context, *mcp.ToolListChangedRequest) {
			s.mu.Lock()
			s.toolsValid = false // new tools appear, withdrawn ones become unreachable
			s.mu.Unlock()
			m.mu.Lock()
			fn := m.onToolsChanged
			m.mu.Unlock()
			if fn != nil {
				go fn(name)
			}
		},
		LoggingMessageHandler: func(_ context.Context, req *mcp.LoggingMessageRequest) {
			if m.log != nil && req != nil && req.Params != nil {
				m.log.Printf("[%s] %s %s: %v", name, req.Params.Level, req.Params.Logger, req.Params.Data)
			}
		},
	})
}

func (m *Manager) connect(ctx context.Context, s *server) (*mcp.ClientSession, func(), error) {
	connectCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()
	if s.cfg.Transport == "http" {
		return m.connectHTTP(connectCtx, s)
	}
	return m.connectStdio(connectCtx, s)
}

func (m *Manager) connectStdio(ctx context.Context, s *server) (*mcp.ClientSession, func(), error) {
	cfg := s.cfg
	env := os.Environ()
	for key, value := range cfg.Env {
		expanded, err := expandValue(ctx, value, m.lookupEnv)
		if err != nil {
			return nil, nil, fmt.Errorf("env %s: %w", key, err)
		}
		env = append(env, key+"="+expanded)
	}
	args := make([]string, len(cfg.Args))
	for i, a := range cfg.Args {
		args[i] = expandHome(a)
	}
	// The process outlives this connect call; it is stopped by closeConn.
	cmd := exec.Command(expandHome(cfg.Command), args...)
	cmd.Env = env
	cmd.Dir = m.resolveCwd(cfg.Cwd)
	cmd.Stderr = s.stderr
	setProcessGroup(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start %s: %w", cfg.Command, err)
	}
	exited := make(chan struct{})
	go func() { _ = cmd.Wait(); close(exited) }()
	var once sync.Once
	closeConn := func() {
		once.Do(func() {
			// Pi: close stdin, then SIGTERM, then SIGKILL to the process group.
			// The group is signalled even when the server itself exits, since
			// wrappers such as npx or uvx can leave children behind.
			_ = stdin.Close()
			select {
			case <-exited:
			case <-time.After(terminateGrace):
			}
			signalProcessGroup(cmd, false)
			// Give the group the grace period to exit after SIGTERM.
			deadline := time.Now().Add(terminateGrace)
			for time.Now().Before(deadline) && processGroupAlive(cmd) {
				time.Sleep(25 * time.Millisecond)
			}
			if processGroupAlive(cmd) {
				signalProcessGroup(cmd, true)
			}
			<-exited
		})
	}
	session, err := m.clientFor(s).Connect(ctx, &mcp.IOTransport{Reader: stdout, Writer: stdin, MaxLineLength: maxFrameBytes}, nil)
	if err != nil {
		closeConn()
		if tail := strings.TrimSpace(s.stderr.String()); tail != "" {
			return nil, nil, fmt.Errorf("initialize: %w (stderr: %s)", err, lastLine(tail))
		}
		return nil, nil, fmt.Errorf("initialize: %w", err)
	}
	return session, func() { _ = session.Close(); closeConn() }, nil
}

func (m *Manager) connectHTTP(ctx context.Context, s *server) (*mcp.ClientSession, func(), error) {
	// Revalidate explicit in-memory configurations as well as parsed files.
	if s.cfg.Auth != nil {
		if err := validateProviderURL(s.cfg.URL); err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(s.cfg.Auth.Provider) == "" {
			return nil, nil, errors.New("provider auth requires a provider")
		}
		if len(s.cfg.OAuth) > 0 && string(s.cfg.OAuth) != "null" {
			return nil, nil, errors.New("auth.provider and oauth are mutually exclusive")
		}
		for key := range s.cfg.Headers {
			if strings.EqualFold(key, "Authorization") {
				return nil, nil, errors.New("auth.provider and Authorization header are mutually exclusive")
			}
		}
	}
	headers := http.Header{}
	for key, value := range s.cfg.Headers {
		expanded, err := expandValue(ctx, value, m.lookupEnv)
		if err != nil {
			return nil, nil, fmt.Errorf("header %s: %w", key, err)
		}
		headers.Set(key, expanded)
	}
	var rt http.RoundTripper = headerTransport{base: http.DefaultTransport, headers: headers}
	if s.cfg.UsesOAuth() && m.credentials != nil {
		if s.auth == nil {
			cfg := s.cfg
			s.auth = &connectionAuth{name: cfg.Name, serverURL: normalizeServerURL(cfg.URL), store: m.credentials,
				settings: func() (OAuthSettings, error) { return m.oauthSettings(context.Background(), cfg) }}
		}
		rt = &oauthTransport{base: rt, auth: s.auth, onNeedsAuth: func() { s.needsAuth.Store(true) }}
	}
	client := &http.Client{Transport: rt}
	if s.cfg.Auth != nil {
		endpoint, _ := url.Parse(s.cfg.URL)
		client.Transport = providerTransport{base: rt, endpoint: endpoint, provider: s.cfg.Auth.Provider, token: inference.ProviderToken}
		client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return errors.New("provider-auth MCP redirects are disabled")
		}
	}
	var lastErr error
	for attempt := 0; attempt < httpConnectTries; attempt++ {
		transport := &mcp.StreamableClientTransport{Endpoint: s.cfg.URL, HTTPClient: client}
		session, err := m.clientFor(s).Connect(ctx, transport, nil)
		if err == nil {
			return session, func() { _ = session.Close() }, nil
		}
		lastErr = err
		if !transientHTTPError(err) || ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(attempt+1) * 250 * time.Millisecond):
		}
	}
	return nil, nil, fmt.Errorf("initialize: %w", lastErr)
}

// transientHTTPError matches network failures and the statuses Pi retries
// (408, 429, 5xx) when connecting or listing; tool calls are never retried.
func transientHTTPError(err error) bool {
	msg := err.Error()
	for _, code := range []string{"408", "429", "500", "502", "503", "504"} {
		if strings.Contains(msg, code) {
			return true
		}
	}
	return strings.Contains(msg, "connection refused") || strings.Contains(msg, "connection reset") || strings.Contains(msg, "EOF")
}

type headerTransport struct {
	base    http.RoundTripper
	headers http.Header
}

func (t headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for key, values := range t.headers {
		for _, v := range values {
			req.Header.Set(key, v)
		}
	}
	return t.base.RoundTrip(req)
}

func (m *Manager) resolveCwd(cwd string) string {
	cwd = expandHome(strings.TrimSpace(cwd))
	switch {
	case cwd == "":
		return m.workspace
	case filepath.IsAbs(cwd):
		return cwd
	default: // Pi: relative cwd values resolve against the session directory
		return filepath.Join(m.workspace, cwd)
	}
}

func listAllTools(ctx context.Context, session *mcp.ClientSession, timeout time.Duration) ([]*mcp.Tool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var tools []*mcp.Tool
	cursor := ""
	for page := 0; page < maxToolPages; page++ {
		res, err := session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return nil, err
		}
		tools = append(tools, res.Tools...)
		if res.NextCursor == "" {
			return tools, nil
		}
		cursor = res.NextCursor
	}
	return nil, fmt.Errorf("more than %d pages of tools", maxToolPages)
}

func lastLine(s string) string {
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return s
}

// tailBuffer keeps the last max bytes written (stderr of a stdio server).
type tailBuffer struct {
	mu  sync.Mutex
	max int
	buf []byte
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, p...)
	if len(t.buf) > t.max {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-t.max:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.buf)
}

// rotatingLog appends lines to a file and moves it to <file>.1 past maxBytes
// (Pi: mcp.log -> mcp.log.1 at 5 MB).
type rotatingLog struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	f        *os.File
}

func newRotatingLog(path string, maxBytes int64) *rotatingLog {
	return &rotatingLog{path: path, maxBytes: maxBytes}
}

func (l *rotatingLog) Printf(format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
			return
		}
		f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return
		}
		l.f = f
	}
	line := time.Now().UTC().Format(time.RFC3339) + " " + fmt.Sprintf(format, args...)
	_, _ = l.f.WriteString(strings.ReplaceAll(line, "\n", " ") + "\n")
	if info, err := l.f.Stat(); err == nil && info.Size() > l.maxBytes {
		_ = l.f.Close()
		l.f = nil
		_ = os.Rename(l.path, l.path+".1")
	}
}

func (l *rotatingLog) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f != nil {
		_ = l.f.Close()
		l.f = nil
	}
}

// HasResources reports whether a connected server offers resources.
func (m *Manager) HasResources(name string) bool {
	s, err := m.server(name)
	if err != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.session == nil {
		return false
	}
	res := s.session.InitializeResult()
	return res != nil && res.Capabilities != nil && res.Capabilities.Resources != nil
}

// ListResources lists one page of a server's resources. Listing and reading
// resources is retried once after a transient error (Pi).
func (m *Manager) ListResources(ctx context.Context, name, cursor string) (*mcp.ListResourcesResult, error) {
	return withSession(ctx, m, name, func(ctx context.Context, cs *mcp.ClientSession) (*mcp.ListResourcesResult, error) {
		return cs.ListResources(ctx, &mcp.ListResourcesParams{Cursor: cursor})
	})
}

// ListResourceTemplates lists one page of a server's resource templates.
func (m *Manager) ListResourceTemplates(ctx context.Context, name, cursor string) (*mcp.ListResourceTemplatesResult, error) {
	return withSession(ctx, m, name, func(ctx context.Context, cs *mcp.ClientSession) (*mcp.ListResourceTemplatesResult, error) {
		return cs.ListResourceTemplates(ctx, &mcp.ListResourceTemplatesParams{Cursor: cursor})
	})
}

// ReadResource reads one resource by URI.
func (m *Manager) ReadResource(ctx context.Context, name, uri string) (*mcp.ReadResourceResult, error) {
	return withSession(ctx, m, name, func(ctx context.Context, cs *mcp.ClientSession) (*mcp.ReadResourceResult, error) {
		return cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
	})
}

func withSession[T any](ctx context.Context, m *Manager, name string, call func(context.Context, *mcp.ClientSession) (T, error)) (T, error) {
	var zero T
	s, err := m.server(name)
	if err != nil {
		return zero, err
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		session, err := m.ensureSession(ctx, s)
		if err != nil {
			return zero, err
		}
		callCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
		out, err := call(callCtx, session)
		cancel()
		if err == nil {
			return out, nil
		}
		lastErr = err
		m.noteFailure(s, session, err)
		if ctx.Err() != nil || !transientHTTPError(err) {
			break
		}
	}
	return zero, fmt.Errorf("mcp %s: %w", name, lastErr)
}

// Reconnect drops the server's connection and connects again, listing its
// tools (Pi's /mcp reconnect).
func (m *Manager) Reconnect(ctx context.Context, name string) error {
	s, err := m.server(name)
	if err != nil {
		return err
	}
	s.mu.Lock()
	if !s.cfg.Enabled {
		s.mu.Unlock()
		return fmt.Errorf("MCP server %q is disabled.", name)
	}
	closeConn := s.closeConn
	s.session, s.closeConn, s.toolsValid = nil, nil, false
	s.state, s.err = StateDisconnected, nil
	s.mu.Unlock()
	if closeConn != nil {
		closeConn()
	}
	_, err = m.Tools(ctx, name)
	return err
}
