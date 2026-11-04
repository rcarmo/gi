package tui

import (
	"fmt"
	"strings"
	"unicode"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
)

// Pi's ModelSelectorComponent replaces the editor while open:
//
//	──────────────────────── (border)
//
//	Scope: all | scoped
//	tab scope (all/scoped)
//
//	> query
//
//	→ ✓ model-id [provider] · default
//	    other-model [provider]
//	  (3/12)
//
//	  Model Name: Friendly Name
//
//	  Enter to select · Ctrl+S to set as default · Escape/Ctrl+C to cancel
//	────────────────────────
//
// Gi keeps its safety checks: models without credentials or with a context
// window smaller than the session are listed dim with the reason and cannot
// be selected.

const piModelSelectorMaxVisible = 10

var piBorder = piRGB(95, 168, 204)

func splitModelLabel(label string) (provider, id string) {
	if i := strings.Index(label, "/"); i > 0 {
		return label[:i], label[i+1:]
	}
	return "", label
}

// modelMenuScopes returns Pi's scoped (enabled models) and all lists.
func (c *chatTUI) modelMenuScopes() (scoped, all []string) {
	all = c.availableModelChoices()
	if !c.cfg.EnabledModelsConfigured {
		return nil, all
	}
	seen := map[string]bool{}
	for _, label := range all {
		seen[label] = true
	}
	for _, model := range c.cfg.EnabledModels {
		key := canonicalModelRef(c.modelDefaults().Provider, strings.TrimSpace(model))
		if key != "" && seen[key] {
			scoped = append(scoped, key)
			seen[key] = false
		}
	}
	return scoped, all
}

func (c *chatTUI) toggleModelMenuScope() {
	if c.modelMenuKind != "model" {
		return
	}
	scoped, all := c.modelMenuScopes()
	if len(scoped) == 0 {
		return
	}
	if c.modelMenuScope == "scoped" {
		c.modelMenuScope, c.modelMenuAll = "all", all
	} else {
		c.modelMenuScope, c.modelMenuAll = "scoped", scoped
	}
	c.applyModelMenuFilter()
	current := canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel)
	for i, label := range c.modelMenuChoices {
		if label == current && c.modelPickerUnavailable(label) == "" {
			c.modelMenuSelected = i
		}
	}
	c.markDirty()
}

// acceptModelMenuAsDefault is Pi's Ctrl+S: select, then save as the default.
func (c *chatTUI) acceptModelMenuAsDefault() {
	if c.modelMenuKind != "model" {
		return
	}
	c.acceptModelMenuSelection()
	if c.modelMenuOpen {
		return // selection was refused; the error is shown in the selector
	}
	if err := config.PersistModelSelection(c.cfg.WorkspaceRoot, c.cfg.DefaultProvider, c.cfg.DefaultModel, c.cfg.DefaultThinkingLevel, c.cfg.EnabledModels); err != nil {
		c.appendTranscript(fmt.Sprintf("warn: failed to save default model: %v", err))
	}
}

func (c *chatTUI) piModelVisibleRange() (int, int) {
	n := len(c.modelMenuChoices)
	visible := min(piModelSelectorMaxVisible, c.piModelSelectorListRows())
	start := max(0, min(c.modelMenuSelected-visible/2, n-visible))
	return start, min(start+visible, n)
}

// piModelSelectorListRows bounds the list so the selector, spacer and footer
// fit the terminal, like Pi's 10-row list on smaller screens.
func (c *chatTUI) piModelSelectorListRows() int {
	width, height := c.outputWidth, c.outputHeight
	if c.app != nil {
		width, height = c.app.Size()
	}
	if height == 0 {
		height = 24
	}
	if width == 0 {
		width = 80
	}
	chrome := 14 // borders, spacers, scope, search, scroll info, name, hint
	if !c.regularMode {
		chrome += 1 + len(c.footerLines(width)) // spacer above, footer
	}
	return max(1, min(piModelSelectorMaxVisible, height-chrome))
}

