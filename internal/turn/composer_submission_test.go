package turn

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
)

func composerSubmitFixture(t *testing.T, text string, withMedia bool) (*Engine, *store.Store, string, store.TUIComposerDraft) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "composer.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	e := New(s)
	t.Cleanup(func() { e.Close(); s.Close() })
	for _, id := range []string{"A", "B"} {
		if _, _, err = s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: id, Title: id, Allocation: session.AllocateDefaultSession("agent", "gi", "default", id)}); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := s.SaveTUITextDraft(ctx, "A", 0, store.TUITextSnapshot{Text: text, Cursor: 0})
	if err != nil {
		t.Fatal(err)
	}
	if withMedia {
		m, err := s.CreateMedia(ctx, "A", "note.txt", "text/plain", []byte("attached bytes"), nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.StageTUIMedia(ctx, "A", store.MediaRef{ID: store.MediaRefID(m.ID), MediaID: m.ID, SessionID: "A", Filename: m.Filename, Size: m.OriginalSize}); err != nil {
			t.Fatal(err)
		}
	}
	claim, err := s.ClaimTUIComposerDraft(ctx, "A", saved.Revision)
	if err != nil {
		t.Fatal(err)
	}
	return e, s, path, claim
}
func activeComposerFixture(t *testing.T, s *store.Store) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateTurnWithStatus(ctx, "active", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "active", "fixture", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
}

func TestTUIComposerSubmitStoredPromptAndMedia(t *testing.T) {
	for _, media := range []bool{false, true} {
		t.Run(map[bool]string{false: "plain", true: "media"}[media], func(t *testing.T) {
			e, s, _, claim := composerSubmitFixture(t, "  original 中文🙂\nsecond line  ", media)
			ctx := context.Background()
			token := claim.Text.Claim.Token
			result, settled, err := e.SubmitTUIComposer(ctx, "A", token, claim.Text.Revision, "bootstrap")
			if err != nil || result == nil || settled.Text.Claim != nil || settled.Media.Claim != nil {
				t.Fatal(result, settled, err)
			}
			waitRetryDone(t, s, result.TurnID)
			turn, err := s.GetTurn(ctx, result.TurnID)
			if err != nil || turn.Prompt != "original 中文🙂\nsecond line" || turn.Metadata["tui_text_claim"] != token || turn.SessionID != "A" {
				t.Fatal(turn, err)
			}
			if media {
				if turn.Metadata["tui_media_claim"] != token || turn.Metadata["media"] == nil {
					t.Fatal("paired receipt/media missing", turn.Metadata)
				}
			}
			if _, _, err = e.SubmitTUIComposer(ctx, "A", token, claim.Text.Revision, "bootstrap"); err == nil {
				t.Fatal("duplicate accepted")
			}
			turns, err := s.ListTurns(ctx, "A")
			if err != nil || len(turns) != 1 {
				t.Fatal(turns, err)
			}
		})
	}
}

func TestTUIComposerSubmitDispatchAcrossEnginesAdmitsOneSteering(t *testing.T) {
	e, s, path, claim := composerSubmitFixture(t, "follow-up 中文", true)
	ctx := context.Background()
	activeComposerFixture(t, s)
	otherStore, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer otherStore.Close()
	other := New(otherStore)
	defer other.Close()
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, engine := range []*Engine{e, other} {
		wg.Add(1)
		go func(engine *Engine) {
			defer wg.Done()
			<-start
			_, _, err := engine.SubmitTUIComposer(ctx, "A", claim.Text.Claim.Token, claim.Text.Revision, "bootstrap")
			results <- err
		}(engine)
	}
	close(start)
	wg.Wait()
	close(results)
	wins, rejected := 0, 0
	for err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, store.ErrTUIDraftHeld) || errors.Is(err, store.ErrTUIDraftConflict) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || rejected != 1 {
		t.Fatal(wins, rejected)
	}
	var count int
	var token, mediaToken, content string
	if err = s.DB().QueryRow(`select count(*),coalesce(json_extract(payload_json,'$.tui_text_claim'),''),coalesce(json_extract(payload_json,'$.tui_media_claim'),''),content from steering_queue`).Scan(&count, &token, &mediaToken, &content); err != nil || count != 1 || token != claim.Text.Claim.Token || mediaToken != token || content != "follow-up 中文" {
		t.Fatal(count, token, mediaToken, content, err)
	}
	pair, err := s.LoadTUIComposerDraft(ctx, "A")
	if err != nil || pair.Text.Claim != nil || pair.Media.Claim != nil {
		t.Fatal(pair, err)
	}
	turns, err := s.ListTurns(ctx, "A")
	if err != nil || len(turns) != 1 {
		t.Fatal("steering spawned duplicate turn", turns, err)
	}
}

