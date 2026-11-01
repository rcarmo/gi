package turn

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestProviderRetryCloseQueuesCheckpointWithoutAbortOrLifecycleReplay(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	delay, retries := 500, 1
	cfg := config.RuntimeConfig{DefaultModel: "test/provider", Retry: config.ProviderRetrySettings{BaseDelayMS: &delay, MaxRetries: &retries}}
	e := NewWithRuntimeConfig(s, cfg, "")
	defer e.Close()
	var calls, starts, ends atomic.Int32
	withStreamWithToolsHookStub(t, func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("HTTP 503: service unavailable")
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "continued"}}, StopReason: goai.StopReasonStop}, Text: "continued"}, nil
	})
	if _, err := e.RegisterHook(HookAgentStart, "start", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := e.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "close-retry", "close", nil); err != nil {
		t.Fatal(err)
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "close-retry", Prompt: "one prompt", Model: "test/provider"})
	if err != nil {
		t.Fatal(err)
	}
	until := time.Now().Add(2 * time.Second)
	for {
		row, err := s.GetTurn(ctx, run.TurnID)
		if err != nil {
			t.Fatal(err)
		}
		if row.Phase == "retry_wait" {
			break
		}
		if time.Now().After(until) {
			t.Fatal("retry did not sleep", row)
		}
		time.Sleep(time.Millisecond)
	}
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "queued" || !store.IsClosedProviderRetryPhase(row.Phase) || row.FinishedAt != "" {
		t.Fatal("graceful close fabricated terminal outcome", row, err)
	}
	if calls.Load() != 1 || starts.Load() != 1 || ends.Load() != 0 {
		t.Fatal("shutdown ran lifecycle or provider", calls.Load(), starts.Load(), ends.Load())
	}
	resumed := NewWithRuntimeConfig(s, cfg, "")
	defer resumed.Close()
	if calls.Load() != 1 {
		t.Fatal("constructor dispatched checkpoint")
	}
	if _, err = resumed.RegisterHook(HookAgentStart, "start", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	if _, err = resumed.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	if err = resumed.ResumeInterruptedWork(ctx); err != nil {
		t.Fatal(err)
	}
	waitRetryDone(t, s, run.TurnID)
	resumed.runs.Wait() // terminal persistence precedes the once-only end hook
	row, err = s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "completed" {
		t.Fatal(row, err)
	}
	if calls.Load() != 2 || starts.Load() != 1 || ends.Load() != 1 {
		t.Fatal("continuation replayed lifecycle", calls.Load(), starts.Load(), ends.Load())
	}
	var prompts, wakes, failures int
	if err = s.DB().QueryRow(`select count(*) from messages where session_id='close-retry' and role='user'`).Scan(&prompts); err != nil {
		t.Fatal(err)
	}
	if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_resumed'`, run.TurnID).Scan(&wakes); err != nil {
		t.Fatal(err)
	}
	if err = s.DB().QueryRow(`select count(*) from turn_failures where turn_id=?`, run.TurnID).Scan(&failures); err != nil {
		t.Fatal(err)
	}
	if prompts != 1 || wakes != 1 || failures != 0 {
		t.Fatal("close changed prompt/wake/failure state", prompts, wakes, failures)
	}
}
