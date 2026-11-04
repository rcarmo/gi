package tui

import (
	"fmt"
	"slices"
	"strings"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/config"
)

// /scoped-models is Pi's ScopedModelsSelectorComponent (pi-coding-agent
// components/scoped-models-selector.js): enable, disable and order the
// models that Ctrl+P cycles. Changes apply to this run at once; Ctrl+S saves
// them as enabledModels. Golden: scripts/golden-scoped-models.mjs.

const scopedModelsMaxVisible = 8

type scopedModelsState struct {
	all      []string          // available models' full ids, in catalogue order
	names    map[string]string // full id -> model name
	enabled  []string          // Pi's enabledIds; all enabled when allOn
	allOn    bool              // Pi's null: every model enabled
	query    string
	selected int
	dirty    bool
}

func (s *scopedModelsState) isEnabled(id string) bool {
	return s.allOn || slices.Contains(s.enabled, id)
}

func (s *scopedModelsState) available(id string) bool {
	_, ok := s.names[id]
	return ok
}

// set takes Pi's normalizeEnabled: a list naming every model is null.
func (s *scopedModelsState) set(ids []string, normalize bool) {
	if normalize && len(ids) == len(s.all) && !slices.ContainsFunc(ids, func(id string) bool { return !slices.Contains(s.all, id) }) {
		s.enabled, s.allOn = nil, true
		return
	}
	s.enabled, s.allOn = ids, false
}

func (s *scopedModelsState) toggle(id string) {
	switch {
	case s.allOn:
		s.set(slices.DeleteFunc(slices.Clone(s.all), func(m string) bool { return m == id }), false)
	case slices.Contains(s.enabled, id):
		s.set(slices.DeleteFunc(slices.Clone(s.enabled), func(m string) bool { return m == id }), false)
	default:
		s.set(append(slices.Clone(s.enabled), id), true)
	}
}

