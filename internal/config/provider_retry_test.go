package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProviderRetryPiSettings(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, ".pi"), 0700)
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"retry":{"enabled":false,"maxRetries":0,"baseDelayMs":17,"maxAgentDelayMs":99}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p := Load(root).Retry.Policy()
	if p.Enabled || p.MaxRetries != 0 || p.BaseDelayMS != 17 || p.MaxDelayMS != 99 {
		t.Fatalf("settings ignored: %#v", p)
	}
	d := ProviderRetrySettings{}.Policy()
	if !d.Enabled || d.MaxRetries != 3 || d.BaseDelayMS != 2000 || d.MaxDelayMS != 60000 {
		t.Fatal(d)
	}
}

func TestProviderRetryHonoursZeroWithoutEnablingRetries(t *testing.T) {
	zero := 0
	no := false
	p := ProviderRetrySettings{Enabled: &no, MaxRetries: &zero, BaseDelayMS: &zero, MaxAgentDelayMS: &zero}.Policy()
	if p.Enabled || p.MaxRetries != 0 || p.BaseDelayMS != 0 || p.MaxDelayMS != 0 {
		t.Fatal(p)
	}
}

func TestProviderRetryAcceptsUpToOneHundred(t *testing.T) {
	for _, n := range []int{0, 10, 11, 100, 101} {
		p := ProviderRetrySettings{MaxRetries: &n}.Policy()
		if p.MaxRetries != min(n, 100) {
			t.Fatal(n, p)
		}
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"retry":{"maxRetries":100,"baseDelayMs":0,"maxAgentDelayMs":0}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p := Load(root).Retry.Policy()
	if p.MaxRetries != 100 || p.BaseDelayMS != 0 || p.MaxDelayMS != 0 {
		t.Fatal(p)
	}
}
