package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func failureOpContext(ctx, fallback context.Context) context.Context {
	if ctx != nil && ctx.Err() == nil {
		return ctx
	}
	if fallback != nil && fallback.Err() == nil {
		return fallback
	}
	return nil
}

func markTurnFailure(ctx context.Context, s *store.Store, turnID, sessionID, failureKind, summary string) {
	markTurnFailureWithFallback(ctx, nil, s, turnID, sessionID, failureKind, summary)
}

func markTurnFailureWithFallback(ctx, fallback context.Context, s *store.Store, turnID, sessionID, failureKind, summary string) {
	markTurnFailureWithHoldAndFallback(ctx, fallback, s, turnID, sessionID, failureKind, "none", summary)
}

func markTurnFailureWithHold(ctx context.Context, s *store.Store, turnID, sessionID, failureKind, holdState, summary string) {
	markTurnFailureWithHoldAndFallback(ctx, nil, s, turnID, sessionID, failureKind, holdState, summary)
}

func markTurnFailureWithHoldAndFallback(ctx, fallback context.Context, s *store.Store, turnID, sessionID, failureKind, holdState, summary string) {
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
	opCtx := failureOpContext(ctx, fallback)
	if opCtx == nil {
		return
	}
	warnStore("turn failure upsert", s.UpsertTurnFailure(opCtx, turnID, sessionID, failureKind, holdState, summary))
	warnStore("turn failure event append", s.AppendTurnEvent(opCtx, turnID, sessionID, "turn.failure_marked", map[string]any{
		"phase":        "recovery",
		"checkpoint":   true,
		"failure_kind": failureKind,
		"hold_state":   holdState,
		"summary":      summary,
	}))
}
