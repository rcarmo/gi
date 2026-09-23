// Isolated acceptance server. Only the provider is deterministic; HTTP, SSE,
// queue admission, inference checkpoints and SQLite are the production paths.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/httpserver"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
	"github.com/rcarmo/gi/internal/web"
	goai "github.com/rcarmo/go-ai"
)

func main() {
	dir := os.Getenv("GI_UX_STATE_DIR")
	var err error
	if dir == "" {
		dir, err = os.MkdirTemp("", "gi-steer-ux-")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir)
	}
	// Never read or write operator credentials.
	os.Setenv("HOME", dir)
	os.MkdirAll(filepath.Join(dir, ".pi", "agent"), 0700)
	if err = os.WriteFile(filepath.Join(dir, ".pi", "agent", "auth.json"), []byte(`{"ux-local":{"type":"api_key","key":"fixture-only","apiKey":"fixture-only"}}`), 0600); err != nil {
		log.Fatal(err)
	}
	gates := os.Getenv("GI_UX_QUEUE_GATES")
	if gates == "" {
		log.Fatal("GI_UX_QUEUE_GATES required")
	}
	var mu sync.Mutex
	seen := map[string]bool{}
	gatePattern := regexp.MustCompile(`UX steer gate:([a-zA-Z0-9_-]+)`)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		raw, _ := json.Marshal(body)
		matches := gatePattern.FindAllStringSubmatch(string(raw), -1)
		var match []string
		if len(matches) > 0 {
			match = matches[len(matches)-1]
		}
		w.Header().Set("Content-Type", "text/event-stream")
		emit := func(v any) { b, _ := json.Marshal(v); fmt.Fprintf(w, "data: %s\n\n", b); w.(http.Flusher).Flush() }
		emit(map[string]any{"id": "fixture", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"role": "assistant", "content": "provider checkpoint"}, "finish_reason": nil}}})
		if len(match) > 1 {
			token := match[1]
			mu.Lock()
			first := !seen[token]
			seen[token] = true
			mu.Unlock()
			if first {
				deadline := time.After(55 * time.Second)
				tick := time.NewTicker(20 * time.Millisecond)
				defer tick.Stop()
			wait:
				for {
					select {
					case <-r.Context().Done():
						return
					case <-deadline:
						break wait
					case <-tick.C:
						if _, err := os.Stat(filepath.Join(gates, token)); err == nil {
							break wait
						}
					}
				}
			}
		}
		// Reflect only user messages as proof of actual second-request delivery.
		var users []string
		if messages, ok := body["messages"].([]any); ok {
			for _, entry := range messages {
				m, _ := entry.(map[string]any)
				if m["role"] == "user" {
					b, _ := json.Marshal(m["content"])
					users = append(users, string(b))
				}
			}
		}
		if os.Getenv("GI_UX_COMPACTION") != "" {
			time.Sleep(50 * time.Millisecond)
		}
		// Meter scenarios select their fixture measurement in the latest user
		// request. Usage still crosses the real provider/parser/persistence path.
		tokens := 100
		if (os.Getenv("GI_UX_METER") != "" || os.Getenv("GI_UX_COMPACTION") != "") && len(users) > 0 {
			marker := regexp.MustCompile(`UX meter tokens:(\d+)`).FindStringSubmatch(users[len(users)-1])
			if len(marker) > 1 {
				if n, err := strconv.Atoi(marker[1]); err == nil && n >= 0 && n <= 4_000_000 {
					tokens = n
				}
			}
		}
		emit(map[string]any{"id": "fixture", "object": "chat.completion.chunk", "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"content": "\nreceived:" + strings.Join(users, "|")}, "finish_reason": "stop"}}, "usage": map[string]any{"prompt_tokens": tokens, "completion_tokens": 20, "total_tokens": tokens + 20}})
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer provider.Close()
	inference.Init()
	goai.RegisterModel(&goai.Model{ID: "gate", Name: "Local UX Gate", Provider: goai.Provider("ux-local"), Api: goai.ApiOpenAICompletions, BaseURL: provider.URL, Input: []string{"text"}, ContextWindow: 32000, MaxTokens: 1024})
	if os.Getenv("GI_UX_INDEX_CONFIG") != "" {
		// Seed an actual startup settings file; do not fabricate index state or
		// bypass config.Load/production scan, worker, or HTTP routes.
		if err := os.MkdirAll(filepath.Join(dir, "docs"), 0700); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".pi", "settings.json"), []byte(`{"workspaceIndex":{"extraRoots":["docs"],"extraExtensions":["nim"],"optionalRoots":["notes",".pi/skills"]}}`), 0600); err != nil {
			log.Fatal(err)
		}
	}
	cfg := config.Load(dir)
	cfg.DefaultModel = "ux-local/gate"
	cfg.EnabledModels = []string{"ux-local/gate", "test-model", "bootstrap"}
	if os.Getenv("GI_UX_CONTEXT") != "" {
		for _, entry := range []struct {
			id     string
			window int
		}{{"small", 80}, {"equal", 100}, {"large", 200}} {
			goai.RegisterModel(&goai.Model{ID: entry.id, Name: entry.id, Provider: goai.Provider("ux-local"), Api: goai.ApiOpenAICompletions, BaseURL: provider.URL, Input: []string{"text"}, ContextWindow: entry.window, MaxTokens: 32})
			cfg.EnabledModels = append(cfg.EnabledModels, "ux-local/"+entry.id)
		}
	}
	if os.Getenv("GI_UX_SETTINGS_CATALOGUE") != "" {
		for i := 0; i < 60; i++ {
			id := fmt.Sprintf("settings-%02d", i)
			goai.RegisterModel(&goai.Model{ID: id, Name: id, Provider: goai.Provider("ux-local"), Api: goai.ApiOpenAICompletions, BaseURL: provider.URL, Input: []string{"text"}, ContextWindow: 32000, MaxTokens: 32})
			cfg.EnabledModels = append(cfg.EnabledModels, "ux-local/"+id)
		}
	}
	if os.Getenv("GI_UX_METER") != "" {
		goai.RegisterModel(&goai.Model{ID: "meter", Name: "Meter", Provider: goai.Provider("ux-local"), Api: goai.ApiOpenAICompletions, BaseURL: provider.URL, Input: []string{"text"}, ContextWindow: 2_000_000, MaxTokens: 32})
		cfg.EnabledModels = append(cfg.EnabledModels, "ux-local/meter")
		cfg.DefaultModel = "ux-local/meter"
	}
	cfg.DefaultProvider = "ux-local"
	cfg.SystemPrompt = "Local acceptance fixture. Answer user messages."
	cfg.WorkspaceRoot = dir
	s, err := store.Open(filepath.Join(dir, "gi.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if os.Getenv("GI_UX_COMPACTION") != "" {
		cfg.Compaction = config.CompactionSettings{Enabled: true, ThresholdTokens: 30, KeepRecentTokens: 10}
		cfg.Hooks.TimeoutMS = 55000
	}
	cfg.Hooks.TimeoutMS = 55000
	engine := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)
	defer engine.Close()
	// Hold a real terminal-state hook before claim cleanup. Admission/SSE/store
	// remain production paths; the file gate controls only hook completion.
	var heldCompletion sync.Map
	_, err = engine.RegisterHook(turn.HookSessionState, "ux-completion-admission", func(ctx context.Context, req turn.HookRequest) (turn.HookResponse, error) {
		if req.SessionStatus != "idle" {
			return turn.HookResponse{}, nil
		}
		id, _, claimErr := s.GetSessionActiveTurn(ctx, req.SessionID)
		if claimErr != nil {
			return turn.HookResponse{}, nil
		}
		record, getErr := s.GetTurn(ctx, id)
		if getErr != nil || record.Status != "completed" || !strings.HasPrefix(record.Prompt, "UX completed claim:") {
			return turn.HookResponse{}, nil
		}
		if _, loaded := heldCompletion.LoadOrStore(id, true); loaded {
			return turn.HookResponse{}, nil
		}
		token := strings.TrimSpace(strings.TrimPrefix(record.Prompt, "UX completed claim:"))
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(token) {
			return turn.HookResponse{}, fmt.Errorf("invalid completion token")
		}
		path := filepath.Join(gates, token)
		if err := os.WriteFile(path+".held", []byte(id), 0600); err != nil {
			return turn.HookResponse{}, err
		}
		tick := time.NewTicker(20 * time.Millisecond)
		defer tick.Stop()
		for {
			if _, err := os.Stat(path); err == nil {
				return turn.HookResponse{}, nil
			}
			select {
			case <-ctx.Done():
				return turn.HookResponse{}, ctx.Err()
			case <-tick.C:
			}
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	if os.Getenv("GI_UX_COMPACTION") != "" {
		_, err = engine.RegisterHook(turn.HookSessionBeforeCompact, "ux-compaction-gate", func(ctx context.Context, req turn.HookRequest) (turn.HookResponse, error) {
			// Real hook gate; no synthetic lifecycle events or database seeding.
			if req.Payload["reason"] == "manual" {
				tick := time.NewTicker(20 * time.Millisecond)
				defer tick.Stop()
				for {
					if _, err := os.Stat(filepath.Join(gates, "manual-"+req.SessionID)); err == nil {
						break
					}
					select {
					case <-ctx.Done():
						return turn.HookResponse{}, ctx.Err()
					case <-tick.C:
					}
				}
				return turn.HookResponse{Payload: map[string]any{"summary": "Manual summary of decisions and pending work."}}, nil
			}
			last := ""
			if len(req.Messages) > 0 {
				last = goai.GetTextContent(&req.Messages[len(req.Messages)-1])
			}
			m := regexp.MustCompile(`UX compact (complete|suppress):([a-zA-Z0-9_-]+)`).FindStringSubmatch(last)
			if len(m) > 2 {
				tick := time.NewTicker(20 * time.Millisecond)
				defer tick.Stop()
				for {
					if _, err := os.Stat(filepath.Join(gates, m[2])); err == nil {
						break
					}
					select {
					case <-ctx.Done():
						return turn.HookResponse{}, ctx.Err()
					case <-tick.C:
					}
				}
				if m[1] == "suppress" {
					return turn.HookResponse{Block: true}, nil
				}
			}
			return turn.HookResponse{Payload: map[string]any{"summary": "Preserve user requirements and pending work."}}, nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	server := web.New(s, engine, cfg)
	addr := os.Getenv("GI_UX_LISTEN")
	if addr == "" {
		addr = "127.0.0.1:19092"
	}
	httpServer := &http.Server{Addr: addr, Handler: server.Handler()}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := server.StartWorkspaceIndex(ctx); err != nil {
		log.Fatal(err)
	}
	defer server.CloseWorkspaceIndex()
	if err := httpserver.Run(ctx, stop, 5*time.Second, httpserver.Listener{Server: httpServer, Serve: httpServer.ListenAndServe, Label: "ux"}); err != nil {
		log.Printf("UX HTTP stopped: %v", err)
	}
}
