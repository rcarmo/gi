package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rcarmo/gi/internal/store"
)

// Turn IDs are session-owned; never accept a cross-session ID or use a prompt
// fallback. These wrappers also serve frontends that do not own a store handle.
func (e *Engine) HeldRetryStatus(ctx context.Context, sessionID, turnID string) (*store.TurnFailure, error) {
	if sessionID == "" || turnID == "" {
		return nil, fmt.Errorf("retry: active session and full turn ID required")
	}
	turn, err := e.store.GetTurn(ctx, turnID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("retry: turn not found in this session")
	}
	if err != nil {
		return nil, err
	}
	if turn.SessionID != sessionID {
		return nil, fmt.Errorf("retry: turn not found in this session")
	}
	f, err := e.store.GetTurnFailure(ctx, turnID)
	if err != nil {
		return nil, err
	}
	if f.SessionID != sessionID {
		return nil, fmt.Errorf("retry: turn not found in this session")
	}
	return f, nil
}

func (e *Engine) RetryHeldTurnInSession(ctx context.Context, sessionID, turnID, summary string) (*SubmitResult, error) {
	if _, err := e.HeldRetryStatus(ctx, sessionID, turnID); err != nil {
		return nil, err
	}
	return e.RetryHeldTurn(ctx, turnID, summary)
}

func (e *Engine) ReleaseHeldRetryInSession(ctx context.Context, sessionID, turnID, token string) error {
	ctx = store.CoordinationContext(ctx, e.backgroundContext())
	if ctx == nil {
		return context.Canceled
	}
	if _, err := e.HeldRetryStatus(ctx, sessionID, turnID); err != nil {
		return err
	}
	if err := e.store.ReleaseUnadmittedHeldRetry(ctx, sessionID, turnID, token); err != nil {
		return err
	}
	e.PublishRuntimeTurnEvent("turn_retry_released", sessionID, turnID, "", "", "held_for_retry_or_skip", map[string]any{"reason": "retry_reservation_released", "submitted": false})
	return nil
}
