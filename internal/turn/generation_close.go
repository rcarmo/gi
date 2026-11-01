package turn

import (
	"context"
	"errors"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// checkpointProgressPersistence binds bounded recorders to this physical host.
// The closing incarnation stays inside the coordinator and never reaches tools.
// A replacement host on the same logical turn cannot inherit these callbacks.
func (r *sessionRunner) checkpointProgressPersistence(ctx context.Context, sessionID, turnID string) func(func(context.Context) error) error {
	ctx = context.WithoutCancel(ctx)
	if r.sharedSessionCoord == nil {
		return func(persist func(context.Context) error) error { return persist(ctx) }
	}
	r.mu.Lock()
	owner := r.current
	r.mu.Unlock()
	closingContext := func() (context.Context, bool) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if owner != nil && r.current == owner && owner.turnID == turnID && owner.closingProgressToken != "" {
			return store.WithTurnWriter(ctx, sessionID, turnID, owner.closingProgressToken), true
		}
		return ctx, false
	}
	return func(persist func(context.Context) error) error {
		writeCtx, closing := closingContext()
		err := persist(writeCtx)
		// Reservation can win between reading the token and the DB write.
		if errors.Is(err, store.ErrQueueConflict) && !closing {
			if next, available := closingContext(); available {
				err = persist(next)
			}
		}
		if err != nil {
			r.mu.Lock()
			if owner != nil && r.current == owner && owner.turnID == turnID && owner.closingProgressToken != "" {
				owner.closingProgressErr = errors.Join(owner.closingProgressErr, err)
			}
			r.mu.Unlock()
		}
		return err
	}
}

func (r *sessionRunner) providerPartialPersistence(ctx context.Context, s *store.Store, sessionID, turnID, requestID string) func(string, map[string]any) error {
	if r.sharedSessionCoord == nil {
		return nil
	}
	r.mu.Lock()
	owner := r.current
	r.mu.Unlock()
	if owner == nil || owner.turnID != turnID {
		return nil
	}
	persist := r.checkpointProgressPersistence(ctx, sessionID, turnID)
	return func(content string, payload map[string]any) error {
		return persist(func(writeCtx context.Context) error {
			return s.SaveProviderPartial(writeCtx, turnID, sessionID, requestID, content, payload)
		})
	}
}

// captureClosingProviderOutcome retains a physical terminal without replacing
// committed progress or approving its body as a logical generation result.
func (r *sessionRunner) captureClosingProviderOutcome(ctx context.Context, s *store.Store, sessionID, turnID, model string, result *inference.StreamResult) error {
	if result == nil || result.OccurrenceID == "" {
		return nil
	}
	return r.checkpointProgressPersistence(ctx, sessionID, turnID)(func(writeCtx context.Context) error {
		record, err := s.ProviderPartial(writeCtx, turnID, sessionID, result.OccurrenceID)
		if err != nil {
			return err
		}
		payload := map[string]any{"model": model, "provider_message": &goai.Message{Role: goai.RoleAssistant}}
		content := ""
		if record != nil {
			payload = record.Payload
			content = goai.GetTextContent(record.Message)
		}
		usage := result.Usage
		if (providerUsageUnknown(usage) || store.ValidateReceiptUsage(usage) != nil) && result.Message != nil {
			usage = result.Message.Usage
		}
		if store.ValidateReceiptUsage(usage) != nil {
			usage = nil
		}
		if !providerUsageUnknown(usage) {
			copied := *usage
			payload["provider_received_usage"] = &copied
		}
		if result.Message != nil {
			terminal, truncated := boundedProviderPartial(result.Message)
			payload["provider_received_terminal"] = terminal
			payload["provider_received_terminal_checkpoint"] = map[string]any{"truncated": truncated, "string_bytes_limit": providerPartialStringBytes, "blocks_limit": 64}
		}
		return s.SaveProviderPartial(writeCtx, turnID, sessionID, result.OccurrenceID, content, payload)
	})
}
