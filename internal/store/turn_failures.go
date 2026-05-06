package store

import (
	"context"
	"fmt"
	"strings"
)

type TurnFailure struct {
	TurnID      string `json:"turn_id"`
	SessionID   string `json:"session_id"`
	FailureKind string `json:"failure_kind"`
	HoldState   string `json:"hold_state"`
	Summary     string `json:"summary"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func normalizeTurnFailureKind(kind string) string {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "unknown"
	}
	return kind
}

func normalizeTurnFailureHoldState(holdState string) string {
	holdState = strings.TrimSpace(holdState)
	if holdState == "" {
		return "none"
	}
	return holdState
}

func (s *Store) UpsertTurnFailure(ctx context.Context, turnID, sessionID, failureKind, holdState, summary string) error {
	failureKind = normalizeTurnFailureKind(failureKind)
	holdState = normalizeTurnFailureHoldState(holdState)
	summary = strings.TrimSpace(summary)
	_, err := s.db.ExecContext(ctx, `
		insert into turn_failures (turn_id, session_id, failure_kind, hold_state, summary, created_at, updated_at)
		values (?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
		on conflict(turn_id) do update set
			session_id = excluded.session_id,
			failure_kind = excluded.failure_kind,
			hold_state = excluded.hold_state,
			summary = excluded.summary,
			updated_at = `+defaultNow+`
	`, turnID, sessionID, failureKind, holdState, summary)
	if err != nil {
		return fmt.Errorf("upsert turn failure: %w", err)
	}
	return nil
}

func (s *Store) ClearTurnFailure(ctx context.Context, turnID string) error {
	_, err := s.db.ExecContext(ctx, `delete from turn_failures where turn_id = ?`, turnID)
	if err != nil {
		return fmt.Errorf("clear turn failure: %w", err)
	}
	return nil
}

func (s *Store) GetTurnFailure(ctx context.Context, turnID string) (*TurnFailure, error) {
	row := s.db.QueryRowContext(ctx, `
		select turn_id, session_id, failure_kind, hold_state, summary, created_at, updated_at
		from turn_failures where turn_id = ?
	`, turnID)
	var out TurnFailure
	if err := row.Scan(&out.TurnID, &out.SessionID, &out.FailureKind, &out.HoldState, &out.Summary, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}
