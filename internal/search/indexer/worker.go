package indexer

import (
	"context"
	"errors"
	"fmt"
	"time"

	searchstore "github.com/rcarmo/gi/internal/search/store"
)

// Worker executes an explicit, bounded refresh. It does not schedule itself,
// poll, queue automatic retries or launch work on application startup. A caller
// owns the Run context and must wait for it before closing the store.
type Worker struct {
	store                                                  *searchstore.RefreshStore
	scan                                                   func(context.Context, searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error)
	leaseTTL, runTimeout, operationTimeout, cleanupTimeout time.Duration
}

func NewWorker(store *searchstore.RefreshStore) *Worker {
	return &Worker{store: store, scan: ScanScope, leaseTTL: 30 * time.Second, runTimeout: 5 * time.Minute, operationTimeout: 5 * time.Second, cleanupTimeout: 2 * time.Second}
}

// Run acquires ownership before scanning. Lease-renewal failure cancels the
// scan, then every renewal is joined before Commit or Fail. No goroutine is
// abandoned. Uninterruptible filesystem/SQLite calls may delay Run's return
// beyond its context deadline; the deadline is not a hard syscall kill.
func (w *Worker) Run(ctx context.Context, config searchstore.ScopeConfig) (err error) {
	if w.store == nil || w.scan == nil {
		return fmt.Errorf("index worker requires store and scanner")
	}
	runCtx, cancelTimeout := context.WithTimeout(ctx, w.runTimeout)
	defer cancelTimeout()
	workCtx, cancelWork := context.WithCancelCause(runCtx)
	defer cancelWork(nil)
	acquireCtx, cancelAcquire := context.WithTimeout(workCtx, w.operationTimeout)
	refresh, err := w.store.Begin(acquireCtx, config, w.leaseTTL)
	cancelAcquire()
	if err != nil {
		return fmt.Errorf("begin index refresh: %w", err)
	}
	// Failure reporting uses its own short context because the caller/scan context
	// may already be cancelled. Fencing prevents it from touching a successor.
	defer func() {
		if err == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), w.cleanupTimeout)
		defer cancel()
		if cleanupErr := refresh.Fail(cleanupCtx, err); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("index cleanup: %w", cleanupErr))
		}
	}()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(w.leaseTTL / 3)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-workCtx.Done():
				return
			case <-ticker.C:
				renewCtx, cancel := context.WithTimeout(workCtx, min(w.operationTimeout, w.leaseTTL/3))
				renewErr := refresh.Renew(renewCtx, w.leaseTTL)
				cancel()
				if renewErr != nil {
					cancelWork(fmt.Errorf("renew index ownership: %w", renewErr))
					return
				}
			}
		}
	}()
	snapshot, scanErr := w.scan(workCtx, config)
	close(stop)
	<-done
	if scanErr != nil || context.Cause(workCtx) != nil {
		return errors.Join(scanErr, context.Cause(workCtx))
	}
	// No renewal goroutine competes with the commit transaction or accidentally
	// interprets successful token release as lost ownership. Renew once immediately
	// before the bounded commit; store fences also check expiry at transaction end.
	renewCtx, cancelRenew := context.WithTimeout(workCtx, w.operationTimeout)
	err = refresh.Renew(renewCtx, w.leaseTTL)
	cancelRenew()
	if err != nil {
		return fmt.Errorf("renew before index commit: %w", err)
	}
	commitCtx, cancelCommit := context.WithTimeout(workCtx, w.operationTimeout)
	defer cancelCommit()
	if err = refresh.Commit(commitCtx, snapshot); err != nil {
		return fmt.Errorf("commit index refresh: %w", err)
	}
	// A successful commit wins a simultaneous caller cancellation: never report an
	// uncommitted failure or encourage an automatic duplicate retry after success.
	return nil
}
