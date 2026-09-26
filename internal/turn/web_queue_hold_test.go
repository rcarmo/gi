package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestWebQueueHoldStopCleanupAndExplicitResume(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	if _, err := s.CreateSession(ctx, "hold", "hold", nil); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{}, 1)
	withStreamWithToolsStub(t, func(ctx context.Context, _ string, _ *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
		select {
		case started <- struct{}{}:
		default:
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	first, err := e.SubmitPrompt(ctx, RunInput{SessionID: "hold", Prompt: "first", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("not started")
	}
	second, err := e.SubmitPrompt(ctx, RunInput{SessionID: "hold", Prompt: "second", Intent: "queue", Model: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	third, err := e.SubmitPrompt(ctx, RunInput{SessionID: "hold", Prompt: "third", Intent: "queue", Model: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if err = e.StopWebActiveTurn(ctx, "hold", first.TurnID); err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, 3*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "hold"); return errors.Is(err, sql.ErrNoRows) }, "cleanup releases claim")
	for _, id := range []string{second.TurnID, third.TurnID} {
		v, err := s.GetTurn(ctx, id)
		if err != nil || v.Status != "queued" {
			t.Fatal(v, err)
		}
	}
	if hold, err := s.WebQueueHold(ctx, "hold"); err != nil || hold != first.TurnID {
		t.Fatal(hold, err)
	}
	if _, err = e.ContinueSession(ctx, "hold"); !errors.Is(err, store.ErrQueueConflict) {
		t.Fatal("unfenced continue", err)
	}
	if _, err = e.ResumeWebQueue(ctx, "hold", "stale"); !errors.Is(err, store.ErrQueueConflict) {
		t.Fatal("stale resume", err)
	}
	later, err := e.SubmitPrompt(ctx, RunInput{SessionID: "hold", Prompt: "later", Model: "bootstrap"})
	if err != nil || !later.Queued {
		t.Fatal(later, err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "hold", second.TurnID, "foreign", "foreign"); err != nil || ok {
		t.Fatal("claim bypassed hold", ok, err)
	}
	if ok, err := e.ResumeWebQueue(ctx, "hold", first.TurnID); err != nil || !ok {
		t.Fatal(ok, err)
	}
	waitForCondition(t, 3*time.Second, func() bool {
		v, err := s.GetTurn(ctx, second.TurnID)
		return err == nil && (v.Status == "completed" || v.Status == "failed")
	}, "resumed completion")
	waitForCondition(t, 3*time.Second, func() bool {
		v, err := s.GetTurn(ctx, third.TurnID)
		return err == nil && (v.Status == "completed" || v.Status == "failed")
	}, "resumed completion")
	waitForCondition(t, 3*time.Second, func() bool {
		v, err := s.GetTurn(ctx, later.TurnID)
		return err == nil && (v.Status == "completed" || v.Status == "failed")
	}, "resumed completion")
	waitForCondition(t, 3*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "hold"); return errors.Is(err, sql.ErrNoRows) }, "final cleanup")
	if _, err = e.ResumeWebQueue(ctx, "hold", first.TurnID); !errors.Is(err, store.ErrQueueConflict) {
		t.Fatal("duplicate resume", err)
	}
}

func TestWebQueueHoldCrashBeforeCleanupSurvivesRecovery(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "hold.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateSession(ctx, "held", "held", nil); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "stopped", "held", "running", "old", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "held", "stopped", "dead", "stopped"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if _, err = s.CreateTurnWithStatus(ctx, "next", "held", "queued", "preserve", nil); err != nil {
		t.Fatal(err)
	}
	if err = s.StopWebActiveTurn(ctx, "held", "stopped", "stopped"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := New(s)
	defer e.Close()
	if _, err = e.recoverInterruptedTurns(ctx, "held"); err != nil {
		t.Fatal(err)
	}
	stopped, _ := s.GetTurn(ctx, "stopped")
	next, _ := s.GetTurn(ctx, "next")
	if stopped.Status != "aborted" || next.Status != "queued" {
		t.Fatal(stopped, next)
	}
	if _, _, err = s.GetSessionActiveTurn(ctx, "held"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	if hold, err := s.WebQueueHold(ctx, "held"); err != nil || hold != "stopped" {
		t.Fatal(hold, err)
	}
}

func TestWebQueueHoldResumeFailureKeepsDurableFence(t *testing.T) {
	for _, steering := range []bool{false, true} {
		t.Run(fmt.Sprint(steering), func(t *testing.T) {
			ctx := context.Background()
			s := openTestStore(t)
			defer s.Close()
			e := New(s)
			defer e.Close()
			s.CreateSession(ctx, "held", "held", nil)
			s.CreateTurnWithStatus(ctx, "stopped", "held", "cancelled", "old", nil)
			if steering {
				if _, err := s.EnqueueSteering(ctx, "held", "", "user", "pending", map[string]any{"model": "bootstrap"}, nil, ""); err != nil {
					t.Fatal(err)
				}
			} else {
				s.CreateTurnWithStatus(ctx, "next", "held", "queued", "next", map[string]any{"model": "bootstrap"})
			}
			if _, err := s.DB().Exec(`insert into web_queue_holds values('held','stopped',datetime('now'))`); err != nil {
				t.Fatal(err)
			}
			e.beforeLaunchSessionStateErrorHook = func(context.Context, string, string) error { return errors.New("injected launch failure") }
			if _, err := e.ResumeWebQueue(ctx, "held", "stopped"); err == nil {
				t.Fatal("wanted failure")
			}
			if hold, err := s.WebQueueHold(ctx, "held"); err != nil || hold != "stopped" {
				t.Fatal(hold, err)
			}
			if _, _, err := s.GetSessionActiveTurn(ctx, "held"); !errors.Is(err, sql.ErrNoRows) {
				t.Fatal(err)
			}
			if _, err := e.ContinueSession(ctx, "held"); !errors.Is(err, store.ErrQueueConflict) {
				t.Fatal(err)
			}
			e.beforeLaunchSessionStateErrorHook = nil
			if ok, err := e.ResumeWebQueue(ctx, "held", "stopped"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			waitForCondition(t, 3*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "held"); return errors.Is(err, sql.ErrNoRows) }, "resumed cleanup")
		})
	}
}

func TestWebQueueHoldRetirementFailureRollsBackLaunch(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	s.CreateSession(ctx, "held", "held", nil)
	s.CreateTurnWithStatus(ctx, "next", "held", "queued", "next", map[string]any{"model": "bootstrap"})
	if _, err := s.DB().Exec(`insert into web_queue_holds values('held','stopped',datetime('now'));create trigger reject_resume before delete on web_queue_holds begin select raise(abort,'hold retirement fault');end`); err != nil {
		t.Fatal(err)
	}
	if _, err := e.ResumeWebQueue(ctx, "held", "stopped"); err == nil {
		t.Fatal("wanted failure")
	}
	if v, err := s.GetTurn(ctx, "next"); err != nil || v.Status != "queued" {
		t.Fatal(v, err)
	}
	if _, _, err := s.GetSessionActiveTurn(ctx, "held"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	if hold, err := s.WebQueueHold(ctx, "held"); err != nil || hold != "stopped" {
		t.Fatal(hold, err)
	}
	if _, err := s.DB().Exec(`drop trigger reject_resume`); err != nil {
		t.Fatal(err)
	}
	if ok, err := e.ResumeWebQueue(ctx, "held", "stopped"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	waitForCondition(t, 3*time.Second, func() bool { _, _, err := s.GetSessionActiveTurn(ctx, "held"); return errors.Is(err, sql.ErrNoRows) }, "cleanup")
}
