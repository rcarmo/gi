package turn

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func (r *sessionRunner) projectContextSnapshot(ctx context.Context, sessionID string, snapshot store.ContextSnapshot) []goai.Message {
	messages, _ := reduceContextContributions(r.rawContextSnapshot(ctx, sessionID, snapshot))
	return messages
}

func (r *sessionRunner) rawContextSnapshot(ctx context.Context, sessionID string, snapshot store.ContextSnapshot) []goai.Message {
	var messages []goai.Message
	if snapshot.Summary != "" {
		messages = append(messages, goai.UserMessage(compaction.SummaryPrefix+snapshot.Summary+compaction.SummarySuffix))
	}
	for _, m := range snapshot.Messages {
		if kind, _ := m.Payload["kind"].(string); m.Role == "user" && kind == BranchSummaryKind {
			messages = append(messages, branchSummaryContext(m.Content))
		} else if m.Role == "user" {
			messages = append(messages, r.userMessageWithProviderSafeMedia(ctx, sessionID, m.Content, m.Payload))
		} else if m.Role == "tool_result" {
			id, _ := m.Payload["tool_call_id"].(string)
			name, _ := m.Payload["tool_name"].(string)
			failed, _ := m.Payload["is_error"].(bool)
			blocks := []goai.ContentBlock{{Type: "text", Text: m.Content}}
			if saved, ok := decodeReplayValue[[]goai.ContentBlock](m.Payload["provider_content_blocks"]); ok {
				blocks = saved
			}
			added, _ := decodeReplayValue[[]string](m.Payload["added_tool_names"])
			messages = append(messages, goai.Message{Role: goai.RoleToolResult, ToolCallID: id, ToolName: name, IsError: failed, Content: blocks, AddedToolNames: added, Details: m.Payload["details"]})
		} else if saved, ok := decodeAssistantReplay(m.Payload["provider_message"]); ok {
			messages = append(messages, saved)
		} else if m.Payload["kind"] == "tool_calls" {
			text, _ := m.Payload["display_text"].(string)
			content := []goai.ContentBlock{{Type: "text", Text: text}}
			var calls []struct {
				ID             string         `json:"id"`
				Name           string         `json:"name"`
				ArgumentsError string         `json:"argumentsError"`
				Arguments      map[string]any `json:"arguments"`
			}
			if raw, err := json.Marshal(m.Payload["tool_calls"]); err == nil && json.Unmarshal(raw, &calls) == nil {
				for _, call := range calls {
					content = append(content, goai.ContentBlock{Type: "toolCall", ID: call.ID, Name: call.Name, Arguments: call.Arguments, ArgumentsError: call.ArgumentsError})
				}
			}
			messages = append(messages, goai.Message{Role: goai.RoleAssistant, Content: content})
		} else {
			messages = append(messages, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
		}
		// Legacy receipts may carry their terminal reason only in display metadata.
		if last := &messages[len(messages)-1]; last.Role == goai.RoleAssistant && last.StopReason == "" {
			if stop, ok := m.Payload["stop_reason"].(string); ok {
				last.StopReason = goai.StopReason(stop)
			}
		}
	}
	return messages
}

func (r *sessionRunner) compactionBoundary(ctx context.Context, sessionID string, conv *goai.Context, snapshot store.ContextSnapshot, payload map[string]any) (*store.ContextBoundary, error) {
	// Hook edits and live tool exchanges cannot be mapped merely by message count.
	// Preserve run-local compaction but do not advance durable history coverage.
	projected, boundaries := reduceContextContributions(r.rawContextSnapshot(ctx, sessionID, snapshot))
	if !reflect.DeepEqual(conv.Messages, projected) {
		return nil, nil
	}
	count, ok := payload["messages_to_summarize"].(int)
	if !ok {
		return nil, nil
	}
	rawCount, boundaryOK := boundaries[count]
	if !boundaryOK {
		// In particular, do not split a batch or one of its synthetic results.
		return nil, nil
	}
	nativeCount := rawCount
	if snapshot.Summary != "" {
		nativeCount--
	}
	if nativeCount < 0 || nativeCount > len(snapshot.Messages) {
		return nil, store.ErrContextChanged
	}
	// Never checkpoint between a tool-call batch and its results. Keeping the
	// complete exchange in the suffix avoids orphaned results on next-turn replay.
	pending := map[string]bool{}
	for _, m := range conv.Messages[:count] {
		for _, call := range goai.GetToolCalls(&m) {
			pending[call.ID] = true
		}
		if m.Role == goai.RoleToolResult {
			delete(pending, m.ToolCallID)
		}
	}
	if len(pending) != 0 {
		return nil, nil
	}
	// The summariser has a text transcript, not image bytes. Do not permanently
	// hide media-bearing older messages behind a text-only summary.
	for _, m := range snapshot.Messages[:nativeCount] {
		if blocks, ok := decodeReplayValue[[]goai.ContentBlock](m.Payload["provider_content_blocks"]); ok {
			for _, block := range blocks {
				if block.Type == "image" {
					return nil, nil
				}
			}
		}
		refs, err := store.NormalizeMediaReferences(m.Payload["media"])
		if err != nil || len(refs) > 0 {
			return nil, nil
		}
	}
	return store.PrepareContextBoundary(snapshot, rawCount)
}

func (r *sessionRunner) compactSnapshot(ctx context.Context, sessionID, turnID, model, agentID string, convCtx *goai.Context, snapshot store.ContextSnapshot, force bool) error {
	return r.compactSnapshotWithInstructions(ctx, sessionID, turnID, model, agentID, convCtx, snapshot, force, "")
}

func (r *sessionRunner) compactSnapshotWithInstructions(ctx context.Context, sessionID, turnID, model, agentID string, convCtx *goai.Context, snapshot store.ContextSnapshot, force bool, instructions string) error {
	return compaction.MaybeCompactContext(ctx, compaction.RuntimeRequest{SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Settings: r.engine.CompactionPolicy(), Force: force, Instructions: instructions, ModelMaxTokens: inference.ModelMaxTokens(model)}, convCtx, compaction.RuntimeOps{BackgroundContext: r.engine.backgroundContext, Summarize: r.compactionSummarizer(turnID, model), BeforeCompact: func(ctx context.Context, payload map[string]any, messages []goai.Message) (compaction.HookDecision, error) {
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

// compactionSummarizer asks the session model for a compaction summary (Pi):
// no tools, the turn's thinking level, a capped response and no cache
// writes.
func (r *sessionRunner) compactionSummarizer(turnID, model string) compaction.Summarizer {
	return func(ctx context.Context, req compaction.SummaryRequest) (compaction.SummaryResponse, error) {
		thinking := ""
		if rec, err := r.store.GetTurn(ctx, turnID); err == nil {
			thinking, _ = rec.Metadata["selected_thinking_level"].(string)
		}
		return summarizeWith(ctx, model, thinking, req, goai.Transport(r.engine.runtimeCfg.Transport))
	}
}

// summarizeWith asks model for a summary: no tools, the given thinking
// level, a capped response and no cache writes.
func summarizeWith(ctx context.Context, model, thinking string, req compaction.SummaryRequest, transport ...goai.Transport) (compaction.SummaryResponse, error) {
	if thinking == "off" {
		thinking = ""
	}
	convCtx := &goai.Context{SystemPrompt: req.SystemPrompt, Messages: []goai.Message{goai.UserMessage(req.Prompt)}}
	hooks := &inference.StreamHooks{Thinking: thinking, MaxTokens: req.MaxTokens, CacheRetention: goai.CacheRetentionNone}
	if len(transport) > 0 {
		hooks.Transport = transport[0]
	}
	result, err := streamWithToolsWithHooks(ctx, model, convCtx, nil, hooks)
	if err != nil {
		return compaction.SummaryResponse{}, err
	}
	resp := compaction.SummaryResponse{Text: result.Text, Usage: result.Usage}
	if result.Message != nil {
		resp.StopReason, resp.ErrorMessage = result.Message.StopReason, result.Message.ErrorMessage
		for _, b := range result.Message.Content {
			if b.Type == "toolCall" {
				resp.ToolCall = true
			}
		}
	}
	return resp, nil
}

// Releases through v0.2.0 wrote StreamResult envelopes. Accept those as well
// as the direct protocol message written now; never discard signed blocks just
// because the legacy envelope has no top-level role.
func decodeAssistantReplay(value any) (goai.Message, bool) {
	if message, ok := decodeReplayValue[goai.Message](value); ok && message.Role == goai.RoleAssistant {
		return message, true
	}
	if result, ok := decodeReplayValue[inference.StreamResult](value); ok && result.Message != nil && result.Message.Role == goai.RoleAssistant {
		return *result.Message, true
	}
	return goai.Message{}, false
}

func decodeReplayValue[T any](value any) (T, bool) {
	var out T
	if value == nil {
		return out, false
	}
	if typed, ok := value.(T); ok {
		return typed, true
	}
	raw, err := json.Marshal(value)
	if err != nil || json.Unmarshal(raw, &out) != nil {
		return out, false
	}
	return out, true
}
