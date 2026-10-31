package turn

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func TestProviderRetryFreshCommittedPreparationAndSubsequentToolPhase(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
	settings := filepath.Join(root, ".pi", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(path, body string) error { return os.WriteFile(path, []byte(body), 0600) }
	if err := write(settings, `{"defaultModel":"fresh-test/old","retry":{"maxRetries":1,"baseDelayMs":0,"maxAgentDelayMs":0}}`); err != nil {
		t.Fatal(err)
	}
	if err := write(filepath.Join(root, "AGENTS.md"), "original committed instructions"); err != nil {
		t.Fatal(err)
	}
	s := openTestStore(t)
	t.Cleanup(func() { s.Close() })
	e := NewWithRuntimeConfig(s, config.Load(root), "")
	t.Cleanup(func() { e.Close() })
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", map[string]any{"model": "fresh-test/old", "selected_model": "fresh-test/old"}); err != nil {
		t.Fatal(err)
	}
	inference.Init()
	high := "high"
	goai.RegisterModel(&goai.Model{ID: "new", Provider: "fresh-test", Api: goai.ApiOpenAICompletions, ContextWindow: 32000, MaxTokens: 1024, Reasoning: true, ThinkingLevelMap: map[goai.ModelThinkingLevel]*string{"high": &high}})
	var toolCalls, agentStarts, turnStarts, contexts, requestContexts atomic.Int32
	register := func(description string) error {
		return e.RegisterTool(tools.RegisteredTool{Name: "fresh_tool", Description: description, Parameters: json.RawMessage(`{"type":"object","properties":{}}`), Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
			toolCalls.Add(1)
			return "owned tool result", nil
		}})
	}
	if err := register("old definition"); err != nil {
		t.Fatal(err)
	}
	if err := e.RegisterTool(tools.RegisteredTool{Name: "withdrawn_tool", Description: "will be withdrawn", Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		t.Error("withdrawn tool executed")
		return "", nil
	}}); err != nil {
		t.Fatal(err)
	}
	if err := e.SetActiveTools([]string{"fresh_tool", "withdrawn_tool"}); err != nil {
		t.Fatal(err)
	}
	for name, counter := range map[string]*atomic.Int32{HookBeforeAgentStart: &agentStarts, HookTurnStart: &turnStarts, HookContext: &contexts} {
		if _, err := e.RegisterHook(name, "fresh-count", func(_ context.Context, req HookRequest) (HookResponse, error) {
			counter.Add(1)
			if req.Name == HookBeforeAgentStart {
				return HookResponse{Message: "once-only lifecycle prefix", Tools: e.toolDefs()}, nil
			}
			return HookResponse{}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := e.RegisterHook(HookBeforeProviderRequest, "fresh-request", func(_ context.Context, req HookRequest) (HookResponse, error) {
		if req.Payload["stage"] == "context" {
			requestContexts.Add(1)
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	var attempts atomic.Int32
	original := streamWithToolsWithHooks
	t.Cleanup(func() { streamWithToolsWithHooks = original })
	streamWithToolsWithHooks = func(callCtx context.Context, model string, request *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		n := attempts.Add(1)
		if n == 1 {
			if model != "fresh-test/old" {
				t.Errorf("first model %q", model)
			}
			// The failed closure must not be reused, even if an adapter mutated it.
			request.Messages = []goai.Message{goai.UserMessage("poisoned failed closure")}
			if err := s.AddMessage(callCtx, store.NowID("message"), "A", "user", "committed during failure", nil); err != nil {
				return nil, err
			}
			if err := s.TouchSessionState(callCtx, "A", map[string]any{"selected_model": "fresh-test/new", "model": "fresh-test/new", "thinking_model": "fresh-test/new", "thinking_level": "high"}); err != nil {
				return nil, err
			}
			if err := register("new definition"); err != nil {
				return nil, err
			}
			e.tools.Unregister("withdrawn_tool")
			if err := write(filepath.Join(root, "AGENTS.md"), "new committed instructions"); err != nil {
				return nil, err
			}
			if err := write(settings, `{"retry":{"maxRetries":3,"baseDelayMs":0,"maxAgentDelayMs":0},"transport":"sse"}`); err != nil {
				return nil, err
			}
			return &inference.StreamResult{}, errors.New("HTTP 503 service unavailable")
		}
		if model != "fresh-test/new" || hooks.Thinking != "high" || hooks.Transport != goai.Transport("sse") {
			t.Errorf("stale preparation: %q %+v", model, hooks)
		}
		raw, _ := json.Marshal(request)
		if !strings.Contains(string(raw), "committed during failure") || !strings.Contains(string(raw), "once-only lifecycle prefix") || strings.Contains(string(raw), "poisoned failed closure") || !strings.Contains(request.SystemPrompt, "new committed instructions") {
			t.Errorf("stale request %s", raw)
		}
		found := false
		for _, tool := range request.Tools {
			if tool.Name == "withdrawn_tool" {
				t.Error("withdrawn tool remained in retry declaration")
			}
			if tool.Name == "fresh_tool" && tool.Description == "new definition" {
				found = true
			}
		}
		if !found {
			t.Error("stale tool definitions")
		}
		if n == 2 {
			msg := goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "fresh-call", Name: "fresh_tool", Arguments: map[string]any{}}}}
			return &inference.StreamResult{Message: &msg}, nil
		}
		if n != 3 {
			t.Errorf("unexpected provider call %d", n)
		}
		msg := goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "fresh done"}}}
		return &inference.StreamResult{Message: &msg, Text: "fresh done"}, nil
	}
	run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "original prompt", Model: "fresh-test/old"})
	if err != nil {
		t.Fatal(err)
	}
	waitForCondition(t, 5*time.Second, func() bool { row, err := s.GetTurn(ctx, run.TurnID); return err == nil && row.Status == "completed" }, "fresh retry completion")
	if attempts.Load() != 3 || toolCalls.Load() != 1 || agentStarts.Load() != 1 || turnStarts.Load() != 2 || contexts.Load() != 3 || requestContexts.Load() != 3 {
		t.Fatal("replayed or omitted boundaries", attempts.Load(), toolCalls.Load(), agentStarts.Load(), turnStarts.Load(), contexts.Load(), requestContexts.Load())
	}
	events, err := s.ListTurnEvents(ctx, run.TurnID)
	if err != nil {
		t.Fatal(err)
	}
	started := 0
	for _, ev := range events {
		if ev.Type == "inference.started" {
			started++
			if started > 1 && ev.Payload["model"] != "fresh-test/new" {
				t.Fatal("stale durable model", ev.Payload)
			}
		}
	}
	if started != 3 {
		t.Fatal("missing fresh intents", started)
	}
	messages, err := s.ListMessages(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range messages {
		if m.Content == "fresh done" {
			found = true
			if m.Payload["model"] != "fresh-test/new" || m.Payload["thinking_level"] != "high" {
				t.Fatal("stale receipt preparation", m.Payload)
			}
		}
	}
	if !found {
		t.Fatal("final response absent")
	}
}

