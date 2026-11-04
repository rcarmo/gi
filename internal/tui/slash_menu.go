package tui

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
)

// Pi's editor opens a slash-command autocomplete list when "/" is typed at
// the start of the message (pi-tui Editor + CombinedAutocompleteProvider +
// SelectList). It is rendered under the editor's bottom border and fuzzy
// filters as the command name is typed. After the first space it lists the
// command's argument completions (Pi's getArgumentCompletions), or closes
// when the command has none. Typing a letter, digit or ".-_" in a slash
// command, or deleting, opens it again.
//
//	Up/Down  move the selection (wrapping)
//	Tab      complete to "/name " (or the argument)
//	Enter    complete to "/name " and submit; complete an argument only
//	Escape   close the list

const (
	slashMaxVisible     = 5  // Editor autocompleteMaxVisible default
	slashMinPrimaryCol  = 12 // SLASH_COMMAND_SELECT_LIST_LAYOUT
	slashMaxPrimaryCol  = 32
	slashPrimaryGap     = 2
	slashMinDescription = 10
)

// slashItem is a command (name) or an argument completion (name is the
// label, value the text that replaces the typed arguments).
type slashItem struct {
	name, description string
	value             string
}

type slashMenu struct {
	active   bool
	items    []slashItem
	selected int
	argument bool // items complete the command's arguments
	applying bool // a completion is being applied; it never reopens the list
}

var slashArgHint = regexp.MustCompile(`^/([^\s]+)\s*(.*)$`)

// slashCommandItems lists commands as Pi's provider does: built-ins in
// catalogue order, then extension commands, then skills ("skill:name").
// Descriptions are "<hint> — <description>" when a command takes arguments.
func (c *chatTUI) slashCommandItems() []slashItem {
	items := make([]slashItem, 0, len(tuiCommands)+8)
	seen := map[string]bool{}
	add := func(name, hint, desc string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		full := desc
		if hint = strings.TrimSpace(hint); hint != "" {
			if desc != "" {
				full = hint + " — " + desc
			} else {
				full = hint
			}
		}
		items = append(items, slashItem{name: name, description: full})
	}
	for _, cmd := range tuiCommands {
		m := slashArgHint.FindStringSubmatch(cmd.name)
		if m == nil || strings.HasPrefix(m[1], "skill:") {
			continue
		}
		add(m[1], m[2], cmd.hint)
	}
	if c.engine != nil {
		for _, cmd := range c.engine.ExtensionCommandInfos() {
			add(cmd.Name, "", strings.TrimSpace(cmd.Description))
		}
	}
	for _, skill := range c.cfg.Discovery.Skills {
		add("skill:"+skill.Name, "", strings.TrimSpace(skill.Description))
	}
	return items
}

// piFuzzyMatch ports pi-tui fuzzy.js fuzzyMatch (lower score is better).
func piFuzzyMatch(query, text string) (bool, float64) {
	q, t := []rune(strings.ToLower(query)), []rune(strings.ToLower(text))
	match := func(q []rune) (bool, float64) {
		if len(q) == 0 {
			return true, 0
		}
		if len(q) > len(t) {
			return false, 0
		}
		positions := make([]int, 0, len(q))
		last, qi := -1, 0
		for qi < len(q) {
			i := -1
			for j := last + 1; j < len(t); j++ {
				if t[j] == q[qi] {
					i = j
					break
				}
			}
			if i == -1 {
				break
			}
			positions = append(positions, i)
			last = i
			qi++
		}
		if qi < len(q) {
			return false, 0
		}
		scorePositions := func(positions []int) float64 {
			score, last, consecutive := 0.0, -1, 0
			for _, i := range positions {
				if last == i-1 {
					consecutive++
					score -= float64(consecutive * 5)
				} else {
					consecutive = 0
					if last >= 0 {
						score += float64((i - last - 1) * 2)
					}
				}
				if i == 0 || unicode.IsSpace(t[i-1]) || strings.ContainsRune("-_./:", t[i-1]) {
					score -= 10
				}
				score += float64(i) * 0.1
				last = i
			}
			return score
		}
		score := scorePositions(positions)
		// Pi 1.5 also scores contiguous occurrences: a scattered match in a
		// provider prefix must not outrank the actual model name later on.
		for start := 0; start+len(q) <= len(t); start++ {
			if !slices.Equal(t[start:start+len(q)], q) {
				continue
			}
			for k := range positions {
				positions[k] = start + k
			}
			if candidate := scorePositions(positions); candidate < score {
				score = candidate
			}
		}
		if string(q) == string(t) {
			score -= 100
		}
		return true, score
	}
	if ok, score := match(q); ok {
		return true, score
	}
	if m := regexp.MustCompile(`^([a-z]+)([0-9]+)$`).FindStringSubmatch(string(q)); m != nil {
		if ok, score := match([]rune(m[2] + m[1])); ok {
			return true, score + 5
		}
	} else if m := regexp.MustCompile(`^([0-9]+)([a-z]+)$`).FindStringSubmatch(string(q)); m != nil {
		if ok, score := match([]rune(m[2] + m[1])); ok {
			return true, score + 5
		}
	}
	return false, 0
}

