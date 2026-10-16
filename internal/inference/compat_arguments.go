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
		repaired := repairCompatToolJSON(raw)
		if repaired == raw || json.Unmarshal([]byte(repaired), &value) != nil {
			return map[string]any{}, fmt.Sprintf("%v; received: %s", err, raw)
		}
	}
	object, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}, "Expected a JSON object; received: " + raw
	}
	return object, ""
}

// Repair matches go-ai v1.3.0 internal/jsonparse/tool_arguments.go (MIT).
func repairCompatToolJSON(text string) string {
	var out strings.Builder
	inside := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		if !inside {
			out.WriteByte(c)
			if c == '"' {
				inside = true
			}
			continue
		}
		if c == '"' {
			inside = false
			out.WriteByte(c)
			continue
		}
		if c == '\\' {
			if i+1 < len(text) {
				next := text[i+1]
				valid := strings.ContainsRune(`"\/bfnrt`, rune(next))
				if next == 'u' && i+5 < len(text) {
					valid = true
					for j := i + 2; j < i+6; j++ {
						h := text[j]
						if !(h >= '0' && h <= '9' || h >= 'a' && h <= 'f' || h >= 'A' && h <= 'F') {
							valid = false
							break
						}
					}
				}
				if valid {
					out.WriteByte(c)
					out.WriteByte(next)
					i++
					continue
				}
			}
			out.WriteString(`\\`)
			continue
		}
		if c < 0x20 {
			fmt.Fprintf(&out, `\u%04x`, c)
		} else {
			out.WriteByte(c)
		}
	}
	return out.String()
}
