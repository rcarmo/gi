package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	goai "github.com/rcarmo/go-ai"
)

// Disposable native TUI/provider harness, compiled only into the test binary.
func TestMarkdownTablePTYFixture(t *testing.T) {
	dir := os.Getenv("GI_TABLE_PTY_DIR")
	if dir == "" {
		t.Skip("PTY driver only")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "chunks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var chunks []string
	if err = json.Unmarshal(raw, &chunks); err != nil {
		t.Fatal(err)
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		f := w.(http.Flusher)
		for i, chunk := range chunks {
			deadline := time.Now().Add(40 * time.Second)
			for {
				if _, err := os.Stat(filepath.Join(dir, fmt.Sprintf("gate-%d", i))); err == nil {
					break
				}
				if time.Now().After(deadline) {
					return
				}
				select {
				case <-r.Context().Done():
					return
				case <-time.After(15 * time.Millisecond):
				}
			}
			data, _ := json.Marshal(map[string]any{"id": "table", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"content": chunk}, "finish_reason": nil}}})
			fmt.Fprintf(w, "data: %s\n\n", data)
			f.Flush()
			os.WriteFile(filepath.Join(dir, fmt.Sprintf("sent-%d", i)), nil, 0600)
		}
		for {
			if _, err := os.Stat(filepath.Join(dir, "finish")); err == nil {
				break
			}
			select {
			case <-r.Context().Done():
				return
			case <-time.After(15 * time.Millisecond):
			}
		}
		fmt.Fprint(w, "data: {\"id\":\"table\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")
		f.Flush()
	}))
	defer provider.Close()
	inference.Init()
	goai.RegisterModel(&goai.Model{ID: "table-fixture", Name: "Table fixture", Provider: "table-local", Api: goai.ApiOpenAICompletions, BaseURL: provider.URL, MaxTokens: 4096, ContextWindow: 16384})
	home := filepath.Join(dir, "home")
	os.MkdirAll(filepath.Join(home, ".pi", "agent"), 0700)
	t.Setenv("HOME", home)
	t.Setenv("GI_CODING_AGENT_DIR", "")
	t.Setenv("PI_CODING_AGENT_DIR", "")
	os.WriteFile(filepath.Join(home, ".pi", "agent", "auth.json"), []byte(`{"table-local":{"type":"api_key","apiKey":"fixture"}}`), 0600)
	cfg := config.Load(dir)
	cfg.DefaultModel = "table-local/table-fixture"
	cfg.DefaultProvider = "table-local"
	s, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	engine := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)
	defer engine.Close()
	if err = runWithEngineMode(s, engine, cfg, os.Getenv("GI_TABLE_PTY_MODE") == "regular"); err != nil {
		t.Fatal(err)
	}
}
