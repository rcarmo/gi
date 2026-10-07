package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	gotui "github.com/grindlemire/go-tui"
)

// /tree is Pi's TreeSelectorComponent (components/tree-selector.js): the
// session's entries as a tree with branch connectors, the active path
// marked, filters, search, folding, labels and copy. Golden:
// scripts/golden-tree.mjs. gi's branches are forked sessions; see
// session_tree.go and docs/internal/tui-tree.md.

const (
	treeGutterWidth                = 2
	treeMinVisibleAnchorContent    = 4
	treeMaxVisibleAnchorContent    = 20
	treeMinAnchorContext           = 2
	treeMaxAnchorContext           = 12
	treeToolCommandPreview         = 50
	treeCustomToolArgumentsPreview = 40
)

var treeFilterModes = []string{"default", "no-tools", "user-only", "labeled-only", "all"}

type treeToolCall struct {
	name string
	args map[string]any
}

// treeEntry holds the Pi session entry fields the selector shows.
type treeEntry struct {
	id, parentID string
	typ          string // Pi's entry type
	role         string // message role
	content      any    // message/custom content: string or blocks
	calls        []struct {
		id, name string
		args     map[string]any
	}
	toolCallID, toolName        string
	stopReason, errorMessage    string
	command                     string // bashExecution
	tokensBefore                float64
	summary, modelID, thinking  string
	customType, name, targetID  string
	replacementNull, labelUnset bool
	labelText                   string
}

// treeEntryFromPi reads a Pi session entry (as JSON-decoded maps).
func treeEntryFromPi(e map[string]any) treeEntry {
	str := func(m map[string]any, k string) string { s, _ := m[k].(string); return s }
	t := treeEntry{id: str(e, "id"), parentID: str(e, "parentId"), typ: str(e, "type")}
	switch t.typ {
	case "message":
		msg, _ := e["message"].(map[string]any)
		t.role, t.content = str(msg, "role"), msg["content"]
		t.toolCallID, t.toolName = str(msg, "toolCallId"), str(msg, "toolName")
		t.stopReason, t.errorMessage, t.command = str(msg, "stopReason"), str(msg, "errorMessage"), str(msg, "command")
		if blocks, ok := msg["content"].([]any); ok {
			for _, b := range blocks {
				block, _ := b.(map[string]any)
				if str(block, "type") == "toolCall" {
					args, _ := block["arguments"].(map[string]any)
					t.calls = append(t.calls, struct {
						id, name string
						args     map[string]any
					}{str(block, "id"), str(block, "name"), args})
				}
			}
		}
	case "custom_message":
		t.customType, t.content = str(e, "customType"), e["content"]
	case "compaction":
		t.tokensBefore, _ = e["tokensBefore"].(float64)
		t.summary = str(e, "summary")
	case "branch_summary":
		t.summary = str(e, "summary")
	case "model_change":
		t.modelID = str(e, "modelId")
	case "thinking_level_change":
		t.thinking = str(e, "thinkingLevel")
	case "custom":
		t.customType = str(e, "customType")
	case "session_info":
		t.name = str(e, "name")
	case "context_edit":
		t.targetID, t.replacementNull = str(e, "targetId"), e["replacement"] == nil
	case "label":
		t.labelText, t.labelUnset = str(e, "label"), e["label"] == nil
	}
	return t
}

type treeNode struct {
	entry     treeEntry
	children  []*treeNode
	label     string
	labelTime string // RFC 3339, when labelled
}

type treeGutter struct {
	position int
	show     bool
}

type treeFlatNode struct {
	node               *treeNode
	indent             int
	showConnector      bool
	isLast             bool
	gutters            []treeGutter
	isVirtualRootChild bool
}

// treeList is Pi's TreeList.
type treeList struct {
	flat, filtered  []*treeFlatNode
	selected        int
	leafID          string
	maxVisible      int
	filterMode      string
	query           string
	toolCalls       map[string]treeToolCall
	multipleRoots   bool
	showLabelTimes  bool
	activePath      map[string]bool
	visibleParent   map[string]string // "" for roots
	visibleChildren map[string][]string
	lastSelectedID  string
	folded          map[string]bool
	byID            map[string]*treeFlatNode
	labelEditing    *treeLabelInput
	now             func() time.Time
}

type treeLabelInput struct {
	entryID, value string
}

func newTreeList(roots []*treeNode, leafID string, maxVisible int, initialSelected, filterMode string) *treeList {
	if filterMode == "" {
		filterMode = "default"
	}
	l := &treeList{leafID: leafID, maxVisible: maxVisible, filterMode: filterMode, toolCalls: map[string]treeToolCall{}, folded: map[string]bool{}, now: time.Now}
	l.multipleRoots = len(roots) > 1
	l.flat = l.flatten(roots)
	l.byID = map[string]*treeFlatNode{}
	for _, f := range l.flat {
		l.byID[f.node.entry.id] = f
	}
	l.buildActivePath()
	l.applyFilter()
	target := initialSelected
	if target == "" {
		target = leafID
	}
	l.selected = l.nearestVisible(target)
	if l.selected < len(l.filtered) {
		l.lastSelectedID = l.filtered[l.selected].node.entry.id
	}
	return l
}

