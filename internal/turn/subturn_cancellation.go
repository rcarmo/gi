package turn

import (
	"context"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

type subTurnCancellationDirective struct {
	enabled        bool
	cancelCritical bool
	reason         string
}

func directiveForParentTerminalStatus(parentStatus, failureKind string) subTurnCancellationDirective {
	status := strings.ToLower(strings.TrimSpace(parentStatus))
	failureKind = strings.ToLower(strings.TrimSpace(failureKind))
	switch {
	case isTimeoutFailureKind(failureKind):
		return subTurnCancellationDirective{enabled: true, cancelCritical: true, reason: "parent_timeout"}
	case status == "cancelled" || status == "aborted":
		return subTurnCancellationDirective{enabled: true, cancelCritical: true, reason: "parent_hard_abort"}
	case status == "completed":
		return subTurnCancellationDirective{enabled: true, cancelCritical: false, reason: "parent_finished_gracefully"}
	default:
		return subTurnCancellationDirective{}
	}
}

func isTimeoutFailureKind(failureKind string) bool {
	failureKind = strings.ToLower(strings.TrimSpace(failureKind))
	if failureKind == "" {
		return false
	}
	return strings.Contains(failureKind, "timeout") || strings.Contains(failureKind, "deadline")
}

func isCriticalSubTurn(sub store.SubTurn) bool {
	return boolValue(sub.Metadata["subturn_critical"]) || boolValue(sub.Metadata["critical"])
}

func isCancellableSubTurnStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "running", "cancelling":
		return true
	default:
		return false
	}
}

func (r *sessionRunner) propagateChildSubTurnCancellation(ctx context.Context, parentTurnID, parentStatus, failureKind string) {
	parentTurnID = strings.TrimSpace(parentTurnID)
	if parentTurnID == "" {
		return
	}
	directive := directiveForParentTerminalStatus(parentStatus, failureKind)
	if !directive.enabled {
		return
	}
	visited := map[string]bool{}
	var walk func(string)
	walk = func(turnID string) {
		subs, err := r.store.ListSubTurnsByParent(ctx, turnID)
		if err != nil {
			warnStore("list child subturns", err)
			return
		}
		for _, sub := range subs {
			if visited[sub.ChildTurnID] {
				continue
			}
			visited[sub.ChildTurnID] = true
			walk(sub.ChildTurnID)
			if !isCancellableSubTurnStatus(sub.Status) {
				continue
			}
			if isCriticalSubTurn(sub) && !directive.cancelCritical {
				continue
			}
			warnStore("update child subturn cancellation metadata", r.store.UpdateSubTurnMetadataByChild(ctx, sub.ChildTurnID, map[string]any{
				"cancel_requested_by_parent":     true,
				"cancel_requested_at":            time.Now().UTC().Format(time.RFC3339Nano),
				"cancel_requested_parent_turn":   turnID,
				"cancel_requested_parent_status": parentStatus,
				"cancel_requested_failure_kind":  failureKind,
				"cancel_reason":                  directive.reason,
			}))
			if err := r.engine.CancelTurn(ctx, sub.ChildSessionID, sub.ChildTurnID); err != nil {
				warnStore("cancel child subturn", err)
				continue
			}
			r.engine.broadcast(sub.ParentSessionID, map[string]any{
				"type":                "subturn_cancel_requested",
				"chat_jid":            "gi:" + sub.ParentSessionID,
				"parent_turn_id":      sub.ParentTurnID,
				"parent_session":      sub.ParentSessionID,
				"child_turn_id":       sub.ChildTurnID,
				"child_session":       sub.ChildSessionID,
				"reason":              directive.reason,
				"parent_status":       parentStatus,
				"parent_failure_kind": failureKind,
			})
		}
	}
	walk(parentTurnID)
}
