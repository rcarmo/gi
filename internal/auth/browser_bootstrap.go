package auth

import (
	"errors"
	"os"
	"time"
)

const browserSetupTTL = 10 * time.Minute
const maxBrowserSetups = 8

var (
	ErrBrowserSetup      = errors.New("browser setup expired, consumed or invalid; start again")
	ErrOwnerExists       = errors.New("owner is already configured")
	ErrBrowserSetupLimit = errors.New("too many pending browser setups; try again later")
)

func emptyOwnerState(s State) bool {
	return s.Username == "" && !s.TOTPEnabled && s.TOTPSecret == "" && len(s.Passkeys) == 0 && len(s.Sessions) == 0
}

// Browser bootstrap tokens are held only as hashes in a bounded, process-local
// map. Restart discards pending setup. The persistent owner is rechecked under
// the auth writer lock when finishing, including against legacy enrollment.
func (m *Manager) BeginBrowserSetup(previous string) (PendingEnrollment, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	state, err := m.load()
	if err != nil && !os.IsNotExist(err) {
		return PendingEnrollment{}, "", err
	}
	if !emptyOwnerState(state) {
		return PendingEnrollment{}, "", ErrOwnerExists
	}
	now := time.Now().UTC()
	for key, pending := range m.browserPending {
		if !now.Before(pending.CreatedAt.Add(browserSetupTTL)) {
			delete(m.browserPending, key)
		}
	}
	if previous != "" {
		delete(m.browserPending, hashToken(previous))
	}
	if len(m.browserPending) >= maxBrowserSetups {
		return PendingEnrollment{}, "", ErrBrowserSetupLimit
	}
	secret, err := GenerateTOTPSecret()
	if err != nil {
		return PendingEnrollment{}, "", err
	}
	token, hash, err := newToken()
	if err != nil {
		return PendingEnrollment{}, "", err
	}
	pending := PendingEnrollment{Username: "admin", Secret: secret, URL: TOTPURL(m.issuer, "admin", secret), CreatedAt: now}
	if m.browserPending == nil {
		m.browserPending = make(map[string]PendingEnrollment)
	}
	m.browserPending[hash] = pending
	return pending, token, nil
}

func (m *Manager) CancelBrowserSetup(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.browserPending, hashToken(token))
}

// Every finish attempt consumes its bound setup before verification. Incorrect
// codes and write failures require explicit new setup; no pending secret is
// persisted and a replay cannot mint another owner session.
func (m *Manager) FinishBrowserSetup(binding, code string) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := hashToken(binding)
	pending, ok := m.browserPending[key]
	delete(m.browserPending, key)
	now := time.Now().UTC()
	if !ok || !now.Before(pending.CreatedAt.Add(browserSetupTTL)) || pending.CreatedAt.After(now) {
		return "", time.Time{}, ErrBrowserSetup
	}
	if !VerifyTOTP(pending.Secret, code, now, 1) {
		return "", time.Time{}, ErrInvalidFactorProof
	}
	var token string
	var expires time.Time
	err := m.updateState(func(s *State, _ bool) error {
		if !emptyOwnerState(*s) {
			return ErrOwnerExists
		}
		if !time.Now().Before(pending.CreatedAt.Add(browserSetupTTL)) {
			return ErrBrowserSetup
		}
		value, hash, err := newToken()
		if err != nil {
			return err
		}
		token = value
		now = time.Now().UTC()
		expires = now.Add(12 * time.Hour)
		s.Username = "admin"
		s.TOTPSecret = pending.Secret
		s.TOTPEnabled = true
		s.LoginPolicy = "either"
		s.CreatedAt = now
		s.UpdatedAt = now
		s.Sessions = []Session{{TokenHash: hash, Purpose: browserOwnerPurpose, AuthFactor: "totp", CreatedAt: now, AuthenticatedAt: now, ExpiresAt: expires}}
		return nil
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}
