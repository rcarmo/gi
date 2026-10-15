package tui

import (
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/version"
)

func headerText(t *testing.T, c *chatTUI, width int) string {
	t.Helper()
	blocks := c.buildTranscriptRenderableBlocks(c.visibleTranscript())
	if len(blocks) == 0 || blocks[0].Kind != startupHeaderKind {
		t.Fatalf("first block %+v", blocks)
	}
	rows := c.transcriptRowsAtWidth(width)
	var out []string
	for _, r := range rows {
		out = append(out, strings.TrimRight(r.text, " "))
	}
	return strings.Join(out, "\n")
}

func TestStartupHeaderBlockLogoAndFallback(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "ghostty")
	c := &chatTUI{startupHeader: true}
	text := headerText(t, c, 100)
	for _, want := range []string{" █▀▀▀ gi " + version.String(), "\n █▄██\n"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing block logo %q:\n%s", want, text)
		}
	}
	if supportsGiBlockLogo("darwin", "Apple_Terminal") || !supportsGiBlockLogo("darwin", "iTerm.app") || !supportsGiBlockLogo("linux", "Apple_Terminal") {
		t.Fatal("Apple Terminal fallback differs from Pi")
	}
	for _, width := range []int{8, 20, 80} {
		for _, row := range c.transcriptRowsAtWidth(width) {
			if len([]rune(row.text)) > width {
				t.Fatalf("header spills at width %d: %q", width, row.text)
			}
		}
	}
}

// gi's startup header: name, version and gi's keys collapsed; the full key
// list and loaded resources on Ctrl+O; quietStartup like Pi.
func TestStartupHeader(t *testing.T) {
	defer func(keys piKeys) { defaultPiKeys = keys }(defaultPiKeys)
	defaultPiKeys = piKeys{} // Pi's Linux keys
	c := &chatTUI{startupHeader: true, transcriptExpanded: map[string]bool{},
		cfg: config.RuntimeConfig{AssistantName: "Gi", EnabledModels: []string{"a/one", "b/two"},
			Discovery: skills.Discovery{Skills: []skills.Skill{{Name: "zeta", Path: "/s/zeta/SKILL.md"}, {Name: "alpha", Path: "/s/alpha/SKILL.md"}}}}}
	text := headerText(t, c, 100)
	for _, want := range []string{
		"gi " + version.String(),
		"escape interrupt · ctrl+c/ctrl+d clear/exit · / commands · ! bash · ctrl+o more",
		"Press ctrl+o to show full startup help and loaded resources.",
		"Model scope: a/one, b/two (Ctrl+P to cycle)",
		"[Skills]\n  alpha, zeta",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("collapsed header lacks %q:\n%s", want, text)
		}
	}
	c.toggleToolOutput()
	text = headerText(t, c, 100)
	for _, want := range []string{"escape to interrupt", "ctrl+c twice to exit", "shift+tab to cycle thinking level", "ctrl+p/shift+ctrl+p to cycle models", "ctrl+o to expand tools", "!! to run bash (no context)", "drop files to attach", "[Skills]\n  /s/alpha/SKILL.md\n  /s/zeta/SKILL.md"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expanded header lacks %q:\n%s", want, text)
		}
	}
	c.toggleToolOutput()
	c.cfg.QuietStartup = "header"
	text = headerText(t, c, 100)
	if strings.Contains(text, "[Skills]") || strings.Contains(text, "Model scope") || !strings.Contains(text, "Press ctrl+o to show full startup help.") {
		t.Fatalf("quietStartup header:\n%s", text)
	}
	c.cfg.QuietStartup = "true"
	if blocks := c.buildTranscriptRenderableBlocks(c.visibleTranscript()); len(blocks) > 0 && blocks[0].Kind == startupHeaderKind {
		t.Fatal("quietStartup true still shows the header")
	}
}
