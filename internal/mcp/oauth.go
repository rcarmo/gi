package mcp

// OAuth sign-in for remote MCP servers, ported from Pi (pi-coding-agent
// extensions/mcp/oauth.js and pi-mcp oauth/{flow,discovery,provider}.js).
//
// Connections never start a browser flow on their own. They send the stored
// access token and, after a 401, try the stored refresh token. When that is
// not possible the server is marked needs-auth and the user signs in with
// `gi mcp login` or `/mcp login`, which runs the authorization code flow
// (PKCE, dynamic client registration) against a loopback callback.
//
// Credentials live in mcp-auth.json (Pi's file and format, keyed by server
// URL), so gi and Pi share sign-ins.

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rcarmo/gi/internal/lockdir"
)

const (
	oauthCallbackHost       = "127.0.0.1"
	oauthCallbackPath       = "/callback"
	oauthFallbackRedirect   = "http://" + oauthCallbackHost + oauthCallbackPath
	oauthRefreshSkew        = 30 * time.Second
	oauthRefreshTimeout     = 15 * time.Second
	oauthRefreshLockStale   = 20 * time.Second
	oauthRefreshLockWait    = 25 * time.Second
	oauthFileLockStale      = 10 * time.Second
	mcpProtocolVersionOAuth = "2025-11-25"
)

// ErrAuthorizationRequired means the user has to sign in (again).
var ErrAuthorizationRequired = errors.New("MCP server requires sign-in")

// ErrSignInCancelled is returned when a sign-in is cancelled or times out.
var ErrSignInCancelled = errors.New("Sign-in cancelled")

// OAuthError is an OAuth error response.
type OAuthError struct{ Code, Description string }

func (e *OAuthError) Error() string { return e.Description }

type insecureEndpointError struct{ url string }

func (e *insecureEndpointError) Error() string {
	return "Refusing to send OAuth credentials to non-HTTPS endpoint " + e.url
}

// issuerMismatchError ports Pi's OAuthIssuerMismatchError.
func issuerMismatchError(expected string, received *string) error {
	got := "none"
	if received != nil {
		got = strconvQuote(*received)
	}
	return fmt.Errorf("OAuth issuer mismatch: expected %s, received %s", strconvQuote(expected), got)
}

func strconvQuote(s string) string { b, _ := json.Marshal(s); return string(b) }

// OAuthSettings are a server's "oauth" settings.
type OAuthSettings struct {
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	CallbackPort *int   `json:"callbackPort,omitempty"`
	CallbackURL  string `json:"callbackUrl,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ClientName   string `json:"clientName,omitempty"`
	// ClientRegistration is "dcr" (dynamic client registration, the
	// default) or "cimd": gi identifies with pi's Client ID Metadata
	// Document instead of registering (Pi 1.0.1).
	ClientRegistration string `json:"clientRegistration,omitempty"`
	// AuthServerMetadataURL is an authorization server metadata document
	// used instead of discovery (servers that advertise a wrong one or none).
	AuthServerMetadataURL string `json:"authServerMetadataUrl,omitempty"`
}

// UsesOAuth ports Pi's usesOAuth: HTTP servers without an Authorization
// header.
func (s ServerConfig) UsesOAuth() bool {
	if s.Transport != "http" || s.Auth != nil {
		return false
	}
	for key := range s.Headers {
		if strings.EqualFold(key, "authorization") {
			return false
		}
	}
	return true
}

// oauthSettings resolves the server's oauth settings (clientSecret may be
// ${NAME} or !command).
func (m *Manager) oauthSettings(ctx context.Context, s ServerConfig) (OAuthSettings, error) {
	var settings OAuthSettings
	if len(s.OAuth) > 0 && string(s.OAuth) != "null" {
		if err := json.Unmarshal(s.OAuth, &settings); err != nil {
			return settings, fmt.Errorf("MCP server %q oauth: %w", s.Name, err)
		}
	}
	if settings.ClientSecret != "" {
		secret, err := expandValue(ctx, settings.ClientSecret, m.lookupEnv)
		if err != nil {
			return settings, fmt.Errorf("MCP server %q oauth.clientSecret: %w", s.Name, err)
		}
		settings.ClientSecret = secret
	}
	return settings, nil
}

// --- credential store -----------------------------------------------------

type oauthTokens struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    *float64 `json:"expires_in,omitempty"`
	Scope        string   `json:"scope,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	IDToken      string   `json:"id_token,omitempty"`
}

type discoveryState struct {
	AuthorizationServerURL      string          `json:"authorizationServerUrl"`
	AuthorizationServerMetadata json.RawMessage `json:"authorizationServerMetadata,omitempty"`
	ResourceMetadata            json.RawMessage `json:"resourceMetadata,omitempty"`
	ResourceMetadataURL         string          `json:"resourceMetadataUrl,omitempty"`
}

// oauthState is one server's entry in mcp-auth.json (Pi's shape).
type oauthState struct {
	ServerURL         string          `json:"serverUrl"`
	Discovery         *discoveryState `json:"discovery,omitempty"`
	ClientInformation json.RawMessage `json:"clientInformation,omitempty"`
	Tokens            *oauthTokens    `json:"tokens,omitempty"`
	TokensExpireAt    *float64        `json:"tokensExpireAt,omitempty"`
	CodeVerifier      string          `json:"codeVerifier,omitempty"`
	OAuthState        string          `json:"oauthState,omitempty"`
}

// CredentialStore is mcp-auth.json.
type CredentialStore struct {
	Path    string
	LockDir string // refresh locks; empty: refreshes serialized in this process only
	mu      sync.Mutex
}

// NewCredentialStore opens the store at path, with refresh locks beside it.
func NewCredentialStore(path string) *CredentialStore {
	return &CredentialStore{Path: path, LockDir: filepath.Dir(path)}
}

// normalizeServerURL is JavaScript's String(new URL(u)) for the cases that
// matter: lower-case scheme and host, "/" for an empty path.
func normalizeServerURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		u.Host = u.Hostname()
	}
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}

func (c *CredentialStore) withFile(update func(states *orderedObject) bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o700); err != nil {
		return err
	}
	return lockdir.With(c.Path+".lock", oauthFileLockStale, oauthFileLockStale, func() error {
		states := &orderedObject{values: map[string]json.RawMessage{}}
		if data, err := os.ReadFile(c.Path); err == nil && strings.TrimSpace(string(data)) != "" {
			if parsed, ok := parseOrderedObject(data); ok {
				states = parsed
			}
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		if !update(states) {
			return nil
		}
		compact, _ := states.MarshalJSON()
		var out bytes.Buffer
		if err := json.Indent(&out, compact, "", "  "); err != nil {
			return err
		}
		out.WriteByte('\n')
		tmp := c.Path + ".tmp"
		if err := os.WriteFile(tmp, out.Bytes(), 0o600); err != nil {
			return err
		}
		return os.Rename(tmp, c.Path)
	})
}

