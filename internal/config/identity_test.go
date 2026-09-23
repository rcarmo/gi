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

func identityFixture(t *testing.T, raw string) (string, string) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".piclaw")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(raw), 0640); err != nil {
		t.Fatal(err)
	}
	return root, path
}

func TestIdentitySavePreservesUnknownFieldsAndActivatesOnReload(t *testing.T) {
	root, path := identityFixture(t, `{"assistant":{"assistantName":"Before","assistantAvatar":"avatar.png","unknown":[1,2]},"user":{"userName":"Reader","userAvatar":"photo.jpg"},"secret":"keep-me","large":12345678901234567890}`)
	before := Load(root)
	snap, err := ReadIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := SaveIdentity(root, snap.Revision, IdentityNames{" After ", " User Two "})
	if err != nil {
		t.Fatal(err)
	}
	if saved.AssistantName != "After" || saved.UserName != "User Two" || saved.Revision == snap.Revision {
		t.Fatalf("saved: %+v", saved)
	}
	if before.AssistantName != "Before" {
		t.Fatal("active config mutated")
	}
	after := Load(root)
	if after.AssistantName != "After" || after.UserName != "User Two" || after.AssistantAvatar != "avatar.png" || after.UserAvatar != "photo.jpg" {
		t.Fatalf("reload: %+v", after)
	}
	raw, _ := os.ReadFile(path)
	var doc map[string]json.RawMessage
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if string(doc["secret"]) != `"keep-me"` || string(doc["large"]) != "12345678901234567890" {
		t.Fatalf("unknown keys changed: %s", raw)
	}
	var assistant map[string]json.RawMessage
	_ = json.Unmarshal(doc["assistant"], &assistant)
	var unknown []int
	_ = json.Unmarshal(assistant["unknown"], &unknown)
	if len(unknown) != 2 {
		t.Fatal("nested unknown removed")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0640 {
		t.Fatalf("mode changed: %o", info.Mode().Perm())
	}
	again, err := ReadIdentity(root)
	if err != nil || again != saved {
		t.Fatalf("reopen: %+v %v", again, err)
	}
	if _, err = SaveIdentity(root, snap.Revision, IdentityNames{"Lost", "Update"}); !errors.Is(err, ErrIdentityConflict) {
		t.Fatalf("stale revision accepted: %v", err)
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".piclaw", "*.tmp"))
	if len(matches) != 0 {
		t.Fatalf("temp leaked: %v", matches)
	}
}

func TestIdentityMissingAndConcurrentCAS(t *testing.T) {
	root := t.TempDir()
	snap, err := ReadIdentity(root)
	if err != nil || snap.Revision != "missing" {
		t.Fatalf("missing: %+v %v", snap, err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, name := range []string{"One", "Two"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := SaveIdentity(root, snap.Revision, IdentityNames{name, "User"})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrIdentityConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("results %d %d", success, conflict)
	}
}

func TestIdentityRejectsUnsafeInvalidAndFailedWritesWithoutReplacement(t *testing.T) {
	for _, raw := range []string{`broken`, `null`, `[]`, `{"assistant":null}`, `{"user":42}`, `{"assistant":{"assistantName":9}}`, strings.Repeat(" ", identityMaxBytes+1)} {
		t.Run(raw[:min(len(raw), 40)], func(t *testing.T) {
			root, path := identityFixture(t, raw)
			if _, err := ReadIdentity(root); err == nil {
				t.Fatal("invalid document accepted")
			}
			if _, err := SaveIdentity(root, "missing", IdentityNames{"New", "Name"}); err == nil {
				t.Fatal("invalid config overwritten")
			}
			after, _ := os.ReadFile(path)
			if string(after) != raw {
				t.Fatal("config changed")
			}
		})
	}
	root, path := identityFixture(t, `{"assistant":{"assistantName":"Old"}}`)
	snap, _ := ReadIdentity(root)
	before, _ := os.ReadFile(path)
	for _, names := range []IdentityNames{{"", "User"}, {" \t", "User"}, {strings.Repeat("a", 129), "User"}, {"Valid", "new\nline"}, {"\xff", "User"}} {
		if _, err := SaveIdentity(root, snap.Revision, names); !errors.Is(err, ErrIdentityInvalid) {
			t.Fatalf("invalid names accepted: %v", err)
		}
	}
	// A nonregular lock prevents writing and never removes or replaces config.
	if err := os.Mkdir(filepath.Join(root, ".piclaw", ".gi-identity.lock"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveIdentity(root, snap.Revision, IdentityNames{"New", "Name"}); err == nil {
		t.Fatal("write succeeded with unusable lock")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Fatal("failed write altered config")
	}
}

func TestIdentityRejectsConfigAndDirectorySymlinks(t *testing.T) {
	root, path := identityFixture(t, `{}`)
	outside := filepath.Join(t.TempDir(), "outside.json")
	os.WriteFile(outside, []byte(`{"secret":"outside"}`), 0600)
	os.Remove(path)
	if err := os.Symlink(outside, path); err != nil {
		t.Skip(err)
	}
	if _, err := ReadIdentity(root); err == nil {
		t.Fatal("symlink read")
	}
	if _, err := SaveIdentity(root, "missing", IdentityNames{"New", "User"}); err == nil {
		t.Fatal("symlink write")
	}
	raw, _ := os.ReadFile(outside)
	if string(raw) != `{"secret":"outside"}` {
		t.Fatal("outside changed")
	}
	other := t.TempDir()
	if err := os.Symlink(filepath.Join(root, ".piclaw"), filepath.Join(other, ".piclaw")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadIdentity(other); err == nil {
		t.Fatal("symlink directory read")
	}
	os.Remove(path)
	os.Mkdir(path, 0700)
	if _, err := ReadIdentity(root); err == nil {
		t.Fatal("directory accepted as config")
	}
}
