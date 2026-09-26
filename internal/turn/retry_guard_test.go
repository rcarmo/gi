package turn

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func heldRetryFixture(t *testing.T) (*Engine, *store.Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "retry.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	e := New(s)
	t.Cleanup(func() { e.Close(); s.Close() })
	ctx := context.Background()
	if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "old", "A", "failed", "original", map[string]any{"model": "bootstrap", "custom": "preserved"}); err != nil {
		t.Fatal(err)
	}
	if err = e.HoldTurnFailure(ctx, "old", "review", "test"); err != nil {
		t.Fatal(err)
	}
	return e, s, path
}
func retryCount(t *testing.T, s *store.Store) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRow(`SELECT COUNT(*) FROM turns WHERE json_extract(metadata_json,'$.retry_of_turn_id')='old'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
func waitRetryDone(t *testing.T, s *store.Store, id string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		r, err := s.GetTurn(context.Background(), id)
		if err == nil && r.FinishedAt != "" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("retry unfinished")
}

func TestRetryHeldGuardConcurrentClientsAdmitOne(t *testing.T) {
	e, s, path := heldRetryFixture(t)
	otherStore, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer otherStore.Close()
	other := New(otherStore)
	defer other.Close()
	start := make(chan struct{})
	var wg sync.WaitGroup
	results := make(chan *SubmitResult, 2)
	for _, engine := range []*Engine{e, other} {
		wg.Add(1)
		go func(engine *Engine) {
			defer wg.Done()
			<-start
			r, _ := engine.RetryHeldTurn(context.Background(), "old", "retry")
			results <- r
		}(engine)
	}
	close(start)
	wg.Wait()
	close(results)
	id := ""
	for result := range results {
		if result != nil {
			if id != "" && id != result.TurnID {
				t.Fatal("two retry IDs")
			}
			id = result.TurnID
		}
	}
	if id == "" || retryCount(t, s) != 1 {
		t.Fatal("duplicate/missing retry", id)
	}
	waitRetryDone(t, s, id)
	f, err := s.GetTurnFailure(context.Background(), "old")
	if err != nil || f.ResolutionState != "retried" || f.ResolvedTurnID != id {
		t.Fatal(f, err)
	}
}

func TestRetryHeldGuardFailuresBeforeAndAfterAdmission(t *testing.T) {
	for _, boundary := range []string{"before", "post-insert", "resolution"} {
		t.Run(boundary, func(t *testing.T) {
			e, s, _ := heldRetryFixture(t)
			ctx := context.Background()
			trigger := ""
			switch boundary {
			case "before":
				trigger = `CREATE TRIGGER retry_fail BEFORE INSERT ON turns WHEN json_extract(NEW.metadata_json,'$.retry_admission_token') IS NOT NULL BEGIN SELECT RAISE(ABORT,'before'); END`
			case "post-insert":
				trigger = `CREATE TRIGGER retry_fail BEFORE UPDATE ON sessions BEGIN SELECT RAISE(ABORT,'after insert'); END`
			case "resolution":
				trigger = `CREATE TRIGGER retry_fail BEFORE UPDATE ON turn_failures WHEN NEW.resolution_state='retried' BEGIN SELECT RAISE(ABORT,'resolution'); END`
			}
			if _, err := s.DB().Exec(trigger); err != nil {
				t.Fatal(err)
			}
			result, err := e.RetryHeldTurn(ctx, "old", "retry")
			if err == nil || result != nil {
				t.Fatal("atomic fault must roll back admission", boundary, result, err)
			}
			if _, err := s.DB().Exec(`DROP TRIGGER retry_fail`); err != nil {
				t.Fatal(err)
			}
			if retryCount(t, s) != 0 {
				t.Fatal("rollback left an admission")
			}
			f, getErr := s.GetTurnFailure(ctx, "old")
			if getErr != nil || f.ResolutionState != "" || f.HoldState != "review" {
				t.Fatal("reservation not released", f, getErr)
			}
			result, err = e.RetryHeldTurn(ctx, "old", "recover")
			if err != nil || result == nil {
				t.Fatal(result, err)
			}
			if retryCount(t, s) != 1 {
				t.Fatal("duplicate on recover")
			}
			waitRetryDone(t, s, result.TurnID)
		})
	}
}

func TestRetryHeldGuardRecoveredReservationDoesNotResend(t *testing.T) {
	e, s, _ := heldRetryFixture(t)
	ctx := context.Background()
	ok, err := s.ReserveHeldRetry(ctx, "old", "crashed")
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err = e.RetryHeldTurn(ctx, "old", "recover"); !errors.Is(err, store.ErrRetryPending) {
		t.Fatal("unknown replay", err)
	}
	if err = e.SkipHeldTurn(ctx, "old", "skip"); !errors.Is(err, store.ErrFailureConflict) {
		t.Fatal("skip overwrote reservation", err)
	}
	if err = e.HoldTurnFailure(ctx, "old", "review", "again"); !errors.Is(err, store.ErrFailureConflict) {
		t.Fatal("rehold overwrote reservation", err)
	}
	if retryCount(t, s) != 0 {
		t.Fatal("unexpected retry")
	}
	// Admission by the original caller may become visible after recovery checks.
	if _, err = s.CreateTurnWithStatus(ctx, "admitted", "A", "queued", "original", map[string]any{"retry_of_turn_id": "old", "retry_admission_token": "crashed"}); err != nil {
		t.Fatal(err)
	}
	if _, err = e.RetryHeldTurn(ctx, "old", "recover"); !errors.Is(err, store.ErrRetryPending) {
		t.Fatal("insert accepted before rollback boundary", err)
	}
	if err = s.AppendTurnEvent(ctx, "admitted", "A", "turn.submitted", nil); err != nil {
		t.Fatal(err)
	}
	result, err := e.RetryHeldTurn(ctx, "old", "recover")
	if err != nil || result.TurnID != "admitted" || retryCount(t, s) != 1 {
		t.Fatal(result, err)
	}
}

func TestRetryHeldGuardQueuesInsteadOfSteeringIntoActiveTurn(t *testing.T) {
	e, s, _ := heldRetryFixture(t)
	ctx := context.Background()
	if _, err := s.CreateTurnWithStatus(ctx, "active", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "active", "test", "active-claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	result, err := e.RetryHeldTurn(ctx, "old", "retry")
	if err != nil || result == nil || !result.Queued || result.TurnID == "active" {
		t.Fatal(result, err)
	}
	rec, _ := s.GetTurn(ctx, result.TurnID)
	if rec.Metadata["custom"] != "preserved" || rec.Metadata["retry_admission_token"] == "" || rec.Metadata["intent"] != "queue" {
		t.Fatal(rec.Metadata)
	}
	var count int
	if err := s.DB().QueryRow(`SELECT COUNT(*) FROM steering_queue`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("retry steered")
	}
}

func TestRetryHeldGuardStaleClaimDoesNotReplayOriginal(t *testing.T) {
	for _, pending := range []bool{false, true} {
		t.Run(fmt.Sprint(pending), func(t *testing.T) {
			e, s, _ := heldRetryFixture(t)
			ctx := context.Background()
			if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "old", "dead-worker", "stale"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			if _, err := s.DB().Exec(`UPDATE session_active_turns SET updated_at='2000-01-01T00:00:00Z' WHERE session_id='A'`); err != nil {
				t.Fatal(err)
			}
			if pending {
				if ok, err := s.ReserveHeldRetry(ctx, "old", "dead-reservation"); err != nil || !ok {
					t.Fatal(ok, err)
				}
			}
			if _, err := e.recoverInterruptedTurns(ctx, "A"); err != nil {
				t.Fatal(err)
			}
			old, err := s.GetTurn(ctx, "old")
			if err != nil || old.Status != "failed" || old.Phase != "held_for_retry_or_skip" {
				t.Fatal("held original replayed", old, err)
			}
			if _, _, err := s.GetSessionActiveTurn(ctx, "A"); err != sql.ErrNoRows {
				t.Fatal("stale claim retained", err)
			}
			if retryCount(t, s) != 0 {
				t.Fatal("recovery admitted work")
			}
			result, err := e.RetryHeldTurn(ctx, "old", "explicit")
			if pending {
				if !errors.Is(err, store.ErrRetryPending) || result != nil {
					t.Fatal("unknown replay", result, err)
				}
				return
			}
			if err != nil || result == nil {
				t.Fatal(result, err)
			}
			waitRetryDone(t, s, result.TurnID)
			if retryCount(t, s) != 1 {
				t.Fatal("explicit retry duplicated")
			}
		})
	}
}

func TestRetryHeldGuardDoesNotReplayContinuationOrRoutingMetadata(t *testing.T) {
	e, s, _ := heldRetryFixture(t)
	ctx := context.Background()
	metadata := map[string]any{"intent": "prompt", "model": "bootstrap", "custom": "preserved", "continue": true, "initial_steering": []any{map[string]any{"content": "old-steering-must-not-replay", "role": "user"}}, "source_session_id": "source", "target_session_id": "A", "source_agent_id": "source-agent", "target_agent_id": "old-target", "route_mode": "prompt", "routing_enabled": true, "routed_from_prompt": true, "ingress_role": "system", "tui_media_claim": "old-claim"}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`UPDATE turns SET metadata_json=? WHERE id='old'`, string(encoded)); err != nil {
		t.Fatal(err)
	}
	result, err := e.RetryHeldTurn(ctx, "old", "retry")
	if err != nil || result == nil {
		t.Fatal(result, err)
	}
	waitRetryDone(t, s, result.TurnID)
	rec, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"initial_steering", "continue", "source_session_id", "source_agent_id", "target_session_id", "target_agent_id", "route_mode", "routing_enabled", "routed_from_prompt", "ingress_role", "tui_media_claim"} {
		if _, ok := rec.Metadata[key]; ok {
			t.Fatalf("copied execution bookkeeping %s", key)
		}
	}
	if rec.Metadata["intent"] != "queue" || rec.Metadata["custom"] != "preserved" {
		t.Fatal(rec.Metadata)
	}
	var n int
	if err = s.DB().QueryRow(`SELECT count(*) FROM messages WHERE content LIKE '%old-steering-must-not-replay%'`).Scan(&n); err != nil || n != 0 {
		t.Fatal("steering replayed", n, err)
	}
	for _, state := range []string{"held", "resolved"} {
		if err = e.HoldTurnFailure(ctx, "old", "review", state); !errors.Is(err, store.ErrFailureConflict) {
			t.Fatal("rehold reopened resolution", err)
		}
	}
	if recovered, err := e.RetryHeldTurn(ctx, "old", "again"); err != nil || recovered.TurnID != result.TurnID {
		t.Fatal("resolved admission not recovered", recovered, err)
	}
	if retryCount(t, s) != 1 {
		t.Fatal("duplicate retry")
	}
}

func TestRetryHeldGuardConcurrentObserverCannotResolveRollbackableSubturn(t *testing.T) {
	e, s, path := heldRetryFixture(t)
	ctx := context.Background()
	if _, err := s.CreateTurnWithStatus(ctx, "parent", "A", "completed", "parent", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`UPDATE turns SET metadata_json=json_set(metadata_json,'$.parent_turn_id','parent') WHERE id='old'`); err != nil {
		t.Fatal(err)
	}
	otherStore, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer otherStore.Close()
	other := New(otherStore)
	defer other.Close()
	inserted := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	e.beforeCreateSubTurnErrorHook = func(context.Context, string, string) error {
		close(inserted)
		<-release
		return errors.New("injected subturn rollback")
	}
	go func() { _, err := e.RetryHeldTurn(ctx, "old", "owner"); done <- err }()
	<-inserted
	// Subturn validation now precedes the atomic admission; the observer must
	// see no provisional turn while the owner is paused here.
	if retryCount(t, s) != 0 {
		close(release)
		<-done
		t.Fatal("provisional retry visible")
	}
	result, observerErr := other.RetryHeldTurn(ctx, "old", "observer")
	close(release)
	ownerErr := <-done
	if !errors.Is(observerErr, store.ErrRetryPending) || result != nil {
		t.Fatal("observer resolved rollbackable row", result, observerErr)
	}
	if ownerErr == nil || retryCount(t, s) != 0 {
		t.Fatal("rollback/release failed", ownerErr)
	}
	f, err := s.GetTurnFailure(ctx, "old")
	if err != nil || f.ResolutionState != "" || f.RetryAdmissionToken != "" || f.HoldState != "review" {
		t.Fatal(f, err)
	}
}

func TestRetryHeldGuardMissingSubmissionAuditRecoversAtomicAdmissionAfterReopen(t *testing.T) {
	e, s, path := heldRetryFixture(t)
	ctx := context.Background()
	// Keep follow-on queued so no background execution can confound the audit
	// failure. Failed resolution now rolls back the whole admission.
	if _, err := s.CreateTurnWithStatus(ctx, "active", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "active", "live", "active-token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	for _, trigger := range []string{
		`CREATE TRIGGER lose_retry_audit BEFORE INSERT ON turn_events WHEN NEW.event_type='turn.submitted' BEGIN SELECT RAISE(ABORT,'lost audit'); END`,
		`CREATE TRIGGER lose_retry_resolution BEFORE UPDATE ON turn_failures WHEN NEW.resolution_state='retried' BEGIN SELECT RAISE(ABORT,'lost resolution'); END`,
	} {
		if _, err := s.DB().Exec(trigger); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := e.RetryHeldTurn(ctx, "old", "owner"); err == nil || result != nil {
		t.Fatal("expected unresolved result", result, err)
	}
	if retryCount(t, s) != 0 {
		t.Fatal("resolution failure left a retry turn")
	}
	if _, err := s.DB().Exec(`DROP TRIGGER lose_retry_resolution`); err != nil {
		t.Fatal(err)
	}
	admitted, err := e.RetryHeldTurn(ctx, "old", "owner")
	if err != nil || admitted == nil {
		t.Fatal(admitted, err)
	}
	if retryCount(t, s) != 1 {
		t.Fatal("missing atomic admission")
	}
	e.Close()
	s.Close()
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	recovery := New(reopened)
	defer recovery.Close()
	for _, trigger := range []string{"lose_retry_audit"} {
		if _, err = reopened.DB().Exec(`DROP TRIGGER ` + trigger); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := recovery.RetryHeldTurn(ctx, "old", "recover"); err != nil || result == nil || result.TurnID != admitted.TurnID {
		t.Fatal("committed admission not recovered", result, err)
	}
	if err := recovery.SkipHeldTurn(ctx, "old", "skip"); err == nil {
		t.Fatal("resolved admission skipped")
	}
	var audits int
	if err = reopened.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='turn.submitted'`, admitted.TurnID).Scan(&audits); err != nil || audits != 0 {
		t.Fatal("audit failure not exercised", audits, err)
	}
	if retryCount(t, reopened) != 1 {
		t.Fatal("duplicate admission")
	}
}

