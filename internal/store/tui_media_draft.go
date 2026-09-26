package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const tuiMediaNamespace = "tui_pending_media_v1"
const MaxTUIPendingMedia = 6

type TUIMediaClaim struct {
	Refs  []MediaRef `json:"refs"`
	Token string     `json:"token"`
}
type TUIMediaDraft struct {
	Pending []MediaRef     `json:"pending"`
	Claim   *TUIMediaClaim `json:"claim,omitempty"`
}

// Serialize read/modify/write with SQLite's immediate transaction. Never save
// an entire frontend snapshot: other clients may have staged newer references.
func (s *Store) updateTUIMediaDraft(ctx context.Context, sessionID string, change func(*sql.Tx, *TUIMediaDraft) error) (TUIMediaDraft, error) {
	var state TUIMediaDraft
	if sessionID == "" {
		return state, errors.New("pending media requires a session")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return state, err
	}
	defer tx.Rollback()
	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT value FROM kv_store WHERE namespace=? AND key=?`, tuiMediaNamespace, sessionID).Scan(&raw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return state, err
	}
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &state); err != nil {
			return state, fmt.Errorf("invalid pending media journal: %w", err)
		}
	}
	count := len(state.Pending)
	if state.Claim != nil {
		count += len(state.Claim.Refs)
		if state.Claim.Token == "" || len(state.Claim.Refs) == 0 {
			return state, errors.New("invalid pending media claim")
		}
	}
	if count > MaxTUIPendingMedia {
		return state, errors.New("pending media journal exceeds limit")
	}
	// A read/recheck with no transition must not rewrite timestamps or acquire
	// unnecessary journal writes on ordinary text submissions.
	before, err := json.Marshal(state)
	if err != nil {
		return TUIMediaDraft{}, err
	}
	if change != nil {
		if err = change(tx, &state); err != nil {
			return TUIMediaDraft{}, err
		}
	}
	after, err := json.Marshal(state)
	if err != nil {
		return TUIMediaDraft{}, err
	}
	if string(before) == string(after) {
		if err = tx.Commit(); err != nil {
			return TUIMediaDraft{}, err
		}
		return state, nil
	}
	if len(state.Pending) == 0 && state.Claim == nil {
		_, err = tx.ExecContext(ctx, `DELETE FROM kv_store WHERE namespace=? AND key=?`, tuiMediaNamespace, sessionID)
	} else {
		raw, err = json.Marshal(state)
		if err != nil {
			return TUIMediaDraft{}, err
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		_, err = tx.ExecContext(ctx, `INSERT INTO kv_store(namespace,key,value,created_at,updated_at) VALUES(?,?,?,?,?) ON CONFLICT(namespace,key) DO UPDATE SET value=excluded.value,updated_at=excluded.updated_at`, tuiMediaNamespace, sessionID, raw, now, now)
	}
	if err != nil {
		return TUIMediaDraft{}, err
	}
	if err = tx.Commit(); err != nil {
		return TUIMediaDraft{}, err
	}
	return state, nil
}

func tuiMediaAdmitted(ctx context.Context, tx *sql.Tx, sessionID, token string) (bool, error) {
	var admitted bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS (
 SELECT 1 FROM turns WHERE session_id=? AND json_extract(metadata_json,'$.tui_media_claim')=?
 UNION ALL SELECT 1 FROM steering_queue WHERE session_id=? AND json_extract(payload_json,'$.tui_media_claim')=?
 UNION ALL SELECT 1 FROM messages WHERE session_id=? AND json_extract(payload_json,'$.tui_media_claim')=?)`, sessionID, token, sessionID, token, sessionID, token).Scan(&admitted)
	return admitted, err
}

