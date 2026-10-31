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
	}{{`Post "https://api.enterprise.githubcopilot.com/responses": http2: timeout awaiting response headers`, true}, {"HTTP 503: service unavailable", true}, {"HTTP 501: not implemented", false}, {"HTTP 505: version unsupported", false}, {"HTTP 429: slow down", true}, {"HTTP 401: timeout in auth", true}, {"HTTP 400: timeout argument invalid", true}, {"invalid_request_error: schema mismatch", false}, {"Unsupported parameter: max_output_tokens", false}, {"context_length exceeded", false}, {"model not found", false}, {"arbitrary provider error", false}, {"connection reset by peer", false}, {"server_busy", true}, {"servers are currently busy", true}, {"Selected model is at capacity", true}, {"HTTP 400 server_busy", true}, {"unsupported model: selected model is at capacity", true}, {"HTTP 524", true}, {"HTTP 429 insufficient_quota", false}}
	for _, tc := range cases {
		if native := goai.IsRetryableAssistantError(&goai.Message{StopReason: goai.StopReasonError, ErrorMessage: tc.text}); native != tc.want {
			t.Fatalf("fixture disagrees with native classifier: %q", tc.text)
		}
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
			}, func(n int, d time.Duration, s string) (time.Time, error) {
				waits++
				if tc.cancel {
					cancel()
				}
				return time.Now().Add(d), nil
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
	ctx = store.WithTurnWriter(ctx, "retry-snapshot", "retry-active", "claim")
	policy := config.ProviderRetryPolicy{Enabled: true, MaxRetries: 3, BaseDelayMS: 60000, MaxDelayMS: 60000}
	var checkpoint *store.ProviderRetryWait
	_, err := retryProviderRequest(ctx, policy, func() (*inference.StreamResult, error, bool) {
		if err := s.AdmitTurnEffect(ctx, "retry-snapshot", "retry-active", "running", "inference.started", map[string]any{"request_id": "attempt"}); err != nil {
			t.Fatal(err)
		}
		checkpoint = newProviderRetryWait(policy, "test/provider", 1, 1, time.Minute, "timeout")
		if _, err := s.SettleProviderAttempt(ctx, "retry-active", "retry-snapshot", "attempt", store.ProviderReceipt{Decision: "error_boundary", Error: "timeout awaiting response headers", Unknown: true, RetryWait: checkpoint, Iteration: 1}); err != nil {
			t.Fatal(err)
		}
		return nil, errors.New("timeout awaiting response headers"), false
	}, func(n int, d time.Duration, class string) (time.Time, error) {
		r.publishProviderRetryWait("retry-active", "retry-snapshot", checkpoint)
		state, err := s.SessionActivity(ctx, "retry-snapshot")
		if err != nil {
			t.Fatal(err)
		}
		retry, ok := state["retry"].(map[string]any)
		if !ok || retry["failure_category"] != "timeout" || !strings.Contains(fmt.Sprint(retry["title"]), "1/3") {
			t.Fatalf("missing reload snapshot %#v", state)
		}
		cancel()
		return checkpoint.RetryAt, nil
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

func waitProviderTerminal(t *testing.T, s *store.Store, id, status string, timeout ...time.Duration) {
	t.Helper()
	limit := 3 * time.Second
	if len(timeout) > 0 {
		limit = timeout[0]
	}
	end := time.Now().Add(limit)
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
	}, func(_ int, d time.Duration, _ string) (time.Time, error) { return time.Now().Add(d), nil }, func() error { return nil })
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
	_, err = retryProviderRequest(context.Background(), p, func() (*inference.StreamResult, error, bool) { return nil, original, false }, func(int, time.Duration, string) (time.Time, error) { return time.Time{}, storage }, nil)
	if !errors.Is(err, original) || !errors.Is(err, storage) {
		t.Fatal("original/callback error lost", err)
	}
}

func TestProviderRetryUsesCommittedDeadlineRatherThanAnotherDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	calls := 0
	result, err := retryProviderRequest(ctx, config.ProviderRetryPolicy{Enabled: true, MaxRetries: 1, BaseDelayMS: 60000, MaxDelayMS: 60000}, func() (*inference.StreamResult, error, bool) {
		calls++
		if calls == 1 {
			return nil, errors.New("HTTP 503"), false
		}
		return &inference.StreamResult{Text: "ok"}, nil, false
	}, func(int, time.Duration, string) (time.Time, error) { return time.Now().Add(-time.Second), nil }, func() error { return nil })
	if err != nil || calls != 2 || result.Text != "ok" {
		t.Fatal("committed deadline ignored", result, err, calls)
	}
}
func TestProviderRetryReturnedPartialWithoutDeltasDoesNotReplay(t *testing.T) {
	for _, kind := range []string{"text", "tool", "partial", "incomplete"} {
		t.Run(kind, func(t *testing.T) {
			s := openChildJoinStore(t)
			zero := 0
			e := NewWithRuntimeConfig(s, config.RuntimeConfig{Retry: config.ProviderRetrySettings{BaseDelayMS: &zero, MaxAgentDelayMS: &zero}}, "")
			defer e.Close()
			var calls atomic.Int32
			withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
				calls.Add(1)
				m := &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonError, ErrorMessage: "HTTP 503", Usage: &goai.Usage{Input: 5, Output: 2, TotalTokens: 7}}
				result := &inference.StreamResult{Message: m, Usage: m.Usage}
				switch kind {
				case "text":
					m.Content = []goai.ContentBlock{{Type: "text", Text: "observed partial"}}
				case "tool":
					m.Content = []goai.ContentBlock{{Type: "toolCall", ID: "must-not-run", Name: "shell", Arguments: map[string]any{"command": "false"}}}
				case "partial":
					result.Message = nil
					result.Partial = m
				case "incomplete":
					result.Incomplete = true
				}
				return result, errors.New("HTTP 503")
			})
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "partial-retry", "partial", nil); err != nil {
				t.Fatal(err)
			}
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "partial-retry", Prompt: "one input", Model: "fixture/model"})
			if err != nil {
				t.Fatal(err)
			}
			waitProviderTerminal(t, s, run.TurnID, "failed")
			if calls.Load() != 1 {
				t.Fatal("returned partial was replayed", calls.Load())
			}
			accounting, err := s.TurnAccounting(ctx, run.TurnID)
			if err != nil || accounting.Usage.TotalTokens != 7 {
				t.Fatal("known partial accounting lost", accounting, err)
			}
			var n int
			if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_scheduled'`, run.TurnID).Scan(&n); err != nil || n != 0 {
				t.Fatal(n, err)
			}
		})
	}
}

