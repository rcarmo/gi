package indexer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	core "github.com/rcarmo/gi/internal/store"
)

func scanFixture(t *testing.T) (string, func(string, string)) {
	t.Helper()
	root := t.TempDir()
	write := func(p, text string) {
		t.Helper()
		p = filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, write
}
func scanConfig(t *testing.T, root, scope string, roots []string) searchstore.ScopeConfig {
	t.Helper()
	c, err := searchstore.NewScopeConfig(root, scope, roots, []string{"md", "txt", "go"}, chunking.LineVersion)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestScanScopeCompleteSkillsRootsFiltersAndCommittedIdentity(t *testing.T) {
	root, write := scanFixture(t)
	write("notes/alpha.md", "# heading\n"+strings.Repeat("界 words\n", 250))
	write("notes/.hidden.txt", "hidden included")
	write(".pi/skills/example/SKILL.md", "skill available")
	for _, p := range []string{"notes/.git/skip.md", "notes/node_modules/skip.md", "notes/.cache/skip.md", "notes/generated/skip.md", "notes/no.exe"} {
		write(p, "skipped")
	}
	write("notes/binary.txt", "nul\x00text")
	write("notes/invalid.txt", string([]byte{255}))
	write("notes/large.txt", strings.Repeat("x", searchstore.MaxDocumentBytes+1))
	write("notes/empty.txt", "")
	if err := os.Symlink("alpha.md", filepath.Join(root, "notes/link.md")); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "notes/outside")); err != nil {
		t.Fatal(err)
	}
	c := scanConfig(t, root, "all", []string{"notes/sub", ".pi/skills", "notes"})
	snapshot, err := ScanScope(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, d := range snapshot.Documents {
		paths = append(paths, d.Path)
	}
	want := []string{".pi/skills/example/SKILL.md", "notes/.hidden.txt", "notes/alpha.md", "notes/empty.txt"}
	if !snapshot.Complete || !reflect.DeepEqual(paths, want) {
		t.Fatal(paths)
	}
	again, err := ScanScope(t.Context(), c)
	if err != nil || !reflect.DeepEqual(snapshot, again) {
		t.Fatal("unstable scan", err)
	}
	db, err := core.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := searchstore.NewRefreshStore(db.DB())
	commit := func(snap searchstore.CompleteSnapshot) {
		t.Helper()
		r, err := store.Begin(t.Context(), c, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if err = r.Commit(t.Context(), snap); err != nil {
			t.Fatal(err)
		}
	}
	commit(snapshot)
	ids := func() []int64 {
		t.Helper()
		rows, err := db.DB().Query("SELECT id FROM workspace_index_chunks ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var ids []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		return ids
	}
	before := ids()
	commit(again)
	if !reflect.DeepEqual(before, ids()) {
		t.Fatal("unchanged chunk IDs moved")
	}
	// A same-size edit with restored mtime is read and hashed, not skipped.
	p := filepath.Join(root, ".pi/skills/example/SKILL.md")
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	write(".pi/skills/example/SKILL.md", "skill different")
	if err := os.Chtimes(p, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	changed, err := ScanScope(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	commit(changed)
	var n int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH 'different'").Scan(&n); err != nil || n != 1 {
		t.Fatal(n, err)
	}
}

func TestScanFailuresReturnNoCompleteSnapshot(t *testing.T) {
	root, write := scanFixture(t)
	write("notes/a.md", "original")
	c := scanConfig(t, root, "notes", []string{"notes"})
	cases := []struct {
		name   string
		limits scanLimits
		mutate func()
		want   error
	}{
		{"entries", scanLimits{entries: 0, depth: 64, totalText: 1000, files: 10, chunks: 10}, nil, ErrScanLimit},
		{"text", scanLimits{entries: 10, depth: 64, totalText: 1, files: 10, chunks: 10}, nil, ErrScanLimit},
		{"files", scanLimits{entries: 10, depth: 64, totalText: 1000, files: 0, chunks: 10}, nil, ErrScanLimit},
		{"chunks", scanLimits{entries: 10, depth: 64, totalText: 1000, files: 10, chunks: 0}, nil, ErrScanLimit},
		{"metadata_collision", defaultScanLimits(), func() {
			p := filepath.Join(root, "notes/a.md")
			info, _ := os.Stat(p)
			write("notes/a.md", "modified")
			os.Chtimes(p, info.ModTime(), info.ModTime())
		}, ErrScanChanged},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snap, err := scanScope(t.Context(), c, tc.limits, tc.mutate)
			if !errors.Is(err, tc.want) || snap.Complete || len(snap.Documents) != 0 {
				t.Fatal(snap.Complete, err)
			}
		})
	}
	for _, p := range []string{"missing", "notes/a.md"} {
		bad := scanConfig(t, root, "bad", []string{p})
		if snap, err := ScanScope(t.Context(), bad); err == nil || snap.Complete {
			t.Fatal(p, err)
		}
	}
	if err := os.Symlink("notes", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}
	if snap, err := ScanScope(t.Context(), scanConfig(t, root, "bad", []string{"alias"})); err == nil || snap.Complete {
		t.Fatal("symlink scope accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if snap, err := ScanScope(ctx, c); !errors.Is(err, context.Canceled) || snap.Complete {
		t.Fatal(err)
	}
	write("notes/deep/file.md", "deep")
	limits := defaultScanLimits()
	limits.depth = 0
	if snap, err := scanScope(t.Context(), c, limits, nil); !errors.Is(err, ErrScanLimit) || snap.Complete {
		t.Fatal(err)
	}
	wrong, err := searchstore.NewScopeConfig(root, "bad", []string{"notes"}, []string{"md"}, "other-version")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ScanScope(t.Context(), wrong); err == nil {
		t.Fatal("mismatched chunker")
	}
}

func TestScanMutationAndCancellationPreserveCommittedIndex(t *testing.T) {
	for _, mode := range []string{"replace", "delete", "new-file", "cancel", "unreadable"} {
		t.Run(mode, func(t *testing.T) {
			root, write := scanFixture(t)
			write("notes/a.md", "retained orchid")
			c := scanConfig(t, root, "notes", []string{"notes"})
			db, err := core.Open(filepath.Join(t.TempDir(), "index.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			store := searchstore.NewRefreshStore(db.DB())
			snap, err := ScanScope(t.Context(), c)
			if err != nil {
				t.Fatal(err)
			}
			r, err := store.Begin(t.Context(), c, time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if err := r.Commit(t.Context(), snap); err != nil {
				t.Fatal(err)
			}
			r, err = store.Begin(t.Context(), c, time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			snap, scanErr := scanScope(ctx, c, defaultScanLimits(), func() {
				p := filepath.Join(root, "notes/a.md")
				switch mode {
				case "replace":
					os.Rename(p, p+".old")
					write("notes/a.md", "replaced orchid")
				case "delete":
					os.Remove(p)
				case "new-file":
					write("notes/b.md", "missed if unchecked")
				case "cancel":
					cancel()
				case "unreadable":
					os.Chmod(p, 0)
					t.Cleanup(func() { os.Chmod(p, 0600) })
				}
			})
			if scanErr == nil || snap.Complete {
				t.Fatal("mutation not rejected", mode)
			}
			if err := r.Commit(t.Context(), snap); !errors.Is(err, searchstore.ErrIncompleteSnapshot) {
				t.Fatal(err)
			}
			if err := r.Fail(t.Context(), scanErr); err != nil {
				t.Fatal(err)
			}
			var n int
			if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH 'orchid'").Scan(&n); err != nil || n != 1 {
				t.Fatal(n, err)
			}
			st, err := store.Status(t.Context(), c)
			if err != nil || st.State != "failed" || st.Generation != 1 || st.IndexedFileCount != 1 {
				t.Fatal(st, err)
			}
		})
	}
}

func TestScanDefaultSkillsIsNotBlanketHiddenExcluded(t *testing.T) {
	root, write := scanFixture(t)
	write("notes/note.md", "note")
	write(".pi/skills/skill/SKILL.md", "skill")
	for scope, count := range map[string]int{"notes": 1, "skills": 1, "all": 2} {
		c, err := searchstore.DefaultScopeConfig(root, scope, nil, nil, chunking.LineVersion)
		if err != nil {
			t.Fatal(err)
		}
		snap, err := ScanScope(t.Context(), c)
		if err != nil || len(snap.Documents) != count {
			t.Fatal(scope, err)
		}
	}
	c := scanConfig(t, root, "custom", []string{"."})
	snap, err := ScanScope(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, d := range snap.Documents {
		paths = append(paths, d.Path)
	}
	sort.Strings(paths)
	if len(paths) != 2 {
		t.Fatal(paths)
	}
}

func TestScanWorkspaceAndRootAncestorReplacement(t *testing.T) {
	for _, mode := range []string{"workspace", "ancestor"} {
		t.Run(mode, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "workspace")
			if err := os.MkdirAll(filepath.Join(root, "notes/nested"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "notes/nested/a.md"), []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			c := scanConfig(t, root, "nested", []string{"notes/nested"})
			snap, err := scanScope(t.Context(), c, defaultScanLimits(), func() {
				target := root
				if mode == "ancestor" {
					target = filepath.Join(root, "notes")
				}
				if err := os.Rename(target, target+".old"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
			})
			if !errors.Is(err, ErrScanChanged) || snap.Complete || len(snap.Documents) != 0 {
				t.Fatal(mode, err)
			}
		})
	}
}

func TestScanSuccessfulEmptyInventoryCleansOnlyItsScope(t *testing.T) {
	root, write := scanFixture(t)
	write("notes/a.md", "shared content")
	db, err := core.Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := searchstore.NewRefreshStore(db.DB())
	first := scanConfig(t, root, "first", []string{"notes"})
	second := scanConfig(t, root, "second", []string{"notes"})
	commit := func(c searchstore.ScopeConfig) {
		t.Helper()
		snap, err := ScanScope(t.Context(), c)
		if err != nil {
			t.Fatal(err)
		}
		r, err := store.Begin(t.Context(), c, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Commit(t.Context(), snap); err != nil {
			t.Fatal(err)
		}
	}
	commit(first)
	commit(second)
	if err := os.Remove(filepath.Join(root, "notes/a.md")); err != nil {
		t.Fatal(err)
	}
	commit(first)
	var count int
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH 'shared'").Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
	commit(second)
	if err := db.DB().QueryRow("SELECT count(*) FROM workspace_index_fts WHERE workspace_index_fts MATCH 'shared'").Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
	if _, err := db.DB().Exec("INSERT INTO workspace_index_fts(workspace_index_fts,rank) VALUES('integrity-check',1)"); err != nil {
		t.Fatal(err)
	}
}