// piFuzzyFilter ports fuzzyFilter: every whitespace/slash token must match;
// results are ordered by total score (stable for ties).
func piFuzzyFilter(items []slashItem, query string, text func(slashItem) string) []slashItem {
	tokens := strings.FieldsFunc(strings.TrimSpace(query), func(r rune) bool { return r == '/' || unicode.IsSpace(r) })
	if len(tokens) == 0 {
		return items
	}
	type scored struct {
		item  slashItem
		score float64
	}
	var results []scored
	for _, item := range items {
		total, all := 0.0, true
		for _, token := range tokens {
			ok, score := piFuzzyMatch(token, text(item))
			if !ok {
				all = false
				break
			}
			total += score
		}
		if all {
			results = append(results, scored{item, total})
		}
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].score < results[j].score })
	out := make([]slashItem, len(results))
	for i, r := range results {
		out[i] = r.item
	}
	return out
}

func (c *chatTUI) slashMatches(prefix string) []slashItem {
	items := c.slashCommandItems()
	bare := piFuzzyFilter(items, prefix, func(it slashItem) string { return strings.TrimPrefix(it.name, "skill:") })
	inBare := map[string]bool{}
	for _, it := range bare {
		inBare[it.name] = true
	}
	var skillsOnly []slashItem
	for _, it := range items {
		if strings.HasPrefix(it.name, "skill:") && !inBare[it.name] {
			skillsOnly = append(skillsOnly, it)
		}
	}
	return append(bare, piFuzzyFilter(skillsOnly, prefix, func(it slashItem) string { return it.name })...)
}

// slashContext returns the text before the cursor, without leading
// whitespace, when the cursor is on the first line and that text is a slash
// command (pi-tui isInSlashCommandContext).
func slashContext(text string, cursor int) (string, bool) {
	runes := []rune(text)
	cursor = max(0, min(cursor, len(runes)))
	before := string(runes[:cursor])
	if strings.Contains(before, "\n") {
		return "", false
	}
	before = strings.TrimLeft(before, " \t")
	return before, strings.HasPrefix(before, "/")
}

// slashSuggestions ports CombinedAutocompleteProvider.getSuggestions for
// slash commands: command names before the first space, then the command's
// argument completions for the text after it.
func (c *chatTUI) slashSuggestions(text string, cursor int) (items []slashItem, argument bool) {
	before, ok := slashContext(text, cursor)
	if !ok {
		return nil, false
	}
	space := strings.Index(before, " ")
	if space < 0 {
		return c.slashMatches(before[1:]), false
	}
	return c.slashArgumentCompletions(before[1:space], before[space+1:]), true
}

