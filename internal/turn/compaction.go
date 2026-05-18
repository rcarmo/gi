package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/compaction"
	goai "github.com/rcarmo/go-ai"
)

func (r *sessionRunner) maybeCompactContext(ctx context.Context, sessionID, turnID, agentID, model string, convCtx *goai.Context) {
	settings := r.engine.runtimeCfg.Compaction
	if !settings.Enabled || len(convCtx.Messages) < 6 {
		return
	}
	tokens := compaction.EstimateMessagesTokens(convCtx.Messages)
	if tokens <= settings.ThresholdTokens {
		return
	}
	prep := compaction.Prepare(convCtx.Messages, tokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
	if prep.MessagesToSummarize <= 0 {
		return
	}
	payload := map[string]any{
		"reason":      "threshold",
		"preparation": prep,
		"settings":    map[string]any{"enabled": settings.Enabled, "context_window": settings.ContextWindow, "reserve_tokens": settings.ReserveTokens, "keep_recent_tokens": settings.KeepRecentTokens, "threshold_tokens": settings.ThresholdTokens, "strategy": settings.Strategy},
	}
	warnStore("update turn phase compacting", r.store.UpdateTurnStatusAndPhase(ctx, turnID, "running", "compacting"))
	warnStore("append compaction.started event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "compaction.started", map[string]any{"phase": "compacting", "checkpoint": true, "reason": "threshold", "tokens_before": tokens, "messages_before": prep.MessagesBefore}))
	bgCtx := r.engine.backgroundContext()
	defer func() {
		warnStore("touch active turn during compaction restore", r.store.TouchSessionActiveTurn(bgCtx, sessionID, turnID))
		warnStore("restore running phase after compaction", r.store.UpdateTurnStatusAndPhase(bgCtx, turnID, "running", "running"))
	}()
	resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookSessionBeforeCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload, Messages: convCtx.Messages})
	if err != nil || resp.Cancel || resp.Block {
		return
	}
	summary := ""
	if resp.Payload != nil {
		if v, ok := resp.Payload["summary"].(string); ok {
			summary = strings.TrimSpace(v)
		}
	}
	if summary == "" {
		summary = compaction.DefaultSummary(prep)
	}
	if strings.TrimSpace(summary) == "" {
		return
	}
	wrapped := compaction.SummaryPrefix + summary + compaction.SummarySuffix
	compacted := []goai.Message{goai.UserMessage(wrapped)}
	compacted = append(compacted, convCtx.Messages[len(convCtx.Messages)-prep.RecentMessages:]...)
	convCtx.Messages = compacted
	compactPayload := map[string]any{"reason": "threshold", "summary": summary, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)}
	r.engine.broadcast(sessionID, map[string]any{"type": "compaction", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)})
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookSessionCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: compactPayload})
	warnStore("append compaction.completed event", r.store.AppendTurnEvent(ctx, turnID, sessionID, "compaction.completed", map[string]any{"phase": "compacting", "checkpoint": true, "reason": "threshold", "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)}))
	warnStore("add compaction summary message", r.store.AddMessage(ctx, "msg_"+turnID+"_compaction", sessionID, "assistant", summary, map[string]any{"kind": "compaction", "turn_id": turnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted), "from_hook": resp.Payload != nil && resp.Payload["summary"] != nil}))
}
