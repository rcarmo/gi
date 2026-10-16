package inference

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Completed compatibility calls must be objects. Never turn malformed input
// into an unmarked empty map that the host could execute.
func completedCompatToolArguments(raw string) (map[string]any, string) {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, ""
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return map[string]any{}, fmt.Sprintf("%v; received: %s", err, raw)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}, "Expected a JSON object; received: " + raw
	}
	return object, ""
}
