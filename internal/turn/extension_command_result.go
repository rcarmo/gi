package turn

import (
	"encoding/json"
	"strings"
)

func extensionCommandResultFromScript(raw string) (ExtensionCommandResult, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ExtensionCommandResult{Type: "noop"}, nil
	}
	var res ExtensionCommandResult
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return ExtensionCommandResult{Type: "message", Lines: []string{raw}}, nil
	}
	if strings.TrimSpace(res.Type) == "" {
		res.Type = "message"
	}
	return res, nil
}
