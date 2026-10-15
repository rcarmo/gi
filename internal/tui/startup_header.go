package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/version"
)

// gi's startup header, adapted from Pi's BuiltInHeader and loaded-resource
// listing (pi-coding-agent interactive-mode). Pi shows its logo, version
// and key hints, then the loaded context files, skills, prompts and
// extensions; Ctrl+O (app.tools.expand) expands both. gi shows its own
// name, version and keys, and the resources gi loads: AGENTS.md, skills,
// extensions and MCP servers. quietStartup works as in Pi: true hides
// everything, "header" keeps the header and hides the resource listing.

const (
	startupHeaderKey  = "startup:header"
	startupHeaderKind = "startup"
)

// giLogoBlue is the blue of gi's gopher avatar (fixtures-vibes ui/classic static/icon-*.png).
// The four-cell, two-row half-block logo spells G/i in the avatar's blue
// and the theme's text colour. Apple Terminal uses the text wordmark, as Pi does.
var giLogoBlue = piRGB(64, 128, 192)

func supportsGiBlockLogo(goos, termProgram string) bool {
	return goos != "darwin" || termProgram != "Apple_Terminal"
}

func (c *chatTUI) startupHeaderShown() bool {
	return c.startupHeader && c.cfg.QuietStartup != "true"
}

func (c *chatTUI) startupDetailsShown() bool {
	return c.cfg.QuietStartup == ""
}

// startupHeaderMarker is the virtual first transcript line that renders the
// header (never stored in the transcript).
func startupHeaderMarker() string {
	return encodeTranscriptBlockMarker(transcriptBlockMeta{Key: startupHeaderKey, Kind: startupHeaderKind, Title: "startup"})
}

type startupSection struct {
	name      string
	collapsed []string // one compact line
	expanded  []string // one line per item
}

// startupSections lists the resources gi loaded, like Pi's loaded-resource
// listing (collapsed: one compact line; expanded: one line per item).
func (c *chatTUI) startupSections() []startupSection {
	var out []startupSection
	if root := strings.TrimSpace(c.cfg.WorkspaceRoot); root != "" {
		path := filepath.Join(root, "AGENTS.md")
		if _, err := os.Stat(path); err == nil {
			out = append(out, startupSection{name: "Context", collapsed: []string{"AGENTS.md"}, expanded: []string{displayPath(path)}})
		}
	}
	if skills := c.cfg.Discovery.Skills; len(skills) > 0 {
		names := make([]string, 0, len(skills))
		var paths []string
		for _, s := range skills {
			names = append(names, s.Name)
			paths = append(paths, displayPath(s.Path))
		}
		sort.Strings(names)
		sort.Strings(paths)
		out = append(out, startupSection{name: "Skills", collapsed: []string{strings.Join(names, ", ")}, expanded: paths})
	}
	if c.engine != nil {
		if exts := c.engine.ExtensionInfos(); len(exts) > 0 {
			var names, paths []string
			for _, e := range exts {
				names = append(names, filepath.Base(e.Path))
				paths = append(paths, displayPath(e.Path))
			}
			sort.Strings(names)
			sort.Strings(paths)
			out = append(out, startupSection{name: "Extensions", collapsed: []string{strings.Join(names, ", ")}, expanded: paths})
		}
		if statuses, _ := c.engine.MCPStatus(); len(statuses) > 0 {
			var names, lines []string
			for _, st := range statuses {
				names = append(names, st.Name)
				lines = append(lines, fmt.Sprintf("%s (%s, %s)", st.Name, st.Exposure, displayPath(st.Source)))
			}
			out = append(out, startupSection{name: "MCP", collapsed: []string{strings.Join(names, ", ")}, expanded: lines})
		}
	}
	return out
}

