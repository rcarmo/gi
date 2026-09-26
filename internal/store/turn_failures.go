package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store/internalx"
)

type TurnFailure struct {
	TurnID                string `json:"turn_id"`
	SessionID             string `json:"session_id"`
	FailureKind           string `json:"failure_kind"`
	HoldState             string `json:"hold_state"`
	Summary               string `json:"summary"`
	ResolutionState       string `json:"resolution_state,omitempty"`
	ResolutionSummary     string `json:"resolution_summary,omitempty"`
	ResolvedAt            string `json:"resolved_at,omitempty"`
	ResolvedTurnID        string `json:"resolved_turn_id,omitempty"`
	RetryAdmissionToken   string `json:"retry_admission_token,omitempty"`
	RetryAdmissionVersion int    `json:"retry_admission_version,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

func normalizeTurnFailureKind(kind string) string {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "unknown"
	}
	return kind
}

func normalizeTurnFailureHoldState(holdState string) string {
	holdState = strings.TrimSpace(strings.ToLower(holdState))
	if holdState == "" {
		return "none"
	}
	return holdState
}

func normalizeTurnFailureResolutionState(state string) string {
	return strings.TrimSpace(strings.ToLower(state))
}

func (s *Store) UpsertTurnFailure(ctx context.Context, turnID, sessionID, failureKind, holdState, summary string) error {
	return upsertTurnFailure(ctx, s.db, turnID, sessionID, failureKind, holdState, summary)
}

func upsertTurnFailure(ctx context.Context, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, turnID, sessionID, failureKind, holdState, summary string) error {
	failureKind = normalizeTurnFailureKind(failureKind)
	holdState = normalizeTurnFailureHoldState(holdState)
	summary = strings.TrimSpace(summary)
	result, err := db.ExecContext(ctx, `
		insert into turn_failures (turn_id, session_id, failure_kind, hold_state, summary, resolution_state, resolution_summary, resolved_at, resolved_turn_id, created_at, updated_at)
		values (?, ?, ?, ?, ?, '', '', null, null, `+defaultNow+`, `+defaultNow+`)
		on conflict(turn_id) do update set
			session_id = excluded.session_id,
			failure_kind = excluded.failure_kind,
			hold_state = excluded.hold_state,
			summary = excluded.summary,
			resolution_state = '',
			resolution_summary = '',
			resolved_at = null,
			resolved_turn_id = null,
			retry_admission_token = '',
			retry_admission_version = 0,
			updated_at = `+defaultNow+`
		where coalesce(turn_failures.resolution_state,'') = ''
	`, turnID, sessionID, failureKind, holdState, summary)
	if err != nil {
		return fmt.Errorf("upsert turn failure: %w", err)
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return ErrFailureConflict
	}
	return nil
}

func (s *Store) ClearTurnFailure(ctx context.Context, turnID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var protected int
	if err = tx.QueryRowContext(ctx, `select count(*) from turn_failures where turn_id = ? and (hold_state <> 'none' or coalesce(resolution_state,'') <> '')`, turnID).Scan(&protected); err != nil {
		return err
	}
	if protected != 0 {
		return ErrFailureConflict
	}
	if _, err = tx.ExecContext(ctx, `delete from turn_failures where turn_id = ?`, turnID); err != nil {
		return fmt.Errorf("clear turn failure: %w", err)
	}
	return tx.Commit()
}

func (s *Store) GetTurnFailure(ctx context.Context, turnID string) (*TurnFailure, error) {
	row := s.db.QueryRowContext(ctx, `
		select turn_id, session_id, failure_kind, hold_state, summary,
		       coalesce(resolution_state,''), coalesce(resolution_summary,''), coalesce(resolved_at,''), coalesce(resolved_turn_id,''), coalesce(retry_admission_token,''), retry_admission_version,
		       created_at, updated_at
		from turn_failures where turn_id = ?
	`, turnID)
	var out TurnFailure
	if err := row.Scan(&out.TurnID, &out.SessionID, &out.FailureKind, &out.HoldState, &out.Summary, &out.ResolutionState, &out.ResolutionSummary, &out.ResolvedAt, &out.ResolvedTurnID, &out.RetryAdmissionToken, &out.RetryAdmissionVersion, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Store) HoldTurnFailure(ctx context.Context, turnID, holdState, summary string) error {
	holdState = normalizeTurnFailureHoldState(holdState)
	if holdState == "none" {
		return fmt.Errorf("hold turn failure: hold state must not be none")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status, sessionID string
	if err = tx.QueryRowContext(ctx, `select status, session_id from turns where id = ?`, turnID).Scan(&status, &sessionID); err != nil {
		return err
	}
	if status != "failed" && status != "aborted" && status != "cancelled" {
		return fmt.Errorf("hold turn failure: turn %s status %s is not terminal", turnID, status)
	}
	failureKind := "manual_hold"
	if err = tx.QueryRowContext(ctx, `select failure_kind from turn_failures where turn_id = ?`, turnID).Scan(&failureKind); err != nil && err != sql.ErrNoRows {
		return err
	}
	if strings.TrimSpace(summary) == "" {
		summary = fmt.Sprintf("turn %s placed on %s hold", turnID, holdState)
	}
	if err := upsertTurnFailure(ctx, tx, turnID, sessionID, failureKind, holdState, summary); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `update turns set phase = 'held_for_retry_or_skip', updated_at = `+defaultNow+` where id = ?`, turnID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ResolveTurnFailure(ctx context.Context, turnID, resolutionState, resolutionSummary, resolvedTurnID string) error {
	resolutionState = normalizeTurnFailureResolutionState(resolutionState)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status, phase string
	if err = tx.QueryRowContext(ctx, `select status, phase from turns where id = ?`, turnID).Scan(&status, &phase); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		update turn_failures
		set hold_state = 'none',
		    resolution_state = ?,
		    resolution_summary = ?,
		    resolved_at = `+defaultNow+`,
		    resolved_turn_id = ?,
		    updated_at = `+defaultNow+`
		where turn_id = ? and hold_state <> 'none' and coalesce(resolution_state,'') = ''
	`, resolutionState, strings.TrimSpace(resolutionSummary), internalx.NilIfEmpty(resolvedTurnID), turnID)
	if err != nil {
		return fmt.Errorf("resolve turn failure: %w", err)
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return ErrFailureConflict
	}
	if phase == "held_for_retry_or_skip" {
		if _, err = tx.ExecContext(ctx, `update turns set phase = ?, updated_at = `+defaultNow+` where id = ?`, turnPhaseForStatus(status), turnID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
