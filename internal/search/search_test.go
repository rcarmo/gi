package search

import "testing"

func TestClassifyQuery(t *testing.T) {
	cases := map[string]QueryMode{
		"SubmitPromptRouted":              QueryModeLexical,
		"vfs://reference/foo":             QueryModeLexical,
		"where do we handle tool retries": QueryModeSemantic,
		"tool retries":                    QueryModeSemantic,
		"route":                           QueryModeMixed,
	}
	for in, want := range cases {
		if got := ClassifyQuery(in); got != want {
			t.Fatalf("ClassifyQuery(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMergeAndRank(t *testing.T) {
	fts := []SearchHit{{ChunkID: 1, Path: "a.go", FTSScore: 0.8}, {ChunkID: 2, Path: "b.go", FTSScore: 0.2}}
	vec := []SearchHit{{ChunkID: 2, Path: "b.go", VecScore: 0.9}, {ChunkID: 3, Path: "c.go", VecScore: 0.4}}
	got := MergeAndRank(fts, vec, 10)
	if len(got) != 3 {
		t.Fatalf("expected 3 hits, got %d", len(got))
	}
	if got[0].ChunkID != 2 {
		t.Fatalf("expected chunk 2 to rank first, got %+v", got[0])
	}
}
