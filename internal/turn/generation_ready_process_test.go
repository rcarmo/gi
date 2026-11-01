package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestGenerationReadyProcessHelper(t *testing.T) {
	path := os.Getenv("GI_GENERATION_READY_DB")
	if path == "" {
		t.Skip("subprocess helper")
	}
	ctx := context.Background()
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	e := NewWithRuntimeConfig(s, config.Load(filepath.Dir(path)), "")
	defer e.Close()
	if _, err = s.CreateSession(ctx, "ready", "ready", nil); err != nil {
		t.Fatal(err)
	}
	if err = e.RegisterTool(tools.RegisteredTool{Name: "ready-effect", Parameters: json.RawMessage(`{"type":"object"}`), Executor: func(_ context.Context, _ tools.ToolRuntime, call goai.ToolCall) (string, error) {
		if err := os.WriteFile(path+".effect", []byte(call.Arguments["value"].(string)), 0600); err != nil {
			return "", err
		}
		return "committed tool result", nil
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err = e.RegisterHook(HookBeforeAgentStart, "fixture", func(context.Context, HookRequest) (HookResponse, error) {
		return HookResponse{Message: "once-only lifecycle prefix"}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = e.RegisterHook(HookApproveTool, "fixture", func(_ context.Context, req HookRequest) (HookResponse, error) {
		call := *req.ToolCall
		call.Arguments = map[string]any{"value": "prepared"}
		return HookResponse{ToolCall: &call}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = e.RegisterHook(HookBeforeProviderRequest, "fixture", func(_ context.Context, req HookRequest) (HookResponse, error) {
		if req.Iteration == 2 {
			if err := os.WriteFile(path+".ready", []byte(req.TurnID), 0600); err != nil {
				return HookResponse{}, err
			}
			select {}
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "one-call", Name: "ready-effect", Arguments: map[string]any{"value": "original"}}}, Usage: &goai.Usage{Input: 7, Output: 2}}}, nil
	})
	if _, err = e.SubmitPrompt(ctx, RunInput{SessionID: "ready", Prompt: "original prompt", Model: "mock-stream"}); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestGenerationReadySurvivesKilledOwnerWithoutReplayingTools(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires process kill")
	}
	root := t.TempDir()
	path := filepath.Join(root, "generation.db")
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pi", "settings.json"), []byte(`{"defaultModel":"mock-stream","maxIterations":4,"retry":{"enabled":false}}`), 0600); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(binary, "-test.run=^TestGenerationReadyProcessHelper$")
	child.Env = append(os.Environ(), "GI_GENERATION_READY_DB="+path)
	var output bytes.Buffer
	child.Stdout = &output
	child.Stderr = &output
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	joined := false
	defer func() {
		if !joined {
			_ = child.Process.Kill()
			_ = child.Wait()
		}
	}()
	waitForCondition(t, 8*time.Second, func() bool { _, err := os.Stat(path + ".ready"); return err == nil }, "settled tool batch generation boundary")
	if err = child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err = child.Wait()
	joined = true
	if err == nil {
		t.Fatal("helper was not killed", output.String())
	}
	turnBytes, err := os.ReadFile(path + ".ready")
	if err != nil {
		t.Fatal(err)
	}
	turnID := string(turnBytes)
	effect, err := os.ReadFile(path + ".effect")
	if err != nil || string(effect) != "prepared" {
		t.Fatal("tool did not run prepared arguments", string(effect), err)
	}
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	row, err := s.GetTurn(context.Background(), turnID)
	if err != nil || row.Phase != "generation_ready" {
		t.Fatal(row, err)
	}
	if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("fresh generation instructions"), 0600); err != nil {
		t.Fatal(err)
	}
	var calls, toolsRun, starts atomic.Int32
	withStreamWithToolsStub(t, func(_ context.Context, _ string, request *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
		calls.Add(1)
		raw, _ := json.Marshal(request)
		for _, want := range []string{"once-only lifecycle prefix", "committed tool result", "fresh generation instructions", "current tool description"} {
			if !strings.Contains(string(raw), want) {
				t.Errorf("fresh continuation lacks %q: %s", want, raw)
			}
		}
		if strings.Count(string(raw), "original prompt") != 1 || strings.Count(string(raw), "once-only lifecycle prefix") != 1 {
			t.Error("prompt/lifecycle lost or replayed", string(raw))
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "continued settled generation"}}, Usage: &goai.Usage{Input: 11, Output: 3}}}, nil
	})
	e := NewWithRuntimeConfig(s, config.Load(root), "")
	defer e.Close()
	if err = e.RegisterTool(tools.RegisteredTool{Name: "ready-effect", Description: "current tool description", Parameters: json.RawMessage(`{"type":"object"}`), Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		toolsRun.Add(1)
		return "must not execute again", nil
	}}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{HookBeforeAgentStart, HookAgentStart, HookToolCall, HookApproveTool} {
		if _, err = e.RegisterHook(name, "fixture", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(30 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatal("provider dispatched before runtime-ready barrier")
	}
	if err = e.ResumeInterruptedWork(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, 5*time.Second, func() bool {
		rec, err := s.GetTurn(context.Background(), turnID)
		return err == nil && rec.Status == "completed"
	}, "recovered generation completion")
	e.runs.Wait()
	if calls.Load() != 1 || toolsRun.Load() != 0 || starts.Load() != 0 {
		t.Fatal("continuation replayed work", calls.Load(), toolsRun.Load(), starts.Load())
	}
	messages, err := s.ListMessages(context.Background(), "ready")
	if err != nil {
		t.Fatal(err)
	}
	prompts, results := 0, 0
	for _, m := range messages {
		if m.Role == "user" && m.Content == "original prompt" {
			prompts++
		}
		if m.Role == "tool_result" {
			results++
		}
	}
	if prompts != 1 || results != 1 {
		t.Fatal("duplicated prompt/tool result", prompts, results)
	}
	accounting, err := s.TurnAccounting(context.Background(), turnID)
	if err != nil || accounting.Iterations != 2 || accounting.Usage.Input != 18 || accounting.Unknown != 0 {
		t.Fatal("receipt accounting changed", accounting, err)
	}
}

func TestGenerationReadyFailurePreventsProviderDispatch(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`create trigger reject_ready before insert on turn_events when new.event_type='generation.ready' begin select raise(abort,'ready unavailable');end`); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		calls.Add(1)
		return nil, nil
	})
	e := New(s)
	defer e.Close()
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "must not dispatch", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	e.runs.Wait()
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "failed" || calls.Load() != 0 {
		t.Fatal("provider crossed an uncommitted frontier", row, calls.Load(), err)
	}
}
