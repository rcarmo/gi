package auth

import (
	"errors"
	"time"
)

var (
	ErrLoginPolicy   = errors.New("invalid login policy")
	ErrPolicyLockout = errors.New("selected policy would leave no usable sign-in method")
)

type LoginPolicy struct {
	Policy            string `json:"policy"`
	Revision          string `json:"revision"`
	TOTPConfigured    bool   `json:"totp_configured"`
	PasskeyConfigured bool   `json:"passkey_configured"`
	PasskeyUsable     bool   `json:"passkey_usable"`
}

func policyRevision(s *State) string {
	if s.PolicyRevision == "" {
		return "initial"
	}
	return s.PolicyRevision
}
func (m *Manager) loginPolicy(s *State, origin string) LoginPolicy {
	policy := s.LoginPolicy
	if policy == "" {
		policy = "either"
	}
	configured := m.PasskeysAvailable(origin)
	usable := false
	for _, p := range s.Passkeys {
		if p.RPID == m.passkeyConfig.RPID {
			usable = configured
			break
		}
	}
	return LoginPolicy{Policy: policy, Revision: policyRevision(s), TOTPConfigured: s.TOTPEnabled && s.TOTPSecret != "", PasskeyConfigured: configured, PasskeyUsable: usable}
}
func (m *Manager) ReadLoginPolicy(token, origin string) (LoginPolicy, error) {
	s, err := m.load()
	if err != nil {
		return LoginPolicy{}, err
	}
	if _, err = m.passkeyOwner(&s, token, time.Now().UTC(), false); err != nil {
		return LoginPolicy{}, err
	}
	return m.loginPolicy(&s, origin), nil
}

// Policy and key mutations share auth.lock. Evaluate the remaining accepted
// factors against the current snapshot, never a prior UI list or active session.
func (m *Manager) UpdateLoginPolicy(token, origin, policy, revision string) (LoginPolicy, error) {
	if policy != "either" && policy != "totp-only" && policy != "passkey-only" {
		return LoginPolicy{}, ErrLoginPolicy
	}
	var result LoginPolicy
	err := m.updateState(func(s *State, _ bool) error {
		now := time.Now().UTC()
		owner, err := m.passkeyOwner(s, token, now, true)
		if err != nil {
			return err
		}
		if owner.AuthFactor == "webauthn" && !m.PasskeysAvailable(origin) {
			return ErrRecentProofRequired
		}
		if revision == "" || revision != policyRevision(s) {
			return ErrStateConflict
		}
		current := m.loginPolicy(s, origin)
		if !((policy == "either" || policy == "totp-only") && current.TOTPConfigured || (policy == "either" || policy == "passkey-only") && current.PasskeyUsable) {
			return ErrPolicyLockout
		}
		next, _, err := newToken()
		if err != nil {
			return err
		}
		s.LoginPolicy, s.PolicyRevision, s.UpdatedAt = policy, next, now
		result = m.loginPolicy(s, origin)
		return nil
	})
	if err != nil {
		return LoginPolicy{}, err
	}
	return result, nil
}