// enableAll and clearAll are Pi's, over targets (nil: every model).
func (s *scopedModelsState) enableAll(targets []string) {
	if s.allOn {
		return
	}
	if targets == nil {
		targets = s.all
	}
	ids := slices.Clone(s.enabled)
	for _, id := range targets {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	s.set(ids, true)
}

func (s *scopedModelsState) clearAll(targets []string) {
	if s.allOn {
		if targets == nil {
			s.set([]string{}, false)
		} else {
			s.set(slices.DeleteFunc(slices.Clone(s.all), func(id string) bool { return slices.Contains(targets, id) }), false)
		}
		return
	}
	if targets == nil {
		targets = s.enabled
	}
	s.set(slices.DeleteFunc(slices.Clone(s.enabled), func(id string) bool { return slices.Contains(targets, id) }), false)
}

// items are Pi's sorted ids (enabled first, in order), fuzzy filtered.
func (s *scopedModelsState) items() []string {
	ids := s.all
	if !s.allOn {
		ids = slices.Clone(s.enabled)
		for _, id := range s.all {
			if !slices.Contains(s.enabled, id) {
				ids = append(ids, id)
			}
		}
	}
	if strings.TrimSpace(s.query) == "" {
		return ids
	}
	items := make([]slashItem, len(ids))
	for i, id := range ids {
		text := id
		if s.available(id) {
			provider, model := splitModelLabel(id)
			text = modelSearchText(provider, model, s.names[id])
		}
		items[i] = slashItem{name: id, description: text}
	}
	var out []string
	for _, it := range piFuzzyFilter(items, s.query, func(it slashItem) string { return it.description }) {
		out = append(out, it.name)
	}
	return out
}

// modelSearchText is Pi's getModelSearchText.
func modelSearchText(provider, id, name string) string {
	if name != "" {
		name = " " + name
	}
	return fmt.Sprintf("%s %s %s/%s %s %s%s", id, provider, provider, id, provider, id, name)
}

func (c *chatTUI) openScopedModels() {
	var all []string
	names := map[string]string{}
	for _, option := range c.sessionModelCatalogue() {
		id := canonicalModelRef(c.modelDefaults().Provider, option.Label)
		if id == "" || slices.Contains(all, id) {
			continue
		}
		all = append(all, id)
		names[id] = strings.TrimSpace(option.Name)
		if names[id] == "" {
			_, names[id] = splitModelLabel(id)
		}
	}
	s := &scopedModelsState{all: all, names: names, allOn: true}
	if c.cfg.EnabledModelsConfigured {
		var ids []string
		for _, model := range c.cfg.EnabledModels {
			if id := canonicalModelRef(c.modelDefaults().Provider, strings.TrimSpace(model)); id != "" && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		s.enabled, s.allOn = ids, false
	}
	c.scopedModels = s
	c.modelMenuOpen = true
	c.modelMenuKind = "scoped-models"
	c.modelMenuError = ""
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

// applyScopedModels is Pi's updateSessionModels: the scope holds the enabled
// models unless none is available or every available one is enabled.
func (c *chatTUI) applyScopedModels() {
	s := c.scopedModels
	anyAvailable := slices.ContainsFunc(s.enabled, s.available)
	every := !slices.ContainsFunc(s.all, func(id string) bool { return !slices.Contains(s.enabled, id) })
	if s.allOn || !anyAvailable || every {
		c.cfg.EnabledModelsConfigured = false
		return
	}
	c.cfg.EnabledModels = slices.Clone(s.enabled)
	c.cfg.EnabledModelsConfigured = true
}

func (c *chatTUI) scopedModelsKeys() gotui.KeyMap {
	s := c.scopedModels
	changed := func() {
		s.dirty = true
		s.selected = min(s.selected, max(0, len(s.items())-1))
		c.applyScopedModels()
		c.markDirty()
	}
	current := func() (string, bool) {
		items := s.items()
		if s.selected < len(items) {
			return items[s.selected], true
		}
		return "", false
	}
	targets := func() []string {
		if s.query == "" {
			return nil
		}
		return s.items()
	}
	move := func(delta int) {
		if n := len(s.items()); n > 0 {
			s.selected = (s.selected + delta + n) % n
			c.markDirty()
		}
	}
	reorder := func(delta int) {
		id, ok := current()
		if s.allOn || !ok || !s.isEnabled(id) {
			return
		}
		i := slices.Index(s.enabled, id)
		if j := i + delta; j >= 0 && j < len(s.enabled) {
			s.enabled[i], s.enabled[j] = s.enabled[j], s.enabled[i]
			s.selected += delta
			changed()
		}
	}
	search := func(query string) {
		s.query = query
		s.selected = min(s.selected, max(0, len(s.items())-1))
		c.markDirty()
	}
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyUp, func(gotui.KeyEvent) { move(-1) }),
		gotui.OnPreemptStop(gotui.KeyDown, func(gotui.KeyEvent) { move(1) }),
		gotui.OnPreemptStop(gotui.KeyPageUp, func(gotui.KeyEvent) {
			s.selected = selectPageIndex(s.selected, len(s.items()), scopedModelsMaxVisible, -1)
			c.markDirty()
		}),
		gotui.OnPreemptStop(gotui.KeyPageDown, func(gotui.KeyEvent) {
			s.selected = selectPageIndex(s.selected, len(s.items()), scopedModelsMaxVisible, 1)
			c.markDirty()
		}),
		gotui.OnPreemptStop(gotui.KeyUp.Alt(), func(gotui.KeyEvent) { reorder(-1) }),
		gotui.OnPreemptStop(gotui.KeyDown.Alt(), func(gotui.KeyEvent) { reorder(1) }),
		gotui.OnPreemptStop(gotui.KeyEnter, func(gotui.KeyEvent) {
			if id, ok := current(); ok {
				s.toggle(id)
				changed()
			}
		}),
		gotui.OnPreemptStop(gotui.Rune('a').Ctrl(), func(gotui.KeyEvent) { s.enableAll(targets()); changed() }),
		gotui.OnPreemptStop(gotui.Rune('x').Ctrl(), func(gotui.KeyEvent) { s.clearAll(targets()); changed() }),
		gotui.OnPreemptStop(gotui.Rune('p').Ctrl(), func(gotui.KeyEvent) {
			id, ok := current()
			if !ok || !s.available(id) {
				return
			}
			provider, _ := splitModelLabel(id)
			var ids []string
			for _, m := range s.all {
				if p, _ := splitModelLabel(m); p == provider {
					ids = append(ids, m)
				}
			}
			if !slices.ContainsFunc(ids, func(id string) bool { return !s.isEnabled(id) }) {
				s.clearAll(ids)
			} else {
				s.enableAll(ids)
			}
			changed()
		}),
		gotui.OnPreemptStop(gotui.Rune('s').Ctrl(), func(gotui.KeyEvent) { c.saveScopedModels() }),
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) {
			if s.query != "" {
				search("")
			} else {
				c.closeModelMenu()
			}
		}),
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.closeModelMenu() }),
		gotui.OnPreemptStop(gotui.KeyBackspace, func(gotui.KeyEvent) {
			if r := []rune(s.query); len(r) > 0 {
				search(string(r[:len(r)-1]))
			}
		}),
		gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { search(s.query + string(ke.Rune)) }),
	}
}

