package store_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/search/chunking"
	search "github.com/rcarmo/gi/internal/search/store"
	core "github.com/rcarmo/gi/internal/store"
)

func document(path, text string) search.RefreshDocument {
	return search.RefreshDocument{Path: path, Content: text, MtimeNS: 100, Chunks: []chunking.Chunk{{ChunkIndex: 0, StartByte: 0, EndByte: len(text), StartLine: 1, EndLine: 1 + strings.Count(text, "\n"), Content: text}}}
}
func dbFixture(t *testing.T) (*core.Store, *search.RefreshStore, string, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	s, err := core.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s, search.NewRefreshStore(s.DB()), t.TempDir(), dbPath
}
func config(t *testing.T, root, scope string) search.ScopeConfig {
	t.Helper()
	c, err := search.DefaultScopeConfig(root, scope, nil, nil, "lines-v1")
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func begin(t *testing.T, s *search.RefreshStore, c search.ScopeConfig) *search.Refresh {
	t.Helper()
	r, err := s.Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func commit(t *testing.T, r *search.Refresh, docs ...search.RefreshDocument) {
	t.Helper()
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: docs}); err != nil {
		t.Fatal(err)
	}
}
func status(t *testing.T, s *search.RefreshStore, c search.ScopeConfig) search.ScopeStatus {
	t.Helper()
	st, err := s.Status(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
func matches(t *testing.T, s *core.Store, term string) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH ?", term).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRefreshConfigAndScopeOwnership(t *testing.T) {
	root := t.TempDir()
	c, err := search.NewScopeConfig(root, "custom", []string{"notes/sub", "notes", "notes", ".pi/skills"}, []string{"MD", ".txt", ".md"}, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c.Roots(), []string{".pi/skills", "notes"}) {
		t.Fatal(c.Roots())
	}
	changed := c.Roots()
	changed[0] = "outside"
	if c.Roots()[0] == "outside" {
		t.Fatal("mutable config")
	}
	same, err := search.NewScopeConfig(root, "custom", []string{".pi/skills", "notes"}, []string{".md", ".txt"}, "v1")
	if err != nil || same.Fingerprint() != c.Fingerprint() {
		t.Fatal(err, "unstable fingerprint")
	}
	for _, path := range []string{"../notes", "/etc", "notes/../outside", "vfs://x", "notes\\x"} {
		if _, err := search.NewScopeConfig(root, "bad", []string{path}, []string{"md"}, "v1"); err == nil {
			t.Fatal(path)
		}
	}
	skills := config(t, root, "skills")
	if !reflect.DeepEqual(skills.Roots(), []string{".pi/skills"}) {
		t.Fatal(skills.Roots())
	}
	all, err := search.DefaultScopeConfig(root, "all", []string{"src"}, []string{"nim"}, "v1")
	if err != nil || len(all.Roots()) != 3 {
		t.Fatal(all.Roots(), err)
	}
}

func TestRefreshIncrementalScopesAndHashCollision(t *testing.T) {
	db, s, root, _ := dbFixture(t)
	notes, all, skills := config(t, root, "notes"), config(t, root, "all"), config(t, root, "skills")
	n := document("notes/first.md", "orchid original")
	sk := document(".pi/skills/tool/SKILL.md", "skill cosmos")
	commit(t, begin(t, s, notes), n)
	commit(t, begin(t, s, all), n, sk)
	commit(t, begin(t, s, skills), sk)
	var documentID, chunkID int64
	if err := db.DB().QueryRow("SELECT d.id,c.id FROM workspace_index_documents d JOIN workspace_index_chunks c ON c.document_id=d.id WHERE d.path=?", n.Path).Scan(&documentID, &chunkID); err != nil {
		t.Fatal(err)
	}
	commit(t, begin(t, s, notes), n)
	var afterDoc, afterChunk int64
	if err := db.DB().QueryRow("SELECT d.id,c.id FROM workspace_index_documents d JOIN workspace_index_chunks c ON c.document_id=d.id WHERE d.path=?", n.Path).Scan(&afterDoc, &afterChunk); err != nil || afterDoc != documentID || afterChunk != chunkID {
		t.Fatal("unchanged IDs differ", err)
	}
	previous := status(t, s, notes)
	// Same size/mtime, different bytes: commit hashes actual content.
	n = document(n.Path, "dahlia revised")
	commit(t, begin(t, s, notes), n)
	if matches(t, db, "orchid") != 0 || matches(t, db, "dahlia") != 1 {
		t.Fatal("hash verification failed")
	}
	if status(t, s, all).State != "stale" || status(t, s, skills).State != "ready" {
		t.Fatal("scope invalidation incorrect")
	}
	if status(t, s, notes).Generation != previous.Generation+1 {
		t.Fatal("generation")
	}
	commit(t, begin(t, s, notes)) // all still owns the document
	if matches(t, db, "dahlia") != 1 {
		t.Fatal("overlapping content lost")
	}
	commit(t, begin(t, s, all), sk)
	if matches(t, db, "dahlia") != 0 || matches(t, db, "cosmos") != 1 {
		t.Fatal("orphan cleanup incorrect")
	}
	if status(t, s, skills).State != "ready" {
		t.Fatal("unrelated scope changed")
	}
	if _, err := db.DB().Exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)"); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshRollbackValidationAndConfigChange(t *testing.T) {
	db, s, root, _ := dbFixture(t)
	c := config(t, root, "notes")
	baseline := document("notes/kept.md", "retained violet")
	commit(t, begin(t, s, c), baseline)
	before := status(t, s, c)
	r := begin(t, s, c)
	if err := r.Commit(t.Context(), search.CompleteSnapshot{}); !errors.Is(err, search.ErrIncompleteSnapshot) {
		t.Fatal(err)
	}
	for _, doc := range []search.RefreshDocument{document("outside.md", "outside"), document("notes/x.bin", "outside"), document("notes/../../bad.md", "outside"), document("notes/bad.md", string([]byte{0xff}))} {
		if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{doc}}); err == nil {
			t.Fatal("invalid candidate accepted")
		}
	}
	bad := document("notes/new.md", "unicode 界")
	bad.Chunks[0].StartByte = 1
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{bad}}); err == nil {
		t.Fatal("bad chunk accepted")
	}
	if _, err := db.DB().Exec("CREATE TRIGGER reject_chunks BEFORE INSERT ON workspace_index_chunks BEGIN SELECT raise(abort,'injected chunk failure'); END"); err != nil {
		t.Fatal(err)
	}
	err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{document(baseline.Path, "replacement iris")}})
	if err == nil {
		t.Fatal("write failure absent")
	}
	if matches(t, db, "violet") != 1 || matches(t, db, "iris") != 0 {
		t.Fatal("rollback lost committed FTS")
	}
	if err := r.Fail(t.Context(), err); err != nil {
		t.Fatal(err)
	}
	failed := status(t, s, c)
	if failed.State != "failed" || failed.LastError == "" || failed.Generation != before.Generation || failed.LastIndexedAtMS != before.LastIndexedAtMS || failed.IndexedFileCount != 1 {
		t.Fatal(failed)
	}
	if _, err := db.DB().Exec("DROP TRIGGER reject_chunks"); err != nil {
		t.Fatal(err)
	}
	revised, err := search.NewScopeConfig(root, "notes", []string{"notes/sub"}, []string{"md"}, "lines-v2")
	if err != nil {
		t.Fatal(err)
	}
	if st := status(t, s, revised); st.State == "ready" {
		t.Fatal(st)
	}
	pending := begin(t, s, revised)
	if st := status(t, s, revised); st.ConfigHash != before.ConfigHash {
		t.Fatal("begin overwrote committed config")
	}
	commit(t, pending, document("notes/sub/new.md", "new cosmos"))
	if st := status(t, s, revised); st.State != "ready" || st.ConfigHash != revised.Fingerprint() || st.Generation != before.Generation+1 {
		t.Fatal(st)
	}
	if matches(t, db, "violet") != 0 {
		t.Fatal("changed roots did not clean membership")
	}
	cancel, cancelFn := context.WithCancel(t.Context())
	next := begin(t, s, revised)
	cancelFn()
	if err := next.Commit(cancel, search.CompleteSnapshot{Complete: true}); err == nil {
		t.Fatal("cancel committed")
	}
	if err := next.Fail(t.Context(), cancel.Err()); err != nil {
		t.Fatal(err)
	}
	if matches(t, db, "cosmos") != 1 {
		t.Fatal("cancel erased snapshot")
	}
}

