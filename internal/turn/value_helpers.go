package turn

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func coordinationContext(ctx context.Context, fallback context.Context) context.Context {
	if ctx != nil && ctx.Err() == nil {
		return ctx
	}
	if fallback != nil && fallback.Err() == nil {
		return fallback
	}
	return nil
}

func (e *Engine) normalizeRunningSessionState(ctx context.Context, sessionID, activeTurnID string, syncQueue bool, overrideModel string) error {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || strings.TrimSpace(activeTurnID) == "" {
		return nil
	}
	opCtx := coordinationContext(ctx, e.backgroundContext())
	sessionState := map[string]any{"status": "running", "active_turn_id": activeTurnID}
	if turnRec, err := e.store.GetTurn(opCtx, activeTurnID); err == nil {
		if model := strings.TrimSpace(store.StringValue(turnRec.Metadata["model"], "")); model != "" {
			sessionState["model"] = model
		}
	}
	if model := strings.TrimSpace(overrideModel); model != "" {
		sessionState["model"] = model
	}
	if syncQueue {
		if err := e.store.SyncSessionQueueCount(opCtx, sessionID); err != nil {
			return fmt.Errorf("sync queue count on running session normalization: %w", err)
		}
	}
	if err := e.store.TouchSessionState(opCtx, sessionID, sessionState); err != nil {
		return fmt.Errorf("touch running session normalization: %w", err)
	}
	return nil
}

func (e *Engine) normalizeInactiveSessionState(ctx context.Context, sessionID, status, model string, syncQueue bool) error {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	opCtx := coordinationContext(ctx, e.backgroundContext())
	sessionState := map[string]any{"status": status, "active_turn_id": nil}
	if model = strings.TrimSpace(model); model != "" {
		sessionState["model"] = model
	}
	if syncQueue {
		if err := e.store.SyncSessionQueueCount(opCtx, sessionID); err != nil {
			return fmt.Errorf("sync queue count on inactive session normalization: %w", err)
		}
	}
	if err := e.store.TouchSessionState(opCtx, sessionID, sessionState); err != nil {
		return fmt.Errorf("touch inactive session normalization: %w", err)
	}
	return nil
}

func runtimeTurnPhaseForStatus(status string) string {
	switch status {
	case "queued":
		return "queued"
	case "running":
		return "setup"
	case "cancelling":
		return "cancelling"
	case "completed":
		return "completed"
	case "failed":
		return "failed"
	case "aborted", "cancelled":
		return "aborted"
	default:
		return status
	}
}

func normalizeSubTurnDeliveryMode(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return "sync", nil
	}
	switch mode {
	case "sync", "async":
		return mode, nil
	default:
		return "", fmt.Errorf("invalid subturn delivery mode: %s", mode)
	}
}

func SortQueuedTurns(turns []store.Turn) {
	sort.SliceStable(turns, func(i, j int) bool { return turns[i].CreatedAt < turns[j].CreatedAt })
}