func TestProviderRetryExplicitZeroAndOneHundredRetriesUseDistinctReceipts(t *testing.T) {
	s := openChildJoinStore(t)
	zero, retries := 0, 100
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{Retry: config.ProviderRetrySettings{MaxRetries: &retries, BaseDelayMS: &zero, MaxAgentDelayMS: &zero}}, "")
	defer e.Close()
	var calls atomic.Int32
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		calls.Add(1)
		return &inference.StreamResult{Usage: &goai.Usage{Input: 1, TotalTokens: 1}}, errors.New("HTTP 503")
	})
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "hundred", "hundred", nil); err != nil {
		t.Fatal(err)
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "hundred", Prompt: "one input", Model: "fixture/model"})
	if err != nil {
		t.Fatal(err)
	}
	// Race instrumentation makes this 101-attempt stress case much slower
	// than the ordinary three-retry fixtures; keep a bounded, case-local guard.
	waitProviderTerminal(t, s, run.TurnID, "failed", 20*time.Second)
	if calls.Load() != 101 {
		t.Fatal("retry count or explicit zero ignored", calls.Load())
	}
	var scheduled, resumed, receipts int
	if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_scheduled'`, run.TurnID).Scan(&scheduled); err != nil {
		t.Fatal(err)
	}
	if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_resumed'`, run.TurnID).Scan(&resumed); err != nil {
		t.Fatal(err)
	}
	if err := s.DB().QueryRow(`select count(*) from turn_receipts where turn_id=? and kind='provider'`, run.TurnID).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if scheduled != 100 || resumed != 100 || receipts != 101 {
		t.Fatal(scheduled, resumed, receipts)
	}
	accounting, err := s.TurnAccounting(ctx, run.TurnID)
	if err != nil || accounting.Usage.TotalTokens != 101 || accounting.Unknown != 0 {
		t.Fatal(accounting, err)
	}
}

func TestProviderRetryCancellationPublishesCommittedPartialBeforeStopping(t *testing.T) {
	s, e := intentFixture(t)
	ctx := store.WithTurnWriter(context.Background(), "A", "t", "claim")
	if err := s.UpdateTurnStatusAndPhase(ctx, "t", "cancelling", "cancelling"); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	ch := e.Subscribe("A")
	defer e.Unsubscribe("A", ch)
	message := testProviderPartial()
	err := e.runner("A").settleProviderFailure(cancelled, s, "t", "A", "fixture/model", 1, &inference.StreamResult{OccurrenceID: "cancelled-partial", Message: message, Usage: message.Usage, Incomplete: true}, context.Canceled)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	select {
	case ev := <-ch:
		if ev["type"] != "new_post" || ev["content"] != goai.GetTextContent(message) {
			t.Fatal("cancelled partial not published", ev)
		}
	default:
		t.Fatal("committed cancelled partial lost its live post")
	}
	accounting, err := s.TurnAccounting(context.Background(), "t")
	if err != nil || accounting.Usage.Input != 17 {
		t.Fatal(accounting, err)
	}
}
