package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// RestoreQueuedTUITextDraft moves all eligible text-only queued turns into the
// current terminal draft. The journal write and queue cancellation commit in
// one transaction, so a failed save or competing claim cannot lose delivery.
// The caller must supply the editor snapshot it has actually displayed.
func (s *Store) RestoreQueuedTUITextDraft(ctx context.Context, sessionID string, expected TUITextDraft) (TUITextDraft, []string, error) {
	return s.restoreQueuedTUITextDraft(ctx, sessionID, "", "", expected, false)
}

// AbortActiveTurnAndRestoreQueuedTUITextDraft is called only by the runner
// owning activeTurnID. The cancellation request, draft write and queue
// removals share one SQLite writer transaction; a conflict cancels nothing.
// The runner must hold its session lock until it signals its worker after
// this transaction commits.
func (s *Store) AbortActiveTurnAndRestoreQueuedTUITextDraft(ctx context.Context, sessionID, activeTurnID, claimToken string, expected TUITextDraft) (TUITextDraft, []string, error) {
	if activeTurnID == "" || claimToken == "" {
		return TUITextDraft{}, nil, ErrQueueConflict
	}
	return s.restoreQueuedTUITextDraft(ctx, sessionID, activeTurnID, claimToken, expected, false)
}

// RestoreAllQueuedTUITextDraft is Alt+Up: withdraw still-pending text steering
// first, then follow-ups, without aborting the current run. Already consumed
// steering, media and maintenance work are never silently retracted.
func (s *Store) RestoreAllQueuedTUITextDraft(ctx context.Context, sessionID string, expected TUITextDraft) (TUITextDraft, []string, error) {
	return s.restoreQueuedTUITextDraft(ctx, sessionID, "", "", expected, true)
}

