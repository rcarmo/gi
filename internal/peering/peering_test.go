package peering

import (
	"testing"

	"github.com/rcarmo/gi/internal/config"
)

func TestManagerStatusDisabledByDefault(t *testing.T) {
	m := NewManager(config.PeeringSettings{}, t.TempDir())
	st := m.Status()
	if st.Enabled || st.Backend != "tsnet" || st.State != "disabled" || st.StateDir == "" {
		t.Fatalf("unexpected status: %#v", st)
	}
}

func TestManagerReportsKeychainPlaceholder(t *testing.T) {
	m := NewManager(config.PeeringSettings{Enabled: true, AuthKeyKeychain: "tailscale/authkey"}, t.TempDir())
	if err := m.Start(t.Context()); err == nil {
		t.Fatal("expected keychain placeholder error")
	}
	st := m.Status()
	if st.State != "needs_keychain" || st.Error == "" {
		t.Fatalf("unexpected status: %#v", st)
	}
}
