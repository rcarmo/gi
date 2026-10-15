package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	gisession "github.com/rcarmo/gi/internal/session"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/turn"
)

// The Make harness owns the PTY, private files and deterministic event trigger.
func TestStartupAndHookPTYFixture(t *testing.T) {
	root := os.Getenv("GI_STARTUP_PTY_DIR")
	if root == "" {
		t.Skip("make test-startup-hooks-pty")
	}
	cfg := config.Load(root)
	cfg.DefaultModel = "bootstrap"
	cfg.QuietStartup = os.Getenv("GI_STARTUP_PTY_QUIET")
	s, err := store.Open(filepath.Join(root, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	allocation := gisession.AllocateDefaultSession(defaultTUIAgentID, "gi", "default", "startup-session")
	session, _, err := s.ResolveOrCreateMainSessionFromAllocation(ctx, store.ResolveOrCreateSessionFromAllocationInput{ID: "startup-session", Title: "Startup", State: map[string]any{"model": "bootstrap"}, Allocation: allocation})
	if err != nil {
		t.Fatal(err)
	}
	engine := turn.NewWithRuntimeConfig(s, cfg, cfg.SystemPrompt)
	defer engine.Close()
	done := make(chan struct{})
	joined := make(chan struct{})
	go func() {
		defer close(joined)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if _, err := os.Stat(filepath.Join(root, "fire-hooks")); err != nil {
					continue
				}
				req := turn.HookRequest{Name: turn.HookBeforeAgentStart, SessionID: session.ID}
				engine.PublishRuntimeHookEvent("hook_invocation", req, "fixture", "continue", 1, nil)
				engine.PublishRuntimeHookDecisionEvent("hook_modify", req, map[string]any{"reason": "fixture modification"})
				engine.PublishRuntimeHookDecisionEvent("hook_respond", req, map[string]any{"reason": "fixture response"})
				engine.PublishRuntimeHookDecisionEvent("hook_deny", req, map[string]any{"reason": "fixture denial remains visible"})
				return
			}
		}
	}()
	defer func() { close(done); <-joined }()
	if err := runWithEngineMode(s, engine, cfg, os.Getenv("GI_STARTUP_PTY_MODE") == "regular", os.Getenv("GI_STARTUP_PTY_DEBUG") == "1"); err != nil {
		t.Fatal(err)
	}
}