func (l *treeList) nearestVisible(id string) int {
	if len(l.filtered) == 0 {
		return 0
	}
	index := map[string]int{}
	for i, f := range l.filtered {
		index[f.node.entry.id] = i
	}
	for id != "" {
		if i, ok := index[id]; ok {
			return i
		}
		f := l.byID[id]
		if f == nil {
			break
		}
		id = f.node.entry.parentID
	}
	return len(l.filtered) - 1
}

func (l *treeList) buildActivePath() {
	l.activePath = map[string]bool{}
	for id := l.leafID; id != ""; {
		l.activePath[id] = true
		f := l.byID[id]
		if f == nil {
			break
		}
		id = f.node.entry.parentID
	}
}

func (l *treeList) flatten(roots []*treeNode) []*treeFlatNode {
	containsActive := map[*treeNode]bool{}
	var mark func(*treeNode) bool
	mark = func(n *treeNode) bool {
		has := l.leafID != "" && n.entry.id == l.leafID
		for _, c := range n.children {
			if mark(c) {
				has = true
			}
		}
		containsActive[n] = has
		return has
	}
	for _, r := range roots {
		mark(r)
	}
	type frame struct {
		node                                                  *treeNode
		indent                                                int
		justBranched, showConnector, isLast, virtualRootChild bool
		gutters                                               []treeGutter
	}
	multiple := len(roots) > 1
	ordered := slices.Clone(roots)
	slices.SortStableFunc(ordered, func(a, b *treeNode) int { return boolInt(containsActive[b]) - boolInt(containsActive[a]) })
	var stack []frame
	for i := len(ordered) - 1; i >= 0; i-- {
		indent := 0
		if multiple {
			indent = 1
		}
		stack = append(stack, frame{ordered[i], indent, multiple, multiple, i == len(ordered)-1, multiple, nil})
	}
	var out []*treeFlatNode
	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		e := f.node.entry
		if e.typ == "message" && e.role == "assistant" {
			for _, c := range e.calls {
				l.toolCalls[c.id] = treeToolCall{c.name, c.args}
			}
		}
		out = append(out, &treeFlatNode{node: f.node, indent: f.indent, showConnector: f.showConnector, isLast: f.isLast, gutters: f.gutters, isVirtualRootChild: f.virtualRootChild})
		var children []*treeNode
		for _, c := range f.node.children {
			if containsActive[c] {
				children = append(children, c)
			}
		}
		for _, c := range f.node.children {
			if !containsActive[c] {
				children = append(children, c)
			}
		}
		childIndent, childGutters := l.childLayout(f.indent, f.justBranched, f.showConnector, f.isLast, f.virtualRootChild, f.gutters, len(children) > 1, multiple)
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, frame{children[i], childIndent, len(children) > 1, len(children) > 1, i == len(children)-1, false, childGutters})
		}
	}
	return out
}

