package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/search/indexer"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

func newIndexTestServer(t *testing.T, db *store.Store, cfg config.RuntimeConfig) *Server {
	t.Helper()
	engine := turn.New(db)
	t.Cleanup(func() { engine.Close() })
	srv := New(db, engine, cfg)
	// Invalid configurations retain native 400 validation without disabling chat.
	err := srv.StartWorkspaceIndex(t.Context())
	if _, configErr := srv.configuredWorkspaceScope("all"); configErr == nil && err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.CloseWorkspaceIndex)
	return srv
}
func indexLifecycleFixture(t *testing.T) (*Server, *store.Store, searchstore.ScopeConfig) {
	t.Helper()
	root := t.TempDir()
	for _, path := range []string{"notes", ".pi/skills"} {
		if err := os.MkdirAll(filepath.Join(root, path), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes/a.md"), []byte("native orchid"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "lifecycle.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	srv := newIndexTestServer(t, db, config.RuntimeConfig{WorkspaceRoot: root})
	c, err := srv.configuredWorkspaceScope("all")
	if err != nil {
		t.Fatal(err)
	}
	return srv, db, c
}
func apiCall(s *Server, ctx context.Context, method, path string) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, httptest.NewRequest(method, path, nil).WithContext(ctx))
	return res
}

func TestIndexLifecycleStartupReadsAndCloseNeverRefresh(t *testing.T) {
	srv, db, c := indexLifecycleFixture(t)
	scheduler := srv.indexScheduler
	if err := srv.StartWorkspaceIndex(t.Context()); err != nil || srv.indexScheduler != scheduler {
		t.Fatal("start not idempotent", err)
	}
	for _, path := range []string{"/api/workspace/index", "/api/workspace/search?q=orchid"} {
		if res := apiCall(srv, t.Context(), "GET", path); res.Code != 200 {
			t.Fatal(res.Code)
		}
	}
	var count int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_workspaces").Scan(&count); err != nil || count != 0 {
		t.Fatal("startup/read created work", count, err)
	}
	srv.CloseWorkspaceIndex()
	srv.CloseWorkspaceIndex()
	if err := srv.StartWorkspaceIndex(t.Context()); !errors.Is(err, indexer.ErrSchedulerClosed) {
		t.Fatal("resurrected", err)
	}
	if res := apiCall(srv, t.Context(), "POST", "/api/workspace/index"); res.Code != 503 {
		t.Fatal(res.Code, res.Body.String())
	}
	st, err := searchstore.NewRefreshStore(db.DB()).Status(t.Context(), c)
	if err != nil || st.State != "never_indexed" {
		t.Fatal(st, err)
	}
}

func TestIndexHTTPDisconnectDoesNotCancelSharedRefresh(t *testing.T) {
	srv, db, c := indexLifecycleFixture(t)
	storage := searchstore.NewRefreshStore(db.DB())
	peer, err := storage.Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	// Seed a real shared pending batch behind the peer lease. A cancelled HTTP
	// caller cannot cancel it, and another HTTP caller waits for its publication.
	ticket, err := srv.requestWorkspaceRefresh("all")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	disconnected := make(chan *httptest.ResponseRecorder, 1)
	go func() { disconnected <- apiCall(srv, ctx, "POST", "/api/workspace/index") }()
	waiting := make(chan *httptest.ResponseRecorder, 1)
	go func() { waiting <- apiCall(srv, t.Context(), "POST", "/api/workspace/index") }()
	select {
	case res := <-waiting:
		t.Fatal("refresh returned before peer release", res.Code)
	case <-time.After(75 * time.Millisecond):
	}
	cancel()
	select {
	case <-disconnected:
	case <-time.After(time.Second):
		t.Fatal("cancelled wait retained HTTP handler")
	}
	if err := peer.Fail(t.Context(), os.ErrClosed); err != nil {
		t.Fatal(err)
	}
	select {
	case res := <-waiting:
		var st workspaceIndexStatus
		if err := json.Unmarshal(res.Body.Bytes(), &st); err != nil || res.Code != 200 || st.State != "ready" || st.Generation != 1 {
			t.Fatal(st, res.Code, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shared refresh stuck")
	}
	status, err := ticket.Wait(t.Context())
	if err != nil || status.Generation != 1 {
		t.Fatal(status, err)
	}
	hits, err := storage.Query(t.Context(), c, "orchid", 10, 0)
	if err != nil || len(hits.Hits) != 1 {
		t.Fatal(hits, err)
	}
}

func TestIndexApplicationCancelDrainsActiveQueuedAndRejectsPost(t *testing.T) {
	srv, db, c := indexLifecycleFixture(t)
	peer, err := searchstore.NewRefreshStore(db.DB()).Begin(t.Context(), c, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Fail(t.Context(), os.ErrClosed)
	active, err := srv.requestWorkspaceRefresh("all")
	if err != nil {
		t.Fatal(err)
	}
	queued, err := srv.requestWorkspaceRefresh("notes")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); srv.CloseWorkspaceIndex() }()
	}
	wg.Wait()
	for _, ticket := range []*indexer.RefreshTicket{active, queued} {
		if _, err := ticket.Wait(t.Context()); err == nil {
			t.Fatal("closed work reported success")
		}
	}
	if res := apiCall(srv, t.Context(), "POST", "/api/workspace/index"); res.Code != 503 {
		t.Fatal(res.Code)
	}
	// Close did not fail or release a different owner's lease.
	if err := peer.Renew(t.Context(), time.Minute); err != nil {
		t.Fatal(err)
	}
}

func TestIndexLifecycleRequiresExplicitStartAndRespectsParentCancel(t *testing.T) {
	srv, db, _ := indexLifecycleFixture(t)
	other := New(db, srv.turns, srv.cfg)
	defer other.CloseWorkspaceIndex()
	if res := apiCall(other, t.Context(), "POST", "/api/workspace/index"); res.Code != http.StatusServiceUnavailable {
		t.Fatal(res.Code)
	}
	if err := other.StartWorkspaceIndex(nil); err == nil {
		t.Fatal("nil owner accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	if err := other.StartWorkspaceIndex(ctx); err != nil {
		t.Fatal(err)
	}
	cancel()
	other.CloseWorkspaceIndex()
	if _, err := other.requestWorkspaceRefresh("all"); !errors.Is(err, indexer.ErrSchedulerClosed) {
		t.Fatal(err)
	}
}

func TestIndexScopesRemainBoundToStartupWorkspaceIdentity(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	link := filepath.Join(root, "current")
	for _, dir := range []string{first, second} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(first, link); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "identity.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	srv := newIndexTestServer(t, db, config.RuntimeConfig{WorkspaceRoot: link})
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(second, link); err != nil {
		t.Fatal(err)
	}
	c, err := srv.workspaceScope(httptest.NewRequest("GET", "/api/workspace/index", nil))
	if err != nil || c.Workspace() != first {
		t.Fatal("GET/POST config diverged from startup scheduler", c.Workspace(), err)
	}
}

func TestIndexAlreadyCancelledPostDoesNotEnqueue(t *testing.T) {
	srv, db, _ := indexLifecycleFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	res := apiCall(srv, ctx, "POST", "/api/workspace/index")
	if res.Code == 200 {
		t.Fatal("cancel ignored")
	}
	srv.CloseWorkspaceIndex()
	var n int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_workspaces").Scan(&n); err != nil || n != 0 {
		t.Fatal("cancelled POST enqueued work", n, err)
	}
}
