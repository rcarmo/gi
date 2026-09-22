package tui

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type terminalCompaction struct {
	turnID      string
	sequence    int
	started     time.Time
	active      bool
	cancelling  bool
	notice      string
	noticeUntil time.Time
}

// startCompaction does not touch editor, cursor, undo, history or queued drafts.
// Only admission is synchronous; the shared engine owns asynchronous execution.
func (c *chatTUI) startCompaction() {
	if c.engine == nil || c.store == nil || c.sessionID == "" {
		c.compactionFeedback("Compact unavailable")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	state, err := c.engine.ManualCompactionState(ctx, c.sessionID)
	if err != nil {
		c.compactionFeedback("Compact failed: " + err.Error())
		return
	}
	if state["available"] != true {
		reason, _ := state["reason"].(string)
		c.compactionFeedback("Compact unavailable: " + reason)
		return
	}
	token, _ := state["token"].(string)
	_, err = c.engine.SubmitManualCompaction(ctx, c.sessionID, token)
	if err != nil {
		c.compactionFeedback("Compact failed: " + err.Error())
		return
	}
	c.compaction = terminalCompaction{}
	c.syncCompactionActivity()
	if c.app != nil {
		c.app.MarkDirty()
	}
}
func (c *chatTUI) compactionFeedback(text string) {
	c.compaction.notice = text
	c.compaction.noticeUntil = time.Now().Add(4 * time.Second)
	if c.app != nil {
		c.app.MarkDirty()
	}
}

// Topic events only invalidate this native snapshot. Delayed events cannot
// repaint an old occurrence as active or restore a previous session's outcome.
func (c *chatTUI) syncCompactionActivity() {
	if c.store == nil || c.sessionID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	state, err := c.store.SessionActivity(ctx, c.sessionID)
	if err != nil {
		c.compactionFeedback("Compact state unavailable: " + err.Error())
		return
	}
	status, _ := state["status"].(string)
	raw, _ := state["compaction"].(map[string]any)
	run, _ := state["turn_id"].(string)
	active := raw != nil && raw["active"] == true
	if active {
		start, _ := raw["timestamp"].(string)
		ts, _ := time.Parse(time.RFC3339Nano, start)
		c.compaction = terminalCompaction{turnID: run, sequence: intFromAny(raw["seq"]), started: ts, active: true, cancelling: status == "cancelling", notice: c.compaction.notice, noticeUntil: c.compaction.noticeUntil}
		c.running = true
		c.clearThinkingIndicator()
		return
	}
	wasActive := c.compaction.active
	c.compaction.active = false
	if wasActive {
		c.running = status == "running" || status == "cancelling"
	}
	if raw == nil {
		return
	}
	seq := intFromAny(raw["seq"])
	if c.compaction.turnID == run && c.compaction.sequence == seq {
		return
	}
	c.compaction.turnID = run
	c.compaction.sequence = seq
	timestamp, _ := raw["timestamp"].(string)
	at, _ := time.Parse(time.RFC3339Nano, timestamp)
	if at.IsZero() || time.Since(at) > 4*time.Second {
		return
	}
	event, _ := raw["event_type"].(string)
	detail, _ := raw["detail"].(string)
	label := ""
	switch event {
	case "compaction.completed":
		label = "Context compacted"
	case "compaction.cancelled":
		label = "Compaction cancelled"
	case "compaction.suppressed":
		label = "Compaction temporarily suppressed"
	case "compaction.failed":
		label = "Compaction failed"
	}
	if label != "" {
		if strings.TrimSpace(detail) != "" {
			label += " · " + detail
		}
		c.compaction.notice = label
		c.compaction.noticeUntil = at.Add(4 * time.Second)
	}
}
func (c *chatTUI) compactionInline() string {
	if !c.compaction.active {
		return ""
	}
	seconds := 0
	if !c.compaction.started.IsZero() {
		seconds = max(0, int(time.Since(c.compaction.started).Seconds()))
	}
	label := "Compacting"
	if c.compaction.cancelling {
		label = "Cancelling compact"
	}
	return fmt.Sprintf("%s %d:%02d", label, seconds/60, seconds%60)
}
func (c *chatTUI) handleCompactionEscape() bool {
	if !c.compaction.active {
		return false
	}
	c.stopCompaction()
	return true
}

func (c *chatTUI) stopCompaction() {
	if c.engine == nil || !c.compaction.active {
		return
	}
	// Do not clear running/editor optimistically. Outcome comes from native state.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := c.engine.CancelActiveTurn(ctx, c.sessionID, c.compaction.turnID)
	c.syncCompactionActivity()
	if err != nil {
		c.compactionFeedback("Compact stop failed: " + err.Error())
	}
}