func TestRetryHeldGuardExplicitReleaseFencesPausedSubmitter(t *testing.T) {
	e, s, path := heldRetryFixture(t)
	ctx := context.Background()
	if _, err := s.CreateTurnWithStatus(ctx, "parent", "A", "completed", "parent", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`UPDATE turns SET metadata_json=json_set(metadata_json,'$.parent_turn_id','parent') WHERE id='old'`); err != nil {
		t.Fatal(err)
	}
	other, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	paused := make(chan struct{})
	resume := make(chan struct{})
	done := make(chan error, 1)
	e.beforeCreateSubTurnErrorHook = func(context.Context, string, string) error { close(paused); <-resume; return nil }
	go func() { _, err := e.RetryHeldTurn(ctx, "old", "owner"); done <- err }()
	<-paused
	f, readErr := other.GetTurnFailure(ctx, "old")
	if readErr != nil {
		close(resume)
		<-done
		t.Fatal(readErr)
	}
	releaseErr := other.ReleaseUnadmittedHeldRetry(ctx, "A", "old", f.RetryAdmissionToken)
	close(resume)
	ownerErr := <-done
	if releaseErr != nil || !errors.Is(ownerErr, store.ErrFailureConflict) {
		t.Fatal("fence failure", releaseErr, ownerErr)
	}
	if retryCount(t, s) != 0 {
		t.Fatal("late caller admitted")
	}
	e.beforeCreateSubTurnErrorHook = nil
	result, err := e.RetryHeldTurn(ctx, "old", "new explicit retry")
	if err != nil || result == nil {
		t.Fatal(result, err)
	}
	waitRetryDone(t, s, result.TurnID)
	if retryCount(t, s) != 1 {
		t.Fatal("duplicate after release")
	}
}

