package turn

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestProviderRetryCloseAfterWakeJoinsActualHookBeforeRequeue(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	delay, retries := 20, 1
	cfg := config.RuntimeConfig{DefaultModel: "test/provider", Retry: config.ProviderRetrySettings{BaseDelayMS: &delay, MaxRetries: &retries}}
	e := NewWithRuntimeConfig(s, cfg, "")
	release, entered, cancelled := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer func() { once.Do(func() { close(release) }); _ = e.Close() }()
	var calls, assemblies, ends atomic.Int32
	withStreamWithToolsHookStub(t, func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("HTTP 503: service unavailable")
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "resumed after close"}}, StopReason: goai.StopReasonStop}, Text: "resumed after close"}, nil
	})
	if _, err := e.RegisterHook(HookContext, "ignore-cancellation", func(ctx context.Context, _ HookRequest) (HookResponse, error) {
		if assemblies.Add(1) == 2 {
			close(entered)
			<-ctx.Done()
			close(cancelled)
			<-release
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "join-retry", "join", nil); err != nil {
		t.Fatal(err)
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "join-retry", Prompt: "once", Model: "test/provider"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("wake did not enter fresh preparation")
	}
	closed := make(chan error, 2)
	go func() { closed <- e.Close() }()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("hook not cancelled")
	}
	go func() { closed <- e.Close() }()
	select {
	case err := <-closed:
		t.Fatal("close released an actual owned callback", err)
	case <-time.After(30 * time.Millisecond):
	}
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "running" || row.Phase != "running" {
		t.Fatal("checkpoint published before hook join", row, err)
	}
	if ends.Load() != 0 || calls.Load() != 1 {
		t.Fatal("shutdown invoked end/provider", ends.Load(), calls.Load())
	}
	// Another engine sharing the Store must not mistake this live reservation
	// for a crash, nor abort it when that other engine closes.
	observer := NewWithRuntimeConfig(s, cfg, "")
	if err = observer.ResumeInterruptedWork(ctx); err != nil {
		t.Fatal(err)
	}
	if err = observer.Close(); err != nil {
		t.Fatal(err)
	}
	once.Do(func() { close(release) })
	for range 2 {
		select {
		case err = <-closed:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("close did not join hook")
		}
	}
	row, err = s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "queued" || row.Phase != store.ClosedProviderRetryReadyPhase {
		t.Fatal(row, err)
	}
	next := NewWithRuntimeConfig(s, cfg, "")
	defer next.Close()
	if _, err = next.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
		t.Fatal(err)
	}
	if err = next.ResumeInterruptedWork(ctx); err != nil {
		t.Fatal(err)
	}
	waitRetryDone(t, s, run.TurnID)
	next.runs.Wait() // inspect lifecycle only after its actual host has joined
	if calls.Load() != 2 || ends.Load() != 1 {
		t.Fatal("provider/end continuation duplicated", calls.Load(), ends.Load())
	}
	var wakes int
	if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_resumed'`, run.TurnID).Scan(&wakes); err != nil || wakes != 1 {
		t.Fatal("wake replayed", wakes, err)
	}
}
