package tui

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestModelPickerMetadataEnabledNavigationAndOwnership(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	inference.Init()
	for _, entry := range []struct {
		id, name  string
		window    int
		reasoning bool
	}{
		{"picker-pine", "Forest Pine", 32000, true}, {"picker-small", "Forest Small", 80, false}, {"picker-oak", "Forest Oak", 200, false},
	} {
		goai.RegisterModel(&goai.Model{ID: entry.id, Name: entry.name, Provider: "opencode-zen", Api: goai.ApiOpenAICompletions, ContextWindow: entry.window, Reasoning: entry.reasoning})
	}
	c, _, _ := modelTestChat(t)
	s := c.store
	c.cfg.EnabledModels = []string{"test/test-model", "opencode-zen/picker-small", "opencode-zen/picker-pine", "test/unavailable", "opencode-zen/picker-oak", "test/bootstrap"}
	ctx := context.Background()
	if _, err := s.CreateTurn(ctx, "measured", "A", "measurement fixture", nil); err != nil {
		t.Fatal(err)
	}
	measure := func(n int) {
		t.Helper()
		if err := s.AppendTurnEvent(ctx, "measured", "A", "context.measured", map[string]any{"tokens": n, "input": n, "model": "test/test-model"}); err != nil {
			t.Fatal(err)
		}
	}
	measure(100)
	before := *c.input
	c.openModelMenu()
	openInput := *c.input // Opening deliberately blurs the editor; rendering must not change it.
	for _, width := range []int{28, 60, 100, 140} {
		rows := collectElementTexts(c.renderModelMenu(width))
		for _, row := range rows {
			if gotui.StringWidth(row) > width {
				t.Fatalf("width %d overflow: %q", width, row)
			}
		}
		if width >= 60 {
			text := strings.Join(rows, "\n")
			if !strings.Contains(text, "opencode-zen/picker-pine · 32K ctx") || !strings.Contains(text, "opencode-zen/picker-small · context too small") || strings.Contains(text, "80 ctx") {
				t.Fatalf("metadata/blocked priority: %s", text)
			}
		}
		if c.modelMenuVisibleRows() > 6 || c.modelMenuHeight() > 8 || openInput.text != c.input.text || openInput.cursorPos != c.input.cursorPos || openInput.undoText != c.input.undoText || openInput.yankText != c.input.yankText || openInput.focused != c.input.focused {
			t.Fatal("metadata changed footprint or editor")
		}
	}
	for _, check := range []struct {
		query string
		want  []string
	}{
		{"FOREST pine", []string{"opencode-zen/picker-pine"}},
		{"32K ctx reasoning", []string{"opencode-zen/picker-pine"}},
		{"opencode-zen forest", []string{"opencode-zen/picker-small", "opencode-zen/picker-pine", "opencode-zen/picker-oak"}},
		{"missing", nil},
	} {
		c.modelMenuQuery = check.query
		c.applyModelMenuFilter()
		if !reflect.DeepEqual(c.modelMenuChoices, check.want) {
			t.Fatalf("%s: %v", check.query, c.modelMenuChoices)
		}
	}
	c.modelMenuQuery = "picker-small"
	c.applyModelMenuFilter()
	if c.modelMenuSelected != -1 {
		t.Fatal("disabled-only selection")
	}
	c.moveModelMenuSelection(1)
	c.acceptModelMenuSelection()
	if !strings.Contains(c.modelMenuError, "context too small") || !c.modelMenuOpen {
		t.Fatal(c.modelMenuError)
	}
	c.modelMenuQuery = ""
	c.applyModelMenuFilter()
	enabled := c.enabledModelMenuIndices()
	for i, want := range enabled {
		c.setModelMenuSelection(0)
		for step := 0; step < i; step++ {
			c.moveModelMenuSelection(1)
		}
		if c.modelMenuSelected != want {
			t.Fatalf("step%d=%d want%d", i, c.modelMenuSelected, want)
		}
	}
	c.setModelMenuSelection(0)
	c.moveModelMenuSelection(-1)
	if c.modelMenuSelected != enabled[len(enabled)-1] {
		t.Fatal("wrap")
	}
	c.moveModelMenuSelection(1)
	if c.modelMenuSelected != enabled[0] {
		t.Fatal("wrap down")
	}
	c.moveModelMenuSelection(5)
	if c.modelMenuSelected != enabled[min(5, len(enabled)-1)] {
		t.Fatal("page")
	}
	c.moveModelMenuSelection(-5)
	if c.modelMenuSelected != enabled[0] {
		t.Fatal("page up")
	}
	c.setModelMenuSelection(len(c.modelMenuChoices) - 1)
	if c.modelMenuSelected != enabled[len(enabled)-1] {
		t.Fatal("end")
	}
	// Runtime context changes after opening: the cached row remains advisory.
	c.modelMenuQuery = "picker-oak"
	c.applyModelMenuFilter()
	measure(300)
	c.acceptModelMenuSelection()
	if !strings.Contains(c.modelMenuError, "smaller than latest measured") || !c.modelMenuOpen {
		t.Fatal(c.modelMenuError)
	}
	c.closeModelMenu()
	c.openModelMenu()
	c.modelMenuQuery = "picker-oak"
	c.applyModelMenuFilter()
	if c.modelMenuSelected != -1 {
		t.Fatal("reopen did not refresh fit")
	}
	c.closeModelMenu()
	c.openModelMenu()
	c.modelMenuQuery = "bootstrap"
	c.applyModelMenuFilter()
	c.sessionID = "B"
	c.acceptModelMenuSelection()
	if c.modelMenuError != "session changed; reopen picker" {
		t.Fatal("retargeted selection")
	}
	b, _ := s.GetSession(ctx, "B")
	if b.State["selected_model"] != nil {
		t.Fatal("changed B")
	}
	c.sessionID = "A"
	c.sessionGeneration++
	c.acceptModelMenuSelection()
	if c.modelMenuError != "session changed; reopen picker" {
		t.Fatal("A-B-A retargeted selection")
	}
	c.closeModelMenu()
	if before.text != c.input.text || before.cursorPos != c.input.cursorPos || before.undoText != c.input.undoText || before.yankText != c.input.yankText {
		t.Fatal("editor changed")
	}
	if c.modelMenuMetadata != nil || c.modelMenuSession != (sessionScope{}) || c.modelMenuHeight() != 0 {
		t.Fatal("closed metadata/chrome retained")
	}
}

