package turn

import (
	"context"
	"fmt"

	"github.com/rcarmo/gi/internal/store"
)

// settleCancelledToolBatch closes every provider-visible call without entering
// preparation, hooks or executors. Each result is atomic/idempotent; a recovered
// terminal host reads the journal again and skips already settled slots.
func (r *sessionRunner) settleCancelledToolBatch(ctx context.Context, s *store.Store, sessionID, turnID string) error {
	state, err := s.LoadCancelledToolBatchCheckpoint(ctx, sessionID, turnID)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("cancelled tool checkpoint is not proven")
	}
	generation := state.Checkpoint.Generation
	for index, call := range state.Calls {
		if call.Done {
			continue
		}
		slot := store.WithToolBatchSlot(ctx, turnID, state.Checkpoint.ProviderOccurrenceID, index, generation.Iteration, generation.LastToolFailureSignature, generation.RepeatedToolFailures, call.Original)
		occurrence := call.OccurrenceID
		physical := occurrence != "" && !call.ReplayPending
		name, id := call.Original.Name, call.Original.ID
		text := "Tool call cancelled before executor admission. Preparation and execution were not repeated."
		if physical {
			name, id = call.Prepared.Name, call.Prepared.ID
			text = "Tool call cancelled after interruption. Verify its effects before repeating; its executor and result policy were not repeated."
			if call.CommittedOutput != "" {
				text = call.CommittedOutput + "\n\n<harness>\n[error] " + text + "\n</harness>"
			}
		}
		event := map[string]any{"phase": "tool", "checkpoint": true, "reason": "checkpoint_cancel", "tool": name, "tool_call_id": id, "replay_pending": false}
		if !physical {
			occurrence = store.NowID("tool")
			event["unadmitted"], event["previous_occurrence_id"] = true, call.OccurrenceID
		}
		event["occurrence_id"] = occurrence
		payload := map[string]any{"kind": "tool_result", "turn_id": turnID, "tool_call_id": call.Original.ID, "tool_name": call.Original.Name, "occurrence_id": occurrence, "is_error": true, "cancelled": true, "interrupted": true, "outcome_unknown": physical, "usage_unknown": false}
		if physical {
			payload["usage"], payload["usage_unknown"] = call.ReceivedUsage, !call.CallbackReturned || call.ReceivedUsageUnknown
			if call.OutputTruncated {
				payload["output_truncated"] = true
			}
		}
		if err = s.SettleToolResult(slot, turnID, sessionID, "tool.interrupted", event, store.NowID("msg"), text, payload); err != nil {
			return err
		}
	}
	return nil
}
