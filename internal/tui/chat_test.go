package tui

import (
	"context"
	"os"
	"path/filepath"
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

func TestTreeLinesShowsParentChildSessions(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	ctx := context.Background()
	root, _ := s.CreateSession(ctx, "session_root", "@agent", map[string]any{"model": "bootstrap", "status": "idle"})
	child, _ := s.CloneSession(ctx, root.ID, "session_child", "@agent1", "agent1")
	c := &chatTUI{store: s, sessionID: child.ID}
	lines := strings.Join(c.treeLines(), "\n")
	for _, want := range []string{"tree: sessions:", "@gi session_root", "* @agent1 session_child"} {
		if !strings.Contains(lines, want) {
			t.Fatalf("tree missing %q:\n%s", want, lines)
		}
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

func TestHandleEventStreamsDraftIntoTranscript(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": "hello"})
	if c.status != "⏳ hello…" || len(c.transcript) != 1 || c.transcript[0] != "Neo: hello" {
		t.Fatalf("unexpected first draft state: status=%q transcript=%#v", c.status, c.transcript)
	}
	c.handleEvent(map[string]any{"type": "agent_draft_delta", "delta": " world"})
	if len(c.transcript) != 1 || c.transcript[0] != "Neo: hello world" {
		t.Fatalf("unexpected merged draft transcript: %#v", c.transcript)
	}
	c.handleEvent(map[string]any{"type": "new_post", "data": map[string]any{"content": "hello world"}})
	if c.running || c.draft != "" || c.draftLineIndex != -1 || c.transcript[0] != "Neo: hello world" {
		t.Fatalf("unexpected final streamed state: running=%v draft=%q draftLineIndex=%d transcript=%#v", c.running, c.draft, c.draftLineIndex, c.transcript)
	}
}

func TestHandleEventStatusRendering(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo", DefaultModel: "bootstrap"}, stickToBottom: true, draftLineIndex: -1}
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

func TestOnSubmitRequiresModelSelectionBeforeFirstPrompt(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: "session_1", cfg: config.RuntimeConfig{AssistantName: "Neo"}, draftLineIndex: -1}
	c.input = gotui.NewInput()
	c.onSubmit("hello")
	if c.running {
		t.Fatal("expected prompt submission to be blocked without a model")
	}
	joined := strings.Join(c.transcript, "\n")
	if !strings.Contains(joined, "no model selected") {
		t.Fatalf("expected model-selection guidance, got:\n%s", joined)
	}
}

func TestInitUsesConfiguredGiDefaultModelWithoutFirstUsePrompt(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	sess, err := s.CreateSession(context.Background(), "session_1", "@agent", map[string]any{"status": "idle"})
	if err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{store: s, engine: turn.New(s), sessionID: sess.ID, cfg: config.RuntimeConfig{AssistantName: "Gi", DefaultModel: "opencode-zen/minimax-m2.5-free"}}
	cleanup := c.Init()
	defer cleanup()
	if c.status != "Gi · opencode-zen/minimax-m2.5-free" {
		t.Fatalf("unexpected status: %q", c.status)
	}
	joined := strings.Join(c.transcript, "\n")
	if strings.Contains(joined, "no model selected") {
		t.Fatalf("unexpected first-use prompt with default model: %s", joined)
	}
}

func TestModelCommandPersistsSelection(t *testing.T) {
	root := t.TempDir()
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c := &chatTUI{store: s, sessionID: "session_1", cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultThinkingLevel: "medium", EnabledModels: []string{"qwen3:latest"}}}
	lines := c.modelCommand([]string{"/model", "ollama/gemma4:latest"})
	if len(lines) == 0 || !strings.Contains(lines[0], "model set to ollama/gemma4:latest") {
		t.Fatalf("unexpected model command output: %#v", lines)
	}
	cfg := config.Load(root)
	if cfg.DefaultProvider != "ollama" || cfg.DefaultModel != "ollama/gemma4:latest" {
		t.Fatalf("unexpected persisted model config: %#v", cfg)
	}
}

func TestSettingsLinesExposeRuntimeState(t *testing.T) {
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := turn.New(s)
	c := &chatTUI{store: s, engine: engine, cfg: config.RuntimeConfig{WorkspaceRoot: "/tmp/ws", DefaultProvider: "test", DefaultModel: "m", DefaultThinkingLevel: "low", MaxIterations: 64}}
	lines := strings.Join(c.settingsLines(), "\n")
	for _, want := range []string{"settings: runtime:", "provider=test model=m thinking=low", "workspace=/tmp/ws", "tools active="} {
		if !strings.Contains(lines, want) {
			t.Fatalf("settings missing %q:\n%s", want, lines)
		}
	}
}

func TestRenderMessageLineFoldsToolAndCompaction(t *testing.T) {
	c := &chatTUI{cfg: config.RuntimeConfig{AssistantName: "Neo"}}
	toolLine := c.renderMessageLine(store.Message{Role: "tool_result", Content: "long tool output", Payload: map[string]any{"kind": "tool_result", "tool_name": "shell", "is_error": false}})
	if toolLine != "tool[shell/ok]: long tool output" {
		t.Fatalf("tool line = %q", toolLine)
	}
	compactionLine := c.renderMessageLine(store.Message{Role: "assistant", Content: "summary", Payload: map[string]any{"kind": "compaction", "tokens_before": 1200}})
	if compactionLine != "compact: summary (tokens_before=1200)" {
		t.Fatalf("compaction line = %q", compactionLine)
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

func TestPluginLinesShowsExtensionsAndHooks(t *testing.T) {
	rootDir := t.TempDir()
	extDir := filepath.Join(rootDir, ".gi", "extensions")
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extDir, "demo.joke"), []byte("nil"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	engine := turn.NewWithRuntimeConfig(s, config.RuntimeConfig{WorkspaceRoot: rootDir, DefaultModel: "bootstrap"}, "")
	engine.RegisterHook("tool_call", "test-hook", func(context.Context, turn.HookRequest) (turn.HookResponse, error) { return turn.HookResponse{}, nil })
	c := &chatTUI{store: s, engine: engine}
	lines := strings.Join(c.pluginLines(), "\n")
	for _, want := range []string{"plugins: extensions:", "loaded joker .gi/extensions/demo.joke", "plugins: hooks:", "tool_call from test-hook"} {
		if !strings.Contains(lines, want) {
			t.Fatalf("plugins missing %q:\n%s", want, lines)
		}
	}
}
