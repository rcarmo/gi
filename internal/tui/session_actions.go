package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func (c *chatTUI) sessionPickerLabel(sess *store.Session) string {
	status, _ := sess.State["status"].(string)
	if status == "" {
		status = "idle"
	}
	if archived, _ := sess.State["archived_at"].(string); archived != "" {
		status = "archived"
	}
	if pinned, _ := sess.State["pinned"].(bool); pinned {
		status += " · pinned"
	}
	return fmt.Sprintf("@%s · %s · %s (%s)", c.agentIDForSession(sess), status, strings.TrimSpace(sess.Title), sess.ID)
}

type sessionActions struct {
	target           string
	targetLabel      string
	labels, choices  []string
	values           map[string]string
	query            string
	selected, scroll int
	mutations        map[string]store.SessionMutation
}

func (c *chatTUI) openSessionActions() {
	if !c.modelMenuOpen || c.modelMenuKind != "session" || c.modelMenuSelected < 0 || c.modelMenuSelected >= len(c.modelMenuChoices) {
		return
	}
	if !c.ownsScope(c.modelMenuSession) {
		c.modelMenuError = "session changed; reopen picker"
		return
	}
	target := c.modelMenuValues[c.modelMenuChoices[c.modelMenuSelected]]
	caps, err := c.store.SessionDisplayCapabilities(context.Background(), target)
	if err != nil {
		c.modelMenuError = "session unavailable; Right to retry or Esc to close"
		return
	}
	state := sessionActions{target: target, targetLabel: c.modelMenuChoices[c.modelMenuSelected], labels: c.modelMenuAll, choices: c.modelMenuChoices, values: c.modelMenuValues, query: c.modelMenuQuery, selected: c.modelMenuSelected, scroll: c.modelMenuScroll, mutations: map[string]store.SessionMutation{}}
	labels := []string{}
	if caps.CanPin {
		label := "Pin"
		pinned := true
		if caps.Pinned {
			label = "Unpin"
			pinned = false
		}
		labels = append(labels, label)
		state.mutations[label] = store.SessionMutation{Action: "pin", Pinned: &pinned}
	}
	if caps.CanArchive {
		labels = append(labels, "Archive")
		state.mutations["Archive"] = store.SessionMutation{Action: "archive"}
	}
	if caps.CanRestore {
		labels = append(labels, "Restore")
		state.mutations["Restore"] = store.SessionMutation{Action: "restore"}
	}
	c.sessionActions = state
	c.modelMenuKind = "session-actions"
	c.modelMenuAll = labels
	c.modelMenuChoices = labels
	c.modelMenuValues = nil
	c.modelMenuQuery = ""
	c.modelMenuSelected = 0
	c.modelMenuScroll = 0
	c.modelMenuError = ""
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) backFromSessionActions() {
	if c.modelMenuKind != "session-actions" {
		c.closeModelMenu()
		return
	}
	state := c.sessionActions
	c.sessionActions = sessionActions{}
	c.modelMenuKind = "session"
	c.modelMenuAll = state.labels
	c.modelMenuChoices = state.choices
	c.modelMenuValues = state.values
	c.modelMenuQuery = state.query
	c.modelMenuSelected = state.selected
	c.modelMenuScroll = state.scroll
	c.modelMenuError = ""
	if c.app != nil {
		c.app.MarkDirty()
	}
}

func (c *chatTUI) applySessionAction() {
	if !c.ownsScope(c.modelMenuSession) {
		c.modelMenuError = "session changed; reopen picker"
		return
	}
	if c.modelMenuSelected < 0 || c.modelMenuSelected >= len(c.modelMenuChoices) {
		return
	}
	mutation, ok := c.sessionActions.mutations[c.modelMenuChoices[c.modelMenuSelected]]
	if !ok {
		return
	}
	if err := c.store.MutateSession(context.Background(), c.sessionActions.target, mutation); err != nil {
		// Return to the original picker selection. A fresh Right re-reads capabilities,
		// so a concurrent archive/claim cannot leave a stale action retry loop.
		c.backFromSessionActions()
		if errors.Is(err, store.ErrSessionMutationConflict) || errors.Is(err, store.ErrSessionMutationInvalid) {
			c.modelMenuError = "action unavailable; Right to refresh or Esc to close"
		} else {
			c.modelMenuError = "storage error; not confirmed, reopen to check"
		}
		return
	}
	target := c.sessionActions.target
	c.backFromSessionActions()
	// Update only the captured target's label, without reordering or filtering
	// away the selected row. Preserve the query and viewport for the next action.
	sess, err := c.store.GetSession(context.Background(), target)
	if err != nil {
		c.modelMenuError = "saved; metadata unavailable, reopen picker"
		return
	}
	label := c.sessionPickerLabel(sess)
	for old, id := range c.modelMenuValues {
		if id == target {
			for i, v := range c.modelMenuAll {
				if v == old {
					c.modelMenuAll[i] = label
				}
			}
			for i, v := range c.modelMenuChoices {
				if v == old {
					c.modelMenuChoices[i] = label
				}
			}
			delete(c.modelMenuValues, old)
			c.modelMenuValues[label] = target
			break
		}
	}
}
