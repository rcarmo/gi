package turn

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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

func normalizedLowerString(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func (e *Engine) normalizeRunningSessionState(ctx context.Context, sessionID, activeTurnID string, syncQueue bool, overrideModel string) error {
	if e == nil || e.store == nil || strings.TrimSpace(sessionID) == "" || strings.TrimSpace(activeTurnID) == "" {
		return nil
	}
	opCtx := coordinationContext(ctx, e.backgroundContext())
	sessionState := map[string]any{"status": "running", "active_turn_id": activeTurnID}
	if turnRec, err := e.store.GetTurn(opCtx, activeTurnID); err == nil {
		if model := strings.TrimSpace(stringValue(turnRec.Metadata["model"], "")); model != "" {
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

func stringValue(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func boolValue(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		s = strings.ToLower(strings.TrimSpace(s))
		return s == "true" || s == "1" || s == "yes"
	}
	return false
}

func boolValueOr(v any, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if v == nil {
		return fallback
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			return fallback
		}
		return s == "true" || s == "1" || s == "yes"
	}
	switch num := v.(type) {
	case float64:
		return num != 0
	case int:
		return num != 0
	case int64:
		return num != 0
	case int32:
		return num != 0
	case uint:
		return num != 0
	case uint64:
		return num != 0
	case uint32:
		return num != 0
	default:
		return fallback
	}
}

func intValueOr(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case uint:
		return int(n)
	case uint64:
		return int(n)
	case uint32:
		return int(n)
	case float64:
		return int(n)
	case string:
		n = strings.TrimSpace(n)
		if n == "" {
			return fallback
		}
		if parsed, err := strconv.Atoi(n); err == nil {
			return parsed
		}
	}
	return fallback
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
