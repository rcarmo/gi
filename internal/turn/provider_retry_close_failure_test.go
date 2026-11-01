package turn

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	goai "github.com/rcarmo/go-ai"
)

func TestProviderRetryClosePreservesAbortAndReportsPersistenceFailure(t *testing.T) {
	for _, mode := range []string{"abort", "reserve", "commit"} {
		t.Run(mode, func(t *testing.T) {
			s := openTestStore(t)
			defer s.Close()
			delay, retries := 60000, 1
			e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/provider", Retry: config.ProviderRetrySettings{BaseDelayMS: &delay, MaxRetries: &retries}}, "")
			defer e.Close()
			var ends atomic.Int32
			withStreamWithToolsHookStub(t, func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
				return nil, errors.New("HTTP 503: service unavailable")
			})
			if _, err := e.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "failure-close", "close", nil); err != nil {
				t.Fatal(err)
			}
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "failure-close", Prompt: "once", Model: "test/provider"})
			if err != nil {
				t.Fatal(err)
			}
			until := time.Now().Add(time.Second)
			for {
				row, err := s.GetTurn(ctx, run.TurnID)
				if err != nil {
					t.Fatal(err)
				}
				if row.Phase == "retry_wait" {
					break
				}
				if time.Now().After(until) {
					t.Fatal("no wait", row)
				}
				time.Sleep(time.Millisecond)
			}
			switch mode {
			case "abort":
				if err = e.CancelActiveTurn(ctx, "failure-close", run.TurnID); err != nil {
					t.Fatal(err)
				}
			case "reserve":
				_, err = s.DB().Exec(`create trigger reject_close before update on session_active_turns when new.worker_id='retry-closing' begin select raise(abort,'reserve rejected');end`)
			case "commit":
				_, err = s.DB().Exec(`create trigger reject_close before insert on turn_events when new.event_type='turn.checkpoint_queued' begin select raise(abort,'commit rejected');end`)
			}
			if err != nil {
				t.Fatal(err)
			}
			err = e.Close()
			row, readErr := s.GetTurn(ctx, run.TurnID)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if mode == "abort" {
				if err != nil || row.Status != "cancelled" || ends.Load() != 1 {
					t.Fatal("user abort became resumable", row, ends.Load(), err)
				}
			} else {
				if err == nil || row.Status != "running" || row.Phase != "retry_wait" || ends.Load() != 0 {
					t.Fatal("persistence failure became success/abort", row, ends.Load(), err)
				}
				if second := e.Close(); second != err {
					t.Fatal("concurrent/idempotent Close lost its error", err, second)
				}
			}
			var queued int
			if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='turn.checkpoint_queued'`, run.TurnID).Scan(&queued); err != nil || queued != 0 {
				t.Fatal("abort/failed close queued checkpoint", queued, err)
			}
		})
	}
}
