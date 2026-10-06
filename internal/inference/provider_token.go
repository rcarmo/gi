package inference

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/rcarmo/go-ai/oauth"
)

// ProviderToken resolves current credentials on every call. OAuth refreshes
// cooperate with Pi's auth.json lock and preserve provider-specific fields.
// No caller receives refresh credentials or credential-store error contents.
func ProviderToken(ctx context.Context, provider string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := lockCredentialsContext(ctx); err != nil {
		return "", err
	}
	doc, err := func() (credentialDocument, error) {
		defer credentialMu.Unlock()
		root, e := openCredentialRoot(false)
		if e != nil {
			return credentialDocument{}, e
		}
		defer root.Close()
		return readCredentialDocument(root)
	}()
	if err != nil {
		return "", errors.New("provider credentials unavailable")
	}
	var entry authEntry
	if raw := doc.entries[provider]; raw == nil || json.Unmarshal(raw, &entry) != nil {
		return "", errors.New("provider credentials unavailable")
	}
	if entryAuthType(entry) != "oauth" {
		key := entry.Key
		if key == "" {
			key = entry.APIKey
		}
		return validProviderToken(key)
	}
	p := oauth.GetProvider(provider)
	if p == nil {
		token := entry.Access
		if token == "" {
			token = entry.Token
		}
		if entry.Expires > 0 && entry.Expires <= time.Now().UnixMilli() {
			return "", errors.New("provider credentials expired")
		}
		return validProviderToken(token)
	}
	// Valid OAuth credentials need no write lock or filesystem mutation.
	creds := providerOAuthCredentials(doc.entries[provider], entry)
	if !providerNeedsRefresh(creds) {
		return validProviderToken(p.GetAPIKey(creds))
	}
	var token string
	_, err = updateCredentialsContext(ctx, "", func(raw map[string]json.RawMessage) error {
		fields := map[string]json.RawMessage{}
		if data := raw[provider]; data == nil {
			return errors.New("provider credentials unavailable")
		} else if json.Unmarshal(data, &fields) != nil {
			return errors.New("provider credentials unavailable")
		}
		var current authEntry
		if json.Unmarshal(raw[provider], &current) != nil {
			return errors.New("provider credentials unavailable")
		}
		// Re-read under the lock: logout/rotation must not refresh stale credentials.
		if entryAuthType(current) != "oauth" {
			var e error
			key := current.Key
			if key == "" {
				key = current.APIKey
			}
			token, e = validProviderToken(key)
			if e != nil {
				return e
			}
			return errCredentialUnchanged
		}
		creds := providerOAuthCredentials(raw[provider], current)
		// Providers without a context refresh cannot be allowed to block a cancelled
		// MCP call while holding the shared credential lock.
		if !providerNeedsRefresh(creds) {
			var e error
			token, e = validProviderToken(p.GetAPIKey(creds))
			if e != nil {
				return e
			}
			return errCredentialUnchanged
		}
		if _, ok := p.(oauth.ContextRefreshProvider); !ok {
			return errors.New("provider refresh does not support cancellation")
		}
		updated, key, e := oauth.GetAPIKeyWithContext(ctx, provider, creds)
		if e != nil {
			return errors.New("provider credential refresh failed")
		}
		token, e = validProviderToken(key)
		if e != nil {
			return e
		}
		if e := ctx.Err(); e != nil {
			return e
		}
		set := func(name string, v any) { fields[name], _ = json.Marshal(v) }
		for k, v := range updated.Extra {
			encoded, e := json.Marshal(v)
			if e != nil {
				return errors.New("invalid refreshed provider credentials")
			}
			fields[k] = encoded
		}
		set("type", "oauth")
		set("access", updated.Access)
		set("refresh", updated.Refresh)
		set("expires", updated.Expires)
		raw[provider], e = json.Marshal(fields)
		return e
	})
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", errors.New("provider credential refresh failed")
	}
	return token, nil
}

func providerNeedsRefresh(c *oauth.Credentials) bool {
	return c.Expires > 0 && c.Expires <= time.Now().Add(5*time.Minute).UnixMilli() || c.Access == "" && c.Refresh != ""
}

func providerOAuthCredentials(raw json.RawMessage, entry authEntry) *oauth.Credentials {
	c := authEntryToOAuthCredentials(entry)
	if c.Extra == nil {
		c.Extra = map[string]any{}
	}
	fields := map[string]any{}
	_ = json.Unmarshal(raw, &fields)
	for k, v := range fields {
		switch k {
		case "type", "access", "refresh", "expires":
		default:
			c.Extra[k] = v
		}
	}
	return c
}

func validProviderToken(token string) (string, error) {
	if token == "" || strings.IndexFunc(token, func(r rune) bool { return r <= 32 || r == 127 }) >= 0 {
		return "", errors.New("provider token unavailable or invalid")
	}
	return token, nil
}
