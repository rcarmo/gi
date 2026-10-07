package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/sessionexport"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// Pi's /tree keeps a session's branches in one file and moves its leaf.
// gi keeps one message list per session, so the tree is the session's
// branch family: the root session and the copies /tree made of it
// (store.BranchSessionBefore). A copied message has the ID of the message it
// copies, so branches share their common history; navigating switches to the
// session that ends at the chosen point, or copies one up to it. See
// docs/internal/tui-tree.md.

// treePath is one family session: its tree entry IDs, oldest first, the
// session messages they came from, and all its messages (for cut points).
type treePath struct {
	session string
	ids     []string
	local   []store.Message
	all     []store.Message
}

// sessionTree is the branch family of a session as Pi's session tree.
type sessionTree struct {
	root   string // the family's root session (labels live in its state)
	roots  []*treeNode
	nodes  map[string]*treeNode
	paths  []treePath // the current session's first
	leaf   string     // the current session's last entry ("" when empty)
	labels map[string]store.TreeLabel
}

// loadSessionTree builds the tree of sessionID's branch family.
func loadSessionTree(ctx context.Context, s *store.Store, sessionID string) (*sessionTree, error) {
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	byID := map[string]store.Session{}
	children := map[string][]string{}
	for _, sess := range sessions {
		byID[sess.ID] = sess
		if parent := store.TreeParent(sess); parent != "" {
			children[parent] = append(children[parent], sess.ID)
		}
	}
	root := sessionID
	for seen := map[string]bool{}; !seen[root]; {
		seen[root] = true
		parent := store.TreeParent(byID[root])
		if _, ok := byID[parent]; !ok {
			break
		}
		root = parent
	}
	t := &sessionTree{root: root, nodes: map[string]*treeNode{}, labels: store.TreeLabels(byID[root])}
	// Family sessions oldest first, so a message's original is always seen
	// before its copies.
	family := []string{root}
	for i := 0; i < len(family); i++ {
		family = append(family, children[family[i]]...)
	}
	canonical := map[string]string{} // session message ID -> entry ID
	for _, id := range family {
		entries, err := sessionexport.Entries(ctx, s, id)
		if err != nil {
			return nil, err
		}
		all, err := s.ListMessages(ctx, id)
		if err != nil {
			return nil, err
		}
		p := treePath{session: id, all: all}
		parent := ""
		for _, se := range entries {
			m := se.Message
			entryID := m.ID
			if from, _ := m.Payload["forked_from_message_id"].(string); canonical[from] != "" {
				entryID = canonical[from]
			}
			canonical[m.ID] = entryID
			if t.nodes[entryID] == nil {
				t.addNode(treeEntryFromMessage(se), entryID, parent)
			}
			p.ids = append(p.ids, entryID)
			p.local = append(p.local, m)
			parent = entryID
		}
		if id == sessionID {
			t.paths = append([]treePath{p}, t.paths...)
			if len(p.ids) > 0 {
				t.leaf = p.ids[len(p.ids)-1]
			}
		} else {
			t.paths = append(t.paths, p)
		}
	}
	return t, nil
}

// treeEntryFromMessage reads a session message's Pi entry; gi's compaction
// and branch summary messages are Pi's compaction and branch_summary entries.
func treeEntryFromMessage(se sessionexport.SourcedEntry) treeEntry {
	m := se.Message
	switch kind, _ := m.Payload["kind"].(string); {
	case m.Role == "assistant" && kind == "compaction":
		return treeEntry{typ: "compaction", tokensBefore: float64(toInt(m.Payload["tokens_before"], 0)), summary: m.Content}
	case m.Role == "user" && kind == turn.BranchSummaryKind:
		return treeEntry{typ: "branch_summary", summary: m.Content}
	}
	return treeEntryFromPi(se.Entry)
}

