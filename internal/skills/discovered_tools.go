package skills

import (
	"encoding/json"
	"strings"
)

func RegisterDiscoveredTools(manifests []ToolManifest, register func(ToolManifest, json.RawMessage) error, onError func(string, error)) {
	for _, manifest := range manifests {
		tool := manifest
		if strings.TrimSpace(tool.Name) == "" || (strings.TrimSpace(tool.Script) == "" && strings.TrimSpace(tool.Path) == "") {
			continue
		}
		params := tool.Parameters
		if len(params) == 0 {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		if err := register(tool, params); err != nil && onError != nil {
			onError(tool.Name, err)
		}
	}
}