// storeKeys ports Pi's: credentials are keyed by server name and URL, so
// servers sharing a URL keep separate accounts; older versions keyed them by
// URL alone (the legacy key).
func storeKeys(name, serverURL string) (key, legacy string) {
	legacy = normalizeServerURL(serverURL)
	return Namespace(name) + "|" + legacy, legacy
}

// load returns the server's state; the first server to load legacy state
// takes it over (others with the same URL sign in again).
func (c *CredentialStore) load(name, serverURL string) *oauthState {
	key, legacy := storeKeys(name, serverURL)
	var state *oauthState
	_ = c.withFile(func(states *orderedObject) bool {
		raw, ok := states.values[key]
		migrate := false
		if !ok {
			if raw, ok = states.values[legacy]; ok {
				migrate = true
			}
		}
		if ok {
			var s oauthState
			if json.Unmarshal(raw, &s) == nil {
				state = &s
			}
		}
		if migrate {
			states.remove(legacy)
			states.set(key, raw)
		}
		return migrate
	})
	return state
}

func (c *CredentialStore) save(name, serverURL string, state *oauthState) error {
	key, _ := storeKeys(name, serverURL)
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return c.withFile(func(states *orderedObject) bool {
		states.set(key, raw)
		return true
	})
}

// Remove deletes a server's credentials (or the legacy state it would take
// over); it reports whether any were stored.
func (c *CredentialStore) Remove(name, serverURL string) bool {
	key, legacy := storeKeys(name, serverURL)
	removed := false
	_ = c.withFile(func(states *orderedObject) bool {
		removed = states.remove(key) || states.remove(legacy)
		return removed
	})
	return removed
}

func (c *CredentialStore) withRefreshLock(name, serverURL string, fn func() error) error {
	if c.LockDir == "" {
		return fn()
	}
	key, _ := storeKeys(name, serverURL)
	sum := sha256.Sum256([]byte(key))
	lock := filepath.Join(c.LockDir, "mcp-auth-refresh-"+hex.EncodeToString(sum[:])[:16]) + ".lock"
	return lockdir.With(lock, oauthRefreshLockStale, oauthRefreshLockWait, fn)
}

// --- provider (Pi's McpOAuthProvider) --------------------------------------

type oauthProvider struct {
	name           string
	serverURL      string
	redirectURL    string
	clientMetadata *orderedObject
	configured     json.RawMessage // configured client (clientId/clientSecret)
	store          *CredentialStore
	onRedirect     func(*url.URL)
	// clientMetadataDocument picks the Client ID Metadata Document for
	// clientRegistration "cimd"; nil otherwise.
	clientMetadataDocument func(*authServerMeta) (clientDocument, error)
}

// clientDocument is a Client ID Metadata Document: its URL is the client
// ID, and its redirect URI may be specific to the MCP server.
type clientDocument struct{ url, redirectURL string }

// clientMetadataBaseURL is where pi.dev serves pi's Client ID Metadata
// Documents: client.json and <callback ID>/client.json. gi identifies as pi.
const clientMetadataBaseURL = "https://pi.dev/oauth"

// callbackID is Pi's: 12 characters identifying an MCP server URL in
// callback paths (computed like Codex does).
func callbackID(serverURL string) string {
	sum := sha256.Sum256([]byte(urlHref(serverURL)))
	return base64.RawURLEncoding.EncodeToString(sum[:9])
}

// urlHref is a URL without its fragment, normalized as WHATWG URL's href:
// lowercase scheme and host, no default port, "/" for an empty path.
func urlHref(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Fragment, u.RawFragment = "", ""
	u.Scheme, u.Host = strings.ToLower(u.Scheme), strings.ToLower(u.Host)
	if port := u.Port(); (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		u.Host = u.Hostname()
	}
	if u.Path == "" && u.RawPath == "" {
		u.Path = "/"
	}
	return u.String()
}

// pickClientMetadataDocument is Pi's clientMetadataDocument, chosen like
// Codex chooses its own. Without the iss parameter in authorization
// responses (RFC 9207), the redirect URI and the document are specific to
// the MCP server, so a response cannot be mixed up with one from another
// authorization server (RFC 9700 section 4.4.2.2).
func pickClientMetadataDocument(serverURL, redirectURL string, meta *authServerMeta) (clientDocument, error) {
	if meta == nil || !meta.ClientIDMetadataDocumentSupported || !contains(meta.TokenEndpointAuthMethodsSupported, "none") {
		return clientDocument{}, errors.New(`The authorization server does not support Client ID Metadata Documents for public clients; remove oauth.clientRegistration "cimd"`)
	}
	if meta.IssParameterSupported {
		return clientDocument{url: clientMetadataBaseURL + "/client.json", redirectURL: redirectURL}, nil
	}
	id := callbackID(serverURL)
	redirect, err := url.Parse(redirectURL)
	if err != nil {
		return clientDocument{}, err
	}
	redirect.Path, redirect.RawPath = oauthCallbackPath+"/"+id, ""
	return clientDocument{url: clientMetadataBaseURL + "/" + id + "/client.json", redirectURL: redirect.String()}, nil
}

func newOAuthProvider(name, serverURL string, store *CredentialStore, settings OAuthSettings, redirectURL string, onRedirect func(*url.URL)) *oauthProvider {
	displayName := settings.ClientName
	if displayName == "" {
		displayName = clientName
	}
	method := "none"
	if settings.ClientSecret != "" {
		method = "client_secret_post"
	}
	meta := &orderedObject{values: map[string]json.RawMessage{}}
	meta.set("client_name", jsonValue(displayName))
	meta.set("redirect_uris", jsonValue([]string{redirectURL}))
	meta.set("grant_types", jsonValue([]string{"authorization_code", "refresh_token"}))
	meta.set("response_types", jsonValue([]string{"code"}))
	meta.set("token_endpoint_auth_method", jsonValue(method))
	p := &oauthProvider{name: name, serverURL: normalizeServerURL(serverURL), redirectURL: redirectURL, clientMetadata: meta, store: store, onRedirect: onRedirect}
	if settings.ClientRegistration == "cimd" {
		p.clientMetadataDocument = func(meta *authServerMeta) (clientDocument, error) {
			return pickClientMetadataDocument(serverURL, redirectURL, meta)
		}
	}
	if settings.ClientID != "" {
		client := map[string]string{"client_id": settings.ClientID}
		if settings.ClientSecret != "" {
			client["client_secret"] = settings.ClientSecret
		}
		p.configured, _ = json.Marshal(client)
	}
	return p
}

