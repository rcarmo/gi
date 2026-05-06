package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func markTurnFailure(s *store.Store, turnID, sessionID, failureKind, summary string) {
	failureKind = strings.TrimSpace(failureKind)
	if failureKind == "" {
		failureKind = "unknown"
	}
	summary = strings.TrimSpace(summary)
	if summary == "" {
		summary = failureKind
	}
	_ = s.UpsertTurnFailure(context.Background(), turnID, sessionID, failureKind, "none", summary)
	_ = s.AppendTurnEvent(context.Background(), turnID, sessionID, "turn.failure_marked", map[string]any{
		"phase":        "recovery",
		"checkpoint":   true,
		"failure_kind": failureKind,
		"hold_state":   "none",
		"summary":      summary,
	})
}
