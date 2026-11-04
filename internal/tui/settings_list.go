package tui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	gotui "github.com/grindlemire/go-tui"
)

// pi-tui's SettingsList (settings-list.js) and Pi's SelectSubmenu
// (settings-submenu.js) and ThemeSubmenu (settings-selector.js), the parts
// /settings uses. Keys arrive by name: up, down, enter, esc (Escape and
// Ctrl+C, Pi's tui.select.cancel), backspace, or a typed rune.

// settingsComponent is a component inside /settings.
type settingsComponent interface {
	rows(width int) spanRows
	handle(key string, r rune)
}

const settingsMaxVisible = 10

type settingItem struct {
	id, label, description, value string
	values                        []string
	// submenu opens a component for the value; done closes it, with the
	// chosen value or nil.
	submenu func(current string, done func(value *string)) settingsComponent
}

// settingsListState is pi-tui's SettingsList.
type settingsListState struct {
	items       []settingItem
	search      bool
	query       string
	selected    int
	submenu     settingsComponent
	submenuItem int
	onChange    func(id, value string)
	onCancel    func()
}

func (s *settingsListState) display() []settingItem {
	if !s.search || s.query == "" {
		return s.items
	}
	wrapped := make([]slashItem, len(s.items))
	for i, it := range s.items {
		wrapped[i] = slashItem{value: strconv.Itoa(i), description: it.label}
	}
	var out []settingItem
	for _, w := range piFuzzyFilter(wrapped, s.query, func(it slashItem) string { return it.description }) {
		i, _ := strconv.Atoi(w.value)
		out = append(out, s.items[i])
	}
	return out
}

func (s *settingsListState) handle(key string, r rune) {
	if s.submenu != nil {
		s.submenu.handle(key, r)
		return
	}
	n := len(s.display())
	switch {
	case key == "up" && n > 0:
		s.selected = (s.selected - 1 + n) % n
	case key == "down" && n > 0:
		s.selected = (s.selected + 1) % n
	case (key == "pgup" || key == "pgdn") && n > 0:
		direction := 1
		if key == "pgup" {
			direction = -1
		}
		s.selected = selectPageIndex(s.selected, n, settingsMaxVisible, direction)
	case key == "enter" || (key == "rune" && r == ' ' && (!s.search || s.query == "")):
		s.activate()
	case key == "esc":
		if s.onCancel != nil {
			s.onCancel()
		}
	case s.search && key == "backspace":
		if q := []rune(s.query); len(q) > 0 {
			s.query, s.selected = string(q[:len(q)-1]), 0
		}
	case s.search && key == "rune":
		s.query, s.selected = s.query+string(r), 0
	}
}

func (s *settingsListState) activate() {
	items := s.display()
	if s.selected >= len(items) {
		return
	}
	id := items[s.selected].id
	i := slices.IndexFunc(s.items, func(it settingItem) bool { return it.id == id })
	it := &s.items[i]
	switch {
	case it.submenu != nil:
		s.submenuItem = s.selected
		s.submenu = it.submenu(it.value, func(value *string) {
			if value != nil {
				it.value = *value
				if s.onChange != nil {
					s.onChange(it.id, *value)
				}
			}
			s.submenu = nil
			s.selected = s.submenuItem
		})
	case len(it.values) > 0:
		it.value = it.values[(slices.Index(it.values, it.value)+1)%len(it.values)]
		if s.onChange != nil {
			s.onChange(it.id, it.value)
		}
	}
}

