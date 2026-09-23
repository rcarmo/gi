package indexer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	searchstore "github.com/rcarmo/gi/internal/search/store"
	core "github.com/rcarmo/gi/internal/store"
)

var testSchedulerPolicy = schedulerPolicy{4, 10 * time.Second, time.Second, 5 * time.Millisecond, 20 * time.Millisecond}

func testScheduler(t *testing.T, store *searchstore.RefreshStore, configs []searchstore.ScopeConfig, run func(context.Context, searchstore.ScopeConfig) error, policy schedulerPolicy) *Scheduler {
	t.Helper()
	s, err := newScheduler(t.Context(), store, configs, run, policy)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}
func request(t *testing.T, s *Scheduler, scope string) *RefreshTicket {
	t.Helper()
	ticket, err := s.Request(scope)
	if err != nil {
		t.Fatal(err)
	}
	return ticket
}
func waitTicket(t *testing.T, ticket *RefreshTicket) (searchstore.ScopeStatus, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	return ticket.Wait(ctx)
}
func requireReady(t *testing.T, ticket *RefreshTicket, generation int64) searchstore.ScopeStatus {
	t.Helper()
	status, err := waitTicket(t, ticket)
	if err != nil || status.State != "ready" || status.RequestedRevision != status.AcknowledgedRevision || status.Generation != generation {
		t.Fatal(status, err)
	}
	return status
}

func TestSchedulerIdleCoalescingWaitCancellationAndExplicitReadyRefresh(t *testing.T) {
	_, store, c, _ := workerFixture(t)
	w := NewWorker(store)
	entered, release := make(chan struct{}), make(chan struct{})
	var scans atomic.Int32
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		if scans.Add(1) == 1 {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
				return searchstore.CompleteSnapshot{}, ctx.Err()
			}
		}
		return ScanScope(ctx, c)
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, w.Run, testSchedulerPolicy)
	if status := workerStatus(t, store, c); status.State != "never_indexed" || scans.Load() != 0 {
		t.Fatal("startup performed work", status)
	}
	if _, err := s.Request("missing"); err == nil {
		t.Fatal("unconfigured scope accepted")
	}
	first := request(t, s, c.Scope())
	awaitSignal(t, entered)
	const count = 64
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticket, err := s.Request(c.Scope())
			if err != nil || ticket != first {
				t.Errorf("requests not coalesced: %p %v", ticket, err)
			}
		}()
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := first.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if scans.Load() != 1 {
		t.Fatal("parallel scan")
	}
	close(release)
	status := requireReady(t, first, 1)
	status.Roots[0] = "caller mutation"
	if got := requireReady(t, first, 1); got.Roots[0] == "caller mutation" {
		t.Fatal("shared ticket slice leaked")
	}
	requireReady(t, request(t, s, c.Scope()), 2) // explicit request must refresh even ready
	if scans.Load() != 2 {
		t.Fatal(scans.Load())
	}
}

