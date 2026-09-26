package turn

import (
	"context"
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/logutil"
	"github.com/rcarmo/gi/internal/store"
)

// A retry carries prompt inputs and tool restrictions, not prior delivery or
// continuation bookkeeping. SubmitPrompt revalidates media and parent tools.
func retryMetadata(original map[string]any) map[string]any {
	metadata := make(map[string]any, len(original))
	for k, v := range original {
		if strings.HasPrefix(k, "retry_") || strings.HasPrefix(k, "route_") || strings.HasPrefix(k, "ingress_") {
			continue
		}
		switch k {
		case "intent", "initial_steering", "continue", "steering_mode", "failure_resolution",
			"source_session_id", "source_agent_id", "target_session_id", "target_agent_id",
			"routing_enabled", "routed_from_prompt", "tui_media_claim":
			continue
		}
		metadata[k] = v
	}
	return metadata
}

// Recovery notifications invalidate views; consumers must read the stored
// resolution. They can repeat when two callers observe the same admission.
func (e *Engine) publishRetryResolution(ctx context.Context, original *store.Turn, admitted string) {
	failure, err := e.store.GetTurnFailure(ctx, original.ID)
	if err != nil {
		logutil.WarnIfErr("read retry resolution notification", err)
		return
	}
	if failure.ResolutionState != "retried" || failure.ResolvedTurnID != admitted {
		return
	}
	phase := original.Phase
	if phase == "held_for_retry_or_skip" {
		phase = store.RuntimeTurnPhaseForStatus(original.Status)
	}
	payload := map[string]any{
		"phase": "recovery", "checkpoint": true, "reason": "failure_resolved",
		"resolution_state": "retried", "resolution_summary": failure.ResolutionSummary, "resolved_turn_id": admitted,
	}
	logutil.WarnIfErr("append turn.failure_resolved event", e.store.AppendTurnEvent(ctx, original.ID, original.SessionID, "turn.failure_resolved", payload))
	e.PublishRuntimeTurnEvent("turn_failure_resolved", original.SessionID, original.ID, "", original.Status, phase, payload)
}

func (e *Engine) recoveredRetryResult(ctx context.Context, sessionID, id string) (*SubmitResult, error) {
	rec, err := e.store.GetTurn(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.SessionID != sessionID {
		return nil, fmt.Errorf("recovered retry belongs to another session")
	}
	return &SubmitResult{TurnID: rec.ID, SessionID: sessionID, Status: rec.Status, Queued: rec.Status == "queued"}, nil
}
