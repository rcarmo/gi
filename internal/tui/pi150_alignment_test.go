package tui

import (
	"fmt"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
)

func TestPi150SelectorPagingOwnsInput(t *testing.T) {
	for _, regular := range []bool{false, true} {
		for _, size := range [][2]int{{60, 18}, {100, 22}, {140, 36}} {
			t.Run(fmt.Sprintf("regular=%v/%dx%d", regular, size[0], size[1]), func(t *testing.T) {
				c := sessionTestChat(t)
				c.regularMode, c.outputWidth, c.outputHeight = regular, size[0], size[1]
				c.input.SetText("unsent Unicode 草稿")
				before := c.input.Text()
				labels := make([]string, 30)
				for i := range labels {
					labels[i] = fmt.Sprintf("test/model-%02d", i)
				}
				c.modelMenuOpen, c.modelMenuKind = true, "model"
				c.modelMenuAll, c.modelMenuChoices = labels, labels
				c.modelMenuMetadata = map[string]modelPickerMetadata{}
				for _, label := range labels {
					c.modelMenuMetadata[label] = modelPickerMetadata{}
				}
				page := func(key gotui.Key) { pressAppKey(t, c, gotui.KeyEvent{Key: key}) }
				page(gotui.KeyPageDown)
				if c.modelMenuSelected != c.piModelSelectorListRows() {
					t.Fatalf("model page=%d want %d", c.modelMenuSelected, c.piModelSelectorListRows())
				}
				for range 40 {
					page(gotui.KeyPageDown)
				}
				if c.modelMenuSelected != 29 {
					t.Fatal("page wrapped", c.modelMenuSelected)
				}
				for range 40 {
					page(gotui.KeyPageUp)
				}
				if c.modelMenuSelected != 0 {
					t.Fatal("page above first", c.modelMenuSelected)
				}

				c.modelMenuKind = "fork"
				page(gotui.KeyPageDown)
				if c.modelMenuSelected != forkMaxVisible {
					t.Fatal("fork did not page")
				}
				c.modelMenuKind, c.modelMenuSelected = "scoped-models", 0
				c.scopedModels = &scopedModelsState{all: labels, allOn: true}
				page(gotui.KeyPageDown)
				if c.scopedModels.selected != scopedModelsMaxVisible {
					t.Fatal("scope did not page")
				}
				c.modelMenuKind = "auth-selector"
				c.authSelector = &authSelectorState{all: make([]inference.LoginOption, 30)}
				page(gotui.KeyPageDown)
				if c.authSelector.selected != 8 {
					t.Fatal("login did not page")
				}
				c.modelMenuKind = "settings"
				c.settingsList = &settingsListState{items: make([]settingItem, 30)}
				page(gotui.KeyPageDown)
				if c.settingsList.selected != settingsMaxVisible {
					t.Fatal("settings did not page")
				}
				c.modelMenuOpen = false
				c.slash = slashMenu{active: true, items: make([]slashItem, 30)}
				page(gotui.KeyPageDown)
				if c.slash.selected != slashMaxVisible {
					t.Fatal("autocomplete did not page")
				}
				page(gotui.KeyPageUp)
				if c.slash.selected != 0 {
					t.Fatal("autocomplete did not page back")
				}
				if c.input.Text() != before {
					t.Fatal("paging changed draft")
				}
			})
		}
	}
}

func TestPi150PagingEmptyListsAndThemePreview(t *testing.T) {
	if got := selectPageIndex(0, 0, 10, 1); got != 0 {
		t.Fatal(got)
	}
	c := &chatTUI{modelMenuKind: "model", modelMenuSelected: -1}
	c.pageModelMenuSelection(1)
	if c.modelMenuSelected != -1 {
		t.Fatal("selected empty model list")
	}
	preview := ""
	options := make([]selectOption, 24)
	for i := range options {
		options[i].value = fmt.Sprint(i)
	}
	menu := newSelectSubmenu("Themes", "", options, "0", nil, nil, func(value string) { preview = value })
	menu.handle("pgdn", 0)
	if menu.selected != 10 || preview != "10" {
		t.Fatalf("page did not preview theme: %+v %q", menu, preview)
	}
	menu.handle("pgup", 0)
	if menu.selected != 0 || preview != "0" {
		t.Fatal("page up preview")
	}
}

func TestPi150DistinctModelNames(t *testing.T) {
	for _, tc := range []struct{ id, name, want string }{
		{"claude-sonnet-4-5", "Claude Sonnet 4.5", ""},
		{"eu.anthropic.claude-sonnet-4-5-20250929-v1:0", "Claude Sonnet 4.5 (EU)", ""},
		{"minimax/minimax-m3", "MiniMax M3 [work, proxy]", "MiniMax M3 [work, proxy]"},
		{"mistral-small-2603", "Mistral Small 4", "Mistral Small 4"},
		{"モデル", "モデル", ""},
		{"model", "  Friendly Alias  ", "Friendly Alias"},
	} {
		if got := distinctModelName(tc.id, tc.name); got != tc.want {
			t.Errorf("%q: %q want %q", tc.id, got, tc.want)
		}
	}
	c := &chatTUI{scopedModels: &scopedModelsState{all: []string{"test/model"}, names: map[string]string{"test/model": "Friendly Alias"}, allOn: true}}
	var text strings.Builder
	for _, row := range c.piScopedModelsRows(100) {
		for _, span := range row {
			text.WriteString(span.Text)
			text.WriteByte('\n')
		}
	}
	if !strings.Contains(text.String(), " (Friendly Alias)") {
		t.Fatal("scoped selector omitted distinct name")
	}
}

func TestPi150FuzzyContiguousMatchesWin(t *testing.T) {
	items := []slashItem{
		{name: "gpt-oss", description: modelSelectorSearchText("openrouter", "openai/gpt-oss-120b", "GPT OSS")},
		{name: "sonnet", description: modelSelectorSearchText("openrouter", "anthropic/claude-sonnet-4.5", "Claude Sonnet")},
		{name: "opus", description: modelSelectorSearchText("openrouter", "anthropic/claude-opus-4.1", "Claude Opus")},
	}
	got := piFuzzyFilter(items, "opus", func(it slashItem) string { return it.description })
	if len(got) == 0 || got[0].name != "opus" {
		t.Fatalf("real opus match lost to scattered provider letters: %+v", got)
	}
}

func TestPi150DoubleClickPathTokens(t *testing.T) {
	for _, token := range []string{"~/src/a+b.go", `\Users\me\file.txt`, "../.config/pi", "a_b-c+d", "~/éclair/é.txt"} {
		c := &chatTUI{transcript: []string{token + " next"}, outputHeight: 20}
		rows := c.renderedTranscriptRows(100)
		for col := 0; col < gotui.StringWidth(token); col++ {
			r, ok := transcriptWordRange(0, rows[0], col)
			if !ok {
				t.Fatalf("%q col%d: no word", token, col)
			}
			s := transcriptSelection{active: true, moved: true, rows: rows, width: 100, start: r.start, end: r.end}
			if got := s.text(); got != token {
				t.Fatalf("%q col%d selected %q", token, col, got)
			}
		}
	}
}
