package turn

import (
	"encoding/json"
	"strings"

	goai "github.com/rcarmo/go-ai"
)

func hookResponseFromScript(result string) (HookResponse, error) {
	var resp HookResponse
	if strings.TrimSpace(result) == "" {
		return resp, nil
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(result), &raw); err != nil {
		return resp, nil
	}
	b, _ := json.Marshal(raw)
	_ = json.Unmarshal(b, &resp)
	if tc, ok := raw["tool_call"]; ok {
		if call := decodeToolCall(tc); call != nil {
			resp.ToolCall = call
		}
	}
	if payload, ok := raw["payload"].(map[string]any); ok {
		if tc, ok := payload["tool_call"]; ok {
			if call := decodeToolCall(tc); call != nil {
				resp.ToolCall = call
			}
		}
	}
	resp.Action = strings.ToLower(strings.TrimSpace(resp.Action))
	if resp.Action == "" {
		if action := stringValue(raw["action"], ""); strings.TrimSpace(action) != "" {
			resp.Action = strings.ToLower(strings.TrimSpace(action))
		}
	}
	switch resp.Action {
	case "continue":
		// no-op
	case "modify":
		// modifications are carried through regular response fields
	case "respond":
		resp.Handled = true
		if strings.TrimSpace(resp.Message) == "" {
			if response := stringValue(raw["response"], ""); strings.TrimSpace(response) != "" {
				resp.Message = response
			} else if payload, ok := raw["payload"].(map[string]any); ok {
				resp.Message = stringValue(payload["response"], "")
			}
		}
	case "deny":
		resp.Block = true
		if strings.TrimSpace(resp.Reason) == "" {
			resp.Reason = "denied by hook"
		}
	case "abort_turn":
		resp.Cancel = true
		resp.Block = true
		if strings.TrimSpace(resp.Reason) == "" {
			resp.Reason = "aborted by hook"
		}
	case "hard_abort":
		resp.Cancel = true
		resp.Block = true
		if resp.Payload == nil {
			resp.Payload = map[string]any{}
		}
		resp.Payload["hard_abort"] = true
		if strings.TrimSpace(resp.Reason) == "" {
			resp.Reason = "hard aborted by hook"
		}
	}
	return resp, nil
}

func decodeToolCall(raw any) *goai.ToolCall {
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var call goai.ToolCall
	if err := json.Unmarshal(b, &call); err != nil {
		return nil
	}
	if call.Name == "" {
		var alt struct {
			ID        string         `json:"id"`
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(b, &alt); err != nil || alt.Name == "" {
			return nil
		}
		call.ID = alt.ID
		call.Name = alt.Name
		call.Arguments = alt.Arguments
	}
	if call.Arguments == nil {
		call.Arguments = map[string]any{}
	}
	if call.Type == "" {
		call.Type = "toolCall"
	}
	return &call
}
