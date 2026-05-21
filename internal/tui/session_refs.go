package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func (c *chatTUI) switchSession(sessionID string) {
	c.bindSession(sessionID)
	c.transcript = c.loadTranscript()
	c.draft = ""
	c.draftLineIndex = -1
	c.running = false
	c.status = fmt.Sprintf("%s · %s", c.cfg.AssistantName, c.cfg.DefaultModel)
	c.scrollTranscriptToBottom()
}

func (c *chatTUI) listAgentLines() []string {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	if len(sessions) == 0 {
		return []string{"sys: no sessions"}
	}
	agentIDs := c.sessionAgentIDIndex(sessions)
	lines := []string{"sys: agents:"}
	for _, sess := range sessions {
		marker := " "
		if sess.ID == c.sessionID {
			marker = "*"
		}
		parent := ""
		if sess.ParentSessionID != "" {
			parent = fmt.Sprintf(" parent=%s", sess.ParentSessionID)
		}
		lines = append(lines, fmt.Sprintf("%s @%s %s%s", marker, c.agentIDForSessionFromIndex(&sess, agentIDs), sess.ID, parent))
	}
	return lines
}

func (c *chatTUI) treeLines() []string {
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
	if len(sessions) == 0 {
		return []string{"tree: no sessions"}
	}
	agentIDs := c.sessionAgentIDIndex(sessions)
	children := map[string][]store.Session{}
	roots := []store.Session{}
	for _, sess := range sessions {
		if sess.ParentSessionID == "" {
			roots = append(roots, sess)
			continue
		}
		children[sess.ParentSessionID] = append(children[sess.ParentSessionID], sess)
	}
	lines := []string{"tree: sessions:"}
	seen := map[string]bool{}
	var walk func(sess store.Session, prefix string, last bool)
	walk = func(sess store.Session, prefix string, last bool) {
		seen[sess.ID] = true
		branch := "├─"
		nextPrefix := prefix + "│  "
		if last {
			branch = "└─"
			nextPrefix = prefix + "   "
		}
		marker := " "
		if sess.ID == c.sessionID {
			marker = "*"
		}
		lines = append(lines, fmt.Sprintf("%s%s%s @%s %s", prefix, branch, marker, c.agentIDForSessionFromIndex(&sess, agentIDs), sess.ID))
		kids := children[sess.ID]
		for i, child := range kids {
			walk(child, nextPrefix, i == len(kids)-1)
		}
	}
	for i, root := range roots {
		walk(root, "", i == len(roots)-1)
	}
	for _, sess := range sessions {
		if !seen[sess.ID] {
			lines = append(lines, fmt.Sprintf("? @%s %s parent=%s", c.agentIDForSessionFromIndex(&sess, agentIDs), sess.ID, sess.ParentSessionID))
		}
	}
	return lines
}

func (c *chatTUI) resolveSessionRef(ref string) (*store.Session, error) {
	ref = strings.TrimSpace(strings.TrimPrefix(ref, "@"))
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		return nil, err
	}
	agentIDs := c.sessionAgentIDIndex(sessions)
	for i := range sessions {
		if sessions[i].ID == ref || c.agentIDForSessionFromIndex(&sessions[i], agentIDs) == ref {
			return &sessions[i], nil
		}
	}
	return nil, fmt.Errorf("unknown session or agent: %s", ref)
}

func (c *chatTUI) nextForkAgentID() string {
	current, err := c.store.GetSession(context.Background(), c.sessionID)
	if err != nil {
		return "agent1"
	}
	sessions, err := c.store.ListSessions(context.Background())
	if err != nil {
		base := strings.TrimRight(c.agentIDForSession(current), "0123456789")
		if base == "" {
			base = c.agentIDForSession(current)
		}
		return base + "1"
	}
	agentIDs := c.sessionAgentIDIndex(sessions)
	base := strings.TrimRight(c.agentIDForSessionFromIndex(current, agentIDs), "0123456789")
	if base == "" {
		base = c.agentIDForSessionFromIndex(current, agentIDs)
	}
	used := map[string]bool{}
	for _, sess := range sessions {
		used[c.agentIDForSessionFromIndex(&sess, agentIDs)] = true
	}
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		if !used[candidate] {
			return candidate
		}
	}
	return base + "999"
}

func (c *chatTUI) sessionAgentIDIndex(sessions []store.Session) map[string]string {
	index, err := c.store.ListSessionAgentIDs(context.Background())
	if err != nil || index == nil {
		index = map[string]string{}
	}
	for _, sess := range sessions {
		agentID := strings.TrimSpace(index[sess.ID])
		if agentID == "" {
			agentID = "agent"
		}
		index[sess.ID] = agentID
	}
	return index
}

func (c *chatTUI) agentIDForSessionFromIndex(sess *store.Session, index map[string]string) string {
	if sess != nil {
		if agentID := strings.TrimSpace(index[sess.ID]); agentID != "" {
			return agentID
		}
	}
	return c.agentIDForSession(sess)
}

func (c *chatTUI) agentIDForSession(sess *store.Session) string {
	if sess != nil {
		if agentID := strings.TrimSpace(c.store.SessionAgentID(context.Background(), sess.ID)); agentID != "" {
			return agentID
		}
	}
	return "agent"
}
