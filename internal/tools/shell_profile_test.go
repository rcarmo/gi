package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestExecuteShellDoesNotSourceLoginProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".profile"), []byte("printf PROFILE-POLLUTION\\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	output, err := ExecuteShell(context.Background(), t.TempDir(), goai.ToolCall{
		Arguments: map[string]any{"command": "printf hello"},
	})
	if err != nil {
		t.Fatalf("ExecuteShell: %v", err)
	}
	if strings.TrimSpace(output) != "hello" {
		t.Fatalf("shell output must not include login-profile output: %q", output)
	}
}
