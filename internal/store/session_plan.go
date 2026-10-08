package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rcarmo/gi/internal/plan"
)

var ErrPlanRevisionRequired = errors.New("plan revision required")
var ErrPlanRevisionConflict = errors.New("plan revision conflict")

type SessionPlan struct {
	Revision    string      `json:"revision"`
	ChatJID     string      `json:"chat_jid"`
	Markdown    string      `json:"markdown"`
	UpdatedAt   *string     `json:"updated_at"`
	Explanation *string     `json:"explanation"`
	Plan        []plan.Item `json:"plan"`
}

func planDetails(session, markdown string, updated *string) SessionPlan {
	parsed := plan.Parse(markdown)
	stamp := ""
	if updated != nil {
		stamp = *updated
	}
	sum := sha256.Sum256([]byte(session + "\x00" + markdown + "\x00" + stamp))
	return SessionPlan{Revision: "plan-v1-" + hex.EncodeToString(sum[:]), ChatJID: "gi:" + session, Markdown: markdown, UpdatedAt: updated, Explanation: parsed.Explanation, Plan: parsed.Plan}
}
func (s *Store) SessionPlan(ctx context.Context, session string) (SessionPlan, error) {
	if _, err := s.GetSession(ctx, session); err != nil {
		return SessionPlan{}, err
	}
	var raw []byte
	err := s.db.QueryRowContext(ctx, "select value from kv_store where namespace='session_plan' and key=?", session).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return planDetails(session, plan.DefaultMarkdown, nil), nil
	}
	if err != nil {
		return SessionPlan{}, err
	}
	var saved SessionPlan
	if err = json.Unmarshal(raw, &saved); err != nil {
		return SessionPlan{}, err
	}
	return planDetails(session, saved.Markdown, saved.UpdatedAt), nil
}
func (s *Store) MutateSessionPlan(ctx context.Context, session string, mutation plan.Mutation) (SessionPlan, error) {
	return s.mutateSessionPlan(ctx, session, mutation, nil)
}
func (s *Store) MutateSessionPlanConditional(ctx context.Context, session string, mutation plan.Mutation, expected string) (SessionPlan, error) {
	if expected == "" {
		if _, err := s.GetSession(ctx, session); err != nil {
			return SessionPlan{}, err
		}
		return SessionPlan{}, ErrPlanRevisionRequired
	}
	return s.mutateSessionPlan(ctx, session, mutation, &expected)
}
func (s *Store) mutateSessionPlan(ctx context.Context, session string, mutation plan.Mutation, expected *string) (SessionPlan, error) {
	var out SessionPlan
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	// Acquire SQLite's writer lock before reading the revision, serialising
	// browser and agent writes without a read-to-write transaction upgrade.
	if _, err = tx.ExecContext(ctx, `update sessions set title=title where id=?`, session); err != nil {
		return out, err
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, "select exists(select 1 from sessions where id=?)", session).Scan(&exists); err != nil {
		return out, err
	}
	if !exists {
		return out, sql.ErrNoRows
	}
	markdown := plan.DefaultMarkdown
	var loadedUpdated *string
	var raw []byte
	err = tx.QueryRowContext(ctx, "select value from kv_store where namespace='session_plan' and key=?", session).Scan(&raw)
	if err == nil {
		var current SessionPlan
		if err = json.Unmarshal(raw, &current); err != nil {
			return out, err
		}
		markdown = current.Markdown
		loadedUpdated = current.UpdatedAt
	} else if !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}
	current := planDetails(session, markdown, loadedUpdated)
	if expected != nil && *expected != current.Revision {
		return current, ErrPlanRevisionConflict
	}
	markdown, err = plan.Apply(markdown, mutation)
	if err != nil {
		return out, fmt.Errorf("%w: %v", plan.ErrInvalid, err)
	}
	updated := time.Now().UTC().Format(time.RFC3339Nano)
	out = planDetails(session, markdown, &updated)
	raw, err = json.Marshal(out)
	if err != nil {
		return SessionPlan{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into kv_store(namespace,key,value,created_at,updated_at) values('session_plan',?,?,?,?)
 on conflict(namespace,key) do update set value=excluded.value,updated_at=excluded.updated_at`, session, raw, updated, updated); err != nil {
		return SessionPlan{}, err
	}
	if err = tx.Commit(); err != nil {
		return SessionPlan{}, err
	}
	return out, nil
}