func (s *Store) restoreQueuedTUITextDraft(ctx context.Context, sessionID, activeTurnID, claimToken string, expected TUITextDraft, allPending bool) (TUITextDraft, []string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	defer tx.Rollback()
	// Claim the writer slot before reading the queue and both draft journals.
	res, err := tx.ExecContext(ctx, `update sessions set id=id where id=?`, sessionID)
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	if changed, err := res.RowsAffected(); err != nil || changed != 1 {
		return TUITextDraft{}, nil, ErrQueueConflict
	}
	if activeTurnID != "" {
		var raw string
		if err := tx.QueryRowContext(ctx, `select t.metadata_json from session_active_turns a join turns t on t.id=a.turn_id
			where a.session_id=? and a.turn_id=? and a.claim_token=? and t.session_id=? and t.status='running'`, sessionID, activeTurnID, claimToken, sessionID).Scan(&raw); err != nil {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
		metadata, err := unmarshalJSONMap(raw)
		if err != nil {
			return TUITextDraft{}, nil, err
		}
		if metadata["media"] != nil || metadata["tui_media_claim"] != nil || metadata["initial_steering"] != nil || metadata["operation"] == "manual_compaction" {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
	}
	var steering []SteeringMessage
	if allPending {
		steering, err = pendingTextSteeringForRestore(ctx, tx, sessionID)
		if err != nil {
			return TUITextDraft{}, nil, err
		}
	}
	rows, err := tx.QueryContext(ctx, `select id, prompt, metadata_json from turns where session_id=? and status='queued' order by queue_position, created_at, id`, sessionID)
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	var ids, prompts []string
	for rows.Next() {
		var id, prompt, raw string
		if err = rows.Scan(&id, &prompt, &raw); err != nil {
			break
		}
		var metadata map[string]any
		metadata, err = unmarshalJSONMap(raw)
		if err != nil {
			break
		}
		if metadata["media"] != nil || metadata["tui_media_claim"] != nil || metadata["operation"] == "manual_compaction" || strings.TrimSpace(prompt) == "" || !utf8.ValidString(prompt) {
			err = fmt.Errorf("%w: queued media or non-text turn requires /queue", ErrQueueConflict)
			break
		}
		ids = append(ids, id)
		prompts = append(prompts, prompt)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	if len(ids) == 0 && len(steering) == 0 && activeTurnID == "" {
		return TUITextDraft{}, nil, sql.ErrNoRows
	}
	// Pi's visual/edit order is steering, follow-up, then the current editor.
	var steeringPrompts []string
	for _, msg := range steering {
		steeringPrompts = append(steeringPrompts, msg.Content)
		if msg.SourceQueueID != "" {
			ids = append(ids, msg.SourceQueueID)
		}
	}
	prompts = append(steeringPrompts, prompts...)
	// Restoration is an editor action, not a cross-session dequeue: leave
	// another frontend's or crashed run's queue untouched until its claim has
	// been settled by the normal recovery path.
	var claimed bool
	if err := tx.QueryRowContext(ctx, `select exists(select 1 from turns t join session_active_turns a on a.turn_id=t.id where t.session_id=? and t.status='queued')`, sessionID).Scan(&claimed); err != nil {
		return TUITextDraft{}, nil, err
	}
	if claimed {
		return TUITextDraft{}, nil, ErrQueueConflict
	}
	if activeTurnID != "" {
		// A second frontend can observe a stale runner after another worker
		// has already claimed queued work. Require the exact active claim.
		var owner string
		if err := tx.QueryRowContext(ctx, `select claim_token from session_active_turns where session_id=? and turn_id=?`, sessionID, activeTurnID).Scan(&owner); err != nil || owner != claimToken {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
	}
	// An active Steer (including one not backed by a queued turn) is a
	// separate delivery path. Do not claim to clear the entire queue while it
	// could still be delivered by the running turn.
	var bound bool
	if err := tx.QueryRowContext(ctx, `select exists(select 1 from steering_queue where session_id=? and status in ('queued','claimed'))`, sessionID).Scan(&bound); err != nil {
		return TUITextDraft{}, nil, err
	}
	if bound && !allPending {
		return TUITextDraft{}, nil, ErrQueueConflict
	}
	if activeTurnID != "" {
		// A completed or asynchronously claimed steering row is still owned
		// by this run. Its prompt cannot be represented by the queued text
		// draft; do not report an edit-all cancellation that drops it.
		var taken bool
		if err := tx.QueryRowContext(ctx, `select exists(select 1 from steering_queue where session_id=? and turn_id=? and status in ('queued','claimed','dequeued'))`, sessionID, activeTurnID).Scan(&taken); err != nil {
			return TUITextDraft{}, nil, err
		}
		if taken {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
	}
	media, err := updateTUIMediaDraftTx(ctx, tx, sessionID, nil)
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	if media.Claim != nil || len(media.Pending) > 0 {
		return TUITextDraft{}, nil, ErrTUIDraftHeld
	}
	prefix := ""
	if len(prompts) > 0 {
		prefix = strings.Join(prompts, "\n\n") + "\n\n"
		if expected.Text == "" {
			prefix = strings.TrimSuffix(prefix, "\n\n")
		}
	}
	restored := TUITextSnapshot{Text: prefix + expected.Text, Cursor: utf8.RuneCountInString(prefix) + expected.Cursor}
	if !validTUITextSnapshot(restored) {
		return TUITextDraft{}, nil, ErrTUIDraftConflict
	}
	updated, err := updateTUITextDraftTx(ctx, tx, sessionID, func(_ *sql.Tx, state *TUITextDraft) error {
		if state.Claim != nil {
			return ErrTUIDraftHeld
		}
		if state.Revision != expected.Revision || state.TUITextSnapshot != expected.TUITextSnapshot {
			return ErrTUIDraftConflict
		}
		if state.Revision >= math.MaxInt64-2 {
			return errors.New("terminal draft revision exhausted")
		}
		state.TUITextSnapshot = restored
		return nil
	})
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	for _, msg := range steering {
		res, err := tx.ExecContext(ctx, `update steering_queue set status='cancelled', updated_at=`+defaultNow+` where id=? and session_id=? and status='queued'`, msg.ID, sessionID)
		if err != nil {
			return TUITextDraft{}, nil, err
		}
		if changed, err := res.RowsAffected(); err != nil || changed != 1 {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
	}
	for _, id := range ids {
		res, err := tx.ExecContext(ctx, `update turns set status='cancelled', phase='aborted', finished_at=`+defaultNow+`, updated_at=`+defaultNow+` where id=? and session_id=? and status in ('queued','steering') and not exists(select 1 from session_active_turns where turn_id=?)`, id, sessionID, id)
		if err != nil {
			return TUITextDraft{}, nil, err
		}
		if changed, err := res.RowsAffected(); err != nil || changed != 1 {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
		payload, err := marshalJSON(map[string]any{"phase": "cancel", "checkpoint": true, "queued": true, "reason": "queue_restore", "status": "cancelled", "turn_phase": "aborted"})
		if err != nil {
			return TUITextDraft{}, nil, err
		}
		if _, err := tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),'turn.cancelled',?,`+defaultNow+`)`, id, sessionID, id, payload); err != nil {
			return TUITextDraft{}, nil, err
		}
	}
	if activeTurnID != "" {
		res, err := tx.ExecContext(ctx, `update turns set status='cancelling', phase='cancelling', updated_at=`+defaultNow+`
			where id=? and session_id=? and status='running' and exists(select 1 from session_active_turns where session_id=? and turn_id=? and claim_token=?)`, activeTurnID, sessionID, sessionID, activeTurnID, claimToken)
		if err != nil {
			return TUITextDraft{}, nil, err
		}
		if changed, err := res.RowsAffected(); err != nil || changed != 1 {
			return TUITextDraft{}, nil, ErrQueueConflict
		}
		payload, err := marshalJSON(map[string]any{"phase": "cancel", "checkpoint": true, "reason": "escape_restore", "status": "cancelling", "turn_phase": "cancelling", "failure_kind": ""})
		if err != nil {
			return TUITextDraft{}, nil, err
		}
		if _, err := tx.ExecContext(ctx, `insert into turn_events(turn_id,session_id,seq,event_type,payload_json,created_at) values(?,?,coalesce((select max(seq)+1 from turn_events where turn_id=?),1),'turn.cancelling',?,`+defaultNow+`)`, activeTurnID, sessionID, activeTurnID, payload); err != nil {
			return TUITextDraft{}, nil, err
		}
	}
	res, err = tx.ExecContext(ctx, `update sessions set state_json=json_set(coalesce(nullif(state_json,''),'{}'),'$.queue_count',0),updated_at=`+defaultNow+` where id=?`, sessionID)
	if err != nil {
		return TUITextDraft{}, nil, err
	}
	if changed, err := res.RowsAffected(); err != nil || changed != 1 {
		return TUITextDraft{}, nil, ErrQueueConflict
	}
	if err := tx.Commit(); err != nil {
		return TUITextDraft{}, nil, err
	}
	for _, msg := range steering {
		if msg.SourceQueueID == "" {
			ids = append(ids, fmt.Sprintf("steering:%d", msg.ID))
		}
	}
	return updated, ids, nil
}
