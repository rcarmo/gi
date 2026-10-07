package tui

import (
	"context"
	"strings"
	"testing"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// treeTestMessage adds a message to a session.
func treeTestMessage(t *testing.T, c *chatTUI, sessionID, role, content string) {
	t.Helper()
	if err := c.store.AddMessage(context.Background(), store.NowID("msg"), sessionID, role, content, map[string]any{"kind": "chat"}); err != nil {
		t.Fatal(err)
	}
}

// treeRows renders the open /tree selector.
func treeRows(t *testing.T, c *chatTUI) string {
	t.Helper()
	if c.treeSelector == nil {
		t.Fatal("tree not open")
	}
	return strings.Join(spanRowsText(c.treeSelector.list.selectorRows(80)), "\n")
}

// treeSelect moves the selection to the entry whose row contains text and
// presses Enter.
func treeSelect(t *testing.T, c *chatTUI, text string) {
	t.Helper()
	l := c.treeSelector.list
	for i, f := range l.filtered {
		if strings.Contains(treeContentText(f.node.entry.content), text) {
			l.selected = i
			pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter})
			return
		}
	}
	t.Fatalf("no entry %q in\n%s", text, treeRows(t, c))
}

// /tree navigates the session's branch family: going back copies the
// session up to the chosen point (a user message's text returns to the
// editor), the copies show as branches, choosing the end of an existing
// branch switches to its session, labels are kept for the family, and a
// branch summary is added to a new copy.
func TestTreeNavigatesBranchFamily(t *testing.T) {
	c := sessionTestChat(t)
	c.cfg.BranchSummarySkipPrompt = true
	for _, m := range [][2]string{{"user", "Plan the parser"}, {"assistant", "Plan ready"}, {"user", "Start with the lexer"}, {"assistant", "Lexer done"}} {
		treeTestMessage(t, c, "A", m[0], m[1])
	}
	c.openTreeSelector("")
	if rows := treeRows(t, c); !strings.Contains(rows, "› ") || !strings.Contains(rows, "assistant: Lexer done") {
		t.Fatalf("rows:\n%s", rows)
	}
	treeSelect(t, c, "Start with the lexer")
	branch := c.sessionID
	if branch == "A" || c.input.Text() != "Start with the lexer" || c.modelMenuOpen {
		t.Fatalf("session %s editor %q menu %v", branch, c.input.Text(), c.modelMenuOpen)
	}
	sess, err := c.store.GetSession(context.Background(), branch)
	if err != nil || store.TreeParent(*sess) != "A" {
		t.Fatalf("branch %+v %v", sess, err)
	}
	original, err := c.store.GetSession(context.Background(), "A")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Scope.AgentID != original.Scope.AgentID || sess.Scope.Channel != original.Scope.Channel || sess.Scope.Account != original.Scope.Account || sess.Scope.Values["chat"] == original.Scope.Values["chat"] {
		t.Fatalf("tree copy changed agent/account or reused chat: %+v", sess.Scope)
	}
	c.input.SetText("")
	treeTestMessage(t, c, branch, "user", "Start with the parser")
	treeTestMessage(t, c, branch, "assistant", "Parser done")

	// Both branches hang off the shared "Plan ready", the active one first.
	c.openTreeSelector("")
	rows := treeRows(t, c)
	for _, want := range []string{"├⊟ • user: Start with the parser", "› │     • assistant: Parser done", "└⊟ user: Start with the lexer"} {
		if !strings.Contains(rows, want) {
			t.Fatalf("rows lack %q:\n%s", want, rows)
		}
	}
	if strings.Count(rows, "Plan ready") != 1 {
		t.Fatalf("shared history repeated:\n%s", rows)
	}
	// Label the shared entry from the branch.
	treeSelectOnly(t, c, "Plan ready")
	for _, ev := range append([]gotui.KeyEvent{{Key: gotui.KeyRune, Rune: 'L'}}, runeEvents("plan")...) {
		pressMenuKey(t, c.KeyMap(), ev)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter})
	// Back to the end of the first branch: its session, no copy.
	treeSelect(t, c, "Lexer done")
	if c.sessionID != "A" {
		t.Fatalf("session %s, want A", c.sessionID)
	}
	c.openTreeSelector("")
	if rows := treeRows(t, c); !strings.Contains(rows, "[plan] assistant: Plan ready") {
		t.Fatalf("label lost:\n%s", rows)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEscape})
	if c.modelMenuOpen || c.treeSelector != nil {
		t.Fatal("escape did not close the tree")
	}

	// A summary of the branch being left goes into a new copy of the target.
	tree, err := loadSessionTree(context.Background(), c.store, c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	var target string
	for id, n := range tree.nodes {
		if treeContentText(n.entry.content) == "Parser done" {
			target = id
		}
	}
	abandoned := tree.abandoned(target)
	if len(abandoned) != 2 || abandoned[0].Content != "Start with the lexer" || abandoned[1].Content != "Lexer done" {
		t.Fatalf("abandoned %+v", abandoned)
	}
	c.finishTreeNavigation(tree, target, "", "SUMMARY")
	if c.sessionID == "A" || c.sessionID == branch {
		t.Fatalf("summary did not get a new copy: %s", c.sessionID)
	}
	summarised, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if summarised.Scope.AgentID != original.Scope.AgentID || summarised.Scope.Values["chat"] == sess.Scope.Values["chat"] {
		t.Fatalf("summary copy changed agent or reused chat: %+v", summarised.Scope)
	}
	messages, err := c.store.ListMessages(context.Background(), c.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	last := messages[len(messages)-1]
	if len(messages) != 5 || last.Role != "user" || last.Content != "SUMMARY" || last.Payload["kind"] != turn.BranchSummaryKind {
		t.Fatalf("messages %+v", messages)
	}
	block := false
	for i, line := range c.transcript {
		if meta, ok := decodeTranscriptBlockMarker(line); ok && meta.Title == "Branch summary" && i+1 < len(c.transcript) && c.transcript[i+1] == "│ SUMMARY" {
			block = true
		}
	}
	if !block {
		t.Fatalf("no branch summary block:\n%s", strings.Join(c.transcript, "\n"))
	}
	c.openTreeSelector("")
	if rows := treeRows(t, c); !strings.Contains(rows, "[branch summary]: SUMMARY") {
		t.Fatalf("rows:\n%s", rows)
	}
}

// Without skipPrompt, choosing an entry asks Pi's "Summarize branch?";
// Escape returns to the tree with the entry selected, "No summary"
// navigates.
func TestTreeAsksAboutBranchSummary(t *testing.T) {
	c := sessionTestChat(t)
	for _, m := range [][2]string{{"user", "one"}, {"assistant", "first"}, {"user", "two"}, {"assistant", "second"}} {
		treeTestMessage(t, c, "A", m[0], m[1])
	}
	c.openTreeSelector("")
	treeSelect(t, c, "first")
	if c.modelMenuKind != "select" || c.selectDialog.title != "Summarize branch?" {
		t.Fatalf("menu %q", c.modelMenuKind)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEscape})
	if c.modelMenuKind != "tree" || treeContentText(c.treeSelector.list.selectedNode().entry.content) != "first" {
		t.Fatalf("escape: menu %q", c.modelMenuKind)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter})
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter}) // No summary
	if c.sessionID == "A" || c.modelMenuOpen {
		t.Fatalf("session %s menu %v", c.sessionID, c.modelMenuOpen)
	}
	// "Summarize with custom prompt" asks for instructions; cancelling them
	// returns to the choice.
	c.openTreeSelector("")
	treeSelect(t, c, "second")
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyDown})
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyDown})
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter})
	if c.modelMenuKind != "editor-dialog" || c.editorDialog.title != "Custom summarization instructions" {
		t.Fatalf("menu %q", c.modelMenuKind)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEscape})
	if c.modelMenuKind != "select" {
		t.Fatalf("cancel: menu %q", c.modelMenuKind)
	}
}

func treeSelectOnly(t *testing.T, c *chatTUI, text string) {
	t.Helper()
	l := c.treeSelector.list
	for i, f := range l.filtered {
		if treeContentText(f.node.entry.content) == text {
			l.selected = i
			return
		}
	}
	t.Fatalf("no entry %q", text)
}

func runeEvents(s string) []gotui.KeyEvent {
	var out []gotui.KeyEvent
	for _, r := range s {
		out = append(out, gotui.KeyEvent{Key: gotui.KeyRune, Rune: r})
	}
	return out
}
