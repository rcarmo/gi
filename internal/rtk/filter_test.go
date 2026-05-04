package rtk

import (
	"strings"
	"testing"
)

func TestFilterGitStatus(t *testing.T) {
	out := "On branch main\nChanges not staged:\n modified: a.go\n modified: b.go\nUntracked files:\n\tc.go\n"
	got := Filter("git status", out)
	if got.Mode != "git-status" || !strings.Contains(got.Output, "modified: 2") || !strings.Contains(got.Output, "added/untracked: 1") {
		t.Fatalf("bad filter: %#v", got)
	}
}

func TestFilterGoTestKeepsFailures(t *testing.T) {
	out := "ok pkg/a\n--- FAIL: TestX (0.00s)\n    x_test.go: boom\nFAIL\nok pkg/b\n"
	got := Filter("go test ./...", out)
	if got.Mode != "go-test" || strings.Contains(got.Output, "ok pkg/a") || !strings.Contains(got.Output, "FAIL") {
		t.Fatalf("bad filter: %#v", got)
	}
}

func TestFilterSearchGroupsFiles(t *testing.T) {
	out := "a.go:1:TODO\na.go:2:TODO\nb.go:3:TODO\n"
	got := Filter("rg TODO", out)
	if !strings.Contains(got.Output, "a.go: 2") || !strings.Contains(got.Output, "b.go: 1") {
		t.Fatalf("bad search filter: %s", got.Output)
	}
}
