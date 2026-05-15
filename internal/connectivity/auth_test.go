package connectivity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

func TestAuthorizeHTTPRequestAllowsLoopbackWithoutAuth(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	if err := AuthorizeHTTPRequest(RouteSpec{}, req, nil); err != nil {
		t.Fatalf("loopback should be allowed without auth: %v", err)
	}
}

func TestAuthorizeHTTPRequestRequiresAuthForExternal(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	if err := AuthorizeHTTPRequest(RouteSpec{}, req, nil); err == nil {
		t.Fatal("expected external request without auth to fail")
	}
}

func TestAuthorizeHTTPRequestAllowsConfiguredExternalBypassEvenWhenAuthExists(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	spec := RouteSpec{
		Auth:    map[string]any{"type": "bearer", "token": "secret"},
		Options: map[string]any{"allow_unauthenticated_external": true},
	}
	if err := AuthorizeHTTPRequest(spec, req, nil); err != nil {
		t.Fatalf("expected allow_unauthenticated_external to bypass auth, got %v", err)
	}
}

func TestAuthorizeHTTPRequestBearer(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("Authorization", "Bearer secret")
	spec := RouteSpec{Auth: map[string]any{"type": "bearer", "token": "secret"}}
	if err := AuthorizeHTTPRequest(spec, req, nil); err != nil {
		t.Fatalf("bearer should pass: %v", err)
	}
}

func TestAuthorizeHTTPRequestBasic(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.SetBasicAuth("rui", "secret")
	spec := RouteSpec{Auth: map[string]any{"type": "basic", "username": "rui", "password": "secret"}}
	if err := AuthorizeHTTPRequest(spec, req, nil); err != nil {
		t.Fatalf("basic should pass: %v", err)
	}
}

func TestAuthorizeHTTPRequestBasicRejectsWrongPassword(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.SetBasicAuth("rui", "wrong")
	spec := RouteSpec{Auth: map[string]any{"type": "basic", "username": "rui", "password": "secret"}}
	if err := AuthorizeHTTPRequest(spec, req, nil); err == nil {
		t.Fatal("expected wrong basic password to fail")
	}
}

func TestAuthorizeHTTPRequestBasicReportsGenericKeychainPlaceholder(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.SetBasicAuth("rui", "secret")
	spec := RouteSpec{Auth: map[string]any{"type": "basic", "username": "rui", "keychain": "connect/basic-password"}}
	if err := AuthorizeHTTPRequest(spec, req, nil); err == nil || err.Error() != "auth keychain references are not wired in gi yet: connect/basic-password" {
		t.Fatalf("expected generic keychain placeholder error, got %v", err)
	}
}

func TestAuthorizeHTTPRequestHMAC(t *testing.T) {
	body := []byte(`{"ok":true}`)
	mac := hmac.New(sha256.New, []byte("topsecret"))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("x-signature", "sha256="+sig)
	spec := RouteSpec{Auth: map[string]any{"type": "hmac", "secret": "topsecret"}}
	if err := AuthorizeHTTPRequest(spec, req, body); err != nil {
		t.Fatalf("hmac should pass: %v", err)
	}
}
