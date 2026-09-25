package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	wa "github.com/go-webauthn/webauthn/webauthn"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func passkeyManager(t *testing.T) *Manager {
	t.Helper()
	old, _ := enrolledManager(t)
	return NewManagerWithPasskeys(filepathWorkspace(old), PasskeyConfig{RPID: "localhost", Origins: []string{"http://localhost:1234"}})
}

const testPasskeyOrigin = "http://localhost:1234"

func fakePasskey(id, rp string) Passkey {
	return Passkey{Name: id, RPID: rp, CreatedAt: time.Now().UTC(), Credential: wa.Credential{ID: []byte(id), PublicKey: []byte("native-metadata-fixture-not-crypto")}}
}
func setPasskeyState(t *testing.T, m *Manager, change func(*State)) {
	t.Helper()
	if err := m.updateState(func(s *State, _ bool) error { change(s); return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestAuthPasskeyConfigFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		cfg    PasskeyConfig
		origin string
		ok     bool
	}{
		{PasskeyConfig{}, testPasskeyOrigin, false},
		{PasskeyConfig{"localhost", []string{testPasskeyOrigin}}, testPasskeyOrigin, true},
		{PasskeyConfig{"example.com", []string{"https://gi.example.com"}}, "https://gi.example.com", true},
		{PasskeyConfig{"example.com", []string{"http://example.com"}}, "http://example.com", false},
		{PasskeyConfig{"com", []string{"https://example.com"}}, "https://example.com", false},
		{PasskeyConfig{"co.uk", []string{"https://gi.co.uk"}}, "https://gi.co.uk", false},
		{PasskeyConfig{"localhost", []string{"http://localhost/path"}}, "http://localhost/path", false},
		{PasskeyConfig{"localhost", []string{"http://user@localhost"}}, "http://user@localhost", false},
		{PasskeyConfig{"localhost", []string{"http://localhost"}}, testPasskeyOrigin, false},
		{PasskeyConfig{"example.com", []string{"https://evil.example"}}, "https://evil.example", false},
		{PasskeyConfig{"127.0.0.1", []string{"http://127.0.0.1"}}, "http://127.0.0.1", false},
	} {
		m := NewManagerWithPasskeys(t.TempDir(), tc.cfg)
		if got := m.PasskeysAvailable(tc.origin); got != tc.ok {
			t.Fatalf("%+v: %v", tc, got)
		}
	}
}
func TestAuthPasskeyCeremonyOwnershipConsumptionAndRestart(t *testing.T) {
	m := passkeyManager(t)
	a, b := browserSession(t, m), browserSession(t, m)
	begin, err := m.BeginPasskey("register", a, testPasskeyOrigin, "Laptop")
	if err != nil {
		t.Fatal(err)
	}
	state, _ := m.load()
	if len(state.WebAuthnUserID) != 32 || len(state.Ceremonies) != 1 {
		t.Fatal("missing ceremony")
	}
	data, _ := os.ReadFile(m.path)
	if strings.Contains(string(data), a) {
		t.Fatal("raw login token persisted")
	}
	if _, err = m.consumeCeremony(begin.ID, "register", b, testPasskeyOrigin); !errors.Is(err, ErrCeremony) {
		t.Fatal(err)
	}
	reopened := NewManagerWithPasskeys(filepathWorkspace(m), m.passkeyConfig)
	if _, err = reopened.consumeCeremony(begin.ID, "register", a, testPasskeyOrigin); err != nil {
		t.Fatal(err)
	}
	if _, err = m.consumeCeremony(begin.ID, "register", a, testPasskeyOrigin); !errors.Is(err, ErrCeremony) {
		t.Fatalf("replay %v", err)
	}
	begin, err = m.BeginPasskey("register", a, testPasskeyOrigin, "Invalid")
	if err != nil {
		t.Fatal(err)
	}
	if err = m.FinishPasskeyRegistration(begin.ID, a, testPasskeyOrigin, []byte(`{}`)); !errors.Is(err, ErrCredential) {
		t.Fatal(err)
	}
	if err = m.FinishPasskeyRegistration(begin.ID, a, testPasskeyOrigin, []byte(`{}`)); !errors.Is(err, ErrCeremony) {
		t.Fatalf("invalid proof reusable %v", err)
	}
	for _, mode := range []string{"expiry", "config", "origin"} {
		begin, err = m.BeginPasskey("register", a, testPasskeyOrigin, "stale")
		if err != nil {
			t.Fatal(err)
		}
		origin := testPasskeyOrigin
		target := m
		if mode == "expiry" {
			setPasskeyState(t, m, func(s *State) { s.Ceremonies[0].ExpiresAt = time.Now().Add(-time.Second) })
		}
		if mode == "config" {
			target = NewManagerWithPasskeys(filepathWorkspace(m), PasskeyConfig{RPID: "localhost", Origins: []string{testPasskeyOrigin, "http://localhost:5678"}})
		}
		if mode == "origin" {
			origin = "http://localhost:5678"
		}
		if _, err = target.consumeCeremony(begin.ID, "register", a, origin); !errors.Is(err, ErrCeremony) {
			t.Fatalf("%s %v", mode, err)
		}
	}
}
func TestAuthPasskeyRemovalExplainsCommittedFactors(t *testing.T) {
	for _, tc := range []struct {
		name, policy, reason, remaining string
		other, oldRP, enabled, secret   bool
	}{
		{"other-key", "passkey-only", "", "passkey", true, false, false, false},
		{"excluded-totp", "passkey-only", "policy-excludes-totp", "", false, false, true, true},
		{"session-only", "passkey-only", "sessions-not-factors", "", false, false, false, false},
		{"old-rp", "passkey-only", "other-rp", "", false, true, false, false},
		{"accepted-totp", "either", "", "totp", false, false, true, true},
		{"no-factor", "either", "no-other-method", "", false, false, false, false},
		{"disabled-secret", "either", "totp-not-enabled", "", false, false, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := passkeyManager(t)
			token := browserSession(t, m)
			id := []byte("selected-key")
			if err := m.updateState(func(s *State, _ bool) error {
				s.LoginPolicy = tc.policy
				s.TOTPEnabled = tc.enabled
				if !tc.secret {
					s.TOTPSecret = ""
				}
				s.Passkeys = []Passkey{{Name: "Laptop", RPID: "localhost", Credential: wa.Credential{ID: id, PublicKey: []byte("public-key")}}}
				if tc.other || tc.oldRP {
					rp := "localhost"
					if tc.oldRP {
						rp = "old.localhost"
					}
					s.Passkeys = append(s.Passkeys, Passkey{Name: "Backup", RPID: rp, Credential: wa.Credential{ID: []byte("backup-key")}})
				}
				for i := range s.Sessions {
					s.Sessions[i].AuthFactor = "webauthn"
					s.Sessions[i].AuthCredentialID = base64.RawURLEncoding.EncodeToString(id)
					s.Sessions[i].AuthRPID = "localhost"
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(m.path)
			if err != nil {
				t.Fatal(err)
			}
			state, _ := m.load()
			result, err := m.RemovePasskeyWithResult(token, testPasskeyOrigin, base64.RawURLEncoding.EncodeToString(id))
			if tc.reason != "" {
				var refusal *PasskeyRemovalError
				if !errors.Is(err, ErrLastFactor) || !errors.As(err, &refusal) || refusal.Reason != tc.reason || err.Error() != ErrLastFactor.Error() || result.RemainingMethod != "" {
					t.Fatalf("result=%+v err=%v refusal=%+v", result, err, refusal)
				}
				after, _ := os.ReadFile(m.path)
				if !bytes.Equal(before, after) {
					t.Fatal("refusal changed state")
				}
				if err = m.RemovePasskey(token, testPasskeyOrigin, base64.RawURLEncoding.EncodeToString(id)); !errors.Is(err, ErrLastFactor) {
					t.Fatal("legacy wrapper lost sentinel")
				}
			} else {
				if err != nil || result.RemainingMethod != tc.remaining {
					t.Fatalf("result=%+v err=%v", result, err)
				}
				after, _ := m.load()
				if len(after.Passkeys) != len(state.Passkeys)-1 || !reflect.DeepEqual(after.Sessions, state.Sessions) {
					t.Fatal("incorrect removal/session change")
				}
			}
		})
	}
	m := passkeyManager(t)
	token := browserSession(t, m)
	result, err := m.RemovePasskeyWithResult(token, testPasskeyOrigin, "not-a-key")
	if !errors.Is(err, ErrPasskeyNotFound) || result.RemainingMethod != "" {
		t.Fatal("uncommitted success detail")
	}
}

func TestAuthPasskeyMetadataAndAtomicLastFactor(t *testing.T) {
	m := passkeyManager(t)
	token := browserSession(t, m)
	setPasskeyState(t, m, func(s *State) { s.Passkeys = []Passkey{fakePasskey("A", "localhost"), fakePasskey("B", "localhost")} })
	state, _ := m.load()
	before := state.Passkeys[0]
	for _, name := range []string{"", "  ", strings.Repeat("Ω", 81), "bad\nname"} {
		if err := m.RenamePasskey(token, testPasskeyOrigin, credentialID(before.Credential), name); !errors.Is(err, ErrPasskeyName) {
			t.Fatalf("name: %v", err)
		}
	}
	name := "<b>" + strings.Repeat("Ω", 70) + "</b>"
	if err := m.RenamePasskey(token, testPasskeyOrigin, credentialID(before.Credential), name); err != nil {
		t.Fatal(err)
	}
	state, _ = m.load()
	before.Name = name
	if !reflect.DeepEqual(state.Passkeys[0], before) {
		t.Fatal("rename changed material")
	}
	// Both recent browser sessions authenticate through distinct passkeys.
	second := browserSession(t, m)
	setPasskeyState(t, m, func(s *State) {
		s.LoginPolicy = "passkey-only"
		for i := range s.Sessions {
			s.Sessions[i].AuthFactor = "webauthn"
			s.Sessions[i].AuthRPID = "localhost"
			s.Sessions[i].AuthCredentialID = credentialID(s.Passkeys[i].Credential)
		}
	})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i, tok := range []string{token, second} {
		wg.Add(1)
		go func(i int, tok string) {
			defer wg.Done()
			errs <- retryState(func() error {
				return m.RemovePasskey(tok, testPasskeyOrigin, credentialID(state.Passkeys[i].Credential))
			})
		}(i, tok)
	}
	wg.Wait()
	close(errs)
	wins := 0
	for err := range errs {
		if err == nil {
			wins++
		} else if !errors.Is(err, ErrLastFactor) {
			t.Fatal(err)
		}
	}
	state, _ = m.load()
	if wins != 1 || len(state.Passkeys) != 1 {
		t.Fatalf("wins%d keys%d", wins, len(state.Passkeys))
	}
	if !m.ValidateToken(token) || !m.ValidateToken(second) {
		t.Fatal("removal revoked existing session")
	}
}
func TestAuthPasskeyLastFactorUsesPolicyRPAndSessionAtWriteTime(t *testing.T) {
	for _, tc := range []struct {
		policy  string
		totp    bool
		otherRP string
		allowed bool
	}{
		{"passkey-only", true, "", false}, {"passkey-only", false, "old.localhost", false}, {"passkey-only", false, "localhost", true}, {"either", true, "", true}, {"either", false, "", false},
	} {
		m := passkeyManager(t)
		token := browserSession(t, m)
		setPasskeyState(t, m, func(s *State) {
			s.LoginPolicy = tc.policy
			s.TOTPEnabled = tc.totp
			s.Passkeys = []Passkey{fakePasskey("A", "localhost")}
			if tc.otherRP != "" {
				s.Passkeys = append(s.Passkeys, fakePasskey("B", tc.otherRP))
			}
			s.Sessions[0].AuthFactor = "webauthn"
			s.Sessions[0].AuthRPID = "localhost"
			s.Sessions[0].AuthCredentialID = base64.RawURLEncoding.EncodeToString([]byte("A"))
		})
		err := m.RemovePasskey(token, testPasskeyOrigin, base64.RawURLEncoding.EncodeToString([]byte("A")))
		if tc.allowed && err != nil || !tc.allowed && !errors.Is(err, ErrLastFactor) {
			t.Fatalf("%+v: %v", tc, err)
		}
		// A configuration loss must not reopen ordinary authenticated routes.
		state, _ := m.load()
		if len(state.Passkeys) > 0 && !stateEnrolled(state) {
			t.Fatal("passkey account became unenrolled")
		}
	}
	m := passkeyManager(t)
	token := browserSession(t, m)
	setPasskeyState(t, m, func(s *State) { s.Passkeys = []Passkey{fakePasskey("A", "localhost")}; s.LoginPolicy = "totp-only" })
	if _, err := m.BeginPasskey("register", token, testPasskeyOrigin, "B"); !errors.Is(err, ErrPasskeysUnavailable) {
		t.Fatal(err)
	}
	setPasskeyState(t, m, func(s *State) {
		s.LoginPolicy = "either"
		s.Sessions[0].CreatedAt = time.Now().Add(-time.Hour)
		s.Sessions[0].AuthenticatedAt = time.Now().Add(-6 * time.Minute)
	})
	if _, err := m.BeginPasskey("register", token, testPasskeyOrigin, "B"); !errors.Is(err, ErrRecentProofRequired) {
		t.Fatal(err)
	}
	if err := m.RemovePasskey(token, testPasskeyOrigin, base64.RawURLEncoding.EncodeToString([]byte("A"))); !errors.Is(err, ErrRecentProofRequired) {
		t.Fatal(err)
	}
}

func TestAuthPasskeyCeremonyBoundsDoNotBlockOwner(t *testing.T) {
	m := passkeyManager(t)
	token := browserSession(t, m)
	setPasskeyState(t, m, func(s *State) {
		s.Passkeys = []Passkey{fakePasskey("A", "localhost")}
		s.WebAuthnUserID = make([]byte, 32)
	})
	owner, err := m.BeginPasskey("register", token, testPasskeyOrigin, "Owner")
	if err != nil {
		t.Fatal(err)
	}
	for range 24 {
		if _, err := m.BeginPasskey("login", "private-cookie-binding", testPasskeyOrigin, ""); err != nil {
			t.Fatal(err)
		}
	}
	state, _ := m.load()
	if len(state.Ceremonies) != maxCeremonies/2+1 {
		t.Fatalf("unbounded %d", len(state.Ceremonies))
	}
	if _, err := m.consumeCeremony(owner.ID, "register", token, testPasskeyOrigin); err != nil {
		t.Fatal("login traffic evicted owner ceremony", err)
	}
}

func TestAuthPasskeyPolicyStatusDoesNotExposeCredentialInventory(t *testing.T) {
	m := passkeyManager(t)
	setPasskeyState(t, m, func(s *State) {
		s.Passkeys = []Passkey{fakePasskey("private-id", "localhost")}
		s.LoginPolicy = "passkey-only"
	})
	p, err := m.StatusForOrigin(testPasskeyOrigin)
	if err != nil {
		t.Fatal(err)
	}
	if p["totp_login_available"] != false || p["passkeys_enabled"] != true || p["passkey_login_available"] != true {
		t.Fatalf("policy %+v", p)
	}
	raw, _ := json.Marshal(p)
	if strings.Contains(string(raw), "private-id") || strings.Contains(string(raw), "public_key") {
		t.Fatal("inventory exposed")
	}
	p, err = m.StatusForOrigin("https://wrong.example")
	if err != nil || p["passkeys_enabled"] != false || p["enrolled"] != true {
		t.Fatalf("wrong origin %+v %v", p, err)
	}
}
