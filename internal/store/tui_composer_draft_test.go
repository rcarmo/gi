package store

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func stageMediaFixture(t *testing.T, s *Store, session string) MediaRef {
	t.Helper()
	media, err := s.CreateMedia(context.Background(), session, "pending.txt", "text/plain", []byte("durable bytes"), nil)
	if err != nil {
		t.Fatal(err)
	}
	return MediaRef{ID: MediaRefID(media.ID), MediaID: media.ID, SessionID: session, Filename: media.Filename, Size: media.OriginalSize}
}

func composerDraftFixture(t *testing.T) (*Store, string, TUITextDraft, MediaRef) {
	t.Helper()
	s, path := textDraftStore(t)
	state := saveTextDraft(t, s, "  paired 中文🙂\ntext  ")
	ref := stageMediaFixture(t, s, "A")
	if _, err := s.StageTUIMedia(context.Background(), "A", ref); err != nil {
		t.Fatal(err)
	}
	return s, path, state, ref
}
func claimComposer(t *testing.T, s *Store, state TUITextDraft) TUIComposerDraft {
	t.Helper()
	out, err := s.ClaimTUIComposerDraft(context.Background(), "A", state.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if out.Text.Claim == nil || !out.Text.Claim.Media || out.Media.Claim == nil || !out.Media.Claim.Text || out.Media.Claim.Token != out.Text.Claim.Token {
		t.Fatal("unpaired claim", out)
	}
	if out.Text.Text != "" || len(out.Media.Pending) != 0 {
		t.Fatal("editable state not cleared", out)
	}
	return out
}

func TestTUIComposerDraftClaimRollbackAndReopen(t *testing.T) {
	for _, fail := range []string{"media", "text", "none"} {
		t.Run(fail, func(t *testing.T) {
			s, path, state, ref := composerDraftFixture(t)
			ctx := context.Background()
			before, err := s.LoadTUIComposerDraft(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if fail != "none" {
				namespace := tuiTextDraftNamespace
				if fail == "media" {
					namespace = tuiMediaNamespace
				}
				if _, err = s.DB().Exec(`create trigger deny_pair before update on kv_store when OLD.namespace='` + namespace + `' begin select raise(abort,'pair failed'); end`); err != nil {
					t.Fatal(err)
				}
			}
			got, err := s.ClaimTUIComposerDraft(ctx, "A", state.Revision)
			if fail != "none" {
				if err == nil {
					t.Fatal("expected claim write failure")
				}
			} else if err != nil || got.Text.Claim == nil {
				t.Fatal(got, err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			loaded, err := s.LoadTUIComposerDraft(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if fail != "none" {
				if !reflect.DeepEqual(before, loaded) {
					t.Fatal("failed claim changed pair", before, loaded)
				}
				return
			}
			if !reflect.DeepEqual(got, loaded) || loaded.Media.Claim.Refs[0].ID != ref.ID {
				t.Fatal("reopen changed pair", got, loaded)
			}
			held, err := s.ReconcileTUIComposerDraft(ctx, "A", loaded.Text.Claim.Token)
			if err != nil || !reflect.DeepEqual(held, loaded) {
				t.Fatal("restart absence changed claim", held, err)
			}
		})
	}
}

func TestTUIComposerDraftIndependentAPIsCannotUnpair(t *testing.T) {
	s, _, state, _ := composerDraftFixture(t)
	ctx := context.Background()
	pair := claimComposer(t, s, state)
	token := pair.Text.Claim.Token
	for name, err := range map[string]error{
		"text finish":    func() error { _, e := s.FinishTUITextDraft(ctx, "A", token, true); return e }(),
		"text reconcile": func() error { _, e := s.ReconcileTUITextDraft(ctx, "A", token); return e }(),
		"text restore":   func() error { _, e := s.RestoreRejectedTUITextDraft(ctx, "A", token, pair.Text.Revision); return e }(),
		"text discard":   func() error { _, e := s.DiscardRejectedTUITextDraft(ctx, "A", token, pair.Text.Revision); return e }(),
		"media settle":   func() error { _, _, e := s.SettleTUIMedia(ctx, "A", token, false, false); return e }(),
		"media detach":   func() error { _, _, e := s.DetachTUIMedia(ctx, "A", "unresolved", token); return e }(),
	} {
		if !errors.Is(err, ErrTUIDraftHeld) {
			t.Fatal(name, err)
		}
	}
	if _, err := s.CreateTurnWithStatus(ctx, "receipt", "A", "queued", "original", map[string]any{"tui_text_claim": token, "tui_media_claim": token}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendTurnEvent(ctx, "receipt", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	media, err := s.LoadTUIMediaDraft(ctx, "A")
	if err != nil || media.Claim == nil {
		t.Fatal("legacy load consumed half", media, err)
	}
	loaded, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil || !reflect.DeepEqual(pair, loaded) {
		t.Fatal("single API mutated pair", loaded, err)
	}
	settled, err := s.ReconcileTUIComposerDraft(ctx, "A", token)
	if err != nil || settled.Text.Claim != nil || settled.Media.Claim != nil {
		t.Fatal(settled, err)
	}
}

func TestTUIComposerDraftReceiptMatrix(t *testing.T) {
	for _, receipt := range []string{"absent", "foreign", "text-only", "media-only", "split", "provisional", "confirmed", "steering", "message"} {
		t.Run(receipt, func(t *testing.T) {
			s, _, state, _ := composerDraftFixture(t)
			ctx := context.Background()
			pair := claimComposer(t, s, state)
			token := pair.Text.Claim.Token
			both := map[string]any{"tui_text_claim": token, "tui_media_claim": token}
			create := func(id, session string, meta map[string]any, audit bool) {
				t.Helper()
				if _, err := s.CreateTurnWithStatus(ctx, id, session, "queued", "original", meta); err != nil {
					t.Fatal(err)
				}
				if audit {
					if err := s.AppendTurnEvent(ctx, id, session, "turn.submitted", nil); err != nil {
						t.Fatal(err)
					}
				}
			}
			switch receipt {
			case "foreign":
				create("foreign", "B", both, true)
			case "text-only":
				create("text", "A", map[string]any{"tui_text_claim": token}, true)
			case "media-only":
				create("media", "A", map[string]any{"tui_media_claim": token}, true)
			case "split":
				create("text", "A", map[string]any{"tui_text_claim": token}, true)
				create("media", "A", map[string]any{"tui_media_claim": token}, true)
			case "provisional":
				create("provisional", "A", both, false)
			case "confirmed":
				create("confirmed", "A", both, true)
			case "steering":
				if _, err := s.EnqueueSteering(ctx, "A", "", "user", "original", both, nil, ""); err != nil {
					t.Fatal(err)
				}
			case "message":
				if err := s.AddMessage(ctx, "msg", "A", "user", "original", both); err != nil {
					t.Fatal(err)
				}
			}
			got, err := s.ReconcileTUIComposerDraft(ctx, "A", token)
			if err != nil {
				t.Fatal(err)
			}
			confirmed := receipt == "confirmed" || receipt == "steering" || receipt == "message"
			if confirmed {
				if got.Text.Claim != nil || got.Media.Claim != nil {
					t.Fatal("confirmed retained", got)
				}
				return
			}
			if !reflect.DeepEqual(pair, got) {
				t.Fatal("unconfirmed changed", got)
			}
			got, err = s.FinishTUIComposerDraft(ctx, "A", token, true)
			if err != nil {
				t.Fatal(err)
			}
			if receipt == "absent" || receipt == "foreign" {
				if got.Text.Text != state.Text || got.Text.Cursor != state.Cursor || got.Text.Claim != nil || got.Media.Claim != nil || len(got.Media.Pending) != 1 {
					t.Fatal("known rejection not restored", got)
				}
			} else if !reflect.DeepEqual(pair, got) {
				t.Fatal("ambiguous error restored/dropped pair", got)
			}
		})
	}
}

func TestTUIComposerDraftSettlementRollbackAndNewerEdits(t *testing.T) {
	s, _, state, ref := composerDraftFixture(t)
	ctx := context.Background()
	pair := claimComposer(t, s, state)
	token := pair.Text.Claim.Token
	if _, err := s.SaveTUITextDraft(ctx, "A", pair.Text.Revision, TUITextSnapshot{Text: "new draft", Cursor: 3}); err != nil {
		t.Fatal(err)
	}
	newer := stageMediaFixture(t, s, "A")
	if _, err := s.StageTUIMedia(ctx, "A", newer); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "receipt", "A", "queued", "original", map[string]any{"tui_text_claim": token, "tui_media_claim": token}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendTurnEvent(ctx, "receipt", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`create trigger deny_pair_settlement before update on kv_store when OLD.namespace='` + tuiTextDraftNamespace + `' begin select raise(abort,'settle failed'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReconcileTUIComposerDraft(ctx, "A", token); err == nil {
		t.Fatal("settlement fault ignored")
	}
	after, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("half settlement committed", after, err)
	}
	if _, err = s.DB().Exec(`drop trigger deny_pair_settlement`); err != nil {
		t.Fatal(err)
	}
	after, err = s.ReconcileTUIComposerDraft(ctx, "A", token)
	if err != nil || after.Text.Text != "new draft" || after.Text.Cursor != 3 || after.Text.Claim != nil || after.Media.Claim != nil || len(after.Media.Pending) != 1 || after.Media.Pending[0].ID != newer.ID {
		t.Fatal("newer text/media lost", after, err)
	}
	if _, err = s.GetMedia(ctx, ref.MediaID); err != nil {
		t.Fatal("stored media deleted", err)
	}
}

func TestTUIComposerDraftRejectedRestoreDiscardAndStaleGuards(t *testing.T) {
	for _, restore := range []bool{false, true} {
		t.Run(map[bool]string{false: "discard", true: "restore"}[restore], func(t *testing.T) {
			s, _, state, ref := composerDraftFixture(t)
			ctx := context.Background()
			pair := claimComposer(t, s, state)
			token := pair.Text.Claim.Token
			changed, err := s.SaveTUITextDraft(ctx, "A", pair.Text.Revision, TUITextSnapshot{Text: "newer", Cursor: 2})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.FinishTUIComposerDraft(ctx, "B", token, true); !errors.Is(err, ErrTUIDraftConflict) {
				t.Fatal("foreign settle", err)
			}
			if _, err = s.ResolveRejectedTUIComposerDraft(ctx, "A", token, changed.Revision, false); !errors.Is(err, ErrTUIDraftHeld) {
				t.Fatal("unknown discard", err)
			}
			rejected, err := s.FinishTUIComposerDraft(ctx, "A", token, true)
			if err != nil || rejected.Text.Claim == nil || !rejected.Text.Claim.Rejected || rejected.Text.Text != "newer" || rejected.Media.Claim == nil {
				t.Fatal(rejected, err)
			}
			if _, err = s.ResolveRejectedTUIComposerDraft(ctx, "A", token, changed.Revision, restore); !errors.Is(err, ErrTUIDraftConflict) {
				t.Fatal("stale restore/discard", err)
			}
			if restore {
				if _, err = s.ResolveRejectedTUIComposerDraft(ctx, "A", token, rejected.Text.Revision, true); !errors.Is(err, ErrTUIDraftConflict) {
					t.Fatal("overwrote newer", err)
				}
				changed, err = s.SaveTUITextDraft(ctx, "A", rejected.Text.Revision, TUITextSnapshot{})
				if err != nil {
					t.Fatal(err)
				}
				rejected.Text = changed
			}
			result, err := s.ResolveRejectedTUIComposerDraft(ctx, "A", token, rejected.Text.Revision, restore)
			if err != nil || result.Text.Claim != nil || result.Media.Claim != nil {
				t.Fatal(result, err)
			}
			if restore {
				if result.Text.Text != state.Text || len(result.Media.Pending) != 1 || result.Media.Pending[0].ID != ref.ID {
					t.Fatal(result)
				}
			} else if result.Text.Text != "newer" || len(result.Media.Pending) != 0 {
				t.Fatal(result)
			}
			if _, err = s.GetMedia(ctx, ref.MediaID); err != nil {
				t.Fatal("deleted stored media", err)
			}
			if _, err = s.ClaimTUIComposerDraft(ctx, "A", result.Text.Revision); err != nil {
				t.Fatal("new draft blocked after resolution", err)
			}
		})
	}
}

func TestTUIComposerDraftConcurrentClaimAndMediaCapacity(t *testing.T) {
	s, path, state, _ := composerDraftFixture(t)
	ctx := context.Background()
	other, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	start := make(chan struct{})
	out := make(chan error, 2)
	var wg sync.WaitGroup
	for _, db := range []*Store{s, other} {
		wg.Add(1)
		go func(db *Store) {
			defer wg.Done()
			<-start
			_, err := db.ClaimTUIComposerDraft(ctx, "A", state.Revision)
			out <- err
		}(db)
	}
	close(start)
	wg.Wait()
	close(out)
	wins, conflicts := 0, 0
	for err := range out {
		if err == nil {
			wins++
		} else if errors.Is(err, ErrTUIDraftHeld) || errors.Is(err, ErrTUIDraftConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatal(wins, conflicts)
	}
	for i := 0; i < 5; i++ {
		ref := stageMediaFixture(t, s, "A")
		if _, err = s.StageTUIMedia(ctx, "A", ref); err != nil {
			t.Fatal(err)
		}
	}
	extra := stageMediaFixture(t, s, "A")
	if _, err = s.StageTUIMedia(ctx, "A", extra); err == nil {
		t.Fatal("paired refs bypassed six-slot cap")
	}
	pair, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	restored, err := s.FinishTUIComposerDraft(ctx, "A", pair.Text.Claim.Token, true)
	if err != nil || len(restored.Media.Pending) != 6 || restored.Media.Claim != nil {
		t.Fatal(restored, err)
	}
}

func TestTUIComposerDraftRefValidationAndInconsistentPairFailClosed(t *testing.T) {
	s, _, state, _ := composerDraftFixture(t)
	ctx := context.Background()
	foreign := stageMediaFixture(t, s, "B")
	if _, err := s.DB().Exec(`update kv_store set value=json_set(value,'$.pending[0].media_id',?) where namespace=? and key='A'`, foreign.MediaID, tuiMediaNamespace); err != nil {
		t.Fatal(err)
	}
	before, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ClaimTUIComposerDraft(ctx, "A", state.Revision); err == nil {
		t.Fatal("foreign media claimed")
	}
	after, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("failed validation changed journals", err)
	}
	// Fresh fixture verifies inconsistent paired tokens never clear either half.
	s2, _ := textDraftStore(t)
	text := saveTextDraft(t, s2, "text")
	ref := stageMediaFixture(t, s2, "A")
	if _, err = s2.StageTUIMedia(ctx, "A", ref); err != nil {
		t.Fatal(err)
	}
	pair := claimComposer(t, s2, text)
	if _, err = s2.DB().Exec(`update kv_store set value=json_set(value,'$.claim.token','foreign-token') where namespace=? and key='A'`, tuiMediaNamespace); err != nil {
		t.Fatal(err)
	}
	if _, err = s2.FinishTUIComposerDraft(ctx, "A", pair.Text.Claim.Token, true); !errors.Is(err, ErrTUIDraftHeld) {
		t.Fatal("inconsistent pair settled", err)
	}
	if _, err = s2.LoadTUIComposerDraft(ctx, "missing"); err == nil {
		t.Fatal("missing session accepted")
	}
}

func TestTUIComposerDraftTextOnlyDoesNotCreateMediaRow(t *testing.T) {
	s, _ := textDraftStore(t)
	ctx := context.Background()
	text := saveTextDraft(t, s, "text-only")
	pair, err := s.ClaimTUIComposerDraft(ctx, "A", text.Revision)
	if err != nil || pair.Text.Claim == nil || pair.Text.Claim.Media || pair.Media.Claim != nil {
		t.Fatal(pair, err)
	}
	pair, err = s.FinishTUIComposerDraft(ctx, "A", pair.Text.Claim.Token, true)
	if err != nil || pair.Text.Text != text.Text || pair.Text.Claim != nil {
		t.Fatal(pair, err)
	}
	var n int
	if err = s.DB().QueryRow(`select count(*) from kv_store where namespace=?`, tuiMediaNamespace).Scan(&n); err != nil || n != 0 {
		t.Fatal("spurious media row", n, err)
	}
}

func TestTUIComposerDraftDispatchPreservesNewerEditsAndFencesRestart(t *testing.T) {
	s, path, original, _ := composerDraftFixture(t)
	ctx := context.Background()
	pair := claimComposer(t, s, original)
	token := pair.Text.Claim.Token
	newer, err := s.SaveTUITextDraft(ctx, "A", pair.Text.Revision, TUITextSnapshot{Text: "new edit", Cursor: 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.BeginTUIComposerSubmission(ctx, "A", token, pair.Text.Revision); !errors.Is(err, ErrTUIDraftConflict) {
		t.Fatal("stale dispatch accepted", err)
	}
	begun, err := s.BeginTUIComposerSubmission(ctx, "A", token, newer.Revision)
	if err != nil || !begun.Text.Claim.Dispatched || begun.Text.Text != "new edit" {
		t.Fatal(begun, err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.BeginTUIComposerSubmission(ctx, "A", token, begun.Text.Revision); !errors.Is(err, ErrTUIDraftHeld) {
		t.Fatal("restarted dispatch accepted", err)
	}
	rejected, err := s.FinishTUIComposerDraft(ctx, "A", token, true)
	if err != nil || rejected.Text.Text != "new edit" || rejected.Text.Claim == nil || !rejected.Text.Claim.Rejected || rejected.Media.Claim == nil {
		t.Fatal("dispatch reset newer-edit fence", rejected, err)
	}
}
