package inference

import (
	"encoding/json"
	"fmt"
	"math"

	goai "github.com/rcarmo/go-ai"
)

// ValidateAssistantResult runs before receipt adoption or tool dispatch, also
// for custom inference adapters. Deferred transport is explicitly unsupported;
// a pending/deferred terminal must never masquerade as a successful answer.
func ValidateAssistantResult(result *StreamResult) error {
	if result == nil || result.Message == nil {
		return fmt.Errorf("provider returned no terminal message")
	}
	if result.Incomplete {
		return fmt.Errorf("provider returned an incomplete response")
	}
	result.Message = objectToolInputs(result.Message)
	message := result.Message
	if message.Role != goai.RoleAssistant {
		return fmt.Errorf("provider terminal role must be assistant")
	}
	calls := 0
	ids := map[string]bool{}
	for _, block := range message.Content {
		switch block.Type {
		case "text", "thinking":
		case "toolCall":
			calls++
			if block.ID == "" || block.Name == "" || ids[block.ID] {
				return fmt.Errorf("invalid or duplicate terminal tool call")
			}
			ids[block.ID] = true
			if _, err := json.Marshal(block.Arguments); err != nil {
				return fmt.Errorf("invalid terminal tool arguments: %w", err)
			}
		default:
			return fmt.Errorf("unsupported assistant content %q", block.Type)
		}
	}
	switch message.StopReason {
	case goai.StopReasonStop, goai.StopReasonLength:
		if calls != 0 {
			return fmt.Errorf("terminal %s contradicts tool calls", message.StopReason)
		}
	case goai.StopReasonToolUse:
		// Empty toolUse is a final answer, matching the native durable contract.
	case goai.StopReasonPending:
		return fmt.Errorf("pending is not a valid terminal stop reason")
	case goai.StopReasonDeferred:
		return fmt.Errorf("deferred provider results are not supported")
	case goai.StopReasonError, goai.StopReasonAborted:
		return fmt.Errorf("provider terminal %s: %s", message.StopReason, message.ErrorMessage)
	default:
		return fmt.Errorf("invalid terminal stop reason %q", message.StopReason)
	}
	if err := validateTerminalUsage(result.Usage); err != nil {
		return err
	}
	return validateTerminalUsage(message.Usage)
}

func validateTerminalUsage(usage *goai.Usage) error {
	if usage == nil {
		return nil
	}
	if usage.Input < 0 || usage.Output < 0 || usage.CacheRead < 0 || usage.CacheWrite < 0 || usage.CacheWrite1h < 0 || usage.Reasoning < 0 || usage.TotalTokens < 0 {
		return fmt.Errorf("invalid terminal usage")
	}
	for _, cost := range []float64{usage.Cost.Input, usage.Cost.Output, usage.Cost.CacheRead, usage.Cost.CacheWrite, usage.Cost.Total} {
		if cost < 0 || math.IsNaN(cost) || math.IsInf(cost, 0) {
			return fmt.Errorf("invalid terminal cost")
		}
	}
	return nil
}
