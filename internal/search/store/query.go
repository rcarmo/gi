package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type LexicalHit struct {
	DocumentID int64   `json:"document_id"`
	ChunkID    int64   `json:"chunk_id"`
	Path       string  `json:"path"`
	StartLine  int     `json:"start_line"`
	EndLine    int     `json:"end_line"`
	StartByte  int     `json:"start_byte"`
	EndByte    int     `json:"end_byte"`
	Snippet    string  `json:"snippet"`
	Rank       float64 `json:"rank"`
}
type LexicalResult struct {
	Hits []LexicalHit `json:"hits"`
	Mode string       `json:"mode"`
}

// Query never starts or awaits a refresh. Membership and configuration must
// match the caller's workspace/scope; stale content under the same configuration
// remains readable. Different configuration/chunker versions yield no old hits.
func (s *RefreshStore) Query(ctx context.Context, c ScopeConfig, q string, limit, offset int) (LexicalResult, error) {
	result := LexicalResult{Hits: []LexicalHit{}, Mode: "fts"}
	q = strings.TrimSpace(q)
	if c.fingerprint == "" {
		return result, fmt.Errorf("unresolved scope config")
	}
	if err := ValidateLexicalQuery(q, limit, offset); err != nil {
		return result, err
	}
	const join = ` JOIN workspace_index_chunks c ON c.id=f.rowid
 JOIN workspace_index_documents d ON d.id=c.document_id
 JOIN workspace_index_memberships m ON m.document_id=d.id AND m.workspace_id=d.workspace_id
 JOIN workspace_index_scopes sc ON sc.workspace_id=m.workspace_id AND sc.scope=m.scope
 JOIN workspace_index_workspaces w ON w.id=d.workspace_id `
	const scope = `w.root_identity=? AND m.scope=? AND sc.config_hash=? AND d.chunker_version=?`
	args := []any{c.workspace, c.scope, c.fingerprint, c.chunker, q, limit, offset}
	rows, err := s.db.QueryContext(ctx, `SELECT d.id,c.id,d.path,c.start_line,c.end_line,c.start_byte,c.end_byte,
 substr(snippet(workspace_index_fts,0,'','',' … ',32),1,512),bm25(workspace_index_fts)
 FROM workspace_index_fts f`+join+` WHERE `+scope+` AND workspace_index_fts MATCH ? ORDER BY bm25(workspace_index_fts),d.path,c.id LIMIT ? OFFSET ?`, args...)
	if err == nil {
		hits, readErr := readLexicalRows(rows)
		if readErr == nil {
			result.Hits = hits
			return result, nil
		}
		err = readErr
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	if !ftsSyntaxError(err) {
		return result, err
	}
	// Only malformed FTS expressions use literal AND substring matching. Do not
	// swallow database/cancellation errors. SQLite lower() is ASCII-only; Unicode
	// tokens use their literal spelling in this deliberately simple fallback.
	terms := strings.Fields(q)
	if len(terms) > 16 {
		return result, fmt.Errorf("literal fallback exceeds 16 terms")
	}
	clauses := make([]string, 0, len(terms))
	args = []any{c.workspace, c.scope, c.fingerprint, c.chunker}
	for _, term := range terms {
		clauses = append(clauses, "instr(lower(c.content),lower(?))>0")
		args = append(args, term)
	}
	args = append(args, limit, offset)
	rows, err = s.db.QueryContext(ctx, `SELECT d.id,c.id,d.path,c.start_line,c.end_line,c.start_byte,c.end_byte,substr(c.content,1,512),0.0
 FROM workspace_index_chunks c JOIN workspace_index_documents d ON d.id=c.document_id
 JOIN workspace_index_memberships m ON m.document_id=d.id AND m.workspace_id=d.workspace_id
 JOIN workspace_index_scopes sc ON sc.workspace_id=m.workspace_id AND sc.scope=m.scope
 JOIN workspace_index_workspaces w ON w.id=d.workspace_id WHERE `+scope+` AND `+strings.Join(clauses, " AND ")+` ORDER BY d.path,c.id LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return result, err
	}
	result.Mode = "literal"
	result.Hits, err = readLexicalRows(rows)
	return result, err
}
func ValidateLexicalQuery(q string, limit, offset int) error {
	if strings.TrimSpace(q) == "" || len(q) > 1024 || !utf8.ValidString(q) || strings.ContainsRune(q, 0) || limit < 1 || limit > 50 || offset < 0 || offset > 10000 {
		return fmt.Errorf("invalid lexical query or bounds")
	}
	return nil
}
func ftsSyntaxError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "fts5: syntax error") || strings.Contains(msg, "unterminated string") || strings.Contains(msg, "fts5: column queries are not supported")
}
func readLexicalRows(rows *sql.Rows) ([]LexicalHit, error) {
	defer rows.Close()
	hits := []LexicalHit{}
	for rows.Next() {
		var hit LexicalHit
		if err := rows.Scan(&hit.DocumentID, &hit.ChunkID, &hit.Path, &hit.StartLine, &hit.EndLine, &hit.StartByte, &hit.EndByte, &hit.Snippet, &hit.Rank); err != nil {
			return nil, err
		}
		hits = append(hits, hit)
	}
	return hits, rows.Err()
}
