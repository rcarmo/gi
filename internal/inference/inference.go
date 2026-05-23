package inference

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	goai "github.com/rcarmo/go-ai"
	_ "github.com/rcarmo/go-ai/inference/provider/anthropic"
	_ "github.com/rcarmo/go-ai/inference/provider/openai"
	_ "github.com/rcarmo/go-ai/inference/provider/openaicodex"
	_ "github.com/rcarmo/go-ai/inference/provider/openairesponses"
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
		registerCustomModels()
	})
}

func registerCustomModels() {
	goai.RegisterModel(&goai.Model{
		ID:            "minimax-m2.5-free",
		Name:          "OpenCode Zen / MiniMax M2.5 Free",
		Api:           goai.ApiOpenAICompletions,
		Provider:      "opencode-zen",
		BaseURL:       "https://opencode.ai/zen/v1",
		Reasoning:     false,
		Input:         []string{"text"},
		ContextWindow: 128000,
		MaxTokens:     16384,
	})
}

func loadAuth(provider string) (string, string, error) {
	if provider == "opencode-zen" {
		return "", "https://opencode.ai/zen/v1", nil
	}
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

type StreamHooks struct {
	OnPayload  func(payload any, model *goai.Model) (any, error)
	OnResponse func(status int, headers map[string]string, model *goai.Model)
}

// StreamWithTools streams a single LLM call (which may produce tool calls).
// It returns the complete assistant message including any tool-call content blocks.
// The broadcast callback receives SSE-shaped events for real-time UI updates.
func StreamWithTools(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any)) (*StreamResult, error) {
	return StreamWithToolsWithHooks(ctx, modelID, convCtx, broadcast, nil)
}

