package auth

import (
	"errors"
	"os"
	"strings"
	"time"
)

const browserOwnerPurpose = "browser-owner"
const recentProofWindow = 5 * time.Minute

var (
	ErrBrowserSessionRequired = errors.New("browser owner sign-in required")
	ErrRecentProofRequired    = errors.New("recent authentication required")
	ErrInvalidFactorProof     = errors.New("invalid authentication code")
)

// BrowserProof is advisory UI state. Credential mutations must recheck authority
// and proof inside updateState, not trust a preceding status response.
type BrowserProof struct {
	ReauthRequired  bool      `json:"reauth_required"`
	AuthenticatedAt time.Time `json:"authenticated_at"`
	FreshUntil      time.Time `json:"fresh_until"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// requireBrowserOwnerSession is usable only against the snapshot being read or
// mutated. Missing purpose (legacy records/older writers) never grants authority.
func requireBrowserOwnerSession(state *State, token string, now time.Time, fresh bool) (*Session, error) {
	if strings.TrimSpace(token) == "" || state.Username == "" {
		return nil, ErrBrowserSessionRequired
	}
	hash := hashToken(token)
	for i := range state.Sessions {
		session := &state.Sessions[i]
		if !constantTimeString(session.TokenHash, hash) {
			continue
		}
		if session.Purpose != browserOwnerPurpose || !session.ExpiresAt.After(now) {
			return nil, ErrBrowserSessionRequired
		}
		if fresh && !hasRecentBrowserProof(state, session, now) {
			return nil, ErrRecentProofRequired
		}
		return session, nil
	}
	return nil, ErrBrowserSessionRequired
}

func hasRecentBrowserProof(state *State, session *Session, now time.Time) bool {
	validFactor := session.AuthFactor == "totp" && totpAccepted(state) && state.TOTPEnabled && state.TOTPSecret != ""
	if session.AuthFactor == "webauthn" && passkeyAccepted(state) {
		for _, p := range state.Passkeys {
			if p.RPID == session.AuthRPID && credentialID(p.Credential) == session.AuthCredentialID {
				validFactor = true
				break
			}
		}
	}
	return validFactor && !session.AuthenticatedAt.IsZero() && !session.AuthenticatedAt.Before(session.CreatedAt) &&
		!session.AuthenticatedAt.After(now) && now.Before(session.AuthenticatedAt.Add(recentProofWindow))
}

func browserProof(state *State, session *Session, now time.Time) BrowserProof {
	until := session.AuthenticatedAt.Add(recentProofWindow)
	if until.After(session.ExpiresAt) {
		until = session.ExpiresAt
	}
	return BrowserProof{ReauthRequired: !hasRecentBrowserProof(state, session, now), AuthenticatedAt: session.AuthenticatedAt, FreshUntil: until, ExpiresAt: session.ExpiresAt}
}

func (m *Manager) BrowserSessionProof(token string) (BrowserProof, error) {
	state, err := m.load()
	if os.IsNotExist(err) {
		return BrowserProof{}, ErrBrowserSessionRequired
	}
	if err != nil {
		return BrowserProof{}, err
	}
	now := time.Now().UTC()
	session, err := requireBrowserOwnerSession(&state, token, now, false)
	if err != nil {
		return BrowserProof{}, err
	}
	proof := browserProof(&state, session, now)
	if session.AuthFactor == "webauthn" && (session.AuthRPID != m.passkeyConfig.RPID || len(m.passkeyConfig.Origins) == 0) {
		proof.ReauthRequired = true
	}
	return proof, nil
}

// ReauthenticateBrowserTOTP refreshes only this existing browser session. It
// never mints/rotates a token, changes expiry or authorises another browser.
func (m *Manager) ReauthenticateBrowserTOTP(token, code string) (BrowserProof, error) {
	var result BrowserProof
	err := m.updateState(func(state *State, _ bool) error {
		now := time.Now().UTC()
		session, err := requireBrowserOwnerSession(state, token, now, false)
		if err != nil {
			return err
		}
		if !totpAccepted(state) || !state.TOTPEnabled || state.TOTPSecret == "" || !VerifyTOTP(state.TOTPSecret, code, now, 1) {
			return ErrInvalidFactorProof
		}
		session.AuthFactor, session.AuthenticatedAt = "totp", now
		state.UpdatedAt = now
		result = browserProof(state, session, now)
		return nil
	})
	if err != nil {
		return BrowserProof{}, err
	}
	return result, nil
}
