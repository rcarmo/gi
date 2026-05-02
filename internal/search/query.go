package search

import "strings"

// QueryMode is a coarse routing hint for hybrid ranking.
type QueryMode string

const (
	QueryModeMixed    QueryMode = "mixed"
	QueryModeLexical  QueryMode = "lexical"
	QueryModeSemantic QueryMode = "semantic"
)

// ClassifyQuery applies simple heuristics to bias lexical vs semantic search.
func ClassifyQuery(q string) QueryMode {
	q = strings.TrimSpace(q)
	if q == "" {
		return QueryModeMixed
	}
	if strings.Contains(q, "::") || strings.Contains(q, "_") || strings.Contains(q, "/") || strings.Contains(q, ".go") || strings.Contains(q, "vfs://") {
		return QueryModeLexical
	}
	upper := 0
	for _, r := range q {
		if r >= 'A' && r <= 'Z' {
			upper++
		}
	}
	if upper >= 2 {
		return QueryModeLexical
	}
	if strings.Contains(q, " ") {
		return QueryModeSemantic
	}
	return QueryModeMixed
}
