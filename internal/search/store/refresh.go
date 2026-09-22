package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/search/chunking"
)

var ErrRefreshBusy = errors.New("workspace index refresh already owned")
var ErrRefreshLost = errors.New("workspace index refresh ownership expired or replaced")
var ErrIncompleteSnapshot = errors.New("workspace index requires a complete snapshot")

const indexNowMS = "CAST((julianday('now') - 2440587.5)*86400000 AS INTEGER)"

type RefreshStore struct{ db *sql.DB }

func NewRefreshStore(db *sql.DB) *RefreshStore { return &RefreshStore{db: db} }

type Refresh struct {
	store       *RefreshStore
	workspaceID int64
	token       string
	config      ScopeConfig
}

type ScopeStatus struct {
	State            string
	ConfigHash       string
	Roots            []string
	Generation       int64
	IndexedFileCount int
	LastIndexedAtMS  sql.NullInt64
	UpdatedAtMS      int64
	LastError        string
}

func (s *RefreshStore) Status(ctx context.Context, c ScopeConfig) (ScopeStatus, error) {
	if c.fingerprint == "" {
		return ScopeStatus{}, fmt.Errorf("unresolved scope config")
	}
	status := ScopeStatus{State: "never_indexed", Roots: c.Roots(), ConfigHash: c.fingerprint}
	var roots string
	var leaseActive bool
	err := s.db.QueryRowContext(ctx, `SELECT sc.state,sc.config_hash,sc.roots_json,sc.committed_generation,sc.indexed_file_count,sc.last_indexed_at_ms,sc.updated_at_ms,sc.last_error,
 EXISTS(SELECT 1 FROM workspace_index_leases l WHERE l.workspace_id=sc.workspace_id AND l.expires_at_ms>`+indexNowMS+`)
 FROM workspace_index_scopes sc JOIN workspace_index_workspaces w ON w.id=sc.workspace_id WHERE w.root_identity=? AND sc.scope=?`, c.workspace, c.scope).Scan(&status.State, &status.ConfigHash, &roots, &status.Generation, &status.IndexedFileCount, &status.LastIndexedAtMS, &status.UpdatedAtMS, &status.LastError, &leaseActive)
	if errors.Is(err, sql.ErrNoRows) {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	if err = json.Unmarshal([]byte(roots), &status.Roots); err != nil {
		return status, err
	}
	if status.State == "indexing" && !leaseActive {
		status.State = "stale"
		status.LastError = "Index refresh interrupted or expired"
	}
	if status.ConfigHash != c.fingerprint && status.State != "indexing" {
		status.State = "stale"
	}
	return status, nil
}

func (s *RefreshStore) Begin(ctx context.Context, c ScopeConfig, ttl time.Duration) (*Refresh, error) {
	if c.fingerprint == "" {
		return nil, fmt.Errorf("unresolved scope config")
	}
	if ttl < time.Second || ttl > 5*time.Minute {
		return nil, fmt.Errorf("index lease TTL must be 1s–5m")
	}
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}
	r := &Refresh{store: s, token: hex.EncodeToString(token), config: c}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO workspace_index_workspaces(root_identity) VALUES(?) ON CONFLICT(root_identity) DO NOTHING", c.workspace); err != nil {
		return nil, err
	}
	if err = tx.QueryRowContext(ctx, "SELECT id FROM workspace_index_workspaces WHERE root_identity=?", c.workspace).Scan(&r.workspaceID); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO workspace_index_leases(workspace_id,owner_token,expires_at_ms) VALUES(?,?,`+indexNowMS+`+?)
 ON CONFLICT(workspace_id) DO UPDATE SET owner_token=excluded.owner_token,expires_at_ms=excluded.expires_at_ms WHERE workspace_index_leases.expires_at_ms<=`+indexNowMS, r.workspaceID, r.token, ttl.Milliseconds())
	if err != nil {
		return nil, err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return nil, ErrRefreshBusy
	}
	// Taking the workspace lease proves no prior worker retains valid ownership;
	// an expired worker may still be running but cannot commit through its fence.
	if _, err = tx.ExecContext(ctx, `UPDATE workspace_index_scopes SET state='stale',last_error='Index refresh interrupted or expired',updated_at_ms=`+indexNowMS+` WHERE workspace_id=? AND state='indexing'`, r.workspaceID); err != nil {
		return nil, err
	}
	roots, _ := json.Marshal(c.roots)
	if _, err = tx.ExecContext(ctx, `INSERT INTO workspace_index_scopes(workspace_id,scope,config_hash,roots_json,state,updated_at_ms) VALUES(?,?,?,?,'indexing',`+indexNowMS+`)
 ON CONFLICT(workspace_id,scope) DO UPDATE SET state='indexing',last_error='',updated_at_ms=excluded.updated_at_ms`, r.workspaceID, c.scope, c.fingerprint, string(roots)); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Refresh) guard(ctx context.Context, tx *sql.Tx) error {
	result, err := tx.ExecContext(ctx, `UPDATE workspace_index_leases SET owner_token=owner_token WHERE workspace_id=? AND owner_token=? AND expires_at_ms>`+indexNowMS, r.workspaceID, r.token)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return ErrRefreshLost
	}
	return nil
}
func (r *Refresh) Renew(ctx context.Context, ttl time.Duration) error {
	if ttl < time.Second || ttl > 5*time.Minute {
		return fmt.Errorf("index lease TTL must be 1s–5m")
	}
	result, err := r.store.db.ExecContext(ctx, `UPDATE workspace_index_leases SET expires_at_ms=`+indexNowMS+`+? WHERE workspace_id=? AND owner_token=? AND expires_at_ms>`+indexNowMS, ttl.Milliseconds(), r.workspaceID, r.token)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return ErrRefreshLost
	}
	return nil
}

// Fail requires live ownership, cannot overwrite a successor, and retains all
// committed content/count/configuration. Call with a cleanup context on cancel.
func (r *Refresh) Fail(ctx context.Context, cause error) error {
	if cause == nil {
		return fmt.Errorf("failure cause required")
	}
	tx, err := r.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = r.guard(ctx, tx); err != nil {
		return err
	}
	detail := strings.ToValidUTF8(cause.Error(), "�")
	if len(detail) > 2048 {
		detail = detail[:2048]
		for !utf8.ValidString(detail) {
			detail = detail[:len(detail)-1]
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE workspace_index_scopes SET state='failed',last_error=?,updated_at_ms=`+indexNowMS+` WHERE workspace_id=? AND scope=?`, detail, r.workspaceID, r.config.scope); err != nil {
		return err
	}
	if err = r.guard(ctx, tx); err != nil {
		return err
	}
	if err = r.release(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *Refresh) release(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM workspace_index_leases WHERE workspace_id=? AND owner_token=?", r.workspaceID, r.token)
	return err
}

