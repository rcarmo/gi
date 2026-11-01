package turn

import (
	"context"
	"log"
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
	e.launchTerminalCleanupLocked(runner, claim.SessionID, claim.TurnID, token, decision)
	return true, nil
}

// The caller holds runner.mu and launchMu. Cancellation and crash recovery share
// this terminal-only host; neither route passes through model/tool setup.
func (e *Engine) launchTerminalCleanupLocked(runner *sessionRunner, sessionID, turnID, token string, decision *store.TurnSettlement) {
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: turnID, claimToken: token, cancel: cancel, recoveredTerminal: decision}
	runner.current = active
	runner.registerHostFenceLocked(active, runCtx)
	e.runs.Add(1)
	go func() {
		defer e.runs.Done()
		defer runner.finishHostFence(active)
		runner.mu.Lock()
		runner.mu.Unlock()
		runner.runTurn(e.store, sessionID, turnID, runCtx, cancel, active)
	}()
}

func (r *sessionRunner) emitAgentEnd(ctx context.Context, s *store.Store, sessionID, turnID, agentID, model string) {
	cleanupCtx := context.WithoutCancel(ctx)
	run, err := s.ClaimAgentEnd(cleanupCtx, sessionID, turnID)
	if err != nil {
		log.Printf("persist agent-end attempt: %v", err)
		return
	}
	if !run {
		return
	}
	reason := "unsettled"
	if rec, err := s.GetTurn(cleanupCtx, turnID); err == nil && isTerminalSubTurnStatus(rec.Status) {
		reason = rec.Status
	}
	_, _ = r.engine.emitHook(cleanupCtx, HookRequest{Name: HookAgentEnd, SessionID: sessionID, TurnID: turnID, AgentID: agentID, Model: model, Payload: map[string]any{"reason": reason}})
}
