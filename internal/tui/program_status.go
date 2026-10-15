package tui

import (
	"encoding/base64"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Wire contract matches pi-tui's program-status.ts (503c605528f9).
const programStatusQuery = "\x1b]7501;?\x1b\\"

type programStatus struct{ State, App, Kind, Message string }

var programAppPattern = regexp.MustCompile(`^[A-Za-z0-9_.+-]{1,32}$`)
var programControlPattern = regexp.MustCompile(`[\x00-\x1f\x7f-\x{009f}]+`)
var programReplyPattern = regexp.MustCompile("^\x1b\\]7501;\\?[^\x07\x1b]*(?:\x07|\x1b\\\\)$")
var programDAPattern = regexp.MustCompile("^\x1b\\[\\?[0-9;]*c$")

func formatProgramStatus(s programStatus) string {
	pairs := []string{"state=" + s.State}
	if programAppPattern.MatchString(s.App) {
		pairs = append(pairs, "app="+s.App)
	}
	if s.State == "blocked" && s.Kind != "" {
		pairs = append(pairs, "kind="+s.Kind)
	}
	msg := strings.TrimSpace(programControlPattern.ReplaceAllString(s.Message, " "))
	if len(msg) > 2048 {
		end := 2048
		for !utf8.RuneStart(msg[end]) {
			end--
		}
		msg = msg[:end]
	}
	if msg != "" {
		pairs = append(pairs, "msg="+base64.StdEncoding.EncodeToString([]byte(msg)))
	}
	return "\x1b]7501;" + strings.Join(pairs, ":") + "\x1b\\"
}

// Owned by the UI loop, not background engine goroutines. Reports are retained
// until support is known; DA1 closes the detection window, as in pi-tui.
type programStatusTerminal struct {
	writer             io.Writer
	supported, pending bool
	current            programStatus
	last               string
}

func newProgramStatusTerminal(w io.Writer, override string) *programStatusTerminal {
	return &programStatusTerminal{writer: w, supported: override == "1", pending: override != "0" && override != "1"}
}
func (t *programStatusTerminal) start() {
	if t.pending {
		_, _ = io.WriteString(t.writer, programStatusQuery+"\x1b[c")
	}
}
func (t *programStatusTerminal) reply(seq string) {
	if t.pending && programReplyPattern.MatchString(seq) {
		t.pending = false
		t.supported = true
		if t.current.State != "" {
			t.send(t.current)
		}
	} else if programDAPattern.MatchString(seq) {
		t.pending = false
	}
}
func (t *programStatusTerminal) send(s programStatus) {
	seq := formatProgramStatus(s)
	if seq == t.last {
		return
	}
	if _, err := io.WriteString(t.writer, seq); err == nil {
		t.last = seq
	}
}
func (t *programStatusTerminal) set(s programStatus) {
	t.current = s
	if t.supported {
		t.send(s)
	}
}
func (t *programStatusTerminal) stop() {
	if t.supported && t.current.State != "" {
		_, _ = io.WriteString(t.writer, formatProgramStatus(programStatus{State: "clear"}))
	}
	t.supported, t.pending = false, false
	t.current = programStatus{}
	t.last = ""
}
func firstProgramStatusLine(s string) string {
	s = strings.TrimSpace(strings.SplitN(s, "\n", 2)[0])
	if s == "" {
		return "Error"
	}
	return s
}
func stringFromPayload(p map[string]any, key string) string { s, _ := p[key].(string); return s }

func (c *chatTUI) reportProgramStatus() {
	if c.programStatus == nil {
		return
	}
	active := c.running || c.compaction.active || c.branchSummaryCancel != nil
	if active && !c.programWasActive {
		c.programResting = programStatus{State: "done"}
	}
	c.programWasActive = active
	s := c.programResting
	if s.State == "" {
		s.State = "idle"
	}
	if active {
		s = programStatus{State: "working"}
	}
	if s.State == "working" || s.State == "done" {
		if name := c.programSessionName; name != c.sessionID && !strings.HasPrefix(name, "@") {
			s.Message = name
		}
	}
	if c.compaction.active {
		s.Message = "Compacting context"
	}
	// Only explicit user-blocking dialogs are reported. Browsing settings, models
	// or history is not a blocked agent. Never send dialog input or chat content.
	if c.modelMenuOpen {
		switch c.modelMenuKind {
		case "login-dialog":
			if c.loginDialog != nil {
				s = programStatus{State: "blocked", Kind: "auth", Message: c.loginDialog.title}
			}
		case "select":
			s = programStatus{State: "blocked", Kind: "question", Message: c.selectDialog.title}
		case "editor-dialog":
			if c.editorDialog != nil {
				s = programStatus{State: "blocked", Kind: "question", Message: c.editorDialog.title}
			}
		}
	}
	if c.editorAskActive {
		s = programStatus{State: "blocked", Kind: "question", Message: c.editorAskPrompt}
	}
	s.App = "gi"
	c.programStatus.set(s)
}
