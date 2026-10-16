package tui

import (
	"strings"
	"testing"
)

func TestCommandPaletteLinesFilterCommands(t *testing.T) {
	c := &chatTUI{}
	all := strings.Join(c.commandPaletteLines(""), "\n")
	for _, want := range []string{"commands: palette", "/session", "/skill:name [args]", "/copy [--osc52|--native|--auto|--fallback]", "!!cmd"} {
		if !strings.Contains(all, want) {
			t.Fatalf("palette missing %q:\n%s", want, all)
		}
	}
	filtered := strings.Join(c.commandPaletteLines("model"), "\n")
	if !strings.Contains(filtered, "/model <provider/model>") || strings.Contains(filtered, "/session") {
		t.Fatalf("filtered palette mismatch:\n%s", filtered)
	}
}

func TestHelpLinesAreGroupedAndDiscoverCoreCommands(t *testing.T) {
	c := &chatTUI{}
	joined := strings.Join(c.helpLines(), "\n")
	for _, want := range []string{
		"help",
		"enter send",
		"/commands",
		"/model",
		"/session",
		"/where",
		"/attach",
		"!cmd",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("help missing %q:\n%s", want, joined)
		}
	}
}
