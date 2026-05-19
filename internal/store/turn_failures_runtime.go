package store

import (
	"context"
	"strings"
)

func CoordinationContext(ctx context.Context, fallback context.Context) context.Context {
	if ctx != nil && ctx.Err() == nil {
		return ctx
	}
	if fallback != nil && fallback.Err() == nil {
		return fallback
	}
	return nil
}

func (s *Store) MarkTurnFailureWithFallbackErr(ctx, fallback context.Context, turnID, sessionID, failureKind, holdState, summary string) error {
	failureKind = strings.TrimSpace(failureKind)
	if failureKind == "" {
		failureKind = "unknown"
	}
	holdState = strings.TrimSpace(strings.ToLower(holdState))
	if holdState == "" {
		holdState = "none"
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		summary = failureKind
	}
	opCtx := CoordinationContext(ctx, fallback)
	if opCtx == nil {
		return nil
	}
	if err := s.UpsertTurnFailure(opCtx, turnID, sessionID, failureKind, holdState, summary); err != nil {
		return err
	}
	if err := s.AppendTurnEvent(opCtx, turnID, sessionID, "turn.failure_marked", map[string]any{
		"phase":        "recovery",
		"checkpoint":   true,
		"failure_kind": failureKind,
		"hold_state":   holdState,
		"summary":      summary,
	}); err != nil {
		return err
	}
	return nil
}
