package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/go-webauthn/webauthn/protocol"
	wa "github.com/go-webauthn/webauthn/webauthn"
)

const ceremonyTTL = 5 * time.Minute
const maxPasskeys = 32
const maxCeremonies = 16

var (
	ErrPasskeysUnavailable = errors.New("passkeys are not configured for this origin")
	ErrCeremony            = errors.New("passkey ceremony expired, consumed or invalid")
	ErrCredential          = errors.New("passkey verification failed")
	ErrDuplicateCredential = errors.New("passkey is already registered")
	ErrPasskeyName         = errors.New("passkey name must contain 1 to 80 characters without controls")
)

type PasskeyConfig struct {
	RPID    string   `json:"rp_id"`
	Origins []string `json:"origins"`
}

type Passkey struct {
	Name       string        `json:"name"`
	RPID       string        `json:"rp_id"`
	CreatedAt  time.Time     `json:"created_at"`
	LastUsedAt time.Time     `json:"last_used_at,omitempty"`
	Credential wa.Credential `json:"credential"`
}

type Ceremony struct {
	ID         string         `json:"id"`
	Operation  string         `json:"operation"`
	Binding    string         `json:"binding"`
	ConfigHash string         `json:"config_hash"`
	Origin     string         `json:"origin"`
	Name       string         `json:"name,omitempty"`
	ExpiresAt  time.Time      `json:"expires_at"`
	Session    wa.SessionData `json:"session"`
}

type PasskeyInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	RPID       string    `json:"rp_id"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

type PasskeyOptions struct {
	ID      string `json:"ceremony_id"`
	Options any    `json:"options"`
}

type passkeyUser struct {
	state *State
	rp    string
}

func (u passkeyUser) WebAuthnID() []byte          { return u.state.WebAuthnUserID }
func (u passkeyUser) WebAuthnName() string        { return u.state.Username }
func (u passkeyUser) WebAuthnDisplayName() string { return u.state.Username }
func (u passkeyUser) WebAuthnCredentials() []wa.Credential {
	out := []wa.Credential{}
	for _, p := range u.state.Passkeys {
		if p.RPID == u.rp {
			out = append(out, p.Credential)
		}
	}
	return out
}

// Config is immutable for a running manager. Invalid configuration fails closed;
// no Host/forwarded header or request body can configure the relying party.
func NewManagerWithPasskeys(root string, cfg PasskeyConfig) *Manager {
	m := NewManager(root)
	m.passkeyConfig = PasskeyConfig{RPID: cfg.RPID, Origins: append([]string(nil), cfg.Origins...)}
	return m
}
func (m *Manager) passkeyVerifier(origin string) (*wa.WebAuthn, error) {
	cfg := m.passkeyConfig
	if cfg.RPID == "" || cfg.RPID != strings.ToLower(cfg.RPID) || strings.ContainsAny(cfg.RPID, "/:@ *\\") || strings.HasSuffix(cfg.RPID, ".") || len(cfg.Origins) == 0 {
		return nil, ErrPasskeysUnavailable
	}
	if cfg.RPID != "localhost" {
		if _, err := publicsuffix.EffectiveTLDPlusOne(cfg.RPID); err != nil {
			return nil, ErrPasskeysUnavailable
		}
	}
	allowed := false
	for _, raw := range cfg.Origins {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
			return nil, ErrPasskeysUnavailable
		}
		host := u.Hostname()
		local := host == "localhost" || net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback()
		if u.Scheme != "https" && !(u.Scheme == "http" && local) {
			return nil, ErrPasskeysUnavailable
		}
		if host != cfg.RPID && !strings.HasSuffix(host, "."+cfg.RPID) {
			return nil, ErrPasskeysUnavailable
		}
		if raw == origin {
			allowed = true
		}
	}
	if !allowed {
		return nil, ErrPasskeysUnavailable
	}
	w, err := wa.New(&wa.Config{RPID: cfg.RPID, RPDisplayName: "Gi", RPOrigins: []string{origin}, AuthenticatorSelection: protocol.AuthenticatorSelection{UserVerification: protocol.VerificationRequired, ResidentKey: protocol.ResidentKeyRequirementPreferred}, Timeouts: wa.TimeoutsConfig{Login: wa.TimeoutConfig{Enforce: true, Timeout: ceremonyTTL}, Registration: wa.TimeoutConfig{Enforce: true, Timeout: ceremonyTTL}}})
	if err != nil {
		return nil, ErrPasskeysUnavailable
	}
	return w, nil
}
func (m *Manager) PasskeysAvailable(origin string) bool {
	_, err := m.passkeyVerifier(origin)
	return err == nil
}
func (m *Manager) passkeyConfigHash() string {
	data, _ := json.Marshal(m.passkeyConfig)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
func credentialID(c wa.Credential) string { return base64.RawURLEncoding.EncodeToString(c.ID) }
func validPasskeyName(name string) bool {
	if !utf8.ValidString(name) || strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 80 {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func (m *Manager) passkeyOwner(s *State, token string, now time.Time, fresh bool) (*Session, error) {
	sess, err := requireBrowserOwnerSession(s, token, now, fresh)
	if err != nil {
		return nil, err
	}
	if fresh && sess.AuthFactor == "webauthn" && sess.AuthRPID != m.passkeyConfig.RPID {
		return nil, ErrRecentProofRequired
	}
	return sess, nil
}
func (m *Manager) ListPasskeys(token, origin string) ([]PasskeyInfo, error) {
	if _, err := m.passkeyVerifier(origin); err != nil {
		return nil, err
	}
	s, err := m.load()
	if err != nil {
		return nil, err
	}
	if _, err = m.passkeyOwner(&s, token, time.Now().UTC(), false); err != nil {
		return nil, err
	}
	out := []PasskeyInfo{}
	for _, p := range s.Passkeys {
		if p.RPID == m.passkeyConfig.RPID {
			out = append(out, PasskeyInfo{ID: credentialID(p.Credential), Name: p.Name, RPID: p.RPID, CreatedAt: p.CreatedAt, LastUsedAt: p.LastUsedAt})
		}
	}
	return out, nil
}

// Registration/reauth bind to an owner token. Login binds to a separate random
// HttpOnly ceremony cookie; a supplied credential response is never a session.
func (m *Manager) BeginPasskey(operation, binding, origin, name string) (PasskeyOptions, error) {
	w, err := m.passkeyVerifier(origin)
	if err != nil {
		return PasskeyOptions{}, err
	}
	if strings.TrimSpace(binding) == "" {
		return PasskeyOptions{}, ErrCeremony
	}
	if operation != "register" && operation != "login" && operation != "reauth" {
		return PasskeyOptions{}, ErrCeremony
	}
	if operation == "register" && !validPasskeyName(name) {
		return PasskeyOptions{}, ErrPasskeyName
	}
	var result PasskeyOptions
	err = m.updateState(func(s *State, _ bool) error {
		now := time.Now().UTC()
		if operation != "login" {
			if _, err := m.passkeyOwner(s, binding, now, operation == "register"); err != nil {
				return err
			}
		}
		if !passkeyAccepted(s) {
			return ErrPasskeysUnavailable
		}
		if s.Username == "" {
			return ErrCredential
		}
		if operation == "register" && len(s.Passkeys) >= maxPasskeys {
			return errors.New("passkey limit reached")
		}
		if len(s.WebAuthnUserID) == 0 {
			if operation != "register" {
				return ErrCredential
			}
			s.WebAuthnUserID = make([]byte, 32)
			if _, err := rand.Read(s.WebAuthnUserID); err != nil {
				return err
			}
		}
		u := passkeyUser{s, m.passkeyConfig.RPID}
		var options any
		var session *wa.SessionData
		var err error
		if operation == "register" {
			exclusions := []protocol.CredentialDescriptor{}
			for _, c := range u.WebAuthnCredentials() {
				exclusions = append(exclusions, c.Descriptor())
			}
			options, session, err = w.BeginRegistration(u, wa.WithExclusions(exclusions))
		} else {
			options, session, err = w.BeginLogin(u, wa.WithUserVerification(protocol.VerificationRequired))
		}
		if err != nil {
			return ErrCredential
		}
		session.Expires = now.Add(ceremonyTTL)
		id, _, err := newToken()
		if err != nil {
			return err
		}
		active := s.Ceremonies[:0]
		for _, c := range s.Ceremonies {
			if c.ExpiresAt.After(now) {
				active = append(active, c)
			}
		}
		s.Ceremonies = active
		// Unauthenticated login starts cannot exhaust owner-management slots.
		// Each class has eight slots; a new login only evicts the oldest login.
		count, oldest := 0, -1
		for i, c := range s.Ceremonies {
			if (c.Operation == "login") == (operation == "login") {
				count++
				if oldest < 0 {
					oldest = i
				}
			}
		}
		if count >= maxCeremonies/2 {
			if operation != "login" {
				return errors.New("too many pending passkey ceremonies")
			}
			s.Ceremonies = append(s.Ceremonies[:oldest], s.Ceremonies[oldest+1:]...)
		}
		s.Ceremonies = append(s.Ceremonies, Ceremony{ID: id, Operation: operation, Binding: hashToken(binding), ConfigHash: m.passkeyConfigHash(), Origin: origin, Name: name, ExpiresAt: session.Expires, Session: *session})
		result = PasskeyOptions{ID: id, Options: options}
		return nil
	})
	if err != nil {
		return PasskeyOptions{}, err
	}
	return result, nil
}

// Consume first, in its own committed transaction, so invalid proof cannot be
// retried with the same challenge. Binding mismatch must not consume another
// browser's ceremony. A later crash/failure needs a new explicit ceremony.
func (m *Manager) consumeCeremony(id, operation, binding, origin string) (Ceremony, error) {
	var result Ceremony
	err := m.updateState(func(s *State, _ bool) error {
		for i, c := range s.Ceremonies {
			if c.ID != id {
				continue
			}
			if c.Operation != operation || !constantTimeString(c.Binding, hashToken(binding)) {
				return ErrCeremony
			}
			result = c
			s.Ceremonies = append(s.Ceremonies[:i], s.Ceremonies[i+1:]...)
			return nil
		}
		return ErrCeremony
	})
	if err != nil {
		return Ceremony{}, err
	}
	if !result.ExpiresAt.After(time.Now().UTC()) || result.ConfigHash != m.passkeyConfigHash() || result.Origin != origin {
		return Ceremony{}, ErrCeremony
	}
	return result, nil
}
func (m *Manager) FinishPasskeyRegistration(id, token, origin string, response []byte) error {
	w, err := m.passkeyVerifier(origin)
	if err != nil {
		return err
	}
	c, err := m.consumeCeremony(id, "register", token, origin)
	if err != nil {
		return err
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(response)
	if err != nil {
		return ErrCredential
	}
	return m.updateState(func(s *State, _ bool) error {
		now := time.Now().UTC()
		if !passkeyAccepted(s) {
			return ErrPasskeysUnavailable
		}
		if !c.ExpiresAt.After(now) {
			return ErrCeremony
		}
		if _, err := m.passkeyOwner(s, token, now, true); err != nil {
			return err
		}
		if len(s.Passkeys) >= maxPasskeys {
			return errors.New("passkey limit reached")
		}
		credential, err := w.CreateCredential(passkeyUser{s, m.passkeyConfig.RPID}, c.Session, parsed)
		if err != nil {
			return ErrCredential
		}
		for _, p := range s.Passkeys {
			if bytes.Equal(p.Credential.ID, credential.ID) {
				return ErrDuplicateCredential
			}
		}
		s.Passkeys = append(s.Passkeys, Passkey{Name: c.Name, RPID: m.passkeyConfig.RPID, CreatedAt: now, Credential: *credential})
		s.UpdatedAt = now
		return nil
	})
}
func (m *Manager) FinishPasskeyAssertion(id, operation, binding, origin string, response []byte) (string, time.Time, error) {
	if operation != "login" && operation != "reauth" {
		return "", time.Time{}, ErrCeremony
	}
	w, err := m.passkeyVerifier(origin)
	if err != nil {
		return "", time.Time{}, err
	}
	c, err := m.consumeCeremony(id, operation, binding, origin)
	if err != nil {
		return "", time.Time{}, err
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(response)
	if err != nil {
		return "", time.Time{}, ErrCredential
	}
	var token string
	var expires time.Time
	err = m.updateState(func(s *State, _ bool) error {
		now := time.Now().UTC()
		if !passkeyAccepted(s) {
			return ErrPasskeysUnavailable
		}
		if !c.ExpiresAt.After(now) {
			return ErrCeremony
		}
		var owner *Session
		if operation == "reauth" {
			owner, err = m.passkeyOwner(s, binding, now, false)
			if err != nil {
				return err
			}
		}
		credential, err := w.ValidateLogin(passkeyUser{s, m.passkeyConfig.RPID}, c.Session, parsed)
		if err != nil || credential.Authenticator.CloneWarning {
			return ErrCredential
		}
		found := false
		for i := range s.Passkeys {
			p := &s.Passkeys[i]
			if p.RPID == m.passkeyConfig.RPID && bytes.Equal(p.Credential.ID, credential.ID) {
				p.Credential = *credential
				p.LastUsedAt = now
				found = true
				break
			}
		}
		if !found {
			return ErrCredential
		}
		if operation == "login" {
			var hash string
			token, hash, err = newToken()
			if err != nil {
				return err
			}
			expires = now.Add(12 * time.Hour)
			s.Sessions = append(pruneSessions(s.Sessions, now), Session{TokenHash: hash, Purpose: browserOwnerPurpose, CreatedAt: now, ExpiresAt: expires, AuthFactor: "webauthn", AuthCredentialID: credentialID(*credential), AuthRPID: m.passkeyConfig.RPID, AuthenticatedAt: now})
		} else {
			owner.AuthFactor = "webauthn"
			owner.AuthCredentialID = credentialID(*credential)
			owner.AuthRPID = m.passkeyConfig.RPID
			owner.AuthenticatedAt = now
			expires = owner.ExpiresAt
		}
		s.UpdatedAt = now
		return nil
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

// Empty policy preserves existing TOTP behaviour and permits configured
// passkeys. Unknown policies fail closed.
func totpAccepted(s *State) bool {
	return s.LoginPolicy == "" || s.LoginPolicy == "either" || s.LoginPolicy == "totp-only"
}
func passkeyAccepted(s *State) bool {
	return s.LoginPolicy == "" || s.LoginPolicy == "either" || s.LoginPolicy == "passkey-only"
}

var ErrLastFactor = errors.New("add another accepted sign-in method before removing this passkey")
var ErrPasskeyNotFound = errors.New("passkey not found")

// Removal metadata is derived under the same writer lock as the decision. It
// is explanatory only; it never grants authority for a later request.
type PasskeyRemovalResult struct {
	RemainingMethod string `json:"remaining_method"`
}

type PasskeyRemovalError struct {
	Reason string
}

func (e *PasskeyRemovalError) Error() string { return ErrLastFactor.Error() }
func (e *PasskeyRemovalError) Unwrap() error { return ErrLastFactor }

func (m *Manager) RenamePasskey(token, origin, id, name string) error {
	if _, err := m.passkeyVerifier(origin); err != nil {
		return err
	}
	if !validPasskeyName(name) {
		return ErrPasskeyName
	}
	return m.updateState(func(s *State, _ bool) error {
		if !passkeyAccepted(s) {
			return ErrPasskeysUnavailable
		}
		if _, err := m.passkeyOwner(s, token, time.Now().UTC(), true); err != nil {
			return err
		}
		for i := range s.Passkeys {
			p := &s.Passkeys[i]
			if p.RPID == m.passkeyConfig.RPID && credentialID(p.Credential) == id {
				p.Name = name
				s.UpdatedAt = time.Now().UTC()
				return nil
			}
		}
		return ErrPasskeyNotFound
	})
}
func (m *Manager) RemovePasskey(token, origin, id string) error {
	_, err := m.RemovePasskeyWithResult(token, origin, id)
	return err
}

func (m *Manager) RemovePasskeyWithResult(token, origin, id string) (PasskeyRemovalResult, error) {
	if _, err := m.passkeyVerifier(origin); err != nil {
		return PasskeyRemovalResult{}, err
	}
	var result PasskeyRemovalResult
	err := m.updateState(func(s *State, _ bool) error {
		if !passkeyAccepted(s) {
			return ErrPasskeysUnavailable
		}
		if _, err := m.passkeyOwner(s, token, time.Now().UTC(), true); err != nil {
			return err
		}
		index := -1
		other, otherRP := false, false
		for i, p := range s.Passkeys {
			if p.RPID != m.passkeyConfig.RPID {
				otherRP = true
				continue
			}
			if credentialID(p.Credential) == id {
				index = i
			} else {
				other = true
			}
		}
		if index < 0 {
			return ErrPasskeyNotFound
		}
		totp := totpAccepted(s) && s.TOTPEnabled && s.TOTPSecret != ""
		if !other && !totp {
			reason := "no-other-method"
			switch {
			case otherRP:
				reason = "other-rp"
			case s.TOTPEnabled && s.TOTPSecret != "" && !totpAccepted(s):
				reason = "policy-excludes-totp"
			case s.TOTPSecret != "" && !s.TOTPEnabled:
				reason = "totp-not-enabled"
			case s.LoginPolicy == "passkey-only":
				reason = "sessions-not-factors"
			}
			return &PasskeyRemovalError{Reason: reason}
		}
		result.RemainingMethod = "passkey"
		if !other {
			result.RemainingMethod = "totp"
		}
		s.Passkeys = append(s.Passkeys[:index], s.Passkeys[index+1:]...)
		s.UpdatedAt = time.Now().UTC()
		return nil
	})
	if err != nil {
		return PasskeyRemovalResult{}, err
	}
	return result, nil
}
