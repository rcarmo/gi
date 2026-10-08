package turn

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
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
	text := strings.ToLower(err.Error())
	for _, s := range []string{"unauthorized", "invalid api key", "invalid_request_error", "unsupported parameter", "unsupported model", "model not found", "context length", "context_length", "prompt too long", "schema", "permission denied", "request was aborted"} {
		if strings.Contains(text, s) {
			return "", false
		}
	}
	if code := providerStatusCode.FindStringSubmatch(text); len(code) > 0 {
		if code[1] == "429" {
			return "rate_limit", true
		}
		switch code[1] {
		case "500", "502", "503", "504", "529":
			return "network", true
		default:
			return "", false
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout", true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return "timeout", true
	}
	if strings.Contains(text, "timeout") || strings.Contains(text, "timed out") {
		return "timeout", true
	}
	for _, s := range []string{"econnreset", "econnrefused", "eai_again", "enotfound", "socket hang up", "connection reset", "connection refused", "connection closed", "connection lost", "fetch failed", "unexpected eof", "service unavailable", "bad gateway", "overloaded", "server_busy", "servers are currently busy", "selected model is at capacity"} {
		if strings.Contains(text, s) {
			return "network", true
		}
	}
	if strings.Contains(text, "rate limit") || strings.Contains(text, "too many requests") {
		return "rate_limit", true
	}
	return "", false
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
func retryProviderRequest(ctx context.Context, p config.ProviderRetryPolicy, request func() (result *inference.StreamResult, err error, progress bool), waiting func(int, time.Duration, string) error, resuming func() error) (*inference.StreamResult, error) {
	for attempt := 0; ; attempt++ {
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
		if !p.Enabled || !retry || progress || attempt >= p.MaxRetries {
			return result, err
		}
		delay := providerRetryDelay(p, attempt+1)
		if scheduleErr := waiting(attempt+1, delay, class); scheduleErr != nil {
			return result, errors.Join(err, fmt.Errorf("retry scheduling: %w", scheduleErr))
		}
		timer := time.NewTimer(delay)
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

func (r *sessionRunner) providerRetryWait(ctx context.Context, s *store.Store, turnID, sessionID, model string, iter, attempt int, delay time.Duration, class string) error {
	if err := s.SetClaimedRunningPhase(ctx, sessionID, turnID, "retry_wait"); err != nil {
		return err
	}
	p := r.engine.runtimeCfg.Retry.Policy()
	title := fmt.Sprintf("Provider %s — retrying (attempt %d/%d, %gs delay)", map[string]string{"timeout": "request timed out", "network": "connection failed", "rate_limit": "rate limited"}[class], attempt, p.MaxRetries, delay.Seconds())
	payload := map[string]any{"title": title, "attempt": attempt, "max_attempts": p.MaxRetries, "delay_ms": delay.Milliseconds(), "retry_at": time.Now().Add(delay).UTC().Format(time.RFC3339Nano), "failure_category": class, "iteration": iter, "model": model, "phase": "retry_wait"}
	if err := s.AppendTurnEvent(ctx, turnID, sessionID, "inference.retry_scheduled", payload); err != nil {
		return err
	}
	r.engine.broadcast(sessionID, map[string]any{"type": "agent_status", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "status": "running", "phase": "retry_wait", "title": title, "attempt": attempt, "max_attempts": p.MaxRetries, "retry_at": payload["retry_at"], "failure_category": class})
	return nil
}