func (t *sessionTree) addNode(entry treeEntry, id, parent string) {
	entry.id, entry.parentID = id, parent
	n := &treeNode{entry: entry}
	if l, ok := t.labels[id]; ok {
		n.label, n.labelTime = l.Label, l.Time
	}
	t.nodes[id] = n
	if p := t.nodes[parent]; p != nil {
		p.children = append(p.children, n)
	} else {
		t.roots = append(t.roots, n)
	}
}

// pathTo is the entry IDs from the root to id.
func (t *sessionTree) pathTo(id string) []string {
	var path []string
	for n := t.nodes[id]; n != nil; n = t.nodes[n.entry.parentID] {
		path = append([]string{n.entry.id}, path...)
	}
	return path
}

// target is Pi's navigateTree leaf: a user message's parent, with its text
// for the editor; any other entry itself.
func (t *sessionTree) target(id string) (leaf, editorText string) {
	n := t.nodes[id]
	if n != nil && n.entry.typ == "message" && n.entry.role == "user" {
		return n.entry.parentID, treeContentText(n.entry.content)
	}
	return id, ""
}

// abandoned is Pi's collectEntriesForBranchSummary: the current session's
// messages after its common ancestor with id.
func (t *sessionTree) abandoned(id string) []store.Message {
	if len(t.paths) == 0 || t.leaf == "" {
		return nil
	}
	onTarget := map[string]bool{}
	for _, e := range t.pathTo(id) {
		onTarget[e] = true
	}
	current := t.paths[0]
	start := 0
	for i, e := range current.ids {
		if onTarget[e] {
			start = i + 1
		}
	}
	return current.local[start:]
}

// endingAt is a family session whose messages end at leaf, the current
// session first.
func (t *sessionTree) endingAt(leaf string) string {
	for _, p := range t.paths {
		last := ""
		if len(p.ids) > 0 {
			last = p.ids[len(p.ids)-1]
		}
		if last == leaf {
			return p.session
		}
	}
	return ""
}

// cutAt is a family session holding leaf and the message to copy it up to
// (before; "" copies all of it), the current session first.
func (t *sessionTree) cutAt(leaf string) (source, before string) {
	for _, p := range t.paths {
		if leaf == "" {
			if len(p.all) > 0 {
				return p.session, p.all[0].ID
			}
			continue
		}
		for i, id := range p.ids {
			if id != leaf {
				continue
			}
			local := p.local[i].ID
			for k, m := range p.all {
				if m.ID == local {
					if k+1 < len(p.all) {
						return p.session, p.all[k+1].ID
					}
					return p.session, ""
				}
			}
		}
	}
	return "", ""
}

// openTreeSelector is Pi's showTreeSelector.
func (c *chatTUI) openTreeSelector(initialSelected string) {
	if c.store == nil || c.sessionID == "" {
		c.appendTranscript("sys: No entries in session")
		return
	}
	tree, err := loadSessionTree(context.Background(), c.store, c.sessionID)
	if err != nil {
		c.appendTranscript(fmt.Sprintf("error: tree: %v", err))
		return
	}
	if len(tree.roots) == 0 {
		c.appendTranscript("sys: No entries in session")
		return
	}
	height := 24
	if c.app != nil {
		_, height = c.app.Size()
	}
	sel := newTreeSelector(tree.roots, tree.leaf, height, initialSelected, c.cfg.TreeFilterMode)
	scope := c.selectionScope()
	closeTree := func() {
		c.treeSelector = nil
		c.closeModelMenu()
	}
	sel.onCancel = closeTree
	sel.onSelect = func(id string) {
		closeTree()
		if id == tree.leaf {
			c.appendTranscript("sys: Already at this point")
			return
		}
		if !c.ownsScope(scope) {
			c.appendTranscript("sys: session changed; reopen /tree")
			return
		}
		c.chooseBranchSummary(tree, id)
	}
	sel.onLabel = func(id, label string) {
		if err := c.store.SetTreeLabel(context.Background(), tree.root, id, label, sel.list.now()); err != nil {
			c.selectionNotice("Label not saved: " + err.Error())
		}
	}
	sel.onCopy = func(text string) {
		if text == "" {
			c.selectionNotice("Selected entry has no text to copy")
			return
		}
		lines := c.copyContentLines(text, "selected message")
		if strings.HasPrefix(lines[0], "copy: sent") {
			c.selectionNotice("Copied selected message to clipboard")
			return
		}
		c.appendTranscript(lines...)
		c.selectionNotice(lines[0])
	}
	c.treeSelector = sel
	c.openMenuKind("tree")
}

