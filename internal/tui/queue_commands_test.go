package tui

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

func TestTUIQueueCommandsDurablePagesAndGuardedRemoval(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	c.input.SetText("untouched draft 中文🙂")
	c.input.cursorPos = 4
	c.queuedDrafts = []string{"local shortcut"}
	for i := 0; i < 8; i++ {
		if _, err := c.store.CreateTurnWithStatus(ctx, fmt.Sprintf("queue-%02d", i), c.sessionID, "queued", "line\ncontrol\x1b preview", nil); err != nil {
			t.Fatal(err)
		}
	}
	before := *c.input
	out := c.queueCommand([]string{"/queue"})
	if len(out) != 8 || !strings.Contains(out[0], "8 queued · page 1/2") || strings.Contains(strings.Join(out, ""), "\x1b") {
		t.Fatal(out)
	}
	if len(c.queueCommand([]string{"/queue", "2"})) != 4 {
		t.Fatal("paging")
	}
	if !strings.Contains(c.queueCommand([]string{"/queue", "9999999999999999999999"})[0], "/queue [page]") {
		t.Fatal("overflow page")
	}
	if !strings.Contains(c.queueCommand([]string{"/queue", "3"})[0], "out of range") {
		t.Fatal("page bounds")
	}
	// Another frontend reads the same durable rows without restoring local drafts.
	other := &chatTUI{store: c.store, engine: c.engine, sessionID: c.sessionID}
	if other.queueCommand([]string{"/queue"})[0] != out[0] {
		t.Fatal("durable view")
	}
	ch, cancel := c.engine.Topics().Subscribe(ctx, "session.queue", topics.SubscribeOptions{SessionID: c.sessionID, Buffer: 4})
	defer cancel()
	out = c.queueCommand([]string{"/queue", "remove", "queue-01"})
	if !strings.Contains(out[0], "removed queue-01") {
		t.Fatal(out)
	}
	select {
	case e := <-ch:
		if e.SessionID != c.sessionID {
			t.Fatal(e)
		}
	default:
		t.Fatal("no queue notice")
	}
	if !strings.Contains(c.queueCommand([]string{"/queue", "remove", "queue-01"})[0], "failed") {
		t.Fatal("double removal")
	}
	c.store.CreateTurnWithStatus(ctx, "foreign", "B", "queued", "other", nil)
	if !strings.Contains(c.queueCommand([]string{"/queue", "remove", "foreign"})[0], "failed") {
		t.Fatal("foreign removed")
	}
	c.store.ClaimSessionActiveTurn(ctx, c.sessionID, "queue-00", "worker", "claim")
	if !strings.Contains(c.queueCommand([]string{"/queue", "remove", "queue-00"})[0], "failed") {
		t.Fatal("claimed removed")
	}
	if before.text != c.input.text || before.cursorPos != c.input.cursorPos || c.queuedDrafts[0] != "local shortcut" {
		t.Fatal("draft changed")
	}
}

func TestTUIQueueCommandsRunBoundSteerRetainsMetadataAndNoFallback(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	id := c.sessionID
	c.store.CreateTurnWithStatus(ctx, "running", id, "running", "active", nil)
	if ok, err := c.store.ClaimSessionActiveTurn(ctx, id, "running", "worker", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	media, err := c.store.CreateMedia(ctx, id, "queue.txt", "text/plain", []byte("queued bytes"), nil)
	if err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"media": []string{store.MediaRefID(media.ID)}, "tui_media_claim": "accepted-media", "custom": "kept"}
	c.store.CreateTurnWithStatus(ctx, "queued", id, "queued", "followup", meta)
	if !strings.Contains(c.queueCommand([]string{"/queue"})[0], "active running") {
		t.Fatal("active id missing")
	}
	for _, args := range [][]string{{"/queue", "steer", "queued", "stale"}, {"/queue", "steer", "queued"}, {"/queue", "retry", "queued"}} {
		out := c.queueCommand(args)
		if strings.Contains(out[0], "queue: bound") {
			t.Fatal("unsupported/stale action", out)
		}
	}
	out := c.queueCommand([]string{"/queue", "steer", "queued", "running"})
	if !strings.Contains(out[0], "bound queued to running") {
		t.Fatal(out)
	}
	var payload string
	if err = c.store.DB().QueryRow(`SELECT payload_json FROM steering_queue WHERE session_id=?`, id).Scan(&payload); err != nil || !strings.Contains(payload, "accepted-media") || !strings.Contains(payload, store.MediaRefID(media.ID)) || !strings.Contains(payload, "kept") {
		t.Fatal(payload, err)
	}
	if !strings.Contains(c.queueCommand([]string{"/queue", "steer", "queued", "running"})[0], "failed") {
		t.Fatal("duplicate steer")
	}
	rows, _ := c.store.ListTurns(ctx, id)
	if len(rows) != 2 {
		t.Fatal("steer created prompt", rows)
	}
	// Cancelled active IDs cannot become an implicit fresh prompt.
	c.store.CreateTurnWithStatus(ctx, "later", id, "queued", "later", nil)
	c.store.UpdateTurnStatusAndPhase(ctx, "running", "cancelled", "aborted")
	if !strings.Contains(c.queueCommand([]string{"/queue", "steer", "later", "running"})[0], "failed") {
		t.Fatal("idle fallback")
	}
	later, _ := c.store.GetTurn(ctx, "later")
	if later.Status != "queued" {
		t.Fatal("failed steer consumed row")
	}
}

