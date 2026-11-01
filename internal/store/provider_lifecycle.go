package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	goai "github.com/rcarmo/go-ai"
)

// ProviderLifecyclePolicy retains accepted once-only lifecycle overrides, not
// an old provider request or executable callbacks. Nil is a legacy checkpoint
// and is deliberately ineligible for automatic continuation.
type ProviderLifecyclePolicy struct {
	Version         int             `json:"version"`
	SystemPrompt    string          `json:"system_prompt,omitempty"`
	Message         string          `json:"message,omitempty"`
	Messages        json.RawMessage `json:"messages,omitempty"`
	Tools           json.RawMessage `json:"tools,omitempty"`
	RegisteredTools map[string]bool `json:"registered_tools,omitempty"`
}

func (p *ProviderLifecyclePolicy) Clone() *ProviderLifecyclePolicy {
	if p == nil {
		return nil
	}
	copy := *p
	copy.Messages = append(json.RawMessage(nil), p.Messages...)
	copy.Tools = append(json.RawMessage(nil), p.Tools...)
	if p.RegisteredTools != nil {
		copy.RegisteredTools = make(map[string]bool, len(p.RegisteredTools))
		for k, v := range p.RegisteredTools {
			copy.RegisteredTools[k] = v
		}
	}
	return &copy
}

func (p *ProviderLifecyclePolicy) Validate() error {
	if p == nil || p.Version != 1 {
		return fmt.Errorf("missing or unsupported provider lifecycle policy")
	}
	if p.Messages != nil {
		var messages []goai.Message
		if err := decodeLifecycleArray(p.Messages, &messages); err != nil {
			return fmt.Errorf("lifecycle messages: %w", err)
		}
	}
	if p.Tools != nil {
		var tools []goai.Tool
		if err := decodeLifecycleArray(p.Tools, &tools); err != nil {
			return fmt.Errorf("lifecycle tools: %w", err)
		}
		names := make(map[string]bool, len(tools))
		for _, tool := range tools {
			if _, ok := p.RegisteredTools[tool.Name]; !ok {
				return fmt.Errorf("lifecycle tool %q lacks origin", tool.Name)
			}
			names[tool.Name] = true
		}
		for name := range p.RegisteredTools {
			if !names[name] {
				return fmt.Errorf("lifecycle origin lacks tool %q", name)
			}
		}
	} else if len(p.RegisteredTools) != 0 {
		return fmt.Errorf("lifecycle origins lack tools")
	}
	return nil
}

func decodeLifecycleArray(raw json.RawMessage, out any) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' || !json.Valid(trimmed) {
		return fmt.Errorf("expected JSON array")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	return decoder.Decode(out)
}
