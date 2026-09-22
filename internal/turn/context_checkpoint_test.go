package turn

import (
	"context"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestContextCheckpointEligibilityRejectsHookToolAndMediaAmbiguity(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	s.CreateSession(ctx, "A", "A", nil)
	for _, id := range []string{"old", "recent"} {
		if err := s.AddMessage(ctx, id, "A", "user", id, nil); err != nil {
			t.Fatal(err)
		}
	}
	e := New(s)
	defer e.Close()
	r := e.runner("A")
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	conv := &goai.Context{Messages: r.projectContextSnapshot(ctx, "A", snapshot)}
	payload := map[string]any{"messages_to_summarize": 1}
	boundary, err := r.compactionBoundary(ctx, "A", conv, snapshot, payload)
	if err != nil || boundary == nil {
		t.Fatal(boundary, err)
	}
	conv.Messages[0] = goai.UserMessage("hook replacement")
	boundary, err = r.compactionBoundary(ctx, "A", conv, snapshot, payload)
	if err != nil || boundary != nil {
		t.Fatal("hook rewritten context was mapped", boundary, err)
	}
	conv.Messages = r.projectContextSnapshot(ctx, "A", snapshot)
	conv.Messages = append(conv.Messages, goai.UserMessage("live tool result"))
	boundary, err = r.compactionBoundary(ctx, "A", conv, snapshot, payload)
	if err != nil || boundary != nil {
		t.Fatal("live context was mapped", boundary, err)
	}
	media, err := s.CreateMedia(ctx, "A", "notes.txt", "text/plain", []byte("original attachment"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`update messages set payload_json=json_object('media',json_array(json_object('media_id',?))) where id='old'`, media.ID); err != nil {
		t.Fatal(err)
	}
	snapshot, err = s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	conv.Messages = r.projectContextSnapshot(ctx, "A", snapshot)
	boundary, err = r.compactionBoundary(ctx, "A", conv, snapshot, payload)
	if err != nil || boundary != nil {
		t.Fatal("media covered by text-only summary", boundary, err)
	}
}
