package turn

import (
	"context"
	"fmt"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func (r *sessionRunner) restoreProviderPreparation(wait *store.ProviderRetryWait) (*providerPreparation, error) {
	if wait == nil || wait.Iteration < 1 || wait.Attempt < 1 || wait.Lifecycle == nil {
		return nil, fmt.Errorf("provider retry lacks continuation policy")
	}
	if err := wait.Lifecycle.Validate(); err != nil {
		return nil, err
	}
	// Current settings are read by refreshProviderPreparation after the wake.
	// A transient settings error must not finish a future wait prematurely.
	cfg := r.engine.runtimeCfg
	policy := wait.Lifecycle.Clone()
	return &providerPreparation{initialized: true, startCaptured: true, retryAttempt: wait.Attempt, lastToolFailureSig: wait.LastToolFailureSignature, repeatedToolFailureCount: wait.RepeatedToolFailures, model: wait.Model, cfg: cfg, startPrompt: policy.SystemPrompt, startMessage: policy.Message, startMessages: policy.Messages, startTools: policy.Tools, startRegistered: policy.RegisteredTools}, nil
}

func (r *sessionRunner) waitRecoveredProviderRetry(ctx context.Context, s *store.Store, sessionID, turnID string, checkpoint *store.ProviderRetryCheckpoint) error {
	if checkpoint.Resumed {
		return ctx.Err()
	} // fresh preparation follows the already committed wake
	r.publishProviderRetryWait(turnID, sessionID, checkpoint.Wait)
	timer := time.NewTimer(max(0, time.Until(checkpoint.Wait.RetryAt)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.ResumeProviderRetry(ctx, sessionID, turnID, checkpoint.OccurrenceID)
}

func (e *Engine) recoverProviderRetry(ctx context.Context, claim store.ActiveTurnClaim) (bool, error) {
	if (claim.Phase != "retry_wait" && claim.Phase != "running") || claim.Status != "running" {
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
	checkpoint, err := e.store.AcquireProviderRetryRecovery(ctx, claim, token, time.Now().Add(-interruptedTurnStaleAfter))
	if err != nil {
		return true, err
	}
	if checkpoint == nil {
		return false, nil
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: claim.TurnID, claimToken: token, cancel: cancel, recoveredRetry: checkpoint}
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

// ResumeInterruptedWork opens effectful recovery only after the host has
// registered tools/hooks and enabled MCP. Constructors may recover terminal
// projections, but cannot dispatch a provider before runtime setup is complete.
func (e *Engine) ResumeInterruptedWork(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("nil recovery context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if e.closing.Load() {
		return context.Canceled
	}
	e.awaitDirectMCPTools(ctx)
	if err := ctx.Err(); err != nil {
		return err
	}
	e.providerRecoveryEnabled.Store(true)
	_, err := e.recoverInterruptedTurns(ctx, "")
	return err
}
