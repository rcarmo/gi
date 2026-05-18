package turn

import (
	"context"
	"github.com/rcarmo/gi/internal/logutil"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func markTurnFailure(ctx context.Context, s *store.Store, turnID, sessionID, failureKind, summary string) {
	logutil.WarnIfErr("turn failure mark", markTurnFailureWithFallbackErr(ctx, nil, s, turnID, sessionID, failureKind, "none", summary))
}

func markTurnFailureWithFallback(ctx, fallback context.Context, s *store.Store, turnID, sessionID, failureKind, summary string) {
	logutil.WarnIfErr("turn failure mark", markTurnFailureWithFallbackErr(ctx, fallback, s, turnID, sessionID, failureKind, "none", summary))
}

func markTurnFailureWithHold(ctx context.Context, s *store.Store, turnID, sessionID, failureKind, holdState, summary string) {
	logutil.WarnIfErr("turn failure mark", markTurnFailureWithFallbackErr(ctx, nil, s, turnID, sessionID, failureKind, holdState, summary))
}

func markTurnFailureWithHoldAndFallback(ctx, fallback context.Context, s *store.Store, turnID, sessionID, failureKind, holdState, summary string) {
	logutil.WarnIfErr("turn failure mark", markTurnFailureWithFallbackErr(ctx, fallback, s, turnID, sessionID, failureKind, holdState, summary))
}

func markTurnFailureWithFallbackErr(ctx, fallback context.Context, s *store.Store, turnID, sessionID, failureKind, holdState, summary string) error {
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
	opCtx := coordinationContext(ctx, fallback)
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
