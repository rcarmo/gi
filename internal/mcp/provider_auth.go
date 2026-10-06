package mcp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// providerTransport reads the provider credential for each HTTP request,
// including SSE reconnects, instead of capturing it when connecting.
type providerTransport struct {
	base     http.RoundTripper
	endpoint *url.URL
	provider string
	token    func(context.Context, string) (string, error)
}

func (t providerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	if err := validateProviderURL(req.URL.String()); err != nil {
		return nil, err
	}
	// Redirects must not replay provider credentials at another origin or path.
	if req.URL.Scheme != t.endpoint.Scheme || !strings.EqualFold(req.URL.Host, t.endpoint.Host) || req.URL.EscapedPath() != t.endpoint.EscapedPath() || req.URL.RawQuery != t.endpoint.RawQuery || req.Host != "" && !strings.EqualFold(req.Host, t.endpoint.Host) {
		return nil, errors.New("provider-auth MCP endpoint changed")
	}
	token, err := t.token(req.Context(), t.provider)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, req.Context().Err()
		}
		return nil, errors.New("provider credentials unavailable; sign in to the configured provider")
	}
	if token == "" || strings.IndexFunc(token, func(r rune) bool { return r <= 32 || r == 127 }) >= 0 {
		return nil, errors.New("provider token unavailable or invalid")
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+token)
	response, err := t.base.RoundTrip(clone)
	// A RoundTripper error may contain response URLs or headers from a custom
	// transport; preserve cancellation but never surface credential canaries.
	if err != nil {
		if req.Context().Err() != nil {
			return nil, req.Context().Err()
		}
		return nil, errors.New("provider-auth MCP HTTP request failed")
	}
	return response, nil
}
