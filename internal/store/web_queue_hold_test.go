package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
)

type steeringHoldSnapshot struct {
	ID          int64
	TurnID      string
	Content     string
	PayloadJSON string
	MediaJSON   string
	Status      string
	CreatedAt   string
	UpdatedAt   string
}

func openWebQueueHoldStore(t *testing.T, path string) *Store {
	t.Helper()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	installWebQueueHoldTable(t, s)
	return s
}

func installWebQueueHoldTable(t *testing.T, s *Store) {
	t.Helper()
}

func holdRowCount(t *testing.T, s *Store, sessionID string) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRow(`select count(*) from web_queue_holds where session_id=?`, sessionID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func readSteeringHoldSnapshot(t *testing.T, s *Store, id int64) steeringHoldSnapshot {
	t.Helper()
	var snap steeringHoldSnapshot
	if err := s.DB().QueryRow(`select id, coalesce(turn_id,''), content, payload_json, media_json, status, created_at, updated_at from steering_queue where id=?`, id).Scan(
		&snap.ID, &snap.TurnID, &snap.Content, &snap.PayloadJSON, &snap.MediaJSON, &snap.Status, &snap.CreatedAt, &snap.UpdatedAt,
	); err != nil {
		t.Fatal(err)
	}
	return snap
}

func TestWebQueueHoldStopExactOwnershipAndResumeFence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "web-queue-hold.db")
	s := openWebQueueHoldStore(t, path)
	defer s.Close()
	for _, id := range []string{"A", "B"} {
		if _, err := s.CreateSession(ctx, id, id, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []struct{ id, session, status string }{{"runA", "A", "running"}, {"runB", "B", "running"}, {"qA", "A", "queued"}} {
		if _, err := s.CreateTurnWithStatus(ctx, v.id, v.session, v.status, v.id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "runA", "runner", "tokenA"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "B", "runB", "runner", "tokenB"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err := s.DB().Exec(`insert into web_queue_holds(session_id,stop_turn_id,created_at) values('A','stale',datetime('now'))`); err != nil {
		t.Fatal(err)
	}
	if err := s.StopWebActiveTurn(ctx, "A", "runA", "tokenA"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	runA, err := s.GetTurn(ctx, "runA")
	if err != nil {
		t.Fatal(err)
	}
	if runA.Status != "running" || runA.Phase != "setup" {
		t.Fatal(runA)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "stale" {
		t.Fatal(hold, err)
	}
	if _, err := s.DB().Exec(`delete from web_queue_holds where session_id='A'`); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ session, turn, token string }{
		{session: "A", turn: "runA", token: "wrong"},
		{session: "A", turn: "missing", token: "tokenA"},
		{session: "B", turn: "runA", token: "tokenA"},
	} {
		if err := s.StopWebActiveTurn(ctx, tc.session, tc.turn, tc.token); !errors.Is(err, ErrQueueConflict) {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
	if err := s.StopWebActiveTurn(ctx, "A", "runA", "tokenA"); err != nil {
		t.Fatal(err)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "runA" {
		t.Fatal(hold, err)
	}
	runA, err = s.GetTurn(ctx, "runA")
	if err != nil {
		t.Fatal(err)
	}
	if runA.Status != "cancelling" || runA.Phase != "cancelling" {
		t.Fatal(runA)
	}
	if err := s.StopWebActiveTurn(ctx, "A", "runA", "tokenA"); err != nil {
		t.Fatal(err)
	}
	if n := holdRowCount(t, s, "A"); n != 1 {
		t.Fatal(n)
	}
	if err := s.ResumeWebQueue(ctx, "A", "stale"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	if err := s.ResumeWebQueue(ctx, "A", "runA"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "runA" {
		t.Fatal(hold, err)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "A", "tokenA"); err != nil {
		t.Fatal(err)
	}
	if err := s.ResumeWebQueue(ctx, "A", "runA"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("cleared before claim", err)
	}
	if ok, err := s.ClaimWebResumedTurn(ctx, "A", "qA", "runner", "qA", "runA"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := s.ResumeWebQueue(ctx, "A", "runA"); err != nil {
		t.Fatal(err)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "" {
		t.Fatal(hold, err)
	}
	if err := s.ResumeWebQueue(ctx, "A", "runA"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
}

func TestWebQueueHoldPreservesQueuedRowsOrderMediaAndSteeringMetadata(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "preserve.db")
	s := openWebQueueHoldStore(t, path)
	defer s.Close()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	media, err := s.CreateMedia(ctx, "A", "queue.txt", "text/plain", []byte("queued bytes"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "run", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "q1", "A", "queued", "first", map[string]any{"media": []string{MediaRefID(media.ID)}, "custom": map[string]any{"nested": "kept"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "q2", "A", "queued", "second", map[string]any{"client_request_id": "keep"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "q3", "A", "queued", "third", map[string]any{"custom": "three"}); err != nil {
		t.Fatal(err)
	}
	if err := s.ReorderQueuedTurns(ctx, "A", []string{"q1", "q2", "q3"}, []string{"q3", "q1", "q2"}); err != nil {
		t.Fatal(err)
	}
	queuedBefore, err := s.ListQueuedTurns(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := []string{"q3", "q1", "q2"}
	gotOrder := []string{queuedBefore[0].ID, queuedBefore[1].ID, queuedBefore[2].ID}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatal(gotOrder)
	}
	steerID, err := s.EnqueueSteering(ctx, "A", "", "user", "queued steer", map[string]any{"kind": "steering", "media": []string{MediaRefID(media.ID)}, "custom": map[string]any{"keep": true}}, []string{MediaRefID(media.ID)}, "all")
	if err != nil {
		t.Fatal(err)
	}
	steeringBefore := readSteeringHoldSnapshot(t, s, steerID)
	if err := s.AppendTurnEvent(ctx, "run", "A", "turn.checkpoint", map[string]any{"phase": "turn", "checkpoint": true}); err != nil {
		t.Fatal(err)
	}
	var eventCountBefore int
	if err := s.DB().QueryRow(`select count(*) from turn_events where session_id='A'`).Scan(&eventCountBefore); err != nil {
		t.Fatal(err)
	}
	if err := s.StopWebActiveTurn(ctx, "A", "run", "token"); err != nil {
		t.Fatal(err)
	}
	queuedAfter, err := s.ListQueuedTurns(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(queuedBefore, queuedAfter) {
		t.Fatalf("queued rows mutated\nbefore=%#v\nafter=%#v", queuedBefore, queuedAfter)
	}
	steeringAfter := readSteeringHoldSnapshot(t, s, steerID)
	if !reflect.DeepEqual(steeringBefore, steeringAfter) {
		t.Fatalf("steering mutated\nbefore=%#v\nafter=%#v", steeringBefore, steeringAfter)
	}
	var eventCountAfter int
	if err := s.DB().QueryRow(`select count(*) from turn_events where session_id='A'`).Scan(&eventCountAfter); err != nil {
		t.Fatal(err)
	}
	if eventCountAfter != eventCountBefore {
		t.Fatal(eventCountBefore, eventCountAfter)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "run" {
		t.Fatal(hold, err)
	}
}

func TestWebQueueHoldNoHoldWhenQueueEmpty(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "empty.db")
	s := openWebQueueHoldStore(t, path)
	defer s.Close()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "run", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := s.StopWebActiveTurn(ctx, "A", "run", "token"); err != nil {
		t.Fatal(err)
	}
	if err := s.StopWebActiveTurn(ctx, "A", "run", "token"); err != nil {
		t.Fatal(err)
	}
	run, err := s.GetTurn(ctx, "run")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != "cancelling" || run.Phase != "cancelling" {
		t.Fatal(run)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "" {
		t.Fatal(hold, err)
	}
	if n := holdRowCount(t, s, "A"); n != 0 {
		t.Fatal(n)
	}
	if err := s.ResumeWebQueue(ctx, "A", "run"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
}

func TestWebQueueHoldRollbackAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "rollback.db")
	s := openWebQueueHoldStore(t, path)
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "run", "A", "running", "active", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "q1", "A", "queued", "later", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "token"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err := s.DB().Exec(`create trigger reject_web_queue_hold before insert on web_queue_holds begin select raise(abort,'hold failed'); end`); err != nil {
		t.Fatal(err)
	}
	err := s.StopWebActiveTurn(ctx, "A", "run", "token")
	if err == nil || errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
	run, getErr := s.GetTurn(ctx, "run")
	if getErr != nil {
		t.Fatal(getErr)
	}
	if run.Status != "running" || run.Phase != "setup" {
		t.Fatal(run)
	}
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "" {
		t.Fatal(hold, err)
	}
	if _, err := s.DB().Exec(`drop trigger reject_web_queue_hold`); err != nil {
		t.Fatal(err)
	}
	if err := s.StopWebActiveTurn(ctx, "A", "run", "token"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = openWebQueueHoldStore(t, path)
	defer s.Close()
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "run" {
		t.Fatal(hold, err)
	}
	run, err = s.GetTurn(ctx, "run")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != "cancelling" || run.Phase != "cancelling" {
		t.Fatal(run)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "A", "token"); err != nil {
		t.Fatal(err)
	}
	if err := s.ResumeWebQueue(ctx, "A", "run"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal("cleared before claim", err)
	}
	if ok, err := s.ClaimWebResumedTurn(ctx, "A", "q1", "runner", "q1", "run"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if err := s.ResumeWebQueue(ctx, "A", "run"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseSessionActiveTurn(ctx, "A", "q1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = openWebQueueHoldStore(t, path)
	defer s.Close()
	if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "" {
		t.Fatal(hold, err)
	}
	var active sql.NullString
	if err := s.DB().QueryRow(`select turn_id from session_active_turns where session_id='A'`).Scan(&active); err != sql.ErrNoRows {
		t.Fatal(active, err)
	}
}

func TestWebQueueHoldConcurrentResumeClaimAndHoldRetirement(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "race.db")
	a := openWebQueueHoldStore(t, path)
	defer a.Close()
	b := openWebQueueHoldStore(t, path)
	defer b.Close()
	a.CreateSession(ctx, "A", "A", nil)
	a.CreateTurnWithStatus(ctx, "q", "A", "queued", "next", nil)
	if _, err := a.DB().Exec(`insert into web_queue_holds values('A','stop',datetime('now'))`); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan bool, 2)
	errs := make(chan error, 2)
	for _, s := range []*Store{a, b} {
		go func(s *Store) {
			<-start
			ok, err := s.ClaimWebResumedTurn(ctx, "A", "q", "runner", "q", "stop")
			results <- ok
			errs <- err
		}(s)
	}
	close(start)
	winners := 0
	for i := 0; i < 2; i++ {
		if <-results {
			winners++
		}
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatal(winners)
	}
	if hold, err := a.WebQueueHold(ctx, "A"); err != nil || hold != "stop" {
		t.Fatal(hold, err)
	}
	if err := a.ResumeWebQueue(ctx, "A", "stop"); err != nil {
		t.Fatal(err)
	}
	if err := b.ResumeWebQueue(ctx, "A", "stop"); !errors.Is(err, ErrQueueConflict) {
		t.Fatal(err)
	}
}

func TestWebQueueHoldReturnedSteeringDoesNotTrapResume(t *testing.T) {
	for _, bound := range []bool{false, true} {
		t.Run(fmt.Sprint(bound), func(t *testing.T) {
			ctx := context.Background()
			s := openWebQueueHoldStore(t, filepath.Join(t.TempDir(), "returned.db"))
			defer s.Close()
			s.CreateSession(ctx, "A", "A", nil)
			s.CreateTurnWithStatus(ctx, "run", "A", "running", "run", nil)
			s.CreateTurnWithStatus(ctx, "q", "A", "queued", "keep", map[string]any{"marker": "unchanged"})
			s.ClaimSessionActiveTurn(ctx, "A", "run", "runner", "run")
			if bound {
				if err := s.SteerQueuedTurn(ctx, "A", "q", "run"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := s.UpdateTurnStatusAndPhase(ctx, "q", "queued", "steer_returned"); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.StopWebActiveTurn(ctx, "A", "run", "run"); err != nil {
				t.Fatal(err)
			}
			hold, err := s.WebQueueHold(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if bound && hold != "run" {
				t.Fatal(hold)
			}
			if !bound && hold != "" {
				t.Fatal("nonrunnable row created hold", hold)
			}
			if err := s.ReleaseSessionActiveTurn(ctx, "A", "run"); err != nil {
				t.Fatal(err)
			}
			if bound {
				if err := s.ResumeWebQueue(ctx, "A", "run"); err != nil {
					t.Fatal(err)
				}
			}
			q, err := s.GetTurn(ctx, "q")
			if err != nil || q.Status != "queued" || q.Phase != "steer_returned" || q.Metadata["marker"] != "unchanged" {
				t.Fatal(q, err)
			}
			if _, err := s.GetNextQueuedTurn(ctx, "A"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("returned row became runnable", err)
			}
			if hold, err := s.WebQueueHold(ctx, "A"); err != nil || hold != "" {
				t.Fatal(hold, err)
			}
		})
	}
}
