package turn

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestTransientProviderFailureClassifier(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{{`Post "https://api.enterprise.githubcopilot.com/responses": http2: timeout awaiting response headers`, true}, {"HTTP 503: service unavailable", true}, {"HTTP 501: not implemented", false}, {"HTTP 505: version unsupported", false}, {"HTTP 429: slow down", true}, {"HTTP 401: timeout in auth", false}, {"HTTP 400: timeout argument invalid", false}, {"invalid_request_error: schema mismatch", false}, {"Unsupported parameter: max_output_tokens", false}, {"context_length exceeded", false}, {"model not found", false}, {"arbitrary provider error", false}, {"connection reset by peer", true}, {"server_busy", true}, {"servers are currently busy", true}, {"Selected model is at capacity", true}, {"HTTP 400 server_busy", false}, {"unsupported model: selected model is at capacity", false}}
	for _, tc := range cases {
		if _, got := transientProviderFailure(errors.New(tc.text)); got != tc.want {
			t.Errorf("%q retry=%v", tc.text, got)
		}
	}
	for _, err := range []error{context.Canceled, hookAbortError{reason: "timeout"}} {
		if _, ok := transientProviderFailure(err); ok {
			t.Fatalf("retry unsafe error %v", err)
		}
	}
	p := config.ProviderRetrySettings{}.Policy()
	for i, want := range []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second, 60 * time.Second} {
		if got := providerRetryDelay(p, i+1); got != want {
			t.Fatal(i, got, want)
		}
	}
}

func TestProviderRetryBoundsProgressCancellationAndDisabled(t *testing.T) {
	for _, tc := range []struct {
		name             string
		policy           config.ProviderRetryPolicy
		progress, cancel bool
		want             int
	}{{"exhausted", config.ProviderRetryPolicy{Enabled: true, MaxRetries: 3}, false, false, 4}, {"partial", config.ProviderRetryPolicy{Enabled: true, MaxRetries: 3}, true, false, 1}, {"cancel", config.ProviderRetryPolicy{Enabled: true, MaxRetries: 3, BaseDelayMS: 1000, MaxDelayMS: 1000}, false, true, 1}, {"disabled", config.ProviderRetryPolicy{}, false, false, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls, waits := 0, 0
			_, err := retryProviderRequest(ctx, tc.policy, func() (*inference.StreamResult, error, bool) {
				calls++
				return nil, errors.New("timeout awaiting response headers"), tc.progress
			}, func(n int, d time.Duration, s string) error {
				waits++
				if tc.cancel {
					cancel()
				}
				return nil
			}, func() error { return nil })
			if err == nil || calls != tc.want {
				t.Fatalf("calls=%d waits=%d err=%v", calls, waits, err)
			}
			if tc.cancel && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		})
	}
}

func TestProviderHeaderTimeoutRetriesWithinSameTurnAndPersistsStatus(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	zero, retries := 1, 3
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{DefaultModel: "test/provider", Retry: config.ProviderRetrySettings{BaseDelayMS: &zero, MaxRetries: &retries}}, "")
	defer e.Close()
	calls := 0
	withStreamWithToolsHookStub(t, func(ctx context.Context, model string, c *goai.Context, broadcast func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		calls++
		if calls == 1 {
			broadcast(map[string]any{"type": "error", "error": "http2: timeout awaiting response headers"})
			return nil, errors.New("http2: timeout awaiting response headers")
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "recovered"}}, StopReason: goai.StopReasonStop}, Text: "recovered"}, nil
	})
	if _, err := s.CreateSession(context.Background(), "retry-session", "retry", nil); err != nil {
		t.Fatal(err)
	}
	events := e.Subscribe("retry-session")
	defer e.Unsubscribe("retry-session", events)
	result, err := e.SubmitPrompt(context.Background(), RunInput{SessionID: "retry-session", Prompt: "original only once", Model: "test/provider"})
	if err != nil {
		t.Fatal(err)
	}
	waitRetryDone(t, s, result.TurnID)
	tr, err := s.GetTurn(context.Background(), result.TurnID)
	if err != nil || tr.Status != "completed" {
		t.Fatal(tr, err)
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
	var scheduled int
	if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_scheduled'`, result.TurnID).Scan(&scheduled); err != nil || scheduled != 1 {
		t.Fatal(scheduled, err)
	}
	msgs, err := s.ListMessages(context.Background(), "retry-session")
	if err != nil {
		t.Fatal(err)
	}
	users, errorsSeen := 0, 0
	for _, m := range msgs {
		if m.Role == "user" {
			users++
		}
		if strings.Contains(m.Content, "Inference error") {
			errorsSeen++
		}
	}
	if users != 1 || errorsSeen != 0 {
		t.Fatalf("users=%d errors=%d", users, errorsSeen)
	}
	var retryTitle bool
	for {
		select {
		case ev := <-events:
			if ev["type"] == "error" {
				t.Fatal("terminal error during recovery", ev)
			}
			if ev["phase"] == "retry_wait" {
				retryTitle = true
			}
		default:
			if !retryTitle {
				t.Fatal("no retry status")
			}
			return
		}
	}
}

