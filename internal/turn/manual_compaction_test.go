package turn

import (
	"context"
	"database/sql"
	"errors"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
	"strings"
	"testing"
	"time"
)

func TestManualCompactionIdleTokenNoProviderAndCancel(t *testing.T) {
	for _, mode := range []string{"complete", "cancel", "conflict"} {
		t.Run(mode, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			e := New(s)
			defer e.Close()
			ctx := context.Background()
			s.CreateSession(ctx, "A", "A", nil)
			for i := 0; i < 4; i++ {
				if err := s.AddMessage(ctx, store.NowID("msg"), "A", "user", strings.Repeat("original history ", 20), nil); err != nil {
					t.Fatal(err)
				}
			}
			state, err := e.ManualCompactionState(ctx, "A")
			if err != nil || state["available"] != true {
				t.Fatal(state, err)
			}
			entered := make(chan struct{}, 1)
			release := make(chan struct{})
			defer close(release)
			e.RegisterHook(HookSessionBeforeCompact, "manual-gate", func(ctx context.Context, req HookRequest) (HookResponse, error) {
				if req.Payload["reason"] != "manual" {
					t.Error(req.Payload)
				}
				entered <- struct{}{}
				select {
				case <-release:
				case <-ctx.Done():
					return HookResponse{}, ctx.Err()
				}
				return HookResponse{Payload: map[string]any{"summary": "manual summary"}}, nil
			})
			withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
				t.Fatal("manual compaction made a provider call")
				return nil, nil
			})
			if _, err = e.SubmitManualCompaction(ctx, "A", "stale"); !errors.Is(err, store.ErrContextChanged) {
				t.Fatal(err)
			}
			res, err := e.SubmitManualCompaction(ctx, "A", state["token"].(string))
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				t.Fatal("no manual hook")
			}
			if _, err = e.SubmitManualCompaction(ctx, "A", state["token"].(string)); !errors.Is(err, store.ErrQueueConflict) {
				t.Fatal(err)
			}
			if _, err = e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "must not steer maintenance"}); !errors.Is(err, store.ErrQueueConflict) {
				t.Fatal(err)
			}
			want := "completed"
			if mode == "cancel" {
				want = "cancelled"
				if err = e.CancelActiveTurn(ctx, "A", res.TurnID); err != nil {
					t.Fatal(err)
				}
			} else {
				if mode == "conflict" {
					want = "failed"
					s.AddMessage(ctx, "new", "A", "user", "race", nil)
				}
				release <- struct{}{}
			}
			waitForCondition(t, 3*time.Second, func() bool { rec, err := s.GetTurn(ctx, res.TurnID); return err == nil && rec.Status == want }, "manual finish")
			waitForCondition(t, 3*time.Second, func() bool { r := e.runner("A"); r.mu.Lock(); defer r.mu.Unlock(); return r.current == nil }, "cleanup")
			snapshot, err := s.ContextSnapshot(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			if mode == "complete" && snapshot.Summary != "manual summary" || mode != "complete" && snapshot.Summary != "" {
				t.Fatal(snapshot)
			}
			if mode == "complete" {
				if _, err = e.SubmitManualCompaction(ctx, "A", state["token"].(string)); !errors.Is(err, store.ErrContextChanged) {
					t.Fatal(err)
				}
			}
			msgs, _ := s.ListMessages(ctx, "A")
			for _, m := range msgs {
				if m.Content == "/compact" {
					t.Fatal("manual command persisted as prompt")
				}
			}
		})
	}
}

func TestManualCompactionRecoveryDoesNotReplay(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	s.CreateSession(ctx, "A", "A", nil)
	e := New(s)
	defer e.Close()
	s.AddMessage(ctx, "one", "A", "user", "one", nil)
	s.AddMessage(ctx, "two", "A", "assistant", "two", nil)
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.AdmitManualCompaction(ctx, "A", "manual", store.ContextToken(snapshot), "test-model"); err != nil {
		t.Fatal(err)
	}
	if next, err := s.GetNextQueuedTurn(ctx, "A"); err != sql.ErrNoRows || next != nil {
		t.Fatal("unclaimed maintenance must not auto-run", next, err)
	}
	s.ClaimSessionActiveTurn(ctx, "A", "manual", "runner", "manual")
	s.UpdateTurnStatusAndPhase(ctx, "manual", "running", "compacting")
	if err = e.recoverInterruptedTurn(ctx, store.ActiveTurnClaim{SessionID: "A", TurnID: "manual", ClaimToken: "manual", Status: "running", Phase: "compacting"}); err != nil {
		t.Fatal(err)
	}
	rec, err := s.GetTurn(ctx, "manual")
	if err != nil || rec.Status != "aborted" {
		t.Fatal(rec, err)
	}
}
