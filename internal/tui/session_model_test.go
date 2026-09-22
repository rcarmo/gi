package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/store"
)

func modelTestChat(t *testing.T) (*chatTUI, string, string) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "models.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	settings := filepath.Join(root, ".pi", "settings.json")
	os.MkdirAll(filepath.Dir(settings), 0700)
	if err := os.WriteFile(settings, []byte(`{"defaultProvider":"test","defaultModel":"test-model","defaultThinkingLevel":"high","enabledModels":["test/test-model","test/bootstrap","test/unavailable"],"untouched":{"flag":true}}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(context.Background(), id, id, map[string]any{"model": "test-model", "provider": "test", "status": "idle", "thinking_level": "high"}); err != nil {
			t.Fatal(err)
		}
	}
	c := &chatTUI{store: s, cfg: config.RuntimeConfig{WorkspaceRoot: root, DefaultProvider: "test", DefaultModel: "test-model", DefaultThinkingLevel: "high", EnabledModels: []string{"test/test-model", "test/bootstrap", "test/unavailable"}}, draftLineIndex: -1}
	c.bindSession("A")
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.ensureInput()
	c.input.SetText("multiline draft\n中文🙂")
	c.input.cursorPos = 4
	c.input.undoText = "undo A"
	c.input.hasUndo = true
	c.input.yankText = "yank A"
	c.history = []string{"history A"}
	return c, path, settings
}

func TestTerminalModelSelectionIsLocalAndRestoresAtStartup(t *testing.T) {
	c, path, settings := modelTestChat(t)
	before, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	lines := c.modelCommand([]string{"/model", "test/bootstrap"})
	if strings.Contains(strings.Join(lines, " "), "error:") {
		t.Fatal(lines)
	}
	if c.cfg.DefaultModel != "bootstrap" || c.cfg.DefaultThinkingLevel != "" {
		t.Fatal("selection not applied")
	}
	c.switchSession("B")
	if c.cfg.DefaultModel != "test-model" {
		t.Fatal("A leaked into B")
	}
	c.switchSession("A")
	if c.input.Text() != "multiline draft\n中文🙂" || c.input.cursorPos != 4 || c.input.undoText != "undo A" || c.input.yankText != "yank A" || !reflect.DeepEqual(c.history, []string{"history A"}) {
		t.Fatal("draft state damaged")
	}
	if err := c.store.TouchSessionState(context.Background(), "A", map[string]any{"model": "old-runtime", "provider": "old-provider"}); err != nil {
		t.Fatal(err)
	}
	if c.contextSummaryData().model != "bootstrap" {
		t.Fatal("footer showed old runtime model")
	}
	// Reopen a real DB and use startup binding rather than switching from another chat.
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	resumed := &chatTUI{store: s, cfg: config.RuntimeConfig{DefaultProvider: "test", DefaultModel: "test-model"}}
	resumed.bindSession("A")
	if resumed.cfg.DefaultModel != "bootstrap" || resumed.cfg.DefaultProvider != "test" {
		t.Fatal("startup did not restore selected model")
	}
	for _, bad := range []string{"unknown", "test/unavailable", "99"} {
		out := c.modelCommand([]string{"/model", bad})
		if !strings.Contains(strings.Join(out, " "), "error:") {
			t.Fatal("invalid selection accepted", bad)
		}
		if c.cfg.DefaultModel != "bootstrap" {
			t.Fatal("invalid selection mutated model")
		}
	}
	after, _ := os.ReadFile(settings)
	if string(before) != string(after) {
		t.Fatal("global settings changed")
	}
	for _, id := range []string{"A", "B"} {
		turns, _ := s.ListTurns(context.Background(), id)
		if len(turns) > 0 {
			t.Fatal("selection submitted work")
		}
	}
}

func TestTerminalModelPickerRetainsDraftAndFootprintOnSuccessAndError(t *testing.T) {
	for _, size := range [][2]int{{60, 18}, {100, 22}, {140, 36}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			c, _, settings := modelTestChat(t)
			c.outputWidth, c.outputHeight = size[0], size[1]
			before, _ := os.ReadFile(settings)
			footer := len(c.footerLines(size[0]))
			transcript := append([]string(nil), c.transcript...)
			c.openModelMenu()
			c.modelMenuQuery = "test/unavailable"
			c.applyModelMenuFilter()
			c.acceptModelMenuSelection()
			if !c.modelMenuOpen || c.modelMenuError == "" || c.cfg.DefaultModel != "test-model" {
				t.Fatal("invalid selection didn't retain menu/model")
			}
			for _, line := range collectElementTexts(c.renderModelMenu(size[0])) {
				if gotui.StringWidth(line) > size[0] {
					t.Fatal("overflow", line)
				}
			}
			if c.modelMenuHeight() > 8 {
				t.Fatal("error added rows")
			}
			c.closeModelMenu()
			c.openModelMenu()
			c.modelMenuQuery = "test/bootstrap"
			c.applyModelMenuFilter()
			c.acceptModelMenuSelection()
			if c.modelMenuOpen || c.cfg.DefaultModel != "bootstrap" || !c.inputActive {
				t.Fatal("selection did not close/focus")
			}
			if c.input.Text() != "multiline draft\n中文🙂" || c.input.cursorPos != 4 || !reflect.DeepEqual(c.transcript, transcript) || c.modelMenuHeight() != 0 || len(c.footerLines(size[0])) != footer {
				t.Fatal("model picker altered draft/transcript/idle rows")
			}
			c.cycleModel(1) // unavailable: retain selection, report failure in requested transcript.
			if c.cfg.DefaultModel != "bootstrap" {
				t.Fatal("cycle accepted unavailable model")
			}
			c.cycleModel(-1)
			if c.cfg.DefaultModel != "test-model" {
				t.Fatal("valid reverse cycle failed")
			}
			after, _ := os.ReadFile(settings)
			if string(before) != string(after) {
				t.Fatal("picker/cycle wrote global settings")
			}
		})
	}
}

func TestTerminalModelStoreFailureDoesNotDeclareSuccess(t *testing.T) {
	c, _, _ := modelTestChat(t)
	c.store.Close()
	lines := c.modelCommand([]string{"/model", "test/bootstrap"})
	if !strings.Contains(strings.Join(lines, " "), "error:") || c.cfg.DefaultModel != "test-model" {
		t.Fatal("storage failure lost selection", lines)
	}
}