func (p *oauthProvider) load() *oauthState {
	if s := p.store.load(p.name, p.serverURL); s != nil && s.ServerURL == p.serverURL {
		return s
	}
	return &oauthState{ServerURL: p.serverURL}
}

func (p *oauthProvider) update(fn func(*oauthState)) error {
	s := p.load()
	fn(s)
	return p.store.save(p.name, p.serverURL, s)
}

func (p *oauthProvider) state() (string, error) {
	if s := p.load().OAuthState; s != "" {
		return s, nil
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	value := hex.EncodeToString(b)
	return value, p.update(func(s *oauthState) { s.OAuthState = value })
}

func (p *oauthProvider) clientInformation() json.RawMessage {
	if p.configured != nil {
		return p.configured
	}
	return p.load().ClientInformation
}

// withScope ports Pi's: a token response without scope grants the requested
// scope (code exchange) or keeps the grant's (refresh), recorded so a
// step-up sign-in can keep it.
func withScope(tokens *oauthTokens, scope string) *oauthTokens {
	if tokens.Scope == "" && scope != "" {
		t := *tokens
		t.Scope = scope
		return &t
	}
	return tokens
}

// parseTokens ports Pi's parseOAuthTokens: null and "" are absent, and
// expires_in may be a numeric string.
func parseTokens(text []byte) (*oauthTokens, error) {
	var raw map[string]any
	if err := json.Unmarshal(text, &raw); err != nil {
		return nil, errors.New("invalid OAuth token response")
	}
	str := func(k string) string { v, _ := raw[k].(string); return v }
	t := &oauthTokens{AccessToken: str("access_token"), TokenType: str("token_type"), Scope: str("scope"), RefreshToken: str("refresh_token"), IDToken: str("id_token")}
	if t.AccessToken == "" {
		return nil, errors.New("OAuth token response: access_token must be a string")
	}
	if t.TokenType == "" {
		return nil, errors.New("OAuth token response: token_type must be a string")
	}
	switch v := raw["expires_in"].(type) {
	case float64:
		t.ExpiresIn = &v
	case string:
		if v != "" {
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, errors.New("Invalid expires_in")
			}
			t.ExpiresIn = &n
		}
	}
	return t, nil
}

func (p *oauthProvider) saveTokens(tokens *oauthTokens) error {
	return p.update(func(s *oauthState) {
		s.Tokens = tokens
		s.TokensExpireAt = nil
		if tokens.ExpiresIn != nil {
			at := float64(time.Now().UnixMilli()) + *tokens.ExpiresIn*1000
			s.TokensExpireAt = &at
		}
	})
}

func (p *oauthProvider) invalidate(kind string) error {
	return p.update(func(s *oauthState) {
		if kind == "all" || kind == "client" {
			s.ClientInformation = nil
		}
		if kind == "all" || kind == "tokens" {
			s.Tokens, s.TokensExpireAt = nil, nil
		}
		if kind == "all" {
			s.CodeVerifier, s.Discovery, s.OAuthState = "", nil, ""
		}
	})
}

// --- discovery -------------------------------------------------------------

type wwwAuthChallenge struct {
	ResourceMetadataURL string
	Scope               string
	Error               string
}

func challengeField(header, name string) string {
	re := regexp.MustCompile(`(?i)(?:^|[,\s])` + name + `=(?:"([^"]*)"|([^\s,]+))`)
	m := re.FindStringSubmatch(header)
	if m == nil {
		return ""
	}
	if m[1] != "" {
		return m[1]
	}
	return m[2]
}

func parseWWWAuthenticate(header string) wwwAuthChallenge {
	fields := strings.Fields(header)
	if len(fields) == 0 {
		return wwwAuthChallenge{}
	}
	if scheme := strings.ToLower(fields[0]); scheme != "bearer" && scheme != "dpop" {
		return wwwAuthChallenge{}
	}
	c := wwwAuthChallenge{Scope: challengeField(header, "scope"), Error: challengeField(header, "error")}
	if rm := challengeField(header, "resource_metadata"); rm != "" {
		if u, err := url.Parse(rm); err == nil && u.Scheme != "" {
			c.ResourceMetadataURL = u.String()
		}
	}
	return c
}

func isDiscoveryMiss(status int) bool { return (status >= 400 && status < 500) || status == 502 }

func pathSuffix(p string) string { return strings.TrimSuffix(p, "/") }

func fetchMetadata(ctx context.Context, client *http.Client, u string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("MCP-Protocol-Version", mcpProtocolVersionOAuth)
	return client.Do(req)
}

func origin(u *url.URL) string { return u.Scheme + "://" + u.Host }

func discoverResourceMetadata(ctx context.Context, client *http.Client, serverURL, resourceMetadataURL string) (json.RawMessage, error) {
	server, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	target := resourceMetadataURL
	if target == "" {
		target = origin(server) + "/.well-known/oauth-protected-resource" + pathSuffix(server.Path)
	}
	resp, err := fetchMetadata(ctx, client, target)
	if err != nil {
		return nil, err
	}
	if resourceMetadataURL == "" && server.Path != "/" && server.Path != "" && isDiscoveryMiss(resp.StatusCode) {
		resp.Body.Close()
		if resp, err = fetchMetadata(ctx, client, origin(server)+"/.well-known/oauth-protected-resource"); err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("HTTP %d loading OAuth protected resource metadata", resp.StatusCode)
	}
	var raw json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&raw); err != nil {
		return nil, err
	}
	var check struct {
		Resource string `json:"resource"`
	}
	if json.Unmarshal(raw, &check) != nil || check.Resource == "" {
		return nil, errors.New("invalid OAuth protected resource metadata")
	}
	return raw, nil
}

