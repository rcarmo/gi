package store_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	search "github.com/rcarmo/gi/internal/search/store"
	core "github.com/rcarmo/gi/internal/store"
)

func TestInvalidationDuringScanRetainsNewerRevisionAndSnapshot(t *testing.T) {
	db, s, root, _ := dbFixture(t)
	c := config(t, root, "notes")
	d := document("notes/a.md", "old orchid")
	commit(t, begin(t, s, c), d)
	previous := status(t, s, c)
	changed, err := s.Invalidate(t.Context(), c, []string{"notes/a.md"})
	if err != nil || !changed {
		t.Fatal(changed, err)
	}
	st := status(t, s, c)
	if st.State != "stale" || st.RequestedRevision != 1 || st.AcknowledgedRevision != 0 || st.Generation != previous.Generation || st.LastIndexedAtMS != previous.LastIndexedAtMS {
		t.Fatal(st)
	}
	r := begin(t, s, c)
	for i := 0; i < 3; i++ {
		if ok, err := s.Invalidate(t.Context(), c, []string{"notes/a.md"}); err != nil || !ok {
			t.Fatal(ok, err)
		}
	}
	if st := status(t, s, c); st.State != "indexing" || st.RequestedRevision != 4 {
		t.Fatal(st)
	}
	commit(t, r, document("notes/a.md", "new violet"))
	st = status(t, s, c)
	if st.State != "stale" || st.RequestedRevision != 4 || st.AcknowledgedRevision != 1 || st.Generation != 2 {
		t.Fatal(st)
	}
	if matches(t, db, "violet") != 1 || matches(t, db, "orchid") != 0 {
		t.Fatal("commit content wrong")
	}
	commit(t, begin(t, s, c), document("notes/a.md", "new violet"))
	if st := status(t, s, c); st.State != "ready" || st.RequestedRevision != 4 || st.AcknowledgedRevision != 4 || st.Generation != 3 {
		t.Fatal(st)
	}
	// An event after release must survive as a fresh unacknowledged request.
	if _, err := s.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	if st := status(t, s, c); st.State != "stale" || st.RequestedRevision != 5 || st.AcknowledgedRevision != 4 {
		t.Fatal(st)
	}
}

func TestInvalidationScopePathAndConfigurationIsolation(t *testing.T) {
	_, s, root, _ := dbFixture(t)
	notes, skills, all := config(t, root, "notes"), config(t, root, "skills"), config(t, root, "all")
	n, sk := document("notes/a.md", "note"), document(".pi/skills/demo/SKILL.md", "skill")
	for _, c := range []search.ScopeConfig{notes, all} {
		commit(t, begin(t, s, c), n)
	}
	commit(t, begin(t, s, skills), sk)
	for _, c := range []search.ScopeConfig{notes, skills, all} {
		hit, err := s.Invalidate(t.Context(), c, []string{"notes"})
		if err != nil || hit != (c.Scope() != "skills") {
			t.Fatal(c.Scope(), hit, err)
		}
	}
	if status(t, s, skills).State != "ready" {
		t.Fatal("unrelated scope invalidated")
	}
	if hit, err := s.Invalidate(t.Context(), skills, []string{".pi"}); err != nil || !hit {
		t.Fatal("ancestor ignored", hit, err)
	}
	foreign := config(t, t.TempDir(), "notes")
	if status(t, s, foreign).RequestedRevision != 0 {
		t.Fatal("workspace leak")
	}
	if hit, err := s.Invalidate(t.Context(), foreign, []string{"other/a.md"}); err != nil || hit {
		t.Fatal(hit, err)
	}
	if hit, err := s.Invalidate(t.Context(), foreign, []string{"notes/new.md"}); err != nil || !hit {
		t.Fatal(hit, err)
	}
	if st := status(t, s, foreign); st.State != "never_indexed" || st.RequestedRevision != 1 || st.Generation != 0 {
		t.Fatal(st)
	}
	changed, err := search.NewScopeConfig(root, "notes", []string{"docs"}, []string{"md"}, "lines-v1")
	if err != nil {
		t.Fatal(err)
	}
	if hit, err := s.Invalidate(t.Context(), changed, []string{"notes/deleted.md"}); err != nil || !hit {
		t.Fatal("old committed root ignored", hit, err)
	}
	for _, p := range []string{"../outside", "/abs", "notes/../outside", ""} {
		if _, err := s.Invalidate(t.Context(), notes, []string{p}); err == nil {
			t.Fatal("bad event accepted", p)
		}
	}
	if _, err := s.Invalidate(t.Context(), notes, make([]string, 10001)); err == nil {
		t.Fatal("unbounded events")
	}
}