func TestProviderRetryReloadsPolicyAfterCommittedFailure(t *testing.T) {
	for _, tc := range []struct {
		name, updated string
		want          int
		beforeFailure bool
	}{
		{"disable", `{"retry":{"enabled":false,"baseDelayMs":0,"maxAgentDelayMs":0}}`, 1, false},
		{"disable_before_decision", `{"retry":{"enabled":false,"baseDelayMs":0,"maxAgentDelayMs":0}}`, 1, true},
		{"malformed", `{"retry":`, 1, false},
		{"malformed_before_decision", `{"retry":`, 1, true},
		{"zero_count", `{"retry":{"maxRetries":0,"baseDelayMs":0,"maxAgentDelayMs":0}}`, 1, false},
		{"increase_budget", `{"retry":{"maxRetries":3,"baseDelayMs":0,"maxAgentDelayMs":0}}`, 4, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
			path := filepath.Join(root, ".pi", "settings.json")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(`{"retry":{"maxRetries":1,"baseDelayMs":0,"maxAgentDelayMs":0}}`), 0600); err != nil {
				t.Fatal(err)
			}
			s := openTestStore(t)
			t.Cleanup(func() { s.Close() })
			e := NewWithRuntimeConfig(s, config.Load(root), "")
			t.Cleanup(func() { e.Close() })
			ctx := context.Background()
			if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			var calls, changed atomic.Int32
			if _, err := e.RegisterHook(HookAfterProviderResponse, "change-current-policy", func(ctx context.Context, req HookRequest) (HookResponse, error) {
				if changed.CompareAndSwap(0, 1) {
					row, err := s.GetTurn(ctx, req.TurnID)
					if err != nil {
						return HookResponse{}, err
					}
					if row.Phase != "retry_wait" {
						t.Errorf("failure was not committed before policy update: %s", row.Phase)
					}
					if err = os.WriteFile(path, []byte(tc.updated), 0600); err != nil {
						return HookResponse{}, err
					}
				}
				return HookResponse{}, nil
			}); err != nil {
				t.Fatal(err)
			}
			original := streamWithToolsWithHooks
			t.Cleanup(func() { streamWithToolsWithHooks = original })
			streamWithToolsWithHooks = func(context.Context, string, *goai.Context, func(map[string]any), *inference.StreamHooks) (*inference.StreamResult, error) {
				if calls.Add(1) < 4 {
					if tc.beforeFailure && changed.CompareAndSwap(0, 1) {
						if err := os.WriteFile(path, []byte(tc.updated), 0600); err != nil {
							return nil, err
						}
					}
					return &inference.StreamResult{}, errors.New("HTTP 503 service unavailable")
				}
				msg := goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "current policy"}}}
				return &inference.StreamResult{Message: &msg, Text: "current policy"}, nil
			}
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "policy", Model: "mock-stream"})
			if err != nil {
				t.Fatal(err)
			}
			wantStatus := "failed"
			if tc.want == 4 {
				wantStatus = "completed"
			}
			waitForCondition(t, 5*time.Second, func() bool { row, err := s.GetTurn(ctx, run.TurnID); return err == nil && row.Status == wantStatus }, "current policy settlement")
			if int(calls.Load()) != tc.want || changed.Load() != 1 {
				t.Fatal("stale retry policy", calls.Load(), changed.Load())
			}
			var intents int
			if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.started'`, run.TurnID).Scan(&intents); err != nil || intents != tc.want {
				t.Fatal("unexpected admitted effects", intents, err)
			}
			if tc.beforeFailure {
				var scheduled int
				if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.retry_scheduled'`, run.TurnID).Scan(&scheduled); err != nil || scheduled != 0 {
					t.Fatal("stale policy scheduled a retry", scheduled, err)
				}
			}
		})
	}
}

