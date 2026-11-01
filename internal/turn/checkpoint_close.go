package turn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

type closingRetry struct {
	sessionID, turnID, token string
	reserved                 bool
	stopHeartbeat            func() error
}

// The reservation revokes the old writer without publishing queued work. Keep
// its new lease alive until even cancellation-ignoring callbacks have joined.
func (e *Engine) keepClosingClaim(sessionID, token string) func() error {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	failures := make(chan error, 1)
	go func() {
		defer close(done)
		ticker := time.NewTicker(activeTurnHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := e.store.TouchSessionActiveTurn(ctx, sessionID, token); err != nil && ctx.Err() == nil {
					select {
					case failures <- err:
					default:
					}
				}
			}
		}
	}()
	return func() error {
		cancel()
		<-done
		select {
		case err := <-failures:
			return err
		default:
			return nil
		}
	}
}

func (e *Engine) stopActiveTurnsForClose(grace time.Duration) ([]closingRetry, error) {
	var plans []closingRetry
	var result error
	ctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	e.sessions.Range(func(key, value any) bool {
		sessionID, runner := key.(string), value.(*sessionRunner)
		runner.mu.Lock()
		active := runner.current
		if active == nil || active.owner != e {
			runner.mu.Unlock()
			return true
		}
		token, err := store.NewActiveTurnClaimToken()
		var checkpoint *store.ProviderRetryCheckpoint
		if err == nil {
			checkpoint, err = e.store.ReserveProviderRetryClose(ctx, sessionID, active.turnID, active.claimToken, token)
		}
		if err != nil || checkpoint != nil {
			// A failed reservation is not permission to fabricate an abort. Preserve
			// its original claim for conservative recovery and return the error.
			active.checkpointClosing = true
			active.terminalPending = true
			reserved := err == nil
			if !reserved {
				token = active.claimToken
				result = errors.Join(result, fmt.Errorf("reserve close checkpoint %s: %w", active.turnID, err))
			}
			plans = append(plans, closingRetry{sessionID: sessionID, turnID: active.turnID, token: token, reserved: reserved, stopHeartbeat: e.keepClosingClaim(sessionID, token)})
			active.cancel()
			runner.mu.Unlock()
			return true
		}
		runner.mu.Unlock()
		// Other phases keep their existing conservative abort settlement until a
		// verified continuation route exists. Do not queue a whole-prompt replay.
		if err = e.cancelTurn(ctx, sessionID, active.turnID, false, true, true); err != nil {
			log.Printf("shutdown: abort %s: %v", active.turnID, err)
			if !errors.Is(err, store.ErrQueueConflict) {
				result = errors.Join(result, fmt.Errorf("abort during close %s: %w", active.turnID, err))
			}
			active.cancel()
		}
		return true
	})
	done := make(chan struct{})
	go func() { e.runs.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		log.Printf("shutdown: turns still finalizing after %s", grace)
	}
	return plans, result
}

func (r *sessionRunner) checkpointClosing(turnID string) bool {
	if r.sharedSessionCoord == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current != nil && r.current.turnID == turnID && r.current.checkpointClosing
}
