package store

import (
	"context"
	"errors"
	"testing"
)

func TestRestoreAllQueuedTextSteeringAndFollowUpAtomic(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			ctx := context.Background()
			s, err := Open(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			draft, err := s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: "newer 草稿", Cursor: 2})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateTurnWithStatus(ctx, "follow", "A", "queued", "follow up", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := s.EnqueueSteering(ctx, "A", "", "user", "steer first", nil, nil, "one-at-a-time"); err != nil {
				t.Fatal(err)
			}
			if fail {
				if _, err := s.DB().Exec(`create trigger reject_queue_restore before update on turns when new.status='cancelled' begin select raise(abort,'injected rollback'); end`); err != nil {
					t.Fatal(err)
				}
			}
			got, ids, err := s.RestoreAllQueuedTUITextDraft(ctx, "A", draft)
			if fail {
				if err == nil {
					t.Fatal("expected failure")
				}
				d, _ := s.LoadTUITextDraft(ctx, "A")
				queued, _ := s.ListQueuedTurns(ctx, "A")
				pending, _ := s.ListPendingTUIMessages(ctx, "A")
				if d.Text != draft.Text || len(queued) != 1 || len(pending) != 2 {
					t.Fatal("failed restore partially committed", d, queued, pending)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Text != "steer first\n\nfollow up\n\nnewer 草稿" || len(ids) != 2 {
				t.Fatalf("wrong edit-all order: %+v %v", got, ids)
			}
			pending, _ := s.ListPendingTUIMessages(ctx, "A")
			queued, _ := s.ListQueuedTurns(ctx, "A")
			if len(pending) != 0 || len(queued) != 0 {
				t.Fatal("restored text remains deliverable")
			}
		})
	}
}

func TestRestoreAllQueuedTextRejectsMedia(t *testing.T) {
	ctx := context.Background()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	draft, _ := s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: "draft", Cursor: 5})
	if _, err := s.EnqueueSteering(ctx, "A", "", "user", "with media", nil, []string{"image-id"}, "one-at-a-time"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.RestoreAllQueuedTUITextDraft(ctx, "A", draft); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("media lost", err)
	}
	pending, _ := s.ListPendingTUIMessages(ctx, "A")
	if len(pending) != 1 {
		t.Fatal("media queue changed")
	}
}

func TestRestoreAllQueuedBoundSteeringOwnership(t *testing.T) {
	for _, claimed := range []bool{false, true} {
		s, err := Open(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
			t.Fatal(err)
		}
		for _, row := range []struct{ id, status string }{{"run", "running"}, {"source", "queued"}} {
			if _, err := s.CreateTurnWithStatus(ctx, row.id, "A", row.status, row.id, nil); err != nil {
				t.Fatal(err)
			}
		}
		if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "fixture", "token"); err != nil || !ok {
			t.Fatal(ok, err)
		}
		if err := s.SteerQueuedTurn(ctx, "A", "source", "run"); err != nil {
			t.Fatal(err)
		}
		draft, _ := s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: "draft", Cursor: 5})
		if claimed {
			if _, err := s.DequeueSteeringForTurn(ctx, "A", "run"); err != nil {
				t.Fatal(err)
			}
		}
		got, ids, err := s.RestoreAllQueuedTUITextDraft(ctx, "A", draft)
		if claimed {
			if !errors.Is(err, ErrQueueConflict) {
				t.Fatal("withdrew already claimed steering", err)
			}
			stored, _ := s.LoadTUITextDraft(ctx, "A")
			if stored.Text != "draft" {
				t.Fatal("failed restore changed draft")
			}
		} else {
			if err != nil || got.Text != "source\n\ndraft" || len(ids) != 1 || ids[0] != "source" {
				t.Fatal(got, ids, err)
			}
			source, _ := s.GetTurn(ctx, "source")
			if source.Status != "cancelled" {
				t.Fatal("bound source can redeliver")
			}
			messages, err := s.DequeueSteeringForTurn(ctx, "A", "run")
			if err == nil && len(messages) != 0 {
				t.Fatal("steering redelivered", messages)
			}
		}
		run, _ := s.GetTurn(ctx, "run")
		if run.Status != "running" {
			t.Fatal("Alt+Up aborted the active run")
		}
		s.Close()
	}
}
