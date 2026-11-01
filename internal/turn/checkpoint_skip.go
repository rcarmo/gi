package turn

import (
	"context"

	"github.com/rcarmo/gi/internal/store"
)

// Explicit skip owns terminal cleanup only. It cannot enter model/tool setup,
// revive a consumed child, or turn an interrupted failure into cancellation.
func (e *Engine) skipHeldCheckpoint(ctx context.Context, sessionID, turnID, summary string) (bool, error) {
	runner := e.runner(sessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil || len(runner.hosts) != 0 {
		return true, store.ErrRetryPending
	}
	e.launchMu.Lock()
	defer e.launchMu.Unlock()
	if e.closing.Load() {
		return true, context.Canceled
	}
	token, err := store.NewActiveTurnClaimToken()
	if err != nil {
		return true, err
	}
	decision, handled, err := e.store.StageHeldCheckpointSkip(ctx, turnID, token, summary)
	if err != nil || !handled {
		return handled, err
	}
	if decision == nil {
		return true, store.ErrFailureConflict
	}
	e.launchTerminalCleanupLocked(runner, sessionID, turnID, token, decision)
	return true, nil
}