// childLayout is flattenTree's child indent and gutters.
func (l *treeList) childLayout(indent int, justBranched, showConnector, isLast, virtualRootChild bool, gutters []treeGutter, multipleChildren, multipleRoots bool) (int, []treeGutter) {
	childIndent := indent
	if multipleChildren || justBranched && indent > 0 {
		childIndent = indent + 1
	}
	if showConnector && !virtualRootChild {
		display := indent
		if multipleRoots {
			display = max(0, indent-1)
		}
		gutters = append(slices.Clone(gutters), treeGutter{position: max(0, display-1), show: !isLast})
	}
	return childIndent, gutters
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (l *treeList) applyFilter() {
	if len(l.filtered) > 0 && l.selected < len(l.filtered) {
		l.lastSelectedID = l.filtered[l.selected].node.entry.id
	}
	tokens := strings.Fields(strings.ToLower(l.query))
	l.filtered = nil
	for _, f := range l.flat {
		e := f.node.entry
		if e.typ == "usage" {
			continue
		}
		if e.typ == "message" && e.role == "assistant" && e.id != l.leafID {
			errorOrAborted := e.stopReason != "" && e.stopReason != "stop" && e.stopReason != "toolUse"
			if !hasTreeText(e.content) && !errorOrAborted {
				continue
			}
		}
		settings := e.typ == "label" || e.typ == "context_edit" || e.typ == "custom" || e.typ == "model_change" || e.typ == "thinking_level_change" || e.typ == "session_info"
		pass := !settings
		switch l.filterMode {
		case "user-only":
			pass = e.typ == "message" && e.role == "user"
		case "no-tools":
			pass = !settings && !(e.typ == "message" && e.role == "toolResult")
		case "labeled-only":
			pass = f.node.label != ""
		case "all":
			pass = true
		}
		if !pass {
			continue
		}
		if len(tokens) > 0 {
			text := strings.ToLower(l.searchText(f.node))
			if slices.ContainsFunc(tokens, func(t string) bool { return !strings.Contains(text, t) }) {
				continue
			}
		}
		l.filtered = append(l.filtered, f)
	}
	if len(l.folded) > 0 {
		skip := map[string]bool{}
		for _, f := range l.flat {
			if p := f.node.entry.parentID; p != "" && (l.folded[p] || skip[p]) {
				skip[f.node.entry.id] = true
			}
		}
		l.filtered = slices.DeleteFunc(l.filtered, func(f *treeFlatNode) bool { return skip[f.node.entry.id] })
	}
	l.recalculate()
	if l.lastSelectedID != "" {
		l.selected = l.nearestVisible(l.lastSelectedID)
	} else if l.selected >= len(l.filtered) {
		l.selected = max(0, len(l.filtered)-1)
	}
	if len(l.filtered) > 0 {
		l.lastSelectedID = l.filtered[l.selected].node.entry.id
	}
}

// recalculate is Pi's recalculateVisualStructure over the visible entries.
func (l *treeList) recalculate() {
	if len(l.filtered) == 0 {
		return
	}
	visible := map[string]bool{}
	for _, f := range l.filtered {
		visible[f.node.entry.id] = true
	}
	ancestor := func(id string) string {
		for cur := l.byID[id].node.entry.parentID; cur != ""; {
			if visible[cur] {
				return cur
			}
			f := l.byID[cur]
			if f == nil {
				return ""
			}
			cur = f.node.entry.parentID
		}
		return ""
	}
	l.visibleParent, l.visibleChildren = map[string]string{}, map[string][]string{"": nil}
	for _, f := range l.filtered {
		id := f.node.entry.id
		a := ancestor(id)
		l.visibleParent[id] = a
		l.visibleChildren[a] = append(l.visibleChildren[a], id)
	}
	roots := l.visibleChildren[""]
	l.multipleRoots = len(roots) > 1
	type frame struct {
		id                                                    string
		indent                                                int
		justBranched, showConnector, isLast, virtualRootChild bool
		gutters                                               []treeGutter
	}
	var stack []frame
	for i := len(roots) - 1; i >= 0; i-- {
		indent := 0
		if l.multipleRoots {
			indent = 1
		}
		stack = append(stack, frame{roots[i], indent, l.multipleRoots, l.multipleRoots, i == len(roots)-1, l.multipleRoots, nil})
	}
	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := l.byID[f.id]
		node.indent, node.showConnector, node.isLast, node.gutters, node.isVirtualRootChild = f.indent, f.showConnector, f.isLast, f.gutters, f.virtualRootChild
		children := l.visibleChildren[f.id]
		childIndent, childGutters := l.childLayout(f.indent, f.justBranched, f.showConnector, f.isLast, f.virtualRootChild, f.gutters, len(children) > 1, l.multipleRoots)
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, frame{children[i], childIndent, len(children) > 1, len(children) > 1, i == len(children)-1, false, childGutters})
		}
	}
}

func treeContentText(content any) string {
	switch c := content.(type) {
	case string:
		return c
	case []any:
		var b strings.Builder
		for _, block := range c {
			if m, ok := block.(map[string]any); ok && m["type"] == "text" {
				s, _ := m["text"].(string)
				b.WriteString(s)
			}
		}
		return b.String()
	}
	return ""
}

func hasTreeText(content any) bool {
	switch c := content.(type) {
	case string:
		return strings.TrimSpace(c) != ""
	case []any:
		for _, block := range c {
			if m, ok := block.(map[string]any); ok && m["type"] == "text" {
				if s, _ := m["text"].(string); strings.TrimSpace(s) != "" {
					return true
				}
			}
		}
	}
	return false
}

// treeExtract is Pi's extractContent: the text, cut to 200 UTF-16 units.
func treeExtract(content any) string {
	return jsSlice(treeContentText(content), 200)
}

// jsSlice cuts s to n UTF-16 code units, as JavaScript's slice does.
func jsSlice(s string, n int) string {
	units := 0
	for i, r := range s {
		w := 1
		if r > 0xFFFF {
			w = 2
		}
		if units+w > n {
			return s[:i]
		}
		units += w
	}
	return s
}

