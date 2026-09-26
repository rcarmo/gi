package store

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestTUIMediaDraftRestartClaimAndAdmissionRecovery(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "a", "test", nil); err != nil {
		t.Fatal(err)
	}
	media, err := s.CreateMedia(ctx, "a", "pending.txt", "text/plain", []byte("durable bytes"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ref := MediaRef{ID: MediaRefID(media.ID), MediaID: media.ID, SessionID: "a", Filename: media.Filename, Size: media.OriginalSize}
	if _, err = s.StageTUIMedia(ctx, "a", ref); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	state, err := s.LoadTUIMediaDraft(ctx, "a")
	if err != nil || len(state.Pending) != 1 {
		t.Fatal(state, err)
	}
	if _, err = s.ClaimTUIMedia(ctx, "a", "claim", false); err == nil {
		t.Fatal("directed claim consumed refs")
	}
	if _, err = s.ClaimTUIMedia(ctx, "a", "claim", true); err != nil {
		t.Fatal(err)
	}
	state, err = s.LoadTUIMediaDraft(ctx, "a")
	if err != nil || state.Claim == nil || len(state.Pending) != 0 {
		t.Fatal("unconfirmed restored", state, err)
	}
	state, restored, err := s.SettleTUIMedia(ctx, "a", "claim", true, false)
	if err != nil || restored || state.Claim == nil {
		t.Fatal("restart absence unsafe", state, err)
	}
	if _, _, err = s.DetachTUIMedia(ctx, "a", "unresolved", "old-token"); err == nil {
		t.Fatal("stale discard")
	}
	if _, err = s.CreateTurn(ctx, "accepted", "a", "prompt", map[string]any{"tui_media_claim": "claim"}); err != nil {
		t.Fatal(err)
	}
	state, err = s.LoadTUIMediaDraft(ctx, "a")
	if err != nil || state.Claim != nil || len(state.Pending) != 0 {
		t.Fatal("admitted retained", state, err)
	}
	if _, err = s.StageTUIMedia(ctx, "a", ref); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ClaimTUIMedia(ctx, "a", "uncertain", true); err != nil {
		t.Fatal(err)
	}
	if _, n, err := s.DetachTUIMedia(ctx, "a", "unresolved", "uncertain"); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	if _, err = s.GetMedia(ctx, media.ID); err != nil {
		t.Fatal("discard deleted media", err)
	}
}

func TestTUIMediaDraftConcurrentClientsAndFailureRollback(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	other, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	s.CreateSession(ctx, "a", "test", nil)
	s.CreateSession(ctx, "b", "test", nil)
	media, err := s.CreateMedia(ctx, "a", "one", "text/plain", []byte("bytes"), nil)
	if err != nil {
		t.Fatal(err)
	}
	ref := MediaRef{ID: MediaRefID(media.ID), MediaID: media.ID, SessionID: "a"}
	if _, err = s.StageTUIMedia(ctx, "b", ref); err == nil {
		t.Fatal("cross session")
	}
	var ok atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db := s
			if i%2 == 1 {
				db = other
			}
			if _, err := db.StageTUIMedia(ctx, "a", ref); err == nil {
				ok.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if ok.Load() != 6 {
		t.Fatal("lost updates/overflow", ok.Load())
	}
	state, err := s.LoadTUIMediaDraft(ctx, "a")
	if err != nil || len(state.Pending) != 6 {
		t.Fatal(state, err)
	}
	if _, err = s.DB().Exec(`CREATE TRIGGER deny_journal BEFORE UPDATE ON kv_store BEGIN SELECT RAISE(ABORT,'denied'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ClaimTUIMedia(ctx, "a", "failed", true); err == nil {
		t.Fatal("claim should fail")
	}
	s.DB().Exec(`DROP TRIGGER deny_journal`)
	state, err = s.LoadTUIMediaDraft(ctx, "a")
	if err != nil || len(state.Pending) != 6 || state.Claim != nil {
		t.Fatal("failed claim lost refs", state, err)
	}
	if _, err = s.ClaimTUIMedia(ctx, "a", "live", true); err != nil {
		t.Fatal(err)
	}
	if _, err = other.ClaimTUIMedia(ctx, "a", "racing", true); err == nil {
		t.Fatal("double claim")
	}
	if _, _, err = other.SettleTUIMedia(ctx, "a", "stale", false, false); err != nil {
		t.Fatal(err)
	}
	state, err = s.LoadTUIMediaDraft(ctx, "a")
	if err != nil || state.Claim.Token != "live" {
		t.Fatal("stale settlement")
	}
	state, restored, err := s.SettleTUIMedia(ctx, "a", "live", true, true)
	if err != nil || !restored || len(state.Pending) != 6 {
		t.Fatal("live rejection", state, err)
	}
}
