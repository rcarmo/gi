package tui

import (
	"encoding/json"
	"fmt"
	"strings"
)

func toolInvocationText(toolName string, args any) string {
	m, ok := args.(map[string]any)
	if !ok || len(m) == 0 {
		if raw, ok := args.(map[string]interface{}); ok && len(raw) > 0 {
			m = map[string]any{}
			for k, v := range raw {
				m[k] = v
			}
		} else {
			return ""
		}
	}
	for _, key := range []string{"command", "cmd", "script", "path", "query"} {
		if v, ok := m[key]; ok {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return truncate(strings.Join(strings.Fields(s), " "), 500)
			}
		}
	}
	payload, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return truncate(string(payload), 500)
}

func (c *chatTUI) toolResultBody(payload map[string]any, turnID, toolCallID, toolName string) []string {
	if output, _ := payload["output"].(string); strings.TrimSpace(output) != "" {
		if toolName == "codemode" {
			output = stripScriptHeader(output)
		}
		return toolOutputBodyLines(output)
	}
	return c.latestToolResultBody(turnID, toolCallID, toolName)
}

func toolOutputBodyLines(text string) []string {
	text = strings.TrimRight(plainTerminalOutput(text), "\r\n")
	if strings.TrimSpace(text) == "" {
		return []string{"(empty)"}
	}
	parts := strings.Split(text, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.TrimRight(part, "\r"))
	}
	return out
}
