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

func setupCode(p PendingEnrollment) string { return totpCode(p.Secret, time.Now().Unix()/30) }
func invalidSetupCode(secret string) string {
	for _, code := range []string{"000000", "111111", "222222", "333333"} {
		if !VerifyTOTP(secret, code, time.Now(), 1) {
			return code
		}
	}
	panic("all test codes valid")
}
func noSetupState(t *testing.T, m *Manager) {
	t.Helper()
	if _, err := os.Stat(m.path); !os.IsNotExist(err) {
		t.Fatalf("pending setup persisted state: %v", err)
	}
}
func TestBrowserSetupBindingAndAtomicOwner(t *testing.T) {
	root := t.TempDir()
	m := NewManager(root)
	a, at, err := m.BeginBrowserSetup("")
	if err != nil {
		t.Fatal(err)
	}
	b, bt, err := m.BeginBrowserSetup("")
	if err != nil {
		t.Fatal(err)
	}
	if a.Username != "admin" || a.Secret == "" || a.Secret == b.Secret || a.URL != TOTPURL("gi", "admin", a.Secret) || len(at) != 64 || at == bt {
		t.Fatal("invalid independent pending setups")
	}
	if _, ok := m.browserPending[at]; ok {
		t.Fatal("raw binding retained")
	}
	if _, ok := m.browserPending[hashToken(at)]; !ok {
		t.Fatal("hashed binding absent")
	}
	noSetupState(t, m)
	if token, expires, err := m.FinishBrowserSetup("unknown-binding", setupCode(a)); !errors.Is(err, ErrBrowserSetup) || token != "" || !expires.IsZero() {
		t.Fatal("unknown binding accepted")
	}
	if len(m.browserPending) != 2 {
		t.Fatal("foreign binding consumed real setups")
	}
	noSetupState(t, m)
	token, expires, err := m.FinishBrowserSetup(at, setupCode(a))
	if err != nil {
		t.Fatal(err)
	}
	state, err := m.load()
	if err != nil {
		t.Fatal(err)
	}
	if state.Username != "admin" || !state.TOTPEnabled || state.TOTPSecret != a.Secret || state.LoginPolicy != "either" || len(state.Sessions) != 1 {
		t.Fatalf("incomplete owner commit: username=%q sessions=%d", state.Username, len(state.Sessions))
	}
	session := state.Sessions[0]
	if session.TokenHash != hashToken(token) || session.Purpose != browserOwnerPurpose || session.AuthFactor != "totp" || !session.ExpiresAt.Equal(expires) || expires.Sub(session.CreatedAt) != 12*time.Hour || !session.AuthenticatedAt.Equal(session.CreatedAt) {
		t.Fatal("owner/session not issued together")
	}
	restarted := NewManager(root)
	proof, err := restarted.BrowserSessionProof(token)
	if err != nil || proof.ReauthRequired || !restarted.ValidateToken(token) {
		t.Fatal("committed owner authority unusable after restart")
	}
	before, _ := os.ReadFile(m.path)
	if bytes.Contains(before, []byte(at)) || bytes.Contains(before, []byte(token)) || bytes.Contains(before, []byte(b.Secret)) {
		t.Fatal("binding/raw session/uncommitted secret persisted")
	}
	for _, binding := range []string{at, bt} {
		code := setupCode(a)
		want := ErrBrowserSetup
		if binding == bt {
			code = setupCode(b)
			want = ErrOwnerExists
		}
		value, expiry, err := m.FinishBrowserSetup(binding, code)
		if !errors.Is(err, want) || value != "" || !expiry.IsZero() {
			t.Fatalf("repeat/stale finish: %v", err)
		}
		after, _ := os.ReadFile(m.path)
		if !bytes.Equal(before, after) {
			t.Fatal("stale setup replaced owner")
		}
	}
	if _, _, err = m.BeginBrowserSetup(""); !errors.Is(err, ErrOwnerExists) {
		t.Fatal("new setup after owner")
	}
}