func (c *chatTUI) piModelSelectorRows(width int) [][]gotui.TextSpan {
	muted, dim, accent := piFg(piMuted), piFg(piDim), piFg(piAccent)
	rule := []gotui.TextSpan{{Text: strings.Repeat("─", max(1, width)), Style: piFg(piBorder)}}
	blank := []gotui.TextSpan{}
	rows := [][]gotui.TextSpan{rule, blank}
	if scoped, _ := c.modelMenuScopes(); len(scoped) > 0 {
		scope := func(name string) gotui.TextSpan {
			if c.modelMenuScope == name {
				return gotui.TextSpan{Text: name, Style: accent}
			}
			return gotui.TextSpan{Text: name, Style: muted}
		}
		rows = append(rows,
			[]gotui.TextSpan{{Text: "Scope: ", Style: muted}, scope("all"), {Text: " | ", Style: muted}, scope("scoped")},
			[]gotui.TextSpan{{Text: "tab", Style: dim}, {Text: " scope", Style: muted}, {Text: " (all/scoped)", Style: muted}})
	} else {
		rows = append(rows, []gotui.TextSpan{{Text: truncateCells("Only showing models from configured providers. Use /login to add providers.", width), Style: piFg(piWarning)}})
	}
	rows = append(rows, blank, []gotui.TextSpan{{Text: "> " + truncateCells(c.modelMenuQuery, max(0, width-3))}, {Text: " ", Style: gotui.NewStyle().Reverse()}}, blank)

	current := canonicalModelRef(c.cfg.DefaultProvider, c.cfg.DefaultModel)
	start, end := c.piModelVisibleRange()
	for i := start; i < end; i++ {
		label := c.modelMenuChoices[i]
		provider, id := splitModelLabel(label)
		selected := i == c.modelMenuSelected
		reason := c.modelPickerUnavailable(label)
		row := []gotui.TextSpan{{Text: "  "}, {Text: "  "}}
		if selected {
			row[0] = gotui.TextSpan{Text: "→ ", Style: accent}
		}
		if label == current {
			row[1] = gotui.TextSpan{Text: "✓ ", Style: accent}
		}
		idStyle := gotui.NewStyle()
		switch {
		case reason != "":
			idStyle = dim
		case selected:
			idStyle = accent
		}
		row = append(row, gotui.TextSpan{Text: id, Style: idStyle})
		if name := distinctModelName(id, c.modelMenuName(label)); name != "" {
			// An unavailable row's safety reason outranks the optional alias
			// on narrow screens. The detail slot still shows its full name.
			suffix := " [" + provider + "]"
			if label == c.modelMenuDefault {
				suffix += " · default"
			}
			if reason != "" {
				suffix += " · " + reason
			}
			if reason == "" || gotui.StringWidth("    "+id+" ("+name+")"+suffix) <= width {
				row = append(row, gotui.TextSpan{Text: " (" + name + ")", Style: muted})
			}
		}
		if provider != "" {
			row = append(row, gotui.TextSpan{Text: " [" + provider + "]", Style: muted})
		}
		if label == c.modelMenuDefault {
			row = append(row, gotui.TextSpan{Text: " · default", Style: muted})
		}
		if reason != "" {
			row = append(row, gotui.TextSpan{Text: " · " + reason, Style: muted})
		}
		rows = append(rows, clipSpans(row, width))
	}
	if start > 0 || end < len(c.modelMenuChoices) {
		rows = append(rows, []gotui.TextSpan{{Text: fmt.Sprintf("  (%d/%d)", c.modelMenuSelected+1, len(c.modelMenuChoices)), Style: muted}})
	}
	// One status slot under the list keeps the selector height stable: the
	// selected model's name, or (in Pi's error colour) why selection failed.
	switch {
	case c.modelMenuError != "":
		line := strings.SplitN(c.modelMenuError, "\n", 2)[0]
		rows = append(rows, blank, []gotui.TextSpan{{Text: truncateCells("  "+line, width), Style: piFg(piError)}})
	case len(c.modelMenuChoices) == 0:
		rows = append(rows, []gotui.TextSpan{{Text: "  No matching models", Style: muted}}, blank)
	case c.modelMenuSelected >= 0 && c.modelMenuSelected < len(c.modelMenuChoices):
		name := c.modelMenuName(c.modelMenuChoices[c.modelMenuSelected])
		rows = append(rows, blank, []gotui.TextSpan{{Text: truncateCells("  Model Name: "+name, width), Style: muted}})
	default:
		rows = append(rows, blank, blank)
	}
	rows = append(rows, blank,
		[]gotui.TextSpan{{Text: truncateCells("  Enter to select · Ctrl+S to set as default · Escape/Ctrl+C to cancel", width), Style: dim}},
		rule)
	return rows
}

// distinctModelName follows Pi 1.5's getDistinctModelName: only show names
// that add words, not differences in casing, separators or version suffixes.
func distinctModelName(id, name string) string {
	words := func(text string) []string {
		return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
	}
	idWords := words(id)
	tokens := make(map[string]bool, len(idWords))
	for _, word := range idWords {
		tokens[word] = true
	}
	compact := strings.Join(idWords, "")
	for _, word := range words(name) {
		if !tokens[word] && !(len([]rune(word)) > 1 && strings.Contains(compact, word)) {
			return strings.TrimSpace(name)
		}
	}
	return ""
}

// modelSelectorSearchText keeps the provider before the bare ID, so a direct
// provider/id query ranks ahead of the same ID through a proxy provider.
func modelSelectorSearchText(provider, id, name string) string {
	return fmt.Sprintf("%s %s/%s %s %s %s", provider, provider, id, provider, id, name)
}

func (c *chatTUI) modelMenuName(label string) string {
	for _, option := range c.sessionModelCatalogue() {
		if option.Label == label && strings.TrimSpace(option.Name) != "" {
			return option.Name
		}
	}
	_, id := splitModelLabel(label)
	return id
}

func clipSpans(spans []gotui.TextSpan, width int) []gotui.TextSpan {
	out := make([]gotui.TextSpan, 0, len(spans))
	used := 0
	for _, span := range spans {
		if used >= width {
			break
		}
		text := truncateCells(span.Text, width-used)
		used += gotui.StringWidth(text)
		span.Text = text
		out = append(out, span)
	}
	return out
}

func (c *chatTUI) renderPiModelSelector(width int) *gotui.Element {
	rows := c.piModelSelectorRows(width)
	block := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(len(rows)))
	for _, spans := range rows {
		block.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithWrap(false), gotui.WithRichText(spans...)))
	}
	return block
}

// modelMenuDefaultLabel is the model saved in settings (Pi's "default" badge).
func (c *chatTUI) modelMenuDefaultLabel() string {
	saved := config.Load(c.cfg.WorkspaceRoot)
	if strings.TrimSpace(saved.DefaultModel) == "" {
		return ""
	}
	return canonicalModelRef(saved.DefaultProvider, (inference.SessionModelChoice{Model: saved.DefaultModel, Provider: saved.DefaultProvider}).Label())
}
