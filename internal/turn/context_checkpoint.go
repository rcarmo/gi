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