func TestBrowserSetupCodeCannotFinishAnotherBinding(t *testing.T) {
	m := NewManager(t.TempDir())
	a, at, err := m.BeginBrowserSetup("")
	if err != nil {
		t.Fatal(err)
	}
	b, bt, err := m.BeginBrowserSetup("")
	if err != nil {
		t.Fatal(err)
	}
	// Avoid an accidental six-digit overlap between independently generated
	// secrets; the verifier accepts the normal adjacent TOTP windows.
	for i := 0; i < 8 && VerifyTOTP(b.Secret, setupCode(a), time.Now(), 1); i++ {
		b, bt, err = m.BeginBrowserSetup(bt)
		if err != nil {
			t.Fatal(err)
		}
	}
	if VerifyTOTP(b.Secret, setupCode(a), time.Now(), 1) {
		t.Fatal("could not generate distinct test codes")
	}
	if token, expires, err := m.FinishBrowserSetup(bt, setupCode(a)); !errors.Is(err, ErrInvalidFactorProof) || token != "" || !expires.IsZero() {
		t.Fatal("code crossed setup bindings")
	}
	noSetupState(t, m)
	if _, _, err = m.FinishBrowserSetup(bt, setupCode(b)); !errors.Is(err, ErrBrowserSetup) {
		t.Fatal("incorrect-code attempt did not consume its own binding")
	}
	if _, _, err = m.FinishBrowserSetup(at, setupCode(a)); err != nil {
		t.Fatal("foreign attempt consumed original binding", err)
	}
}

func TestBrowserSetupConsumptionCancellationExpiryAndLimit(t *testing.T) {
	for _, kind := range []string{"wrong-code", "cancel", "expired", "future", "restart", "replace"} {
		t.Run(kind, func(t *testing.T) {
			m := NewManager(t.TempDir())
			p, token, err := m.BeginBrowserSetup("")
			if err != nil {
				t.Fatal(err)
			}
			other, otherToken, err := m.BeginBrowserSetup("")
			if err != nil {
				t.Fatal(err)
			}
			target := m
			code := setupCode(p)
			want := ErrBrowserSetup
			switch kind {
			case "wrong-code":
				code = invalidSetupCode(p.Secret)
				want = ErrInvalidFactorProof
			case "cancel":
				m.CancelBrowserSetup(token)
				m.CancelBrowserSetup(token)
			case "expired":
				p.CreatedAt = time.Now().Add(-browserSetupTTL)
				m.browserPending[hashToken(token)] = p
			case "future":
				p.CreatedAt = time.Now().Add(time.Minute)
				m.browserPending[hashToken(token)] = p
			case "restart":
				target = NewManager(filepath.Dir(filepath.Dir(m.path)))
			case "replace":
				_, replacement, err := m.BeginBrowserSetup(token)
				if err != nil || replacement == token {
					t.Fatal("replacement failed")
				}
			}
			value, expiry, err := target.FinishBrowserSetup(token, code)
			if !errors.Is(err, want) || value != "" || !expiry.IsZero() {
				t.Fatalf("bad result: %v", err)
			}
			if _, _, err = target.FinishBrowserSetup(token, setupCode(p)); !errors.Is(err, ErrBrowserSetup) {
				t.Fatal("single use violated")
			}
			noSetupState(t, m)
			// No rejected token can consume a different browser's pending secret.
			if _, _, err = m.FinishBrowserSetup(otherToken, setupCode(other)); err != nil {
				t.Fatal(err)
			}
		})
	}
	m := NewManager(t.TempDir())
	var tokens []string
	for i := 0; i < maxBrowserSetups; i++ {
		_, token, err := m.BeginBrowserSetup("")
		if err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, token)
	}
	if _, _, err := m.BeginBrowserSetup(""); !errors.Is(err, ErrBrowserSetupLimit) {
		t.Fatal("pending limit not enforced")
	}
	_, replacement, err := m.BeginBrowserSetup(tokens[0])
	if err != nil || len(m.browserPending) != maxBrowserSetups {
		t.Fatal("own replacement must work at limit")
	}
	if _, ok := m.browserPending[hashToken(tokens[0])]; ok {
		t.Fatal("old setup survived replacement")
	}
	m.CancelBrowserSetup(replacement)
	if _, _, err = m.BeginBrowserSetup(""); err != nil {
		t.Fatal("cancel did not free capacity")
	}
	for key, p := range m.browserPending {
		p.CreatedAt = time.Now().Add(-browserSetupTTL)
		m.browserPending[key] = p
	}
	if _, _, err = m.BeginBrowserSetup(""); err != nil || len(m.browserPending) != 1 {
		t.Fatal("expiry did not free capacity")
	}
	noSetupState(t, m)
}