// saveScopedModels is Pi's onPersist: enabledModels, or none when every
// available model is enabled.
func (c *chatTUI) saveScopedModels() {
	s := c.scopedModels
	ids := slices.Clone(s.enabled)
	if s.allOn || len(ids) == len(s.all) && !slices.ContainsFunc(ids, func(id string) bool { return !s.available(id) }) {
		ids = nil
	}
	if err := config.PersistEnabledModels(c.cfg.WorkspaceRoot, ids); err != nil {
		c.selectionNotice(fmt.Sprintf("Could not save model selection: %v", err))
		return
	}
	s.dirty = false
	c.selectionNotice("Model selection saved to settings")
	c.markDirty()
}

func (c *chatTUI) piScopedModelsRows(width int) spanRows {
	s := c.scopedModels
	keys := defaultPiKeys
	muted, accent := piFg(piMuted), piFg(piAccent)
	text := func(spans ...gotui.TextSpan) spanRows { return piWrapLine(spans, max(1, width)) }
	rows := spanRows{piRule(width, piBorder), nil}
	rows = append(rows, text(gotui.TextSpan{Text: "Model Configuration", Style: accent.Bold()})...)
	rows = append(rows, text(gotui.TextSpan{Text: "Session-only. " + keys.display("Ctrl+S") + " to save to settings.", Style: muted})...)
	rows = append(rows, nil, piSearchRow(s.query, width), nil)

	items := s.items()
	if len(items) == 0 {
		rows = append(rows, text(gotui.TextSpan{Text: "  No matching models", Style: muted})...)
	} else {
		start := max(0, min(s.selected-scopedModelsMaxVisible/2, len(items)-scopedModelsMaxVisible))
		end := min(start+scopedModelsMaxVisible, len(items))
		for i := start; i < end; i++ {
			id := items[i]
			line := []gotui.TextSpan{{Text: "  "}}
			if i == s.selected {
				line[0] = gotui.TextSpan{Text: "→ ", Style: accent}
			}
			if s.available(id) && s.isEnabled(id) {
				line = append(line, gotui.TextSpan{Text: "✓ ", Style: accent})
			} else {
				line = append(line, gotui.TextSpan{Text: "  "})
			}
			provider, name := splitModelLabel(id)
			style, badge := gotui.NewStyle(), " ["+provider+"]"
			if !s.available(id) {
				name, style, badge = id, style.Strikethrough(), " [unavailable]"
			}
			if i == s.selected {
				style = style.Foreground(piAccent)
			}
			line = append(line, gotui.TextSpan{Text: name, Style: style})
			if s.available(id) {
				if distinct := distinctModelName(name, s.names[id]); distinct != "" {
					line = append(line, gotui.TextSpan{Text: " (" + distinct + ")", Style: muted})
				}
			}
			line = append(line, gotui.TextSpan{Text: badge, Style: muted})
			rows = append(rows, text(line...)...)
		}
		if start > 0 || end < len(items) {
			rows = append(rows, text(gotui.TextSpan{Text: fmt.Sprintf("  (%d/%d)", s.selected+1, len(items)), Style: muted})...)
		}
		detail := "  Model unavailable"
		if id := items[s.selected]; s.available(id) {
			detail = "  Model Name: " + s.names[id]
		}
		rows = append(rows, nil)
		rows = append(rows, text(gotui.TextSpan{Text: detail, Style: muted})...)
	}

	count := "all enabled"
	if !s.allOn {
		enabled, gone := 0, 0
		for _, id := range s.enabled {
			if s.available(id) {
				enabled++
			} else {
				gone++
			}
		}
		count = fmt.Sprintf("%d/%d enabled", enabled, len(s.all))
		if gone > 0 {
			count += fmt.Sprintf(" · %d unavailable", gone)
		}
	}
	footer := "  " + strings.Join([]string{
		keys.display("Enter") + " toggle",
		keys.display("Ctrl+A") + " all",
		keys.display("Ctrl+X") + " clear",
		keys.display("Ctrl+P") + " provider",
		keys.display("Alt+Up") + "/" + keys.display("Alt+Down") + " reorder",
		keys.display("Ctrl+S") + " save",
		count,
	}, " · ")
	footerSpans := []gotui.TextSpan{{Text: footer, Style: piFg(piDim)}}
	if s.dirty {
		footerSpans = []gotui.TextSpan{{Text: footer + " ", Style: piFg(piDim)}, {Text: "(unsaved)", Style: piFg(piWarning)}}
	}
	rows = append(rows, nil)
	rows = append(rows, text(footerSpans...)...)
	return append(rows, piRule(width, piBorder))
}
