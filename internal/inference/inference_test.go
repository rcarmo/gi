package inference

import (
	"context"
	"os"
	"path/filepath"
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
	if os.Getenv("GI_RUN_LIVE_INFERENCE_SMOKE") == "" {
		t.Skip("set GI_RUN_LIVE_INFERENCE_SMOKE=1 to run live provider smoke tests")
	}
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

func TestListRuntimeOptionsIncludesAuthBackedProviderModels(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	if err := os.MkdirAll(filepath.Join(root, ".pi", "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "agent", "auth.json"), []byte(`{"github-copilot":{"type":"oauth","refresh":"token","access":"token","expires":9999999999999}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	providers, models := ListRuntimeOptions("github-copilot", "github-copilot/gpt-5-mini", nil)
	foundProvider := false
	for _, provider := range providers {
		if provider.ID == "github-copilot" && provider.Authenticated {
			foundProvider = true
			break
		}
	}
	if !foundProvider {
		t.Fatalf("expected github-copilot provider in %#v", providers)
	}
	foundModel := false
	for _, model := range models {
		if model.Label == "github-copilot/gpt-5-mini" {
			foundModel = true
			break
		}
	}
	if !foundModel {
		t.Fatalf("expected github-copilot/gpt-5-mini in %#v", models)
	}
}

func TestListAuthStatusAndRemoveEntry(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	if err := os.MkdirAll(filepath.Join(root, ".pi", "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	authPath := filepath.Join(root, ".pi", "agent", "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"github-copilot":{"type":"oauth","refresh":"r","access":"a","expires":9999999999999},"custom-key":{"apiKey":"sk-x"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	statuses := ListAuthStatus()
	var copilot, custom *AuthStatus
	for i := range statuses {
		switch statuses[i].ID {
		case "github-copilot":
			copilot = &statuses[i]
		case "custom-key":
			custom = &statuses[i]
		}
	}
	if copilot == nil || !copilot.Authenticated || copilot.Kind != "oauth" {
		t.Fatalf("github-copilot status wrong: %#v", copilot)
	}
	if custom == nil || !custom.Authenticated || custom.Kind != "api-key" {
		t.Fatalf("custom-key status wrong: %#v", custom)
	}
	// Remove an entry and confirm it is gone.
	removed, err := RemoveAuthEntry("custom-key")
	if err != nil || !removed {
		t.Fatalf("remove custom-key: removed=%v err=%v", removed, err)
	}
	if again, _ := RemoveAuthEntry("custom-key"); again {
		t.Fatalf("expected second remove to be a no-op")
	}
	for _, s := range ListAuthStatus() {
		if s.ID == "custom-key" {
			t.Fatalf("custom-key should be gone after logout: %#v", s)
		}
	}
}
