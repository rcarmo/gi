package connectivity

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/rcarmo/gi/internal/secrets"
)

// AuthorizeHTTPRequest enforces route auth for HTTP/SSE/WebSocket-style
// request dispatch. Loopback callers may use unauthenticated routes. Non-loopback
// callers must satisfy an auth block unless the route explicitly sets
// options.allow_unauthenticated_external=true.
func AuthorizeHTTPRequest(spec RouteSpec, r *http.Request, body []byte) error {
	return AuthorizeHTTPRequestWithResolver(spec, r, body, secrets.EnvResolver{})
}

func AuthorizeHTTPRequestWithResolver(spec RouteSpec, r *http.Request, body []byte, resolver secrets.Resolver) error {
	if isLoopbackRemote(r.RemoteAddr) {
		if len(spec.Auth) == 0 {
			return nil
		}
		return checkHTTPRequestAuth(spec.Auth, r, body, resolver)
	}
	if boolOption(spec.Options, "allow_unauthenticated_external") {
		return nil
	}
	if len(spec.Auth) == 0 {
		return fmt.Errorf("external request requires route auth")
	}
	return checkHTTPRequestAuth(spec.Auth, r, body, resolver)
}

func checkHTTPRequestAuth(auth map[string]any, r *http.Request, body []byte, resolver secrets.Resolver) error {
	typ := strings.ToLower(strings.TrimSpace(stringMapValue(auth, "type")))
	if typ == "" {
		typ = "bearer"
	}
	switch typ {
	case "none":
		return nil
	case "totp", "webauthn":
		return fmt.Errorf("auth type %s requires host middleware", typ)
	case "bearer":
		expected, err := authSecret(r.Context(), auth, resolver)
		if err != nil {
			return err
		}
		got := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(got), "bearer ") {
			got = strings.TrimSpace(got[len("bearer "):])
		}
		if constantTimeEqual(got, expected) {
			return nil
		}
		return fmt.Errorf("invalid bearer token")
	case "basic":
		expectedUser := firstNonEmpty(stringMapValue(auth, "username"), stringMapValue(auth, "user"))
		if expectedUser == "" {
			return fmt.Errorf("basic auth username is required")
		}
		expectedPassword, err := authPassword(r.Context(), auth, resolver)
		if err != nil {
			return err
		}
		gotUser, gotPassword, ok := r.BasicAuth()
		if !ok {
			return fmt.Errorf("missing basic auth")
		}
		if constantTimeEqual(gotUser, expectedUser) && constantTimeEqual(gotPassword, expectedPassword) {
			return nil
		}
		return fmt.Errorf("invalid basic auth")
	case "header":
		expected, err := authSecret(r.Context(), auth, resolver)
		if err != nil {
			return err
		}
		header := firstNonEmpty(stringMapValue(auth, "header"), stringMapValue(auth, "name"))
		if header == "" {
			return fmt.Errorf("header auth requires header")
		}
		if constantTimeEqual(r.Header.Get(header), expected) {
			return nil
		}
		return fmt.Errorf("invalid header auth")
	case "query":
		expected, err := authSecret(r.Context(), auth, resolver)
		if err != nil {
			return err
		}
		param := firstNonEmpty(stringMapValue(auth, "param"), stringMapValue(auth, "name"), "token")
		if constantTimeEqual(r.URL.Query().Get(param), expected) {
			return nil
		}
		return fmt.Errorf("invalid query auth")
	case "hmac":
		secret, err := authSecret(r.Context(), auth, resolver)
		if err != nil {
			return err
		}
		header := firstNonEmpty(stringMapValue(auth, "header"), "x-signature")
		got := strings.TrimSpace(r.Header.Get(header))
		if strings.Contains(got, "=") {
			parts := strings.SplitN(got, "=", 2)
			got = strings.TrimSpace(parts[1])
		}
		mac := hmac.New(sha256.New, []byte(secret))
		if _, err := mac.Write(body); err != nil {
			return fmt.Errorf("hmac write: %w", err)
		}
		expected := hex.EncodeToString(mac.Sum(nil))
		if constantTimeEqual(got, expected) {
			return nil
		}
		return fmt.Errorf("invalid hmac signature")
	default:
		return fmt.Errorf("unsupported auth type: %s", typ)
	}
}

func authSecret(ctx context.Context, auth map[string]any, resolver secrets.Resolver) (string, error) {
	if v := firstNonEmpty(stringMapValue(auth, "token"), stringMapValue(auth, "value"), stringMapValue(auth, "secret")); v != "" {
		return v, nil
	}
	if env := firstNonEmpty(stringMapValue(auth, "env"), stringMapValue(auth, "token_env"), stringMapValue(auth, "secret_env")); env != "" {
		if value := os.Getenv(env); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("auth env %s is not set", env)
	}
	if keychain := stringMapValue(auth, "keychain"); keychain != "" {
		if resolver == nil {
			return "", fmt.Errorf("auth secret resolver is not available for %s", keychain)
		}
		value, err := resolver.Resolve(ctx, keychain)
		if err != nil {
			return "", fmt.Errorf("resolve auth keychain %s: %w", keychain, err)
		}
		return value, nil
	}
	return "", fmt.Errorf("auth secret is required")
}

func authPassword(ctx context.Context, auth map[string]any, resolver secrets.Resolver) (string, error) {
	if v := firstNonEmpty(stringMapValue(auth, "password"), stringMapValue(auth, "pass"), stringMapValue(auth, "secret")); v != "" {
		return v, nil
	}
	if env := firstNonEmpty(stringMapValue(auth, "password_env"), stringMapValue(auth, "pass_env"), stringMapValue(auth, "secret_env"), stringMapValue(auth, "env")); env != "" {
		if value := os.Getenv(env); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("auth env %s is not set", env)
	}
	if keychain := firstNonEmpty(stringMapValue(auth, "password_keychain"), stringMapValue(auth, "keychain")); keychain != "" {
		if resolver == nil {
			return "", fmt.Errorf("auth password resolver is not available for %s", keychain)
		}
		value, err := resolver.Resolve(ctx, keychain)
		if err != nil {
			return "", fmt.Errorf("resolve auth keychain %s: %w", keychain, err)
		}
		return value, nil
	}
	return "", fmt.Errorf("basic auth password is required")
}

func isLoopbackRemote(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func stringMapValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

func boolOption(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	switch v := m[key].(type) {
	case bool:
		return v
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		return s == "true" || s == "1" || s == "yes"
	default:
		return false
	}
}
