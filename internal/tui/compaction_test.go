package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
	"github.com/rcarmo/gi/internal/turn"
)

func TestTerminalManualCompactionPreservesEditorAndScopes(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		if err := c.store.AddMessage(ctx, fmt.Sprint(i), "A", "user", strings.Repeat("history ", 20), nil); err != nil {
			t.Fatal(err)
		}
	}
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	c.engine.RegisterHook(turn.HookSessionBeforeCompact, "gate", func(ctx context.Context, _ turn.HookRequest) (turn.HookResponse, error) {
		entered <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			return turn.HookResponse{}, ctx.Err()
		}
		return turn.HookResponse{Payload: map[string]any{"summary": "summary"}}, nil
	})
	c.input.SetText("unsent draft")
	c.input.cursorPos = 3
	c.input.undoText = "undo"
	c.input.hasUndo = true
	c.input.yankText = "yank"
	c.history = []string{"one"}
	c.queuedDrafts = []string{"pending"}
	editor := func() sessionEditorState { c.saveSessionEditor(); return c.sessionEditors[c.sessionID] }
	before := editor()
	widths := []int{60, 100, 140}
	rows := map[int]int{}
	for _, w := range widths {
		rows[w] = len(c.footerLines(w))
	}
	c.startCompaction()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("not started")
	}
	c.syncCompactionActivity()
	if !c.compaction.active || !strings.Contains(c.compactionInline(), "Compacting") {
		t.Fatal(c.compaction)
	}
	if !reflect.DeepEqual(before, editor()) {
		t.Fatal("editor changed")
	}
	for _, w := range widths {
		lines := c.footerLines(w)
		if len(lines) != rows[w] {
			t.Fatal("active compaction added rows", lines)
		}
		for _, line := range lines {
			if gotui.StringWidth(line) > w {
				t.Fatal("wrapped", line)
			}
		}
	}
	run := c.compaction.turnID
	scope := c.selectionScope()
	c.startCompaction()
	turns, _ := c.store.ListTurns(ctx, "A")
	if len(turns) != 1 {
		t.Fatal("duplicate work")
	}
	c.switchSession("B")
	c.input.SetText("B draft")
	c.handleSessionTopicEvent(sessionTopicEvent{scope: scope, envelope: topics.Envelope{Topic: "session.compaction", SessionID: "A", Payload: map[string]any{"type": "compaction_completed"}}})
	if c.compaction.active || c.input.Text() != "B draft" {
		t.Fatal("foreign event changed target")
	}
	c.switchSession("A")
	c.syncCompactionActivity()
	if c.compaction.turnID != run || !c.compaction.active {
		t.Fatal(c.compaction)
	}
	c.stopCompaction()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rec, _ := c.store.GetTurn(ctx, run)
		if rec.Status == "cancelled" {
			break
		}
		time.Sleep(time.Millisecond)
	}
	c.syncCompactionActivity()
	rec, _ := c.store.GetTurn(ctx, run)
	if rec.Status != "cancelled" || c.compaction.active {
		t.Fatal(rec, c.compaction)
	}
	if !reflect.DeepEqual(before, editor()) {
		t.Fatal("cancel changed editor")
	}
	c.compaction.noticeUntil = time.Now().Add(-time.Second)
	c.status = ""
	c.running = false
	for _, w := range widths {
		if len(c.footerLines(w)) != rows[w] {
			t.Fatal("extra idle rows")
		}
	}
	close(release)
}

func TestCompactionNonCompletionCannotRenderSuccess(t *testing.T) {
	c := &chatTUI{}
	for _, kind := range []string{"compaction_started", "compaction_suppressed", "compaction_failed", "compaction_cancelled"} {
		c.renderCompactionEvent(map[string]any{"type": kind}, time.Now())
	}
	if len(c.transcript) != 0 || c.status != "" {
		t.Fatal(c.transcript, c.status)
	}
}

// Opt-in fixture launched inside a real PTY by test-tui-compaction.mjs. It adds
// only a blocking native hook to the production terminal loop, never fake UI.
func TestTerminalCompactionPTYFixture(t *testing.T) {
	dir := os.Getenv("GI_TUI_COMPACTION_FIXTURE")
	if dir == "" {
		t.Skip("PTY fixture only")
	}
	cfg := config.Load(dir)
	cfg.DefaultModel = "test-model"
	cfg.EnabledModels = []string{"test-model", "bootstrap"}
	cfg.Compaction.Enabled = false
	cfg.Hooks.TimeoutMS = 55000
	s, err := store.Open(filepath.Join(dir, "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)
	defer e.Close()
	_, err = e.RegisterHook(turn.HookSessionBeforeCompact, "pty-gate", func(ctx context.Context, req turn.HookRequest) (turn.HookResponse, error) {
		path := filepath.Join(dir, "release")
		tick := time.NewTicker(20 * time.Millisecond)
		defer tick.Stop()
		for {
			if _, err := os.Stat(path); err == nil {
				os.Remove(path)
				break
			}
			select {
			case <-ctx.Done():
				return turn.HookResponse{}, ctx.Err()
			case <-tick.C:
			}
		}
		return turn.HookResponse{Payload: map[string]any{"summary": "Native terminal compaction summary"}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runWithEngine(s, e, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestTerminalCompactionOutcomeSequenceAndFailureDetail(t *testing.T) {
	for _, outcome := range []string{"completed", "cancelled", "suppressed", "failed"} {
		t.Run(outcome, func(t *testing.T) {
			c := sessionTestChat(t)
			ctx := context.Background()
			c.store.CreateTurn(ctx, "run", "A", "prompt", nil)
			c.store.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run")
			seq, err := c.store.BeginCompaction(ctx, "A", "run", nil)
			if err != nil {
				t.Fatal(err)
			}
			c.syncCompactionActivity()
			if !c.compaction.active || c.compaction.sequence != seq {
				t.Fatal(c.compaction)
			}
			if _, err = c.store.FinishCompaction(ctx, "A", "run", seq, outcome, "summary", map[string]any{"detail": "native detail"}); err != nil {
				t.Fatal(err)
			}
			c.syncCompactionActivity()
			if c.compaction.active || c.compaction.sequence <= seq || !strings.Contains(c.compaction.notice, "native detail") {
				t.Fatal(c.compaction)
			}
			until := c.compaction.noticeUntil
			c.syncCompactionActivity()
			if c.compaction.noticeUntil != until {
				t.Fatal("duplicate event extended notice")
			}
			if len(c.transcript) != 0 {
				t.Fatal("lifecycle inflated transcript", c.transcript)
			}
		})
	}
}
