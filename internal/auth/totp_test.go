package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTOTPVerifyCurrentCode(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("secret: %v", err)
	}
	now := time.Now().UTC()
	code := totpCode(secret, now.Unix()/totpPeriod)
	if !VerifyTOTP(secret, code, now, 0) {
		t.Fatal("expected generated code to verify")
	}
}

func TestVerifyEnrollmentRemovesExpiredPendingEntry(t *testing.T) {
	m := NewManager(t.TempDir())
	m.pending["rui"] = PendingEnrollment{
		Username:  "rui",
		Secret:    "expired-secret",
		CreatedAt: time.Now().UTC().Add(-11 * time.Minute),
	}
	if _, err := m.VerifyEnrollment("rui", "000000"); err == nil || err.Error() != "pending enrollment expired" {
		t.Fatalf("expected expired enrollment error, got %v", err)
	}
	if _, ok := m.pending["rui"]; ok {
		t.Fatalf("expected expired pending enrollment to be removed, got %#v", m.pending)
	}
}

func TestStartEnrollmentPrunesExpiredPendingEntries(t *testing.T) {
	m := NewManager(t.TempDir())
	m.pending["stale"] = PendingEnrollment{
		Username:  "stale",
		Secret:    "expired-secret",
		CreatedAt: time.Now().UTC().Add(-11 * time.Minute),
	}
	pending, err := m.StartEnrollment("rui")
	if err != nil {
		t.Fatalf("start enrollment: %v", err)
	}
	if pending.Username != "rui" {
		t.Fatalf("unexpected pending enrollment: %#v", pending)
	}
	if _, ok := m.pending["stale"]; ok {
		t.Fatalf("expected stale pending enrollment to be pruned, got %#v", m.pending)
	}
	if _, ok := m.pending["rui"]; !ok {
		t.Fatalf("expected fresh pending enrollment to remain, got %#v", m.pending)
	}
}

func TestStartEnrollmentReturnsLoadErrorForCorruptState(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gi", "auth.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir auth dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt auth file: %v", err)
	}
	m := NewManager(root)
	if _, err := m.StartEnrollment("rui"); err == nil {
		t.Fatal("expected start enrollment to return corrupt state load error")
	}
}

func TestManagerSaveLoadRoundTrip(t *testing.T) {
	m := NewManager(t.TempDir())
	now := time.Now().UTC().Truncate(time.Second)
	state := State{
		Username:    "rui",
		TOTPSecret:  "secret",
		TOTPEnabled: true,
		CreatedAt:   now,
		UpdatedAt:   now,
		Sessions: []Session{{
			TokenHash: "hash",
			CreatedAt: now,
			ExpiresAt: now.Add(time.Hour),
		}},
	}
	if err := m.save(state); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := m.load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Username != state.Username || loaded.TOTPSecret != state.TOTPSecret || !loaded.TOTPEnabled || len(loaded.Sessions) != 1 || loaded.Sessions[0].TokenHash != "hash" {
		t.Fatalf("unexpected loaded state: %#v", loaded)
	}
}

func TestManagerEnrollmentAndLogin(t *testing.T) {
	m := NewManager(t.TempDir())
	status, err := m.Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status["enrollment_required"] != true {
		t.Fatalf("expected enrollment required: %#v", status)
	}
	pending, err := m.StartEnrollment("rui")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	code := totpCode(pending.Secret, time.Now().UTC().Unix()/totpPeriod)
	if _, err := m.VerifyEnrollment("rui", code); err != nil {
		t.Fatalf("verify enrollment: %v", err)
	}
	loginCode := totpCode(pending.Secret, time.Now().UTC().Unix()/totpPeriod)
	token, expires, err := m.VerifyLogin("rui", loginCode)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" || !expires.After(time.Now()) {
		t.Fatalf("bad token/expires: %q %v", token, expires)
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if !m.ValidateBearerRequest(req) {
		t.Fatal("expected bearer token to validate")
	}
}
