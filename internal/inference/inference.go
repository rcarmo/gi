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

	// GitHub Copilot: exchange the refresh token for a Copilot session token
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

func Complete(ctx context.Context, modelID string, systemPrompt string, messages []goai.Message) (string, error) {
	return CompleteWithBroadcast(ctx, modelID, systemPrompt, messages, nil)
}

func CompleteWithBroadcast(ctx context.Context, modelID string, systemPrompt string, messages []goai.Message, broadcast func(map[string]any)) (string, error) {
	Init()

	// Parse provider/model from "provider/model" format
	provider, modelName := splitModelID(modelID)

	model := goai.GetModel(goai.Provider(provider), modelName)
	if model == nil {
		return "", fmt.Errorf("model not found: %s/%s", provider, modelName)
	}

	apiKey, baseURLOverride, err := loadAuth(provider)
	if err != nil {
		return "", fmt.Errorf("load auth for %s: %w", provider, err)
	}

	// Override the base URL if the auth provider gives us one
	// (e.g., enterprise vs individual Copilot endpoints)
	if baseURLOverride != "" {
		model.BaseURL = baseURLOverride
	}

	convCtx := &goai.Context{
		SystemPrompt: systemPrompt,
		Messages:     messages,
	}

	opts := &goai.StreamOptions{
		APIKey: apiKey,
	}

	// Add Copilot headers for GitHub Copilot providers
	if provider == "github-copilot" {
		opts.Headers = goai.CopilotHeaders()
	}

	// Stream and collect
	var fullText string
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
		case *goai.DoneEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "done", "model": modelID})
			}
			return fullText, nil
		case *goai.ErrorEvent:
			if broadcast != nil {
				broadcast(map[string]any{"type": "error", "error": e.Err.Error()})
			}
			return fullText, e.Err
		}
	}
	return fullText, nil
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
