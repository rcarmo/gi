package inference

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func credentialFixture(t *testing.T, raw string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".pi", "agent")
	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, "auth.json")
	if raw != "" {
		if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}
func TestProviderKeysPrivateAtomicPreservationAndNativeConsumption(t *testing.T) {
	path := credentialFixture(t, `{"other":{"type":"oauth","access":"fixture-other-secret","refresh":"fixture-refresh","extra":{"large":12345678901234567890}},"anthropic":{"type":"api_key","apiKey":"fixture-anthropic","extra":"preserve"}}`)
	before, err := ReadProviderSettings()
	if err != nil {
		t.Fatal(err)
	}
	if err = SaveProviderKey("openai", before.Revision, "fixture-openai-key"); err != nil {
		t.Fatal(err)
	}
	key, _, err := loadAuth("openai")
	if err != nil || key != "fixture-openai-key" {
		t.Fatal("native key consumption failed")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
	snap, err := ReadProviderSettings()
	if err != nil {
		t.Fatal(err)
	}
	blob, _ := json.Marshal(snap)
	for _, secret := range []string{"fixture-openai-key", "fixture-other-secret", "fixture-refresh", "fixture-anthropic", "12345678901234567890"} {
		if bytes.Contains(blob, []byte(secret)) {
			t.Fatal("credential metadata leak")
		}
	}
	if snap.Revision == before.Revision {
		t.Fatal("revision unchanged")
	}
	if err = SaveProviderKey("openai", before.Revision, "overwrite"); !errors.Is(err, ErrCredentialConflict) {
		t.Fatalf("stale save %v", err)
	}
	if err = RemoveProviderKey("anthropic", before.Revision); !errors.Is(err, ErrCredentialConflict) {
		t.Fatalf("stale remove %v", err)
	}
	if err = SaveProviderKey("anthropic", snap.Revision, "fixture-new"); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	var doc map[string]json.RawMessage
	json.Unmarshal(raw, &doc)
	if !bytes.Contains(doc["anthropic"], []byte("preserve")) || !bytes.Contains(doc["other"], []byte("12345678901234567890")) {
		t.Fatal("unknown fields lost")
	}
	// Native TUI logout shares the writer and preserves unknown entries too.
	removed, err := RemoveAuthEntry("openai")
	if err != nil || !removed {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(path)
	if !bytes.Contains(raw, []byte("fixture-other-secret")) {
		t.Fatal("logout lost other credential")
	}
	latest, _ := ReadProviderSettings()
	if err = RemoveProviderKey("anthropic", latest.Revision); err != nil {
		t.Fatal(err)
	}
	if _, _, err = loadAuth("anthropic"); err == nil {
		t.Fatal("removed credential still usable")
	}
}
func TestProviderKeyValidationAndOAuthProtection(t *testing.T) {
	path := credentialFixture(t, `{"openai":{"type":"oauth","access":"fixture-token"},"anthropic":{"type":"api_key","apiKey":"fixture","token":"override"}}`)
	snap, _ := ReadProviderSettings()
	before, _ := os.ReadFile(path)
	for _, provider := range []string{"openai", "anthropic", "unknown", "github-copilot"} {
		if err := SaveProviderKey(provider, snap.Revision, "value"); !errors.Is(err, ErrCredentialInvalid) {
			t.Fatal("protected entry accepted")
		}
		if err := RemoveProviderKey(provider, snap.Revision); !errors.Is(err, ErrCredentialInvalid) {
			t.Fatal("protected entry removed")
		}
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("protected file changed")
	}
	os.WriteFile(path, []byte(`{}`), 0600)
	snap, _ = ReadProviderSettings()
	for _, key := range []string{"", "white space", "new\nline", strings.Repeat("x", 4097), "\xff"} {
		if err := SaveProviderKey("openai", snap.Revision, key); !errors.Is(err, ErrCredentialInvalid) {
			t.Fatal("invalid key accepted")
		}
	}
}
func TestProviderCredentialUnsafeFileFailureAndConflict(t *testing.T) {
	for _, raw := range []string{"broken", "null", "[]", `{"openai":null}`, strings.Repeat(" ", credentialMaxBytes+1)} {
		t.Run(raw[:min(20, len(raw))], func(t *testing.T) {
			path := credentialFixture(t, raw)
			if _, err := ReadProviderSettings(); err == nil {
				t.Fatal("bad file accepted")
			}
			if err := SaveProviderKey("openai", "revision", "fixture"); err == nil {
				t.Fatal("bad file replaced")
			}
			after, _ := os.ReadFile(path)
			if string(after) != raw {
				t.Fatal("failure modified file")
			}
		})
	}
	path := credentialFixture(t, `{}`)
	snap, _ := ReadProviderSettings()
	os.Mkdir(filepath.Join(filepath.Dir(path), ".gi-auth.lock"), 0700)
	if err := SaveProviderKey("openai", snap.Revision, "fixture"); err == nil {
		t.Fatal("nonregular lock accepted")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "{}" {
		t.Fatal("failed write replaced file")
	}
	os.Remove(path)
	external := filepath.Join(t.TempDir(), "external.json")
	os.WriteFile(external, []byte(`{}`), 0600)
	if err := os.Symlink(external, path); err != nil {
		t.Skip(err)
	}
	if _, err := ReadProviderSettings(); err == nil {
		t.Fatal("symlink read")
	}
	if _, err := RemoveAuthEntry("openai"); err == nil {
		t.Fatal("logout followed symlink")
	}
}
func TestProviderCredentialConcurrentRevisions(t *testing.T) {
	credentialFixture(t, `{}`)
	snap, _ := ReadProviderSettings()
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, provider := range []string{"openai", "anthropic"} {
		wg.Add(1)
		go func() { defer wg.Done(); results <- SaveProviderKey(provider, snap.Revision, "fixture-key") }()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrCredentialConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}
