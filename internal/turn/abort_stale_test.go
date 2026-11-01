package turn

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// A turn left 'running' by a crash is finalized as aborted by /abort instead
// of keeping the session busy or being replayed; a fresh claim from another
// live process is refused.
func TestAbortStaleActiveTurn(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	engine := New(s)
	defer engine.Close()

	sess, err := s.CreateSession(ctx, "session_abort_stale", "Abort", map[string]any{"status": "running"})
	if err != nil {
		t.Fatal(err)
	}
	tr, err := s.CreateTurnWithStatus(ctx, "turn_abort_stale", sess.ID, "running", "make it so", map[string]any{"intent": "prompt", "model": "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, sess.ID, tr.ID, "other-worker", "claim-abort-stale"); err != nil || !ok {
		t.Fatalf("claim: ok=%v err=%v", ok, err)
	}
	if err := s.TouchSessionState(ctx, sess.ID, map[string]any{"active_turn_id": tr.ID, "status": "running"}); err != nil {
		t.Fatal(err)
	}

	// Fresh heartbeat: another live process owns it.
	if _, err := engine.AbortStaleActiveTurn(ctx, sess.ID); !errors.Is(err, ErrTurnOwnedElsewhere) {
		t.Fatalf("fresh claim: err=%v, want ErrTurnOwnedElsewhere", err)
	}

	stale := time.Now().Add(-(interruptedTurnStaleAfter + 5*time.Second)).UTC().Format(time.RFC3339Nano)
	if _, err := s.DB().ExecContext(ctx, `update session_active_turns set updated_at=? where session_id=?`, stale, sess.ID); err != nil {
		t.Fatal(err)
	}
	busy, err := s.CompactionBusy(ctx, sess.ID)
	if err != nil || !busy {
		t.Fatalf("precondition: stale claim should make the session busy (busy=%v err=%v)", busy, err)
	}
	aborted, err := engine.AbortStaleActiveTurn(ctx, sess.ID)
	if err != nil || aborted != tr.ID {
		t.Fatalf("abort stale: id=%q err=%v", aborted, err)
	}
	got, err := s.GetTurn(ctx, tr.ID)
	if err != nil || got.Status != "aborted" {
		t.Fatalf("turn status=%q err=%v, want aborted", got.Status, err)
	}
	if id, _, err := s.GetSessionActiveTurn(ctx, sess.ID); !errors.Is(err, sql.ErrNoRows) && (err != nil || id != "") {
		t.Fatalf("claim not released: %q %v", id, err)
	}
	if busy, err := s.CompactionBusy(ctx, sess.ID); err != nil || busy {
		t.Fatalf("session still busy after abort (busy=%v err=%v)", busy, err)
	}
	if again, err := engine.AbortStaleActiveTurn(ctx, sess.ID); err != nil || again != "" {
		t.Fatalf("second abort: id=%q err=%v, want nothing", again, err)
	}
}

// A clean Close joins the live host and retains its verified continuation,
// never granting the ordinary queue authority to replay the original prompt.
func TestClosePreservesLiveGenerationCheckpoint(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-ctx.Done() // a long request that only ends when cancelled
		return nil, ctx.Err()
	})
	engine := New(s)
	sess, err := s.CreateSession(ctx, "session_close_live", "Close", map[string]any{"status": "idle"})
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.SubmitPrompt(ctx, RunInput{SessionID: sess.ID, Prompt: "make it so", Model: "mock-close"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("turn never reached inference")
	}
	if err = engine.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetTurn(ctx, res.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "queued" || got.Phase != store.ClosedGenerationReadyPhase || got.FinishedAt != "" {
		t.Fatalf("Close lost continuation: %#v", got)
	}
	if _, err = s.GetNextQueuedTurn(ctx, sess.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("checkpoint became ordinary prompt replay: %v", err)
	}
	if id, token, err := s.GetSessionActiveTurn(ctx, sess.ID); err != nil || id != res.TurnID || token == "" {
		t.Fatalf("checkpoint reservation lost: %q %q %v", id, token, err)
	}
	s.Close()
}