func (l *treeList) searchText(n *treeNode) string {
	e := n.entry
	var parts []string
	if n.label != "" {
		parts = append(parts, n.label)
	}
	switch e.typ {
	case "message":
		parts = append(parts, e.role)
		if e.content != nil {
			if text := treeExtract(e.content); text != "" || true {
				parts = append(parts, text)
			}
		}
		if e.role == "bashExecution" && e.command != "" {
			parts = append(parts, e.command)
		}
	case "custom_message":
		parts = append(parts, e.customType, treeExtract(e.content))
	case "compaction":
		parts = append(parts, "compaction")
	case "branch_summary":
		parts = append(parts, "branch summary", e.summary)
	case "session_info":
		parts = append(parts, "title")
		if e.name != "" {
			parts = append(parts, e.name)
		}
	case "model_change":
		parts = append(parts, "model", e.modelID)
	case "thinking_level_change":
		parts = append(parts, "thinking", e.thinking)
	case "custom":
		parts = append(parts, "custom", e.customType)
	case "context_edit":
		kind := "replace"
		if e.replacementNull {
			kind = "omit"
		}
		parts = append(parts, "context edit", kind, e.targetID)
	case "label":
		parts = append(parts, "label", e.labelText)
	}
	return strings.Join(parts, " ")
}

func (l *treeList) statusLabels() string {
	s := ""
	switch l.filterMode {
	case "no-tools":
		s += " [no-tools]"
	case "user-only":
		s += " [user]"
	case "labeled-only":
		s += " [labeled]"
	case "all":
		s += " [all]"
	}
	if l.showLabelTimes {
		s += " [+label time]"
	}
	return s
}

func (l *treeList) foldable(id string) bool {
	if len(l.visibleChildren[id]) == 0 {
		return false
	}
	parent, ok := l.visibleParent[id]
	if !ok || parent == "" {
		return true
	}
	return len(l.visibleChildren[parent]) > 1
}

type treeRow struct {
	gutter, body []gotui.TextSpan
	anchor       int
	selected     bool
}

func (l *treeList) rows(width int) spanRows {
	muted := piFg(piMuted)
	if len(l.filtered) == 0 {
		return spanRows{
			piTruncate([]gotui.TextSpan{{Text: "  No entries found", Style: muted}}, width),
			piTruncate([]gotui.TextSpan{{Text: "  (0/0)" + l.statusLabels(), Style: muted}}, width),
		}
	}
	start := max(0, min(l.selected-l.maxVisible/2, len(l.filtered)-l.maxVisible))
	end := min(start+l.maxVisible, len(l.filtered))
	var rendered []treeRow
	for i := start; i < end; i++ {
		f := l.filtered[i]
		e := f.node.entry
		selected := i == l.selected
		cursor := gotui.TextSpan{Text: "  "}
		if selected {
			cursor = gotui.TextSpan{Text: "› ", Style: piFg(piAccent)}
		}
		display := f.indent
		if l.multipleRoots {
			display = max(0, f.indent-1)
		}
		connector := f.showConnector && !f.isVirtualRootChild
		connectorPos := -1
		if connector {
			connectorPos = display - 1
		}
		folded := l.folded[e.id]
		var prefix strings.Builder
		for c := 0; c < display*3; c++ {
			level, pos := c/3, c%3
			gutter := slices.IndexFunc(f.gutters, func(g treeGutter) bool { return g.position == level })
			switch {
			case gutter >= 0:
				if pos == 0 && f.gutters[gutter].show {
					prefix.WriteString("│")
				} else {
					prefix.WriteString(" ")
				}
			case connector && level == connectorPos:
				switch pos {
				case 0:
					if f.isLast {
						prefix.WriteString("└")
					} else {
						prefix.WriteString("├")
					}
				case 1:
					switch {
					case folded:
						prefix.WriteString("⊞")
					case l.foldable(e.id):
						prefix.WriteString("⊟")
					default:
						prefix.WriteString("─")
					}
				default:
					prefix.WriteString(" ")
				}
			default:
				prefix.WriteString(" ")
			}
		}
		body := []gotui.TextSpan{{Text: prefix.String(), Style: piFg(piDim)}}
		if folded && !connector {
			body = append(body, gotui.TextSpan{Text: "⊞ ", Style: piFg(piAccent)})
		}
		if l.activePath[e.id] {
			body = append(body, gotui.TextSpan{Text: "• ", Style: piFg(piAccent)})
		}
		anchor := spansWidth(body)
		if f.node.label != "" {
			body = append(body, gotui.TextSpan{Text: "[" + f.node.label + "] ", Style: piFg(piWarning)})
			if l.showLabelTimes && f.node.labelTime != "" {
				body = append(body, gotui.TextSpan{Text: l.labelTimestamp(f.node.labelTime) + " ", Style: muted})
			}
		}
		body = append(body, l.displayText(f.node, selected)...)
		if selected {
			cursor.Style = cursor.Style.Background(piSelectedBg)
			for j := range body {
				body[j].Style = body[j].Style.Background(piSelectedBg)
			}
		}
		rendered = append(rendered, treeRow{gutter: []gotui.TextSpan{cursor}, body: body, anchor: anchor, selected: selected})
	}
	rows := treeHorizontalViewport(rendered, width)
	return append(rows, piTruncate([]gotui.TextSpan{{Text: fmt.Sprintf("  (%d/%d)%s", l.selected+1, len(l.filtered), l.statusLabels()), Style: muted}}, width))
}

