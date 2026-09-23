package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func validPolicy() CompactionPolicyEdit { return CompactionPolicyEdit{true, 64000, 4000, 8000, 50000} }
func settingsFixture(t *testing.T, raw string) (string, string) {
	t.Helper()
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".pi"), 0700)
	path := filepath.Join(root, ".pi", "settings.json")
	if err := os.WriteFile(path, []byte(raw), 0640); err != nil {
		t.Fatal(err)
	}
	return root, path
}

func TestCompactionPolicyPersistenceDefaultsAndWriters(t *testing.T) {
	for _, raw := range []string{"", `{}`, `{"compaction":{"enabled":true,"threshold_tokens":90000}}`} {
		root := t.TempDir()
		if raw != "" {
			os.Mkdir(filepath.Join(root, ".pi"), 0700)
			os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(raw), 0600)
		}
		snap, err := ReadCompactionPolicy(root)
		if err != nil {
			t.Fatal(err)
		}
		if snap.Policy != Load(root).Compaction {
			t.Fatalf("default mismatch %q: %+v != %+v", raw, snap.Policy, Load(root).Compaction)
		}
	}
	root, path := settingsFixture(t, `{"defaultModel":"test","hooks":{"timeout_ms":1234},"huge":12345678901234567890,"compaction":{"enabled":false,"strategy":"custom-label","extra":{"keep":true}}}`)
	original := Load(root).Compaction
	snap, err := ReadCompactionPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := SaveCompactionPolicy(root, snap.Revision, validPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if saved.Policy.Strategy != "custom-label" || saved.Policy.ThresholdTokens != 50000 || Load(root).Compaction != saved.Policy || original == saved.Policy {
		t.Fatalf("saved %+v", saved)
	}
	calls := []func() error{
		func() error { return PersistModelSelection(root, "test", "bootstrap", "high", []string{"test-model"}) },
		func() error { return PersistClipboardMode(root, "native") }, func() error { return PersistScrollbackLimit(root, 231) },
		func() error { return PersistTUIHistoryLimit(root, 444) }, func() error { return PersistTUIScrollbar(root, true) },
	}
	var wg sync.WaitGroup
	for _, call := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := call(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	cfg := Load(root)
	if cfg.Compaction != saved.Policy || cfg.TUIClipboardMode != "native" || cfg.ScrollbackLimit != 231 || cfg.TUIHistoryLimit != 444 || !cfg.TUIScrollbar || cfg.DefaultModel != "bootstrap" {
		t.Fatalf("writer lost fields: %+v", cfg)
	}
	if _, err = SaveCompactionPolicy(root, saved.Revision, validPolicy()); !errors.Is(err, ErrSettingsConflict) {
		t.Fatalf("stale policy accepted: %v", err)
	}
	raw, _ := os.ReadFile(path)
	var doc map[string]json.RawMessage
	json.Unmarshal(raw, &doc)
	if string(doc["huge"]) != "12345678901234567890" || !strings.Contains(string(doc["compaction"]), `"extra"`) {
		t.Fatalf("unknown lost %s", raw)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0640 {
		t.Fatal("mode changed")
	}
}

func TestCompactionPolicyRejectsInvalidUnsafeAndFailedWrites(t *testing.T) {
	root, path := settingsFixture(t, `{"compaction":{"enabled":true}}`)
	snap, _ := ReadCompactionPolicy(root)
	before, _ := os.ReadFile(path)
	for _, patch := range []func(*CompactionPolicyEdit){func(p *CompactionPolicyEdit) { p.ContextWindow = 0 }, func(p *CompactionPolicyEdit) { p.ContextWindow = MaxCompactionTokens + 1 }, func(p *CompactionPolicyEdit) { p.ReserveTokens = p.ContextWindow }, func(p *CompactionPolicyEdit) { p.KeepRecentTokens = p.ThresholdTokens + 1 }, func(p *CompactionPolicyEdit) { p.ThresholdTokens = p.ContextWindow }, func(p *CompactionPolicyEdit) { p.KeepRecentTokens = -1 }} {
		p := validPolicy()
		patch(&p)
		if _, err := SaveCompactionPolicy(root, snap.Revision, p); !errors.Is(err, ErrCompactionPolicyInvalid) {
			t.Fatal(err)
		}
	}
	os.Mkdir(filepath.Join(root, ".pi", ".gi-settings.lock"), 0700)
	if _, err := SaveCompactionPolicy(root, snap.Revision, validPolicy()); err == nil {
		t.Fatal("lock dir accepted")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("failure changed settings")
	}
	for _, raw := range []string{`bad`, `null`, `[]`, `{"compaction":null}`, `{"compaction":{"enabled":"yes"}}`, strings.Repeat(" ", settingsMaxBytes+1)} {
		t.Run(raw[:min(32, len(raw))], func(t *testing.T) {
			root, path := settingsFixture(t, raw)
			if _, err := ReadCompactionPolicy(root); err == nil {
				t.Fatal("bad read")
			}
			if _, err := SaveCompactionPolicy(root, "missing", validPolicy()); err == nil {
				t.Fatal("bad write")
			}
			got, _ := os.ReadFile(path)
			if string(got) != raw {
				t.Fatal("bad config replaced")
			}
		})
	}
	other, path := settingsFixture(t, `{}`)
	target := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(target, []byte(`{}`), 0600)
	os.Remove(path)
	if err := os.Symlink(target, path); err != nil {
		t.Skip(err)
	}
	if _, err := ReadCompactionPolicy(other); err == nil {
		t.Fatal("symlink read")
	}
	if err := PersistClipboardMode(other, "native"); err == nil {
		t.Fatal("legacy writer followed symlink")
	}
}

func TestCompactionPolicyConcurrentMergeCannotLoseModelPreference(t *testing.T) {
	for i := 0; i < 20; i++ {
		root, _ := settingsFixture(t, `{"compaction":{"enabled":true},"keep":{"value":7}}`)
		before, err := ReadCompactionPolicy(root)
		if err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			<-start
			done <- PersistModelSelection(root, "test", "bootstrap", "high", []string{"test-model"})
		}()
		close(start)
		_, policyErr := SaveCompactionPolicy(root, before.Revision, validPolicy())
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		if errors.Is(policyErr, ErrSettingsConflict) {
			current, err := ReadCompactionPolicy(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = SaveCompactionPolicy(root, current.Revision, validPolicy()); err != nil {
				t.Fatal(err)
			}
		} else if policyErr != nil {
			t.Fatal(policyErr)
		}
		cfg := Load(root)
		if cfg.DefaultModel != "bootstrap" || cfg.DefaultThinkingLevel != "high" || cfg.Compaction.ThresholdTokens != 50000 {
			t.Fatalf("lost concurrent write %+v", cfg)
		}
	}
}