func TestTUIQueueMoveSnapshotOwnershipAndPreservedHistory(t *testing.T) {
	c := sessionTestChat(t)
	ctx := context.Background()
	for _, id := range []string{"one", "two", "three"} {
		if _, err := c.store.CreateTurnWithStatus(ctx, id, c.sessionID, "queued", id, map[string]any{"media": []string{"media:1"}, "custom": "kept"}); err != nil {
			t.Fatal(err)
		}
	}
	c.input.SetText("retain editor Ω")
	c.input.cursorPos = 4
	before, _ := c.store.GetTurn(ctx, "one")
	move := []string{"/queue", "move", "three", "before", "one"}
	if !strings.Contains(c.queueCommand(move)[0], "inspect /queue") {
		t.Fatal("move without snapshot")
	}
	c.queueCommand([]string{"/queue"})
	if !strings.Contains(c.queueCommand(move)[0], "moved three before one") {
		t.Fatal("move")
	}
	if c.queueSnapshot != nil {
		t.Fatal("success retained snapshot")
	}
	rows, _ := c.store.ListQueuedTurns(ctx, c.sessionID)
	if rows[0].ID != "three" || rows[1].ID != "one" || rows[2].ID != "two" {
		t.Fatal(rows)
	}
	after, _ := c.store.GetTurn(ctx, "one")
	if before.CreatedAt != after.CreatedAt || before.UpdatedAt != after.UpdatedAt || before.Prompt != after.Prompt || !reflect.DeepEqual(before.Metadata, after.Metadata) {
		t.Fatal("history/metadata rewritten")
	}
	c.queueCommand([]string{"/queue"})
	c.sessionGeneration += 2 // A-B-A must not reuse the first visit's display.
	if !strings.Contains(c.queueCommand([]string{"/queue", "move", "one", "before", "three"})[0], "inspect /queue") {
		t.Fatal("old visit reordered")
	}
	c.queueCommand([]string{"/queue"})
	c.store.CreateTurnWithStatus(ctx, "later", c.sessionID, "queued", "later", nil)
	if !strings.Contains(c.queueCommand([]string{"/queue", "move", "one", "before", "three"})[0], "move failed") {
		t.Fatal("stale queue reordered")
	}
	if c.queueSnapshot != nil {
		t.Fatal("conflict kept snapshot")
	}
	c.queueCommand([]string{"/queue"})
	if !strings.Contains(c.queueCommand([]string{"/queue", "move", "three", "after", "two"})[0], "moved") {
		t.Fatal("after move")
	}
	c.queueCommand([]string{"/queue"})
	if !strings.Contains(c.queueCommand([]string{"/queue", "move", "later", "after", "two"})[0], "moved") {
		t.Fatal("tail move")
	}
	for _, fields := range [][]string{{"/queue", "move", "one", "before", "one"}, {"/queue", "move", "missing", "before", "one"}, {"/queue", "move", "one", "around", "two"}} {
		c.queueCommand([]string{"/queue"})
		if strings.Contains(c.queueCommand(fields)[0], "queue: moved") {
			t.Fatal("invalid move")
		}
	}
	if c.input.Text() != "retain editor Ω" || c.input.cursorPos != 4 {
		t.Fatal("draft touched")
	}
}