// chooseBranchSummary is Pi's "Summarize branch?" loop: Escape returns to
// the tree; cancelling the custom prompt returns to the choice.
func (c *chatTUI) chooseBranchSummary(tree *sessionTree, id string) {
	if c.cfg.BranchSummarySkipPrompt {
		c.navigateTree(tree, id, false, "")
		return
	}
	c.openSelect("Summarize branch?", []string{"No summary", "Summarize", "Summarize with custom prompt"}, func(choice string) {
		switch choice {
		case "Summarize with custom prompt":
			c.openEditorDialog("Custom summarization instructions", "", func(instructions string) {
				c.navigateTree(tree, id, true, instructions)
			}, func() { c.chooseBranchSummary(tree, id) })
		default:
			c.navigateTree(tree, id, choice == "Summarize", "")
		}
	}, func() { c.openTreeSelector(id) })
}

// navigateTree is Pi's navigateTree with gi's sessions: with a summary of
// the branch being left, or without one.
func (c *chatTUI) navigateTree(tree *sessionTree, id string, summarize bool, instructions string) {
	leaf, editorText := tree.target(id)
	var abandoned []store.Message
	if summarize {
		abandoned = tree.abandoned(id)
	}
	if len(abandoned) == 0 {
		c.finishTreeNavigation(tree, leaf, editorText, "")
		return
	}
	if c.engine == nil {
		c.appendTranscript("error: No model available for summarization")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.branchSummaryCancel = cancel
	c.markDirty()
	sessionID, model := c.sessionID, c.sessionModelLabel()
	go func() {
		summary, err := c.engine.SummarizeBranch(ctx, sessionID, model, abandoned, instructions)
		cancelled := ctx.Err() != nil
		cancel()
		c.runOnUI(func() {
			c.branchSummaryCancel = nil
			switch {
			case cancelled:
				c.appendTranscript("sys: Branch summarization cancelled")
				c.openTreeSelector(id)
			case err != nil:
				c.appendTranscript("error: " + err.Error())
			case c.sessionID != sessionID:
				c.appendTranscript("sys: session changed; navigation cancelled")
			default:
				c.finishTreeNavigation(tree, leaf, editorText, summary)
			}
		})
	}()
}

// finishTreeNavigation moves to leaf: the session that ends there, or a new
// copy up to it, with the branch summary (if any) added to a new copy.
func (c *chatTUI) finishTreeNavigation(tree *sessionTree, leaf, editorText, summary string) {
	ctx := context.Background()
	target := ""
	if summary == "" {
		target = tree.endingAt(leaf)
	}
	if target == "" {
		source, before := tree.cutAt(leaf)
		if source == "" {
			source = c.sessionID
		}
		branch, err := c.store.BranchSessionBefore(ctx, source, store.NowID("session"), "", "", before)
		if err != nil {
			c.appendTranscript("error: tree: " + err.Error())
			return
		}
		if summary != "" {
			if err := c.store.AddMessage(ctx, store.NowID("msg"), branch.ID, "user", summary, map[string]any{"kind": turn.BranchSummaryKind, "from_id": tree.leaf}); err != nil {
				c.appendTranscript("error: tree: " + err.Error())
				return
			}
		}
		target = branch.ID
	}
	if target != c.sessionID && !c.switchSession(target) {
		return
	}
	c.ensureInput()
	if editorText != "" && strings.TrimSpace(c.input.Text()) == "" {
		c.input.SetText(editorText)
	}
	c.appendTranscript("sys: Navigated to selected point")
	c.markDirty()
}
