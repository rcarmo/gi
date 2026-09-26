package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
	"unicode/utf8"
)

const tuiTextDraftNamespace = "tui_text_draft_v1"

var ErrTUIDraftConflict = errors.New("terminal draft changed; reload before editing or submitting")
var ErrTUIDraftHeld = errors.New("terminal draft admission unresolved; never resend the held snapshot")

type TUITextSnapshot struct {
	Text   string `json:"text"`
	Cursor int    `json:"cursor"` // rune offset, matching the terminal editor
}
type TUITextClaim struct {
	TUITextSnapshot
	Token      string `json:"token"`
	Revision   int64  `json:"revision"`             // editable revision immediately after claiming
	Rejected   bool   `json:"rejected,omitempty"`   // live caller proved admission absent
	Media      bool   `json:"media,omitempty"`      // must settle together with media journal
	Dispatched bool   `json:"dispatched,omitempty"` // only one caller may enter submission
}
type TUITextDraft struct {
	TUITextSnapshot
	Revision int64         `json:"revision"`
	Claim    *TUITextClaim `json:"claim,omitempty"`
}

func validTUITextSnapshot(s TUITextSnapshot) bool {
	return utf8.ValidString(s.Text) && s.Cursor >= 0 && s.Cursor <= utf8.RuneCountInString(s.Text)
}
func decodeTUITextDraft(raw []byte) (TUITextDraft, error) {
	var state TUITextDraft
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &state); err != nil {
			return state, fmt.Errorf("invalid terminal draft journal: %w", err)
		}
	}
	if !validTUITextSnapshot(state.TUITextSnapshot) || state.Revision < 0 {
		return TUITextDraft{}, errors.New("invalid terminal draft snapshot")
	}
	if c := state.Claim; c != nil {
		if c.Token == "" || c.Text == "" || !validTUITextSnapshot(c.TUITextSnapshot) || c.Revision <= 0 || c.Revision > state.Revision {
			return TUITextDraft{}, errors.New("invalid terminal draft claim")
		}
	}
	return state, nil
}

// Load is read-only. Admission recovery is a separate explicit operation;
// unknown outcomes are never returned as editable text.
func (s *Store) LoadTUITextDraft(ctx context.Context, sessionID string) (TUITextDraft, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `select value from kv_store where namespace=? and key=? and exists(select 1 from sessions where id=?)`, tuiTextDraftNamespace, sessionID, sessionID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err = s.GetSession(ctx, sessionID); err != nil {
			return TUITextDraft{}, err
		}
		return TUITextDraft{}, nil
	}
	if err != nil {
		return TUITextDraft{}, err
	}
	return decodeTUITextDraft(raw)
}

func (s *Store) updateTUITextDraft(ctx context.Context, sessionID string, change func(*sql.Tx, *TUITextDraft) error) (TUITextDraft, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TUITextDraft{}, err
	}
	defer tx.Rollback()
	state, err := updateTUITextDraftTx(ctx, tx, sessionID, change)
	if err != nil {
		return TUITextDraft{}, err
	}
	if err = tx.Commit(); err != nil {
		return TUITextDraft{}, err
	}
	return state, nil
}

func updateTUITextDraftTx(ctx context.Context, tx *sql.Tx, sessionID string, change func(*sql.Tx, *TUITextDraft) error) (TUITextDraft, error) {
	var owner string
	var err error
	if err = tx.QueryRowContext(ctx, `select id from sessions where id=?`, sessionID).Scan(&owner); err != nil {
		return TUITextDraft{}, err
	}
	var raw []byte
	err = tx.QueryRowContext(ctx, `select value from kv_store where namespace=? and key=?`, tuiTextDraftNamespace, sessionID).Scan(&raw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return TUITextDraft{}, err
	}
	state, err := decodeTUITextDraft(raw)
	if err != nil {
		return TUITextDraft{}, err
	}
	before, err := json.Marshal(state)
	if err != nil {
		return TUITextDraft{}, err
	}
	if err = change(tx, &state); err != nil {
		return TUITextDraft{}, err
	}
	after, err := json.Marshal(state)
	if err != nil {
		return TUITextDraft{}, err
	}
	if string(before) != string(after) {
		if state.Revision == math.MaxInt64 {
			return TUITextDraft{}, errors.New("terminal draft revision exhausted")
		}
		state.Revision++
		after, err = json.Marshal(state)
		if err != nil {
			return TUITextDraft{}, err
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		// Retain empty revision records: deleting would let a stale revision-zero
		// writer silently replace a newer draft after a clear/claim cycle (ABA).
		if _, err = tx.ExecContext(ctx, `insert into kv_store(namespace,key,value,created_at,updated_at) values(?,?,?,?,?) on conflict(namespace,key) do update set value=excluded.value,updated_at=excluded.updated_at`, tuiTextDraftNamespace, sessionID, after, now, now); err != nil {
			return TUITextDraft{}, err
		}
	}
	return state, nil
}

func (s *Store) SaveTUITextDraft(ctx context.Context, sessionID string, expected int64, snapshot TUITextSnapshot) (TUITextDraft, error) {
	if !validTUITextSnapshot(snapshot) {
		return TUITextDraft{}, errors.New("invalid terminal draft text/cursor")
	}
	return s.updateTUITextDraft(ctx, sessionID, func(_ *sql.Tx, state *TUITextDraft) error {
		if state.Revision != expected {
			return ErrTUIDraftConflict
		}
		if state.Revision >= math.MaxInt64-2 {
			return errors.New("terminal draft revision exhausted")
		}
		state.TUITextSnapshot = snapshot
		return nil
	})
}

// Claim must commit before the frontend clears its editor or calls submission.
// Its token belongs in native admission metadata as tui_text_claim.
func (s *Store) ClaimTUITextDraft(ctx context.Context, sessionID string, expected int64) (TUITextDraft, error) {
	return s.updateTUITextDraft(ctx, sessionID, func(_ *sql.Tx, state *TUITextDraft) error { return claimTUIText(state, expected) })
}
func claimTUIText(state *TUITextDraft, expected int64) error {
	if state.Revision != expected {
		return ErrTUIDraftConflict
	}
	if state.Claim != nil {
		return ErrTUIDraftHeld
	}
	if state.Text == "" || state.Revision >= math.MaxInt64-2 {
		return errors.New("nonempty claimable draft with revision capacity required")
	}
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return err
	}
	state.Claim = &TUITextClaim{TUITextSnapshot: state.TUITextSnapshot, Token: "tui-text-" + hex.EncodeToString(token[:]), Revision: state.Revision + 1}
	state.TUITextSnapshot = TUITextSnapshot{}
	return nil
}