func (s *settingsListState) rows(width int) spanRows {
	if s.submenu != nil {
		return s.submenu.rows(width)
	}
	var rows spanRows
	if s.search {
		rows = append(rows, piSearchRow(s.query, width), nil)
	}
	hint := func() {
		text := "  Enter/Space to change · Esc to cancel"
		if s.search {
			text = "  Type to search · Enter/Space to change · Esc to cancel"
		}
		rows = append(rows, nil, piTruncate([]gotui.TextSpan{{Text: text, Style: piFg(piDim)}}, width))
	}
	items := s.display()
	if len(items) == 0 {
		rows = append(rows, piTruncate([]gotui.TextSpan{{Text: "  No matching settings", Style: piFg(piDim)}}, width))
		hint()
		return rows
	}
	labelWidth := 0
	for _, it := range s.items {
		labelWidth = max(labelWidth, gotui.StringWidth(it.label))
	}
	labelWidth = min(36, labelWidth)
	start := max(0, min(s.selected-settingsMaxVisible/2, len(items)-settingsMaxVisible))
	end := min(start+settingsMaxVisible, len(items))
	for i := start; i < end; i++ {
		it := items[i]
		prefix, labelStyle, valueStyle := gotui.TextSpan{Text: "  "}, gotui.NewStyle(), piFg(piMuted)
		if i == s.selected {
			prefix, labelStyle, valueStyle = gotui.TextSpan{Text: "→ ", Style: piFg(piAccent)}, piFg(piAccent), piFg(piAccent)
		}
		label := it.label + strings.Repeat(" ", max(0, labelWidth-gotui.StringWidth(it.label)))
		value := truncateCells(it.value, width-2-labelWidth-2-2)
		rows = append(rows, piTruncate([]gotui.TextSpan{prefix, {Text: label, Style: labelStyle}, {Text: "  "}, {Text: value, Style: valueStyle}}, width))
	}
	if start > 0 || end < len(items) {
		rows = append(rows, []gotui.TextSpan{{Text: truncateCells(fmt.Sprintf("  (%d/%d)", s.selected+1, len(items)), width-2), Style: piFg(piDim)}})
	}
	if description := items[s.selected].description; description != "" {
		rows = append(rows, nil)
		for _, line := range piWrapLine([]gotui.TextSpan{{Text: description}}, max(1, width-4)) {
			rows = append(rows, append([]gotui.TextSpan{{Text: "  ", Style: piFg(piDim)}}, restyle(line, piFg(piDim))...))
		}
	}
	hint()
	return rows
}

// selectOption is a SelectList item.
type selectOption struct{ value, label, description string }

// selectSubmenu is Pi's SelectSubmenu (not searchable): a title, a
// description, a SelectList and a hint.
type selectSubmenu struct {
	title, description string
	options            []selectOption
	selected           int
	onSelect           func(value string)
	onCancel           func()
	onSelectionChange  func(value string)
}

func newSelectSubmenu(title, description string, options []selectOption, current string, onSelect func(string), onCancel func(), onSelectionChange func(string)) *selectSubmenu {
	m := &selectSubmenu{title: title, description: description, options: options, onSelect: onSelect, onCancel: onCancel, onSelectionChange: onSelectionChange}
	if i := slices.IndexFunc(options, func(o selectOption) bool { return o.value == current }); i >= 0 {
		m.selected = i
	}
	return m
}

func (m *selectSubmenu) handle(key string, _ rune) {
	n := len(m.options)
	switch {
	case (key == "up" || key == "down") && n > 0:
		delta := 1
		if key == "up" {
			delta = -1
		}
		m.selected = (m.selected + delta + n) % n
		if m.onSelectionChange != nil {
			m.onSelectionChange(m.options[m.selected].value)
		}
	case (key == "pgup" || key == "pgdn") && n > 0:
		direction := 1
		if key == "pgup" {
			direction = -1
		}
		m.selected = selectPageIndex(m.selected, n, 10, direction)
		if m.onSelectionChange != nil {
			m.onSelectionChange(m.options[m.selected].value)
		}
	case key == "enter" && n > 0:
		m.onSelect(m.options[m.selected].value)
	case key == "esc":
		m.onCancel()
	}
}

func (m *selectSubmenu) rows(width int) spanRows {
	rows := piWrapLine([]gotui.TextSpan{{Text: m.title, Style: piFg(piAccent).Bold()}}, max(1, width))
	if m.description != "" {
		rows = append(rows, nil)
		rows = append(rows, piWrapLine([]gotui.TextSpan{{Text: m.description, Style: piFg(piMuted)}}, max(1, width))...)
	}
	rows = append(rows, nil)
	labels := make([]string, len(m.options))
	descriptions := make([]string, len(m.options))
	for i, o := range m.options {
		labels[i], descriptions[i] = o.label, o.description
	}
	rows = append(rows, piSelectRows(labels, descriptions, m.selected, width, min(len(m.options), 10))...)
	rows = append(rows, nil)
	return append(rows, piWrapLine([]gotui.TextSpan{{Text: "  Enter to select · Esc to go back", Style: piFg(piDim)}}, max(1, width))...)
}

