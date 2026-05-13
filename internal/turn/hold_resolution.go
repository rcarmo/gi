package turn

import (
	"context"
	"fmt"
)

func (e *Engine) HoldTurnFailure(ctx context.Context, turnID, holdState, summary string) error {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return err
	}
	if err := e.store.HoldTurnFailure(opCtx, turnID, holdState, summary); err != nil {
		return err
	}
	payload := map[string]any{
		"phase":      "recovery",
		"checkpoint": true,
		"reason":     "failure_held",
		"hold_state": holdState,
		"summary":    summary,
	}
	if err := e.store.AppendTurnEvent(opCtx, turnID, turnRec.SessionID, "turn.failure_held", payload); err != nil {
		return err
	}
	e.PublishRuntimeTurnEvent("turn_failure_held", turnRec.SessionID, turnID, "", turnRec.Status, turnRec.Phase, payload)
	return nil
}

func (e *Engine) RetryHeldTurn(ctx context.Context, turnID, summary string) (*SubmitResult, error) {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return nil, err
	}
	failureRec, err := e.store.GetTurnFailure(opCtx, turnID)
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
	result, err := e.SubmitPrompt(opCtx, RunInput{
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
	if err := e.store.ResolveTurnFailure(opCtx, turnID, "retried", summary, result.TurnID); err != nil {
		return nil, err
	}
	payload := map[string]any{
		"phase":              "recovery",
		"checkpoint":         true,
		"reason":             "failure_resolved",
		"resolution_state":   "retried",
		"resolution_summary": summary,
		"resolved_turn_id":   result.TurnID,
	}
	warnStore("append turn.failure_resolved event", e.store.AppendTurnEvent(opCtx, turnID, turnRec.SessionID, "turn.failure_resolved", payload))
	e.PublishRuntimeTurnEvent("turn_failure_resolved", turnRec.SessionID, turnID, "", turnRec.Status, turnRec.Phase, payload)
	return result, nil
}

func (e *Engine) SkipHeldTurn(ctx context.Context, turnID, summary string) error {
	opCtx := ctx
	if opCtx == nil || opCtx.Err() != nil {
		opCtx = e.backgroundContext()
	}
	turnRec, err := e.store.GetTurn(opCtx, turnID)
	if err != nil {
		return err
	}
	failureRec, err := e.store.GetTurnFailure(opCtx, turnID)
	if err != nil {
		return err
	}
	if failureRec.HoldState == "none" {
		return fmt.Errorf("skip held turn: turn %s is not currently held", turnID)
	}
	if err := e.store.ResolveTurnFailure(opCtx, turnID, "skipped", summary, ""); err != nil {
		return err
	}
	payload := map[string]any{
		"phase":              "recovery",
		"checkpoint":         true,
		"reason":             "failure_resolved",
		"resolution_state":   "skipped",
		"resolution_summary": summary,
	}
	if err := e.store.AppendTurnEvent(opCtx, turnID, turnRec.SessionID, "turn.failure_resolved", payload); err != nil {
		return err
	}
	e.PublishRuntimeTurnEvent("turn_failure_resolved", turnRec.SessionID, turnID, "", turnRec.Status, turnRec.Phase, payload)
	return nil
}
