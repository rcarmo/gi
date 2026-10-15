package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAgentsInTheCloudContextOrderAndSymlinkDeduplication(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "empty-gi"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(root, "empty-pi"))
	files := []struct{ path, text string }{{"AGENTS.md", "root instructions"}, {".agents-in-the-cloud/AGENTS.md", "root cloud instructions"}, {"repo/AGENTS.md", "repo instructions"}, {"repo/.agents-in-the-cloud/AGENTS.md", "repo cloud instructions"}}
	for _, f := range files {
		p := filepath.Join(root, f.path)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(f.text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var got []ContextFile
	for _, f := range LoadContextFiles(filepath.Join(root, "repo")) {
		if len(f.Path) > len(root) && f.Path[:len(root)] == root {
			got = append(got, f)
		}
	}
	if len(got) != len(files) {
		t.Fatalf("contexts=%v", got)
	}
	for i, want := range files {
		if got[i].Path != filepath.Join(root, want.path) || got[i].Content != want.text {
			t.Fatalf("context[%d]=%+v", i, got[i])
		}
	}
	// The compatibility link must not inject the same instructions twice.
	cloud := filepath.Join(root, "repo/.agents-in-the-cloud/AGENTS.md")
	if err := os.Remove(cloud); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "repo/AGENTS.md"), cloud); err != nil {
		t.Fatal(err)
	}
	got = nil
	for _, f := range LoadContextFiles(filepath.Join(root, "repo")) {
		if len(f.Path) > len(root) && f.Path[:len(root)] == root {
			got = append(got, f)
		}
	}
	if len(got) != 3 {
		t.Fatalf("linked instructions duplicated: %v", got)
	}
}
