package turn

import (
	"context"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func (r *sessionRunner) ownedTerminalDecision(turnID string, decision store.TurnSettlement) store.TurnSettlement {
	if r.sharedSessionCoord == nil {
		return decision
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.current != nil && r.current.turnID == turnID {
		if r.current.recoveredTerminal != nil {
			decision = *r.current.recoveredTerminal
		}
		decision.ClaimToken = r.current.claimToken
	}
	return decision
}

// The heartbeat is a DB writer owned by the main host. Stop and join it before
// cleanup releases the claim, including when finalisation itself is recovered.
func (r *sessionRunner) ownTurnHeartbeat(ctx context.Context, sessionID, token string, cancel context.CancelFunc) func() {
	heartbeatCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); r.heartbeatActiveTurn(heartbeatCtx, sessionID, token, cancel) }()
	return func() { stop(); <-done }
}

func (e *Engine) recoverStagedTerminal(ctx context.Context, claim store.ActiveTurnClaim) (bool, error) {
	record, err := e.store.GetTurn(ctx, claim.TurnID)
	if err != nil {
		return false, err
	}
	staged := claim.Phase == "settling" || record.Phase == "settling"
	if !staged && record.Phase == "cancelling" {
		staged, err = e.store.HasStagedTerminalCheckpoint(ctx, claim.TurnID)
		if err != nil {
			return false, err
		}
	}
	if !staged {
		return false, nil
	}
	runner := e.runner(claim.SessionID)
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.current != nil || runner.hosts[claim.TurnID] != nil {
		return true, store.ErrQueueConflict
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
	decision, err := e.store.AcquireStagedTerminalRecovery(ctx, claim, token, time.Now().Add(-interruptedTurnStaleAfter), claim.Phase == "cancelling")
	if err != nil {
		return true, err
	}
	if decision == nil {
		return true, store.ErrQueueConflict
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: claim.TurnID, claimToken: token, cancel: cancel, recoveredTerminal: decision}
	runner.current = active
	runner.registerHostFenceLocked(active, runCtx)
	e.runs.Add(1)
	go func() {
		defer e.runs.Done()
		defer runner.finishHostFence(active)
		runner.mu.Lock()
		runner.mu.Unlock()
		runner.runTurn(e.store, claim.SessionID, claim.TurnID, runCtx, cancel, active)
	}()
	return true, nil
}
