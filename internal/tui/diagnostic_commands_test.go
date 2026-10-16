package tui

import (
	"strings"
	"testing"
)

func TestOfflineDiagnosticCommands(t *testing.T) {
	c := durableTestChat(t)
	c.cfg.WorkspaceRoot = "/private/do-not-show"
	c.cfg.DefaultModel = "secret-model-marker"
	c.cfg.DefaultProvider = "secret-provider-marker"
	text := strings.Join(c.bugPreviewLines(nil), "\n")
	for _, secret := range []string{c.cfg.WorkspaceRoot, c.cfg.DefaultModel, c.cfg.DefaultProvider, c.sessionID} {
		if secret != "" && strings.Contains(text, secret) {
			t.Fatalf("diagnostic leaks %q", secret)
		}
	}
	if !strings.Contains(text, "local preview only") || len(text) > diagnosticLimit {
		t.Fatal(text)
	}
	for _, args := range [][]string{{"save"}, {"upload"}, {"preview", "extra"}} {
		if !strings.Contains(strings.Join(c.bugPreviewLines(args), " "), "usage:") {
			t.Fatal(args)
		}
	}
	notes := strings.Join(changelogLines(nil), "\n")
	if !strings.Contains(notes, "# Unreleased") || !strings.Contains(notes, "do not announce a release") {
		t.Fatal(notes)
	}
	for _, name := range []string{"../../etc/passwd", "v0.1.1", "https://example.org"} {
		if !strings.Contains(strings.Join(changelogLines([]string{name}), " "), "no bundled notes") {
			t.Fatal(name)
		}
	}
	c.input.SetText("private unsent draft")
	for _, command := range []string{"/bug", "/bug save", "/changelog", "/changelog ../../etc/passwd", "/trust"} {
		c.onSubmit(command)
		if c.input.Text() != "private unsent draft" {
			t.Fatalf("%s changed draft: %q", command, c.input.Text())
		}
	}
	if !strings.Contains(strings.Join(c.transcript, "\n"), "project MCP configuration is ignored, not trusted") {
		t.Fatal("trust limitation not explicit")
	}
}