func TestRetryHeldGuardPublicSubmissionCannotForgeReceipt(t *testing.T) {
	e, s, _ := heldRetryFixture(t)
	ctx := context.Background()
	ok, err := s.ReserveAtomicHeldRetry(ctx, "old", "private-token")
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	result, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "ordinary", Intent: "queue", Model: "bootstrap", Metadata: map[string]any{"retry_of_turn_id": "old", "retry_admission_token": "private-token", "custom": "preserved"}})
	if err != nil || result == nil {
		t.Fatal(result, err)
	}
	waitRetryDone(t, s, result.TurnID)
	row, err := s.GetTurn(ctx, result.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Metadata["retry_of_turn_id"] != nil || row.Metadata["retry_admission_token"] != nil || row.Metadata["custom"] != "preserved" {
		t.Fatal(row.Metadata)
	}
	if retryCount(t, s) != 0 {
		t.Fatal("public submit forged retry receipt")
	}
	if err = s.ReleaseUnadmittedHeldRetry(ctx, "A", "old", "private-token"); err != nil {
		t.Fatal(err)
	}
}

func TestRetryHeldGuardShutdownDoesNotPassNilContextToQueueHandoff(t *testing.T) {
	e, _, _ := heldRetryFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	e.Close()
	runner := e.runner("A")
	runner.mu.Lock()
	launched, err := e.startNextQueuedTurnLocked(ctx, runner, "A")
	runner.mu.Unlock()
	if launched || !errors.Is(err, context.Canceled) {
		t.Fatal(launched, err)
	}
	if _, err = e.RetryHeldTurn(ctx, "old", "shutdown"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err = e.ReleaseHeldRetryInSession(ctx, "A", "old", "token"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRetryHeldGuardShutdownSkipsSubTurnLifecycleWithoutLiveContext(t *testing.T) {
	e, s, _ := heldRetryFixture(t)
	ctx := context.Background()
	if _, err := s.CreateTurnWithStatus(ctx, "parent", "A", "completed", "parent", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSubTurn(ctx, "parent", "A", "old", "A", "async", 1, nil); err != nil {
		t.Fatal(err)
	}
	before, err := s.GetSubTurnByChild(ctx, "old")
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	e.Close()
	// Previously panicked inside database/sql while holding its internal mutex,
	// then blocked cleanup and Store.Close until the Go test timeout.
	e.runner("A").publishSubTurnLifecycle(canceled, "old", "completed")
	after, err := s.GetSubTurnByChild(ctx, "old")
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("shutdown lifecycle mutated stored subturn", after, err)
	}
}
