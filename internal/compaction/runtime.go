package compaction

import (
	"context"
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
}

type RuntimeOps struct {
	BackgroundContext        func() context.Context
	BeforeCompact            func(context.Context, map[string]any, []goai.Message) (HookDecision, error)
	AfterCompact             func(context.Context, map[string]any)
	UpdateTurnStatusAndPhase func(context.Context, string, string, string) error
	AppendTurnEvent          func(context.Context, string, string, string, map[string]any) error
	TouchSessionActiveTurn   func(context.Context, string, string) error
	AddMessage               func(context.Context, string, string, string, string, map[string]any) error
	Broadcast                func(string, map[string]any)
	Warn                     func(string, error)
}

func MaybeCompactContext(ctx context.Context, req RuntimeRequest, convCtx *goai.Context, ops RuntimeOps) {
	settings := req.Settings
	if !settings.Enabled || len(convCtx.Messages) < 6 {
		return
	}
	tokens := EstimateMessagesTokens(convCtx.Messages)
	if tokens <= settings.ThresholdTokens {
		return
	}
	prep := Prepare(convCtx.Messages, tokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
	if prep.MessagesToSummarize <= 0 {
		return
	}
	payload := map[string]any{
		"reason":      "threshold",
		"preparation": prep,
		"settings":    map[string]any{"enabled": settings.Enabled, "context_window": settings.ContextWindow, "reserve_tokens": settings.ReserveTokens, "keep_recent_tokens": settings.KeepRecentTokens, "threshold_tokens": settings.ThresholdTokens, "strategy": settings.Strategy},
	}
	warn(ops, "update turn phase compacting", maybeCallUpdate(ops.UpdateTurnStatusAndPhase, ctx, req.TurnID, "running", "compacting"))
	warn(ops, "append compaction.started event", maybeCallAppend(ops.AppendTurnEvent, ctx, req.TurnID, req.SessionID, "compaction.started", map[string]any{"phase": "compacting", "checkpoint": true, "reason": "threshold", "tokens_before": tokens, "messages_before": prep.MessagesBefore}))
	bgCtx := ctx
	if ops.BackgroundContext != nil {
		bgCtx = ops.BackgroundContext()
	}
	defer func() {
		warn(ops, "touch active turn during compaction restore", maybeCallTouch(ops.TouchSessionActiveTurn, bgCtx, req.SessionID, req.TurnID))
		warn(ops, "restore running phase after compaction", maybeCallUpdate(ops.UpdateTurnStatusAndPhase, bgCtx, req.TurnID, "running", "running"))
	}()
	decision := HookDecision{}
	var err error
	if ops.BeforeCompact != nil {
		decision, err = ops.BeforeCompact(ctx, payload, convCtx.Messages)
	}
	if err != nil || decision.Cancel || decision.Block {
		return
	}
	summary := ""
	if decision.Payload != nil {
		if v, ok := decision.Payload["summary"].(string); ok {
			summary = strings.TrimSpace(v)
		}
	}
	if summary == "" {
		summary = DefaultSummary(prep)
	}
	if strings.TrimSpace(summary) == "" {
		return
	}
	wrapped := SummaryPrefix + summary + SummarySuffix
	compacted := []goai.Message{goai.UserMessage(wrapped)}
	compacted = append(compacted, convCtx.Messages[len(convCtx.Messages)-prep.RecentMessages:]...)
	convCtx.Messages = compacted
	compactPayload := map[string]any{"reason": "threshold", "summary": summary, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)}
	if ops.Broadcast != nil {
		ops.Broadcast(req.SessionID, map[string]any{"type": "compaction", "chat_jid": "gi:" + req.SessionID, "turn_id": req.TurnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)})
	}
	if ops.AfterCompact != nil {
		ops.AfterCompact(ctx, compactPayload)
	}
	warn(ops, "append compaction.completed event", maybeCallAppend(ops.AppendTurnEvent, ctx, req.TurnID, req.SessionID, "compaction.completed", map[string]any{"phase": "compacting", "checkpoint": true, "reason": "threshold", "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)}))
	fromHook := decision.Payload != nil && decision.Payload["summary"] != nil
	warn(ops, "add compaction summary message", maybeCallAdd(ops.AddMessage, ctx, "msg_"+req.TurnID+"_compaction", req.SessionID, "assistant", summary, map[string]any{"kind": "compaction", "turn_id": req.TurnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted), "from_hook": fromHook}))
}

func warn(ops RuntimeOps, label string, err error) {
	if ops.Warn != nil && err != nil {
		ops.Warn(label, err)
	}
}
func maybeCallUpdate(fn func(context.Context, string, string, string) error, ctx context.Context, turnID, status, phase string) error {
	if fn == nil {
		return nil
	}
	return fn(ctx, turnID, status, phase)
}
func maybeCallAppend(fn func(context.Context, string, string, string, map[string]any) error, ctx context.Context, turnID, sessionID, typ string, payload map[string]any) error {
	if fn == nil {
		return nil
	}
	return fn(ctx, turnID, sessionID, typ, payload)
}
func maybeCallTouch(fn func(context.Context, string, string) error, ctx context.Context, sessionID, turnID string) error {
	if fn == nil {
		return nil
	}
	return fn(ctx, sessionID, turnID)
}
func maybeCallAdd(fn func(context.Context, string, string, string, string, map[string]any) error, ctx context.Context, id, sessionID, role, content string, payload map[string]any) error {
	if fn == nil {
		return nil
	}
	return fn(ctx, id, sessionID, role, content, payload)
}
