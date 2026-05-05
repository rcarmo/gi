package tui

import (
	"context"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func TestFocusInputActivatesInput(t *testing.T) {
	c := &chatTUI{}
	c.focusInput()
	if !c.inputActive {
		t.Fatal("expected inputActive to be true")
	}
}

func TestHandleMouseOnInputRegionFocusesInput(t *testing.T) {
	el := gotui.New(gotui.WithWidth(20), gotui.WithHeight(1))
	buf := gotui.NewBuffer(80, 25)
	el.Render(buf, 80, 25)
	c := &chatTUI{inputRegion: el}
	c.inputActive = false
	consumed := c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseLeft, Action: gotui.MousePress, X: 2, Y: 0})
	if !consumed {
		t.Fatal("expected mouse click to be consumed")
	}
	if !c.inputActive {
		t.Fatal("expected inputActive to become true after click")
	}
}

func TestHandleMouseWheelScrollsTranscript(t *testing.T) {
	c := &chatTUI{
		transcript:       []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		transcriptScroll: 3,
	}
	if !c.HandleMouse(gotui.MouseEvent{Button: gotui.MouseWheelUp, Action: gotui.MousePress}) {
		t.Fatal("expected wheel event to be consumed")
	}
	if c.transcriptScroll >= 3 {
		t.Fatalf("expected transcript scroll to move up, got %d", c.transcriptScroll)
	}
}

func TestScrollTranscriptClampsToBounds(t *testing.T) {
	c := &chatTUI{transcript: []string{"1", "2", "3", "4", "5", "6", "7"}}
	c.scrollTranscript(-10)
	if c.transcriptScroll != 0 {
		t.Fatalf("expected top clamp at 0, got %d", c.transcriptScroll)
	}
	c.scrollTranscript(100)
	if c.transcriptScroll != 3 {
		t.Fatalf("expected bottom clamp at 3, got %d", c.transcriptScroll)
	}
}

func TestHandleForkCommandSwitchesSession(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, err := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: root.ID, cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, eventCh: make(chan map[string]any, 64)}
	c.handleCommand("/fork @agent1")
	if c.sessionID == root.ID {
		t.Fatalf("expected fork command to switch session")
	}
	sess, err := s.GetSession(ctx, c.sessionID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if sess.Scope == nil || sess.Scope.AgentID != "agent1" {
		t.Fatalf("unexpected forked session: %#v", sess)
	}
}

func TestResolveSessionRefByAgentID(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, _ := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	_, _ = s.CloneSession(ctx, root.ID, "session_child", "@agent1", "agent1")
	c := &chatTUI{store: s, sessionID: root.ID}
	sess, err := c.resolveSessionRef("@agent1")
	if err != nil {
		t.Fatalf("resolve session ref: %v", err)
	}
	if sess.Scope == nil || sess.Scope.AgentID != "agent1" {
		t.Fatalf("unexpected session: %#v", sess)
	}
}

func TestHandleEventStatusRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true}
	c.handleEvent(map[string]any{"type": "agent_thought_delta"})
	if c.status != "Thinking…" {
		t.Fatalf("thinking status = %q", c.status)
	}
	c.handleEvent(map[string]any{"type": "tool_finished", "tool": "read"})
	if c.status != "Tool finished: read" {
		t.Fatalf("tool finished status = %q", c.status)
	}
	c.handleEvent(map[string]any{"type": "tool_failed", "tool": "shell", "error": "boom"})
	if c.status != "Tool failed: shell" || len(c.transcript) == 0 || c.transcript[len(c.transcript)-1] != "sys: tool failed: shell: boom" {
		t.Fatalf("tool failed status/transcript = %q %#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "compaction", "messages_before": 10, "messages_after": 4, "tokens_before": 1234})
	if c.status != "Compacted context" || c.transcript[len(c.transcript)-1] != "sys: compacted context: messages 10→4, tokens_before=1234" {
		t.Fatalf("compaction status/transcript = %q %#v", c.status, c.transcript)
	}
}

func TestFooterTextContainsStableHints(t *testing.T) {
	c := &chatTUI{}
	footer := c.footerText()
	for _, want := range []string{"Hints:", "/help", "/tools active|activate|reset", "Tab focus", "F2/F3 history", "Ctrl-D quit"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("footer missing %q: %s", want, footer)
		}
	}
}
