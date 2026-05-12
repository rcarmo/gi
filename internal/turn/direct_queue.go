package turn

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

func directEnvelopeFromInput(in DirectInput) map[string]any {
	envelope := map[string]any{
		"kind":            in.Kind,
		"session_id":      in.SessionID,
		"session_key":     in.SessionKey,
		"target_agent_id": in.TargetAgentID,
		"prompt":          in.Prompt,
		"intent":          in.Intent,
		"model":           in.Model,
		"parent_turn_id":  in.ParentTurnID,
		"metadata":        in.Metadata,
		"origin": map[string]any{
			"source_kind": in.Origin.SourceKind,
			"source_id":   in.Origin.SourceID,
			"role":        in.Origin.Role,
			"label":       in.Origin.Label,
		},
	}
	return envelope
}

func directInputFromEnvelope(envelope map[string]any) DirectInput {
	in := DirectInput{
		Kind:          stringValue(envelope["kind"], ""),
		SessionID:     stringValue(envelope["session_id"], ""),
		SessionKey:    stringValue(envelope["session_key"], ""),
		TargetAgentID: stringValue(envelope["target_agent_id"], ""),
		Prompt:        stringValue(envelope["prompt"], ""),
		Intent:        stringValue(envelope["intent"], ""),
		Model:         stringValue(envelope["model"], ""),
		ParentTurnID:  stringValue(envelope["parent_turn_id"], ""),
		Metadata:      map[string]any{},
	}
	if metadata, ok := envelope["metadata"].(map[string]any); ok && metadata != nil {
		in.Metadata = metadata
	}
	if origin, ok := envelope["origin"].(map[string]any); ok && origin != nil {
		in.Origin = DirectOrigin{
			SourceKind: stringValue(origin["source_kind"], ""),
			SourceID:   stringValue(origin["source_id"], ""),
			Role:       stringValue(origin["role"], ""),
			Label:      stringValue(origin["label"], ""),
		}
	}
	return in
}

const (
	inboundWorkMaxAttempts = 3
	inboundWorkRetryDelay  = 2 * time.Second
)

func (e *Engine) EnqueueDirectInbound(ctx context.Context, in DirectInput) (*store.InboundWorkItem, error) {
	if e.store == nil {
		return nil, fmt.Errorf("direct inbound queue requires store")
	}
	sourceKind := normalizeDirectSourceKind(in.Origin.SourceKind)
	envelope := directEnvelopeFromInput(in)
	item, err := e.store.EnqueueInboundWork(ctx, sourceKind, strings.TrimSpace(in.SessionID), strings.TrimSpace(in.SessionKey), envelope)
	if err != nil {
		return nil, err
	}
	e.PublishRuntimeInboundWorkEvent("inbound_work_enqueued", item, nil)
	return item, nil
}

func (e *Engine) ProcessNextInboundWork(ctx context.Context, claimedBy string) (*store.InboundWorkItem, *SubmitResult, error) {
	if e.store == nil {
		return nil, nil, fmt.Errorf("inbound work processing requires store")
	}
	item, err := e.store.ClaimNextInboundWork(ctx, claimedBy)
	if err != nil {
		return nil, nil, err
	}
	in := directInputFromEnvelope(item.Envelope)
	if strings.TrimSpace(in.Origin.SourceKind) == "" {
		in.Origin.SourceKind = item.SourceKind
	}
	if strings.TrimSpace(in.SessionID) == "" {
		in.SessionID = item.SessionID
	}
	if strings.TrimSpace(in.SessionKey) == "" {
		in.SessionKey = item.ExplicitSessionKey
	}
	result, processErr := e.ProcessDirect(ctx, in)
	if processErr != nil {
		attemptCount := item.AttemptCount + 1
		var updateErr error
		if attemptCount >= inboundWorkMaxAttempts {
			updateErr = e.store.RecordInboundWorkFailure(ctx, item.ID, attemptCount, processErr.Error())
		} else {
			updateErr = e.store.RecordInboundWorkRetry(ctx, item.ID, attemptCount, processErr.Error(), inboundWorkRetryDelay*time.Duration(attemptCount))
		}
		statusEvent := map[bool]string{true: "inbound_work_failed", false: "inbound_work_retry_scheduled"}[attemptCount >= inboundWorkMaxAttempts]
		if updateErr != nil {
			return item, result, updateErr
		}
		updated, getErr := e.store.GetInboundWork(ctx, item.ID)
		if getErr == nil {
			item = updated
		}
		e.PublishRuntimeInboundWorkEvent(statusEvent, item, map[string]any{"error": processErr.Error()})
		return item, result, processErr
	}
	if err := e.store.UpdateInboundWorkStatus(ctx, item.ID, "completed"); err != nil {
		return item, result, err
	}
	updated, getErr := e.store.GetInboundWork(ctx, item.ID)
	if getErr == nil {
		item = updated
	}
	e.PublishRuntimeInboundWorkEvent("inbound_work_completed", item, nil)
	return item, result, nil
}

func (e *Engine) ProcessNextInboundWorkIfQueued(ctx context.Context, claimedBy string) (*store.InboundWorkItem, *SubmitResult, bool, error) {
	item, result, err := e.ProcessNextInboundWork(ctx, claimedBy)
	if err == sql.ErrNoRows {
		return nil, nil, false, nil
	}
	if err != nil {
		return item, result, true, err
	}
	return item, result, true, nil
}

func (e *Engine) ProcessQueuedInboundWork(ctx context.Context, claimedBy string, limit int) ([]*store.InboundWorkItem, []*SubmitResult, error) {
	if limit <= 0 {
		limit = 1
	}
	items := make([]*store.InboundWorkItem, 0, limit)
	results := make([]*SubmitResult, 0, limit)
	for i := 0; i < limit; i++ {
		item, result, ok, err := e.ProcessNextInboundWorkIfQueued(ctx, claimedBy)
		if err != nil {
			return items, results, err
		}
		if !ok {
			break
		}
		items = append(items, item)
		results = append(results, result)
	}
	return items, results, nil
}
