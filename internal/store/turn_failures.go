package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store/internalx"
)

type TurnFailure struct {
	TurnID            string `json:"turn_id"`
	SessionID         string `json:"session_id"`
	FailureKind       string `json:"failure_kind"`
	HoldState         string `json:"hold_state"`
	Summary           string `json:"summary"`
	ResolutionState   string `json:"resolution_state,omitempty"`
	ResolutionSummary string `json:"resolution_summary,omitempty"`
	ResolvedAt        string `json:"resolved_at,omitempty"`
	ResolvedTurnID    string `json:"resolved_turn_id,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
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
	failureKind = normalizeTurnFailureKind(failureKind)
	holdState = normalizeTurnFailureHoldState(holdState)
	summary = strings.TrimSpace(summary)
	_, err := s.db.ExecContext(ctx, `
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
		select turn_id, session_id, failure_kind, hold_state, summary,
		       coalesce(resolution_state,''), coalesce(resolution_summary,''), coalesce(resolved_at,''), coalesce(resolved_turn_id,''),
		       created_at, updated_at
		from turn_failures where turn_id = ?
	`, turnID)
	var out TurnFailure
	if err := row.Scan(&out.TurnID, &out.SessionID, &out.FailureKind, &out.HoldState, &out.Summary, &out.ResolutionState, &out.ResolutionSummary, &out.ResolvedAt, &out.ResolvedTurnID, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Store) HoldTurnFailure(ctx context.Context, turnID, holdState, summary string) error {
	holdState = normalizeTurnFailureHoldState(holdState)
	if holdState == "none" {
		return fmt.Errorf("hold turn failure: hold state must not be none")
	}
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	if turnRec.Status != "failed" && turnRec.Status != "aborted" && turnRec.Status != "cancelled" {
		return fmt.Errorf("hold turn failure: turn %s status %s is not terminal", turnID, turnRec.Status)
	}
	failureKind := "manual_hold"
	if failureRec, err := s.GetTurnFailure(ctx, turnID); err == nil {
		failureKind = failureRec.FailureKind
	}
	if strings.TrimSpace(summary) == "" {
		summary = fmt.Sprintf("turn %s placed on %s hold", turnID, holdState)
	}
	if err := s.UpsertTurnFailure(ctx, turnID, turnRec.SessionID, failureKind, holdState, summary); err != nil {
		return err
	}
	if err := s.UpdateTurnStatusAndPhase(ctx, turnID, turnRec.Status, "held_for_retry_or_skip"); err != nil {
		return err
	}
	return nil
}

func (s *Store) ResolveTurnFailure(ctx context.Context, turnID, resolutionState, resolutionSummary, resolvedTurnID string) error {
	resolutionState = normalizeTurnFailureResolutionState(resolutionState)
	turnRec, err := s.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	if _, err := s.GetTurnFailure(ctx, turnID); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		update turn_failures
		set hold_state = 'none',
		    resolution_state = ?,
		    resolution_summary = ?,
		    resolved_at = `+defaultNow+`,
		    resolved_turn_id = ?,
		    updated_at = `+defaultNow+`
		where turn_id = ?
	`, resolutionState, strings.TrimSpace(resolutionSummary), internalx.NilIfEmpty(resolvedTurnID), turnID)
	if err != nil {
		return fmt.Errorf("resolve turn failure: %w", err)
	}
	if turnRec.Phase == "held_for_retry_or_skip" {
		if err := s.UpdateTurnStatusAndPhase(ctx, turnID, turnRec.Status, turnPhaseForStatus(turnRec.Status)); err != nil {
			return err
		}
	}
	return nil
}
