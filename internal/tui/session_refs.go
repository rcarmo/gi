package tui

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

const defaultForkAgentID = "agent"
const minForkAgentIDSuffix = 1
const maxForkAgentIDSuffixExclusive = 1000

func forkAgentIDCandidate(base string, suffix int) string {
	return fmt.Sprintf("%s%d", base, suffix)
}

func firstForkAgentID(base string) string {
	return forkAgentIDCandidate(normalizeForkAgentBase(base), minForkAgentIDSuffix)
}

func lastForkAgentID(base string) string {
	return forkAgentIDCandidate(normalizeForkAgentBase(base), maxForkAgentIDSuffixExclusive-1)
}

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
	agentIDs, err := c.sessionAgentIDIndex(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
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
	agentIDs, err := c.sessionAgentIDIndex(context.Background())
	if err != nil {
		return []string{fmt.Sprintf("error: %v", err)}
	}
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

func normalizeIndexedSessionID(sessionID string) string {
	return strings.TrimSpace(sessionID)
}

func indexedAgentID(sessionID string, index map[string]string) string {
	return strings.TrimSpace(index[normalizeIndexedSessionID(sessionID)])
}

func indexedAgentIDLower(sessionID string, index map[string]string) string {
	return strings.ToLower(indexedAgentIDOrDefault(sessionID, index))
}

func indexedAgentIDOrDefault(sessionID string, index map[string]string) string {
	if normalizeIndexedSessionID(sessionID) == "" {
		return defaultForkAgentID
	}
	agentID := indexedAgentID(sessionID, index)
	if agentID == "" {
		return defaultForkAgentID
	}
	return agentID
}

func unknownSessionOrAgentError(ref string) error {
	return fmt.Errorf("unknown session or agent: %s", ref)
}

func normalizeSessionRef(ref string) string {
	ref = strings.TrimSpace(ref)
	return strings.TrimPrefix(ref, "@")
}

func normalizeSessionRefLower(ref string) string {
	return strings.ToLower(normalizeSessionRef(ref))
}

func (c *chatTUI) resolveSessionRef(ref string) (*store.Session, error) {
	ctx := context.Background()
	ref = normalizeSessionRef(ref)
	if ref == "" {
		return nil, unknownSessionOrAgentError(ref)
	}
	if sess, err := c.store.GetSession(ctx, ref); err == nil {
		return sess, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}
	sessionIDs, agentIDs, err := c.loadSessionIdentityIndex(ctx)
	if err != nil {
		return nil, err
	}
	if sessionID, ok := findSessionIDByNormalizedAgentRef(sessionIDs, agentIDs, normalizeSessionRefLower(ref)); ok {
		return c.store.GetSession(ctx, sessionID)
	}
	return nil, unknownSessionOrAgentError(ref)
}

func normalizeForkAgentBase(agentID string) string {
	agentID = strings.TrimSpace(agentID)
	base := strings.TrimRight(agentID, "0123456789")
	if base == "" {
		base = agentID
	}
	if base == "" {
		base = defaultForkAgentID
	}
	return base
}

func buildUsedForkAgentIDs(sessionIDs []string, agentIDs map[string]string) map[string]bool {
	if len(sessionIDs) == 0 {
		return map[string]bool{}
	}
	used := make(map[string]bool, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		normalizedSessionID := normalizeIndexedSessionID(sessionID)
		if normalizedSessionID == "" {
			continue
		}
		used[indexedAgentIDOrDefault(normalizedSessionID, agentIDs)] = true
	}
	return used
}

func findSessionIDByNormalizedAgentRef(sessionIDs []string, agentIDs map[string]string, normalizedRef string) (string, bool) {
	if normalizedRef == "" {
		return "", false
	}
	for _, sessionID := range sessionIDs {
		if indexedAgentIDLower(sessionID, agentIDs) == normalizedRef {
			return sessionID, true
		}
	}
	return "", false
}

func chooseNextForkAgentID(base string, used map[string]bool) (string, bool) {
	base = normalizeForkAgentBase(base)
	if used == nil {
		used = map[string]bool{}
	}
	for i := minForkAgentIDSuffix; i < maxForkAgentIDSuffixExclusive; i++ {
		candidate := forkAgentIDCandidate(base, i)
		if !used[candidate] {
			return candidate, true
		}
	}
	return "", false
}

func (c *chatTUI) nextForkAgentID() string {
	ctx := context.Background()
	base := c.store.SessionAgentID(ctx, c.sessionID)
	sessionIDs, agentIDs, err := c.loadSessionIdentityIndex(ctx)
	if err != nil {
		return firstForkAgentID(base)
	}
	used := buildUsedForkAgentIDs(sessionIDs, agentIDs)
	if candidate, ok := chooseNextForkAgentID(base, used); ok {
		return candidate
	}
	return lastForkAgentID(base)
}

func (c *chatTUI) sessionAgentIDIndex(ctx context.Context) (map[string]string, error) {
	index, err := c.store.ListSessionAgentIDs(ctx)
	if err != nil {
		return nil, err
	}
	if index == nil {
		return map[string]string{}, nil
	}
	return index, nil
}

func (c *chatTUI) loadSessionIdentityIndex(ctx context.Context) ([]string, map[string]string, error) {
	sessionIDs, err := c.store.ListSessionIDs(ctx)
	if err != nil {
		return nil, nil, err
	}
	agentIDs, err := c.sessionAgentIDIndex(ctx)
	if err != nil {
		return nil, nil, err
	}
	return sessionIDs, agentIDs, nil
}

func (c *chatTUI) agentIDForSessionFromIndex(sess *store.Session, index map[string]string) string {
	if sess != nil {
		return indexedAgentIDOrDefault(sess.ID, index)
	}
	return defaultForkAgentID
}

func (c *chatTUI) agentIDForSession(sess *store.Session) string {
	if sess != nil {
		return c.store.SessionAgentID(context.Background(), sess.ID)
	}
	return defaultForkAgentID
}