// slashArgumentCompletions is Pi's command.getArgumentCompletions; nil when
// the command has none or nothing matches.
func (c *chatTUI) slashArgumentCompletions(command, args string) []slashItem {
	switch command {
	case "mcp":
		return c.mcpArgumentCompletions(args)
	case "model":
		return c.modelArgumentCompletions(args)
	case "thinking":
		return fuzzyCompletions(c.thinkingMenuLevels(), args, func(level string) string { return level }, func(level string) slashItem {
			return slashItem{name: level, value: level}
		})
	case "login":
		return loginArgumentCompletions(inference.LoginOptions(""), args)
	}
	return nil
}

// fuzzyCompletions is Pi's createFuzzyAutocompleteItems.
func fuzzyCompletions[T any](items []T, prefix string, text func(T) string, item func(T) slashItem) []slashItem {
	wrapped := make([]slashItem, len(items))
	for i, it := range items {
		wrapped[i] = slashItem{value: strconv.Itoa(i), description: text(it)}
	}
	var out []slashItem
	for _, w := range piFuzzyFilter(wrapped, prefix, func(it slashItem) string { return it.description }) {
		i, _ := strconv.Atoi(w.value)
		out = append(out, item(items[i]))
	}
	return out
}

// modelArgumentCompletions is Pi's /model completion: the scoped models,
// else every available one, as provider/id.
func (c *chatTUI) modelArgumentCompletions(prefix string) []slashItem {
	models, all := c.modelMenuScopes()
	if len(models) == 0 {
		models = all
	}
	return fuzzyCompletions(models, prefix, func(label string) string {
		provider, id := splitModelLabel(label)
		return modelSearchText(provider, id, c.modelMenuName(label))
	}, func(label string) slashItem {
		provider, id := splitModelLabel(label)
		return slashItem{name: id, value: label, description: provider}
	})
}

// loginArgumentCompletions is Pi's /login completion: one item per
// provider, with its sign-in methods.
func loginArgumentCompletions(options []inference.LoginOption, prefix string) []slashItem {
	type provider struct {
		id, name     string
		types        []string
		subscription bool
	}
	var providers []*provider
	byID := map[string]*provider{}
	for _, o := range options {
		if p := byID[o.ID]; p != nil {
			if !slices.Contains(p.types, o.AuthType) {
				p.types = append(p.types, o.AuthType)
				slices.SortFunc(p.types, func(a, b string) int { return strings.Compare(b, a) }) // oauth, then api_key
			}
			continue
		}
		p := &provider{id: o.ID, name: o.Name, types: []string{o.AuthType}, subscription: o.Subscription}
		byID[o.ID] = p
		providers = append(providers, p)
	}
	slices.SortStableFunc(providers, func(a, b *provider) int { return inference.LocaleCompare(a.name, b.name) })
	labels := func(p *provider, sep string, withType bool) string {
		parts := make([]string, len(p.types))
		for i, t := range p.types {
			parts[i] = authTypeLabel(t, p.subscription)
			if withType {
				parts[i] = t + " " + parts[i]
			}
		}
		return strings.Join(parts, sep)
	}
	return fuzzyCompletions(providers, prefix, func(p *provider) string {
		return p.id + " " + p.name + " " + labels(p, " ", true)
	}, func(p *provider) slashItem {
		description := labels(p, "/", false)
		if p.name != p.id {
			description = p.name + " · " + description
		}
		return slashItem{name: p.id, value: p.id, description: description}
	})
}

// slashLead is the whitespace before the command, kept when a completion is
// applied (Pi's beforePrefix).
func slashLead(text string) string {
	return text[:len(text)-len(strings.TrimLeft(text, " \t"))]
}

var slashTriggerChar = regexp.MustCompile(`^[a-zA-Z0-9.\-_]$`)

// cjkBreak is pi-tui's cjkBreakRegex by script; Go has no Script_Extensions,
// so CJK punctuation shared between scripts (、, ー) does not match.
var cjkBreak = regexp.MustCompile(`^[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}\p{Bopomofo}]$`)

