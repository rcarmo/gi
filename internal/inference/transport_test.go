package inference

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	goai "github.com/rcarmo/go-ai"
)

func TestPiConfiguredSSEDoesNotUpgradeAndAcceptsLargeProviderEvent(t *testing.T) {
	agent, workspace := t.TempDir(), t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(agent, "absent-gi"))
	claim := base64.RawURLEncoding.EncodeToString([]byte(`{"https://api.openai.com/auth":{"chatgpt_account_id":"fixture-account"}}`))
	auth, _ := json.Marshal(map[string]any{"openai-codex": map[string]any{"type": "oauth", "access": "fixture." + claim + ".fixture"}})
	for name, body := range map[string][]byte{"settings.json": []byte(`{"transport":"sse"}`), "auth.json": auth} {
		if err := os.WriteFile(filepath.Join(agent, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	Init()
	var upgrades, posts atomic.Int32
	text := strings.Repeat("x", 40*1024) // exceeds coder/websocket's default 32 KiB message limit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			upgrades.Add(1)
			http.Error(w, "WebSocket forbidden by configured transport", 400)
			return
		}
		posts.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "event: response.output_item.added\ndata: {\"type\":\"response.output_item.added\",\"item\":{\"type\":\"message\",\"id\":\"msg\"}}\n\n")
		delta, _ := json.Marshal(map[string]any{"type": "response.output_text.delta", "delta": text})
		fmt.Fprintf(w, "event: response.output_text.delta\ndata: %s\n\n", delta)
		fmt.Fprint(w, "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"fixture\",\"status\":\"completed\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1,\"total_tokens\":2}}}\n\n")
	}))
	defer server.Close()
	id := fmt.Sprintf("gi-sse-transport-%d", time.Now().UnixNano())
	goai.RegisterModel(&goai.Model{ID: id, Provider: "openai-codex", Api: goai.ApiOpenAICodexResponses, BaseURL: server.URL, ContextWindow: 65536, MaxTokens: 16384, Input: []string{"text"}})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg := config.Load(workspace)
	result, err := StreamWithToolsWithHooks(ctx, "openai-codex/"+id, &goai.Context{Messages: []goai.Message{goai.UserMessage("hello")}}, nil, &StreamHooks{Transport: goai.Transport(cfg.Transport)})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.Text != text {
		t.Fatal("large SSE provider event was not preserved")
	}
	if upgrades.Load() != 0 || posts.Load() != 1 {
		t.Fatalf("upgrades=%d posts=%d; managed SSE was ignored", upgrades.Load(), posts.Load())
	}
}
