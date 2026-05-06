package turn

import (
	"context"
	"fmt"
)

func (e *Engine) HoldTurnFailure(ctx context.Context, turnID, holdState, summary string) error {
	turnRec, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	if err := e.store.HoldTurnFailure(ctx, turnID, holdState, summary); err != nil {
		return err
	}
	return e.store.AppendTurnEvent(ctx, turnID, turnRec.SessionID, "turn.failure_held", map[string]any{
		"phase":      "recovery",
		"checkpoint": true,
		"hold_state": holdState,
		"summary":    summary,
	})
}

func (e *Engine) RetryHeldTurn(ctx context.Context, turnID, summary string) (*SubmitResult, error) {
	turnRec, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return nil, err
	}
	failureRec, err := e.store.GetTurnFailure(ctx, turnID)
	if err != nil {
		return nil, err
	}
	if failureRec.HoldState == "none" {
		return nil, fmt.Errorf("retry held turn: turn %s is not currently held", turnID)
	}
	metadata := map[string]any{}
	for k, v := range turnRec.Metadata {
		metadata[k] = v
	}
	metadata["retry_of_turn_id"] = turnID
	metadata["retry_failure_kind"] = failureRec.FailureKind
	metadata["retry_hold_state"] = failureRec.HoldState
	metadata["failure_resolution"] = "retry"
	result, err := e.SubmitPrompt(ctx, RunInput{
		SessionID:    turnRec.SessionID,
		Prompt:       turnRec.Prompt,
		Intent:       stringValue(turnRec.Metadata["intent"], "prompt"),
		Model:        stringValue(turnRec.Metadata["model"], ""),
		ParentTurnID: stringValue(turnRec.Metadata["parent_turn_id"], ""),
		Metadata:     metadata,
	})
	if err != nil {
		return nil, err
	}
	if err := e.store.ResolveTurnFailure(ctx, turnID, "retried", summary, result.TurnID); err != nil {
		return nil, err
	}
	_ = e.store.AppendTurnEvent(ctx, turnID, turnRec.SessionID, "turn.failure_resolved", map[string]any{
		"phase":              "recovery",
		"checkpoint":         true,
		"resolution_state":   "retried",
		"resolution_summary": summary,
		"resolved_turn_id":   result.TurnID,
	})
	return result, nil
}

func (e *Engine) SkipHeldTurn(ctx context.Context, turnID, summary string) error {
	turnRec, err := e.store.GetTurn(ctx, turnID)
	if err != nil {
		return err
	}
	failureRec, err := e.store.GetTurnFailure(ctx, turnID)
	if err != nil {
		return err
	}
	if failureRec.HoldState == "none" {
		return fmt.Errorf("skip held turn: turn %s is not currently held", turnID)
	}
	if err := e.store.ResolveTurnFailure(ctx, turnID, "skipped", summary, ""); err != nil {
		return err
	}
	return e.store.AppendTurnEvent(ctx, turnID, turnRec.SessionID, "turn.failure_resolved", map[string]any{
		"phase":              "recovery",
		"checkpoint":         true,
		"resolution_state":   "skipped",
		"resolution_summary": summary,
	})
}