type RefreshDocument struct {
	Path     string
	Content  string
	MtimeNS  int64
	Language string
	Chunks   []chunking.Chunk
}
type CompleteSnapshot struct {
	// Set only after every configured root was inventoried successfully. A missing
	// optional root must be an explicit scanner policy, not an ignored IO error.
	Complete  bool
	Documents []RefreshDocument
}

func (r *Refresh) validate(snapshot CompleteSnapshot) error {
	if !snapshot.Complete {
		return ErrIncompleteSnapshot
	}
	if len(snapshot.Documents) > MaxSnapshotFiles {
		return fmt.Errorf("index file limit exceeded")
	}
	seen := map[string]bool{}
	total, chunks, indexedBytes := 0, 0, 0
	for _, doc := range snapshot.Documents {
		if !r.config.eligible(doc.Path) || seen[doc.Path] {
			return fmt.Errorf("invalid, duplicate or out-of-scope path %q", doc.Path)
		}
		seen[doc.Path] = true
		total += len(doc.Content)
		chunks += len(doc.Chunks)
		if len(doc.Content) > MaxDocumentBytes || total > MaxSnapshotText || chunks > MaxSnapshotChunks {
			return fmt.Errorf("index text/chunk limit exceeded")
		}
		if len(doc.Language) > 128 || !utf8.ValidString(doc.Language) || strings.ContainsRune(doc.Language, 0) {
			return fmt.Errorf("invalid document language")
		}
		if !utf8.ValidString(doc.Content) || strings.ContainsRune(doc.Content, 0) {
			return fmt.Errorf("index content must be UTF-8 text")
		}
		if len(doc.Chunks) == 0 {
			return fmt.Errorf("document requires at least one chunk")
		}
		var newlines []int
		for pos, ch := range doc.Content {
			if ch == '\n' {
				newlines = append(newlines, pos)
			}
		}
		lineAt := func(offset int) int { return 1 + sort.SearchInts(newlines, offset) }
		lastStart := -1
		for n, c := range doc.Chunks {
			if c.ChunkIndex != n || c.StartByte < 0 || c.StartByte < lastStart || c.EndByte < c.StartByte || c.EndByte > len(doc.Content) || c.StartLine < 1 || c.EndLine < c.StartLine {
				return fmt.Errorf("invalid chunk bounds")
			}
			indexedBytes += len(c.Content) + len(c.Heading)
			if indexedBytes > MaxSnapshotText*2 {
				return fmt.Errorf("index chunk text limit exceeded")
			}
			if (c.StartByte < len(doc.Content) && !utf8.RuneStart(doc.Content[c.StartByte])) || (c.EndByte < len(doc.Content) && !utf8.RuneStart(doc.Content[c.EndByte])) {
				return fmt.Errorf("chunk cuts a UTF-8 code point")
			}
			text := doc.Content[c.StartByte:c.EndByte]
			if !utf8.ValidString(text) || c.Content != text || !utf8.ValidString(c.Heading) || len(c.Heading) > 4096 {
				return fmt.Errorf("chunk differs from source")
			}
			if c.StartLine != lineAt(c.StartByte) || c.EndLine != lineAt(c.EndByte) {
				return fmt.Errorf("chunk source line mismatch")
			}
			lastStart = c.StartByte
		}
	}
	return nil
}

