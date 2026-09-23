package indexer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	core "github.com/rcarmo/gi/internal/store"
)

func workerFixture(t *testing.T) (*core.Store, *searchstore.RefreshStore, searchstore.ScopeConfig, string) {
	t.Helper()
	root, write := scanFixture(t)
	write("notes/a.md", "committed orchid")
	path := filepath.Join(t.TempDir(), "worker.db")
	db, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	c := scanConfig(t, root, "notes", []string{"notes"})
	return db, searchstore.NewRefreshStore(db.DB()), c, path
}
func awaitWorker(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not finish")
		return nil
	}
}
func awaitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not reach scanner")
	}
}
func workerStatus(t *testing.T, s *searchstore.RefreshStore, c searchstore.ScopeConfig) searchstore.ScopeStatus {
	t.Helper()
	st, err := s.Status(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
func workerMatches(t *testing.T, db *core.Store, term string) int {
	t.Helper()
	var n int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH ?", term).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestWorkerNativeScanCommitAndRenewalAcrossTwoStores(t *testing.T) {
	db, s, c, path := workerFixture(t)
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := searchstore.NewRefreshStore(other.DB())
	w := NewWorker(s)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	baseline := workerStatus(t, s, c)
	if baseline.State != "ready" || baseline.Generation != 1 || workerMatches(t, db, "orchid") != 1 {
		t.Fatal(baseline)
	}
	w.leaseTTL = time.Second
	entered, release := make(chan struct{}), make(chan struct{})
	var exited atomic.Bool
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		close(entered)
		select {
		case <-release:
		case <-ctx.Done():
			return searchstore.CompleteSnapshot{}, ctx.Err()
		}
		defer exited.Store(true)
		return ScanScope(ctx, c)
	}
	done := make(chan error, 1)
	go func() { done <- w.Run(t.Context(), c) }()
	awaitSignal(t, entered)
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	var firstExpiry int64
	var token string
	if err := db.DB().QueryRow("SELECT expires_at_ms,owner_token FROM workspace_index_leases").Scan(&firstExpiry, &token); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(4 * time.Second)
	for {
		var current, now int64
		var currentToken string
		if err := db.DB().QueryRow("SELECT expires_at_ms,owner_token,CAST((julianday('now')-2440587.5)*86400000 AS INTEGER) FROM workspace_index_leases").Scan(&current, &currentToken, &now); err != nil {
			t.Fatal(err)
		}
		if now > firstExpiry && current > now && currentToken == token {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("lease did not survive original expiry")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err := NewWorker(peer).Run(t.Context(), c); !errors.Is(err, searchstore.ErrRefreshBusy) {
		t.Fatal("competitor acquired", err)
	}
	if st := workerStatus(t, s, c); st.State != "indexing" || st.Generation != 1 || st.LastIndexedAtMS != baseline.LastIndexedAtMS {
		t.Fatal(st)
	}
	close(release)
	if err := awaitWorker(t, done); err != nil {
		t.Fatal(err)
	}
	if !exited.Load() {
		t.Fatal("scanner still active after return")
	}
	if st := workerStatus(t, s, c); st.State != "ready" || st.Generation != 2 {
		t.Fatal(st)
	}
	var leases int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_leases").Scan(&leases); err != nil || leases != 0 {
		t.Fatal(leases, err)
	}
}

func TestWorkerCancellationTimeoutAndNativeFailureRetainSnapshot(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline", "scan", "commit", "incomplete"} {
		t.Run(mode, func(t *testing.T) {
			db, s, c, _ := workerFixture(t)
			w := NewWorker(s)
			if err := w.Run(t.Context(), c); err != nil {
				t.Fatal(err)
			}
			baseline := workerStatus(t, s, c)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch mode {
			case "cancel", "deadline":
				entered := make(chan struct{})
				if mode == "deadline" {
					w.runTimeout = 100 * time.Millisecond
				}
				w.scan = func(ctx context.Context, _ searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
					close(entered)
					<-ctx.Done()
					return searchstore.CompleteSnapshot{}, ctx.Err()
				}
				done := make(chan error, 1)
				go func() { done <- w.Run(ctx, c) }()
				awaitSignal(t, entered)
				if mode == "cancel" {
					cancel()
				}
				err := awaitWorker(t, done)
				want := context.Canceled
				if mode == "deadline" {
					want = context.DeadlineExceeded
				}
				if !errors.Is(err, want) {
					t.Fatal(err)
				}
			case "scan":
				if err := os.RemoveAll(filepath.Join(c.Workspace(), "notes")); err != nil {
					t.Fatal(err)
				}
				if err := w.Run(ctx, c); err == nil {
					t.Fatal("scan failure absent")
				}
			case "commit":
				if err := os.WriteFile(filepath.Join(c.Workspace(), "notes/a.md"), []byte("replacement iris"), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := db.DB().Exec("CREATE TRIGGER reject_worker_chunk BEFORE INSERT ON workspace_index_chunks BEGIN SELECT raise(abort,'worker write failed'); END"); err != nil {
					t.Fatal(err)
				}
				if err := w.Run(ctx, c); err == nil || !strings.Contains(err.Error(), "worker write failed") {
					t.Fatal(err)
				}
			case "incomplete":
				w.scan = func(context.Context, searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
					return searchstore.CompleteSnapshot{}, nil
				}
				if err := w.Run(ctx, c); !errors.Is(err, searchstore.ErrIncompleteSnapshot) {
					t.Fatal(err)
				}
			}
			failed := workerStatus(t, s, c)
			if failed.State != "failed" || failed.Generation != baseline.Generation || failed.IndexedFileCount != 1 || failed.LastIndexedAtMS != baseline.LastIndexedAtMS || failed.LastError == "" {
				t.Fatal(failed)
			}
			if workerMatches(t, db, "orchid") != 1 || workerMatches(t, db, "iris") != 0 {
				t.Fatal("failure changed committed content")
			}
		})
	}
}

func TestWorkerRenewalLossCancelsScanWithoutChangingSuccessor(t *testing.T) {
	db, s, c, path := workerFixture(t)
	w := NewWorker(s)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := searchstore.NewRefreshStore(other.DB())
	entered := make(chan struct{})
	w.leaseTTL = time.Second
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		close(entered)
		<-ctx.Done()
		return searchstore.CompleteSnapshot{}, ctx.Err()
	}
	done := make(chan error, 1)
	go func() { done <- w.Run(t.Context(), c) }()
	awaitSignal(t, entered)
	if _, err := db.DB().Exec("UPDATE workspace_index_leases SET expires_at_ms=0"); err != nil {
		t.Fatal(err)
	}
	successor, err := peer.Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := awaitWorker(t, done); !errors.Is(err, searchstore.ErrRefreshLost) {
		t.Fatal(err)
	}
	if st := workerStatus(t, peer, c); st.State != "indexing" || st.LastError != "" || st.Generation != 1 {
		t.Fatal("old worker changed successor", st)
	}
	snap, err := ScanScope(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if err := successor.Commit(t.Context(), snap); err != nil {
		t.Fatal(err)
	}
	if workerStatus(t, peer, c).Generation != 2 {
		t.Fatal("successor did not commit")
	}
}

func TestWorkerFinalRenewalAndCleanupErrors(t *testing.T) {
	db, s, c, _ := workerFixture(t)
	w := NewWorker(s)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		snap, err := ScanScope(ctx, c)
		if err != nil {
			return snap, err
		}
		_, err = db.DB().Exec("UPDATE workspace_index_leases SET expires_at_ms=0")
		return snap, err
	}
	if err := w.Run(t.Context(), c); !errors.Is(err, searchstore.ErrRefreshLost) {
		t.Fatal(err)
	}
	if st := workerStatus(t, s, c); st.State != "stale" || st.Generation != 1 {
		t.Fatal(st)
	}
	w = NewWorker(s)
	if _, err := db.DB().Exec("CREATE TRIGGER reject_failure BEFORE UPDATE OF state ON workspace_index_scopes WHEN new.state='failed' BEGIN SELECT raise(abort,'cleanup rejected'); END"); err != nil {
		t.Fatal(err)
	}
	w.scan = func(context.Context, searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		return searchstore.CompleteSnapshot{}, fmt.Errorf("original scan error")
	}
	if err := w.Run(t.Context(), c); err == nil || !strings.Contains(err.Error(), "original scan error") || !strings.Contains(err.Error(), "cleanup rejected") {
		t.Fatal(err)
	}
	if workerMatches(t, db, "orchid") != 1 {
		t.Fatal("failed cleanup lost data")
	}
}

// This helper runs only in a child process, holds a real database lease, and is
// killed by the parent to test recovery without Go defers or cleanup callbacks.
func TestWorkerCrashHelper(t *testing.T) {
	path := os.Getenv("GI_INDEX_WORKER_TEST_DB")
	if path == "" {
		t.Skip("child only")
	}
	root := os.Getenv("GI_INDEX_WORKER_TEST_ROOT")
	ready := os.Getenv("GI_INDEX_WORKER_TEST_READY")
	db, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	c, err := searchstore.DefaultScopeConfig(root, "notes", nil, nil, chunking.LineVersion)
	if err != nil {
		t.Fatal(err)
	}
	w := NewWorker(searchstore.NewRefreshStore(db.DB()))
	w.leaseTTL = time.Second
	w.scan = func(ctx context.Context, _ searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		if err := os.WriteFile(ready, []byte("owned"), 0600); err != nil {
			return searchstore.CompleteSnapshot{}, err
		}
		<-ctx.Done()
		return searchstore.CompleteSnapshot{}, ctx.Err()
	}
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerProcessCrashExpiresAndRecovers(t *testing.T) {
	db, s, c, path := workerFixture(t)
	if err := NewWorker(s).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	ready := filepath.Join(t.TempDir(), "ready")
	child := exec.Command(os.Args[0], "-test.run=^TestWorkerCrashHelper$", "-test.timeout=15s")
	child.Env = append(os.Environ(), "GI_INDEX_WORKER_TEST_DB="+path, "GI_INDEX_WORKER_TEST_ROOT="+c.Workspace(), "GI_INDEX_WORKER_TEST_READY="+ready)
	var output strings.Builder
	child.Stdout = &output
	child.Stderr = &output
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if child.ProcessState == nil {
			child.Process.Kill()
			child.Wait()
		}
	}()
	deadline := time.Now().Add(8 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child did not own lease")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if st := workerStatus(t, s, c); st.State != "indexing" {
		t.Fatal(st)
	}
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := child.Wait(); err == nil {
		t.Fatal("child unexpectedly exited cleanly")
	}
	deadline = time.Now().Add(5 * time.Second)
	for workerStatus(t, s, c).State != "stale" {
		if time.Now().After(deadline) {
			t.Fatal("crashed lease did not expire", output.String())
		}
		time.Sleep(20 * time.Millisecond)
	}
	if workerMatches(t, db, "orchid") != 1 {
		t.Fatal("crash erased snapshot")
	}
	if err := NewWorker(s).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	if st := workerStatus(t, s, c); st.State != "ready" || st.Generation != 2 || st.IndexedFileCount != 1 {
		t.Fatal(st)
	}
}

func TestWorkerEntryAndPublicationCancellation(t *testing.T) {
	db, s, c, _ := workerFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	w := NewWorker(s)
	var scanned atomic.Bool
	w.scan = func(context.Context, searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		scanned.Store(true)
		return searchstore.CompleteSnapshot{}, nil
	}
	if err := w.Run(ctx, c); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if scanned.Load() || workerStatus(t, s, c).State != "never_indexed" {
		t.Fatal("cancelled acquisition started a scan")
	}
	if err := NewWorker(s).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	baseline := workerStatus(t, s, c)
	ctx, cancel = context.WithCancel(t.Context())
	defer cancel()
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		snap, err := ScanScope(ctx, c)
		cancel()
		return snap, err
	}
	if err := w.Run(ctx, c); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	failed := workerStatus(t, s, c)
	if failed.Generation != baseline.Generation || failed.LastIndexedAtMS != baseline.LastIndexedAtMS || failed.State != "failed" || workerMatches(t, db, "orchid") != 1 {
		t.Fatal(failed)
	}
	// Explicit retry works only after cancellation cleanup; there is no automatic rerun.
	if err := NewWorker(s).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	if workerStatus(t, s, c).Generation != 2 {
		t.Fatal("retry did not commit once")
	}
	if err := NewWorker(nil).Run(t.Context(), c); err == nil {
		t.Fatal("nil store accepted")
	}
}

