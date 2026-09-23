package store_test

import (
	"context"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"strings"
	"testing"
)

func TestScopedLexicalQueryLimitsFallbackAndIsolation(t *testing.T) {
	db, s, root, _ := dbFixture(t)
	notes, skills, all := config(t, root, "notes"), config(t, root, "skills"), config(t, root, "all")
	n := document("notes/a.md", "orchid orchid 世界 literal \"unfinished 100%_ specimen")
	n2 := document("notes/b.md", "orchid violet")
	sk := document(".pi/skills/a/SKILL.md", "orchid skill")
	commit(t, begin(t, s, notes), n, n2)
	commit(t, begin(t, s, skills), sk)
	commit(t, begin(t, s, all), n, n2, sk)
	foreign := config(t, t.TempDir(), "all")
	commit(t, begin(t, s, foreign), document("notes/foreign.md", "orchid otherworkspace"))
	result, err := s.Query(t.Context(), notes, "orchid", 50, 0)
	if err != nil || len(result.Hits) != 2 || result.Mode != "fts" {
		t.Fatal(result, err)
	}
	for _, hit := range result.Hits {
		if !strings.HasPrefix(hit.Path, "notes/") || hit.DocumentID <= 0 || hit.ChunkID <= 0 || hit.StartLine != 1 || len([]rune(hit.Snippet)) > 512 {
			t.Fatal(hit)
		}
	}
	one, err := s.Query(t.Context(), notes, "orchid", 1, 0)
	if err != nil || len(one.Hits) != 1 {
		t.Fatal(one, err)
	}
	two, err := s.Query(t.Context(), notes, "orchid", 1, 1)
	if err != nil || len(two.Hits) != 1 || one.Hits[0].ChunkID == two.Hits[0].ChunkID {
		t.Fatal(two, err)
	}
	unicode, err := s.Query(t.Context(), all, "世界", 10, 0)
	if err != nil || len(unicode.Hits) != 1 {
		t.Fatal(unicode, err)
	}
	fallback, err := s.Query(t.Context(), notes, `"unfinished`, 10, 0)
	if err != nil || fallback.Mode != "literal" || len(fallback.Hits) != 1 {
		t.Fatal(fallback, err)
	}
	if foreignHits, err := s.Query(t.Context(), notes, "otherworkspace", 10, 0); err != nil || len(foreignHits.Hits) != 0 {
		t.Fatal(foreignHits, err)
	}
	// Config mismatch must not use memberships from an older/broader scope.
	changed, err := searchstore.NewScopeConfig(root, "notes", []string{"notes/sub"}, []string{"md"}, "lines-v1")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.Query(t.Context(), changed, "orchid", 10, 0); err != nil || len(got.Hits) != 0 {
		t.Fatal("config leak", got, err)
	}
	cancel, stop := context.WithCancel(t.Context())
	stop()
	if _, err := s.Query(cancel, notes, "orchid", 10, 0); err == nil {
		t.Fatal("cancel swallowed")
	}
	for _, q := range []string{"", " ", "x\x00y", strings.Repeat("x", 1025)} {
		if _, err := s.Query(t.Context(), notes, q, 10, 0); err == nil {
			t.Fatal("invalid query", q)
		}
	}
	if _, err := s.Query(t.Context(), notes, "orchid", 51, 0); err == nil {
		t.Fatal("limit")
	}
	if _, err := s.Query(t.Context(), notes, "orchid", 10, -1); err == nil {
		t.Fatal("offset")
	}
	// Missing FTS storage must remain an error, never disguised as literal search.
	if _, err := db.DB().Exec("DROP TABLE workspace_index_fts"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Query(t.Context(), notes, "orchid", 10, 0); err == nil {
		t.Fatal("schema failure hidden")
	}
}
