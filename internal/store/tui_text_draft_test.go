package store

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

func textDraftStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "text-draft.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	for _, id := range []string{"A", "B"} {
		if _, err = s.CreateSession(context.Background(), id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	return s, path
}
func saveTextDraft(t *testing.T, s *Store, text string) TUITextDraft {
	t.Helper()
	state, err := s.SaveTUITextDraft(context.Background(), "A", 0, TUITextSnapshot{Text: text, Cursor: utf8.RuneCountInString(text)})
	if err != nil {
		t.Fatal(err)
	}
	return state
}
func claimTextDraft(t *testing.T, s *Store, state TUITextDraft) TUITextDraft {
	t.Helper()
	state, err := s.ClaimTUITextDraft(context.Background(), "A", state.Revision)
	if err != nil {
		t.Fatal(err)
	}
	return state
}
func textDraftRaw(t *testing.T, s *Store) string {
	t.Helper()
	var raw string
	err := s.DB().QueryRow(`select value from kv_store where namespace=? and key='A'`, tuiTextDraftNamespace).Scan(&raw)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestTUITextDraftExactUnicodeCASAndRestart(t *testing.T) {
	s, path := textDraftStore(t)
	ctx := context.Background()
	text := "  中文e\u0301🙂\n" + strings.Repeat("long🙂", 1<<17) + " trailing  "
	state, err := s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: text, Cursor: 4})
	if err != nil || state.Text != text || state.Cursor != 4 || state.Revision != 1 {
		t.Fatal("full draft not saved", err)
	}
	before := textDraftRaw(t, s)
	loaded, err := s.LoadTUITextDraft(ctx, "A")
	if err != nil || !reflect.DeepEqual(state, loaded) || textDraftRaw(t, s) != before {
		t.Fatal("read changed state", err)
	}
	same, err := s.SaveTUITextDraft(ctx, "A", state.Revision, state.TUITextSnapshot)
	if err != nil || same.Revision != state.Revision {
		t.Fatal("noop changed revision", err)
	}
	if _, err = s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: "stale", Cursor: 2}); !errors.Is(err, ErrTUIDraftConflict) {
		t.Fatal("stale writer accepted", err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	loaded, err = s.LoadTUITextDraft(ctx, "A")
	if err != nil || !reflect.DeepEqual(loaded, state) {
		t.Fatal("reopen changed bytes/cursor", err)
	}
	other, err := s.LoadTUITextDraft(ctx, "B")
	if err != nil || other.Text != "" || other.Revision != 0 {
		t.Fatal("session bleed", other, err)
	}
	cleared, err := s.SaveTUITextDraft(ctx, "A", state.Revision, TUITextSnapshot{})
	if err != nil || cleared.Revision != state.Revision+1 {
		t.Fatal(cleared, err)
	}
	if _, err = s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: "ABA", Cursor: 3}); !errors.Is(err, ErrTUIDraftConflict) {
		t.Fatal("clear deleted revision fence", err)
	}
}

func TestTUITextDraftTwoWritersHaveOneWinner(t *testing.T) {
	s, path := textDraftStore(t)
	other, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	ctx := context.Background()
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, db := range []*Store{s, other} {
		wg.Add(1)
		go func(db *Store) {
			defer wg.Done()
			<-start
			_, err := db.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{Text: "draft", Cursor: 2})
			results <- err
		}(db)
	}
	close(start)
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, ErrTUIDraftConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatal(wins, conflicts)
	}
	state, err := s.LoadTUITextDraft(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	start = make(chan struct{})
	results = make(chan error, 2)
	for _, db := range []*Store{s, other} {
		wg.Add(1)
		go func(db *Store) {
			defer wg.Done()
			<-start
			_, err := db.ClaimTUITextDraft(ctx, "A", state.Revision)
			results <- err
		}(db)
	}
	close(start)
	wg.Wait()
	close(results)
	wins, conflicts = 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, ErrTUIDraftConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatal("claim race", wins, conflicts)
	}
}

