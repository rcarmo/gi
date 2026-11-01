package turn

import (
	"context"
	"time"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func (e *Engine) recoverGenerationReady(ctx context.Context, claim store.ActiveTurnClaim) (bool, error) {
	if claim.Status != "running" || (claim.Phase != "generation_ready" && claim.Phase != "running") {
		return false, nil
	}
	if !e.providerRecoveryEnabled.Load() {
		return true, nil
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
	checkpoint, err := e.store.AcquireGenerationReadyRecovery(ctx, claim, token, time.Now().Add(-interruptedTurnStaleAfter))
	if err != nil {
		return true, err
	}
	if checkpoint == nil {
		return false, nil
	}
	runCtx, cancel := context.WithCancel(e.backgroundContext())
	active := &runningTurn{turnID: claim.TurnID, claimToken: token, cancel: cancel, recoveredGeneration: checkpoint}
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

func (r *sessionRunner) restoreGenerationPreparation(ctx context.Context, s *store.Store, turnID, sessionID, agentID string, checkpoint *store.GenerationReadyCheckpoint) (*providerPreparation, *goai.Context, error) {
	if err := checkpoint.Validate(); err != nil {
		return nil, nil, err
	}
	policy := checkpoint.Lifecycle.Clone()
	p := &providerPreparation{initialized: true, startCaptured: true, model: checkpoint.Model, cfg: r.engine.runtimeCfg, lastToolFailureSig: checkpoint.LastToolFailureSignature, repeatedToolFailureCount: checkpoint.RepeatedToolFailures, startPrompt: policy.SystemPrompt, startMessage: policy.Message, startMessages: policy.Messages, startTools: policy.Tools, startRegistered: policy.RegisteredTools}
	cfg, err := r.engine.runtimeCfg.ForGeneration()
	if err != nil {
		return nil, nil, err
	}
	conv := &goai.Context{}
	if err = r.rebuildProviderPreparation(ctx, s, turnID, sessionID, agentID, checkpoint.Iteration, p, conv, cfg); err != nil {
		return nil, nil, err
	}
	return p, conv, nil
}