func TestModelPickerContextReadFailureStaysRecoverable(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	c, _, _ := modelTestChat(t)
	s := c.store
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	c.openModelMenu()
	c.modelMenuQuery = "bootstrap"
	c.applyModelMenuFilter()
	c.acceptModelMenuSelection()
	if !strings.Contains(c.modelMenuError, "context unavailable") || !c.modelMenuOpen {
		t.Fatal(c.modelMenuError)
	}
	reopened, err := store.Open(filepath.Join(c.cfg.WorkspaceRoot, "models.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	c.store = reopened
	c.acceptModelMenuSelection()
	if !c.modelMenuOpen || c.modelMenuSelected < 0 || c.cfg.DefaultModel != "test-model" {
		t.Fatal("retry must enable, not activate")
	}
	c.acceptModelMenuSelection()
	if c.modelMenuOpen || c.cfg.DefaultModel != "bootstrap" {
		t.Fatal("did not recover")
	}
}

func TestModelPickerInlineContextPreservesIdentityAndWidth(t *testing.T) {
	c := &chatTUI{modelMenuKind: "model", modelMenuMetadata: map[string]modelPickerMetadata{
		"provider/pine":       {context: "32K ctx"},
		"provider/中文🙂e\u0301": {context: "1.0M ctx"},
		"provider/blocked":    {context: "80 ctx", unavailable: "context too small"},
	}}
	for _, key := range []string{"provider/pine", "provider/中文🙂e\u0301", "provider/blocked", "provider/unknown"} {
		for _, width := range []int{0, 1, 8, 20, 28, 60, 100, 140} {
			got := c.modelPickerRowLabel(key, width)
			if gotui.StringWidth(got) > width || strings.ContainsAny(got, "\r\n\x1b") {
				t.Fatalf("width %d: %q", width, got)
			}
			meta := c.modelMenuMetadata[key]
			want := key
			full := key + " · " + meta.context
			if meta.context != "" && meta.unavailable == "" && gotui.StringWidth(full) <= width {
				want = full
			}
			if got != selectorText(want, width) {
				t.Fatalf("width %d got %q want %q", width, got, selectorText(want, width))
			}
		}
	}
	// Other selectors share the renderer but must never inherit model metadata.
	c.modelMenuKind = "session"
	if got := c.modelPickerRowLabel("provider/pine", 100); got != "provider/pine" {
		t.Fatal(got)
	}
}