func TestProviderRetryActivitySnapshotAndCancelWait(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e := New(s)
	defer e.Close()
	if _, err := s.CreateSession(ctx, "retry-snapshot", "snapshot", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTurnWithStatus(ctx, "retry-active", "retry-snapshot", "running", "prompt", nil); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.ClaimSessionActiveTurn(ctx, "retry-snapshot", "retry-active", "fixture", "claim"); err != nil || !ok {
		t.Fatal(ok, err)
	}
	r := &sessionRunner{engine: e}
	_, err := retryProviderRequest(ctx, config.ProviderRetryPolicy{Enabled: true, MaxRetries: 3, BaseDelayMS: 60000, MaxDelayMS: 60000}, func() (*inference.StreamResult, error, bool) {
		return nil, errors.New("timeout awaiting response headers"), false
	}, func(n int, d time.Duration, class string) error {
		if err := r.providerRetryWait(ctx, s, "retry-active", "retry-snapshot", "test/provider", 1, n, d, class); err != nil {
			return err
		}
		state, err := s.SessionActivity(ctx, "retry-snapshot")
		if err != nil {
			t.Fatal(err)
		}
		retry, ok := state["retry"].(map[string]any)
		if !ok || retry["failure_category"] != "timeout" || !strings.Contains(fmt.Sprint(retry["title"]), "1/3") {
			t.Fatalf("missing reload snapshot %#v", state)
		}
		cancel()
		return nil
	}, func() error { t.Fatal("cancelled retry resumed"); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.UpdateTurnStatus(context.Background(), "retry-active", "cancelled"); err != nil {
		t.Fatal(err)
	}
	state, err := s.SessionActivity(context.Background(), "retry-snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if state["retry"] != nil {
		t.Fatal("terminal retry resurrected", state)
	}
}

func TestProviderRetryExhaustionEmitsOneDurableFailure(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	zero, max := 1, 3
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{Retry: config.ProviderRetrySettings{BaseDelayMS: &zero, MaxRetries: &max}}, "")
	defer e.Close()
	calls := 0
	withStreamWithToolsHookStub(t, func(ctx context.Context, _ string, _ *goai.Context, broadcast func(map[string]any), _ *inference.StreamHooks) (*inference.StreamResult, error) {
		calls++
		broadcast(map[string]any{"type": "error", "error": "http2: timeout awaiting response headers"})
		return nil, errors.New("http2: timeout awaiting response headers")
	})
	if _, err := s.CreateSession(context.Background(), "exhaust", "exhaust", nil); err != nil {
		t.Fatal(err)
	}
	ch := e.Subscribe("exhaust")
	defer e.Unsubscribe("exhaust", ch)
	result, err := e.SubmitPrompt(context.Background(), RunInput{SessionID: "exhaust", Prompt: "one input", Model: "test/provider"})
	if err != nil {
		t.Fatal(err)
	}
	waitProviderTerminal(t, s, result.TurnID, "failed")
	if calls != 4 {
		t.Fatal(calls)
	}
	var retries int
	if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_scheduled'`, result.TurnID).Scan(&retries); err != nil || retries != 3 {
		t.Fatal(retries, err)
	}
	// FinalizeRunningTurn commits failed status before the terminal system
	// message. Wait for that record rather than treating status as a barrier.
	waitForCondition(t, 3*time.Second, func() bool {
		msgs, err := s.ListMessages(context.Background(), "exhaust")
		if err != nil {
			return false
		}
		for _, m := range msgs {
			if m.Role == "system" && strings.Contains(m.Content, "Inference error") {
				return true
			}
		}
		return false
	}, "provider retry terminal system message")
	msgs, err := s.ListMessages(context.Background(), "exhaust")
	if err != nil {
		t.Fatal(err)
	}
	systems, users := 0, 0
	for _, m := range msgs {
		if m.Role == "system" && strings.Contains(m.Content, "Inference error") {
			systems++
		}
		if m.Role == "user" {
			users++
		}
	}
	if systems != 1 || users != 1 {
		t.Fatalf("systems=%d users=%d", systems, users)
	}
	for {
		select {
		case ev := <-ch:
			if ev["type"] == "error" {
				t.Fatal("duplicate transient error broadcast", ev)
			}
		default:
			return
		}
	}
}

func waitProviderTerminal(t *testing.T, s *store.Store, id, status string) {
	t.Helper()
	end := time.Now().Add(3 * time.Second)
	for time.Now().Before(end) {
		tr, err := s.GetTurn(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if tr.Status == status {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	tr, _ := s.GetTurn(context.Background(), id)
	t.Fatalf("expected %s got %#v", status, tr)
}

func TestProviderRetryCancelStopsRealTurnWithoutAnotherRequest(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	delay := 60000
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{Retry: config.ProviderRetrySettings{BaseDelayMS: &delay}}, "")
	defer e.Close()
	var calls atomic.Int32
	withStreamWithToolsHookStub(t, func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
		calls.Add(1)
		return nil, errors.New("timeout awaiting response headers")
	})
	if _, err := s.CreateSession(context.Background(), "cancel-wait", "cancel", nil); err != nil {
		t.Fatal(err)
	}
	result, err := e.SubmitPrompt(context.Background(), RunInput{SessionID: "cancel-wait", Prompt: "cancel this", Model: "test/provider"})
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now().Add(2 * time.Second)
	waiting := false
	for time.Now().Before(end) {
		state, err := s.SessionActivity(context.Background(), "cancel-wait")
		if err == nil && state["retry"] != nil {
			waiting = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !waiting {
		t.Fatal("no retry wait")
	}
	if err := e.CancelTurn(context.Background(), "cancel-wait", result.TurnID); err != nil {
		t.Fatal(err)
	}
	waitProviderTerminal(t, s, result.TurnID, "cancelled")
	if calls.Load() != 1 {
		t.Fatal("cancel replayed request", calls.Load())
	}
}

func TestProviderRetryAttemptDeadlineVersusParentCancellation(t *testing.T) {
	calls := 0
	p := config.ProviderRetryPolicy{Enabled: true, MaxRetries: 1}
	result, err := retryProviderRequest(context.Background(), p, func() (*inference.StreamResult, error, bool) {
		calls++
		if calls == 1 {
			return nil, context.DeadlineExceeded, false
		}
		return &inference.StreamResult{Text: "ok"}, nil, false
	}, func(int, time.Duration, string) error { return nil }, func() error { return nil })
	if err != nil || calls != 2 || result.Text != "ok" {
		t.Fatal(result, err, calls)
	}
	_, err = retryProviderRequest(context.Background(), config.ProviderRetryPolicy{}, func() (*inference.StreamResult, error, bool) { return nil, context.DeadlineExceeded, false }, nil, nil)
	if err == nil || isCancellationError(err) {
		t.Fatal("attempt timeout treated as cancelled turn", err)
	}
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	calls = 0
	_, err = retryProviderRequest(parent, p, func() (*inference.StreamResult, error, bool) { calls++; return nil, nil, false }, nil, nil)
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatal(err, calls)
	}
	original := errors.New("timeout awaiting response headers")
	storage := errors.New("lost ownership")
	_, err = retryProviderRequest(context.Background(), p, func() (*inference.StreamResult, error, bool) { return nil, original, false }, func(int, time.Duration, string) error { return storage }, nil)
	if !errors.Is(err, original) || !errors.Is(err, storage) {
		t.Fatal("original/callback error lost", err)
	}
}
