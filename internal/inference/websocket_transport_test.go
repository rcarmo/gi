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
	"github.com/rcarmo/go-ai/transports/websocket"
)

func TestPiConfiguredWebSocketAcceptsLargeProviderEvent(t *testing.T) {
	for _, transport := range []string{"websocket", "auto"} {
		t.Run(transport, func(t *testing.T) {
			agent, workspace := t.TempDir(), t.TempDir()
			t.Setenv("PI_CODING_AGENT_DIR", agent)
			t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(agent, "absent-gi"))
			claim := base64.RawURLEncoding.EncodeToString([]byte(`{"https://api.openai.com/auth":{"chatgpt_account_id":"fixture-account"}}`))
			auth, _ := json.Marshal(map[string]any{"openai-codex": map[string]any{"type": "oauth", "access": "fixture." + claim + ".fixture"}})
			settings, _ := json.Marshal(map[string]string{"transport": transport})
			for name, body := range map[string][]byte{"settings.json": settings, "auth.json": auth} {
				if err := os.WriteFile(filepath.Join(agent, name), body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			Init()
			text := strings.Repeat("x", 1024*1024)
			var upgrades, posts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					posts.Add(1)
					http.Error(w, "unexpected SSE fallback", http.StatusBadRequest)
					return
				}
				conn, err := websocket.Accept(w, r, nil)
				if err != nil {
					t.Errorf("accept: %v", err)
					return
				}
				defer conn.CloseNow()
				upgrades.Add(1)
				if _, _, err := conn.Read(r.Context()); err != nil {
					t.Errorf("request: %v", err)
					return
				}
				for _, event := range []any{
					map[string]any{"type": "response.output_item.added", "item": map[string]any{"type": "message", "id": "msg"}},
					map[string]any{"type": "response.output_text.delta", "delta": text},
					map[string]any{"type": "response.completed", "response": map[string]any{"id": "fixture", "status": "completed"}},
				} {
					data, _ := json.Marshal(event)
					if err := conn.Write(r.Context(), websocket.MessageText, data); err != nil {
						t.Errorf("event: %v", err)
						return
					}
				}
			}))
			defer server.Close()
			id := fmt.Sprintf("gi-ws-transport-%d", time.Now().UnixNano())
			goai.RegisterModel(&goai.Model{ID: id, Provider: "openai-codex", Api: goai.ApiOpenAICodexResponses, BaseURL: server.URL, ContextWindow: 65536, MaxTokens: 16384, Input: []string{"text"}})
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cfg := config.Load(workspace)
			result, err := StreamWithToolsWithHooks(ctx, "openai-codex/"+id, &goai.Context{Messages: []goai.Message{goai.UserMessage("hello")}}, nil, &StreamHooks{Transport: goai.Transport(cfg.Transport)})
			if err != nil {
				t.Fatal(err)
			}
			if result == nil || result.Text != text {
				t.Fatal("large WebSocket event was not preserved")
			}
			if upgrades.Load() != 1 || posts.Load() != 0 {
				t.Fatalf("upgrades=%d posts=%d; WebSocket setting was ignored", upgrades.Load(), posts.Load())
			}
		})
	}
}
