package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestQueuedResetAtomicAcknowledgement(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "reset-queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`insert into session_active_turns(session_id,turn_id,claim_token,claimed_at,updated_at) values('s','t','claim',` + defaultNow + `,` + defaultNow + `)`); err != nil {
		t.Fatal(err)
	}
	queued, err := s.RequestContextReset(ctx, "s", "fresh")
	if err != nil || !queued {
		t.Fatalf("admission: %t %v", queued, err)
	}
	if _, err := s.EnqueueSteering(ctx, "s", "t", "user", "after", nil, nil, "all"); err != nil {
		t.Fatal(err)
	}
	early, err := s.DequeueSteeringForTurn(ctx, "s", "t")
	if err != nil || len(early) != 0 {
		t.Fatalf("reset bypassed: %+v %v", early, err)
	}
	if _, err := s.db.Exec(`create trigger fail_reset_ack before update on steering_queue begin select raise(abort,'ack failed'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ApplyPendingContextReset(ctx, "s", "t"); err == nil {
		t.Fatal("expected ack failure")
	}
	var markers int
	if err := s.db.QueryRow(`select count(*) from messages where json_extract(payload_json,'$.kind')='context_reset'`).Scan(&markers); err != nil {
		t.Fatal(err)
	}
	if markers != 0 {
		t.Fatal("reset committed without ack")
	}
	if _, err := s.db.Exec(`drop trigger fail_reset_ack`); err != nil {
		t.Fatal(err)
	}
	applied, err := s.ApplyPendingContextReset(ctx, "s", "t")
	if err != nil || !applied {
		t.Fatalf("apply: %t %v", applied, err)
	}
	applied, err = s.ApplyPendingContextReset(ctx, "s", "t")
	if err != nil || applied {
		t.Fatalf("duplicate apply: %t %v", applied, err)
	}
	follow, err := s.DequeueSteeringForTurn(ctx, "s", "t")
	if err != nil || len(follow) != 1 || follow[0].Content != "fresh" {
		t.Fatalf("reset follow-up: %+v %v", follow, err)
	}
}

func TestQueuedResetsSurviveReopenWithoutCrossingFIFO(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reset-recover.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "test", "claim"); err != nil || !ok {
		t.Fatalf("claim %t %v", ok, err)
	}
	for _, text := range []string{"", ""} {
		if queued, err := s.RequestContextReset(ctx, "s", text); err != nil || !queued {
			t.Fatalf("queue %t %v", queued, err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for i := 0; i < 2; i++ {
		applied, err := s.ApplyPendingContextReset(ctx, "s", "t")
		if err != nil || !applied {
			t.Fatalf("apply %d %t %v", i, applied, err)
		}
	}
	applied, err := s.ApplyPendingContextReset(ctx, "s", "t")
	if err != nil || applied {
		t.Fatalf("unexpected third reset: %t %v", applied, err)
	}
	snapshot, err := s.ContextSnapshot(ctx, "s")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Version != 2 || len(snapshot.Messages) != 0 {
		t.Fatalf("reset versions: %+v", snapshot)
	}
}

func TestResetDoesNotCrossEarlierAllModeSteering(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "reset-order.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.CreateSession(ctx, "s", "s", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurn(ctx, "t", "s", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "s", "t", "test", "claim"); err != nil || !ok {
		t.Fatalf("claim %t %v", ok, err)
	}
	if _, err := s.EnqueueSteering(ctx, "s", "t", "user", "before", nil, nil, "all"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RequestContextReset(ctx, "s", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.EnqueueSteering(ctx, "s", "t", "user", "after", nil, nil, "all"); err != nil {
		t.Fatal(err)
	}
	applied, err := s.ApplyPendingContextReset(ctx, "s", "t")
	if err != nil || applied {
		t.Fatalf("reset crossed earlier steering %t %v", applied, err)
	}
	first, err := s.DequeueSteeringForTurn(ctx, "s", "t")
	if err != nil || len(first) != 1 || first[0].Content != "before" {
		t.Fatalf("all crossed reset: %+v %v", first, err)
	}
	applied, err = s.ApplyPendingContextReset(ctx, "s", "t")
	if err != nil || !applied {
		t.Fatalf("reset not applied %t %v", applied, err)
	}
	after, err := s.DequeueSteeringForTurn(ctx, "s", "t")
	if err != nil || len(after) != 1 || after[0].Content != "after" {
		t.Fatalf("post-reset ordering: %+v %v", after, err)
	}
}
