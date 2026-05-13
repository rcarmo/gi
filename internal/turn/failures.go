package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func markTurnFailure(ctx context.Context, s *store.Store, turnID, sessionID, failureKind, summary string) {
	markTurnFailureWithHold(ctx, s, turnID, sessionID, failureKind, "none", summary)
}

func markTurnFailureWithHold(ctx context.Context, s *store.Store, turnID, sessionID, failureKind, holdState, summary string) {
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
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = context.Background()
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
