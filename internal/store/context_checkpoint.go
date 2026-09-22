package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrContextChanged = errors.New("context history changed during compaction")

type ContextFingerprint struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}
type ContextSnapshot struct {
	// Version fences concurrent boundary changes even when coverage fell back.
	Version  int64
	Summary  string
	Covered  []ContextFingerprint
	Messages []Message
	Expected []ContextFingerprint
}
type ContextBoundary struct {
	Version      int64
	SummaryCount int
	Expected     []ContextFingerprint
	Covered      []ContextFingerprint
}
type contextQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// ContextToken includes checkpoint version and all provider-visible history.
// It is a conflict token, not a credential or an exactly-once delivery receipt.
func ContextToken(snapshot ContextSnapshot) string {
	raw, _ := json.Marshal([]any{snapshot.Version, snapshot.Expected})
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}

func fingerprintMessage(m Message) (ContextFingerprint, error) {
	// Provider projection depends on role, content and media payload. Include all
	// payload metadata conservatively, so edits cannot leave a stale summary live.
	raw, err := json.Marshal([]any{m.Role, m.Content, m.Payload, m.CreatedAt})
	if err != nil {
		return ContextFingerprint{}, err
	}
	hash := sha256.Sum256(raw)
	return ContextFingerprint{ID: m.ID, Hash: hex.EncodeToString(hash[:])}, nil
}
func readContextSnapshot(ctx context.Context, q contextQuerier, sessionID string) (ContextSnapshot, error) {
	var out ContextSnapshot
	var coveredJSON string
	err := q.QueryRowContext(ctx, `select version,summary,covered_json from context_checkpoints where session_id=?`, sessionID).Scan(&out.Version, &out.Summary, &coveredJSON)
	if err != nil && err != sql.ErrNoRows {
		return out, err
	}
	if err == nil {
		if err = json.Unmarshal([]byte(coveredJSON), &out.Covered); err != nil {
			return out, fmt.Errorf("decode context checkpoint: %w", err)
		}
	}
	rows, err := q.QueryContext(ctx, `select id,session_id,role,content,payload_json,created_at from messages where session_id=? and role in ('user','assistant') and coalesce(json_extract(payload_json,'$.kind'),'')!='compaction' order by created_at,id`, sessionID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	var all []Message
	for rows.Next() {
		var m Message
		var payload string
		if err = rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &payload, &m.CreatedAt); err != nil {
			return out, err
		}
		m.Payload, err = unmarshalJSONMap(payload)
		if err != nil {
			return out, err
		}
		all = append(all, m)
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	current := map[string]string{}
	for _, m := range all {
		fp, err := fingerprintMessage(m)
		if err != nil {
			return out, err
		}
		out.Expected = append(out.Expected, fp)
		current[fp.ID] = fp.Hash
	}
	valid := out.Summary != "" && len(out.Covered) > 0
	covered := map[string]bool{}
	for _, fp := range out.Covered {
		if covered[fp.ID] || current[fp.ID] != fp.Hash {
			valid = false
		}
		covered[fp.ID] = true
	}
	if !valid {
		out.Summary = ""
		out.Covered = nil
		covered = map[string]bool{}
	}
	for _, m := range all {
		if !covered[m.ID] {
			out.Messages = append(out.Messages, m)
		}
	}
	return out, nil
}
func (s *Store) ContextSnapshot(ctx context.Context, sessionID string) (ContextSnapshot, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ContextSnapshot{}, err
	}
	defer tx.Rollback()
	out, err := readContextSnapshot(ctx, tx, sessionID)
	if err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return out, nil
}

// PrepareContextBoundary covers the prefix actually summarised, extending the
// existing checkpoint when its virtual summary is the first input message.
func PrepareContextBoundary(snapshot ContextSnapshot, summaryCount int) (*ContextBoundary, error) {
	count := summaryCount
	covered := append([]ContextFingerprint(nil), snapshot.Covered...)
	if snapshot.Summary != "" {
		count--
	}
	if count < 0 || count > len(snapshot.Messages) {
		return nil, ErrContextChanged
	}
	for _, m := range snapshot.Messages[:count] {
		fp, err := fingerprintMessage(m)
		if err != nil {
			return nil, err
		}
		covered = append(covered, fp)
	}
	if len(covered) == 0 {
		return nil, ErrContextChanged
	}
	return &ContextBoundary{Version: snapshot.Version, SummaryCount: summaryCount, Expected: append([]ContextFingerprint(nil), snapshot.Expected...), Covered: covered}, nil
}
func commitContextBoundary(ctx context.Context, tx *sql.Tx, sessionID, summary string, boundary *ContextBoundary) error {
	current, err := readContextSnapshot(ctx, tx, sessionID)
	if err != nil {
		return err
	}
	if current.Version != boundary.Version || len(current.Expected) != len(boundary.Expected) {
		return ErrContextChanged
	}
	// Coverage is previous coverage + an exact prefix of the projected native
	// messages, not an arbitrary subset. Backdated *new* IDs remain uncovered.
	prepared, err := PrepareContextBoundary(current, boundary.SummaryCount)
	if err != nil {
		return err
	}
	if len(prepared.Covered) != len(boundary.Covered) {
		return ErrContextChanged
	}
	for i, fp := range prepared.Covered {
		if fp != boundary.Covered[i] {
			return ErrContextChanged
		}
	}
	available := map[string]string{}
	for i, fp := range current.Expected {
		if fp != boundary.Expected[i] {
			return ErrContextChanged
		}
		available[fp.ID] = fp.Hash
	}
	seen := map[string]bool{}
	for _, fp := range boundary.Covered {
		if seen[fp.ID] || available[fp.ID] != fp.Hash {
			return ErrContextChanged
		}
		seen[fp.ID] = true
	}
	if len(seen) == 0 {
		return ErrContextChanged
	}
	raw, err := json.Marshal(boundary.Covered)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `insert into context_checkpoints(session_id,version,summary,covered_json,created_at) values(?,?,?,?,`+defaultNow+`) on conflict(session_id) do update set version=excluded.version,summary=excluded.summary,covered_json=excluded.covered_json,created_at=excluded.created_at`, sessionID, current.Version+1, summary, string(raw))
	return err
}
