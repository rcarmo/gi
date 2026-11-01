package turn

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

var providerStatusCode = regexp.MustCompile(`(?i)(?:http(?: status)?[ :]+|status(?: code)?[ :=]+)([45][0-9][0-9])\b`)

// Classify at the provider boundary only, never assistant prose. Automatic
// retries below are restricted further to attempts with no streamed progress.
func transientProviderFailure(err error) (string, bool) {
	if err == nil || errors.Is(err, context.Canceled) {
		return "", false
	}
	var abort hookAbortError
	if errors.As(err, &abort) {
		return "", false
	}
	text := err.Error()
	if errors.Is(err, context.DeadlineExceeded) {
		text = "provider request timeout: " + text
	}
	if !goai.IsRetryableAssistantError(&goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonError, ErrorMessage: text}) {
		return "", false
	}
	// Categories are presentation only; the provider-boundary diagnostic above
	// is classified by go-ai, never by assistant content or a local status list.
	text = strings.ToLower(text)
	if code := providerStatusCode.FindStringSubmatch(text); len(code) > 0 && code[1] == "429" {
		return "rate_limit", true
	}
	if strings.Contains(text, "rate") || strings.Contains(text, "too many requests") {
		return "rate_limit", true
	}
	if strings.Contains(text, "timeout") || strings.Contains(text, "timed out") {
		return "timeout", true
	}
	return "network", true
}

func providerRetryDelay(p config.ProviderRetryPolicy, attempt int) time.Duration {
	ms := int64(p.BaseDelayMS)
	for i := 1; i < attempt && ms < int64(p.MaxDelayMS); i++ {
		ms *= 2
	}
	if ms > int64(p.MaxDelayMS) {
		ms = int64(p.MaxDelayMS)
	}
	return time.Duration(ms) * time.Millisecond
}

// request repeats only the current inference request; executed tools and prior
// iterations are never replayed. Progress or hook failure disables blind retry.
func retryProviderRequest(ctx context.Context, p config.ProviderRetryPolicy, request func() (result *inference.StreamResult, err error, progress bool), waiting func(int, time.Duration, string) (time.Time, error), resuming func() error) (*inference.StreamResult, error) {
	return retryProviderRequestDynamic(ctx, func() config.ProviderRetryPolicy { return p }, request, waiting, resuming)
}

func retryProviderRequestDynamic(ctx context.Context, policy func() config.ProviderRetryPolicy, request func() (result *inference.StreamResult, err error, progress bool), waiting func(int, time.Duration, string) (time.Time, error), resuming func() error) (*inference.StreamResult, error) {
	return retryProviderRequestDynamicFrom(ctx, policy, 0, request, waiting, resuming)
}

func retryProviderRequestDynamicFrom(ctx context.Context, policy func() config.ProviderRetryPolicy, firstAttempt int, request func() (result *inference.StreamResult, err error, progress bool), waiting func(int, time.Duration, string) (time.Time, error), resuming func() error) (*inference.StreamResult, error) {
	if firstAttempt < 0 || firstAttempt > 100 {
		return nil, fmt.Errorf("invalid recovered retry attempt")
	}
	for attempt := firstAttempt; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result, err, progress := request()
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		// A transport's attempt-scoped deadline is not cancellation of the
		// owning turn. Avoid the outer turn handler misclassifying exhaustion.
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("provider request timeout: %v", err)
		}
		if err == nil {
			return result, nil
		}
		class, retry := transientProviderFailure(err)
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		p := policy()
		if !p.Enabled || !retry || progress || attempt >= p.MaxRetries {
			return result, err
		}
		delay := providerRetryDelay(p, attempt+1)
		deadline, scheduleErr := waiting(attempt+1, delay, class)
		if scheduleErr != nil {
			return result, errors.Join(err, fmt.Errorf("retry scheduling: %w", scheduleErr))
		}
		timer := time.NewTimer(max(0, time.Until(deadline)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return result, ctx.Err()
		case <-timer.C:
		}
		if resumeErr := resuming(); resumeErr != nil {
			return result, errors.Join(err, fmt.Errorf("retry resumption: %w", resumeErr))
		}
	}
}

func newProviderRetryWait(p config.ProviderRetryPolicy, model string, iter, attempt int, delay time.Duration, class string) *store.ProviderRetryWait {
	title := fmt.Sprintf("Provider %s — retrying (attempt %d/%d, %gs delay)", map[string]string{"timeout": "request timed out", "network": "connection failed", "rate_limit": "rate limited"}[class], attempt, p.MaxRetries, delay.Seconds())
	return &store.ProviderRetryWait{Attempt: attempt, MaxAttempts: p.MaxRetries, DelayMS: delay.Milliseconds(), RetryAt: time.Now().Add(delay).UTC(), FailureCategory: class, Iteration: iter, Model: model, Title: title}
}
func (r *sessionRunner) publishProviderRetryWait(turnID, sessionID string, p *store.ProviderRetryWait) {
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "status": "running", "phase": "retry_wait", "title": p.Title, "attempt": p.Attempt, "max_attempts": p.MaxAttempts, "retry_at": p.RetryAt.UTC().Format(time.RFC3339Nano), "failure_category": p.FailureCategory})
}
