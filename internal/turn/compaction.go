package turn

import (
	"context"
	"fmt"
	"strings"

	goai "github.com/rcarmo/go-ai"
)

const compactionSummaryPrefix = "The conversation history before this point was compacted into the following summary:\n\n<summary>\n"
const compactionSummarySuffix = "\n</summary>"

type compactionPreparation struct {
	ContextTokens       int              `json:"context_tokens"`
	ThresholdTokens     int              `json:"threshold_tokens"`
	KeepRecentTokens    int              `json:"keep_recent_tokens"`
	ReserveTokens       int              `json:"reserve_tokens"`
	MessagesBefore      int              `json:"messages_before"`
	MessagesToSummarize int              `json:"messages_to_summarize"`
	RecentMessages      int              `json:"recent_messages"`
	Strategy            string           `json:"strategy"`
	Transcript          string           `json:"transcript"`
	RecentTranscript    string           `json:"recent_transcript"`
	Messages            []map[string]any `json:"messages"`
}

func (r *sessionRunner) maybeCompactContext(ctx context.Context, sessionID, turnID, agentID, model string, convCtx *goai.Context) {
	settings := r.engine.runtimeCfg.Compaction
	if !settings.Enabled || len(convCtx.Messages) < 6 {
		return
	}
	tokens := estimateMessagesTokens(convCtx.Messages)
	if tokens <= settings.ThresholdTokens {
		return
	}
	prep := prepareCompaction(convCtx.Messages, tokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
	if prep.MessagesToSummarize <= 0 {
		return
	}
	payload := map[string]any{
		"reason":      "threshold",
		"preparation": prep,
		"settings":    map[string]any{"enabled": settings.Enabled, "context_window": settings.ContextWindow, "reserve_tokens": settings.ReserveTokens, "keep_recent_tokens": settings.KeepRecentTokens, "threshold_tokens": settings.ThresholdTokens, "strategy": settings.Strategy},
	}
	_ = r.store.UpdateTurnStatusAndPhase(ctx, turnID, "running", "compacting")
	_ = r.store.AppendTurnEvent(ctx, turnID, sessionID, "compaction.started", map[string]any{"phase": "compacting", "checkpoint": true, "reason": "threshold", "tokens_before": tokens, "messages_before": prep.MessagesBefore})
	defer func() {
		_ = r.store.TouchSessionActiveTurn(context.Background(), sessionID, turnID)
		_ = r.store.UpdateTurnStatusAndPhase(context.Background(), turnID, "running", "running")
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
		summary = defaultCompactionSummary(prep)
	}
	if strings.TrimSpace(summary) == "" {
		return
	}
	wrapped := compactionSummaryPrefix + summary + compactionSummarySuffix
	compacted := []goai.Message{goai.UserMessage(wrapped)}
	compacted = append(compacted, convCtx.Messages[len(convCtx.Messages)-prep.RecentMessages:]...)
	convCtx.Messages = compacted
	compactPayload := map[string]any{"reason": "threshold", "summary": summary, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)}
	r.engine.broadcast(sessionID, map[string]any{"type": "compaction", "chat_jid": "gi:" + sessionID, "turn_id": turnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)})
	_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookSessionCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: compactPayload})
	_ = r.store.AppendTurnEvent(ctx, turnID, sessionID, "compaction.completed", map[string]any{"phase": "compacting", "checkpoint": true, "reason": "threshold", "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted)})
	_ = r.store.AddMessage(ctx, "msg_"+turnID+"_compaction", sessionID, "assistant", summary, map[string]any{"kind": "compaction", "turn_id": turnID, "tokens_before": tokens, "messages_before": prep.MessagesBefore, "messages_after": len(compacted), "from_hook": resp.Payload != nil && resp.Payload["summary"] != nil})
}

func prepareCompaction(messages []goai.Message, contextTokens, keepRecentTokens, reserveTokens, thresholdTokens int, strategy string) compactionPreparation {
	keepStart := len(messages)
	kept := 0
	for i := len(messages) - 1; i >= 0; i-- {
		kept += estimateMessageTokens(messages[i])
		keepStart = i
		if kept >= keepRecentTokens {
			break
		}
	}
	if keepStart <= 0 {
		keepStart = len(messages) / 2
	}
	prep := compactionPreparation{ContextTokens: contextTokens, ThresholdTokens: thresholdTokens, KeepRecentTokens: keepRecentTokens, ReserveTokens: reserveTokens, MessagesBefore: len(messages), MessagesToSummarize: keepStart, RecentMessages: len(messages) - keepStart, Strategy: strategy}
	prep.Transcript = serializeMessages(messages[:keepStart])
	prep.RecentTranscript = serializeMessages(messages[keepStart:])
	prep.Messages = messageDTOs(messages[:keepStart])
	return prep
}

func defaultCompactionSummary(prep compactionPreparation) string {
	transcript := strings.TrimSpace(prep.Transcript)
	if transcript == "" {
		return "Earlier conversation contained no text content."
	}
	const max = 12000
	if len(transcript) > max {
		transcript = transcript[len(transcript)-max:]
	}
	return fmt.Sprintf("Earlier conversation was compacted by gi's default heuristic. It contained %d messages and about %d tokens before compaction. Preserve user requirements, decisions, file paths, tests, and pending work from this transcript excerpt:\n\n%s", prep.MessagesToSummarize, prep.ContextTokens, transcript)
}

func estimateMessagesTokens(messages []goai.Message) int {
	total := 0
	for _, msg := range messages {
		total += estimateMessageTokens(msg)
	}
	return total
}

func estimateMessageTokens(msg goai.Message) int {
	return estimateTokens(goai.GetTextContent(&msg)) + 4
}

func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len([]rune(text))/4 + 1
}

func serializeMessages(messages []goai.Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		text := strings.TrimSpace(goai.GetTextContent(&msg))
		if text == "" {
			continue
		}
		fmt.Fprintf(&sb, "%s: %s\n\n", msg.Role, text)
	}
	return sb.String()
}

func messageDTOs(messages []goai.Message) []map[string]any {
	out := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		out = append(out, map[string]any{"role": string(msg.Role), "text": goai.GetTextContent(&msg), "tokens": estimateMessageTokens(msg)})
	}
	return out
}