// treeHorizontalViewport is Pi's renderHorizontalViewport: bodies shift
// left only when the selected row's text would start too far right.
func treeHorizontalViewport(rows []treeRow, width int) spanRows {
	viewport := max(0, width-treeGutterWidth)
	widest := 0
	for _, r := range rows {
		widest = max(widest, spansWidth(r.body))
	}
	maxScroll := max(0, widest-viewport)
	scroll := 0
	for _, r := range rows {
		if r.selected && maxScroll > 0 {
			minContent := min(treeMaxVisibleAnchorContent, max(treeMinVisibleAnchorContent, viewport/3))
			if r.anchor > viewport-minContent {
				context := min(treeMaxAnchorContext, max(treeMinAnchorContext, viewport/4))
				scroll = min(maxScroll, r.anchor-context)
			}
		}
	}
	out := make(spanRows, len(rows))
	for i, r := range rows {
		body := r.body
		if scroll > 0 {
			body = sliceSpans(body, scroll, viewport)
		}
		out[i] = clipSpans(append(slices.Clone(r.gutter), body...), width)
	}
	return out
}

// sliceSpans is pi-tui's sliceByColumn: the cells from start, at most width.
func sliceSpans(spans []gotui.TextSpan, start, width int) []gotui.TextSpan {
	var out []gotui.TextSpan
	col := 0
	for _, s := range spans {
		var b strings.Builder
		for _, r := range s.Text {
			w := gotui.StringWidth(string(r))
			if col >= start && col+w <= start+width {
				b.WriteRune(r)
			}
			col += w
		}
		if b.Len() > 0 {
			out = append(out, gotui.TextSpan{Text: b.String(), Style: s.Style})
		}
	}
	return out
}

func (l *treeList) labelTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ""
	}
	now := l.now()
	t = t.In(now.Location())
	clock := t.Format("15:04")
	switch {
	case t.Year() == now.Year() && t.YearDay() == now.YearDay():
		return clock
	case t.Year() == now.Year():
		return fmt.Sprintf("%d/%d %s", int(t.Month()), t.Day(), clock)
	}
	return fmt.Sprintf("%s/%d/%d %s", t.Format("06"), int(t.Month()), t.Day(), clock)
}

func treeNormalize(s string) string {
	return strings.TrimSpace(strings.NewReplacer("\n", " ", "\t", " ").Replace(s))
}

func (l *treeList) displayText(n *treeNode, selected bool) []gotui.TextSpan {
	e := n.entry
	span := func(text string, c gotui.Color) gotui.TextSpan { return gotui.TextSpan{Text: text, Style: piFg(c)} }
	var out []gotui.TextSpan
	switch e.typ {
	case "message":
		switch e.role {
		case "user":
			out = []gotui.TextSpan{span("user: ", piAccent), {Text: treeNormalize(treeExtract(e.content))}}
		case "assistant":
			out = []gotui.TextSpan{span("assistant: ", piSuccess)}
			switch text := treeNormalize(treeExtract(e.content)); {
			case text != "":
				out = append(out, gotui.TextSpan{Text: text})
			case e.stopReason == "aborted":
				out = append(out, span("(aborted)", piMuted))
			case e.errorMessage != "":
				out = append(out, span(jsSlice(treeNormalize(e.errorMessage), 80), piError))
			default:
				out = append(out, span("(no content)", piMuted))
			}
		case "toolResult":
			if call, ok := l.toolCalls[e.toolCallID]; ok && e.toolCallID != "" {
				out = []gotui.TextSpan{span(treeFormatToolCall(call.name, call.args), piMuted)}
			} else {
				name := e.toolName
				if name == "" {
					name = "tool"
				}
				out = []gotui.TextSpan{span("["+name+"]", piMuted)}
			}
		case "bashExecution":
			out = []gotui.TextSpan{span("[bash]: "+treeNormalize(e.command), piDim)}
		default:
			out = []gotui.TextSpan{span("["+e.role+"]", piDim)}
		}
	case "custom_message":
		out = []gotui.TextSpan{span("["+e.customType+"]: ", piAccent), {Text: treeNormalize(treeContentText(e.content))}}
	case "compaction":
		out = []gotui.TextSpan{span(fmt.Sprintf("[compaction: %dk tokens]", int(e.tokensBefore/1000+0.5)), piAccent)}
	case "branch_summary":
		out = []gotui.TextSpan{span("[branch summary]: ", piWarning), {Text: treeNormalize(e.summary)}}
	case "model_change":
		out = []gotui.TextSpan{span("[model: "+e.modelID+"]", piDim)}
	case "thinking_level_change":
		out = []gotui.TextSpan{span("[thinking: "+e.thinking+"]", piDim)}
	case "custom":
		out = []gotui.TextSpan{span("[custom: "+e.customType+"]", piDim)}
	case "context_edit":
		kind := "replace"
		if e.replacementNull {
			kind = "omit"
		}
		out = []gotui.TextSpan{span(fmt.Sprintf("[context %s: %s]", kind, e.targetID), piDim)}
	case "label":
		text := e.labelText
		if e.labelUnset {
			text = "(cleared)"
		}
		out = []gotui.TextSpan{span("[label: "+text+"]", piDim)}
	case "session_info":
		if e.name != "" {
			out = []gotui.TextSpan{span("[title: "+e.name+"]", piDim)}
		} else {
			out = []gotui.TextSpan{span("[title: ", piDim), {Text: "empty", Style: piFg(piDim).Italic()}, span("]", piDim)}
		}
	}
	if selected {
		for i := range out {
			out[i].Style = out[i].Style.Bold()
		}
	}
	return out
}

