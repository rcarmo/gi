package tools

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/search/chunking"
	"github.com/rcarmo/gi/internal/search/indexer"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func producerFixture(t *testing.T) (*store.Store, config.RuntimeConfig, []searchstore.ScopeConfig) {
	t.Helper()
	root := t.TempDir()
	for _, path := range []string{"notes", ".pi/skills", "docs"} {
		if err := os.MkdirAll(filepath.Join(root, path), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes/a.md"), []byte("oldorchid"), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := store.Open(filepath.Join(t.TempDir(), "producer.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := config.RuntimeConfig{WorkspaceRoot: root}
	cfg.WorkspaceIndex.ExtraRoots = []string{"docs"}
	var scopes []searchstore.ScopeConfig
	for _, name := range []string{"all", "notes", "skills"} {
		c, err := searchstore.ConfiguredScopeConfig(root, name, []string{"docs"}, nil, nil, chunking.LineVersion)
		if err != nil {
			t.Fatal(err)
		}
		if err := indexer.NewWorker(searchstore.NewRefreshStore(db.DB())).Run(t.Context(), c); err != nil {
			t.Fatal(err)
		}
		scopes = append(scopes, c)
	}
	return db, cfg, scopes
}
func producerStatus(t *testing.T, db *store.Store, c searchstore.ScopeConfig) searchstore.ScopeStatus {
	t.Helper()
	st, err := searchstore.NewRefreshStore(db.DB()).Status(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	return st
}
func TestNativeAndScriptWritesInvalidateWithoutRefreshing(t *testing.T) {
	for _, via := range []string{"native", "script"} {
		t.Run(via, func(t *testing.T) {
			db, cfg, scopes := producerFixture(t)
			before := producerStatus(t, db, scopes[0])
			if via == "native" {
				if _, err := ExecuteWrite(t.Context(), cfg, db, goai.ToolCall{Arguments: map[string]any{"path": "notes/a.md", "content": "newviolet"}}); err != nil {
					t.Fatal(err)
				}
			} else {
				out := NewScriptTool(db, cfg).Execute(t.Context(), ScriptInput{Engine: "js", Script: `gi.writeFile("notes/a.md", "newviolet"); "ok"`})
				if out.Error != "" {
					t.Fatal(out)
				}
			}
			for _, c := range scopes {
				st := producerStatus(t, db, c)
				if c.Scope() == "skills" {
					if st.State != "ready" || st.RequestedRevision != 0 {
						t.Fatal(st)
					}
					continue
				}
				if st.State != "stale" || st.RequestedRevision != 2 || st.AcknowledgedRevision != 0 || st.Generation != before.Generation {
					t.Fatal(st)
				}
			}
			st := producerStatus(t, db, scopes[0])
			if st.LastIndexedAtMS != before.LastIndexedAtMS || st.IndexedFileCount != before.IndexedFileCount {
				t.Fatal(st)
			}
			storage := searchstore.NewRefreshStore(db.DB())
			old, err := storage.Query(t.Context(), scopes[0], "oldorchid", 10, 0)
			if err != nil || len(old.Hits) != 1 {
				t.Fatal(old, err)
			}
			if err := indexer.NewWorker(storage).Run(t.Context(), scopes[0]); err != nil {
				t.Fatal(err)
			}
			hits, err := storage.Query(t.Context(), scopes[0], "newviolet", 10, 0)
			if err != nil || len(hits.Hits) != 1 {
				t.Fatal(hits, err)
			}
			if st := producerStatus(t, db, scopes[0]); st.State != "ready" || st.AcknowledgedRevision != 2 {
				t.Fatal(st)
			}
		})
	}
}
func TestWriteScopeFanoutAndVFSIsolation(t *testing.T) {
	db, cfg, scopes := producerFixture(t)
	for _, path := range []string{"other.txt", "vfs://scratch/note.md"} {
		if err := WriteFile(t.Context(), cfg, db, path, "content"); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range scopes {
		if st := producerStatus(t, db, c); st.RequestedRevision != 0 {
			t.Fatal(st)
		}
	}
	if err := WriteFile(t.Context(), cfg, db, "docs/extra.md", "extra"); err != nil {
		t.Fatal(err)
	}
	if st := producerStatus(t, db, scopes[0]); st.RequestedRevision != 2 {
		t.Fatal(st)
	}
	for _, c := range scopes[1:] {
		if st := producerStatus(t, db, c); st.RequestedRevision != 0 {
			t.Fatal(st)
		}
	}
	if err := WriteFile(t.Context(), cfg, db, ".pi/skills/new/SKILL.md", "skill"); err != nil {
		t.Fatal(err)
	}
	if st := producerStatus(t, db, scopes[2]); st.RequestedRevision != 2 {
		t.Fatal(st)
	}
}
func TestWritePreFailureRollsBackAllScopesAndPreventsMutation(t *testing.T) {
	db, cfg, scopes := producerFixture(t)
	if _, err := db.DB().Exec(`CREATE TRIGGER reject_notes BEFORE UPDATE ON workspace_index_invalidations WHEN new.scope='notes' BEGIN SELECT raise(abort,'notes rejected'); END`); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(t.Context(), cfg, db, "notes/a.md", "wrong"); err == nil || !strings.Contains(err.Error(), "not attempted") {
		t.Fatal(err)
	}
	for _, c := range scopes {
		if st := producerStatus(t, db, c); st.State != "ready" || st.RequestedRevision != 0 {
			t.Fatal("partial event", st)
		}
	}
	bytes, err := os.ReadFile(filepath.Join(cfg.WorkspaceRoot, "notes/a.md"))
	if err != nil || string(bytes) != "oldorchid" {
		t.Fatal(string(bytes), err)
	}
}
func TestWriteBracketsRefreshAndRetainsPendingOnReopen(t *testing.T) {
	db, cfg, scopes := producerFixture(t)
	storage := searchstore.NewRefreshStore(db.DB())
	notify := func(ctx context.Context) error {
		_, err := storage.InvalidateScopes(ctx, scopes, []string{"notes/a.md"})
		return err
	}
	var captured *searchstore.Refresh
	var snap searchstore.CompleteSnapshot
	if err := writeWithInvalidation(t.Context(), notify, func() error {
		var err error
		captured, err = storage.Begin(t.Context(), scopes[0], time.Minute)
		if err != nil {
			return err
		}
		snap, err = indexer.ScanScope(t.Context(), scopes[0])
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(cfg.WorkspaceRoot, "notes/a.md"), []byte("newviolet"), 0600)
	}); err != nil {
		t.Fatal(err)
	}
	if err := captured.Commit(t.Context(), snap); err != nil {
		t.Fatal(err)
	}
	if st := producerStatus(t, db, scopes[0]); st.State != "stale" || st.RequestedRevision != 2 || st.AcknowledgedRevision != 1 {
		t.Fatal(st)
	}
	// A second connection proves revisions are durable, not process-local hooks.
	var dbPath string
	var seq int
	var name string
	if err := db.DB().QueryRow("PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if st := producerStatus(t, reopened, scopes[0]); st.State != "stale" || st.RequestedRevision != 2 {
		t.Fatal(st)
	}
	old, err := storage.Query(t.Context(), scopes[0], "oldorchid", 10, 0)
	if err != nil || len(old.Hits) != 1 {
		t.Fatal(old, err)
	}
	if err := indexer.NewWorker(storage).Run(t.Context(), scopes[0]); err != nil {
		t.Fatal(err)
	}
	fresh, err := storage.Query(t.Context(), scopes[0], "newviolet", 10, 0)
	if err != nil || len(fresh.Hits) != 1 {
		t.Fatal(fresh, err)
	}
}
func TestWritePostFailureAndCancellationReportChangedBytes(t *testing.T) {
	for _, mode := range []string{"partial", "post", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			calls := 0
			wrote := false
			notify := func(ctx context.Context) error {
				calls++
				if ctx.Err() != nil {
					t.Fatal("post cleanup inherited cancellation")
				}
				if calls == 2 && mode == "post" {
					return errors.New("database unavailable")
				}
				return nil
			}
			err := writeWithInvalidation(ctx, notify, func() error {
				wrote = true
				cancel()
				if mode == "partial" {
					return errors.New("partial write")
				}
				return nil
			})
			if calls != 2 || !wrote {
				t.Fatal(calls, wrote)
			}
			if mode == "post" && (err == nil || !strings.Contains(err.Error(), "bytes may have changed")) {
				t.Fatal(err)
			}
			if mode == "partial" && (err == nil || !strings.Contains(err.Error(), "partial write")) {
				t.Fatal(err)
			}
			if mode == "cancel" && err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWriteRejectsAliasesEscapeInvalidConfigAndCancelledRequest(t *testing.T) {
	db, cfg, scopes := producerFixture(t)
	if err := os.Symlink("notes", filepath.Join(cfg.WorkspaceRoot, "alias")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"alias/a.md", "../escape", "notes"} {
		if err := WriteFile(t.Context(), cfg, db, path, "wrong"); err == nil {
			t.Fatal(path)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := WriteFile(ctx, cfg, db, "notes/a.md", "wrong"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cfg.WorkspaceIndex.ExtraRoots = []string{"../escape"}
	if err := WriteFile(t.Context(), cfg, db, "notes/a.md", "wrong"); err == nil {
		t.Fatal("bad config accepted")
	}
	for _, c := range scopes {
		if st := producerStatus(t, db, c); st.RequestedRevision != 0 {
			t.Fatal(st)
		}
	}
}

func TestWriteNativePostNotificationFailurePreservesBytesAndPreRevision(t *testing.T) {
	db, cfg, scopes := producerFixture(t)
	if _, err := db.DB().Exec(`CREATE TRIGGER reject_post BEFORE UPDATE ON workspace_index_invalidations WHEN new.scope='notes' AND new.requested_revision=2 BEGIN SELECT raise(abort,'post rejected'); END`); err != nil {
		t.Fatal(err)
	}
	err := WriteFile(t.Context(), cfg, db, "notes/a.md", "changed bytes")
	if err == nil || !strings.Contains(err.Error(), "bytes may have changed") {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(filepath.Join(cfg.WorkspaceRoot, "notes/a.md"))
	if err != nil || string(bytes) != "changed bytes" {
		t.Fatal(string(bytes), err)
	}
	for _, c := range scopes[:2] {
		st := producerStatus(t, db, c)
		if st.RequestedRevision != 1 || st.AcknowledgedRevision != 0 || st.State != "stale" {
			t.Fatal("post partially committed", st)
		}
	}
}