// A turn INSERT may still be rolled back during subturn setup. Only its
// post-rollback submitted event confirms it. Steering commits atomically and
// retained messages cover consumed steering. No prompt-text matching fallback.
func tuiTextAdmitted(ctx context.Context, tx *sql.Tx, sessionID, token string, submittedReturned bool) (bool, error) {
	var admitted bool
	err := tx.QueryRowContext(ctx, `select exists(
 select 1 from turns t where session_id=? and json_extract(metadata_json,'$.tui_text_claim')=? and (? or exists(select 1 from turn_events e where e.turn_id=t.id and e.event_type='turn.submitted'))
 union all select 1 from steering_queue where session_id=? and json_extract(payload_json,'$.tui_text_claim')=?
 union all select 1 from messages where session_id=? and json_extract(payload_json,'$.tui_text_claim')=?)`, sessionID, token, submittedReturned, sessionID, token, sessionID, token).Scan(&admitted)
	return admitted, err
}

// Reconcile never restores an absent receipt. A missing audit or crash before
// admission leaves the snapshot held; this is not permission to replay it.
func (s *Store) ReconcileTUITextDraft(ctx context.Context, sessionID, token string) (TUITextDraft, error) {
	return s.updateTUITextDraft(ctx, sessionID, func(tx *sql.Tx, state *TUITextDraft) error {
		if state.Claim == nil || state.Claim.Token != token {
			return ErrTUIDraftConflict
		}
		if state.Claim.Media {
			return ErrTUIDraftHeld
		}
		admitted, err := tuiTextAdmitted(ctx, tx, sessionID, token, false)
		if err != nil {
			return err
		}
		if admitted {
			state.Claim = nil
		}
		return nil
	})
}

// Finish is reserved for the live caller after synchronous SubmitPrompt returns.
// Even success requires a stored receipt. A rejected claim restores only when
// no subsequent edit occurred; otherwise retain it separately for explicit use.
func (s *Store) FinishTUITextDraft(ctx context.Context, sessionID, token string, rejected bool) (TUITextDraft, error) {
	return s.updateTUITextDraft(ctx, sessionID, func(tx *sql.Tx, state *TUITextDraft) error {
		if state.Claim == nil || state.Claim.Token != token {
			return ErrTUIDraftConflict
		}
		if state.Claim.Media {
			return ErrTUIDraftHeld
		}
		admitted, err := tuiTextAdmitted(ctx, tx, sessionID, token, !rejected)
		if err != nil {
			return err
		}
		if admitted {
			state.Claim = nil
			return nil
		}
		if !rejected {
			return nil
		}
		// An error plus a turn without its post-rollback audit is ambiguous.
		// Preserve the held snapshot, neither restoring nor dropping it.
		possible, err := tuiTextAdmitted(ctx, tx, sessionID, token, true)
		if err != nil {
			return err
		}
		if possible {
			return nil
		}
		if state.Revision == state.Claim.Revision {
			state.TUITextSnapshot = state.Claim.TUITextSnapshot
			state.Claim = nil
		} else {
			state.Claim.Rejected = true
		}
		return nil
	})
}

// Explicit restore is permitted only for a proven rejected claim, a blank
// current editor and an exact revision. It cannot restore an unknown outcome.
func (s *Store) RestoreRejectedTUITextDraft(ctx context.Context, sessionID, token string, expected int64) (TUITextDraft, error) {
	return s.updateTUITextDraft(ctx, sessionID, func(_ *sql.Tx, state *TUITextDraft) error {
		if state.Revision != expected || state.Text != "" || state.Claim == nil || state.Claim.Token != token {
			return ErrTUIDraftConflict
		}
		if !state.Claim.Rejected || state.Claim.Media {
			return ErrTUIDraftHeld
		}
		state.TUITextSnapshot = state.Claim.TUITextSnapshot
		state.Claim = nil
		return nil
	})
}

// Discard explicitly abandons only a proven rejected snapshot. Newer editable
// text remains intact, and unknown outcomes cannot be discarded by this API.
func (s *Store) DiscardRejectedTUITextDraft(ctx context.Context, sessionID, token string, expected int64) (TUITextDraft, error) {
	return s.updateTUITextDraft(ctx, sessionID, func(_ *sql.Tx, state *TUITextDraft) error {
		if state.Revision != expected || state.Claim == nil || state.Claim.Token != token {
			return ErrTUIDraftConflict
		}
		if !state.Claim.Rejected || state.Claim.Media {
			return ErrTUIDraftHeld
		}
		state.Claim = nil
		return nil
	})
}
