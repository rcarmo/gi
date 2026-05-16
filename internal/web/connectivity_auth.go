package web

import (
	"fmt"
	"net/http"
	"strings"

	giauth "github.com/rcarmo/gi/internal/auth"
	"github.com/rcarmo/gi/internal/connectivity"
)

func allowUnauthenticatedExternal(options map[string]any) bool {
	if options == nil {
		return false
	}
	switch v := options["allow_unauthenticated_external"].(type) {
	case bool:
		return v
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		return s == "true" || s == "1" || s == "yes"
	default:
		return false
	}
}

func connectivityAuthHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch err.Error() {
	case "external request requires route auth",
		"invalid bearer token",
		"basic auth username is required",
		"missing basic auth",
		"invalid basic auth",
		"header auth requires header",
		"invalid header auth",
		"invalid query auth",
		"invalid hmac signature",
		"invalid or missing TOTP bearer token":
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

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
	if allowUnauthenticatedExternal(spec.Options) && !giauth.LoopbackRequest(r) {
		return nil
	}
	if s.auth == nil {
		return fmt.Errorf("TOTP auth is not available")
	}
	ok, err := s.auth.ValidateBearerRequestWithError(r)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return fmt.Errorf("invalid or missing TOTP bearer token")
}