// StreamWithToolsWithHooks is the hook-aware variant of StreamWithTools.
func StreamWithToolsWithHooks(ctx context.Context, modelID string, convCtx *goai.Context, broadcast func(map[string]any), hooks *StreamHooks) (*StreamResult, error) {
	Init()

	provider, modelName := splitModelID(modelID)
	model := goai.GetModel(goai.Provider(provider), modelName)
	if model == nil {
		return nil, fmt.Errorf("model not found: %s/%s", provider, modelName)
	}
	if provider == "opencode-zen" {
		return streamOpenCodeZen(ctx, model, convCtx, broadcast, hooks)
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
	if hooks != nil {
		opts.OnPayload = hooks.OnPayload
		opts.OnResponse = hooks.OnResponse
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

func responseHeadersMap(header http.Header) map[string]string {
	if len(header) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) == 0 {
			out[key] = ""
			continue
		}
		out[key] = strings.Join(values, ", ")
	}
	return out
}

func splitModelID(id string) (string, string) {
	for i, c := range id {
		if c == '/' {
			return id[:i], id[i+1:]
		}
	}
	return "", id
}

func streamOpenCodeZen(ctx context.Context, model *goai.Model, convCtx *goai.Context, broadcast func(map[string]any), hooks *StreamHooks) (*StreamResult, error) {
	payload := map[string]any{
		"model":    model.ID,
		"stream":   true,
		"messages": openAICompatMessages(convCtx),
	}
	if len(convCtx.Tools) > 0 {
		payload["tools"] = openAICompatTools(convCtx.Tools)
	}
	var payloadValue any = payload
	if hooks != nil && hooks.OnPayload != nil {
		replaced, err := hooks.OnPayload(payloadValue, model)
		if err != nil {
			return nil, err
		}
		if replaced != nil {
			payloadValue = replaced
		}
	}
	body, err := json.Marshal(payloadValue)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(model.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if hooks != nil && hooks.OnResponse != nil {
		hooks.OnResponse(resp.StatusCode, responseHeadersMap(resp.Header), model)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	var fullText, fullThinking string
	usage := &goai.Usage{}
	toolCalls := map[int]*goai.ToolCall{}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string `json:"content"`
					Reasoning string `json:"reasoning"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Usage.TotalTokens > 0 {
			usage.Input = chunk.Usage.PromptTokens
			usage.Output = chunk.Usage.CompletionTokens
			usage.TotalTokens = chunk.Usage.TotalTokens
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Reasoning != "" {
				fullThinking += choice.Delta.Reasoning
				if broadcast != nil {
					broadcast(map[string]any{"type": "thinking_delta", "delta": choice.Delta.Reasoning})
				}
			}
			if choice.Delta.Content != "" {
				fullText += choice.Delta.Content
				if broadcast != nil {
					broadcast(map[string]any{"type": "text_delta", "delta": choice.Delta.Content})
				}
			}
			for _, tc := range choice.Delta.ToolCalls {
				call := toolCalls[tc.Index]
				if call == nil {
					call = &goai.ToolCall{Type: "toolCall"}
					toolCalls[tc.Index] = call
					if broadcast != nil && tc.Function.Name != "" {
						broadcast(map[string]any{"type": "tool_call_start", "name": tc.Function.Name, "index": tc.Index})
					}
				}
				if tc.ID != "" {
					call.ID = tc.ID
				}
				if tc.Function.Name != "" {
					call.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					if call.Arguments == nil {
						call.Arguments = map[string]any{}
					}
					var args map[string]any
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
						call.Arguments = args
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	msg := &goai.Message{Role: goai.RoleAssistant, Model: model.ID, Provider: model.Provider, Api: model.Api, Usage: usage}
	if fullThinking != "" {
		msg.Content = append(msg.Content, goai.ContentBlock{Type: "thinking", Text: fullThinking})
	}
	if fullText != "" {
		msg.Content = append(msg.Content, goai.ContentBlock{Type: "text", Text: fullText})
	}
	for _, idx := range sortedToolCallIndexes(toolCalls) {
		call := toolCalls[idx]
		msg.Content = append(msg.Content, goai.ContentBlock{Type: "toolCall", ID: call.ID, Name: call.Name, Arguments: call.Arguments})
		if broadcast != nil {
			broadcast(map[string]any{"type": "tool_call_end", "name": call.Name, "id": call.ID})
		}
	}
	if broadcast != nil {
		broadcast(map[string]any{"type": "done", "model": string(model.Provider) + "/" + model.ID, "usage": map[string]any{"input": usage.Input, "output": usage.Output, "total": usage.TotalTokens}})
	}
	return &StreamResult{Message: msg, Usage: usage, Text: fullText}, nil
}

func openAICompatMessages(convCtx *goai.Context) []map[string]any {
	messages := make([]map[string]any, 0, len(convCtx.Messages)+1)
	if strings.TrimSpace(convCtx.SystemPrompt) != "" {
		messages = append(messages, map[string]any{"role": "system", "content": convCtx.SystemPrompt})
	}
	for _, msg := range convCtx.Messages {
		role := strings.ToLower(string(msg.Role))
		content := goai.GetTextContent(&msg)
		if role == "tool" || role == "toolresult" {
			entry := map[string]any{"role": "tool", "content": content}
			if msg.ToolCallID != "" {
				entry["tool_call_id"] = msg.ToolCallID
			}
			messages = append(messages, entry)
			continue
		}
		messages = append(messages, map[string]any{"role": role, "content": content})
	}
	return messages
}

func openAICompatTools(tools []goai.Tool) []map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  json.RawMessage(tool.Parameters),
			},
		})
	}
	return out
}

func sortedToolCallIndexes(calls map[int]*goai.ToolCall) []int {
	idx := make([]int, 0, len(calls))
	for k := range calls {
		idx = append(idx, k)
	}
	for i := 0; i < len(idx); i++ {
		for j := i + 1; j < len(idx); j++ {
			if idx[j] < idx[i] {
				idx[i], idx[j] = idx[j], idx[i]
			}
		}
	}
	return idx
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