// automaticThemeValue is the single-mode item that switches to automatic
// mode (Pi's AUTOMATIC_THEME_VALUE).
const automaticThemeValue = "/"

// themeOptions is Pi's themeItems: a check on the current theme.
func themeOptions(available []string, current string) []selectOption {
	out := make([]selectOption, len(available))
	for i, name := range available {
		mark := "  "
		if name == current {
			mark = "✓ "
		}
		out[i] = selectOption{value: name, label: mark + name}
		if name == systemThemeName {
			out[i].description = "Theme created from your terminal's colors"
		}
	}
	return out
}

// singleModeThemeOptions is Pi's singleModeThemeItems: system, automatic,
// then the other themes.
func singleModeThemeOptions(available []string, current string) []selectOption {
	items := themeOptions(available, current)
	var system []selectOption
	if i := slices.IndexFunc(items, func(o selectOption) bool { return o.value == systemThemeName }); i >= 0 {
		system = []selectOption{items[i]}
		items = slices.Delete(items, i, i+1)
	}
	auto := selectOption{value: automaticThemeValue, label: "  automatic", description: "Use separate themes for light and dark terminal appearance"}
	return append(append(system, auto), items...)
}

func preferredTheme(available []string, preferred, fallback string) string {
	if preferred != "" && slices.Contains(available, preferred) {
		return preferred
	}
	if slices.Contains(available, fallback) || len(available) == 0 {
		return fallback
	}
	return available[0]
}

// parseAutoThemeSetting is Pi's: a "light/dark" pair, or ok false.
func parseAutoThemeSetting(setting string) (light, dark string, ok bool) {
	i := strings.Index(setting, "/")
	if i < 0 || strings.Contains(setting[i+1:], "/") {
		return "", "", false
	}
	light, dark = strings.TrimSpace(setting[:i]), strings.TrimSpace(setting[i+1:])
	return light, dark, light != "" && dark != ""
}

// themeSubmenu is Pi's ThemeSubmenu: one theme, or automatic mode with a
// theme for light and one for dark terminals. Moving previews; Escape
// restores the original setting.
type themeSubmenu struct {
	available                     []string
	terminalTheme                 string
	original                      string
	automatic                     bool
	single, lightTheme, darkTheme string
	preview                       func(setting string)
	done                          func(value *string)
	content                       settingsComponent
}

func newThemeSubmenu(current, terminalTheme string, available []string, preview func(string), done func(*string)) *themeSubmenu {
	t := &themeSubmenu{available: available, terminalTheme: terminalTheme, original: current, preview: preview, done: done}
	light, dark, auto := parseAutoThemeSetting(current)
	if auto {
		t.lightTheme, t.darkTheme = light, dark
	} else {
		fixed := ""
		if !strings.Contains(current, "/") {
			fixed = current
		}
		name := preferredTheme(available, fixed, systemThemeName)
		t.lightTheme, t.darkTheme = name, name
	}
	fixed := ""
	if !auto && !strings.Contains(current, "/") {
		fixed = current
	} else if auto {
		fixed = t.activeAutomatic()
	}
	t.single = preferredTheme(available, fixed, systemThemeName)
	if auto {
		t.showAutomatic()
	} else {
		t.showSingle()
	}
	return t
}

func (t *themeSubmenu) rows(width int) spanRows   { return t.content.rows(width) }
func (t *themeSubmenu) handle(key string, r rune) { t.content.handle(key, r) }

func (t *themeSubmenu) activeAutomatic() string {
	if t.terminalTheme == "light" {
		return t.lightTheme
	}
	return t.darkTheme
}

func (t *themeSubmenu) automaticSetting() string { return t.lightTheme + "/" + t.darkTheme }

func (t *themeSubmenu) setting() string {
	if t.automatic {
		return t.automaticSetting()
	}
	return t.single
}

func (t *themeSubmenu) apply(setting string) { t.done(&setting) }

func (t *themeSubmenu) cancel() {
	t.preview(t.original)
	t.done(nil)
}