func TestSchedulerDrainsNewerRevisionWithOneLocalWorker(t *testing.T) {
	db, store, c, _ := workerFixture(t)
	w := NewWorker(store)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	captured, release := make(chan struct{}), make(chan struct{})
	var scans atomic.Int32
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		snapshot, err := ScanScope(ctx, c)
		if scans.Add(1) == 1 {
			close(captured)
			select {
			case <-release:
			case <-ctx.Done():
				return snapshot, ctx.Err()
			}
		}
		return snapshot, err
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, w.Run, testSchedulerPolicy)
	ticket := request(t, s, c.Scope())
	awaitSignal(t, captured)
	if err := os.WriteFile(filepath.Join(c.Workspace(), "notes/a.md"), []byte("fresh dahlia"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Invalidate(t.Context(), c, []string{"notes/a.md"}); err != nil {
		t.Fatal(err)
	}
	if next := request(t, s, c.Scope()); next != ticket {
		t.Fatal("newer request not coalesced")
	}
	close(release)
	status := requireReady(t, ticket, 3)
	if scans.Load() != 2 || status.AcknowledgedRevision != 1 || workerMatches(t, db, "dahlia") != 1 || workerMatches(t, db, "orchid") != 0 {
		t.Fatal(scans.Load(), status)
	}
}

func TestSchedulerSerialisesScopesAndIsolatesWorkspace(t *testing.T) {
	_, store, c, _ := workerFixture(t)
	skills := scanConfig(t, c.Workspace(), "skills", []string{".pi/skills"})
	if err := os.MkdirAll(filepath.Join(c.Workspace(), ".pi/skills/demo"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c.Workspace(), ".pi/skills/demo/SKILL.md"), []byte("skill"), 0600); err != nil {
		t.Fatal(err)
	}
	other := scanConfig(t, t.TempDir(), "notes", []string{"notes"})
	if _, err := NewScheduler(t.Context(), store, []searchstore.ScopeConfig{c, other}); err == nil {
		t.Fatal("mixed workspaces accepted")
	}
	if _, err := NewScheduler(t.Context(), store, []searchstore.ScopeConfig{c, c}); err == nil {
		t.Fatal("duplicate scope accepted")
	}
	if _, err := NewScheduler(t.Context(), store, []searchstore.ScopeConfig{{}}); err == nil {
		t.Fatal("unresolved scope accepted")
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var running atomic.Int32
	var order []string // accessed only in scheduler loop, inspected after Close
	w := NewWorker(store)
	run := func(ctx context.Context, config searchstore.ScopeConfig) error {
		if running.Add(1) != 1 {
			t.Error("local workers overlapped")
		}
		defer running.Add(-1)
		order = append(order, config.Scope())
		if config.Scope() == "notes" {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return w.Run(ctx, config)
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c, skills}, run, testSchedulerPolicy)
	notesTicket := request(t, s, "notes")
	awaitSignal(t, entered)
	skillsTicket := request(t, s, "skills")
	if st := workerStatus(t, store, skills); st.State != "never_indexed" {
		t.Fatal("queued scope started", st)
	}
	close(release)
	requireReady(t, notesTicket, 1)
	requireReady(t, skillsTicket, 1)
	s.Close()
	if len(order) != 2 || order[0] != "notes" || order[1] != "skills" {
		t.Fatal(order)
	}
	if st := workerStatus(t, store, other); st.Generation != 0 {
		t.Fatal("workspace leaked", st)
	}
}

func TestSchedulerPeerCommitSatisfiesContentionWithoutDuplicateScan(t *testing.T) {
	_, store, c, path := workerFixture(t)
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := searchstore.NewRefreshStore(other.DB())
	r, err := peer.Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := ScanScope(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	busy, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	w := NewWorker(store)
	run := func(ctx context.Context, c searchstore.ScopeConfig) error {
		calls.Add(1)
		err := w.Run(ctx, c)
		if errors.Is(err, searchstore.ErrRefreshBusy) {
			close(busy)
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return err
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, run, testSchedulerPolicy)
	ticket := request(t, s, c.Scope())
	awaitSignal(t, busy)
	if err := r.Commit(t.Context(), snapshot); err != nil {
		t.Fatal(err)
	}
	close(release)
	requireReady(t, ticket, 1)
	if calls.Load() != 1 {
		t.Fatal("redundant peer scan", calls.Load())
	}
}

func TestSchedulerUnrelatedPeerLeaseDoesNotSatisfyReadyRefresh(t *testing.T) {
	_, store, c, path := workerFixture(t)
	if err := NewWorker(store).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := searchstore.NewRefreshStore(other.DB())
	skills := scanConfig(t, c.Workspace(), "skills", []string{".pi/skills"})
	r, err := peer.Begin(t.Context(), skills, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	busy, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	w := NewWorker(store)
	run := func(ctx context.Context, c searchstore.ScopeConfig) error {
		n := calls.Add(1)
		err := w.Run(ctx, c)
		if n == 1 {
			if !errors.Is(err, searchstore.ErrRefreshBusy) {
				t.Error(err)
			}
			close(busy)
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return err
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, run, testSchedulerPolicy)
	ticket := request(t, s, c.Scope())
	awaitSignal(t, busy)
	if err := r.Fail(t.Context(), errors.New("peer stopped")); err != nil {
		t.Fatal(err)
	}
	close(release)
	requireReady(t, ticket, 2)
	if calls.Load() != 2 {
		t.Fatal("old ready generation incorrectly satisfied demand", calls.Load())
	}
}

func TestSchedulerBoundsBusyLostOwnershipAndContinuousInvalidation(t *testing.T) {
	for _, mode := range []string{"busy", "lost", "events"} {
		t.Run(mode, func(t *testing.T) {
			_, store, c, _ := workerFixture(t)
			w := NewWorker(store)
			if err := w.Run(t.Context(), c); err != nil {
				t.Fatal(err)
			}
			var calls atomic.Int32
			var times []time.Time
			if mode == "events" {
				w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
					snapshot, err := ScanScope(ctx, c)
					if err == nil {
						_, err = store.Invalidate(ctx, c, nil)
					}
					return snapshot, err
				}
			}
			run := func(ctx context.Context, c searchstore.ScopeConfig) error {
				calls.Add(1)
				times = append(times, time.Now())
				switch mode {
				case "busy":
					return searchstore.ErrRefreshBusy
				case "lost":
					return searchstore.ErrRefreshLost
				default:
					return w.Run(ctx, c)
				}
			}
			s := testScheduler(t, store, []searchstore.ScopeConfig{c}, run, testSchedulerPolicy)
			_, err := waitTicket(t, request(t, s, c.Scope()))
			if !errors.Is(err, ErrRefreshPending) {
				t.Fatal("false success", err)
			}
			s.Close()
			if calls.Load() != int32(testSchedulerPolicy.attempts) {
				t.Fatal(calls.Load())
			}
			for i := 1; i < len(times); i++ {
				minimum := min(testSchedulerPolicy.backoff*time.Duration(1<<(i-1)), testSchedulerPolicy.maxBackoff)
				if times[i].Sub(times[i-1]) < minimum {
					t.Fatal("missing backoff", times)
				}
			}
			if mode == "events" {
				st := workerStatus(t, store, c)
				if st.State != "stale" || st.RequestedRevision != 4 || st.AcknowledgedRevision != 3 {
					t.Fatal(st)
				}
			}
		})
	}
}

func TestSchedulerPermanentFailureStopsAndLaterRequestCanRecover(t *testing.T) {
	db, store, c, _ := workerFixture(t)
	w := NewWorker(store)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var fail atomic.Bool
	fail.Store(true)
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		calls.Add(1)
		if fail.Load() {
			return searchstore.CompleteSnapshot{}, errors.New("scan unavailable")
		}
		return ScanScope(ctx, c)
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, w.Run, testSchedulerPolicy)
	if _, err := waitTicket(t, request(t, s, c.Scope())); err == nil {
		t.Fatal("failed scan reported success")
	}
	st := workerStatus(t, store, c)
	if calls.Load() != 1 || st.State != "failed" || st.RequestedRevision != 1 || st.AcknowledgedRevision != 0 || st.Generation != 1 || workerMatches(t, db, "orchid") != 1 {
		t.Fatal(calls.Load(), st)
	}
	fail.Store(false)
	requireReady(t, request(t, s, c.Scope()), 2)
	if calls.Load() != 2 {
		t.Fatal(calls.Load())
	}
}

func TestSchedulerRequestRacingFinalStatusReadCannotLoseInvalidation(t *testing.T) {
	_, store, c, _ := workerFixture(t)
	w := NewWorker(store)
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, w.Run, testSchedulerPolicy)
	read, release := make(chan struct{}), make(chan struct{})
	var reads atomic.Int32
	// Install before any request, while the loop is idle.
	s.readStatus = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.ScopeStatus, error) {
		st, err := store.Status(ctx, c)
		if reads.Add(1) == 2 {
			close(read)
			select {
			case <-release:
			case <-ctx.Done():
				return st, ctx.Err()
			}
		}
		return st, err
	}
	ticket := request(t, s, c.Scope())
	awaitSignal(t, read)
	if _, err := store.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	if next := request(t, s, c.Scope()); next != ticket {
		t.Fatal("expected same active batch")
	}
	close(release)
	st := requireReady(t, ticket, 2)
	if st.AcknowledgedRevision != 1 {
		t.Fatal(st)
	}
}

func TestSchedulerCloseJoinsScannerCleanupAndQueuedTickets(t *testing.T) {
	db, store, c, _ := workerFixture(t)
	w := NewWorker(store)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	skills := scanConfig(t, c.Workspace(), "skills", []string{".pi/skills"})
	entered, cancelled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		close(entered)
		<-ctx.Done()
		close(cancelled)
		<-release
		return searchstore.CompleteSnapshot{}, ctx.Err()
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c, skills}, w.Run, testSchedulerPolicy)
	active := request(t, s, c.Scope())
	awaitSignal(t, entered)
	queued := request(t, s, skills.Scope())
	closed := make(chan struct{})
	go func() { s.Close(); close(closed) }()
	awaitSignal(t, cancelled)
	select {
	case <-closed:
		t.Fatal("Close abandoned scanner")
	default:
	}
	if _, err := s.Request(c.Scope()); !errors.Is(err, ErrSchedulerClosed) {
		t.Fatal(err)
	}
	close(release)
	awaitSignal(t, closed)
	if _, err := waitTicket(t, active); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := waitTicket(t, queued); !errors.Is(err, ErrSchedulerClosed) {
		t.Fatal(err)
	}
	st := workerStatus(t, store, c)
	if st.State != "failed" || st.Generation != 1 || st.RequestedRevision != 1 || st.AcknowledgedRevision != 0 {
		t.Fatal("cleanup not joined", st)
	}
	if st := workerStatus(t, store, skills); st.State != "never_indexed" {
		t.Fatal("queued work ran", st)
	}
	var leases int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_leases").Scan(&leases); err != nil || leases != 0 {
		t.Fatal(leases, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	s.Close() // safe after DB close, idempotent
}

func TestSchedulerParentCancellationDuringBackoff(t *testing.T) {
	_, store, c, _ := workerFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	called := make(chan struct{})
	policy := testSchedulerPolicy
	policy.backoff = time.Hour
	policy.maxBackoff = time.Hour
	s, err := newScheduler(ctx, store, []searchstore.ScopeConfig{c}, func(context.Context, searchstore.ScopeConfig) error { close(called); return searchstore.ErrRefreshBusy }, policy)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ticket := request(t, s, c.Scope())
	awaitSignal(t, called)
	cancel()
	s.Close()
	if _, err := waitTicket(t, ticket); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := s.Request(c.Scope()); !errors.Is(err, ErrSchedulerClosed) {
		t.Fatal(err)
	}
}

func TestSchedulerNativeLostLeaseAcceptsSuccessorPublication(t *testing.T) {
	db, store, c, path := workerFixture(t)
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := searchstore.NewRefreshStore(other.DB())
	w := NewWorker(store)
	w.leaseTTL = time.Second
	entered, cancelled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var scans atomic.Int32
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		scans.Add(1)
		close(entered)
		<-ctx.Done()
		close(cancelled)
		<-release // give successor time to publish before old worker cleanup
		return searchstore.CompleteSnapshot{}, ctx.Err()
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, w.Run, testSchedulerPolicy)
	ticket := request(t, s, c.Scope())
	awaitSignal(t, entered)
	if _, err := db.DB().Exec("UPDATE workspace_index_leases SET expires_at_ms=0"); err != nil {
		t.Fatal(err)
	}
	awaitSignal(t, cancelled)
	if err := NewWorker(peer).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	close(release)
	requireReady(t, ticket, 1)
	if scans.Load() != 1 {
		t.Fatal("successor reindexed", scans.Load())
	}
}

