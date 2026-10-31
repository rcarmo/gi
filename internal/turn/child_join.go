package turn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rcarmo/gi/internal/store"
)

func (r *sessionRunner) registerHostFenceLocked(active *runningTurn, ctx context.Context) {
	active.done = make(chan struct{})
	active.cancelled = ctx.Done()
	if r.hosts == nil {
		r.hosts = map[string]chan struct{}{}
	}
	r.hosts[active.turnID] = active.done
}
func (r *sessionRunner) finishHostFence(active *runningTurn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hosts, active.turnID)
	close(active.done)
}
func (r *sessionRunner) setTerminalPending(turnID string, pending bool) {
	if r.sharedSessionCoord == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.current != nil && r.current.turnID == turnID {
		r.current.terminalPending = pending
	}
}

// Cancellation requests precede this walk. Join the host, including its final
// hooks and cleanup, even when the child has already committed a terminal row.
// An unowned, nonterminal child cannot be pronounced settled by this engine.
func (r *sessionRunner) joinRequiredChildren(ctx context.Context, parentTurnID, status, failureKind string) error {
	directive := directiveForParentTerminalStatus(status, failureKind)
	if !directive.enabled {
		return nil
	}
	var cancelled <-chan struct{}
	if r.sharedSessionCoord != nil && status != "cancelled" && status != "aborted" {
		r.mu.Lock()
		if r.current != nil && r.current.turnID == parentTurnID {
			cancelled = r.current.cancelled
		}
		r.mu.Unlock()
	}
	visited := map[string]bool{parentTurnID: true}
	visiting := map[string]bool{}
	var walk func(string) error
	walk = func(turnID string) error {
		visiting[turnID] = true
		defer delete(visiting, turnID)
		subs, err := r.store.ListSubTurnsByParent(ctx, turnID)
		if err != nil {
			return err
		}
		for _, sub := range subs {
			if visiting[sub.ChildTurnID] {
				return fmt.Errorf("child ownership cycle at %s", sub.ChildTurnID)
			}
			if visited[sub.ChildTurnID] {
				continue
			}
			visited[sub.ChildTurnID] = true
			if err := walk(sub.ChildTurnID); err != nil {
				return err
			}
			if isCriticalSubTurn(sub) && !directive.cancelCritical {
				continue
			}
			childRunner := r.engine.runner(sub.ChildSessionID)
			childRunner.mu.Lock()
			done := childRunner.hosts[sub.ChildTurnID]
			childRunner.mu.Unlock()
			if done != nil {
				select {
				case <-done:
				case <-cancelled:
					return store.ErrQueueConflict
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if done == nil {
				activeID, _, err := r.store.GetSessionActiveTurn(ctx, sub.ChildSessionID)
				if err == nil && activeID == sub.ChildTurnID {
					return fmt.Errorf("required child %s still has an unjoined owner", sub.ChildTurnID)
				}
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			}
			r.engine.joinTurnHooks(sub.ChildTurnID, false)
			terminal, err := r.store.HasTurnTerminalReceipt(ctx, sub.ChildTurnID)
			if err != nil {
				return err
			}
			if !terminal {
				return fmt.Errorf("required child %s has no joined terminal receipt", sub.ChildTurnID)
			}
		}
		return nil
	}
	return walk(parentTurnID)
}
