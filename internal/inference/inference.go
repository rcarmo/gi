package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	goai "github.com/rcarmo/go-ai"
	_ "github.com/rcarmo/go-ai/provider/anthropic"
	_ "github.com/rcarmo/go-ai/provider/openai"
	_ "github.com/rcarmo/go-ai/provider/openaicodex"
	_ "github.com/rcarmo/go-ai/provider/openairesponses"
)

var once sync.Once

type authEntry struct {
	Type    string `json:"type"`
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
	Expires int64  `json:"expires"`
	Token   string `json:"token"`
	APIKey  string `json:"apiKey"`
}

func Init() {
	once.Do(func() {
		goai.RegisterBuiltinModels()
	})
}

func loadAuth(provider string) (string, string, error) {
	home, _ := os.UserHomeDir()
	authPath := filepath.Join(home, ".pi", "agent", "auth.json")
	data, err := os.ReadFile(authPath)
	if err != nil {
		return "", "", fmt.Errorf("read auth.json: %w", err)
	}
	var entries map[string]authEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return "", "", fmt.Errorf("parse auth.json: %w", err)
	}
	entry, ok := entries[provider]
	if !ok {
		return "", "", fmt.Errorf("no auth entry for provider %q", provider)
	}

	if provider == "github-copilot" {
		refreshToken := entry.Refresh
		if refreshToken == "" {
			refreshToken = entry.Access
		}
		if refreshToken == "" {
			return "", "", fmt.Errorf("no access/refresh token for github-copilot")
		}
		token, baseURL, err := refreshCopilotToken(refreshToken)
		if err != nil {
			return "", "", fmt.Errorf("refresh copilot token: %w", err)
		}
		return token, baseURL, nil
	}

	if entry.Access != "" {
		return entry.Access, "", nil
	}
	if entry.Token != "" {
		return entry.Token, "", nil
	}
	if entry.APIKey != "" {
		return entry.APIKey, "", nil
	}
	return "", "", fmt.Errorf("no token/key in auth entry for %q", provider)
}

// StreamResult holds the full result from a streaming inference call.
type StreamResult struct {
	Message *goai.Message
	Usage   *goai.Usage
	Text    string
}

// StreamWithTools streams a single LLM call (which may produce tool calls).
// It returns the complete assistant message including any tool-call content blocks.
// The broadcast callback receives SSE-shaped events for real-time UI updates.
func StreamWithTools(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*StreamResult, error) {
	Init()

	provider, modelName := splitModelID(modelID)
	model := goai.GetModel(goai.Provider(provider), modelName)
	if model == nil {
		return nil, fmt.Errorf("model not found: %s/%s", provider, modelName)
	}

	apiKey, baseURLOverride, err := loadAuth(provider)
	if err != nil {
		return nil, fmt.Errorf("load auth for %s: %w", provider, err)
	}
	if baseURLOverride != "" {
		model.BaseURL = baseURLOverride
	}

	opts := &goai.StreamOptions{
		APIKey: apiKey,
	}
	if provider == "github-copilot" {
		opts.Headers = goai.CopilotHeaders()
	}

	var fullText string
	var result *StreamResult

	for ev := range goai.Stream(ctx, model, convCtx, opts) {
		switch e := ev.(type) {
		case *goai.TextDeltaEvent:
			fullText += e.Delta
			if broadcast != nil {
				broadcast(map[string]any{"type": "text_delta", "delta": e.Delta})
			}
		case *goai.ThinkingDeltaEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "thinking_delta", "delta": e.Delta})
			}
		case *goai.ToolCallStartEvent:
			if broadcast != nil && e.Partial != nil {
				// Extract tool name from partial message
				for _, b := range e.Partial.Content {
					if b.Type == "toolCall" && b.Name != "" {
						broadcast(map[string]any{"type": "tool_call_start", "name": b.Name, "index": e.ContentIndex})
						break
					}
				}
			}
		case *goai.ToolCallDeltaEvent:
			// accumulation happens inside go-ai; no need to broadcast deltas
		case *goai.ToolCallEndEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "tool_call_end", "name": e.ToolCall.Name, "id": e.ToolCall.ID})
			}
		case *goai.DoneEvent:
			usage := e.Message.Usage
			if broadcast != nil {
				usageMap := map[string]any{}
				if usage != nil {
					usageMap = map[string]any{
						"input": usage.Input, "output": usage.Output,
						"total":      usage.TotalTokens,
						"cache_read": usage.CacheRead, "cache_write": usage.CacheWrite,
						"cost_input": usage.Cost.Input, "cost_output": usage.Cost.Output,
						"cost_total": usage.Cost.Total,
					}
				}
				broadcast(map[string]any{"type": "done", "model": modelID, "usage": usageMap})
			}
			result = &StreamResult{Message: e.Message, Usage: usage, Text: fullText}
			return result, nil
		case *goai.ErrorEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "error", "error": e.Err.Error()})
			}
			return nil, e.Err
		}
	}

	// Stream ended without a DoneEvent — return what we have
	return &StreamResult{Text: fullText}, nil
}

func splitModelID(id string) (string, string) {
	for i, c := range id {
		if c == '/' {
			return id[:i], id[i+1:]
		}
	}
	return "", id
}

func refreshCopilotToken(refreshToken string) (string, string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/copilot_internal/v2/token", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "token "+refreshToken)
	for k, v := range goai.CopilotHeaders() {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	var raw struct {
		Token     string `json:"token"`
		Endpoints struct {
			API string `json:"api"`
		} `json:"endpoints"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", "", err
	}
	baseURL := raw.Endpoints.API
	if baseURL == "" {
		baseURL = "https://api.individual.githubcopilot.com"
	}
	return raw.Token, baseURL, nil
}
