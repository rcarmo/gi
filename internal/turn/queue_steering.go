package turn

import (
	"context"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/topics"
)

// SteerQueuedTurn has no idle-submit fallback. Admission and removal share one
// store transaction; runner ownership prevents cleanup from interleaving locally.
func (e *Engine) SteerQueuedTurn(ctx context.Context, sessionID, queuedID, activeID string) error {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	opCtx := store.CoordinationContext(ctx, e.backgroundContext())
	if err := e.store.SteerQueuedTurn(opCtx, sessionID, queuedID, activeID); err != nil {
		return err
	}
	if bus := e.Topics(); bus != nil {
		bus.Publish(topics.Envelope{Topic: "session.queue", SessionID: sessionID, Type: "notice", Payload: map[string]any{"type": "queue_changed"}})
	}
	return nil
}