// treeFormatToolCall is Pi's formatToolCall.
func treeFormatToolCall(name string, args map[string]any) string {
	home, _ := os.UserHomeDir()
	if h := os.Getenv("HOME"); h != "" {
		home = h
	}
	short := func(p string) string {
		if home != "" && strings.HasPrefix(p, home) {
			return "~" + p[len(home):]
		}
		return p
	}
	arg := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := args[k]; ok && v != nil && v != "" {
				return fmt.Sprint(v)
			}
		}
		return ""
	}
	orDot := func(s string) string {
		if s == "" {
			return "."
		}
		return s
	}
	switch name {
	case "read":
		display := short(arg("path", "file_path"))
		offset, hasOffset := args["offset"].(float64)
		limit, hasLimit := args["limit"].(float64)
		if hasOffset || hasLimit {
			if !hasOffset {
				offset = 1
			}
			display += fmt.Sprintf(":%d", int(offset))
			if hasLimit {
				display += fmt.Sprintf("-%d", int(offset+limit-1))
			}
		}
		return "[read: " + display + "]"
	case "write", "edit":
		return "[" + name + ": " + short(arg("path", "file_path")) + "]"
	case "bash":
		raw := arg("command")
		cmd := jsSlice(treeNormalize(raw), treeToolCommandPreview)
		if len([]rune(raw)) > treeToolCommandPreview {
			cmd += "..."
		}
		return "[bash: " + cmd + "]"
	case "grep":
		return "[grep: /" + arg("pattern") + "/ in " + short(orDot(arg("path"))) + "]"
	case "find":
		return "[find: " + arg("pattern") + " in " + short(orDot(arg("path"))) + "]"
	case "ls":
		return "[ls: " + short(orDot(arg("path"))) + "]"
	}
	raw, _ := json.Marshal(args)
	if args == nil {
		raw = []byte("{}")
	}
	text := jsSlice(string(raw), treeCustomToolArgumentsPreview)
	if len(raw) > treeCustomToolArgumentsPreview {
		text += "..."
	}
	return "[" + name + ": " + text + "]"
}

// copyText is Pi's getEntryCopyText.
func (l *treeList) copyText(n *treeNode) string {
	e := n.entry
	text := ""
	switch e.typ {
	case "message":
		if e.role == "bashExecution" {
			text = e.command
		} else {
			text = treeContentText(e.content)
			if text == "" && e.role == "assistant" {
				text = e.errorMessage
			}
		}
	case "custom_message":
		text = treeContentText(e.content)
	case "compaction", "branch_summary":
		text = e.summary
	}
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return text
}

// segmentStart is Pi's findBranchSegmentStart.
func (l *treeList) segmentStart(down bool) int {
	if l.selected >= len(l.filtered) {
		return l.selected
	}
	index := map[string]int{}
	for i, f := range l.filtered {
		index[f.node.entry.id] = i
	}
	id := l.filtered[l.selected].node.entry.id
	if down {
		for {
			children := l.visibleChildren[id]
			switch {
			case len(children) == 0:
				return index[id]
			case len(children) > 1:
				return index[children[0]]
			}
			id = children[0]
		}
	}
	for {
		parent := l.visibleParent[id]
		if parent == "" {
			return index[id]
		}
		if len(l.visibleChildren[parent]) > 1 {
			if start := index[id]; start < l.selected {
				return start
			}
		}
		id = parent
	}
}

func (l *treeList) setFilter(mode string) {
	l.filterMode = mode
	l.folded = map[string]bool{}
	l.applyFilter()
}

func (l *treeList) toggleFilter(mode string) {
	if l.filterMode == mode {
		mode = "default"
	}
	l.setFilter(mode)
}

