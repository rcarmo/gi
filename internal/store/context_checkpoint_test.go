package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func TestContextCheckpointReopenRepeatedAndHistoryIntact(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "context.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { s.Close() }()
	for _, id := range []string{"A", "B"} {
		if _, err = s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	s.CreateTurn(ctx, "run", "A", "prompt", nil)
	s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run")
	for i := 0; i < 6; i++ {
		if err = s.AddMessage(ctx, fmt.Sprintf("m%d", i), "A", "user", fmt.Sprintf("text%d", i), nil); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := PrepareContextBoundary(snapshot, 4)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out, err := s.FinishCompactionWithBoundary(ctx, "A", "run", seq, "completed", "first", nil, boundary); err != nil || out != "completed" {
		t.Fatal(out, err)
	}
	all, err := s.ListMessages(ctx, "A")
	if err != nil || len(all) != 7 {
		t.Fatal(len(all), err)
	}
	for i := 0; i < 6; i++ {
		var text string
		if err = s.DB().QueryRow(`select content from messages where id=?`, fmt.Sprintf("m%d", i)).Scan(&text); err != nil || text != fmt.Sprintf("text%d", i) {
			t.Fatal(text, err)
		}
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err = s.ContextSnapshot(ctx, "A")
	if err != nil || snapshot.Version != 1 || snapshot.Summary != "first" || len(snapshot.Messages) != 2 {
		t.Fatal(snapshot, err)
	}
	other, err := s.ContextSnapshot(ctx, "B")
	if err != nil || other.Summary != "" || len(other.Messages) != 0 {
		t.Fatal(other, err)
	}
	// Backdated concurrent/new messages are retained by ID; no timestamp cutoff.
	if _, err = s.DB().Exec(`insert into messages(id,session_id,role,content,payload_json,created_at) values('backdated','A','user','new request','{}','2000-01-01')`); err != nil {
		t.Fatal(err)
	}
	snapshot, err = s.ContextSnapshot(ctx, "A")
	if err != nil || snapshot.Messages[0].ID != "backdated" {
		t.Fatal(snapshot, err)
	}
	boundary, err = PrepareContextBoundary(snapshot, 2)
	if err != nil {
		t.Fatal(err)
	} // virtual summary + backdated
	seq, err = s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out, err := s.FinishCompactionWithBoundary(ctx, "A", "run", seq, "completed", "second", nil, boundary); err != nil || out != "completed" {
		t.Fatal(out, err)
	}
	snapshot, err = s.ContextSnapshot(ctx, "A")
	if err != nil || snapshot.Summary != "second" || len(snapshot.Covered) != 5 || len(snapshot.Messages) != 2 {
		t.Fatal(snapshot, err)
	}
	// Covered-history edits invalidate the whole boundary, without deleting it.
	if _, err = s.DB().Exec(`update messages set content='edited requirement' where id='m0'`); err != nil {
		t.Fatal(err)
	}
	snapshot, err = s.ContextSnapshot(ctx, "A")
	if err != nil || snapshot.Summary != "" || len(snapshot.Messages) != 7 || snapshot.Version != 2 {
		t.Fatal(snapshot, err)
	}
	if _, err = s.DB().Exec(`delete from messages where id='m1'`); err != nil {
		t.Fatal(err)
	}
	snapshot, err = s.ContextSnapshot(ctx, "A")
	if err != nil || snapshot.Summary != "" || len(snapshot.Messages) != 6 {
		t.Fatal(snapshot, err)
	}
}

func TestContextCheckpointAtomicFailureAndSnapshotConflict(t *testing.T) {
	for _, mode := range []string{"cancel", "write", "history", "version"} {
		t.Run(mode, func(t *testing.T) {
			s := compactionStore(t)
			ctx := context.Background()
			s.AddMessage(ctx, "m1", "A", "user", "original", nil)
			s.AddMessage(ctx, "m2", "A", "assistant", "answer", nil)
			snapshot, err := s.ContextSnapshot(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			boundary, err := PrepareContextBoundary(snapshot, 1)
			if err != nil {
				t.Fatal(err)
			}
			seq, err := s.BeginCompaction(ctx, "A", "run", nil)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "cancel":
				s.UpdateTurnStatusAndPhase(ctx, "run", "cancelling", "cancelling")
			case "write":
				_, err = s.DB().Exec(`create trigger reject_summary before insert on messages begin select raise(abort,'write failure'); end`)
			case "history":
				err = s.AddMessage(ctx, "raced", "A", "user", "concurrent edit", nil)
			case "version":
				_, err = s.DB().Exec(`insert into context_checkpoints values('A',9,'other','[]','now')`)
			}
			if err != nil {
				t.Fatal(err)
			}
			out, err := s.FinishCompactionWithBoundary(ctx, "A", "run", seq, "completed", "new summary", nil, boundary)
			if mode == "cancel" {
				if err != nil || out != "cancelled" {
					t.Fatal(out, err)
				}
			} else if err == nil {
				t.Fatal("expected rejection")
			}
			if (mode == "history" || mode == "version") && !errors.Is(err, ErrContextChanged) {
				t.Fatal(err)
			}
			var n int
			s.DB().QueryRow(`select count(*) from context_checkpoints where summary='new summary'`).Scan(&n)
			if n != 0 {
				t.Fatal("partial boundary committed")
			}
			events, _ := s.ListTurnEvents(ctx, "run")
			for _, ev := range events {
				if ev.Type == "compaction.completed" {
					t.Fatal("partial completion")
				}
			}
			if mode == "version" {
				var summary string
				s.DB().QueryRow(`select summary from context_checkpoints where session_id='A'`).Scan(&summary)
				if summary != "other" {
					t.Fatal(summary)
				}
			}
		})
	}
}

func TestContextCheckpointRejectsNonPrefixCoverage(t *testing.T) {
	s := compactionStore(t)
	ctx := context.Background()
	s.AddMessage(ctx, "first", "A", "user", "first", nil)
	s.AddMessage(ctx, "middle", "A", "user", "middle", nil)
	s.AddMessage(ctx, "last", "A", "user", "last", nil)
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := PrepareContextBoundary(snapshot, 1)
	if err != nil {
		t.Fatal(err)
	}
	boundary.Covered[0] = snapshot.Expected[1]
	seq, err := s.BeginCompaction(ctx, "A", "run", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.FinishCompactionWithBoundary(ctx, "A", "run", seq, "completed", "invalid", nil, boundary); !errors.Is(err, ErrContextChanged) {
		t.Fatal(err)
	}
}