// displayPath shortens the home directory to ~ (Pi's formatDisplayPath).
func displayPath(path string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" && strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

// startupSignature changes whenever the header's content would.
func (c *chatTUI) startupSignature() []string {
	sig := []string{version.String(), c.cfg.QuietStartup, strings.Join(c.cfg.EnabledModels, ","), os.Getenv("TERM_PROGRAM")}
	if c.startupDetailsShown() {
		for _, s := range c.startupSections() {
			sig = append(sig, s.name+"="+strings.Join(s.expanded, "|"))
		}
	}
	return sig
}

func keyHintSpans(key, description string) []gotui.TextSpan {
	return []gotui.TextSpan{{Text: key, Style: piFg(piDim)}, {Text: " " + description, Style: piFg(piMuted)}}
}

// renderStartupHeader renders gi's header (and, unless quietStartup is
// "header", the loaded resources), collapsed or expanded like Pi's.
func (c *chatTUI) renderStartupHeader(expanded bool) *gotui.Element {
	root := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100))
	indented := func(left int, spans ...gotui.TextSpan) {
		root.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithPaddingTRBL(0, 1, 0, left), gotui.WithRichText(spans...)))
	}
	// Pi's header has one column of padding; its resource sections none,
	// with items indented two columns.
	line := func(spans ...gotui.TextSpan) { indented(1, spans...) }
	blank := func() { root.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1))) }
	blank()
	if supportsGiBlockLogo(runtime.GOOS, os.Getenv("TERM_PROGRAM")) {
		// Four square-pixel rows packed into two terminal rows, like Pi's logo.
		line(gotui.TextSpan{Text: "█▀▀", Style: piFg(giLogoBlue)}, gotui.TextSpan{Text: "▀", Style: piFg(piText)},
			gotui.TextSpan{Text: " gi " + version.String(), Style: piFg(piDim)})
		line(gotui.TextSpan{Text: "█▄█", Style: piFg(giLogoBlue)}, gotui.TextSpan{Text: "█", Style: piFg(piText)})
	} else {
		line(gotui.TextSpan{Text: "g", Style: piFg(giLogoBlue).Bold()}, gotui.TextSpan{Text: "i", Style: piFg(piText).Bold()},
			gotui.TextSpan{Text: " " + version.String(), Style: piFg(piDim)})
	}
	details := c.startupDetailsShown()
	if expanded {
		// Pi's expandedInstructions: Pi's keyText for this platform.
		keys := defaultPiKeys
		suspend := "Ctrl+Z"
		if keys.windows {
			suspend = ""
		}
		_, cycleBack := keys.cycleModelBackward()
		_, followUp := keys.followUp()
		_, dequeue := keys.dequeue()
		_, paste := keys.pasteImage()
		for _, h := range [][2]string{
			{keys.keyText("Escape"), "to interrupt"},
			{keys.keyText("Ctrl+C"), "to clear"},
			{keys.keyText("Ctrl+C") + " twice", "to exit"},
			{keys.keyText("Ctrl+D"), "to exit (empty)"},
			{keys.keyText(suspend), "to suspend"},
			{keys.keyText("Ctrl+K"), "to delete to end"},
			{keys.keyText("Shift+Tab"), "to cycle thinking level"},
			{keys.keyText("Ctrl+P/" + cycleBack), "to cycle models"},
			{keys.keyText("Ctrl+L"), "to select model"},
			{keys.keyText("Ctrl+O"), "to expand tools"},
			{keys.keyText("Ctrl+T"), "to expand thinking"},
			{keys.keyText("Ctrl+G"), "for external editor"},
			{"/", "for commands"},
			{"!", "to run bash"},
			{"!!", "to run bash (no context)"},
			{keys.keyText(followUp), "to queue follow-up"},
			{keys.keyText(dequeue), "to edit all queued messages"},
			{keys.keyText(paste), "to paste files on macOS, images, or text"},
			{"drop files", "to attach"},
		} {
			line(keyHintSpans(h[0], h[1])...)
		}
	} else {
		sep := gotui.TextSpan{Text: " · ", Style: piFg(piMuted)}
		var spans []gotui.TextSpan
		for i, h := range [][2]string{{"escape", "interrupt"}, {"ctrl+c/ctrl+d", "clear/exit"}, {"/", "commands"}, {"!", "bash"}, {"ctrl+o", "more"}} {
			if i > 0 {
				spans = append(spans, sep)
			}
			spans = append(spans, keyHintSpans(h[0], h[1])...)
		}
		line(spans...)
		press := "Press ctrl+o to show full startup help"
		if details {
			press += " and loaded resources"
		}
		line(gotui.TextSpan{Text: press + ".", Style: piFg(piDim)})
	}
	blank()
	line(gotui.TextSpan{Text: "gi reads Pi's settings, models and credentials. /help lists its commands and /hotkeys its keys.", Style: piFg(piDim)})
	if details {
		if models := c.cfg.EnabledModels; len(models) > 0 {
			blank()
			line(gotui.TextSpan{Text: "Model scope: " + strings.Join(models, ", "), Style: piFg(piDim)},
				gotui.TextSpan{Text: " (Ctrl+P to cycle)", Style: piFg(piMuted)})
		}
		for _, s := range c.startupSections() {
			blank()
			indented(0, gotui.TextSpan{Text: "[" + s.name + "]", Style: piFg(piMdHeading)})
			items := s.collapsed
			if expanded {
				items = s.expanded
			}
			for _, item := range items {
				indented(2, gotui.TextSpan{Text: item, Style: piFg(piDim)})
			}
		}
	}
	blank()
	return root
}