func (l *treeList) cycleFilter(delta int) {
	i := slices.Index(treeFilterModes, l.filterMode)
	l.setFilter(treeFilterModes[(i+delta+len(treeFilterModes))%len(treeFilterModes)])
}

func (l *treeList) setQuery(q string) {
	l.query = q
	l.folded = map[string]bool{}
	l.applyFilter()
}

func (l *treeList) move(delta int) {
	n := len(l.filtered)
	if n == 0 {
		return
	}
	l.selected = (l.selected + delta + n) % n
}

func (l *treeList) page(delta int) {
	l.selected = max(0, min(len(l.filtered)-1, l.selected+delta*l.maxVisible))
}

func (l *treeList) foldOrUp() {
	if l.selected >= len(l.filtered) {
		return
	}
	id := l.filtered[l.selected].node.entry.id
	if l.foldable(id) && !l.folded[id] {
		l.folded[id] = true
		l.applyFilter()
		return
	}
	l.selected = l.segmentStart(false)
}

func (l *treeList) unfoldOrDown() {
	if l.selected >= len(l.filtered) {
		return
	}
	id := l.filtered[l.selected].node.entry.id
	if l.folded[id] {
		delete(l.folded, id)
		l.applyFilter()
		return
	}
	l.selected = l.segmentStart(true)
}

func (l *treeList) selectedNode() *treeNode {
	if l.selected < len(l.filtered) {
		return l.filtered[l.selected].node
	}
	return nil
}

// treeHelpRows is Pi's TreeHelp.
func treeHelpRows(width int) spanRows {
	items := []string{"↑/↓ move", "←/→ page", "ctrl+←/→ branch", "ctrl+x copy", "shift+l label", "shift+t label time", "filters ctrl+d/t/u/l/a", "cycle ctrl+o/shift+ctrl+o"}
	if defaultPiKeys.darwin {
		for i := range items {
			items[i] = strings.ReplaceAll(items[i], "alt+", "option+")
		}
	}
	avail := max(1, width)
	var lines []string
	current := ""
	indented := func(item string) string {
		if gotui.StringWidth("  "+item) <= avail {
			return "  " + item
		}
		return item
	}
	flush := func() {
		for _, line := range piWrapLine([]gotui.TextSpan{{Text: strings.TrimRight(current, " ")}}, avail) {
			lines = append(lines, spanText(line))
		}
	}
	for _, item := range items {
		candidate := indented(item)
		if current != "" {
			candidate = current + " · " + item
		}
		if current == "" || gotui.StringWidth(candidate) <= avail {
			current = candidate
			continue
		}
		flush()
		current = indented(item)
	}
	if current != "" {
		flush()
	}
	rows := make(spanRows, len(lines))
	for i, line := range lines {
		rows[i] = []gotui.TextSpan{{Text: line, Style: piFg(piMuted)}}
	}
	return rows
}

func spanText(spans []gotui.TextSpan) string {
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.Text)
	}
	return b.String()
}

// treeSelectorRows is Pi's TreeSelectorComponent layout.
func (l *treeList) selectorRows(width int) spanRows {
	rows := spanRows{nil, piRule(width, piBorder)}
	for _, line := range piWrapLine([]gotui.TextSpan{{Text: "  Session Tree", Style: gotui.NewStyle().Bold()}}, max(1, width-2)) {
		rows = append(rows, append([]gotui.TextSpan{{Text: " "}}, line...))
	}
	rows = append(rows, treeHelpRows(width)...)
	search := []gotui.TextSpan{{Text: "  "}, {Text: "Type to search:", Style: piFg(piMuted)}}
	if l.query != "" {
		search = append(search, gotui.TextSpan{Text: " "}, gotui.TextSpan{Text: l.query, Style: piFg(piAccent)})
	}
	rows = append(rows, piTruncate(search, width), piRule(width, piBorder), nil)
	if in := l.labelEditing; in != nil {
		rows = append(rows,
			piTruncate([]gotui.TextSpan{{Text: "  "}, {Text: "Label (empty to remove):", Style: piFg(piMuted)}}, width),
			piTruncate(append([]gotui.TextSpan{{Text: "  "}}, piSearchRow(in.value, width-2)...), width),
			piTruncate(append(append(append([]gotui.TextSpan{{Text: "  "}}, keyHintSpans("enter", "save")...), gotui.TextSpan{Text: "  "}), keyHintSpans("escape/ctrl+c", "cancel")...), width))
	} else {
		rows = append(rows, l.rows(width)...)
	}
	return append(rows, nil, piRule(width, piBorder))
}

// treeSelector is Pi's TreeSelectorComponent: the tree list, or the label
// input in its place while a label is edited.
type treeSelector struct {
	list     *treeList
	onSelect func(id string)
	onCancel func()
	onLabel  func(id, label string)
	onCopy   func(text string)
}

