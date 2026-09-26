package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Fence only explicit selection fields, not unrelated runtime state updates.
func SessionThinkingToken(sessionID string, state map[string]any) string {
	keys := []string{"selected_model", "selected_provider", "model", "provider", "thinking_level", "thinking_model", "selection_revision"}
	values := []any{sessionID}
	for _, key := range keys {
		values = append(values, state[key])
	}
	raw, _ := json.Marshal(values)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (s *Store) SelectSessionThinking(ctx context.Context, id, expected, model, level string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Take the SQLite writer slot before reading the revision, avoiding a stale
	// read-to-write upgrade across separate Store instances.
	if _, err = tx.ExecContext(ctx, `update sessions set id=id where id=?`, id); err != nil {
		return err
	}
	var raw string
	if err = tx.QueryRowContext(ctx, `select state_json from sessions where id=?`, id).Scan(&raw); err != nil {
		return err
	}
	state, err := unmarshalJSONMap(raw)
	if err != nil {
		return err
	}
	if expected == "" || SessionThinkingToken(id, state) != expected {
		return ErrContextChanged
	}
	patch, err := json.Marshal(map[string]any{"thinking_level": level, "thinking_model": model})
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update sessions set state_json=json_set(json_patch(state_json,?),'$.selection_revision',coalesce(json_extract(state_json,'$.selection_revision'),0)+1),updated_at=`+defaultNow+` where id=?`, string(patch), id); err != nil {
		return err
	}
	return tx.Commit()
}
