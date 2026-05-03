package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rcarmo/gi/internal/connectivity"
)

func (s *Server) authorizeConnectivityRequest(spec connectivity.RouteSpec, r *http.Request, body []byte) error {
	typ := ""
	if spec.Auth != nil {
		if value, ok := spec.Auth["type"].(string); ok {
			typ = strings.ToLower(strings.TrimSpace(value))
		}
	}
	if typ != "totp" {
		return connectivity.AuthorizeHTTPRequest(spec, r, body)
	}
	if s.auth == nil {
		return fmt.Errorf("TOTP auth is not available")
	}
	if s.auth.ValidateBearerRequest(r) {
		return nil
	}
	return fmt.Errorf("invalid or missing TOTP bearer token")
}