func TestRefreshTwoStoresTakeoverAndFinalFence(t *testing.T) {
	db, a, root, path := dbFixture(t)
	second, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	b := search.NewRefreshStore(second.DB())
	c := config(t, root, "notes")
	var wg sync.WaitGroup
	type result struct {
		r   *search.Refresh
		err error
	}
	ch := make(chan result, 2)
	for _, s := range []*search.RefreshStore{a, b} {
		wg.Add(1)
		go func() { defer wg.Done(); r, err := s.Begin(t.Context(), c, time.Minute); ch <- result{r, err} }()
	}
	wg.Wait()
	close(ch)
	var winner *search.Refresh
	busy := 0
	for out := range ch {
		if out.err == nil {
			winner = out.r
		} else if errors.Is(out.err, search.ErrRefreshBusy) {
			busy++
		} else {
			t.Fatal(out.err)
		}
	}
	if winner == nil || busy != 1 {
		t.Fatal("multiple owners", busy)
	}
	if err := winner.Renew(t.Context(), time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Begin(t.Context(), config(t, root, "skills"), time.Minute); !errors.Is(err, search.ErrRefreshBusy) {
		t.Fatal("scope bypassed workspace lease", err)
	}
	if _, err := db.DB().Exec("UPDATE workspace_index_leases SET expires_at_ms=0"); err != nil {
		t.Fatal(err)
	}
	if status(t, b, c).State != "stale" {
		t.Fatal("expired status not stale")
	}
	successor := begin(t, b, c)
	if err := winner.Renew(t.Context(), time.Minute); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal(err)
	}
	if err := winner.Fail(t.Context(), errors.New("old owner")); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal(err)
	}
	if err := winner.Commit(t.Context(), search.CompleteSnapshot{Complete: true}); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal(err)
	}
	commit(t, successor, document("notes/live.md", "live owner"))
	// Expiry introduced after writes but before the final fence must roll back.
	r := begin(t, a, c)
	if _, err := db.DB().Exec("CREATE TRIGGER expire_after_insert AFTER INSERT ON workspace_index_chunks BEGIN UPDATE workspace_index_leases SET expires_at_ms=0; END"); err != nil {
		t.Fatal(err)
	}
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{document("notes/live.md", "forbidden write")}}); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal(err)
	}
	if matches(t, db, "owner") != 1 || matches(t, db, "forbidden") != 0 {
		t.Fatal("final fence lost snapshot")
	}
	if _, err := db.DB().Exec("DROP TRIGGER expire_after_insert"); err != nil {
		t.Fatal(err)
	}
	if err := r.Fail(t.Context(), errors.New("fixture cancelled")); err != nil {
		t.Fatal(err)
	}
	third, err := core.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer third.Close()
	if st := status(t, search.NewRefreshStore(third.DB()), c); st.State != "failed" || st.Generation != 1 || st.IndexedFileCount != 1 {
		t.Fatal(st)
	}
}