// An old process may still finish its admission. Absence after restart is not
// proof of rejection; only confirmed admission can automatically retire a claim.
func (s *Store) LoadTUIMediaDraft(ctx context.Context, sessionID string) (TUIMediaDraft, error) {
	return s.updateTUIMediaDraft(ctx, sessionID, func(tx *sql.Tx, state *TUIMediaDraft) error {
		if state.Claim == nil {
			return nil
		}
		yes, err := tuiMediaAdmitted(ctx, tx, sessionID, state.Claim.Token)
		if err == nil && yes {
			state.Claim = nil
		}
		return err
	})
}
func (s *Store) StageTUIMedia(ctx context.Context, sessionID string, ref MediaRef) (TUIMediaDraft, error) {
	return s.updateTUIMediaDraft(ctx, sessionID, func(tx *sql.Tx, state *TUIMediaDraft) error {
		count := len(state.Pending)
		if state.Claim != nil {
			count += len(state.Claim.Refs)
		}
		if count >= MaxTUIPendingMedia {
			return errors.New("pending limit 6; /attachments or /detach first")
		}
		var owner string
		if err := tx.QueryRowContext(ctx, `SELECT session_id FROM media WHERE id=?`, ref.MediaID).Scan(&owner); err != nil {
			return err
		}
		if owner != sessionID || ref.SessionID != sessionID || ref.ID != MediaRefID(ref.MediaID) {
			return errors.New("pending media belongs to another session")
		}
		state.Pending = append(state.Pending, ref)
		return nil
	})
}
func (s *Store) ClaimTUIMedia(ctx context.Context, sessionID, token string, allow bool) (TUIMediaDraft, error) {
	return s.updateTUIMediaDraft(ctx, sessionID, func(_ *sql.Tx, state *TUIMediaDraft) error {
		if state.Claim != nil {
			return errors.New("attachment admission unresolved; /attachments to inspect")
		}
		if len(state.Pending) == 0 {
			return nil
		}
		if !allow {
			return errors.New("references retained; choose a model or detach before a directed send")
		}
		if token == "" {
			return errors.New("claim token required")
		}
		state.Claim = &TUIMediaClaim{Refs: state.Pending, Token: token}
		state.Pending = nil
		return nil
	})
}
func (s *Store) DetachTUIMedia(ctx context.Context, sessionID, selector, expectedClaim string) (TUIMediaDraft, int, error) {
	removed := 0
	state, err := s.updateTUIMediaDraft(ctx, sessionID, func(_ *sql.Tx, state *TUIMediaDraft) error {
		if selector == "unresolved" {
			if state.Claim != nil && state.Claim.Token != expectedClaim {
				return errors.New("admission changed; inspect /attachments before detaching")
			}
			if state.Claim != nil {
				removed = len(state.Claim.Refs)
				state.Claim = nil
			}
			return nil
		}
		if selector == "all" {
			removed = len(state.Pending)
			state.Pending = nil
			return nil
		}
		for i, ref := range state.Pending {
			if ref.ID == selector {
				state.Pending = append(state.Pending[:i:i], state.Pending[i+1:]...)
				removed = 1
				break
			}
		}
		return nil
	})
	return state, removed, err
}

// Only the still-running frontend that observed SubmitPrompt return may prove
// absence is final and restore refs. Token comparison protects newer claims.
func (s *Store) SettleTUIMedia(ctx context.Context, sessionID, token string, rejected, restoreAbsent bool) (TUIMediaDraft, bool, error) {
	restored := false
	state, err := s.updateTUIMediaDraft(ctx, sessionID, func(tx *sql.Tx, state *TUIMediaDraft) error {
		if state.Claim == nil || state.Claim.Token != token {
			return nil
		}
		if rejected {
			yes, err := tuiMediaAdmitted(ctx, tx, sessionID, token)
			if err != nil {
				return err
			}
			if !yes {
				if !restoreAbsent {
					return nil
				}
				state.Pending = append(state.Claim.Refs, state.Pending...)
				restored = true
			}
		}
		state.Claim = nil
		return nil
	})
	return state, restored, err
}
