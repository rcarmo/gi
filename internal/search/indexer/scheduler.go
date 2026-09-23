package indexer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	searchstore "github.com/rcarmo/gi/internal/search/store"
)

var ErrSchedulerClosed = errors.New("workspace index scheduler closed")
var ErrRefreshPending = errors.New("workspace index refresh remains pending after bounded attempts")

// RefreshTicket is shared by all requests coalesced into one batch. Cancelling
// Wait cancels only that wait. The application context or Close owns the work.
type RefreshTicket struct {
	done   chan struct{}
	status searchstore.ScopeStatus
	err    error
}

func (t *RefreshTicket) Wait(ctx context.Context) (searchstore.ScopeStatus, error) {
	select {
	case <-ctx.Done():
		return searchstore.ScopeStatus{}, ctx.Err()
	case <-t.done:
		// Roots is the only mutable field; callers must not share its backing array.
		status := t.status
		status.Roots = append([]string(nil), status.Roots...)
		return status, t.err
	}
}

type refreshBatch struct {
	config   searchstore.ScopeConfig
	ticket   *RefreshTicket
	requests uint64 // protects a request racing the final status read
}

type schedulerPolicy struct {
	attempts                                       int
	timeout, operationTimeout, backoff, maxBackoff time.Duration
}

var defaultSchedulerPolicy = schedulerPolicy{4, 5 * time.Minute, 5 * time.Second, 250 * time.Millisecond, 2 * time.Second}

// Scheduler serialises local work across fixed startup scopes in one workspace.
// It performs no startup scan, watcher work or query-triggered refresh. Request
// is an explicit refresh demand; callers record durable invalidations separately.
// A batch is bounded, but Close must join any uninterruptible OS/database calls.
// The application must Close before closing the database.
type Scheduler struct {
	readStatus func(context.Context, searchstore.ScopeConfig) (searchstore.ScopeStatus, error)
	run        func(context.Context, searchstore.ScopeConfig) error
	policy     schedulerPolicy
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan struct{}
	wake       chan struct{}
	mu         sync.Mutex
	configs    map[string]searchstore.ScopeConfig
	batches    map[string]*refreshBatch
	queue      []*refreshBatch
}

func NewScheduler(ctx context.Context, store *searchstore.RefreshStore, configs []searchstore.ScopeConfig) (*Scheduler, error) {
	return newScheduler(ctx, store, configs, NewWorker(store).Run, defaultSchedulerPolicy)
}

func newScheduler(ctx context.Context, store *searchstore.RefreshStore, configs []searchstore.ScopeConfig, run func(context.Context, searchstore.ScopeConfig) error, policy schedulerPolicy) (*Scheduler, error) {
	if store == nil || run == nil || len(configs) == 0 || len(configs) > 3 {
		return nil, fmt.Errorf("index scheduler requires store and 1–3 configured scopes")
	}
	if policy.attempts < 1 || policy.timeout <= 0 || policy.operationTimeout <= 0 || policy.backoff <= 0 || policy.maxBackoff < policy.backoff {
		return nil, fmt.Errorf("invalid index scheduler bounds")
	}
	s := &Scheduler{readStatus: store.Status, run: run, policy: policy, done: make(chan struct{}), wake: make(chan struct{}, 1), configs: make(map[string]searchstore.ScopeConfig), batches: make(map[string]*refreshBatch)}
	for _, c := range configs {
		if c.Fingerprint() == "" || c.Workspace() != configs[0].Workspace() {
			return nil, fmt.Errorf("index scheduler requires resolved scopes for one workspace")
		}
		if _, exists := s.configs[c.Scope()]; exists {
			return nil, fmt.Errorf("duplicate index scheduler scope %q", c.Scope())
		}
		s.configs[c.Scope()] = c
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	go s.loop()
	return s, nil
}

// Request does not access the database or wait for scanning. The bounded queue
// has at most one batch per configured scope, including the active batch. A
// request during an active batch shares its outcome, rather than forcing another
// scan. Requests made after recording a durable invalidation cannot be lost in
// the final status-read/completion race. Config changes require a new scheduler.
func (s *Scheduler) Request(scope string) (*RefreshTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx.Err() != nil {
		return nil, ErrSchedulerClosed
	}
	config, ok := s.configs[scope]
	if !ok {
		return nil, fmt.Errorf("unconfigured index scheduler scope %q", scope)
	}
	if batch := s.batches[scope]; batch != nil {
		batch.requests++
		return batch.ticket, nil
	}
	batch := &refreshBatch{config: config, ticket: &RefreshTicket{done: make(chan struct{})}, requests: 1}
	s.batches[scope] = batch
	s.queue = append(s.queue, batch)
	select {
	case s.wake <- struct{}{}:
	default:
	}
	return batch.ticket, nil
}

// Close is idempotent and does not return until active work, including worker
// renewal and failure cleanup, and all queued tickets have terminated.
func (s *Scheduler) Close() { s.cancel(); <-s.done }

func (s *Scheduler) loop() {
	defer close(s.done)
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, batch := range s.batches {
			s.finishLocked(batch, searchstore.ScopeStatus{}, ErrSchedulerClosed)
		}
		s.queue = nil
	}()
	for {
		if s.ctx.Err() != nil {
			return
		}
		s.mu.Lock()
		var batch *refreshBatch
		if len(s.queue) > 0 {
			batch = s.queue[0]
			s.queue = s.queue[1:]
		}
		s.mu.Unlock()
		if batch == nil {
			select {
			case <-s.ctx.Done():
				return
			case <-s.wake:
			}
			continue
		}
		s.process(batch)
	}
}

