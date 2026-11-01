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

func TestProviderRetryRecoveryFreshContinuationWithoutLifecycleOrToolReplay(t *testing.T) {
	for _, mode := range []string{"success", "exhausted", "disabled", "legacy", "invalid", "cancelled", "failure-loop", "woken"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			root := t.TempDir()
			t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
			if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
				t.Fatal(err)
			}
			settings := filepath.Join(root, ".pi", "settings.json")
			if err := os.WriteFile(settings, []byte(`{"defaultModel":"mock-stream","retry":{"maxRetries":3,"baseDelayMs":0,"maxAgentDelayMs":0}}`), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("old instructions"), 0600); err != nil {
				t.Fatal(err)
			}
			s := openChildJoinStore(t)
			if _, err := s.CreateSession(ctx, "recovery", "recovery", map[string]any{"selected_model": "mock-stream", "model": "mock-stream"}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateTurnWithStatus(ctx, "interrupted", "recovery", "running", "original user prompt", map[string]any{"model": "mock-stream"}); err != nil {
				t.Fatal(err)
			}
			if ok, err := s.ClaimSessionActiveTurn(ctx, "recovery", "interrupted", "dead", "old"); err != nil || !ok {
				t.Fatal(ok, err)
			}
			writer := store.WithTurnWriter(ctx, "recovery", "interrupted", "old")
			if err := s.AddMessage(ctx, "original-user", "recovery", "user", "original user prompt", map[string]any{"turn_id": "interrupted"}); err != nil {
				t.Fatal(err)
			}
			callMessage := goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "completed-call", Name: "completed_effect", Arguments: map[string]any{}}}}
			if err := s.AddMessage(ctx, "old-call", "recovery", "assistant", "old call", map[string]any{"turn_id": "interrupted", "provider_message": callMessage}); err != nil {
				t.Fatal(err)
			}
			if err := s.SettleToolResult(writer, "interrupted", "recovery", "tool.finished", map[string]any{"occurrence_id": "completed-effect"}, "old-result", "already did effect", map[string]any{"tool_call_id": "completed-call", "tool_name": "completed_effect", "turn_id": "interrupted"}); err != nil {
				t.Fatal(err)
			}
			if err := s.AdmitTurnEffect(writer, "recovery", "interrupted", "running", "inference.started", map[string]any{"request_id": "failed-request"}); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(300 * time.Millisecond)
			if mode == "cancelled" {
				deadline = time.Now().Add(time.Hour)
			}
			if mode == "woken" {
				deadline = time.Now().Add(-time.Second)
			}
			wait := &store.ProviderRetryWait{Attempt: 2, MaxAttempts: 3, Iteration: 3, Model: "mock-stream", RetryAt: deadline, FailureCategory: "network", Lifecycle: &store.ProviderLifecyclePolicy{Version: 1, Message: "once-only prefix", Tools: json.RawMessage(`[{"name":"completed_effect","description":"old definition","parameters":{"type":"object","properties":{}}}]`), RegisteredTools: map[string]bool{"completed_effect": true}}}
			if mode == "legacy" {
				wait.Lifecycle = nil
			}
			if mode == "failure-loop" {
				wait.RepeatedToolFailures = 3
				wait.LastToolFailureSignature = "completed_effect||same"
			}
			if _, err := s.SettleProviderAttempt(writer, "interrupted", "recovery", "failed-request", store.ProviderReceipt{Decision: "error_boundary", Error: "HTTP 503", Usage: &goai.Usage{Input: 5}, Iteration: 3, RetryWait: wait}); err != nil {
				t.Fatal(err)
			}
			if mode == "invalid" {
				if _, err := s.DB().Exec(`update turn_receipts set decision_json=json_set(decision_json,'$.retry_wait.lifecycle.version',9) where occurrence_id='failed-request'`); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "woken" {
				if err := s.ResumeProviderRetry(writer, "recovery", "interrupted", "failed-request"); err != nil {
					t.Fatal(err)
				}
			}
			staleRecoveryClaim(t, s)
			var requests, lifecycle, toolCalls, setups, ends atomic.Int32
			withStreamWithToolsStub(t, func(_ context.Context, _ string, request *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
				requests.Add(1)
				if time.Now().Before(deadline) {
					t.Error("recovery bypassed absolute deadline")
				}
				prefix, committed, oldResult := false, false, false
				for _, m := range request.Messages {
					for _, block := range m.Content {
						prefix = prefix || block.Text == "once-only prefix"
						committed = committed || block.Text == "committed while stopped"
						oldResult = oldResult || block.Text == "already did effect"
					}
				}
				if !prefix || !committed || !oldResult || !strings.Contains(request.SystemPrompt, "current instructions") {
					t.Errorf("not freshly rebuilt: prefix=%v committed=%v result=%v", prefix, committed, oldResult)
				}
				if len(request.Tools) != 1 || request.Tools[0].Description != "current definition" {
					t.Errorf("tool definition not rebound: %+v", request.Tools)
				}
				if mode == "exhausted" {
					return nil, errors.New("HTTP 503 unavailable")
				}
				if mode == "failure-loop" {
					return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: []goai.ContentBlock{{Type: "toolCall", ID: "fresh-call", Name: "completed_effect", Arguments: map[string]any{}}}}}, nil
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "resumed answer"}}, Usage: &goai.Usage{Output: 2}}}, nil
			})
			e := NewWithRuntimeConfig(s, config.Load(root), "")
			defer e.Close()
			if requests.Load() != 0 {
				t.Fatal("constructor dispatched before runtime-ready barrier")
			}
			e.beforeSetupErrorHook = func(context.Context, string, string) error { setups.Add(1); return errors.New("must not rerun setup") }
			if err := e.RegisterTool(tools.RegisteredTool{Name: "completed_effect", Description: "current definition", Parameters: json.RawMessage(`{"type":"object","properties":{}}`), Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
				toolCalls.Add(1)
				if mode == "failure-loop" {
					return "", errors.New("same")
				}
				return "replayed", nil
			}}); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{HookBeforeAgentStart, HookAgentStart, HookTurnStart} {
				if _, err := e.RegisterHook(name, "no-replay", func(context.Context, HookRequest) (HookResponse, error) { lifecycle.Add(1); return HookResponse{}, nil }); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := e.RegisterHook(HookAgentEnd, "end", func(context.Context, HookRequest) (HookResponse, error) { ends.Add(1); return HookResponse{}, nil }); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("current instructions"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := s.AddMessage(ctx, "later-user", "recovery", "user", "committed while stopped", nil); err != nil {
				t.Fatal(err)
			}
			if mode == "disabled" {
				if err := os.WriteFile(settings, []byte(`{"defaultModel":"mock-stream","retry":{"enabled":false}}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := e.ResumeInterruptedWork(ctx); err != nil {
				t.Fatal(err)
			}
			want := "completed"
			if mode != "success" && mode != "woken" {
				want = "failed"
			}
			if mode == "cancelled" {
				want = "cancelled"
				if err := e.CancelTurn(ctx, "recovery", "interrupted"); err != nil {
					t.Fatal(err)
				}
			}
			waitForCondition(t, 3*time.Second, func() bool { row, err := s.GetTurn(ctx, "interrupted"); return err == nil && row.Status == want }, "retry checkpoint continuation")
			if err := e.Close(); err != nil {
				t.Fatal(err)
			}
			expected := int32(1)
			if mode == "exhausted" {
				expected = 2
			}
			if mode == "legacy" || mode == "invalid" || mode == "disabled" || mode == "cancelled" {
				expected = 0
			}
			wantToolCalls := int32(0)
			if mode == "failure-loop" {
				wantToolCalls = 1
			}
			if requests.Load() != expected || toolCalls.Load() != wantToolCalls || setups.Load() != 0 || lifecycle.Load() != 0 {
				t.Fatal("replayed work or reset retry budget", requests.Load(), toolCalls.Load(), setups.Load(), lifecycle.Load())
			}
			wantEnds := int32(1)
			if mode == "legacy" || mode == "invalid" {
				wantEnds = 0
			}
			if ends.Load() != wantEnds {
				t.Fatal("lost or replayed agent end", ends.Load(), wantEnds)
			}
			if mode == "woken" {
				var wakes int
				if err := s.DB().QueryRow(`select count(*) from turn_events where turn_id='interrupted' and event_type='inference.retry_resumed'`).Scan(&wakes); err != nil || wakes != 1 {
					t.Fatal("wake replayed", wakes, err)
				}
			}
			var originalUsers int
			if err := s.DB().QueryRow(`select count(*) from messages where session_id='recovery' and role='user' and content='original user prompt'`).Scan(&originalUsers); err != nil || originalUsers != 1 {
				t.Fatal("original prompt replayed", originalUsers, err)
			}
			if mode != "legacy" && mode != "invalid" {
				if err := s.AdmitTurnEffect(writer, "recovery", "interrupted", "running", "inference.started", map[string]any{"request_id": "late-old"}); err == nil {
					t.Fatal("old writer revived after recovered cleanup")
				}
			}
			accounting, err := s.TurnAccounting(ctx, "interrupted")
			if err != nil || accounting.Usage.Input != 5 || accounting.Iterations != 3 {
				t.Fatal("iteration/accounting reset", accounting, err)
			}
		})
	}
}

func TestProviderRetryRecoveredAttemptBudget(t *testing.T) {
	var requests, waits int
	_, err := retryProviderRequestDynamicFrom(context.Background(), func() config.ProviderRetryPolicy { return config.ProviderRetryPolicy{Enabled: true, MaxRetries: 3} }, 2, func() (*inference.StreamResult, error, bool) { requests++; return nil, errors.New("HTTP 503"), false }, func(attempt int, _ time.Duration, _ string) (time.Time, error) {
		waits++
		if attempt != 3 {
			t.Error("attempt reset", attempt)
		}
		return time.Now(), nil
	}, func() error { return nil })
	if err == nil || requests != 2 || waits != 1 {
		t.Fatal("recovered attempt budget changed", requests, waits, err)
	}
}

func TestProviderRetryRestoreDoesNotReadSettingsBeforeDeadline(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "agent"))
	if err := os.MkdirAll(filepath.Join(root, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(root, ".pi", "settings.json")
	if err := os.WriteFile(settings, []byte(`{"defaultModel":"mock-stream"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s := openChildJoinStore(t)
	e := NewWithRuntimeConfig(s, config.Load(root), "")
	defer e.Close()
	if err := os.WriteFile(settings, []byte(`{"retry":`), 0600); err != nil {
		t.Fatal(err)
	}
	wait := &store.ProviderRetryWait{Attempt: 1, Iteration: 1, Model: "mock-stream", RetryAt: time.Now().Add(time.Hour), Lifecycle: &store.ProviderLifecyclePolicy{Version: 1}}
	preparation, err := e.runner("recovery").restoreProviderPreparation(wait)
	if err != nil || preparation == nil {
		t.Fatal("settings ended restored wait before its deadline", err)
	}
}
