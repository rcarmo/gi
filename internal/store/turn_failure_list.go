package store

import (
	"context"
	"fmt"
)

// ListHeldTurnFailures is bounded and advisory; all actions check ownership and
// current native failure state again. Resolved IDs remain readable via check.
func (s *Store) ListHeldTurnFailures(ctx context.Context, sessionID string, offset, limit int) ([]TurnFailure, error) {
	if sessionID == "" || offset < 0 || limit < 1 || limit > 100 {
		return nil, fmt.Errorf("invalid held failure page")
	}
	rows, err := s.db.QueryContext(ctx, `select turn_id, session_id, failure_kind, hold_state, summary,
 coalesce(resolution_state,''),coalesce(resolution_summary,''),coalesce(resolved_at,''),coalesce(resolved_turn_id,''),retry_admission_token,retry_admission_version,created_at,updated_at
 from turn_failures where session_id=? and hold_state<>'none' order by created_at,turn_id limit ? offset ?`, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TurnFailure
	for rows.Next() {
		var f TurnFailure
		if err = rows.Scan(&f.TurnID, &f.SessionID, &f.FailureKind, &f.HoldState, &f.Summary, &f.ResolutionState, &f.ResolutionSummary, &f.ResolvedAt, &f.ResolvedTurnID, &f.RetryAdmissionToken, &f.RetryAdmissionVersion, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}
