package rtk

import "testing"

func TestRTKCommandCoverageModes(t *testing.T) {
	covered := map[string]string{
		"ls -la":              "ls",
		"tree .":              "tree",
		"cat README.md":       "generic",
		"head -20 README.md":  "generic",
		"tail -20 log.txt":    "generic",
		"grep -R TODO .":      "grep",
		"rg TODO .":           "rg",
		"find . -name '*.go'": "find",
		"diff a b":            "generic",
		"git status":          "git-status",
		"gh pr list":          "generic",
		"go test ./...":       "go-test",
		"npm test":            "npm-test",
		"pnpm test":           "npm-test",
		"yarn test":           "npm-test",
		"bun test":            "npm-test",
		"jest":                "generic",
		"vitest run":          "generic",
		"pytest":              "pytest",
		"cargo test":          "cargo-test",
		"ruff check .":        "generic",
		"golangci-lint run":   "generic",
		"eslint .":            "generic",
		"biome check .":       "generic",
		"tsc --noEmit":        "generic",
		"docker ps":           "generic",
		"kubectl get pods":    "generic",
		"make test":           "generic",
		"cmake --build build": "generic",
	}
	for command, mode := range covered {
		got := Filter(command, "sample output")
		if got.Mode != mode {
			t.Fatalf("mode mismatch for %q: got=%q want=%q", command, got.Mode, mode)
		}
	}
}

func TestRTKReadStyleCommandsStayGeneric(t *testing.T) {
	if got := Filter("cat README.md", "hello"); got.Mode != "generic" {
		t.Fatalf("cat should remain generic, got %q", got.Mode)
	}
	if got := Filter("read vfs://notes/file.md", "hello"); got.Mode != "generic" {
		t.Fatalf("vfs read should remain generic, got %q", got.Mode)
	}
}