func (t *themeSubmenu) showSingle() {
	t.automatic = false
	t.content = newSelectSubmenu("Theme", "Select a theme, or choose automatic to follow terminal appearance.", singleModeThemeOptions(t.available, t.single), t.single, func(value string) {
		if value == automaticThemeValue {
			t.automatic = true
			t.preview(t.setting())
			t.showAutomatic()
			return
		}
		t.single = value
		t.apply(value)
	}, t.cancel, func(value string) {
		if value == automaticThemeValue {
			t.preview(t.automaticSetting())
		} else {
			t.preview(value)
		}
	})
}

// themeAutomaticMenu is the automatic-mode screen: a heading over a
// SettingsList.
type themeAutomaticMenu struct{ list *settingsListState }

func (m *themeAutomaticMenu) handle(key string, r rune) { m.list.handle(key, r) }

func (m *themeAutomaticMenu) rows(width int) spanRows {
	wrap := func(text string, style gotui.Style) spanRows {
		return piWrapLine([]gotui.TextSpan{{Text: text, Style: style}}, max(1, width))
	}
	rows := wrap("Automatic Theme", piFg(piAccent).Bold())
	rows = append(rows, nil)
	rows = append(rows, wrap("Choose themes for terminal light and dark appearance.", piFg(piMuted))...)
	rows = append(rows, wrap("Light/dark detection requires terminal support.", piFg(piMuted))...)
	rows = append(rows, nil)
	return append(rows, m.list.rows(width)...)
}

func (t *themeSubmenu) showAutomatic() {
	t.automatic = true
	themeSelect := func(title, description string, set func(string)) func(string, func(*string)) settingsComponent {
		return func(current string, done func(*string)) settingsComponent {
			return newSelectSubmenu(title, description, themeOptions(t.available, current), current, func(value string) {
				set(value)
				t.preview(t.setting())
				done(&value)
			}, func() {
				t.preview(t.setting())
				done(nil)
			}, t.preview)
		}
	}
	list := &settingsListState{items: []settingItem{
		{id: "light-theme", label: "Light theme", description: "Theme to use in automatic mode when the terminal is light", value: t.lightTheme,
			submenu: themeSelect("Light Theme", "Select the theme to use for light terminal appearance", func(v string) { t.lightTheme = v })},
		{id: "dark-theme", label: "Dark theme", description: "Theme to use in automatic mode when the terminal is dark", value: t.darkTheme,
			submenu: themeSelect("Dark Theme", "Select the theme to use for dark terminal appearance", func(v string) { t.darkTheme = v })},
		{id: "apply", label: "Apply", description: "Save and go back", value: "save and go back", values: []string{"save and go back"}},
		{id: "single-mode", label: "Change mode", description: "Switch to one theme for light and dark", value: "switch to single theme", values: []string{"switch to single theme"}},
	}}
	list.onChange = func(id, _ string) {
		switch id {
		case "single-mode":
			t.automatic = false
			t.single = t.activeAutomatic()
			t.preview(t.single)
			t.showSingle()
		case "apply":
			t.apply(t.automaticSetting())
		}
	}
	list.onCancel = t.cancel
	t.content = &themeAutomaticMenu{list: list}
}

// settingsListKeys binds the key names a settings component handles.
func settingsListKeys(handle func(key string, r rune), dirty func()) gotui.KeyMap {
	on := func(key string) func(gotui.KeyEvent) {
		return func(gotui.KeyEvent) { handle(key, 0); dirty() }
	}
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyUp, on("up")),
		gotui.OnPreemptStop(gotui.KeyDown, on("down")),
		gotui.OnPreemptStop(gotui.KeyPageUp, on("pgup")),
		gotui.OnPreemptStop(gotui.KeyPageDown, on("pgdn")),
		gotui.OnPreemptStop(gotui.KeyEnter, on("enter")),
		gotui.OnPreemptStop(gotui.KeyEscape, on("esc")),
		gotui.OnPreemptStop(gotui.KeyCtrlC, on("esc")),
		gotui.OnPreemptStop(gotui.KeyBackspace, on("backspace")),
		gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { handle("rune", ke.Rune); dirty() }),
	}
}
