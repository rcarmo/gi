package peering

import (
	"context"
	"errors"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/secrets"
)

func TestManagerStatusDisabledByDefault(t *testing.T) {
	m := NewManager(config.PeeringSettings{}, t.TempDir())
	st := m.Status()
	if st.Enabled || st.Backend != "tsnet" || st.State != "disabled" || st.StateDir == "" {
		t.Fatalf("unexpected status: %#v", st)
	}
}

func TestManagerResolvesKeychainAuthKey(t *testing.T) {
	resolver := secrets.ResolverFunc(func(_ context.Context, name string) (string, error) {
		if name != "tailscale/authkey" {
			t.Fatalf("unexpected name: %s", name)
		}
		return "tskey-auth-test", nil
	})
	m := NewManagerWithResolver(config.PeeringSettings{Enabled: true, AuthKeyKeychain: "tailscale/authkey"}, t.TempDir(), resolver)
	got, err := m.resolveAuthKey(t.Context())
	if err != nil || got != "tskey-auth-test" {
		t.Fatalf("resolveAuthKey = %q, %v", got, err)
	}
}

func TestManagerReportsUnavailableKeychain(t *testing.T) {
	resolver := secrets.ResolverFunc(func(context.Context, string) (string, error) { return "", errors.New("not found") })
	m := NewManagerWithResolver(config.PeeringSettings{Enabled: true, AuthKeyKeychain: "tailscale/authkey"}, t.TempDir(), resolver)
	if _, err := m.resolveAuthKey(t.Context()); err == nil {
		t.Fatal("expected keychain resolution error")
	}
	st := m.Status()
	if st.State != "needs_keychain" || st.Error == "" {
		t.Fatalf("unexpected status: %#v", st)
	}
}

func TestManagerAuthKeyEnvTakesPrecedence(t *testing.T) {
	t.Setenv("TS_AUTHKEY_TEST", "from-env")
	resolver := secrets.ResolverFunc(func(context.Context, string) (string, error) {
		t.Fatal("resolver must not be called when auth_key_env is set")
		return "", nil
	})
	m := NewManagerWithResolver(config.PeeringSettings{Enabled: true, AuthKeyEnv: "TS_AUTHKEY_TEST", AuthKeyKeychain: "tailscale/authkey"}, t.TempDir(), resolver)
	got, err := m.resolveAuthKey(t.Context())
	if err != nil || got != "from-env" {
		t.Fatalf("resolveAuthKey = %q, %v", got, err)
	}
}
