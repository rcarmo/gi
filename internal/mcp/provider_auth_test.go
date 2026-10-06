package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/mcp/mcptest"
)

func TestProviderAuthConfigAndOAuthExclusion(t *testing.T) {
	for _, address := range []string{"https://example.com/mcp", "http://localhost:9090/mcp", "http://127.0.0.2/mcp", "http://[::1]:8080/mcp"} {
		s, err := parseServer("server", json.RawMessage(fmt.Sprintf(`{"url":%q,"auth":{"provider":"radius"}}`, address)), "")
		if err != nil || s.Auth == nil || s.Auth.Provider != "radius" || s.UsesOAuth() {
			t.Fatal(s, err)
		}
	}
	for _, raw := range []string{
		`{"command":"x","auth":{"provider":"radius"}}`,
		`{"url":"http://example.com/mcp","auth":{"provider":"radius"}}`,
		`{"url":"http://127.0.0.1.evil/mcp","auth":{"provider":"radius"}}`,
		`{"url":"https://user:password@example.com/mcp","auth":{"provider":"radius"}}`,
		`{"url":"https://example.com/mcp#fragment","auth":{"provider":"radius"}}`,
		`{"url":"https://example.com/mcp","auth":null}`,
		`{"url":"https://example.com/mcp","auth":{"provider":""}}`,
		`{"url":"https://example.com/mcp","auth":{"provider":123}}`,
		`{"url":"https://example.com/mcp","auth":{"provider":"radius","token":"secret"}}`,
		`{"url":"https://example.com/mcp","auth":{"provider":"radius"},"oauth":{}}`,
		`{"url":"https://example.com/mcp","auth":{"provider":"radius"},"headers":{"authorization":"Bearer secret"}}`,
	} {
		if _, err := parseServer("server", json.RawMessage(raw), ""); err == nil {
			t.Fatal("accepted invalid auth", raw)
		}
	}
}

func TestProviderAuthManagerRotationRevocationAndNoOAuth(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	if err := inference.SaveAPIKeyLogin("provider-fixture", "first-token"); err != nil {
		t.Fatal(err)
	}
	fixture := mcptest.New("")
	defer fixture.Close()
	handler := fixture.HTTP.Config.Handler
	var mu sync.Mutex
	var tokens []string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		tokens = append(tokens, r.Header.Get("Authorization"))
		mu.Unlock()
		if r.Header.Get("Authorization") == "" {
			t.Error("missing token")
		}
		handler.ServeHTTP(w, r)
	}))
	defer provider.Close()
	cfg := LoadConfig(writeJSON(t, filepath.Join(dir, "mcp.json"), fmt.Sprintf(`{"mcpServers":{"remote":{"url":%q,"auth":{"provider":"provider-fixture"}}}}`, provider.URL)), "", false)
	if len(cfg.Errors) != 0 {
		t.Fatal(cfg.Errors)
	}
	m := NewManager(cfg, t.TempDir(), "")
	m.SetCredentials(NewCredentialStore(filepath.Join(dir, "mcp-auth.json")))
	defer m.Close()
	ctx := context.Background()
	if out, err := m.CallTool(ctx, "remote", "echo", map[string]any{"text": "one"}); err != nil || out.Content == nil {
		t.Fatal(out, err)
	}
	if err := inference.SaveAPIKeyLogin("provider-fixture", "rotated-token"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.CallTool(ctx, "remote", "echo", map[string]any{"text": "two"}); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	seen := append([]string(nil), tokens...)
	before := len(tokens)
	mu.Unlock()
	if seen[len(seen)-1] != "Bearer rotated-token" || seen[0] != "Bearer first-token" {
		t.Fatal(seen)
	}
	if err := m.SignIn(ctx, "remote", SignInPrompt{}); err == nil || !strings.Contains(err.Error(), "does not use OAuth") {
		t.Fatal(err)
	}
	if _, err := inference.RemoveAuthEntry("provider-fixture"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.CallTool(ctx, "remote", "echo", map[string]any{"text": "revoked"}); err == nil || strings.Contains(err.Error(), "rotated-token") {
		t.Fatal(err)
	}
	mu.Lock()
	after := len(tokens)
	mu.Unlock()
	if after != before {
		t.Fatal("request sent after revocation", before, after)
	}
}

func TestProviderAuthRedirectDoesNotLeakCredential(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	if err := inference.SaveAPIKeyLogin("provider-fixture", "redirect-canary"); err != nil {
		t.Fatal(err)
	}
	hits := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits++; t.Error("redirect target reached") }))
	defer target.Close()
	from := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer from.Close()
	cfg := LoadConfig(writeJSON(t, filepath.Join(dir, "mcp.json"), fmt.Sprintf(`{"mcpServers":{"remote":{"url":%q,"auth":{"provider":"provider-fixture"}}}}`, from.URL)), "", false)
	if len(cfg.Errors) != 0 {
		t.Fatal(cfg.Errors)
	}
	m := NewManager(cfg, t.TempDir(), "")
	defer m.Close()
	if _, err := m.Tools(context.Background(), "remote"); err == nil || strings.Contains(err.Error(), "redirect-canary") {
		t.Fatal(err)
	}
	if hits != 0 {
		t.Fatal(hits)
	}
	endpoint, _ := url.Parse(from.URL)
	transport := providerTransport{base: http.DefaultTransport, endpoint: endpoint, provider: "fixture", token: func(context.Context, string) (string, error) {
		t.Fatal("read credentials for changed endpoint")
		return "", nil
	}}
	request, _ := http.NewRequest("GET", target.URL, nil)
	if _, err := transport.RoundTrip(request); err == nil {
		t.Fatal("accepted changed endpoint")
	}
}

