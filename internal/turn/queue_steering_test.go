package turn

import (
	"context"
	"encoding/base64"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
	"testing"
)

func TestQueueSteerNativeCheckpointKeepsMediaAndAtMostOnce(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	e := New(s)
	defer e.Close()
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	media, err := s.CreateMedia(ctx, "A", "image.png", "image/png", png, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "run", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "q", "A", "queued", "describe", map[string]any{"media": []any{map[string]any{"id": store.MediaRefID(media.ID)}}, "turn_id": "spoof"}); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if err = e.SteerQueuedTurn(ctx, "A", "q", "run"); err != nil {
		t.Fatal(err)
	}
	runner := e.runner("A")
	msgs, err := runner.dequeueSteeringMessages(ctx, "A", "run")
	if err != nil || len(msgs) != 1 {
		t.Fatal(msgs, err)
	}
	conv := &goai.Context{}
	// Persistence failure must not inject, and release must still recover.
	if _, err = s.DB().Exec(`create trigger reject_message before insert on messages begin select raise(abort,'injected failure'); end`); err != nil {
		t.Fatal(err)
	}
	if n := runner.injectSteeringMessages(ctx, "A", "run", conv, msgs); n != 0 || len(conv.Messages) != 0 {
		t.Fatal(n, conv)
	}
	if _, err = s.DB().Exec(`drop trigger reject_message`); err != nil {
		t.Fatal(err)
	}
	if n := runner.injectSteeringMessages(ctx, "A", "run", conv, msgs); n != 1 {
		t.Fatal(n)
	}
	if n := runner.injectSteeringMessages(ctx, "A", "run", conv, msgs); n != 0 || len(conv.Messages) != 1 {
		t.Fatal(n, conv)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "cancel-q", "A", "queued", "keep on cancel", nil); err != nil {
		t.Fatal(err)
	}
	if err := e.SteerQueuedTurn(ctx, "A", "cancel-q", "run"); err != nil {
		t.Fatal(err)
	}
	pending, err := runner.dequeueSteeringMessages(ctx, "A", "run")
	if err != nil || len(pending) != 1 {
		t.Fatal(pending, err)
	}
	if err := s.UpdateTurnStatus(ctx, "run", "cancelling"); err != nil {
		t.Fatal(err)
	}
	if n := runner.injectSteeringMessages(ctx, "A", "run", conv, pending); n != 0 || len(conv.Messages) != 1 {
		t.Fatal(n, conv)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "A", "run"); err != nil {
		t.Fatal(err)
	}
	q, err := s.GetTurn(ctx, "cancel-q")
	if err != nil || q.Status != "queued" || q.Phase != "steer_returned" {
		t.Fatal(q, err)
	}
	if len(conv.Messages[0].Content) != 2 || conv.Messages[0].Content[1].Data != base64.StdEncoding.EncodeToString(png) {
		t.Fatal(conv)
	}
	saved, err := s.ListMessages(ctx, "A")
	if err != nil || len(saved) != 1 || saved[0].Payload["turn_id"] != "run" {
		t.Fatal(saved, err)
	}
}
