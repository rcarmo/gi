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
const queueUsage = "queue: /queue [page] | remove <id> | steer <id> <active-id> | move <id> before|after <target-id>"

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
		case "move":
			if len(fields) != 5 || (fields[3] != "before" && fields[3] != "after") {
				return []string{queueUsage}
			}
			if !c.ownsScope(c.queueSnapshotScope) || len(c.queueSnapshot) == 0 {
				return []string{"queue: inspect /queue in this session before moving rows"}
			}
			order, err := moveQueuedID(c.queueSnapshot, fields[2], fields[4], fields[3] == "after")
			if err != nil {
				return []string{"queue: " + err.Error()}
			}
			expected := c.queueSnapshot
			c.queueSnapshot = nil
			if err = c.store.ReorderQueuedTurns(ctx, id, expected, order); err != nil {
				return []string{"queue: move failed; refresh /queue: " + err.Error()}
			}
			c.publishQueueCommandChange(id)
			return []string{"queue: moved " + fields[2] + " " + fields[3] + " " + fields[4] + "; /queue to refresh"}
		case "remove":
			if len(fields) != 3 {
				return []string{queueUsage}
			}
			c.queueSnapshot = nil
			if err := c.store.CancelQueuedTurn(ctx, id, fields[2]); err != nil {
				return []string{"queue: remove failed; refresh /queue: " + err.Error()}
			}
			c.publishQueueCommandChange(id)
			return []string{"queue: removed " + fields[2] + "; accepted history and stored media kept"}
		case "steer":
			if len(fields) != 4 {
				return []string{queueUsage}
			}
			c.queueSnapshot = nil
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
	c.queueSnapshot = nil
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
	c.queueSnapshotScope = c.selectionScope()
	c.queueSnapshot = make([]string, len(items))
	for i, item := range items {
		c.queueSnapshot[i] = item.ID
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

func (c *chatTUI) publishQueueCommandChange(id string) {
	if bus := c.engine.Topics(); bus != nil {
		bus.Publish(topics.Envelope{Topic: "session.queue", SessionID: id, Type: "notice", Payload: map[string]any{"type": "queue_changed"}})
	}
}

func moveQueuedID(snapshot []string, id, target string, after bool) ([]string, error) {
	if id == target {
		return nil, fmt.Errorf("choose two different queued IDs")
	}
	found, anchor := false, false
	for _, value := range snapshot {
		found = found || value == id
		anchor = anchor || value == target
	}
	if !found || !anchor {
		return nil, fmt.Errorf("IDs are not both in the displayed snapshot; refresh /queue")
	}
	order := make([]string, 0, len(snapshot))
	for _, value := range snapshot {
		if value == id {
			continue
		}
		if value == target && !after {
			order = append(order, id)
		}
		order = append(order, value)
		if value == target && after {
			order = append(order, id)
		}
	}
	same := true
	for i := range order {
		same = same && order[i] == snapshot[i]
	}
	if same {
		return nil, fmt.Errorf("queue already has that order")
	}
	return order, nil
}