type authServerMeta struct {
	Issuer                            string   `json:"issuer"`
	ClientIDMetadataDocumentSupported bool     `json:"client_id_metadata_document_supported"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	IssParameterSupported             bool     `json:"authorization_response_iss_parameter_supported"`
}

// fetchConfiguredAuthServerMetadata loads a configured metadata document
// (trusted as configured: no issuer check).
func fetchConfiguredAuthServerMetadata(ctx context.Context, client *http.Client, u string) (json.RawMessage, string, error) {
	resp, err := fetchMetadata(ctx, client, u)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, "", fmt.Errorf("HTTP %d loading authorization server metadata from %s", resp.StatusCode, u)
	}
	var raw json.RawMessage
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&raw); err != nil {
		return nil, "", err
	}
	var meta authServerMeta
	if json.Unmarshal(raw, &meta) != nil || meta.Issuer == "" {
		return nil, "", errors.New("invalid authorization server metadata")
	}
	return raw, meta.Issuer, nil
}

func discoverAuthServerMetadata(ctx context.Context, client *http.Client, asURL string) (json.RawMessage, error) {
	issuer, err := url.Parse(asURL)
	if err != nil {
		return nil, err
	}
	p := pathSuffix(issuer.Path)
	urls := []string{origin(issuer) + "/.well-known/oauth-authorization-server" + p, origin(issuer) + "/.well-known/openid-configuration" + p}
	if p != "" {
		urls = append(urls, origin(issuer)+p+"/.well-known/openid-configuration")
	}
	for _, u := range urls {
		resp, err := fetchMetadata(ctx, client, u)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			resp.Body.Close()
			if isDiscoveryMiss(resp.StatusCode) {
				continue
			}
			return nil, fmt.Errorf("HTTP %d loading authorization server metadata from %s", resp.StatusCode, u)
		}
		var raw json.RawMessage
		err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&raw)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		var meta authServerMeta
		if json.Unmarshal(raw, &meta) != nil || meta.Issuer == "" {
			return nil, errors.New("invalid authorization server metadata")
		}
		if strings.TrimSuffix(meta.Issuer, "/") != strings.TrimSuffix(asURL, "/") {
			return nil, issuerMismatchError(asURL, &meta.Issuer)
		}
		return raw, nil
	}
	return nil, nil
}

func selectResource(serverURL string, resourceMetadata json.RawMessage) (string, error) {
	if len(resourceMetadata) == 0 {
		return "", nil
	}
	var meta struct {
		Resource string `json:"resource"`
	}
	_ = json.Unmarshal(resourceMetadata, &meta)
	requested, err1 := url.Parse(serverURL)
	configured, err2 := url.Parse(meta.Resource)
	mismatch := fmt.Errorf("Protected resource %s does not match MCP server %s", meta.Resource, serverURL)
	if err1 != nil || err2 != nil || origin(requested) != origin(configured) {
		return "", mismatch
	}
	withSlash := func(p string) string {
		if strings.HasSuffix(p, "/") {
			return p
		}
		return p + "/"
	}
	if !strings.HasPrefix(withSlash(requested.Path), withSlash(configured.Path)) {
		return "", mismatch
	}
	return meta.Resource, nil
}

// --- token endpoint ----------------------------------------------------------

func isLoopbackHost(h string) bool {
	return h == "localhost" || h == "127.0.0.1" || h == "[::1]" || h == "::1"
}

func secureEndpoint(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" && !isLoopbackHost(u.Hostname()) {
		return nil, &insecureEndpointError{raw}
	}
	return u, nil
}

type clientInfo struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func selectClientAuthMethod(info clientInfo, supported []string) string {
	has := func(m string) bool {
		for _, s := range supported {
			if s == m {
				return true
			}
		}
		return false
	}
	if h := info.TokenEndpointAuthMethod; (h == "client_secret_basic" || h == "client_secret_post" || h == "none") && (len(supported) == 0 || has(h)) {
		return h
	}
	if len(supported) == 0 {
		if info.ClientSecret != "" {
			return "client_secret_basic"
		}
		return "none"
	}
	switch {
	case info.ClientSecret != "" && has("client_secret_basic"):
		return "client_secret_basic"
	case info.ClientSecret != "" && has("client_secret_post"):
		return "client_secret_post"
	case has("none"):
		return "none"
	case info.ClientSecret != "":
		return "client_secret_post"
	}
	return "none"
}

func tokenRequest(ctx context.Context, client *http.Client, asURL string, meta *authServerMeta, info clientInfo, resource string, params url.Values) (*oauthTokens, error) {
	endpoint := ""
	if meta != nil {
		endpoint = meta.TokenEndpoint
	}
	if endpoint == "" {
		base, _ := url.Parse(asURL)
		endpoint = origin(base) + "/token"
	}
	u, err := secureEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if resource != "" {
		params.Set("resource", resource)
	}
	var supported []string
	if meta != nil {
		supported = meta.TokenEndpointAuthMethodsSupported
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	switch selectClientAuthMethod(info, supported) {
	case "client_secret_basic":
		if info.ClientSecret == "" {
			return nil, errors.New("client_secret_basic requires a client secret")
		}
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(info.ClientID+":"+info.ClientSecret)))
	case "client_secret_post":
		params.Set("client_id", info.ClientID)
		if info.ClientSecret != "" {
			params.Set("client_secret", info.ClientSecret)
		}
	default:
		params.Set("client_id", info.ClientID)
	}
	body := params.Encode()
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	text, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var value map[string]any
	_ = json.Unmarshal(text, &value)
	if code, ok := value["error"].(string); ok {
		desc, _ := value["error_description"].(string)
		if desc == "" {
			desc = code
		}
		return nil, &OAuthError{Code: code, Description: desc}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &OAuthError{Code: "server_error", Description: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, text)}
	}
	return parseTokens(text)
}

func registerClient(ctx context.Context, client *http.Client, asURL string, meta *authServerMeta, clientMetadata *orderedObject, scope string) (json.RawMessage, error) {
	endpoint := ""
	if meta != nil {
		if meta.RegistrationEndpoint == "" {
			return nil, errors.New("Authorization server does not support dynamic client registration")
		}
		endpoint = meta.RegistrationEndpoint
	} else {
		base, _ := url.Parse(asURL)
		endpoint = origin(base) + "/register"
	}
	body := &orderedObject{keys: append([]string(nil), clientMetadata.keys...), values: map[string]json.RawMessage{}}
	for k, v := range clientMetadata.values {
		body.values[k] = v
	}
	if scope != "" {
		body.set("scope", jsonValue(scope))
	}
	encoded, _ := body.MarshalJSON()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	text, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("OAuth dynamic client registration failed with status %d: %s", resp.StatusCode, text)
	}
	var info clientInfo
	if json.Unmarshal(text, &info) != nil || info.ClientID == "" {
		return nil, errors.New("invalid OAuth client registration response")
	}
	return text, nil
}

// --- flow (Pi's runFlow / authorizeMcp) ------------------------------------

type flowOptions struct {
	serverURL           string
	resourceMetadataURL string
	metadataURL         string  // configured authorization server metadata
	iss                 *string // RFC 9207 iss of the authorization response
	scope               string
	authorizationCode   string
	skipRefresh         bool
	client              *http.Client
}

// runFlow returns true when authorized, false after redirecting the user.
func runFlow(ctx context.Context, p *oauthProvider, opts flowOptions) (bool, error) {
	client := opts.client
	if opts.metadataURL != "" {
		if _, err := secureEndpoint(opts.metadataURL); err != nil {
			return false, err
		}
	}
	state := p.load()
	var disc discoveryState
	// With a configured metadata URL, discovery is not cached, so changing
	// the URL applies at once.
	if cached := state.Discovery; opts.metadataURL == "" && cached != nil && cached.AuthorizationServerURL != "" {
		disc = *cached
		if len(disc.AuthorizationServerMetadata) == 0 {
			meta, err := discoverAuthServerMetadata(ctx, client, disc.AuthorizationServerURL)
			if err != nil {
				return false, err
			}
			disc.AuthorizationServerMetadata = meta
		}
	} else {
		resourceMeta, err := discoverResourceMetadata(ctx, client, opts.serverURL, opts.resourceMetadataURL)
		if err != nil {
			resourceMeta = nil // Pi: missing resource metadata falls back to the server origin
		}
		disc.ResourceMetadata = resourceMeta
		if opts.metadataURL != "" {
			meta, issuer, err := fetchConfiguredAuthServerMetadata(ctx, client, opts.metadataURL)
			if err != nil {
				return false, err
			}
			disc.AuthorizationServerURL, disc.AuthorizationServerMetadata = issuer, meta
		} else {
			var rm struct {
				AuthorizationServers []string `json:"authorization_servers"`
			}
			_ = json.Unmarshal(resourceMeta, &rm)
			if len(rm.AuthorizationServers) > 0 {
				disc.AuthorizationServerURL = rm.AuthorizationServers[0]
			} else {
				u, _ := url.Parse(opts.serverURL)
				disc.AuthorizationServerURL = origin(u) + "/"
			}
			meta, err := discoverAuthServerMetadata(ctx, client, disc.AuthorizationServerURL)
			if err != nil {
				return false, err
			}
			disc.AuthorizationServerMetadata = meta
		}
	}
	if opts.metadataURL == "" {
		if opts.resourceMetadataURL != "" {
			disc.ResourceMetadataURL = opts.resourceMetadataURL
		}
		if err := p.update(func(s *oauthState) { d := disc; s.Discovery = &d }); err != nil {
			return false, err
		}
	}
	var meta *authServerMeta
	if len(disc.AuthorizationServerMetadata) > 0 && string(disc.AuthorizationServerMetadata) != "null" {
		meta = &authServerMeta{}
		_ = json.Unmarshal(disc.AuthorizationServerMetadata, meta)
	}
	resource, err := selectResource(opts.serverURL, disc.ResourceMetadata)
	if err != nil {
		return false, err
	}
	scope := opts.scope
	if scope == "" {
		var rm struct {
			ScopesSupported []string `json:"scopes_supported"`
		}
		_ = json.Unmarshal(disc.ResourceMetadata, &rm)
		scope = strings.Join(rm.ScopesSupported, " ")
	}
	rawClient := p.clientInformation()
	// A Client ID Metadata Document is not stored: its URL is the client ID
	// of every sign-in.
	var doc *clientDocument
	if len(rawClient) == 0 && p.clientMetadataDocument != nil {
		d, err := p.clientMetadataDocument(meta)
		if err != nil {
			return false, err
		}
		if u, err := url.Parse(d.url); err != nil || u.Scheme != "https" || u.Path == "/" {
			return false, errors.New("Invalid OAuth client metadata URL")
		}
		doc = &d
		rawClient, _ = json.Marshal(map[string]string{"client_id": d.url})
	}
	// The document's redirect URI may differ from the provider's, for
	// example by a server-specific path.
	redirectURL := p.redirectURL
	if doc != nil {
		redirectURL = doc.redirectURL
	}
	if len(rawClient) == 0 {
		if opts.authorizationCode != "" {
			return false, errors.New("OAuth client information is missing during code exchange")
		}
		registered, err := registerClient(ctx, client, disc.AuthorizationServerURL, meta, p.clientMetadata, scope)
		if err != nil {
			return false, err
		}
		if err := p.update(func(s *oauthState) { s.ClientInformation = registered }); err != nil {
			return false, err
		}
		rawClient = registered
	}
	var info clientInfo
	_ = json.Unmarshal(rawClient, &info)
	if opts.authorizationCode != "" {
		// RFC 9207: never send a code from another authorization server.
		if meta != nil && (opts.iss != nil || meta.IssParameterSupported) {
			if opts.iss == nil || *opts.iss != meta.Issuer {
				return false, issuerMismatchError(meta.Issuer, opts.iss)
			}
		}
		verifier := p.load().CodeVerifier
		if verifier == "" {
			return false, errors.New("No OAuth PKCE code verifier is stored")
		}
		tokens, err := tokenRequest(ctx, client, disc.AuthorizationServerURL, meta, info, resource, url.Values{
			"grant_type": {"authorization_code"}, "code": {opts.authorizationCode}, "code_verifier": {verifier}, "redirect_uri": {redirectURL},
		})
		if err != nil {
			return false, err
		}
		return true, p.saveTokens(withScope(tokens, scope))
	}
	if existing := p.load().Tokens; !opts.skipRefresh && existing != nil && existing.RefreshToken != "" {
		tokens, err := tokenRequest(ctx, client, disc.AuthorizationServerURL, meta, info, resource, url.Values{
			"grant_type": {"refresh_token"}, "refresh_token": {existing.RefreshToken},
		})
		if err == nil {
			if tokens.RefreshToken == "" {
				tokens.RefreshToken = existing.RefreshToken
			}
			return true, p.saveTokens(withScope(tokens, existing.Scope))
		}
		var insecure *insecureEndpointError
		var oauthErr *OAuthError
		if errors.As(err, &insecure) || (errors.As(err, &oauthErr) && oauthErr.Code != "server_error") {
			return false, err
		}
	}
	oauthStateValue, err := p.state()
	if err != nil {
		return false, err
	}
	if meta != nil && !contains(meta.ResponseTypesSupported, "code") {
		return false, errors.New("Authorization server does not support authorization codes")
	}
	if meta != nil && meta.CodeChallengeMethodsSupported != nil && !contains(meta.CodeChallengeMethodsSupported, "S256") {
		return false, errors.New("Authorization server does not support PKCE S256")
	}
	endpoint := ""
	if meta != nil {
		endpoint = meta.AuthorizationEndpoint
	}
	if endpoint == "" {
		base, _ := url.Parse(disc.AuthorizationServerURL)
		endpoint = origin(base) + "/authorize"
	}
	authURL, err := url.Parse(endpoint)
	if err != nil {
		return false, err
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	verifier := base64.RawURLEncoding.EncodeToString(b)
	digest := sha256.Sum256([]byte(verifier))
	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", info.ClientID)
	q.Set("code_challenge", base64.RawURLEncoding.EncodeToString(digest[:]))
	q.Set("code_challenge_method", "S256")
	q.Set("redirect_uri", redirectURL)
	if oauthStateValue != "" {
		q.Set("state", oauthStateValue)
	}
	if scope != "" {
		q.Set("scope", scope)
		if contains(strings.Fields(scope), "offline_access") {
			q.Set("prompt", "consent")
		}
	}
	if resource != "" {
		q.Set("resource", resource)
	}
	authURL.RawQuery = q.Encode()
	if err := p.update(func(s *oauthState) { s.CodeVerifier = verifier }); err != nil {
		return false, err
	}
	p.onRedirect(authURL)
	return false, nil
}

func authorizeMCP(ctx context.Context, p *oauthProvider, opts flowOptions) (bool, error) {
	ok, err := runFlow(ctx, p, opts)
	var oauthErr *OAuthError
	if err != nil && errors.As(err, &oauthErr) {
		switch oauthErr.Code {
		case "invalid_client", "unauthorized_client":
			_ = p.invalidate("all")
			return runFlow(ctx, p, opts)
		case "invalid_grant":
			_ = p.invalidate("tokens")
			return runFlow(ctx, p, opts)
		}
	}
	return ok, err
}

// --- connection auth (Pi's createMcpAuthProvider) ---------------------------

func callbackSettingsFor(settings OAuthSettings) (host, redirectHost string, port *int, path, fixedRedirect string) {
	raw := settings.CallbackURL
	if raw == "" {
		raw = "http://" + oauthCallbackHost + oauthCallbackPath
	}
	u, err := url.Parse(raw)
	if err != nil {
		u, _ = url.Parse("http://" + oauthCallbackHost + oauthCallbackPath)
	}
	address := strings.Trim(u.Hostname(), "[]")
	port = settings.CallbackPort
	if u.Port() != "" {
		n, _ := strconv.Atoi(u.Port())
		port = &n
		fixedRedirect = settings.CallbackURL
	} else if port != nil {
		u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(*port))
		fixedRedirect = u.String()
	}
	host = address
	if address == "localhost" {
		host = oauthCallbackHost
	}
	return host, address, port, u.Path, fixedRedirect
}

type connectionAuth struct {
	name      string
	serverURL string
	store     *CredentialStore
	settings  func() (OAuthSettings, error)

	mu         sync.Mutex
	refreshing chan struct{}
	refreshErr error
	challenge  wwwAuthChallenge
	needsAuth  bool
}

func (a *connectionAuth) lastChallenge() wwwAuthChallenge {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.challenge
}

// refresh replaces staleToken; concurrent callers share one refresh and
// other processes are kept out by the store's refresh lock.
func (a *connectionAuth) refresh(ctx context.Context, staleToken string, challenge *wwwAuthChallenge) error {
	a.mu.Lock()
	if ch := a.refreshing; ch != nil {
		a.mu.Unlock()
		<-ch
		a.mu.Lock()
		err := a.refreshErr
		a.mu.Unlock()
		return err
	}
	ch := make(chan struct{})
	a.refreshing = ch
	a.mu.Unlock()
	err := a.store.withRefreshLock(a.name, a.serverURL, func() error {
		state := a.store.load(a.name, a.serverURL)
		current := ""
		if state != nil && state.Tokens != nil {
			current = state.Tokens.AccessToken
		}
		if current != staleToken {
			return nil // replaced meanwhile (another process refreshed, or a sign-in)
		}
		if state == nil || state.Tokens == nil || state.Tokens.RefreshToken == "" {
			return ErrAuthorizationRequired
		}
		settings, err := a.settings()
		if err != nil {
			return err
		}
		_, _, _, _, fixed := callbackSettingsFor(settings)
		redirect := fixed
		if redirect == "" {
			var info clientInfo
			_ = json.Unmarshal(state.ClientInformation, &info)
			if len(info.RedirectURIs) > 0 {
				redirect = info.RedirectURIs[0]
			} else {
				redirect = oauthFallbackRedirect
			}
		}
		p := newOAuthProvider(a.name, a.serverURL, a.store, settings, redirect, func(*url.URL) {})
		opts := flowOptions{serverURL: a.serverURL, metadataURL: settings.AuthServerMetadataURL, client: &http.Client{Timeout: oauthRefreshTimeout}}
		if challenge != nil {
			opts.resourceMetadataURL, opts.scope = challenge.ResourceMetadataURL, challenge.Scope
		}
		ok, err := authorizeMCP(ctx, p, opts)
		if err != nil {
			return err
		}
		if !ok {
			return ErrAuthorizationRequired
		}
		return nil
	})
	a.mu.Lock()
	a.refreshErr, a.refreshing = err, nil
	a.mu.Unlock()
	close(ch)
	return err
}

func (a *connectionAuth) token(ctx context.Context) string {
	a.mu.Lock()
	if ch := a.refreshing; ch != nil {
		a.mu.Unlock()
		<-ch
	} else {
		a.mu.Unlock()
	}
	state := a.store.load(a.name, a.serverURL)
	if state == nil || state.Tokens == nil {
		return ""
	}
	expired := state.TokensExpireAt != nil && time.UnixMilli(int64(*state.TokensExpireAt)).Add(-oauthRefreshSkew).Before(time.Now())
	if !expired || state.Tokens.RefreshToken == "" {
		return state.Tokens.AccessToken
	}
	// Failures fall through: the request goes out with the old token and a
	// 401 decides what happens.
	_ = a.refresh(ctx, state.Tokens.AccessToken, nil)
	if state = a.store.load(a.name, a.serverURL); state != nil && state.Tokens != nil {
		return state.Tokens.AccessToken
	}
	return ""
}

func (a *connectionAuth) onUnauthorized(ctx context.Context, header http.Header, token string) error {
	challenge := parseWWWAuthenticate(header.Get("WWW-Authenticate"))
	a.mu.Lock()
	a.challenge = challenge
	a.mu.Unlock()
	if challenge.Error == "insufficient_scope" {
		return ErrAuthorizationRequired
	}
	return a.refresh(ctx, token, &challenge)
}

// needsAuthorization: 401, or 403 with an insufficient_scope challenge.
func needsAuthorization(resp *http.Response) bool {
	if resp.StatusCode == http.StatusUnauthorized {
		return true
	}
	return resp.StatusCode == http.StatusForbidden && parseWWWAuthenticate(resp.Header.Get("WWW-Authenticate")).Error == "insufficient_scope"
}

// oauthTransport sends the stored token and, after a 401, refreshes and
// retries once; when the user must sign in it calls onNeedsAuth.
type oauthTransport struct {
	base        http.RoundTripper
	auth        *connectionAuth
	onNeedsAuth func()
}

func (t *oauthTransport) send(req *http.Request, token string) (*http.Response, error) {
	r := req.Clone(req.Context())
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		r.Body = body
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return t.base.RoundTrip(r)
}

func (t *oauthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	token := t.auth.token(ctx)
	resp, err := t.send(req, token)
	if err != nil || !needsAuthorization(resp) {
		return resp, err
	}
	header := resp.Header.Clone()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	resp.Body.Close()
	if err := t.auth.onUnauthorized(ctx, header, token); err != nil {
		if errors.Is(err, ErrAuthorizationRequired) {
			t.onNeedsAuth()
		}
		return nil, fmt.Errorf("%w: %v", ErrAuthorizationRequired, err)
	}
	if req.Body != nil && req.GetBody == nil {
		return nil, errors.New("cannot retry the request after refreshing the token")
	}
	resp, err = t.send(req, t.auth.token(ctx))
	if err == nil && needsAuthorization(resp) {
		t.onNeedsAuth()
	}
	return resp, err
}

// --- sign-in (Pi's signInMcpServer) -----------------------------------------

// SignInPrompt shows the authorization URL and may return a pasted redirect
// URL (empty: wait for the browser callback only).
type SignInPrompt struct {
	ShowAuthorizationURL func(string)
	PromptForRedirectURL func(ctx context.Context) string
}

// stepUpScope ports Pi's: the challenged scopes plus the granted ones, since
// a challenge may list only the missing scopes (SEP-2350); "" without
// challenged scopes.
func stepUpScope(granted, challenged string) string {
	if challenged == "" {
		return ""
	}
	return mergeScopes(granted, challenged)
}

func mergeScopes(scopes ...string) string {
	seen := map[string]bool{}
	var out []string
	for _, s := range scopes {
		for _, f := range strings.Fields(s) {
			if !seen[f] {
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	return strings.Join(out, " ")
}

func codeFromRedirectURL(input, state string, redirect *url.URL) (string, *string, error) {
	u, err := url.Parse(strings.TrimSpace(input))
	if err != nil || u.Scheme == "" {
		return "", nil, errors.New("Expected the full redirect URL from the browser address bar")
	}
	// A server-specific redirect URI tells authorization servers apart, so
	// it must match exactly.
	if origin(u) != origin(redirect) || urlPath(u) != urlPath(redirect) {
		return "", nil, errors.New("The redirect URL does not match this sign-in's redirect URI")
	}
	q := u.Query()
	if e := q.Get("error"); e != "" {
		if d := q.Get("error_description"); d != "" {
			return "", nil, errors.New(d)
		}
		return "", nil, errors.New(e)
	}
	if q.Get("state") != state {
		return "", nil, errors.New("The redirect URL belongs to a different sign-in")
	}
	code := q.Get("code")
	if code == "" {
		return "", nil, errors.New("The redirect URL does not contain an authorization code")
	}
	return code, issParam(q), nil
}

func issParam(q url.Values) *string {
	if v := q.Get("iss"); v != "" {
		return &v
	}
	return nil
}

type callbackServer struct {
	listener    net.Listener
	server      *http.Server
	redirectURL string
	results     chan callbackResult
	mu          sync.Mutex
	expected    map[string]string // state -> the redirect URI path its response must arrive on
}

type callbackResult struct {
	code, state string
	iss         *string
	err         error
}

// expect records the path the response with state must arrive on (Pi's
// waitForCallback with a path).
func (c *callbackServer) expect(state, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.expected == nil {
		c.expected = map[string]string{}
	}
	c.expected[state] = path
}

// urlPath is a URL's path as WHATWG URL's pathname: "/" when empty.
func urlPath(u *url.URL) string {
	if u.EscapedPath() == "" {
		return "/"
	}
	return u.EscapedPath()
}

// listenForCallback serves the callback on path and on extraPaths (the
// server-specific redirect URI of a Client ID Metadata Document).
func listenForCallback(host, redirectHost, path string, extraPaths []string, port *int, required bool) (*callbackServer, error) {
	listen := func(p int) (net.Listener, error) { return net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(p))) }
	want := 0
	if port != nil {
		want = *port
	}
	ln, err := listen(want)
	if err != nil {
		if required || port == nil {
			return nil, err
		}
		if ln, err = listen(0); err != nil {
			return nil, err
		}
	}
	actual := ln.Addr().(*net.TCPAddr).Port
	hostPart := redirectHost
	if strings.Contains(hostPart, ":") {
		hostPart = "[" + hostPart + "]"
	}
	cs := &callbackServer{listener: ln, redirectURL: fmt.Sprintf("http://%s:%d%s", hostPart, actual, path), results: make(chan callbackResult, 1)}
	mux := http.NewServeMux()
	handle := func(w http.ResponseWriter, r *http.Request) {
		// Pi's callback replies (pi-mcp oauth/callback.js) and pages.
		q := r.URL.Query()
		reply := func(status int, page string) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(page))
		}
		res := callbackResult{code: q.Get("code"), state: q.Get("state"), iss: issParam(q)}
		cs.mu.Lock()
		want, expected := cs.expected[res.state]
		cs.mu.Unlock()
		if expected && urlPath(r.URL) != want {
			res.err = errors.New("The authorization response arrived on another redirect URI")
			reply(http.StatusBadRequest, oauthErrorHTML("Unexpected redirect URI", ""))
		} else if e := q.Get("error"); e != "" {
			desc := q.Get("error_description")
			if desc == "" {
				desc = e
			}
			res.err = errors.New(desc)
			reply(http.StatusOK, oauthErrorHTML("Authorization failed. You may close this window.", desc))
		} else if res.code == "" {
			res.err = errors.New("OAuth callback did not include an authorization code")
			reply(http.StatusBadRequest, oauthErrorHTML("Missing authorization code", ""))
		} else {
			reply(http.StatusOK, oauthSuccessHTML("Signed in to the MCP server. You may now close this page."))
		}
		select {
		case cs.results <- res:
		default:
		}
	}
	for _, p := range append([]string{path}, extraPaths...) {
		mux.HandleFunc(p, handle)
	}
	cs.server = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = cs.server.Serve(ln) }()
	return cs, nil
}

func (c *callbackServer) close() { _ = c.server.Close() }

// SignIn signs in to an OAuth server: it reuses the stored refresh token
// when possible, else runs the browser authorization code flow.
func (m *Manager) SignIn(ctx context.Context, name string, prompt SignInPrompt) error {
	s, err := m.server(name)
	if err != nil {
		return err
	}
	if !s.cfg.UsesOAuth() {
		return fmt.Errorf("MCP server %q does not use OAuth. Only HTTP servers without an Authorization header do.", name)
	}
	if m.credentials == nil {
		return errors.New("MCP OAuth credentials are not available")
	}
	settings, err := m.oauthSettings(ctx, s.cfg)
	if err != nil {
		return err
	}
	var challenge wwwAuthChallenge
	if s.auth != nil {
		challenge = s.auth.lastChallenge()
	}
	return signIn(ctx, name, s.cfg.URL, m.credentials, settings, challenge, prompt)
}

func signIn(ctx context.Context, name, serverURL string, store *CredentialStore, settings OAuthSettings, challenge wwwAuthChallenge, prompt SignInPrompt) error {
	stored := store.load(name, serverURL)
	stepUp := challenge.Error == "insufficient_scope"
	host, redirectHost, port, path, fixed := callbackSettingsFor(settings)
	preferred := port
	if preferred == nil && stored != nil {
		var info clientInfo
		_ = json.Unmarshal(stored.ClientInformation, &info)
		if len(info.RedirectURIs) > 0 {
			if u, err := url.Parse(info.RedirectURIs[0]); err == nil && u.Port() != "" {
				n, _ := strconv.Atoi(u.Port())
				preferred = &n
			}
		}
	}
	cimd := settings.ClientRegistration == "cimd"
	var extraPaths []string
	if cimd { // the redirect URI of a server-specific Client ID Metadata Document
		extraPaths = []string{oauthCallbackPath + "/" + callbackID(serverURL)}
	}
	callback, err := listenForCallback(host, redirectHost, path, extraPaths, preferred, port != nil)
	if err != nil {
		return err
	}
	defer callback.close()
	redirectURL := fixed
	if redirectURL == "" {
		redirectURL = callback.redirectURL
	}
	if stored != nil {
		next := *stored
		next.OAuthState = "" // every sign-in gets a fresh state
		var info clientInfo
		_ = json.Unmarshal(stored.ClientInformation, &info)
		// A registered client cannot use another redirect URI, and its tokens
		// belong to it. A Client ID Metadata Document is not stored, so with
		// one, a stored client was registered before and is replaced.
		keepClient := settings.ClientID != "" || (!cimd && contains(info.RedirectURIs, redirectURL)) || (cimd && len(stored.ClientInformation) == 0)
		if !keepClient {
			next.ClientInformation, next.Tokens, next.TokensExpireAt = nil, nil, nil
		}
		if err := store.save(name, serverURL, &next); err != nil {
			return err
		}
	}
	var authorizationURL *url.URL
	p := newOAuthProvider(name, serverURL, store, settings, redirectURL, func(u *url.URL) { authorizationURL = u })
	// A server asking for more scope gets it on top of the configured scope
	// and, since the challenge may list only the missing scopes, on top of the
	// scope granted so far.
	requested := challenge.Scope
	if stepUp {
		granted := ""
		if stored != nil && stored.Tokens != nil {
			granted = stored.Tokens.Scope
		}
		requested = stepUpScope(granted, challenge.Scope)
	}
	flow := flowOptions{serverURL: normalizeServerURL(serverURL), resourceMetadataURL: challenge.ResourceMetadataURL,
		metadataURL: settings.AuthServerMetadataURL, scope: mergeScopes(settings.Scope, requested), client: &http.Client{Timeout: 30 * time.Second}}
	skip := flow
	skip.skipRefresh = stepUp
	ok, err := authorizeMCP(ctx, p, skip)
	if err != nil || ok {
		return err
	}
	if authorizationURL == nil {
		return errors.New("OAuth flow did not produce an authorization URL")
	}
	state, err := p.state()
	if err != nil {
		return err
	}
	// The flow picks the redirect URI, which may be specific to the MCP server.
	authorizationRedirect, err := url.Parse(redirectURL)
	if r := authorizationURL.Query().Get("redirect_uri"); r != "" {
		authorizationRedirect, err = url.Parse(r)
	}
	if err != nil {
		return err
	}
	prompt.ShowAuthorizationURL(authorizationURL.String())
	code, iss, err := waitForAuthorizationCode(ctx, callback, state, authorizationRedirect, prompt)
	if err != nil {
		return err
	}
	flow.authorizationCode, flow.iss = code, iss
	_, err = authorizeMCP(ctx, p, flow)
	return err
}

func waitForAuthorizationCode(ctx context.Context, callback *callbackServer, state string, redirect *url.URL, prompt SignInPrompt) (string, *string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	callback.expect(state, urlPath(redirect))
	pasted := make(chan callbackResult, 1)
	if prompt.PromptForRedirectURL != nil {
		go func() {
			input := prompt.PromptForRedirectURL(ctx)
			if strings.TrimSpace(input) == "" {
				if ctx.Err() == nil {
					pasted <- callbackResult{err: ErrSignInCancelled}
				}
				return
			}
			code, iss, err := codeFromRedirectURL(input, state, redirect)
			pasted <- callbackResult{code: code, state: state, iss: iss, err: err}
		}()
	}
	for {
		select {
		case <-ctx.Done():
			return "", nil, ErrSignInCancelled
		case res := <-pasted:
			return res.code, res.iss, res.err
		case res := <-callback.results:
			if res.err != nil {
				return "", nil, res.err
			}
			if res.state != state {
				continue // another sign-in's redirect
			}
			return res.code, res.iss, nil
		}
	}
}
