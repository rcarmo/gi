package inference

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
	"github.com/rcarmo/go-ai/oauth"
)

type fixtureTokenProvider struct {
	refreshes int
	cancel    bool
}

func (p *fixtureTokenProvider) ID() string   { return "mcp-oauth-fixture" }
func (p *fixtureTokenProvider) Name() string { return "MCP fixture" }
func (p *fixtureTokenProvider) Login(oauth.LoginCallbacks) (*oauth.Credentials, error) {
	panic("unexpected login")
}
func (p *fixtureTokenProvider) RefreshToken(*oauth.Credentials) (*oauth.Credentials, error) {
	panic("unbounded refresh")
}
func (p *fixtureTokenProvider) RefreshTokenContext(ctx context.Context, c *oauth.Credentials) (*oauth.Credentials, error) {
	if c.Extra["projectId"] != "project-fixture" {
		return nil, errors.New("provider-specific field lost")
	}
	p.refreshes++
	if p.cancel {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return &oauth.Credentials{Access: "fresh-access", Refresh: "fresh-refresh", Expires: time.Now().Add(time.Hour).UnixMilli(), Extra: map[string]any{"fixtureExtra": true}}, nil
}
func (p *fixtureTokenProvider) GetAPIKey(c *oauth.Credentials) string { return c.Access }
func (p *fixtureTokenProvider) ModifyModels(m []*goai.Model, _ *oauth.Credentials) []*goai.Model {
	return m
}

func TestProviderTokenRefreshPersistsAndPreservesFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	p := &fixtureTokenProvider{}
	oauth.RegisterProvider(p)
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"mcp-oauth-fixture":{"type":"oauth","access":"expired-access","refresh":"old-refresh","expires":1,"preserved":"keep","projectId":"project-fixture"},"other":{"type":"api_key","key":"other-key"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		token, err := ProviderToken(context.Background(), p.ID())
		if err != nil || token != "fresh-access" {
			t.Fatal(token, err)
		}
	}
	if p.refreshes != 1 {
		t.Fatal(p.refreshes)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc[p.ID()]["preserved"] != "keep" || doc[p.ID()]["refresh"] != "fresh-refresh" || doc[p.ID()]["fixtureExtra"] != true || doc["other"]["key"] != "other-key" {
		t.Fatal(doc)
	}
	if _, err := RemoveAuthEntry(p.ID()); err != nil {
		t.Fatal(err)
	}
	if _, err := ProviderToken(context.Background(), p.ID()); err == nil {
		t.Fatal("used token after logout")
	}
}

func TestProviderTokenCancellationAndInvalidTokens(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	p := &fixtureTokenProvider{cancel: true}
	oauth.RegisterProvider(p)
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"mcp-oauth-fixture":{"type":"oauth","refresh":"refresh-canary","expires":1,"projectId":"project-fixture"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := ProviderToken(ctx, p.ID()); err == nil || !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatal(err)
	}
	for _, token := range []string{"", "bad\r\ntoken", "two words", "nul\x00byte"} {
		if _, err := validProviderToken(token); err == nil {
			t.Fatal("accepted token", token)
		}
	}
	for _, raw := range []string{`{"unregistered":{"type":"oauth","access":"expired-canary","expires":1}}`, `{"unregistered":{"type":"api_key","key":"bad\r\nkey"}}`, `not-json`} {
		if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if token, err := ProviderToken(context.Background(), "unregistered"); err == nil || token != "" || strings.Contains(err.Error(), "canary") {
			t.Fatal(token, err)
		}
	}
}

func TestProviderTokenLockCancellation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	p := &fixtureTokenProvider{}
	oauth.RegisterProvider(p)
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"mcp-oauth-fixture":{"type":"oauth","access":"expired","refresh":"refresh","expires":1,"projectId":"project-fixture"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"process", "directory"} {
		t.Run(kind, func(t *testing.T) {
			if kind == "process" {
				credentialMu.Lock()
				defer credentialMu.Unlock()
			} else {
				if err := os.Mkdir(filepath.Join(dir, "auth.json.lock"), 0700); err != nil {
					t.Fatal(err)
				}
				defer os.Remove(filepath.Join(dir, "auth.json.lock"))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			started := time.Now()
			if _, err := ProviderToken(ctx, p.ID()); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
			if time.Since(started) > time.Second {
				t.Fatal("cancellation delayed by credential lock")
			}
		})
	}
	if p.refreshes != 0 {
		t.Fatal("refreshed while locked")
	}
}

func TestProviderTokenSelectedEntryAndLegacyKeys(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	for _, body := range []string{`{"selected":{"type":"api-key","apiKey":"legacy-key"},"other":{"expires":"unrelated-invalid-field"}}`, `{"selected":{"key":"legacy-key"}}`, `{"selected":{"type":"oauth","token":"legacy-key"}}`} {
		if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		key, err := ProviderToken(context.Background(), "selected")
		if err != nil || key != "legacy-key" {
			t.Fatal(key, err)
		}
	}
}

func TestProviderTokenValidOAuthIsReadOnly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	p := &fixtureTokenProvider{}
	oauth.RegisterProvider(p)
	body := []byte(fmt.Sprintf(`{"mcp-oauth-fixture":{"type":"oauth","access":"valid-token","expires":%d,"projectId":"project-fixture"}}`, time.Now().Add(time.Hour).UnixMilli()))
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(path)
	if err := os.Mkdir(filepath.Join(dir, "auth.json.lock"), 0700); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join(dir, "auth.json.lock"))
	key, err := ProviderToken(context.Background(), p.ID())
	if err != nil || key != "valid-token" {
		t.Fatal(key, err)
	}
	after, _ := os.Stat(path)
	data, _ := os.ReadFile(path)
	if p.refreshes != 0 || string(body) != string(data) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("valid token mutated credentials")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gi-auth.lock")); !os.IsNotExist(err) {
		t.Fatal("write lock created on read path", err)
	}
}

func TestProviderTokenConcurrentRefresh(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", dir)
	t.Setenv("PI_CODING_AGENT_DIR", dir)
	p := &fixtureTokenProvider{}
	oauth.RegisterProvider(p)
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"mcp-oauth-fixture":{"type":"oauth","access":"expired","refresh":"refresh","expires":1,"projectId":"project-fixture"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key, err := ProviderToken(context.Background(), p.ID())
			if err != nil || key != "fresh-access" {
				t.Error(key, err)
			}
		}()
	}
	wg.Wait()
	if p.refreshes != 1 {
		t.Fatal("duplicate refreshes", p.refreshes)
	}
}
