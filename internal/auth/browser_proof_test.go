package auth

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func filepathWorkspace(m *Manager) string { return filepath.Dir(filepath.Dir(m.path)) }

func browserSession(t *testing.T, m *Manager) string {
	t.Helper()
	state, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := m.VerifyBrowserSessionLogin(totpCode(state.TOTPSecret, time.Now().Unix()/30))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestBrowserLogoutIsSessionBoundWithoutFreshProof(t *testing.T) {
	m, secret := enrolledManager(t)
	owner, other := browserSession(t, m), browserSession(t, m)
	bearer, _, err := m.VerifyLogin("", totpCode(secret, time.Now().Unix()/30))
	if err != nil {
		t.Fatal(err)
	}
	if err = m.updateState(func(s *State, _ bool) error {
		for i := range s.Sessions {
			if s.Sessions[i].TokenHash == hashToken(owner) {
				s.Sessions[i].AuthFactor = "webauthn"
				s.Sessions[i].AuthCredentialID = "removed-key"
				s.Sessions[i].AuthRPID = "localhost"
				s.Sessions[i].AuthenticatedAt = time.Now().Add(-time.Hour)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	proof, err := m.BrowserSessionProof(owner)
	if err != nil || !proof.ReauthRequired {
		t.Fatalf("proof=%+v err=%v", proof, err)
	}
	before, _ := m.load()
	if err = m.LogoutBrowserSession(owner); err != nil {
		t.Fatal(err)
	}
	if m.ValidateToken(owner) || !m.ValidateToken(other) || !m.ValidateToken(bearer) {
		t.Fatal("logout revoked wrong sessions")
	}
	after, _ := m.load()
	if len(after.Sessions) != len(before.Sessions)-1 || after.TOTPSecret != before.TOTPSecret || after.TOTPEnabled != before.TOTPEnabled {
		t.Fatal("logout changed factors")
	}
	for _, s := range before.Sessions {
		if s.TokenHash == hashToken(owner) {
			continue
		}
		found := false
		for _, a := range after.Sessions {
			if a == s {
				found = true
			}
		}
		if !found {
			t.Fatal("other session changed")
		}
	}
	saved, _ := os.ReadFile(m.path)
	for _, token := range []string{owner, bearer, ""} {
		if err = m.LogoutBrowserSession(token); !errors.Is(err, ErrBrowserSessionRequired) {
			t.Fatalf("unexpected authority: %v", err)
		}
		data, _ := os.ReadFile(m.path)
		if !bytes.Equal(data, saved) {
			t.Fatal("refusal changed state")
		}
	}
	if err = m.updateState(func(s *State, _ bool) error {
		for i := range s.Sessions {
			s.Sessions[i].ExpiresAt = time.Now().Add(-time.Second)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = m.LogoutBrowserSession(other); !errors.Is(err, ErrBrowserSessionRequired) {
		t.Fatal("expired owner accepted")
	}
}

func TestBrowserProofIssuanceAndLegacy(t *testing.T) {
	m, _ := enrolledManager(t)
	owner := browserSession(t, m)
	state, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	bearer, _, err := m.VerifyLogin("", totpCode(state.TOTPSecret, time.Now().Unix()/30))
	if err != nil {
		t.Fatal(err)
	}
	if !m.ValidateToken(owner) || !m.ValidateToken(bearer) {
		t.Fatal("ordinary auth regressed")
	}
	proof, err := m.BrowserSessionProof(owner)
	if err != nil || proof.ReauthRequired {
		t.Fatalf("owner proof: %+v %v", proof, err)
	}
	if _, err = m.BrowserSessionProof(bearer); !errors.Is(err, ErrBrowserSessionRequired) {
		t.Fatalf("bearer grant: %v", err)
	}
	// A restarted manager recovers proof. An older writer dropping provenance
	// must only downgrade to ordinary auth; CreatedAt is not inferred as proof.
	reopened := NewManager(filepathWorkspace(m))
	again, err := reopened.BrowserSessionProof(owner)
	if err != nil || again != proof {
		t.Fatalf("restart %+v %v", again, err)
	}
	if err = m.updateState(func(s *State, _ bool) error {
		s.Sessions[0].Purpose = ""
		s.Sessions[0].AuthenticatedAt = time.Time{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !m.ValidateToken(owner) {
		t.Fatal("legacy ordinary session rejected")
	}
	if _, err = m.BrowserSessionProof(owner); !errors.Is(err, ErrBrowserSessionRequired) {
		t.Fatalf("legacy grant: %v", err)
	}
	before, _ := os.ReadFile(m.path)
	if _, err = m.ReauthenticateBrowserTOTP(owner, totpCode(state.TOTPSecret, time.Now().Unix()/30)); !errors.Is(err, ErrBrowserSessionRequired) {
		t.Fatalf("legacy elevated: %v", err)
	}
	after, _ := os.ReadFile(m.path)
	if string(before) != string(after) {
		t.Fatal("failed reauth wrote state")
	}
}

func TestBrowserProofFreshnessBoundaries(t *testing.T) {
	now := time.Now().UTC()
	token := "test-browser-token"
	base := Session{TokenHash: hashToken(token), Purpose: browserOwnerPurpose, AuthFactor: "totp", AuthenticatedAt: now.Add(-time.Minute), CreatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)}
	for _, tc := range []struct {
		name   string
		change func(*State, *Session)
		want   error
	}{
		{"fresh", func(*State, *Session) {}, nil},
		{"just-under-five", func(_ *State, s *Session) { s.AuthenticatedAt = now.Add(-5*time.Minute + time.Nanosecond) }, nil},
		{"exact-five", func(_ *State, s *Session) { s.AuthenticatedAt = now.Add(-5 * time.Minute) }, ErrRecentProofRequired},
		{"older", func(_ *State, s *Session) { s.AuthenticatedAt = now.Add(-6 * time.Minute) }, ErrRecentProofRequired},
		{"future", func(_ *State, s *Session) { s.AuthenticatedAt = now.Add(time.Nanosecond) }, ErrRecentProofRequired},
		{"zero", func(_ *State, s *Session) { s.AuthenticatedAt = time.Time{} }, ErrRecentProofRequired},
		{"before-session", func(_ *State, s *Session) { s.CreatedAt = now }, ErrRecentProofRequired},
		{"unknown-factor", func(_ *State, s *Session) { s.AuthFactor = "webauthn" }, ErrRecentProofRequired},
		{"disabled-factor", func(s *State, _ *Session) { s.TOTPEnabled = false }, ErrRecentProofRequired},
		{"no-secret", func(s *State, _ *Session) { s.TOTPSecret = "" }, ErrRecentProofRequired},
		{"legacy", func(_ *State, s *Session) { s.Purpose = "" }, ErrBrowserSessionRequired},
		{"unknown-purpose", func(_ *State, s *Session) { s.Purpose = "admin" }, ErrBrowserSessionRequired},
		{"expired", func(_ *State, s *Session) { s.ExpiresAt = now }, ErrBrowserSessionRequired},
		{"no-owner", func(s *State, _ *Session) { s.Username = "" }, ErrBrowserSessionRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := State{Username: "admin", TOTPEnabled: true, TOTPSecret: "present", Sessions: []Session{base}}
			tc.change(&state, &state.Sessions[0])
			_, err := requireBrowserOwnerSession(&state, token, now, true)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestBrowserReauthIsSessionScopedAndReadDoesNotRefresh(t *testing.T) {
	m, _ := enrolledManager(t)
	a, b := browserSession(t, m), browserSession(t, m)
	if err := m.updateState(func(s *State, _ bool) error {
		for i := range s.Sessions {
			s.Sessions[i].CreatedAt = time.Now().Add(-time.Hour)
			s.Sessions[i].AuthenticatedAt = time.Now().Add(-6 * time.Minute)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	before, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	bytesBefore, _ := os.ReadFile(m.path)
	for range 3 {
		proof, err := m.BrowserSessionProof(a)
		if err != nil || !proof.ReauthRequired {
			t.Fatalf("stale %+v %v", proof, err)
		}
	}
	bytesAfter, _ := os.ReadFile(m.path)
	if string(bytesBefore) != string(bytesAfter) {
		t.Fatal("status refreshed auth")
	}
	if _, err := m.ReauthenticateBrowserTOTP(a, "bad"); !errors.Is(err, ErrInvalidFactorProof) {
		t.Fatal(err)
	}
	bytesAfter, _ = os.ReadFile(m.path)
	if string(bytesBefore) != string(bytesAfter) {
		t.Fatal("failed proof changed state")
	}
	proof, err := m.ReauthenticateBrowserTOTP(a, totpCode(before.TOTPSecret, time.Now().Unix()/30))
	if err != nil || proof.ReauthRequired {
		t.Fatalf("reauth %+v %v", proof, err)
	}
	other, err := m.BrowserSessionProof(b)
	if err != nil || !other.ReauthRequired {
		t.Fatalf("other browser granted %+v %v", other, err)
	}
	after, _ := m.load()
	if len(after.Sessions) != len(before.Sessions) || !reflect.DeepEqual(after.Sessions[1], before.Sessions[1]) {
		t.Fatal("another session changed")
	}
	expected := before.Sessions[0]
	expected.AuthenticatedAt = after.Sessions[0].AuthenticatedAt
	if !reflect.DeepEqual(expected, after.Sessions[0]) {
		t.Fatal("reauth rotated token or extended expiry")
	}
	if !proof.AuthenticatedAt.After(before.Sessions[0].AuthenticatedAt) {
		t.Fatal("proof not advanced")
	}
}

func TestBrowserReauthRevokeAndFreshGuardUseCurrentTransaction(t *testing.T) {
	m, _ := enrolledManager(t)
	token := browserSession(t, m)
	state, _ := m.load()
	if _, err := m.BrowserSessionProof(token); err != nil {
		t.Fatal(err)
	}
	other := NewManager(filepathWorkspace(m))
	if err := other.RevokeToken(token); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(m.path)
	err := m.updateState(func(s *State, _ bool) error {
		_, err := requireBrowserOwnerSession(s, token, time.Now().UTC(), true)
		if err != nil {
			return err
		}
		s.Username = "must-not-write"
		return nil
	})
	if !errors.Is(err, ErrBrowserSessionRequired) {
		t.Fatalf("stale preflight admitted: %v", err)
	}
	if _, err = m.ReauthenticateBrowserTOTP(token, totpCode(state.TOTPSecret, time.Now().Unix()/30)); !errors.Is(err, ErrBrowserSessionRequired) {
		t.Fatalf("revoked reauth: %v", err)
	}
	after, _ := os.ReadFile(m.path)
	if string(before) != string(after) {
		t.Fatal("revoked session changed state")
	}
	// Concurrent refresh and revoke cannot resurrect the captured session.
	token = browserSession(t, m)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs <- retryState(func() error {
			_, err := m.ReauthenticateBrowserTOTP(token, totpCode(state.TOTPSecret, time.Now().Unix()/30))
			if errors.Is(err, ErrBrowserSessionRequired) {
				return nil
			}
			return err
		})
	}()
	go func() { defer wg.Done(); errs <- retryState(func() error { return other.RevokeToken(token) }) }()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if m.ValidateToken(token) {
		t.Fatal("revoked token resurrected")
	}
}
