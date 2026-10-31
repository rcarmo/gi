package turn

import goai "github.com/rcarmo/go-ai"

// reduceContextContributions separates retained receipts from model inputs.
// Boundaries map a reduced prefix back to its source prefix for compaction:
// a call batch (including synthesized results) is always one indivisible unit.
func reduceContextContributions(source []goai.Message) ([]goai.Message, map[int]int) {
	out := make([]goai.Message, 0, len(source))
	boundaries := map[int]int{0: 0}
	for i := 0; i < len(source); {
		message := source[i]
		i++
		if message.Role == goai.RoleToolResult || (message.Role == goai.RoleAssistant && (message.StopReason == goai.StopReasonError || message.StopReason == goai.StopReasonAborted || message.StopReason == goai.StopReasonDeferred)) {
			boundaries[len(out)] = i
			continue
		}
		out = append(out, message)
		calls := goai.GetToolCalls(&message)
		if message.Role == goai.RoleAssistant && len(calls) > 0 {
			// A later batch may reuse the ID. Only this contiguous result region can
			// satisfy these calls; unmatched/duplicate receipts remain raw/UI-only.
			results := map[string]goai.Message{}
			for i < len(source) && source[i].Role == goai.RoleToolResult {
				result := source[i]
				i++
				if _, exists := results[result.ToolCallID]; !exists {
					results[result.ToolCallID] = result
				}
			}
			for _, call := range calls {
				result, ok := results[call.ID]
				if !ok {
					result = goai.Message{Role: goai.RoleToolResult, ToolCallID: call.ID, ToolName: call.Name, IsError: true, Content: []goai.ContentBlock{{Type: "text", Text: "Tool call did not produce a durable result."}}, Details: map[string]any{"reason": "missing_result"}}
				}
				out = append(out, result)
			}
		}
		boundaries[len(out)] = i
	}
	return out, boundaries
}