func TestWorkerNativeRenewalErrorStopsScanAndReleasesOwnership(t *testing.T) {
	db, s, c, _ := workerFixture(t)
	w := NewWorker(s)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	w.leaseTTL = time.Second
	entered := make(chan struct{})
	w.scan = func(ctx context.Context, _ searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		close(entered)
		<-ctx.Done()
		return searchstore.CompleteSnapshot{}, ctx.Err()
	}
	done := make(chan error, 1)
	go func() { done <- w.Run(t.Context(), c) }()
	awaitSignal(t, entered)
	// Actual renewal statement fails; acquisition already succeeded and cleanup's
	// fenced owner-token check is a different UPDATE, so it can persist failure.
	if _, err := db.DB().Exec("CREATE TRIGGER reject_renewal BEFORE UPDATE OF expires_at_ms ON workspace_index_leases BEGIN SELECT raise(abort,'renewal rejected'); END"); err != nil {
		t.Fatal(err)
	}
	err := awaitWorker(t, done)
	if err == nil || !strings.Contains(err.Error(), "renewal rejected") {
		t.Fatal(err)
	}
	failed := workerStatus(t, s, c)
	if failed.State != "failed" || failed.Generation != 1 || !strings.Contains(failed.LastError, "renewal rejected") {
		t.Fatal(failed)
	}
	var leases int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_leases").Scan(&leases); err != nil || leases != 0 {
		t.Fatal(leases, err)
	}
	if workerMatches(t, db, "orchid") != 1 {
		t.Fatal("renewal error changed committed data")
	}
}

