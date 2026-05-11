package inference

import (
	"context"
	"strings"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestInitRegistersOpenCodeZenModel(t *testing.T) {
	Init()
	model := goai.GetModel(goai.Provider("opencode-zen"), "minimax-m2.5-free")
	if model == nil {
		t.Fatal("expected opencode-zen/minimax-m2.5-free to be registered")
	}
	if model.BaseURL != "https://opencode.ai/zen/v1" {
		t.Fatalf("unexpected base URL: %q", model.BaseURL)
	}
}

func TestStreamWithToolsOpenCodeZenSmoke(t *testing.T) {
	res, err := StreamWithTools(context.Background(), "opencode-zen/minimax-m2.5-free", &goai.Context{Messages: []goai.Message{goai.UserMessage("Say hello in one word.")}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	norm := strings.ToLower(strings.TrimSpace(res.Text))
	norm = strings.Trim(norm, "!?.:,; \t\r\n")
	if !strings.Contains(norm, "hello") && norm != "hi" {
		t.Fatalf("unexpected smoke response: %q", res.Text)
	}
}

func TestLoadAuthAllowsOpenCodeZenWithoutSecret(t *testing.T) {
	apiKey, baseURL, err := loadAuth("opencode-zen")
	if err != nil {
		t.Fatal(err)
	}
	if apiKey != "" || baseURL != "https://opencode.ai/zen/v1" {
		t.Fatalf("unexpected auth tuple: apiKey=%q baseURL=%q", apiKey, baseURL)
	}
}
