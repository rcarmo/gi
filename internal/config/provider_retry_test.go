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

func TestProviderRetryFileBackedGenerationRefresh(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
	path := filepath.Join(root, ".pi", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"retry":{"maxRetries":1,"baseDelayMs":17}}`)
	original := Load(root)
	write(`{"retry":{"maxRetries":4,"baseDelayMs":0},"transport":"sse"}`)
	fresh, err := original.ForGeneration()
	if err != nil {
		t.Fatal(err)
	}
	if p := fresh.Retry.Policy(); p.MaxRetries != 4 || p.BaseDelayMS != 0 || fresh.Transport != "sse" {
		t.Fatal(p, fresh.Transport)
	}
	if original.Retry.Policy().MaxRetries != 1 {
		t.Fatal("refresh mutated original")
	}
	write(`{}`)
	fresh, err = original.ForGeneration()
	if err != nil {
		t.Fatal(err)
	}
	if p := fresh.Retry.Policy(); p.MaxRetries != 3 || p.BaseDelayMS != 2000 {
		t.Fatal("removed settings retained stale policy", p)
	}
	disabled := false
	explicit := RuntimeConfig{WorkspaceRoot: root, Retry: ProviderRetrySettings{Enabled: &disabled}}
	fresh, err = explicit.ForGeneration()
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Retry.Policy().Enabled {
		t.Fatal("explicit in-memory configuration discarded")
	}
	write(`{"retry":`)
	if _, err = original.ForGeneration(); err == nil {
		t.Fatal("malformed current policy silently enabled retries")
	}
}

func TestProviderRetryGenerationPolicyGlobalProjectLayers(t *testing.T) {
	root := t.TempDir()
	global := filepath.Join(root, "agent")
	t.Setenv("GI_CODING_AGENT_DIR", global)
	if err := os.MkdirAll(global, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(global, "settings.json"), []byte(`{"retry":{"enabled":false,"maxRetries":7,"baseDelayMs":17}}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".pi", "settings.json")
	for _, body := range []string{`{}`, `{"retry":{"enabled":true,"maxRetries":4,"baseDelayMs":0,"maxAgentDelayMs":0}}`} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		loaded := Load(root)
		fresh, err := loaded.ForGeneration()
		if err != nil {
			t.Fatal(err)
		}
		policy := fresh.Retry.Policy()
		if policy != loaded.Retry.Policy() {
			t.Fatal(policy, loaded.Retry.Policy())
		}
		if body != `{}` && (!policy.Enabled || policy.MaxRetries != 4 || policy.BaseDelayMS != 0 || policy.MaxDelayMS != 0) {
			t.Fatal("project overrides lost", policy)
		}
		if body == `{}` && (policy.Enabled || policy.MaxRetries != 7 || policy.BaseDelayMS != 17) {
			t.Fatal("global retry settings lost", policy)
		}
	}
	if err := os.WriteFile(path, []byte(`{"retry":`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root).ForGeneration(); err == nil {
		t.Fatal("malformed retry settings ignored")
	}
	no := false
	explicit := RuntimeConfig{WorkspaceRoot: root, Retry: ProviderRetrySettings{Enabled: &no}}
	if refreshed, err := explicit.ForGeneration(); err != nil || refreshed.Retry.Policy().Enabled {
		t.Fatal("explicit configuration discarded", refreshed.Retry.Policy(), err)
	}
}