// Commit publishes only a complete snapshot for the captured scope. Scanning is
// outside this transaction; both entry and final fences check token and expiry.
func (r *Refresh) Commit(ctx context.Context, snapshot CompleteSnapshot) error {
	if err := r.validate(snapshot); err != nil {
		return err
	}
	tx, err := r.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = r.guard(ctx, tx); err != nil {
		return err
	}
	var generation int64
	if err = tx.QueryRowContext(ctx, "SELECT committed_generation FROM workspace_index_scopes WHERE workspace_id=? AND scope=?", r.workspaceID, r.config.scope).Scan(&generation); err != nil {
		return err
	}
	generation++
	for _, doc := range snapshot.Documents {
		if err = ctx.Err(); err != nil {
			return err
		}
		if err = r.putDocument(ctx, tx, doc, generation); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM workspace_index_memberships WHERE workspace_id=? AND scope=? AND seen_generation<>?", r.workspaceID, r.config.scope, generation); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM workspace_index_documents WHERE workspace_id=? AND NOT EXISTS(SELECT 1 FROM workspace_index_memberships m WHERE m.document_id=workspace_index_documents.id)`, r.workspaceID); err != nil {
		return err
	}
	roots, _ := json.Marshal(r.config.roots)
	if _, err = tx.ExecContext(ctx, `UPDATE workspace_index_scopes SET state='ready',config_hash=?,roots_json=?,committed_generation=?,indexed_file_count=?,last_indexed_at_ms=`+indexNowMS+`,updated_at_ms=`+indexNowMS+`,last_error='' WHERE workspace_id=? AND scope=?`, r.config.fingerprint, string(roots), generation, len(snapshot.Documents), r.workspaceID, r.config.scope); err != nil {
		return err
	}
	if err = r.guard(ctx, tx); err != nil {
		return err
	}
	if err = r.release(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Refresh) putDocument(ctx context.Context, tx *sql.Tx, doc RefreshDocument, generation int64) error {
	sum := sha256.Sum256([]byte(doc.Content))
	hash := hex.EncodeToString(sum[:])
	var id int64
	var previousHash, version, language string
	err := tx.QueryRowContext(ctx, "SELECT id,content_hash,chunker_version,language FROM workspace_index_documents WHERE workspace_id=? AND path=?", r.workspaceID, doc.Path).Scan(&id, &previousHash, &version, &language)
	fresh := errors.Is(err, sql.ErrNoRows)
	if err != nil && !fresh {
		return err
	}
	changed := fresh || hash != previousHash || version != r.config.chunker || language != doc.Language
	if fresh {
		result, err := tx.ExecContext(ctx, `INSERT INTO workspace_index_documents(workspace_id,path,kind,language,size_bytes,mtime_ns,content_hash,chunker_version,indexed_at_ms) VALUES(?,?,'text',?,?,?,?,?,`+indexNowMS+`)`, r.workspaceID, doc.Path, doc.Language, len(doc.Content), doc.MtimeNS, hash, r.config.chunker)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
	} else if changed {
		// Overlapping scopes see shared document edits but must revalidate their own
		// inventories/configuration before advertising ready again.
		if _, err = tx.ExecContext(ctx, `UPDATE workspace_index_scopes SET state='stale',updated_at_ms=`+indexNowMS+` WHERE workspace_id=? AND scope<>? AND scope IN (SELECT scope FROM workspace_index_memberships WHERE document_id=?)`, r.workspaceID, r.config.scope, id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "DELETE FROM workspace_index_chunks WHERE document_id=?", id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE workspace_index_documents SET language=?,size_bytes=?,mtime_ns=?,content_hash=?,chunker_version=?,indexed_at_ms=`+indexNowMS+` WHERE id=?`, doc.Language, len(doc.Content), doc.MtimeNS, hash, r.config.chunker, id); err != nil {
			return err
		}
	} else {
		if _, err = tx.ExecContext(ctx, "UPDATE workspace_index_documents SET mtime_ns=? WHERE id=?", doc.MtimeNS, id); err != nil {
			return err
		}
	}
	if changed {
		for _, c := range doc.Chunks {
			if _, err = tx.ExecContext(ctx, `INSERT INTO workspace_index_chunks(document_id,chunk_index,start_byte,end_byte,start_line,end_line,heading,content) VALUES(?,?,?,?,?,?,?,?)`, id, c.ChunkIndex, c.StartByte, c.EndByte, c.StartLine, c.EndLine, c.Heading, c.Content); err != nil {
				return err
			}
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workspace_index_memberships(workspace_id,scope,document_id,seen_generation) VALUES(?,?,?,?) ON CONFLICT(workspace_id,scope,document_id) DO UPDATE SET seen_generation=excluded.seen_generation`, r.workspaceID, r.config.scope, id, generation)
	return err
}
