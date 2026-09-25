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
	Username       string     `json:"username"`
	TOTPSecret     string     `json:"totp_secret"`
	TOTPEnabled    bool       `json:"totp_enabled"`
	LoginPolicy    string     `json:"login_policy,omitempty"`
	PolicyRevision string     `json:"policy_revision,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Sessions       []Session  `json:"sessions,omitempty"`
	WebAuthnUserID []byte     `json:"webauthn_user_id,omitempty"`
	Passkeys       []Passkey  `json:"passkeys,omitempty"`
	Ceremonies     []Ceremony `json:"webauthn_ceremonies,omitempty"`
	extra          map[string]json.RawMessage
}

type Session struct {
	TokenHash        string    `json:"token_hash"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	Purpose          string    `json:"purpose,omitempty"`
	AuthFactor       string    `json:"auth_factor,omitempty"`
	AuthCredentialID string    `json:"auth_credential_id,omitempty"`
	AuthRPID         string    `json:"auth_rp_id,omitempty"`
	AuthenticatedAt  time.Time `json:"authenticated_at,omitempty"`
}

type PendingEnrollment struct {
	Username  string    `json:"username"`
	Secret    string    `json:"secret"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Manager struct {
	mu             sync.Mutex
	path           string
	issuer         string
	pending        map[string]PendingEnrollment
	browserPending map[string]PendingEnrollment
	passkeyConfig  PasskeyConfig
}

func NewManager(workspaceRoot string) *Manager {
	return &Manager{path: filepath.Join(workspaceRoot, ".gi", "auth.json"), issuer: "gi", pending: make(map[string]PendingEnrollment)}
}

func (m *Manager) Status() (map[string]any, error) { return m.StatusForOrigin("") }

func (m *Manager) StatusForOrigin(origin string) (map[string]any, error) {
	state, err := m.load()
	if os.IsNotExist(err) {
		return map[string]any{"enrolled": false, "enrollment_required": true, "totp_enabled": false, "totp_login_available": false, "passkeys_enabled": false, "passkey_login_available": false}, nil
	}
	if err != nil {
		return nil, err
	}
	configured := m.PasskeysAvailable(origin) && passkeyAccepted(&state)
	usable := false
	for _, p := range state.Passkeys {
		if p.RPID == m.passkeyConfig.RPID {
			usable = true
			break
		}
	}
	return map[string]any{"enrolled": stateEnrolled(state), "enrollment_required": !stateEnrolled(state), "username": state.Username, "totp_enabled": state.TOTPEnabled, "totp_login_available": totpAccepted(&state) && state.TOTPEnabled && state.TOTPSecret != "", "passkeys_enabled": configured, "passkey_login_available": configured && usable}, nil
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
	defer m.mu.Unlock()
	pending, ok := m.pending[username]
	if !ok {
		return State{}, fmt.Errorf("no pending enrollment for user")
	}
	if time.Since(pending.CreatedAt) > 10*time.Minute {
		m.pruneExpiredPendingLocked(time.Now().UTC())
		return State{}, fmt.Errorf("pending enrollment expired")
	}
	if !VerifyTOTP(pending.Secret, code, time.Now().UTC(), 1) {
		return State{}, fmt.Errorf("invalid TOTP code")
	}
	var enrolled State
	if err := m.updateState(func(state *State, _ bool) error {
		if state.TOTPEnabled || state.Username != "" {
			return fmt.Errorf("enrollment is already complete")
		}
		now := time.Now().UTC()
		state.Username, state.TOTPSecret, state.TOTPEnabled = username, pending.Secret, true
		state.CreatedAt, state.UpdatedAt = now, now
		enrolled = *state
		return nil
	}); err != nil {
		return State{}, err
	}
	delete(m.pending, username)
	return enrolled, nil
}

// VerifyLogin issues native/bearer credentials. Copying one into a browser
// cookie does not grant owner credential-management authority.
func (m *Manager) VerifyLogin(username, code string) (string, time.Time, error) {
	return m.verifyLogin(username, code, "")
}

// VerifyBrowserSessionLogin is for the guarded browser login endpoint only.
func (m *Manager) VerifyBrowserSessionLogin(code string) (string, time.Time, error) {
	return m.verifyLogin("", code, browserOwnerPurpose)
}

func (m *Manager) verifyLogin(username, code, purpose string) (string, time.Time, error) {
	var token string
	var expires time.Time
	err := m.updateState(func(state *State, _ bool) error {
		if username != "" && username != state.Username {
			return fmt.Errorf("invalid user")
		}
		if !totpAccepted(state) || !state.TOTPEnabled || state.TOTPSecret == "" {
			return fmt.Errorf("TOTP is not enrolled")
		}
		now := time.Now().UTC()
		if !VerifyTOTP(state.TOTPSecret, code, now, 1) {
			return fmt.Errorf("invalid TOTP code")
		}
		var hash string
		var err error
		token, hash, err = newToken()
		if err != nil {
			return err
		}
		expires = now.Add(12 * time.Hour)
		session := Session{TokenHash: hash, CreatedAt: now, ExpiresAt: expires, Purpose: purpose}
		if purpose == browserOwnerPurpose {
			session.AuthFactor, session.AuthenticatedAt = "totp", now
		}
		state.Sessions = append(pruneSessions(state.Sessions, now), session)
		state.UpdatedAt = now
		return nil
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func (m *Manager) ValidateBearerRequest(r *http.Request) bool {
	ok, _ := m.ValidateBearerRequestWithError(r)
	return ok
}

func (m *Manager) ValidateBearerRequestWithError(r *http.Request) (bool, error) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	if token == "" {
		token = r.URL.Query().Get("auth_token")
	}
	return m.ValidateTokenWithError(token)
}

func (m *Manager) ValidateToken(token string) bool {
	ok, _ := m.ValidateTokenWithError(token)
	return ok
}

func (m *Manager) ValidateTokenWithError(token string) (bool, error) {
	if strings.TrimSpace(token) == "" {
		return false, nil
	}
	state, err := m.load()
	if err != nil {
		return false, err
	}
	hash := hashToken(token)
	now := time.Now().UTC()
	for _, sess := range state.Sessions {
		if sess.ExpiresAt.After(now) && constantTimeString(sess.TokenHash, hash) {
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) Enrolled() (bool, error) {
	state, err := m.load()
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return stateEnrolled(state), nil
}

// Stored passkeys keep authentication enabled even if their RP configuration is
// later removed. Configuration changes must never reopen an enrolled instance.
func stateEnrolled(state State) bool {
	return state.Username != "" && (state.TOTPEnabled || len(state.Passkeys) > 0)
}

// RevokeToken removes only the captured opaque token, never another device's
// session. An already-absent token is idempotent; HTTP callers must authenticate
// and enforce origin/transport policy before exposing this operation.
func (m *Manager) RevokeToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token required")
	}
	return m.updateState(func(state *State, exists bool) error {
		if !exists {
			return os.ErrNotExist
		}
		hash := hashToken(token)
		sessions := state.Sessions[:0]
		for _, session := range state.Sessions {
			if !constantTimeString(session.TokenHash, hash) {
				sessions = append(sessions, session)
			}
		}
		state.Sessions = sessions
		state.UpdatedAt = time.Now().UTC()
		return nil
	})
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
