package auth

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestAuthLoginPolicyReadCASAndPersistence(t *testing.T) {
	m := passkeyManager(t)
	token := browserSession(t, m)
	before, _ := os.ReadFile(m.path)
	p, err := m.ReadLoginPolicy(token, testPasskeyOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if p.Policy != "either" || p.Revision != "initial" || !p.TOTPConfigured || !p.PasskeyConfigured || p.PasskeyUsable {
		t.Fatalf("%+v", p)
	}
	after, _ := os.ReadFile(m.path)
	if string(before) != string(after) {
		t.Fatal("read mutated state")
	}
	state, _ := m.load()
	sessions := state.Sessions
	for _, target := range []string{"totp-only", "either"} {
		p, err = m.UpdateLoginPolicy(token, testPasskeyOrigin, target, p.Revision)
		if err != nil {
			t.Fatal(err)
		}
		if p.Policy != target || p.Revision == "initial" {
			t.Fatal(p)
		}
		if _, err = m.UpdateLoginPolicy(token, testPasskeyOrigin, "totp-only", "initial"); !errors.Is(err, ErrStateConflict) {
			t.Fatal(err)
		}
	}
	reopened := NewManagerWithPasskeys(filepathWorkspace(m), m.passkeyConfig)
	got, err := reopened.ReadLoginPolicy(token, testPasskeyOrigin)
	if err != nil || got != p {
		t.Fatalf("restart %+v %v", got, err)
	}
	state, _ = m.load()
	if !reflect.DeepEqual(state.Sessions, sessions) {
		t.Fatal("policy changed sessions")
	}
}

func TestAuthLoginPolicyRejectsUnsafeChangesWithoutWrites(t *testing.T) {
	for _, tc := range []struct {
		name, policy, origin string
		change               func(*State)
		want                 error
	}{
		{"invalid", "off", testPasskeyOrigin, func(*State) {}, ErrLoginPolicy},
		{"no-key", "passkey-only", testPasskeyOrigin, func(*State) {}, ErrPolicyLockout},
		{"wrong-rp", "passkey-only", testPasskeyOrigin, func(s *State) { s.Passkeys = []Passkey{fakePasskey("A", "old.localhost")} }, ErrPolicyLockout},
		{"wrong-origin", "passkey-only", "http://localhost:4567", func(s *State) { s.Passkeys = []Passkey{fakePasskey("A", "localhost")} }, ErrPolicyLockout},
		{"expired", "totp-only", testPasskeyOrigin, func(s *State) { s.Sessions[0].ExpiresAt = time.Now().Add(-time.Second) }, ErrBrowserSessionRequired},
		{"revoked", "totp-only", testPasskeyOrigin, func(s *State) { s.Sessions = nil }, ErrBrowserSessionRequired},
		{"legacy", "totp-only", testPasskeyOrigin, func(s *State) { s.Sessions[0].Purpose = "" }, ErrBrowserSessionRequired},
		{"stale-proof", "totp-only", testPasskeyOrigin, func(s *State) {
			s.Sessions[0].CreatedAt = time.Now().Add(-time.Hour)
			s.Sessions[0].AuthenticatedAt = time.Now().Add(-6 * time.Minute)
		}, ErrRecentProofRequired},
		{"no-totp", "totp-only", testPasskeyOrigin, func(s *State) {
			s.TOTPEnabled = false
			s.Passkeys = []Passkey{fakePasskey("A", "localhost")}
			s.Sessions[0].AuthFactor = "webauthn"
			s.Sessions[0].AuthRPID = "localhost"
			s.Sessions[0].AuthCredentialID = credentialID(s.Passkeys[0].Credential)
		}, ErrPolicyLockout},
		{"old-policy-rejects-proof", "either", testPasskeyOrigin, func(s *State) { s.LoginPolicy = "passkey-only"; s.Passkeys = []Passkey{fakePasskey("A", "localhost")} }, ErrRecentProofRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := passkeyManager(t)
			token := browserSession(t, m)
			setPasskeyState(t, m, tc.change)
			before, _ := os.ReadFile(m.path)
			if _, err := m.UpdateLoginPolicy(token, tc.origin, tc.policy, "initial"); !errors.Is(err, tc.want) {
				t.Fatalf("got%v want%v", err, tc.want)
			}
			after, _ := os.ReadFile(m.path)
			if string(before) != string(after) {
				t.Fatal("rejected change wrote state")
			}
		})
	}
}

func TestAuthLoginPolicyAndLastKeyRemovalAreAtomic(t *testing.T) {
	for range 12 {
		m := passkeyManager(t)
		token := browserSession(t, m)
		setPasskeyState(t, m, func(s *State) { s.Passkeys = []Passkey{fakePasskey("A", "localhost")} })
		start := make(chan struct{})
		errs := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			errs <- retryState(func() error {
				_, err := m.UpdateLoginPolicy(token, testPasskeyOrigin, "passkey-only", "initial")
				return err
			})
		}()
		other := NewManagerWithPasskeys(filepathWorkspace(m), m.passkeyConfig)
		go func() {
			defer wg.Done()
			<-start
			errs <- retryState(func() error {
				return other.RemovePasskey(token, testPasskeyOrigin, credentialID(fakePasskey("A", "localhost").Credential))
			})
		}()
		close(start)
		wg.Wait()
		close(errs)
		wins := 0
		for err := range errs {
			if err == nil {
				wins++
			} else if !errors.Is(err, ErrPolicyLockout) && !errors.Is(err, ErrRecentProofRequired) && !errors.Is(err, ErrLastFactor) {
				t.Fatal(err)
			}
		}
		s, _ := m.load()
		if wins != 1 {
			t.Fatalf("wins %d", wins)
		}
		if s.LoginPolicy == "passkey-only" && len(s.Passkeys) != 1 {
			t.Fatal("locked out by race")
		}
	}
}

func TestAuthLoginPolicyReadStillWorksWithoutPasskeyConfiguration(t *testing.T) {
	m := passkeyManager(t)
	token := browserSession(t, m)
	setPasskeyState(t, m, func(s *State) { s.LoginPolicy = "totp-only" })
	plain := NewManager(filepathWorkspace(m))
	p, err := plain.ReadLoginPolicy(token, testPasskeyOrigin)
	if err != nil || p.PasskeyConfigured || p.Policy != "totp-only" {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err = plain.UpdateLoginPolicy(token, testPasskeyOrigin, "either", p.Revision); err != nil {
		t.Fatal(err)
	}
}