// updateSlashMenu follows pi-tui's Editor: an open list is refreshed on
// every edit and closes when nothing matches. A closed list opens when "/"
// is typed at the start of the message, when a letter, digit or ".-_" (or
// CJK text) is typed in a slash command, or when a character is deleted in
// one. Programmatic text changes (history, drafts, paste) never open it.
func (c *chatTUI) updateSlashMenu(previous string) {
	if c.input == nil {
		return
	}
	text := c.input.Text()
	cursor := c.input.cursorPos
	if c.slash.applying {
		return
	}
	if !c.slash.active {
		before, inSlash := slashContext(text, cursor)
		typed, deleted := editedRune(previous, text, cursor)
		switch {
		case typed == "/" && strings.TrimSpace(strings.TrimSuffix(before, "/")) == "" && inSlash:
		case typed != "" && inSlash && (slashTriggerChar.MatchString(typed) || cjkBreak.MatchString(typed)):
		case deleted && inSlash:
		default:
			return
		}
	}
	items, argument := c.slashSuggestions(text, cursor)
	if len(items) == 0 {
		c.slash = slashMenu{}
		return
	}
	c.slash = slashMenu{active: true, items: items, argument: argument}
}

// editedRune reports a single rune typed just before the cursor, or a single
// rune deleted, between previous and text.
func editedRune(previous, text string, cursor int) (typed string, deleted bool) {
	prev, next := []rune(previous), []rune(text)
	switch {
	case len(next) == len(prev)+1 && cursor >= 1 && cursor <= len(next) &&
		string(next[:cursor-1])+string(next[cursor:]) == previous:
		return string(next[cursor-1]), false
	case len(next) == len(prev)-1:
		return "", true
	}
	return "", false
}

func (c *chatTUI) moveSlashSelection(delta int) {
	n := len(c.slash.items)
	if n == 0 {
		return
	}
	c.slash.selected = ((c.slash.selected+delta)%n + n) % n
	c.markDirty()
}

// applySlashSelection ports CombinedAutocompleteProvider.applyCompletion.
// A command becomes "/name " (and Enter submits it); an argument completion
// replaces the arguments typed before the cursor and never submits.
func (c *chatTUI) applySlashSelection(submit bool) {
	if !c.slash.active || len(c.slash.items) == 0 {
		return
	}
	item := c.slash.items[c.slash.selected]
	argument := c.slash.argument
	runes := []rune(c.input.Text())
	cursor := max(0, min(c.input.cursorPos, len(runes)))
	after := string(runes[cursor:])
	before, _ := slashContext(c.input.Text(), cursor)
	c.slash = slashMenu{applying: true}
	defer func() { c.slash.applying = false }()
	c.input.snapshotUndo()
	if argument {
		typed := []rune(before[strings.Index(before, " ")+1:])
		head := string(runes[:cursor-len(typed)])
		c.input.text = head + item.value + after
		c.input.cursorPos = len([]rune(head)) + len([]rune(item.value))
		c.input.notifyChanged()
		c.markDirty()
		return
	}
	lead := slashLead(c.input.text)
	c.input.text = lead + "/" + item.name + " " + after
	c.input.cursorPos = len([]rune(lead)) + len([]rune(item.name)) + 2
	c.input.notifyChanged()
	if submit {
		c.input.enter(gotui.KeyEvent{Key: gotui.KeyEnter})
	}
	c.markDirty()
}

func (c *chatTUI) closeSlashMenu() {
	c.slash = slashMenu{}
	c.markDirty()
}

func (c *chatTUI) markDirty() {
	if c.app != nil {
		c.app.MarkDirty()
	}
}

// handleSlashKey is the editor's interceptKey hook.
func (c *chatTUI) handleSlashKey(key gotui.Key) bool {
	if !c.slash.active {
		return false
	}
	switch key {
	case gotui.KeyTab:
		c.applySlashSelection(false)
	case gotui.KeyEnter:
		c.applySlashSelection(true)
	case gotui.KeyEscape:
		c.closeSlashMenu()
	default:
		return false
	}
	return true
}

