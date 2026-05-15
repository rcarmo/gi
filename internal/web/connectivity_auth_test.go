package web

import (
	"net/http"
	"testing"

	"github.com/rcarmo/gi/internal/connectivity"
)

func TestAuthorizeConnectivityRequestAllowsExternalBypassForTOTP(t *testing.T) {
	srv := &Server{}
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	spec := connectivity.RouteSpec{
		Auth:    map[string]any{"type": "totp"},
		Options: map[string]any{"allow_unauthenticated_external": true},
	}
	if err := srv.authorizeConnectivityRequest(spec, req, nil); err != nil {
		t.Fatalf("expected TOTP external bypass to succeed, got %v", err)
	}
}

func TestAuthorizeConnectivityRequestStillRequiresTOTPBearerWithoutBypass(t *testing.T) {
	srv := &Server{}
	req, _ := http.NewRequest(http.MethodPost, "http://example.test/hook", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	spec := connectivity.RouteSpec{Auth: map[string]any{"type": "totp"}}
	if err := srv.authorizeConnectivityRequest(spec, req, nil); err == nil || err.Error() != "TOTP auth is not available" {
		t.Fatalf("expected missing TOTP auth error, got %v", err)
	}
}
