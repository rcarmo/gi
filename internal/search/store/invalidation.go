package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Invalidate records a durable request for one configured scope. Call once per
// relevant scope; an empty path list means an explicit whole-scope event (for
// example watcher overflow). Paths can name files or removed/renamed folders.
// No filesystem lookup, scan, automatic retry, or writer-lease acquisition occurs.
// Returning true means a revision was durably recorded, not that refresh ran.
func (s *RefreshStore) Invalidate(ctx context.Context, c ScopeConfig, paths []string) (bool, error) {
	return s.InvalidateScopes(ctx, []ScopeConfig{c}, paths)
}

// InvalidateScopes atomically fans one event out to a bounded set of configured
// scopes for one workspace. Failure in any scope rolls back the entire event.
func (s *RefreshStore) InvalidateScopes(ctx context.Context, configs []ScopeConfig, paths []string) (bool, error) {
	if len(configs) == 0 || len(configs) > 3 {
		return false, fmt.Errorf("invalidation requires 1–3 scopes")
	}
	seen := map[string]bool{}
	for _, c := range configs {
		if c.fingerprint == "" || c.workspace != configs[0].workspace || seen[c.scope] {
			return false, fmt.Errorf("invalid invalidation scope set")
		}
		seen[c.scope] = true
	}
	if len(paths) > 10000 {
		return false, fmt.Errorf("invalidation path limit exceeded")
	}
	for _, p := range paths {
		if !localIndexPath(p) {
			return false, fmt.Errorf("invalid invalidation path %q", p)
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	affected := false
	for _, c := range configs {
		hit, err := invalidateScope(ctx, tx, c, paths)
		if err != nil {
			return false, err
		}
		affected = affected || hit
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return affected, nil
}

func invalidateScope(ctx context.Context, tx *sql.Tx, c ScopeConfig, paths []string) (bool, error) {
	roots := c.Roots()
	// Include the last committed roots during a configuration transition, so a
	// change to an old root cannot be hidden before the replacement commits.
	var persisted string
	err := tx.QueryRowContext(ctx, `SELECT sc.roots_json FROM workspace_index_scopes sc JOIN workspace_index_workspaces w ON w.id=sc.workspace_id WHERE w.root_identity=? AND sc.scope=?`, c.workspace, c.scope).Scan(&persisted)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == nil {
		var old []string
		if err = json.Unmarshal([]byte(persisted), &old); err != nil {
			return false, err
		}
		roots = append(roots, old...)
	}
	affected := len(paths) == 0
	for _, p := range paths {
		for _, root := range roots {
			if underIndexRoot(p, root) || underIndexRoot(root, p) {
				affected = true
			}
		}
	}
	if !affected {
		return false, nil
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO workspace_index_workspaces(root_identity) VALUES(?) ON CONFLICT(root_identity) DO NOTHING`, c.workspace); err != nil {
		return false, err
	}
	var id int64
	if err = tx.QueryRowContext(ctx, "SELECT id FROM workspace_index_workspaces WHERE root_identity=?", c.workspace).Scan(&id); err != nil {
		return false, err
	}
	raw, _ := json.Marshal(c.roots)
	if _, err = tx.ExecContext(ctx, `INSERT INTO workspace_index_scopes(workspace_id,scope,config_hash,roots_json,state,updated_at_ms) VALUES(?,?,?,?,'never_indexed',`+indexNowMS+`) ON CONFLICT(workspace_id,scope) DO NOTHING`, id, c.scope, c.fingerprint, string(raw)); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO workspace_index_invalidations(workspace_id,scope,requested_revision) VALUES(?,?,1)
 ON CONFLICT(workspace_id,scope) DO UPDATE SET requested_revision=requested_revision+1`, id, c.scope); err != nil {
		return false, err
	}
	// Preserve live ownership and error detail. Neither a change event nor reads
	// clear a prior failure; the scheduler can use revision counters to retry.
	if _, err = tx.ExecContext(ctx, `UPDATE workspace_index_scopes SET state=CASE WHEN state='ready' THEN 'stale' ELSE state END,updated_at_ms=`+indexNowMS+` WHERE workspace_id=? AND scope=?`, id, c.scope); err != nil {
		return false, err
	}
	return true, nil
}