func TestBrowserSetupRejectsExistingPartialAndFailedWrites(t *testing.T) {
	for _, kind := range []string{"username", "disabled-secret", "totp-enabled", "passkey", "session", "corrupt", "lock-failure"} {
		t.Run(kind, func(t *testing.T) {
			m := NewManager(t.TempDir())
			p, binding, err := m.BeginBrowserSetup("")
			if err != nil {
				t.Fatal(err)
			}
			if err = os.MkdirAll(filepath.Dir(m.path), 0700); err != nil {
				t.Fatal(err)
			}
			data := []byte("{")
			if kind != "corrupt" {
				state := State{}
				switch kind {
				case "username":
					state.Username = "already"
				case "disabled-secret":
					state.TOTPSecret = "SECRET"
				case "totp-enabled":
					state.TOTPEnabled = true
				case "passkey":
					state.Passkeys = []Passkey{fakePasskey("key", "localhost")}
				case "session":
					state.Sessions = []Session{{TokenHash: "existing"}}
				}
				data, err = encodeState(state)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = os.WriteFile(m.path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "lock-failure" {
				if err = os.Mkdir(filepath.Join(filepath.Dir(m.path), "auth.lock"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			token, expiry, err := m.FinishBrowserSetup(binding, setupCode(p))
			if err == nil || token != "" || !expiry.IsZero() {
				t.Fatal("failed write returned authority")
			}
			after, _ := os.ReadFile(m.path)
			if !bytes.Equal(data, after) {
				t.Fatal("failed setup changed state")
			}
			if _, _, err = m.FinishBrowserSetup(binding, setupCode(p)); !errors.Is(err, ErrBrowserSetup) {
				t.Fatal("failed attempt not consumed")
			}
			if kind != "lock-failure" {
				if _, _, err = m.BeginBrowserSetup(""); err == nil {
					t.Fatal("start bypassed existing/malformed state")
				}
			}
		})
	}
}

func TestBrowserSetupConcurrentManagersAndLegacyEnrollment(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "browser-browser", true: "browser-legacy"}[legacy], func(t *testing.T) {
			root := t.TempDir()
			a, b := NewManager(root), NewManager(root)
			pa, ta, err := a.BeginBrowserSetup("")
			if err != nil {
				t.Fatal(err)
			}
			var pb PendingEnrollment
			var tb string
			if legacy {
				pb, err = b.StartEnrollment("legacy-owner")
			} else {
				pb, tb, err = b.BeginBrowserSetup("")
			}
			if err != nil {
				t.Fatal(err)
			}
			start := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(2)
			type outcome struct {
				token string
				err   error
			}
			results := make([]outcome, 2)
			go func() {
				defer wg.Done()
				<-start
				results[0].token, _, results[0].err = a.FinishBrowserSetup(ta, setupCode(pa))
			}()
			go func() {
				defer wg.Done()
				<-start
				if legacy {
					_, results[1].err = b.VerifyEnrollment("legacy-owner", setupCode(pb))
				} else {
					results[1].token, _, results[1].err = b.FinishBrowserSetup(tb, setupCode(pb))
				}
			}()
			close(start)
			wg.Wait()
			successes := 0
			for _, r := range results {
				if r.err == nil {
					successes++
				} else if r.token != "" {
					t.Fatal("loser issued token")
				}
			}
			if successes != 1 {
				t.Fatalf("expected one commit, errors=%v/%v", results[0].err, results[1].err)
			}
			saved, err := a.load()
			if err != nil {
				t.Fatal(err)
			}
			expected := pa.Secret
			if results[1].err == nil {
				expected = pb.Secret
			}
			if saved.TOTPSecret != expected {
				t.Fatal("winner secret overwritten")
			}
			if legacy && results[1].err == nil {
				if saved.Username != "legacy-owner" || len(saved.Sessions) != 0 {
					t.Fatal("legacy winner changed")
				}
			} else {
				if saved.Username != "admin" || len(saved.Sessions) != 1 {
					t.Fatal("browser winner incomplete")
				}
			}
			before, _ := os.ReadFile(a.path)
			// Retry a rejected browser finish cannot replay a consumed pending token.
			for i, m := range []*Manager{a, b} {
				if i == 1 && legacy {
					continue
				}
				binding := ta
				if i == 1 {
					binding = tb
				}
				if _, _, err = m.FinishBrowserSetup(binding, setupCode(pa)); !errors.Is(err, ErrBrowserSetup) {
					t.Fatal("browser replay accepted")
				}
			}
			after, _ := os.ReadFile(a.path)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("replay changed winner")
			}
		})
	}
}
