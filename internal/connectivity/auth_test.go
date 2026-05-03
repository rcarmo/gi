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

func TestAuthorizeHTTPRequestBearer(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("Authorization", "Bearer secret")
	spec := RouteSpec{Auth: map[string]any{"type": "bearer", "token": "secret"}}
	if err := AuthorizeHTTPRequest(spec, req, nil); err != nil {
		t.Fatalf("bearer should pass: %v", err)
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