func TestWorkerEventAfterScanRemainsPendingUntilNextRefresh(t *testing.T) {
	db, s, c, _ := workerFixture(t)
	w := NewWorker(s)
	if err := w.Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	scanned, release := make(chan struct{}), make(chan struct{})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	w.scan = func(ctx context.Context, c searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
		snapshot, err := ScanScope(ctx, c)
		if err != nil {
			return snapshot, err
		}
		close(scanned)
		select {
		case <-release:
			return snapshot, nil
		case <-ctx.Done():
			return searchstore.CompleteSnapshot{}, ctx.Err()
		}
	}
	result := make(chan error, 1)
	go func() { result <- w.Run(ctx, c) }()
	awaitSignal(t, scanned)
	if err := os.WriteFile(filepath.Join(c.Workspace(), "notes/a.md"), []byte("new dahlia"), 0600); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.Invalidate(t.Context(), c, []string{"notes/a.md"}); err != nil || !ok {
		t.Fatal(ok, err)
	}
	close(release)
	if err := awaitWorker(t, result); err != nil {
		t.Fatal(err)
	}
	stale := workerStatus(t, s, c)
	if stale.State != "stale" || stale.RequestedRevision != 1 || stale.AcknowledgedRevision != 0 || stale.Generation != 2 {
		t.Fatal(stale)
	}
	if workerMatches(t, db, "orchid") != 1 {
		t.Fatal("captured snapshot unexpectedly changed")
	}
	if err := NewWorker(s).Run(t.Context(), c); err != nil {
		t.Fatal(err)
	}
	ready := workerStatus(t, s, c)
	if ready.State != "ready" || ready.RequestedRevision != 1 || ready.AcknowledgedRevision != 1 || ready.Generation != 3 {
		t.Fatal(ready)
	}
	if workerMatches(t, db, "dahlia") != 1 || workerMatches(t, db, "orchid") != 0 {
		t.Fatal("follow-up refresh wrong")
	}
}
