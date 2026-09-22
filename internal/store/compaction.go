package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrCompactionInactive = errors.New("compaction requires a running claimed turn")

// BeginCompaction commits the checkpoint with its phase; cancellation cannot
// be overwritten by a later unconditional status update.
func (s *Store) BeginCompaction(ctx context.Context, sessionID, turnID string, payload map[string]any) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `update turns set phase='compacting',updated_at=`+defaultNow+` where id=? and session_id=? and status='running' and phase!='compacting' and exists(select 1 from session_active_turns where session_id=? and turn_id=?)`, turnID, sessionID, sessionID, turnID)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, ErrCompactionInactive
	}
	seq, err := appendCompactionEvent(ctx, tx, sessionID, turnID, "started", payload)
	if err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return seq, nil
}

// FinishCompaction persists outcome, optional summary and phase restoration in
// one transaction. It returns the committed outcome (possibly cancelled).
func (s *Store) FinishCompaction(ctx context.Context, sessionID, turnID string, startedSeq int, outcome, summary string, payload map[string]any) (string, error) {
	return s.FinishCompactionWithBoundary(ctx, sessionID, turnID, startedSeq, outcome, summary, payload, nil)
}

func (s *Store) FinishCompactionWithBoundary(ctx context.Context, sessionID, turnID string, startedSeq int, outcome, summary string, payload map[string]any, boundary *ContextBoundary) (string, error) {
	switch outcome {
	case "completed", "cancelled", "suppressed", "failed":
	default:
		return "", fmt.Errorf("invalid compaction outcome %q", outcome)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	// One terminal outcome per occurrence, even when a turn compacts repeatedly.
	var latestSeq int
	var latestType string
	if err = tx.QueryRowContext(ctx, `select seq,event_type from turn_events where turn_id=? and session_id=? and event_type in ('compaction.started','compaction.completed','compaction.cancelled','compaction.failed','compaction.suppressed') order by seq desc limit 1`, turnID, sessionID).Scan(&latestSeq, &latestType); err != nil {
		return "", err
	}
	if startedSeq <= 0 || latestSeq != startedSeq || latestType != "compaction.started" {
		return "", ErrCompactionInactive
	}
	var status, phase string
	var active bool
	err = tx.QueryRowContext(ctx, `select status,phase,exists(select 1 from session_active_turns where session_id=? and turn_id=?) from turns where id=? and session_id=?`, sessionID, turnID, turnID, sessionID).Scan(&status, &phase, &active)
	if err != nil {
		return "", err
	}
	if status != "running" || phase != "compacting" || !active {
		outcome = "cancelled"
	}
	copyPayload := map[string]any{}
	for k, v := range payload {
		copyPayload[k] = v
	}
	copyPayload["durable_context"] = false
	if outcome == "completed" && boundary != nil {
		if err = commitContextBoundary(ctx, tx, sessionID, summary, boundary); err != nil {
			return "", err
		}
		copyPayload["durable_context"] = true
	}
	copyPayload["outcome"] = outcome
	copyPayload["started_seq"] = startedSeq
	if outcome != "completed" {
		delete(copyPayload, "messages_after")
		delete(copyPayload, "from_hook")
	}
	seq, err := appendCompactionEvent(ctx, tx, sessionID, turnID, outcome, copyPayload)
	if err != nil {
		return "", err
	}
	if outcome == "completed" {
		if summary == "" {
			return "", errors.New("empty compaction summary")
		}
		copyPayload["kind"] = "compaction"
		copyPayload["turn_id"] = turnID
		raw, err := marshalJSON(copyPayload)
		if err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `insert into messages(id,session_id,role,content,payload_json,created_at) values(?,?,'assistant',?,?,`+defaultNow+`)`, fmt.Sprintf("msg_%s_compaction_%d", turnID, seq), sessionID, summary, raw); err != nil {
			return "", err
		}
	}
	if _, err = tx.ExecContext(ctx, `update sessions set updated_at = `+defaultNow+` where id = ?`, sessionID); err != nil {
		return "", err
	}
	// Restore only the phase we own; never resurrect cancelling/terminal work.
	if _, err = tx.ExecContext(ctx, `update turns set phase='running',updated_at=`+defaultNow+` where id=? and session_id=? and status='running' and phase='compacting' and exists(select 1 from session_active_turns where session_id=? and turn_id=?)`, turnID, sessionID, sessionID, turnID); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return outcome, nil
}

// SetClaimedRunningPhase is used on return to inference after context hooks.
// It never writes status, so a cancel committed before this statement wins.
func (s *Store) SetClaimedRunningPhase(ctx context.Context, sessionID, turnID, phase string) error {
	result, err := s.db.ExecContext(ctx, `update turns set phase=?,updated_at=`+defaultNow+` where id=? and session_id=? and status='running' and exists(select 1 from session_active_turns where session_id=? and turn_id=?)`, phase, turnID, sessionID, sessionID, turnID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return context.Canceled
	}
	return nil
}

func appendCompactionEvent(ctx context.Context, tx *sql.Tx, sessionID, turnID, outcome string, payload map[string]any) (int, error) {
	raw, err := marshalJSON(payload)
	if err != nil {
		return 0, err
	}
	var seq int
	if err = tx.QueryRowContext(ctx, `select coalesce(max(seq),0)+1 from turn_events where turn_id=?`, turnID).Scan(&seq); err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,?,?,?,`+defaultNow+`)`, turnID, sessionID, seq, "compaction."+outcome, raw)
	return seq, err
}
