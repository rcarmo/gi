package store

import (
	"context"
	"database/sql"
	"errors"
)

// TUIComposerDraft is a transactionally consistent view of both journals. No
// frontend uses it yet: claim provenance and callback ownership are separate.
type TUIComposerDraft struct {
	Text  TUITextDraft
	Media TUIMediaDraft
}

func (s *Store) updateTUIComposerDraft(ctx context.Context, sessionID string, change func(*sql.Tx, *TUITextDraft, *TUIMediaDraft) error) (TUIComposerDraft, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TUIComposerDraft{}, err
	}
	defer tx.Rollback()
	var media TUIMediaDraft
	text, err := updateTUITextDraftTx(ctx, tx, sessionID, func(tx *sql.Tx, text *TUITextDraft) error {
		var err error
		media, err = updateTUIMediaDraftTx(ctx, tx, sessionID, func(tx *sql.Tx, media *TUIMediaDraft) error { return change(tx, text, media) })
		return err
	})
	if err != nil {
		return TUIComposerDraft{}, err
	}
	if err = tx.Commit(); err != nil {
		return TUIComposerDraft{}, err
	}
	return TUIComposerDraft{Text: text, Media: media}, nil
}

// Load never settles or restores a claim. Reading both journals together is
// necessary to avoid observing a mixed pre/post-admission pair.
func (s *Store) LoadTUIComposerDraft(ctx context.Context, sessionID string) (TUIComposerDraft, error) {
	return s.updateTUIComposerDraft(ctx, sessionID, func(_ *sql.Tx, _ *TUITextDraft, _ *TUIMediaDraft) error { return nil })
}

// Claim atomically clears the editable text and staged media. A write failure
// in either journal leaves both unchanged. The same random token owns the pair.
func (s *Store) ClaimTUIComposerDraft(ctx context.Context, sessionID string, expected int64) (TUIComposerDraft, error) {
	return s.updateTUIComposerDraft(ctx, sessionID, func(tx *sql.Tx, text *TUITextDraft, media *TUIMediaDraft) error {
		if media.Claim != nil {
			return ErrTUIDraftHeld
		}
		if err := claimTUIText(text, expected); err != nil {
			return err
		}
		if len(media.Pending) == 0 {
			return nil
		}
		for _, ref := range media.Pending {
			var owner string
			if err := tx.QueryRowContext(ctx, `select session_id from media where id=?`, ref.MediaID).Scan(&owner); err != nil {
				return err
			}
			if owner != sessionID || ref.SessionID != sessionID || ref.ID != MediaRefID(ref.MediaID) {
				return errors.New("composer media belongs to another session")
			}
		}
		text.Claim.Media = true
		media.Claim = &TUIMediaClaim{Token: text.Claim.Token, Refs: media.Pending, Text: true}
		media.Pending = nil
		return nil
	})
}

func checkTUIComposerClaim(text *TUITextDraft, media *TUIMediaDraft, token string) error {
	if text.Claim == nil || text.Claim.Token != token {
		return ErrTUIDraftConflict
	}
	if text.Claim.Media {
		if media.Claim == nil || !media.Claim.Text || media.Claim.Token != token {
			return ErrTUIDraftHeld
		}
	} else if media.Claim != nil {
		return ErrTUIDraftHeld
	}
	return nil
}

// Both tokens must occur in one same-session receipt, not separate turns.
func tuiComposerAdmitted(ctx context.Context, tx *sql.Tx, sessionID, token string, withMedia, returnedSuccess bool) (bool, error) {
	if !withMedia {
		return tuiTextAdmitted(ctx, tx, sessionID, token, returnedSuccess)
	}
	var admitted bool
	err := tx.QueryRowContext(ctx, `select exists(
 select 1 from turns t where session_id=? and json_extract(metadata_json,'$.tui_text_claim')=? and json_extract(metadata_json,'$.tui_media_claim')=? and (? or exists(select 1 from turn_events e where e.turn_id=t.id and e.event_type='turn.submitted'))
 union all select 1 from steering_queue where session_id=? and json_extract(payload_json,'$.tui_text_claim')=? and json_extract(payload_json,'$.tui_media_claim')=?
 union all select 1 from messages where session_id=? and json_extract(payload_json,'$.tui_text_claim')=? and json_extract(payload_json,'$.tui_media_claim')=?)`, sessionID, token, token, returnedSuccess, sessionID, token, token, sessionID, token, token).Scan(&admitted)
	return admitted, err
}

func (s *Store) settleTUIComposerDraft(ctx context.Context, sessionID, token string, returned, rejected bool) (TUIComposerDraft, error) {
	return s.updateTUIComposerDraft(ctx, sessionID, func(tx *sql.Tx, text *TUITextDraft, media *TUIMediaDraft) error {
		if err := checkTUIComposerClaim(text, media, token); err != nil {
			return err
		}
		admitted, err := tuiComposerAdmitted(ctx, tx, sessionID, token, text.Claim.Media, returned && !rejected)
		if err != nil {
			return err
		}
		if admitted {
			if text.Claim.Media {
				media.Claim = nil
			}
			text.Claim = nil
			return nil
		}
		if !returned || !rejected {
			return nil
		}
		possible, err := tuiTextAdmitted(ctx, tx, sessionID, token, true)
		if err != nil {
			return err
		}
		if text.Claim.Media {
			m, err := tuiMediaAdmitted(ctx, tx, sessionID, token)
			if err != nil {
				return err
			}
			possible = possible || m
		}
		if possible {
			return nil
		} // partial receipts are uncertain, never restored
		if text.Revision == text.Claim.Revision {
			text.TUITextSnapshot = text.Claim.TUITextSnapshot
			if text.Claim.Media {
				media.Pending = append(media.Claim.Refs, media.Pending...)
				media.Claim = nil
			}
			text.Claim = nil
		} else {
			text.Claim.Rejected = true
		}
		return nil
	})
}

func (s *Store) ReconcileTUIComposerDraft(ctx context.Context, sessionID, token string) (TUIComposerDraft, error) {
	return s.settleTUIComposerDraft(ctx, sessionID, token, false, false)
}

// Only the original caller after synchronous submission returns may Finish.
func (s *Store) FinishTUIComposerDraft(ctx context.Context, sessionID, token string, rejected bool) (TUIComposerDraft, error) {
	return s.settleTUIComposerDraft(ctx, sessionID, token, true, rejected)
}

// ResolveRejected performs explicit restore (blank editor only) or discard of
// the rejected snapshot and its references. Stored media bytes are never deleted.
func (s *Store) ResolveRejectedTUIComposerDraft(ctx context.Context, sessionID, token string, expected int64, restore bool) (TUIComposerDraft, error) {
	return s.updateTUIComposerDraft(ctx, sessionID, func(_ *sql.Tx, text *TUITextDraft, media *TUIMediaDraft) error {
		if err := checkTUIComposerClaim(text, media, token); err != nil {
			return err
		}
		if text.Revision != expected || (restore && text.Text != "") {
			return ErrTUIDraftConflict
		}
		if !text.Claim.Rejected {
			return ErrTUIDraftHeld
		}
		if restore {
			text.TUITextSnapshot = text.Claim.TUITextSnapshot
		}
		if text.Claim.Media {
			if restore {
				media.Pending = append(media.Claim.Refs, media.Pending...)
			}
			media.Claim = nil
		}
		text.Claim = nil
		return nil
	})
}
