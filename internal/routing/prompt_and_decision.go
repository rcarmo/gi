package routing

import "strings"

func ParseDirectedPrompt(prompt string) (string, string, bool) {
	trimmed := strings.TrimSpace(prompt)
	if !strings.HasPrefix(trimmed, "@") {
		return "", prompt, false
	}
	rest := trimmed[1:]
	end := strings.IndexAny(rest, " \t\r\n:")
	if end < 0 {
		return NormalizeAgentID(rest), "", true
	}
	target := NormalizeAgentID(rest[:end])
	body := strings.TrimSpace(rest[end:])
	body = strings.TrimPrefix(body, ":")
	body = strings.TrimSpace(body)
	return target, body, target != ""
}