func TestInvalidationFailureAndReopenNeverAcknowledgePendingWork(t *testing.T) {
	db, s, root, path := dbFixture(t)
	c := config(t, root, "notes")
	commit(t, begin(t, s, c), document("notes/a.md", "kept orchid"))
	if _, err := s.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	r := begin(t, s, c)
	if _, err := s.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().Exec("CREATE TRIGGER fail_invalid_chunk BEFORE INSERT ON workspace_index_chunks BEGIN SELECT raise(abort,'failed write'); END"); err != nil {
		t.Fatal(err)
	}
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{document("notes/a.md", "new violet")}}); err == nil {
		t.Fatal("commit should fail")
	}
	if err := r.Fail(t.Context(), errors.New("failed write")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	failed := status(t, s, c)
	if failed.State != "failed" || failed.LastError != "failed write" || failed.RequestedRevision != 3 || failed.AcknowledgedRevision != 0 {
		t.Fatal(failed)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := s.Invalidate(ctx, c, nil); err == nil {
		t.Fatal("cancel ignored")
	}
	second, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	peer := search.NewRefreshStore(second.DB())
	if st := status(t, peer, c); st.RequestedRevision != 3 || st.AcknowledgedRevision != 0 || st.Generation != 1 {
		t.Fatal(st)
	}
	if matches(t, db, "orchid") != 1 {
		t.Fatal("failure lost content")
	}
	if _, err := db.DB().Exec("DROP TRIGGER fail_invalid_chunk"); err != nil {
		t.Fatal(err)
	}
	commit(t, begin(t, peer, c), document("notes/a.md", "new violet"))
	if st := status(t, peer, c); st.State != "ready" || st.AcknowledgedRevision != 3 {
		t.Fatal(st)
	}
}

func TestInvalidationTwoStoresAndOverlappingScopeRevisions(t *testing.T) {
	db, s, root, path := dbFixture(t)
	c, all := config(t, root, "notes"), config(t, root, "all")
	d := document("notes/a.md", "orchid")
	commit(t, begin(t, s, c), d)
	commit(t, begin(t, s, all), d)
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := search.NewRefreshStore(other.DB())
	const count = 20
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store := s
			if i%2 == 1 {
				store = peer
			}
			_, err := store.Invalidate(t.Context(), c, []string{"notes/a.md"})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if st := status(t, s, c); st.RequestedRevision != count {
		t.Fatal(st)
	}
	// Both orderings of refresh publication vs mutation preserve the event:
	// either it is captured by Begin, or it remains pending after publication.
	r, err := peer.Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Invalidate(t.Context(), c, nil); err != nil {
		t.Fatal(err)
	}
	commit(t, r, document("notes/a.md", "new rose"))
	if st := status(t, s, c); st.AcknowledgedRevision != count || st.RequestedRevision != count+1 || st.State != "stale" {
		t.Fatal(st)
	}
	if st := status(t, s, all); st.State != "stale" || st.RequestedRevision != 1 || st.AcknowledgedRevision != 0 {
		t.Fatal("shared edit not durable invalidation", st)
	}
	if _, err := db.DB().Exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)"); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidationIsAtomicAndConcurrentCommitCannotClearIt(t *testing.T) {
	db, s, root, path := dbFixture(t)
	c := config(t, root, "notes")
	d := document("notes/a.md", "retained")
	commit(t, begin(t, s, c), d)
	if _, err := db.DB().Exec(`CREATE TRIGGER reject_stale BEFORE UPDATE OF state ON workspace_index_scopes WHEN new.state='stale' BEGIN SELECT raise(abort,'invalidation failure'); END`); err != nil {
		t.Fatal(err)
	}
	if changed, err := s.Invalidate(t.Context(), c, nil); err == nil || changed {
		t.Fatal("partial invalidation accepted", changed, err)
	}
	if st := status(t, s, c); st.RequestedRevision != 0 || st.State != "ready" {
		t.Fatal("failed invalidation leaked revision", st)
	}
	if _, err := db.DB().Exec("DROP TRIGGER reject_stale"); err != nil {
		t.Fatal(err)
	}
	other, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	peer := search.NewRefreshStore(other.DB())
	for n := int64(1); n <= 10; n++ {
		r := begin(t, s, c)
		start := make(chan struct{})
		done := make(chan error, 2)
		go func() {
			<-start
			done <- r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{d}})
		}()
		go func() { <-start; _, err := peer.Invalidate(t.Context(), c, []string{"notes/a.md"}); done <- err }()
		close(start)
		for i := 0; i < 2; i++ {
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		}
		if st := status(t, s, c); st.State != "stale" || st.RequestedRevision != n || st.AcknowledgedRevision != n-1 {
			t.Fatal("concurrent event lost", st)
		}
	}
}