func TestTUITextDraftUnknownAndProvisionalAdmissionStayHeldOnRestart(t *testing.T) {
	s, path := textDraftStore(t)
	ctx := context.Background()
	original := saveTextDraft(t, s, "do not resend 中文")
	state := claimTextDraft(t, s, original)
	token := state.Claim.Token
	if len(token) != len("tui-text-")+64 {
		t.Fatal("not random256 token", token)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	loaded, err := s.LoadTUITextDraft(ctx, "A")
	if err != nil || loaded.Text != "" || loaded.Claim == nil || loaded.Claim.Text != original.Text {
		t.Fatal(loaded, err)
	}
	if _, err = s.RestoreRejectedTUITextDraft(ctx, "A", token, state.Revision); !errors.Is(err, ErrTUIDraftHeld) {
		t.Fatal("unknown restored", err)
	}
	if _, err = s.DiscardRejectedTUITextDraft(ctx, "A", token, state.Revision); !errors.Is(err, ErrTUIDraftHeld) {
		t.Fatal("unknown discarded", err)
	}
	if _, err = s.ClaimTUITextDraft(ctx, "A", state.Revision); !errors.Is(err, ErrTUIDraftHeld) {
		t.Fatal("unknown re-claimed", err)
	}
	if _, err = s.ReconcileTUITextDraft(ctx, "A", "wrong"); !errors.Is(err, ErrTUIDraftConflict) {
		t.Fatal("wrong token", err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "foreign", "B", "queued", original.Text, map[string]any{"tui_text_claim": token}); err != nil {
		t.Fatal(err)
	}
	if err = s.AppendTurnEvent(ctx, "foreign", "B", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	checked, err := s.ReconcileTUITextDraft(ctx, "A", token)
	if err != nil || checked.Claim == nil {
		t.Fatal("foreign admission settled", checked, err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "local", "A", "queued", original.Text, map[string]any{"tui_text_claim": token}); err != nil {
		t.Fatal(err)
	}
	checked, err = s.ReconcileTUITextDraft(ctx, "A", token)
	if err != nil || checked.Claim == nil {
		t.Fatal("provisional INSERT settled", checked, err)
	}
	if err = s.AppendTurnEvent(ctx, "local", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	checked, err = s.ReconcileTUITextDraft(ctx, "A", token)
	if err != nil || checked.Claim != nil || checked.Text != "" {
		t.Fatal("confirmed admission not retired", checked, err)
	}
}

func TestTUITextDraftRejectionPreservesNewEditsAndExplicitRecovery(t *testing.T) {
	for _, action := range []string{"restore", "discard", "untouched"} {
		t.Run(action, func(t *testing.T) {
			s, _ := textDraftStore(t)
			ctx := context.Background()
			initial := saveTextDraft(t, s, "original 🙂")
			state := claimTextDraft(t, s, initial)
			token := state.Claim.Token
			if action != "untouched" {
				var err error
				state, err = s.SaveTUITextDraft(ctx, "A", state.Revision, TUITextSnapshot{Text: "newer", Cursor: 2})
				if err != nil {
					t.Fatal(err)
				}
			}
			var err error
			state, err = s.FinishTUITextDraft(ctx, "A", token, true)
			if err != nil {
				t.Fatal(err)
			}
			if action == "untouched" {
				if state.Text != initial.Text || state.Cursor != initial.Cursor || state.Claim != nil {
					t.Fatal(state)
				}
				return
			}
			if state.Text != "newer" || state.Claim == nil || !state.Claim.Rejected {
				t.Fatal("newer edit overwritten", state)
			}
			stale := state.Revision - 1
			if _, err = s.DiscardRejectedTUITextDraft(ctx, "A", token, stale); !errors.Is(err, ErrTUIDraftConflict) {
				t.Fatal("stale discard accepted", err)
			}
			if action == "restore" {
				if _, err = s.RestoreRejectedTUITextDraft(ctx, "A", token, state.Revision); !errors.Is(err, ErrTUIDraftConflict) {
					t.Fatal("restore overwrote new draft", err)
				}
				state, err = s.SaveTUITextDraft(ctx, "A", state.Revision, TUITextSnapshot{})
				if err != nil {
					t.Fatal(err)
				}
				state, err = s.RestoreRejectedTUITextDraft(ctx, "A", token, state.Revision)
				if err != nil || state.Text != initial.Text || state.Claim != nil {
					t.Fatal(state, err)
				}
			} else {
				state, err = s.DiscardRejectedTUITextDraft(ctx, "A", token, state.Revision)
				if err != nil || state.Text != "newer" || state.Claim != nil {
					t.Fatal(state, err)
				}
			}
			next := claimTextDraft(t, s, state)
			if next.Claim.Token == token {
				t.Fatal("token reused")
			}
		})
	}
}

func TestTUITextDraftConfirmedAdmissionKeepsNewerDraft(t *testing.T) {
	for _, receipt := range []string{"turn", "steering", "message", "absent-success"} {
		t.Run(receipt, func(t *testing.T) {
			s, _ := textDraftStore(t)
			ctx := context.Background()
			state := claimTextDraft(t, s, saveTextDraft(t, s, "original"))
			token := state.Claim.Token
			state, err := s.SaveTUITextDraft(ctx, "A", state.Revision, TUITextSnapshot{Text: "next draft", Cursor: 4})
			if err != nil {
				t.Fatal(err)
			}
			metadata := map[string]any{"tui_text_claim": token}
			switch receipt {
			case "turn":
				_, err = s.CreateTurnWithStatus(ctx, "admitted", "A", "queued", "original", metadata)
			case "steering":
				_, err = s.EnqueueSteering(ctx, "A", "", "user", "original", metadata, nil, "")
			case "message":
				err = s.AddMessage(ctx, "msg", "A", "user", "original", metadata)
			}
			if err != nil {
				t.Fatal(err)
			}
			// Returned submit may confirm even a turn whose warning-only audit was lost.
			state, err = s.FinishTUITextDraft(ctx, "A", token, false)
			if err != nil || state.Text != "next draft" || state.Cursor != 4 {
				t.Fatal(state, err)
			}
			if receipt == "absent-success" {
				if state.Claim == nil {
					t.Fatal("success without receipt dropped claim")
				}
			} else if state.Claim != nil {
				t.Fatal("admission claim retained")
			}
		})
	}
}

func TestTUITextDraftInvalidStateAndWriteFailuresFailClosed(t *testing.T) {
	s, _ := textDraftStore(t)
	ctx := context.Background()
	for _, snapshot := range []TUITextSnapshot{{Text: string([]byte{0xff}), Cursor: 0}, {Text: "🙂", Cursor: 2}, {Text: "text", Cursor: -1}} {
		if _, err := s.SaveTUITextDraft(ctx, "A", 0, snapshot); err == nil {
			t.Fatal("invalid snapshot saved", snapshot)
		}
	}
	if _, err := s.SaveTUITextDraft(ctx, "missing", 0, TUITextSnapshot{Text: "x", Cursor: 1}); err == nil {
		t.Fatal("missing session saved")
	}
	state := saveTextDraft(t, s, "retained")
	before := textDraftRaw(t, s)
	if _, err := s.DB().Exec(`create trigger deny_text_draft before update on kv_store begin select raise(abort,'denied'); end`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimTUITextDraft(ctx, "A", state.Revision); err == nil {
		t.Fatal("claim write succeeded")
	}
	if textDraftRaw(t, s) != before {
		t.Fatal("failed claim changed journal")
	}
	if _, err := s.DB().Exec(`drop trigger deny_text_draft`); err != nil {
		t.Fatal(err)
	}
	state = claimTextDraft(t, s, state)
	before = textDraftRaw(t, s)
	if _, err := s.DB().Exec(`alter table turns rename to hidden_turns`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FinishTUITextDraft(ctx, "A", state.Claim.Token, true); err == nil {
		t.Fatal("read error restored claim")
	}
	if textDraftRaw(t, s) != before {
		t.Fatal("failed receipt read changed journal")
	}
	if _, err := s.DB().Exec(`alter table hidden_turns rename to turns`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`update kv_store set value='not-json' where namespace=?`, tuiTextDraftNamespace); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadTUITextDraft(ctx, "A"); err == nil {
		t.Fatal("corrupt journal loaded")
	}
	if _, err := s.SaveTUITextDraft(ctx, "A", 0, TUITextSnapshot{}); err == nil {
		t.Fatal("corrupt journal overwritten")
	}
}

func TestTUITextDraftRevisionExhaustionRejectsBeforeClaim(t *testing.T) {
	s, _ := textDraftStore(t)
	ctx := context.Background()
	saveTextDraft(t, s, "retained")
	if _, err := s.DB().Exec(`update kv_store set value=json_set(value,'$.revision',?) where namespace=?`, math.MaxInt64-1, tuiTextDraftNamespace); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ClaimTUITextDraft(ctx, "A", math.MaxInt64-1); err == nil {
		t.Fatal("unresolvable claim accepted")
	}
	state, err := s.LoadTUITextDraft(ctx, "A")
	if err != nil || state.Text != "retained" || state.Claim != nil {
		t.Fatal(state, err)
	}
}

func TestTUITextDraftRejectedWithUnauditedTurnStaysHeld(t *testing.T) {
	s, _ := textDraftStore(t)
	ctx := context.Background()
	state := claimTextDraft(t, s, saveTextDraft(t, s, "original"))
	token := state.Claim.Token
	if _, err := s.CreateTurnWithStatus(ctx, "partial", "A", "queued", "original", map[string]any{"tui_text_claim": token}); err != nil {
		t.Fatal(err)
	}
	held, err := s.FinishTUITextDraft(ctx, "A", token, true)
	if err != nil || held.Claim == nil || held.Claim.Rejected || held.Text != "" {
		t.Fatal("ambiguous error lost/restored claim", held, err)
	}
	if err = s.AppendTurnEvent(ctx, "partial", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	settled, err := s.FinishTUITextDraft(ctx, "A", token, true)
	if err != nil || settled.Claim != nil || settled.Text != "" {
		t.Fatal("confirmed admission not retired", settled, err)
	}
}