type providerRoundTripFunc func(*http.Request) (*http.Response, error)

func (f providerRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestProviderAuthTransportGuardsAndRedaction(t *testing.T) {
	endpoint, _ := url.Parse("https://example.com/mcp?fixed=1")
	calls := 0
	rt := providerTransport{endpoint: endpoint, provider: "fixture", token: func(context.Context, string) (string, error) { calls++; return "secret-canary", nil }, base: providerRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer secret-canary" || r.Header.Get("X-Fixture") != "kept" {
			t.Fatal("missing auth or configured header")
		}
		return nil, fmt.Errorf("secret-canary transport failure")
	})}
	req, _ := http.NewRequest("POST", endpoint.String(), nil)
	req.Header.Set("X-Fixture", "kept")
	if _, err := rt.RoundTrip(req); err == nil || strings.Contains(err.Error(), "secret-canary") {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatal("mutated caller headers")
	}
	for _, address := range []string{"http://example.com/mcp?fixed=1", "https://elsewhere.com/mcp?fixed=1", "https://example.com/other?fixed=1", "https://example.com/mcp?changed=1"} {
		r, _ := http.NewRequest("POST", address, nil)
		before := calls
		if _, err := rt.RoundTrip(r); err == nil {
			t.Fatal("unsafe endpoint", address)
		}
		if calls != before {
			t.Fatal("credentials read for unsafe endpoint")
		}
	}
	req.Host = "evil.example"
	before := calls
	if _, err := rt.RoundTrip(req); err == nil || calls != before {
		t.Fatal("Host override accepted", err)
	}
	req.Host = endpoint.Host
	rt.token = func(context.Context, string) (string, error) { return "", fmt.Errorf("refresh-canary") }
	if _, err := rt.RoundTrip(req); err == nil || strings.Contains(err.Error(), "refresh-canary") {
		t.Fatal(err)
	}
	rt.token = func(context.Context, string) (string, error) { return "two words", nil }
	if _, err := rt.RoundTrip(req); err == nil {
		t.Fatal("invalid bearer accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req = req.WithContext(ctx)
	if _, err := rt.RoundTrip(req); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestProviderAuthCLIUsesProviderNotMCPLogin(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	if err := inference.SaveAPIKeyLogin("provider-fixture", "cli-token"); err != nil {
		t.Fatal(err)
	}
	fixture := mcptest.New("")
	defer fixture.Close()
	path := writeJSON(t, filepath.Join(dir, "mcp.json"), fmt.Sprintf(`{"mcpServers":{"remote":{"url":%q,"auth":{"provider":"provider-fixture"}}}}`, fixture.URL))
	var out, errOut bytes.Buffer
	options := CLIOptions{UserPath: path, Cwd: t.TempDir(), CredentialsPath: filepath.Join(dir, "mcp-auth.json"), Log: func(s string) { fmt.Fprintln(&out, s) }, Error: func(s string) { fmt.Fprintln(&errOut, s) }, OpenURL: func(string) { t.Error("MCP OAuth browser opened") }}
	if code := RunCommand([]string{"list", "--json"}, options); code != 0 || !strings.Contains(out.String(), `"state": "connected"`) {
		t.Fatal(code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	if code := RunCommand([]string{"login", "remote"}, options); code == 0 || !strings.Contains(errOut.String(), "does not use OAuth") {
		t.Fatal(code, out.String(), errOut.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "mcp-auth.json")); !os.IsNotExist(err) {
		t.Fatal("MCP OAuth credentials created", err)
	}
}