func TestTUIComposerSubmitPublicMetadataCannotForgeReceipt(t *testing.T) {
	for _, steer := range []bool{false, true} {
		t.Run(map[bool]string{false: "turn", true: "steering"}[steer], func(t *testing.T) {
			e, s, _, claim := composerSubmitFixture(t, "held", true)
			ctx := context.Background()
			token := claim.Text.Claim.Token
			if steer {
				activeComposerFixture(t, s)
			}
			metadata := map[string]any{"tui_text_claim": token, "tui_media_claim": token, "custom": "preserved"}
			result, err := e.SubmitPromptRouted(ctx, RunInput{SessionID: "A", Prompt: "public", Model: "bootstrap", Metadata: metadata})
			if err != nil || result == nil {
				t.Fatal(result, err)
			}
			if !steer {
				waitRetryDone(t, s, result.TurnID)
			}
			var count int
			if err = s.DB().QueryRow(`select (select count(*) from turns where json_extract(metadata_json,'$.tui_text_claim')=? or json_extract(metadata_json,'$.tui_media_claim')=?)+(select count(*) from steering_queue where json_extract(payload_json,'$.tui_text_claim')=? or json_extract(payload_json,'$.tui_media_claim')=?)`, token, token, token, token).Scan(&count); err != nil || count != 0 {
				t.Fatal("forged receipt", count, err)
			}
			if metadata["tui_text_claim"] != token {
				t.Fatal("mutated caller metadata")
			}
			pair, err := s.ReconcileTUIComposerDraft(ctx, "A", token)
			if err != nil || pair.Text.Claim == nil || pair.Media.Claim == nil {
				t.Fatal("forgery settled pair", pair, err)
			}
		})
	}
}

func TestTUIComposerSubmitReadOnlyRejections(t *testing.T) {
	for _, tc := range []struct{ text, session, model string }{{"@peer do work", "A", "bootstrap"}, {"/retry run old", "A", "bootstrap"}, {"!!echo local", "A", "bootstrap"}, {"original", "A", ""}, {"original", "B", "bootstrap"}} {
		t.Run(tc.text+tc.session+tc.model, func(t *testing.T) {
			e, s, _, claim := composerSubmitFixture(t, tc.text, true)
			ctx := context.Background()
			before, err := s.LoadTUIComposerDraft(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if result, _, err := e.SubmitTUIComposer(ctx, tc.session, claim.Text.Claim.Token, claim.Text.Revision, tc.model); err == nil || result != nil {
				t.Fatal(result, err)
			}
			after, err := s.LoadTUIComposerDraft(ctx, "A")
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("preflight changed claim", after, err)
			}
			var sessions, turns, msgs int
			if err = s.DB().QueryRow(`select (select count(*) from sessions),(select count(*) from turns),(select count(*) from messages)`).Scan(&sessions, &turns, &msgs); err != nil || sessions != 2 || turns != 0 || msgs != 0 {
				t.Fatal("preflight routed/mutated", sessions, turns, msgs, err)
			}
		})
	}
}

func TestTUIComposerSubmitRestartAfterDispatchNeverReplays(t *testing.T) {
	e, s, path, claim := composerSubmitFixture(t, "unknown", true)
	ctx := context.Background()
	token := claim.Text.Claim.Token
	begun, err := s.BeginTUIComposerSubmission(ctx, "A", token, claim.Text.Revision)
	if err != nil {
		t.Fatal(err)
	}
	e.Close()
	s.Close()
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e = New(s)
	defer e.Close()
	if result, _, err := e.SubmitTUIComposer(ctx, "A", token, begun.Text.Revision, "bootstrap"); !errors.Is(err, store.ErrTUIDraftHeld) || result != nil {
		t.Fatal("restart replayed dispatch", result, err)
	}
	held, err := s.ReconcileTUIComposerDraft(ctx, "A", token)
	if err != nil || held.Text.Claim == nil || held.Media.Claim == nil || !held.Text.Claim.Dispatched {
		t.Fatal(held, err)
	}
	turns, err := s.ListTurns(ctx, "A")
	if err != nil || len(turns) != 0 {
		t.Fatal(turns, err)
	}
}

