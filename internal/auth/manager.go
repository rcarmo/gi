package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type State struct {
	Username    string    `json:"username"`
	TOTPSecret  string    `json:"totp_secret"`
	TOTPEnabled bool      `json:"totp_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Sessions    []Session `json:"sessions,omitempty"`
}

type Session struct {
	TokenHash string    `json:"token_hash"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PendingEnrollment struct {
	Username  string    `json:"username"`
	Secret    string    `json:"secret"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Manager struct {
	mu      sync.Mutex
	path    string
	issuer  string
	pending map[string]PendingEnrollment
}

func NewManager(workspaceRoot string) *Manager {
	return &Manager{path: filepath.Join(workspaceRoot, ".gi", "auth.json"), issuer: "gi", pending: make(map[string]PendingEnrollment)}
}

func (m *Manager) Status() (map[string]any, error) {
	state, err := m.load()
	if os.IsNotExist(err) {
		return map[string]any{"enrolled": false, "enrollment_required": true, "totp_enabled": false}, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"enrolled": state.TOTPEnabled && state.Username != "", "enrollment_required": !(state.TOTPEnabled && state.Username != ""), "username": state.Username, "totp_enabled": state.TOTPEnabled}, nil
}

func (m *Manager) StartEnrollment(username string) (PendingEnrollment, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}
	enrolled, err := m.Enrolled()
	if err != nil {
		return PendingEnrollment{}, err
	}
	if enrolled {
		return PendingEnrollment{}, fmt.Errorf("enrollment is already complete")
	}
	secret, err := GenerateTOTPSecret()
	if err != nil {
		return PendingEnrollment{}, err
	}
	pending := PendingEnrollment{Username: username, Secret: secret, URL: TOTPURL(m.issuer, username, secret), CreatedAt: time.Now().UTC()}
	m.mu.Lock()
	m.pruneExpiredPendingLocked(pending.CreatedAt)
	m.pending[username] = pending
	m.mu.Unlock()
	return pending, nil
}

func (m *Manager) VerifyEnrollment(username, code string) (State, error) {
	username = strings.TrimSpace(username)
	m.mu.Lock()
	pending, ok := m.pending[username]
	m.mu.Unlock()
	if !ok {
		return State{}, fmt.Errorf("no pending enrollment for user")
	}
	if time.Since(pending.CreatedAt) > 10*time.Minute {
		m.mu.Lock()
		m.pruneExpiredPendingLocked(time.Now().UTC())
		m.mu.Unlock()
		return State{}, fmt.Errorf("pending enrollment expired")
	}
	if !VerifyTOTP(pending.Secret, code, time.Now().UTC(), 1) {
		return State{}, fmt.Errorf("invalid TOTP code")
	}
	now := time.Now().UTC()
	state := State{Username: username, TOTPSecret: pending.Secret, TOTPEnabled: true, CreatedAt: now, UpdatedAt: now}
	if err := m.save(state); err != nil {
		return State{}, err
	}
	m.mu.Lock()
	delete(m.pending, username)
	m.mu.Unlock()
	return state, nil
}

func (m *Manager) VerifyLogin(username, code string) (string, time.Time, error) {
	state, err := m.load()
	if err != nil {
		return "", time.Time{}, err
	}
	if username != "" && username != state.Username {
		return "", time.Time{}, fmt.Errorf("invalid user")
	}
	if !state.TOTPEnabled || state.TOTPSecret == "" {
		return "", time.Time{}, fmt.Errorf("TOTP is not enrolled")
	}
	if !VerifyTOTP(state.TOTPSecret, code, time.Now().UTC(), 1) {
		return "", time.Time{}, fmt.Errorf("invalid TOTP code")
	}
	token, hash, err := newToken()
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now().UTC()
	expires := now.Add(12 * time.Hour)
	state.Sessions = append(pruneSessions(state.Sessions, now), Session{TokenHash: hash, CreatedAt: now, ExpiresAt: expires})
	state.UpdatedAt = now
	if err := m.save(state); err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (m *Manager) ValidateBearerRequest(r *http.Request) bool {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	if token == "" {
		token = r.URL.Query().Get("auth_token")
	}
	return m.ValidateToken(token)
}

func (m *Manager) ValidateToken(token string) bool {
	if strings.TrimSpace(token) == "" {
		return false
	}
	state, err := m.load()
	if err != nil {
		return false
	}
	hash := hashToken(token)
	now := time.Now().UTC()
	for _, sess := range state.Sessions {
		if sess.ExpiresAt.After(now) && constantTimeString(sess.TokenHash, hash) {
			return true
		}
	}
	return false
}

func (m *Manager) Enrolled() (bool, error) {
	state, err := m.load()
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state.TOTPEnabled && state.Username != "", nil
}

func (m *Manager) load() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var state State
	data, err := os.ReadFile(m.path)
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	return state, nil
}

func (m *Manager) save(state State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(m.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0o600)
}

func (m *Manager) pruneExpiredPendingLocked(now time.Time) {
	cutoff := now.Add(-10 * time.Minute)
	for username, pending := range m.pending {
		if pending.CreatedAt.Before(cutoff) {
			delete(m.pending, username)
		}
	}
}

func newToken() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(raw[:])
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func pruneSessions(sessions []Session, now time.Time) []Session {
	out := sessions[:0]
	for _, sess := range sessions {
		if sess.ExpiresAt.After(now) {
			out = append(out, sess)
		}
	}
	return out
}

func LoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