func TestProviderRetryOwnsLifecyclePreparationAndRebindsRegistrations(t *testing.T) {
	_, e := intentFixture(t)
	r := e.runner("A")
	tool := func(name, description string) tools.RegisteredTool {
		return tools.RegisteredTool{Name: name, Description: description, Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) { return "", nil }}
	}
	if err := e.RegisterTool(tool("withdrawn", "original")); err != nil {
		t.Fatal(err)
	}
	p := &providerPreparation{}
	args := map[string]any{"n": json.Number("9007199254740993")}
	resp := HookResponse{SystemPrompt: "once-only policy", Messages: []goai.Message{{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "toolCall", Arguments: args}}}}, Tools: []goai.Tool{{Name: "withdrawn"}, {Name: "newly_registered", Description: "old synthetic description"}}}
	if err := r.captureAgentStartPreparation(p, resp); err != nil {
		t.Fatal(err)
	}
	args["n"] = json.Number("1")
	e.tools.Unregister("withdrawn")
	if err := e.RegisterTool(tool("newly_registered", "current registered definition")); err != nil {
		t.Fatal(err)
	}
	conv := &goai.Context{}
	if err := r.applyAgentStartPreparation(p, conv); err != nil {
		t.Fatal(err)
	}
	if conv.SystemPrompt != "once-only policy" || len(conv.Tools) != 1 || conv.Tools[0].Name != "newly_registered" || conv.Tools[0].Description != "current registered definition" {
		t.Fatal(conv)
	}
	if n, ok := conv.Messages[0].Content[0].Arguments["n"].(json.Number); !ok || n.String() != "9007199254740993" {
		t.Fatal("lifecycle snapshot aliased or rounded", conv.Messages)
	}
	empty := &providerPreparation{}
	if err := r.captureAgentStartPreparation(empty, HookResponse{Messages: []goai.Message{}, Tools: []goai.Tool{}}); err != nil {
		t.Fatal(err)
	}
	if err := r.applyAgentStartPreparation(empty, conv); err != nil {
		t.Fatal(err)
	}
	if conv.Messages == nil || len(conv.Messages) != 0 || conv.Tools == nil || len(conv.Tools) != 0 {
		t.Fatal("empty overrides lost", conv)
	}
}