func TestTUIComposerSubmitWriteAndAdmissionFaults(t *testing.T) {
	for _, boundary := range []string{"dispatch", "before-insert", "post-insert", "settlement"} {
		t.Run(boundary, func(t *testing.T) {
			e, s, _, claim := composerSubmitFixture(t, "retain on failure", true)
			ctx := context.Background()
			token := claim.Text.Claim.Token
			triggers := map[string]string{
				"dispatch":      `create trigger fail_composer before update on kv_store when OLD.namespace='tui_text_draft_v1' and coalesce(json_extract(NEW.value,'$.claim.dispatched'),0)=1 begin select raise(abort,'dispatch denied'); end`,
				"before-insert": `create trigger fail_composer before insert on turns begin select raise(abort,'before admission'); end`,
				"post-insert":   `create trigger fail_composer before update on sessions begin select raise(abort,'post insert'); end`,
				"settlement":    `create trigger fail_composer before update on kv_store when OLD.namespace='tui_text_draft_v1' and json_extract(NEW.value,'$.claim') is null begin select raise(abort,'settlement denied'); end`,
			}
			if _, err := s.DB().Exec(triggers[boundary]); err != nil {
				t.Fatal(err)
			}
			result, _, err := e.SubmitTUIComposer(ctx, "A", token, claim.Text.Revision, "bootstrap")
			if err == nil {
				t.Fatal("expected failure", boundary)
			}
			if _, err = s.DB().Exec(`drop trigger fail_composer`); err != nil {
				t.Fatal(err)
			}
			if result != nil {
				waitRetryDone(t, s, result.TurnID)
			}
			pair, err := s.LoadTUIComposerDraft(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			switch boundary {
			case "dispatch":
				if pair.Text.Claim == nil || pair.Text.Claim.Dispatched || pair.Media.Claim == nil {
					t.Fatal(pair)
				}
			case "before-insert":
				if pair.Text.Claim != nil || pair.Text.Text != "retain on failure" || pair.Media.Claim != nil || len(pair.Media.Pending) != 1 {
					t.Fatal("known error not restored", pair)
				}
			case "post-insert", "settlement":
				if pair.Text.Claim == nil || !pair.Text.Claim.Dispatched || pair.Media.Claim == nil {
					t.Fatal("unknown claim lost", pair)
				}
				if _, _, err = e.SubmitTUIComposer(ctx, "A", token, pair.Text.Revision, "bootstrap"); !errors.Is(err, store.ErrTUIDraftHeld) {
					t.Fatal("error replayed", err)
				}
				recovered, err := s.ReconcileTUIComposerDraft(ctx, "A", token)
				if err != nil {
					t.Fatal(err)
				}
				if boundary == "settlement" && recovered.Text.Claim != nil {
					t.Fatal("confirmed recovery failed", recovered)
				}
				if boundary == "post-insert" && recovered.Text.Claim == nil {
					t.Fatal("unaudited row accepted")
				}
			}
		})
	}
}

func TestTUIComposerSubmitShellShortcutIsExplicit(t *testing.T) {
	e, s, _, claim := composerSubmitFixture(t, "!echo sample", false)
	ctx := context.Background()
	result, _, err := e.SubmitTUIComposer(ctx, "A", claim.Text.Claim.Token, claim.Text.Revision, "bootstrap")
	if err != nil || result == nil {
		t.Fatal(result, err)
	}
	waitRetryDone(t, s, result.TurnID)
	row, err := s.GetTurn(ctx, result.TurnID)
	if err != nil || !strings.HasPrefix(row.Prompt, "Run this shell command and summarize the result: echo sample") {
		t.Fatal(row, err)
	}
}

func TestTUIComposerSubmitMediaOnlyAdapterRejectsPairedAndWrongToken(t *testing.T) {
	e, s, _, pair := composerSubmitFixture(t, "paired", true)
	ctx := context.Background()
	input := RunInput{SessionID: "A", Prompt: "bypass", Model: "bootstrap"}
	if _, err := e.SubmitTUIMediaPrompt(ctx, input, pair.Text.Claim.Token); !errors.Is(err, store.ErrTUIDraftConflict) {
		t.Fatal("media adapter split paired claim", err)
	}
	ref, err := s.CreateMedia(ctx, "B", "other.txt", "text/plain", []byte("bytes"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.StageTUIMedia(ctx, "B", store.MediaRef{ID: store.MediaRefID(ref.ID), MediaID: ref.ID, SessionID: "B"}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ClaimTUIMedia(ctx, "B", "media-only", true); err != nil {
		t.Fatal(err)
	}
	input.SessionID = "B"
	if _, err = e.SubmitTUIMediaPrompt(ctx, input, "wrong"); !errors.Is(err, store.ErrTUIDraftConflict) {
		t.Fatal("wrong token accepted", err)
	}
	input.Metadata = map[string]any{"tui_text_claim": pair.Text.Claim.Token, "media": []string{"media:99999"}}
	result, err := e.SubmitTUIMediaPrompt(ctx, input, "media-only")
	if err != nil || result == nil {
		t.Fatal(result, err)
	}
	waitRetryDone(t, s, result.TurnID)
	row, err := s.GetTurn(ctx, result.TurnID)
	if err != nil || row.Metadata["tui_media_claim"] != "media-only" || row.Metadata["tui_text_claim"] != nil {
		t.Fatal(row, err)
	}
	if _, err = e.SubmitTUIMediaPrompt(ctx, input, "media-only"); err == nil {
		t.Fatal("confirmed media claim reused")
	}
}