func TestRefreshExpiredRestartRecoveryAndLimits(t *testing.T) {
	db, s, root, dbPath := dbFixture(t)
	c := config(t, root, "notes")
	before := status(t, s, c)
	if before.State != "never_indexed" || before.Generation != 0 {
		t.Fatal(before)
	}
	if _, err := s.Begin(t.Context(), c, time.Millisecond); err == nil {
		t.Fatal("short TTL accepted")
	}
	if _, err := s.Begin(t.Context(), c, time.Hour); err == nil {
		t.Fatal("long TTL accepted")
	}
	r := begin(t, s, c)
	if err := r.Renew(t.Context(), 0); err == nil {
		t.Fatal("invalid renewal")
	}
	if err := r.Fail(t.Context(), nil); err == nil {
		t.Fatal("nil failure")
	}
	duplicate := document("notes/duplicate.md", "retained")
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{duplicate, duplicate}}); err == nil {
		t.Fatal("duplicate paths accepted")
	}
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: make([]search.RefreshDocument, search.MaxSnapshotFiles+1)}); err == nil {
		t.Fatal("file bound")
	}
	oversize := document("notes/big.md", strings.Repeat("x", search.MaxDocumentBytes+1))
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{oversize}}); err == nil {
		t.Fatal("byte bound")
	}
	boundary := document("notes/boundary.md", "界")
	boundary.Chunks[0] = chunking.Chunk{StartByte: 1, EndByte: 1, StartLine: 1, EndLine: 1, Content: ""}
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{boundary}}); err == nil {
		t.Fatal("empty chunk within codepoint accepted")
	}
	many := document("notes/many.md", "x")
	many.Chunks = make([]chunking.Chunk, search.MaxSnapshotChunks+1)
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{many}}); err == nil {
		t.Fatal("chunk bound")
	}
	overlap := document("notes/overlap.md", strings.Repeat("x", search.MaxDocumentBytes))
	overlap.Chunks = nil
	for n := 0; n < 65; n++ {
		overlap.Chunks = append(overlap.Chunks, chunking.Chunk{ChunkIndex: n, EndByte: len(overlap.Content), StartLine: 1, EndLine: 1, Content: overlap.Content})
	}
	if err := r.Commit(t.Context(), search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{overlap}}); err == nil {
		t.Fatal("overlapping text amplification accepted")
	}
	commit(t, r, document("notes/unicode.md", "hello\n世界\n"))
	old := begin(t, s, c)
	if _, err := db.DB().Exec("UPDATE workspace_index_leases SET expires_at_ms=0"); err != nil {
		t.Fatal(err)
	}
	if err := old.Renew(t.Context(), time.Minute); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal("expired lease resurrected", err)
	}
	if err := old.Commit(t.Context(), search.CompleteSnapshot{Complete: true}); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal(err)
	}
	if err := old.Fail(t.Context(), errors.New("expired")); !errors.Is(err, search.ErrRefreshLost) {
		t.Fatal(err)
	}
	reopened, err := core.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	other := search.NewRefreshStore(reopened.DB())
	if st := status(t, other, c); st.State != "stale" || st.Generation != 1 || st.IndexedFileCount != 1 {
		t.Fatal(st)
	}
	successor := begin(t, other, config(t, root, "skills"))
	var state string
	if err := db.DB().QueryRow("SELECT state FROM workspace_index_scopes WHERE scope='notes'").Scan(&state); err != nil || state != "stale" {
		t.Fatal("recovery not durable", state, err)
	}
	commit(t, successor, document(".pi/skills/test/SKILL.md", "skill enabled"))
	if matches(t, db, "hello") != 1 {
		t.Fatal("recovery erased prior data")
	}
	foreign := config(t, t.TempDir(), "notes")
	commit(t, begin(t, s, foreign), document("notes/unicode.md", "other workspace"))
	if st := status(t, s, c); st.IndexedFileCount != 1 || st.Generation != 1 {
		t.Fatal("workspace status leaked", st)
	}
	if _, err := db.DB().Exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)"); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshHandleCommitsAtMostOnce(t *testing.T) {
	db, s, root, _ := dbFixture(t)
	c := config(t, root, "notes")
	r := begin(t, s, c)
	result := make(chan error, 2)
	var wg sync.WaitGroup
	snapshot := search.CompleteSnapshot{Complete: true, Documents: []search.RefreshDocument{document("notes/once.md", "single generation")}}
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); result <- r.Commit(t.Context(), snapshot) }()
	}
	wg.Wait()
	close(result)
	success, lost := 0, 0
	for err := range result {
		if err == nil {
			success++
		} else if errors.Is(err, search.ErrRefreshLost) {
			lost++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || lost != 1 || status(t, s, c).Generation != 1 || matches(t, db, "generation") != 1 {
		t.Fatal(success, lost, status(t, s, c))
	}
}