func (s *Scheduler) status(ctx context.Context, c searchstore.ScopeConfig) (searchstore.ScopeStatus, error) {
	op, cancel := context.WithTimeout(ctx, s.policy.operationTimeout)
	defer cancel()
	return s.readStatus(op, c)
}

func (s *Scheduler) finishLocked(batch *refreshBatch, status searchstore.ScopeStatus, err error) {
	batch.ticket.status, batch.ticket.err = status, err
	delete(s.batches, batch.config.Scope())
	close(batch.ticket.done)
}

func (s *Scheduler) finish(batch *refreshBatch, status searchstore.ScopeStatus, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finishLocked(batch, status, err)
}

// completeReady publishes success atomically with local request coalescing. An
// event+Request after the read either makes us re-read or creates a fresh batch.
// Reads are outside the mutex so slow SQLite never blocks Request. Bounded
// rechecks reject a continuous request storm rather than spinning indefinitely.
func (s *Scheduler) completeReady(ctx context.Context, batch *refreshBatch, generation int64) (bool, searchstore.ScopeStatus, error) {
	var status searchstore.ScopeStatus
	for i := 0; i < s.policy.attempts; i++ {
		s.mu.Lock()
		requests := batch.requests
		s.mu.Unlock()
		var err error
		status, err = s.status(ctx, batch.config)
		if err != nil {
			return false, status, err
		}
		if status.State != "ready" || status.ConfigHash != batch.config.Fingerprint() || status.RequestedRevision != status.AcknowledgedRevision || status.Generation <= generation {
			return false, status, nil
		}
		s.mu.Lock()
		if batch.requests == requests {
			s.finishLocked(batch, status, nil)
			s.mu.Unlock()
			return true, status, nil
		}
		s.mu.Unlock()
	}
	return false, status, ErrRefreshPending
}

func (s *Scheduler) process(batch *refreshBatch) {
	ctx, cancel := context.WithTimeout(s.ctx, s.policy.timeout)
	defer cancel()
	baseline, err := s.status(ctx, batch.config)
	if err != nil {
		s.finish(batch, baseline, err)
		return
	}
	status := baseline
	delay := s.policy.backoff
	var lastErr error
	for attempt := 0; attempt < s.policy.attempts; attempt++ {
		if err = ctx.Err(); err != nil {
			s.finish(batch, status, err)
			return
		}
		// Even ready scopes get one explicit attempt. After contention, only a
		// newer compatible generation can satisfy this demand through a peer.
		err = s.run(ctx, batch.config)
		if err != nil && !errors.Is(err, searchstore.ErrRefreshBusy) && !errors.Is(err, searchstore.ErrRefreshLost) {
			s.finish(batch, status, err)
			return // no automatic retries for scan/write failures
		}
		lastErr = err
		var completed bool
		completed, status, err = s.completeReady(ctx, batch, baseline.Generation)
		if completed {
			return
		}
		if err != nil {
			s.finish(batch, status, err)
			return
		}
		if attempt+1 == s.policy.attempts {
			break
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			s.finish(batch, status, ctx.Err())
			return
		case <-timer.C:
		}
		delay = min(delay*2, s.policy.maxBackoff)
		completed, status, err = s.completeReady(ctx, batch, baseline.Generation)
		if completed {
			return
		}
		if err != nil {
			s.finish(batch, status, err)
			return
		}
	}
	s.finish(batch, status, errors.Join(ErrRefreshPending, lastErr))
}
