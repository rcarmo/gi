package auth

import (
	"net/http"
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
