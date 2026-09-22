package turn

import (
	"context"
	"reflect"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func (r *sessionRunner) projectContextSnapshot(ctx context.Context, sessionID string, snapshot store.ContextSnapshot) []goai.Message {
	var messages []goai.Message
	if snapshot.Summary != "" {
		messages = append(messages, goai.UserMessage(compaction.SummaryPrefix+snapshot.Summary+compaction.SummarySuffix))
	}
	for _, m := range snapshot.Messages {
		if m.Role == "user" {
			messages = append(messages, r.userMessageWithProviderSafeMedia(ctx, sessionID, m.Content, m.Payload))
		} else {
			messages = append(messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
		}
	}
	return messages
}

func (r *sessionRunner) compactionBoundary(ctx context.Context, sessionID string, conv *goai.Context, snapshot store.ContextSnapshot, payload map[string]any) (*store.ContextBoundary, error) {
	// Hook edits and live tool exchanges cannot be mapped merely by message count.
	// Preserve run-local compaction but do not advance durable history coverage.
	if !reflect.DeepEqual(conv.Messages, r.projectContextSnapshot(ctx, sessionID, snapshot)) {
		return nil, nil
	}
	count, ok := payload["messages_to_summarize"].(int)
	if !ok {
		return nil, nil
	}
	nativeCount := count
	if snapshot.Summary != "" {
		nativeCount--
	}
	if nativeCount < 0 || nativeCount > len(snapshot.Messages) {
		return nil, store.ErrContextChanged
	}
	// The summariser has a text transcript, not image bytes. Do not permanently
	// hide media-bearing older messages behind a text-only summary.
	for _, m := range snapshot.Messages[:nativeCount] {
		refs, err := store.NormalizeMediaReferences(m.Payload["media"])
		if err != nil || len(refs) > 0 {
			return nil, nil
		}
	}
	return store.PrepareContextBoundary(snapshot, count)
}

func (r *sessionRunner) compactSnapshot(ctx context.Context, sessionID, turnID, model, agentID string, convCtx *goai.Context, snapshot store.ContextSnapshot, force bool) error {
	return compaction.MaybeCompactContext(ctx, compaction.RuntimeRequest{SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Settings: r.engine.runtimeCfg.Compaction, Force: force}, convCtx, compaction.RuntimeOps{BackgroundContext: r.engine.backgroundContext, BeforeCompact: func(ctx context.Context, payload map[string]any, messages []goai.Message) (compaction.HookDecision, error) {
		resp, err := r.engine.emitHook(ctx, HookRequest{Name: HookSessionBeforeCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload, Messages: messages})
		return compaction.HookDecision{Cancel: resp.Cancel, Block: resp.Block, Payload: resp.Payload}, err
	}, AfterCompact: func(ctx context.Context, payload map[string]any) {
		_, _ = r.engine.emitHook(ctx, HookRequest{Name: HookSessionCompact, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: payload})
	}, Begin: r.store.BeginCompaction, Finish: func(finishCtx context.Context, sid, tid string, seq int, outcome, summary string, payload map[string]any) (string, error) {
		var boundary *store.ContextBoundary
		if outcome == "completed" {
			var err error
			boundary, err = r.compactionBoundary(finishCtx, sid, convCtx, snapshot, payload)
			if err != nil {
				return "", err
			}
		}
		if force && outcome == "completed" && boundary == nil {
			return "", store.ErrContextChanged
		}
		accepted, err := r.store.FinishCompactionWithBoundary(finishCtx, sid, tid, seq, outcome, summary, payload, boundary)
		if err == nil {
			payload["durable_context"] = accepted == "completed" && boundary != nil
		}
		return accepted, err
	}, Broadcast: r.engine.broadcast})
}