func TestSchedulerDeadlineCancelsAndJoinsNativeWorker(t *testing.T) {
	_, store, c, _ := workerFixture(t)
	w := NewWorker(store)
	policy := testSchedulerPolicy
	policy.timeout = 100 * time.Millisecond
	var exited atomic.Bool
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		<-ctx.Done()
		exited.Store(true)
		return searchstore.CompleteSnapshot{}, ctx.Err()
	}
	s := testScheduler(t, store, []searchstore.ScopeConfig{c}, w.Run, policy)
	if _, err := waitTicket(t, request(t, s, c.Scope())); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if !exited.Load() {
		t.Fatal("scanner abandoned")
	}
	if st := workerStatus(t, store, c); st.State != "failed" || st.Generation != 0 {
		t.Fatal(st)
	}
}

func TestSchedulerStatusFailureAndRequestStormAreBounded(t *testing.T) {
	for _, mode := range []string{"failure", "storm"} {
		t.Run(mode, func(t *testing.T) {
			_, store, c, _ := workerFixture(t)
			s := testScheduler(t, store, []searchstore.ScopeConfig{c}, NewWorker(store).Run, testSchedulerPolicy)
			var reads atomic.Int32
			s.readStatus = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.ScopeStatus, error) {
				n := reads.Add(1)
				st, err := store.Status(ctx, c)
				if n > 1 {
					if mode == "failure" {
						return st, errors.New("status unavailable")
					}
					if _, err := s.Request(c.Scope()); err != nil {
						t.Error(err)
					} // deterministically race every readiness read
				}
				return st, err
			}
			if _, err := waitTicket(t, request(t, s, c.Scope())); err == nil {
				t.Fatal("unverified success")
			}
			s.Close()
			max := int32(2)
			if mode == "storm" {
				max = int32(1 + testSchedulerPolicy.attempts)
			}
			if reads.Load() != max {
				t.Fatal("unbounded readiness reads", reads.Load())
			}
			if st := workerStatus(t, store, c); st.State != "ready" || st.Generation != 1 {
				t.Fatal("status failure changed publication", st)
			}
		})
	}
}
