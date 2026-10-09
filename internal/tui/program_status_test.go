package tui

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/topics"
)

func TestProgramStatusEncoding(t *testing.T) {
	tests := []struct {
		s    programStatus
		want string
	}{
		{programStatus{State: "blocked", App: "gi", Kind: "permission", Message: "Allow bash?"}, "\x1b]7501;state=blocked:app=gi:kind=permission:msg=" + base64.StdEncoding.EncodeToString([]byte("Allow bash?")) + "\x1b\\"},
		{programStatus{State: "clear"}, "\x1b]7501;state=clear\x1b\\"},
		{programStatus{State: "working", App: "my app", Kind: "auth", Message: " \n "}, "\x1b]7501;state=working\x1b\\"},
		{programStatus{State: "error", Message: "first\nsecond\x1b[31m\u009bthird\t"}, "\x1b]7501;state=error:msg=" + base64.StdEncoding.EncodeToString([]byte("first second [31m third")) + "\x1b\\"},
	}
	for _, tt := range tests {
		if got := formatProgramStatus(tt.s); got != tt.want {
			t.Errorf("got %q want %q", got, tt.want)
		}
	}
	for _, s := range []string{strings.Repeat("é", 2000), strings.Repeat("a", 2047) + "🙂"} {
		seq := formatProgramStatus(programStatus{State: "working", Message: s})
		msg, err := base64.StdEncoding.DecodeString(strings.TrimSuffix(strings.Split(seq, ":msg=")[1], "\x1b\\"))
		if err != nil || len(msg) > 2048 || bytes.Contains(msg, []byte("�")) {
			t.Fatalf("invalid bounded message: %d %v", len(msg), err)
		}
	}
	if !strings.Contains(formatProgramStatus(programStatus{State: "idle", App: strings.Repeat("a", 32)}), ":app=") || strings.Contains(formatProgramStatus(programStatus{State: "idle", App: strings.Repeat("a", 33)}), ":app=") {
		t.Fatal("app limits")
	}
}

func TestProgramStatusNegotiation(t *testing.T) {
	for _, reply := range []string{programStatusQuery, "\x1b]7501;?\x07", "\x1b]7501;?version=2\x1b\\"} {
		var out bytes.Buffer
		term := newProgramStatusTerminal(&out, "")
		term.start()
		term.set(programStatus{State: "working", App: "gi"})
		if out.String() != programStatusQuery+"\x1b[c" {
			t.Fatal("status before support", out.String())
		}
		term.reply("\x1b]11;rgb:0000/0000/0000\x07")
		if term.supported {
			t.Fatal("unrelated OSC accepted")
		}
		term.reply(reply)
		if !strings.Contains(out.String(), "state=working:app=gi") {
			t.Fatal("cached status not reported")
		}
		n := out.Len()
		term.set(programStatus{State: "working", App: "gi"})
		if out.Len() != n {
			t.Fatal("duplicate status")
		}
		term.stop()
		if !strings.HasSuffix(out.String(), "\x1b]7501;state=clear\x1b\\") {
			t.Fatal("missing clear")
		}
	}
	var out bytes.Buffer
	term := newProgramStatusTerminal(&out, "")
	term.start()
	term.reply("\x1b[?62;22c")
	term.reply(programStatusQuery)
	term.set(programStatus{State: "idle"})
	term.stop()
	if strings.Contains(out.String(), "state=") {
		t.Fatal("late reply accepted")
	}
	for _, override := range []string{"0", "1"} {
		out.Reset()
		term = newProgramStatusTerminal(&out, override)
		term.start()
		term.set(programStatus{State: "idle"})
		term.reply(programStatusQuery)
		term.stop()
		if strings.Contains(out.String(), ";?") {
			t.Fatal("override emitted query")
		}
		if (out.Len() > 0) != (override == "1") {
			t.Fatal("override failed", override, out.String())
		}
	}
}

func TestProgramStatusNativeLifecycle(t *testing.T) {
	var out bytes.Buffer
	c := &chatTUI{sessionID: "s", programSessionName: "Named session", programStatus: newProgramStatusTerminal(&out, "1")}
	expect := func(state string) {
		t.Helper()
		c.reportProgramStatus()
		if c.programStatus.current.State != state {
			t.Fatalf("got %+v want %s", c.programStatus.current, state)
		}
	}
	expect("idle")
	c.running = true
	expect("working")
	if c.programStatus.current.Message != "Named session" {
		t.Fatal("session name missing")
	}
	c.compaction.active = true
	expect("working")
	if c.programStatus.current.Message != "Compacting context" {
		t.Fatal("compaction message")
	}
	c.openSelect("Choose action", []string{"a"}, nil, nil)
	expect("blocked")
	if c.programStatus.current.Kind != "question" {
		t.Fatal("kind")
	}
	c.modelMenuOpen = false
	c.compaction.active = false
	c.running = false
	expect("done")
	c.handleTopicEvent(topics.Envelope{Topic: "runtime.turn", SessionID: "s", Payload: map[string]any{"type": "turn_terminal", "status": "cancelled"}})
	expect("idle")
	c.running = true
	expect("working")
	c.handleEvent(map[string]any{"type": "error", "error": "First line\nsecret second line"})
	expect("error")
	if c.programStatus.current.Message != "First line" {
		t.Fatal("error message")
	}
	c.running = true
	expect("working")
	c.running = false
	expect("done")
	c.modelMenuOpen = true
	c.modelMenuKind = "login-dialog"
	c.loginDialog = &loginDialogState{title: "Login"}
	expect("blocked")
	if c.programStatus.current.Kind != "auth" {
		t.Fatal("login kind")
	}
	c.modelMenuOpen = false
	c.editorAskActive = true
	c.editorAskPrompt = "Question"
	expect("blocked")
	c.editorAskActive = false
	c.modelMenuOpen = true
	c.modelMenuKind = "settings"
	expect("done")
	c.bindSession("other")
	c.modelMenuOpen = false
	expect("idle")
	c.programStatus.stop()
	if strings.Contains(out.String(), "secret") {
		t.Fatal("leaked error body")
	}
}
