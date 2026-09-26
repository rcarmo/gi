package tui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/topics"
)

const queuePageSize = 6
const queueUsage = "queue: /queue [page] | /queue remove <turn-id> | /queue steer <queued-id> <active-id>"

func (c *chatTUI) showQueueCommand(lines []string) {
	for i, line := range lines {
		lines[i] = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, line)
	}
	// Explicit command responses cannot wait behind a busy regular transcript:
	// print once above the dock without committing partial model output or
	// adding these lines to the later transcript-flush batch.
	if c.regularMode && c.app != nil && !c.workspaceIndex.active && !c.modelMenuAltScreen {
		c.app.PrintAboveElement(c.renderLineBlock(lines, gotui.NewStyle()))
		return
	}
	c.appendTranscript(lines...)
}

// Deliberately ID-based, not mutable row indices. Native storage/engine guards
// remain authoritative; a failed Steer never becomes a new prompt submission.
func (c *chatTUI) queueCommand(fields []string) []string {
	if c.store == nil || c.engine == nil || c.sessionID == "" {
		return []string{"queue: active session required"}
	}
	ctx := context.Background()
	id := c.sessionID
	page := 1
	if len(fields) > 1 {
		switch fields[1] {
		case "remove":
			if len(fields) != 3 {
				return []string{queueUsage}
			}
			if err := c.store.CancelQueuedTurn(ctx, id, fields[2]); err != nil {
				return []string{"queue: remove failed; refresh /queue: " + err.Error()}
			}
			if bus := c.engine.Topics(); bus != nil {
				bus.Publish(topics.Envelope{Topic: "session.queue", SessionID: id, Type: "notice", Payload: map[string]any{"type": "queue_changed"}})
			}
			return []string{"queue: removed " + fields[2] + "; accepted history and stored media kept"}
		case "steer":
			if len(fields) != 4 {
				return []string{queueUsage}
			}
			if err := c.engine.SteerQueuedTurn(ctx, id, fields[2], fields[3]); err != nil {
				return []string{"queue: steer failed; refresh /queue: " + err.Error()}
			}
			return []string{"queue: bound " + fields[2] + " to " + fields[3] + "; no new prompt submitted"}
		default:
			if len(fields) != 2 {
				return []string{queueUsage}
			}
			value, err := strconv.Atoi(fields[1])
			if err != nil || value < 1 {
				return []string{queueUsage}
			}
			page = value
		}
	}
	items, err := c.store.ListQueuedTurns(ctx, id)
	if err != nil {
		return []string{"queue: read failed: " + err.Error()}
	}
	active := "none"
	running, _, err := c.store.GetSessionActiveTurn(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return []string{"queue: activity read failed: " + err.Error()}
	}
	if running != "" {
		turn, err := c.store.GetTurn(ctx, running)
		if err != nil {
			return []string{"queue: activity read failed: " + err.Error()}
		}
		if turn.Status == "running" {
			active = running
		}
	}
	pages := max(1, (len(items)+queuePageSize-1)/queuePageSize)
	if page > pages {
		return []string{fmt.Sprintf("queue: page out of range; %d page(s). /queue to refresh", pages)}
	}
	lines := []string{fmt.Sprintf("queue: %d queued · page %d/%d · active %s", len(items), page, pages, active)}
	start := (page - 1) * queuePageSize
	for _, item := range items[start:min(len(items), start+queuePageSize)] {
		// IDs are complete; only the advisory prompt preview is shortened/sanitised.
		preview := strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, item.Prompt)
		lines = append(lines, fmt.Sprintf("  %s  %s", item.ID, selectorText(preview, 40)))
	}
	lines = append(lines, queueUsage)
	return lines
}