// newTreeSelector sizes the list as Pi does: half the terminal, at least 5.
func newTreeSelector(roots []*treeNode, leafID string, terminalHeight int, initialSelected, filterMode string) *treeSelector {
	return &treeSelector{list: newTreeList(roots, leafID, max(5, terminalHeight/2), initialSelected, filterMode)}
}

// updateLabel is Pi's updateNodeLabel.
func (l *treeList) updateLabel(id, label string) {
	if f := l.byID[id]; f != nil {
		f.node.label, f.node.labelTime = label, ""
		if label != "" {
			f.node.labelTime = l.now().UTC().Format("2006-01-02T15:04:05.000Z")
		}
	}
}

// treeSelectorKeys is Pi's TreeList, LabelInput and TreeSelectorComponent
// handleInput with Pi's default keys.
func treeSelectorKeys(s *treeSelector, dirty func()) gotui.KeyMap {
	l := s.list
	do := func(f func()) func(gotui.KeyEvent) {
		return func(gotui.KeyEvent) { f(); dirty() }
	}
	// Each binding acts on the label input when it is open, else on the list.
	bind := func(m gotui.KeyMatcher, list func(), label func(in *treeLabelInput)) gotui.KeyBinding {
		return gotui.OnPreemptStop(m, do(func() {
			if in := l.labelEditing; in != nil {
				if label != nil {
					label(in)
				}
				return
			}
			list()
		}))
	}
	closeLabel := func(*treeLabelInput) { l.labelEditing = nil }
	cancel := func() {
		if l.query != "" {
			l.setQuery("")
			return
		}
		if s.onCancel != nil {
			s.onCancel()
		}
	}
	confirm := func() {
		if n := l.selectedNode(); n != nil && s.onSelect != nil {
			s.onSelect(n.entry.id)
		}
	}
	saveLabel := func(in *treeLabelInput) {
		label := strings.TrimSpace(in.value)
		l.updateLabel(in.entryID, label)
		if s.onLabel != nil {
			s.onLabel(in.entryID, label)
		}
		l.labelEditing = nil
	}
	editLabel := func() {
		if n := l.selectedNode(); n != nil {
			l.labelEditing = &treeLabelInput{entryID: n.entry.id, value: n.label}
		}
	}
	copySelected := func() {
		if s.onCopy == nil {
			return
		}
		text := ""
		if n := l.selectedNode(); n != nil {
			text = l.copyText(n)
		}
		s.onCopy(text)
	}
	backspace := func() {
		if r := []rune(l.query); len(r) > 0 {
			l.setQuery(string(r[:len(r)-1]))
		}
	}
	filter := func(mode string) func() { return func() { l.toggleFilter(mode) } }
	none := func(*treeLabelInput) {}
	return gotui.KeyMap{
		bind(gotui.KeyUp, func() { l.move(-1) }, none),
		bind(gotui.KeyDown, func() { l.move(1) }, none),
		bind(gotui.KeyLeft.Ctrl(), l.foldOrUp, none),
		bind(gotui.KeyLeft.Alt(), l.foldOrUp, none),
		bind(gotui.KeyRight.Ctrl(), l.unfoldOrDown, none),
		bind(gotui.KeyRight.Alt(), l.unfoldOrDown, none),
		bind(gotui.KeyLeft, func() { l.page(-1) }, none),
		bind(gotui.KeyPageUp, func() { l.page(-1) }, none),
		bind(gotui.KeyRight, func() { l.page(1) }, none),
		bind(gotui.KeyPageDown, func() { l.page(1) }, none),
		bind(gotui.KeyEnter, confirm, saveLabel),
		bind(gotui.KeyCtrlX, copySelected, none),
		bind(gotui.KeyEscape, cancel, closeLabel),
		bind(gotui.KeyCtrlC, cancel, closeLabel),
		bind(gotui.KeyCtrlD, func() { l.setFilter("default") }, none),
		bind(gotui.KeyCtrlT, filter("no-tools"), none),
		bind(gotui.KeyCtrlU, filter("user-only"), none),
		bind(gotui.KeyCtrlL, filter("labeled-only"), none),
		bind(gotui.KeyCtrlA, filter("all"), none),
		bind(gotui.Rune('o').Ctrl().Shift(), func() { l.cycleFilter(-1) }, none),
		bind(gotui.KeyCtrlO, func() { l.cycleFilter(1) }, none),
		bind(gotui.KeyBackspace, backspace, func(in *treeLabelInput) {
			if r := []rune(in.value); len(r) > 0 {
				in.value = string(r[:len(r)-1])
			}
		}),
		gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) {
			if in := l.labelEditing; in != nil {
				in.value += string(ke.Rune)
			} else {
				switch ke.Rune {
				case 'L': // shift+l
					editLabel()
				case 'T': // shift+t
					l.showLabelTimes = !l.showLabelTimes
				default:
					l.setQuery(l.query + string(ke.Rune))
				}
			}
			dirty()
		}),
	}
}
