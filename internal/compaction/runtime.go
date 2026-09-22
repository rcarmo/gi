package compaction

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/config"
	goai "github.com/rcarmo/go-ai"
)

type HookDecision struct {
	Cancel  bool
	Block   bool
	Payload map[string]any
}

type RuntimeRequest struct {
	SessionID string
	TurnID    string
	AgentID   string
	Model     string
	Settings  config.CompactionSettings
	Force     bool
}

type RuntimeOps struct {
	BackgroundContext func() context.Context
	BeforeCompact     func(context.Context, map[string]any, []goai.Message) (HookDecision, error)
	AfterCompact      func(context.Context, map[string]any)
	Begin             func(context.Context, string, string, map[string]any) (int, error)
	Finish            func(context.Context, string, string, int, string, string, map[string]any) (string, error)
	Broadcast         func(string, map[string]any)
}

func MaybeCompactContext(ctx context.Context, req RuntimeRequest, convCtx *goai.Context, ops RuntimeOps) error {
	settings := req.Settings
	if err := ctx.Err(); err != nil {
		return err
	}
	if (!req.Force && !settings.Enabled) || len(convCtx.Messages) < 2 || (!req.Force && len(convCtx.Messages) < 6) {
		return nil
	}
	tokens := EstimateMessagesTokens(convCtx.Messages)
	if !req.Force && tokens <= settings.ThresholdTokens {
		return nil
	}
	prep := Prepare(convCtx.Messages, tokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
	if prep.MessagesToSummarize <= 0 {
		return nil
	}
	reason := "threshold"
	if req.Force {
		reason = "manual"
	}
	payload := map[string]any{"phase": "compacting", "checkpoint": true, "reason": reason, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "tokens_source": "estimate"}
	startedSeq := 0
	if ops.Begin != nil {
		var err error
		startedSeq, err = ops.Begin(ctx, req.SessionID, req.TurnID, payload)
		if err != nil {
			return err
		}
	}
	payload["started_seq"] = startedSeq
	publish := func(outcome string, fields map[string]any) {
		if ops.Broadcast == nil {
			return
		}
		ev := map[string]any{"type": "compaction_" + outcome, "chat_jid": "gi:" + req.SessionID, "turn_id": req.TurnID, "outcome": outcome}
		for k, v := range fields {
			ev[k] = v
		}
		ops.Broadcast(req.SessionID, ev)
	}
	publish("started", payload)
	durableCtx := ctx
	if ops.BackgroundContext != nil {
		durableCtx = ops.BackgroundContext()
	}
	finish := func(outcome, summary string, fields map[string]any) (string, error) {
		accepted := outcome
		var err error
		if ops.Finish != nil {
			accepted, err = ops.Finish(durableCtx, req.SessionID, req.TurnID, startedSeq, outcome, summary, fields)
		}
		if err != nil {
			return "", err
		}
		published := map[string]any{}
		for k, v := range fields {
			published[k] = v
		}
		published["started_seq"] = startedSeq
		if accepted != "completed" {
			delete(published, "messages_after")
			delete(published, "from_hook")
		}
		publish(accepted, published)
		return accepted, nil
	}
	hookPayload := map[string]any{"reason": reason, "preparation": prep, "settings": map[string]any{"enabled": settings.Enabled, "context_window": settings.ContextWindow, "reserve_tokens": settings.ReserveTokens, "keep_recent_tokens": settings.KeepRecentTokens, "threshold_tokens": settings.ThresholdTokens, "strategy": settings.Strategy}}
	decision := HookDecision{}
	var hookErr error
	if ops.BeforeCompact != nil {
		decision, hookErr = ops.BeforeCompact(ctx, hookPayload, convCtx.Messages)
	}
	outcome := "completed"
	if ctx.Err() != nil {
		outcome = "cancelled"
		payload["detail"] = "Turn cancellation requested"
	} else if hookErr != nil {
		outcome = "failed"
		payload["detail"] = hookErr.Error()
	} else if decision.Cancel || decision.Block {
		outcome = "suppressed"
		payload["detail"] = "Compaction temporarily suppressed by before-compact hook"
	}
	if outcome != "completed" {
		if _, err := finish(outcome, "", payload); err != nil {
			return err
		}
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		_, finishErr := finish("cancelled", "", payload)
		if finishErr != nil {
			return finishErr
		}
		return err
	}
	summary := ""
	if decision.Payload != nil {
		summary, _ = decision.Payload["summary"].(string)
		summary = strings.TrimSpace(summary)
	}
	if summary == "" {
		summary = DefaultSummary(prep)
	}
	if strings.TrimSpace(summary) == "" {
		payload["detail"] = "Compaction produced an empty summary"
		_, err := finish("failed", "", payload)
		return err
	}
	candidate := []goai.Message{goai.UserMessage(SummaryPrefix + summary + SummarySuffix)}
	candidate = append(candidate, convCtx.Messages[len(convCtx.Messages)-prep.RecentMessages:]...)
	payload["messages_after"] = len(candidate)
	payload["messages_to_summarize"] = prep.MessagesToSummarize
	payload["from_hook"] = decision.Payload != nil && decision.Payload["summary"] != nil
	if err := ctx.Err(); err != nil {
		_, finishErr := finish("cancelled", "", payload)
		if finishErr != nil {
			return finishErr
		}
		return err
	}
	accepted, err := finish("completed", summary, payload)
	if err != nil {
		// The completion transaction rolled back. Best-effort terminal failure
		// restores our phase, but the caller must not proceed to inference.
		failed := map[string]any{"phase": "compacting", "checkpoint": true, "reason": "persistence", "detail": err.Error()}
		_, _ = finish("failed", "", failed)
		return fmt.Errorf("persist compaction: %w", err)
	}
	if accepted != "completed" {
		return context.Canceled
	}
	convCtx.Messages = candidate
	if ops.Broadcast != nil {
		// Retain the existing completion notice for TUI/legacy subscribers, only
		// after durable summary and terminal event are committed.
		ops.Broadcast(req.SessionID, map[string]any{"type": "compaction", "chat_jid": "gi:" + req.SessionID, "turn_id": req.TurnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(candidate), "tokens_source": "estimate"})
	}
	if ops.AfterCompact != nil {
		after := map[string]any{}
		for k, v := range payload {
			after[k] = v
		}
		after["summary"] = summary
		ops.AfterCompact(ctx, after)
	}
	return nil
}