func (c *chatTUI) slashMenuKeys() gotui.KeyMap {
	if !c.slash.active {
		return nil
	}
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyUp, func(gotui.KeyEvent) { c.moveSlashSelection(-1) }),
		gotui.OnPreemptStop(gotui.KeyDown, func(gotui.KeyEvent) { c.moveSlashSelection(1) }),
		gotui.OnPreemptStop(gotui.KeyPageUp, func(gotui.KeyEvent) { c.pageSlashSelection(-1) }),
		gotui.OnPreemptStop(gotui.KeyPageDown, func(gotui.KeyEvent) { c.pageSlashSelection(1) }),
		gotui.OnPreemptStop(gotui.KeyTab, func(gotui.KeyEvent) { c.applySlashSelection(false) }),
		gotui.OnPreemptStop(gotui.KeyEnter, func(gotui.KeyEvent) { c.applySlashSelection(true) }),
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.closeSlashMenu() }),
	}
}

func (c *chatTUI) slashVisibleRange() (int, int) {
	n := len(c.slash.items)
	start := max(0, min(c.slash.selected-slashMaxVisible/2, n-slashMaxVisible))
	return start, min(start+slashMaxVisible, n)
}

func (c *chatTUI) slashMenuHeight() int {
	if !c.slash.active {
		return 0
	}
	start, end := c.slashVisibleRange()
	rows := end - start
	if start > 0 || end < len(c.slash.items) {
		rows++
	}
	return rows
}

// slashMenuRows ports SelectList.render/renderItem for the slash layout.
func (c *chatTUI) slashMenuRows(width int) [][]gotui.TextSpan {
	if !c.slash.active {
		return nil
	}
	widest := 0
	for _, it := range c.slash.items {
		widest = max(widest, gotui.StringWidth(it.name)+slashPrimaryGap)
	}
	primary := max(slashMinPrimaryCol, min(widest, slashMaxPrimaryCol))
	if c.slash.argument {
		primary = slashMaxPrimaryCol // SelectList's default layout: a fixed 32-cell column
	}
	accent, muted := piFg(piAccent), piFg(piMuted)
	start, end := c.slashVisibleRange()
	var rows [][]gotui.TextSpan
	for i := start; i < end; i++ {
		it := c.slash.items[i]
		selected := i == c.slash.selected
		prefix := "  "
		if selected {
			prefix = "→ "
		}
		desc := strings.Join(strings.Fields(it.description), " ")
		if desc != "" && width > 40 {
			col := max(1, min(primary, width-2-4))
			value := truncateCells(it.name, max(1, col-slashPrimaryGap))
			spacing := strings.Repeat(" ", max(1, col-gotui.StringWidth(value)))
			remaining := width - (2 + gotui.StringWidth(value) + len(spacing)) - 2
			if remaining > slashMinDescription {
				desc = truncateCells(desc, remaining)
				if selected {
					rows = append(rows, []gotui.TextSpan{{Text: prefix + value + spacing + desc, Style: accent}})
				} else {
					rows = append(rows, []gotui.TextSpan{{Text: prefix + value}, {Text: spacing + desc, Style: muted}})
				}
				continue
			}
		}
		value := truncateCells(it.name, max(1, width-2-2))
		if selected {
			rows = append(rows, []gotui.TextSpan{{Text: prefix + value, Style: accent}})
		} else {
			rows = append(rows, []gotui.TextSpan{{Text: prefix + value}})
		}
	}
	if start > 0 || end < len(c.slash.items) {
		rows = append(rows, []gotui.TextSpan{{Text: truncateCells(fmt.Sprintf("  (%d/%d)", c.slash.selected+1, len(c.slash.items)), max(0, width-2)), Style: muted}})
	}
	return rows
}

func (c *chatTUI) renderSlashMenu(width int) *gotui.Element {
	rows := c.slashMenuRows(width)
	block := gotui.New(gotui.WithDirection(gotui.Column), gotui.WithWidthPercent(100), gotui.WithHeight(len(rows)))
	for _, spans := range rows {
		block.AddChild(gotui.New(gotui.WithWidthPercent(100), gotui.WithHeight(1), gotui.WithWrap(false), gotui.WithRichText(spans...)))
	}
	return block
}
