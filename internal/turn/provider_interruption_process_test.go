package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func TestProviderInterruptionProcessHelper(t *testing.T) {
	path := os.Getenv("GI_PROVIDER_INTERRUPTION_DB")
	if path == "" {
		t.Skip("subprocess helper")
	}
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.CreateSession(context.Background(), "interrupted", "interrupted", nil); err != nil {
		t.Fatal(err)
	}
	e := NewWithRuntimeConfig(s, config.Load(filepath.Dir(path)), "")
	defer e.Close()
	if err = e.RegisterTool(tools.RegisteredTool{Name: "must-not-run", Description: "old definition", Parameters: json.RawMessage(`{"type":"object"}`), Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
		t.Error("partial tool prefix executed")
		return "forbidden", nil
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err = e.RegisterHook(HookBeforeAgentStart, "fixture", func(context.Context, HookRequest) (HookResponse, error) {
		return HookResponse{Message: "once-only accepted policy"}, nil
	}); err != nil {
		t.Fatal(err)
	}
	withStreamWithToolsHookStub(t, func(_ context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
		if _, err := hooks.OnPayload(map[string]any{"messages": "old prepared payload must not be sent again"}, nil); err != nil {
			t.Fatal(err)
		}
		if os.Getenv("GI_PROVIDER_INTERRUPTION_PARTIAL") == "1" {
			partial := &goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: "committed interrupted prefix"}, {Type: "thinking", Thinking: "interrupted private thought", ThinkingSignature: "retained-signature"}, {Type: "toolCall", ID: "not-authorised", Name: "must-not-run", Arguments: map[string]any{"value": "unadmitted argument"}}}, Usage: &goai.Usage{Input: 17}}
			if err := hooks.OnProgress(partial); err != nil {
				t.Fatal(err)
			}
			if _, err := s.DB().Exec(`update messages set content='tampered UI draft' where role='assistant'`); err != nil {
				t.Fatal(err)
			}
		}
		var turnID, requestID string
		if err := s.DB().QueryRow(`select turn_id,json_extract(payload_json,'$.request_id') from turn_events where event_type='inference.started' order by seq desc limit 1`).Scan(&turnID, &requestID); err != nil {
			t.Fatal(err)
		}
		ready, _ := json.Marshal(map[string]string{"turn": turnID, "request": requestID})
		if err := os.WriteFile(path+".ready", ready, 0600); err != nil {
			t.Fatal(err)
		}
		select {}
	})
	if _, err = e.SubmitPrompt(context.Background(), RunInput{SessionID: "interrupted", Prompt: "original interrupted prompt", Model: "old/model"}); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestProviderInterruptionSurvivesKilledOwnerWithFreshRequest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires process kill")
	}
	for _, partial := range []bool{false, true} {
		t.Run(map[bool]string{false: "unknown-without-output", true: "committed-partial"}[partial], func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "interrupted.db")
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
			child := exec.Command(binary, "-test.run=^TestProviderInterruptionProcessHelper$")
			flag := "0"
			if partial {
				flag = "1"
			}
			child.Env = append(os.Environ(), "GI_PROVIDER_INTERRUPTION_DB="+path, "GI_PROVIDER_INTERRUPTION_PARTIAL="+flag)
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
			waitForCondition(t, 8*time.Second, func() bool { _, err := os.Stat(path + ".ready"); return err == nil }, "admitted provider request before process kill")
			if err = child.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			err = child.Wait()
			joined = true
			if err == nil {
				t.Fatal("helper did not die", output.String())
			}
			var ready map[string]string
			data, err := os.ReadFile(path + ".ready")
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(data, &ready); err != nil {
				t.Fatal(err)
			}
			s, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
				t.Fatal(err)
			}
			if err = s.TouchSessionState(context.Background(), "interrupted", map[string]any{"selected_provider": "current", "selected_model": "new-model"}); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("fresh instructions after provider interruption"), 0600); err != nil {
				t.Fatal(err)
			}
			var requests, toolsRun, starts, requestHooks atomic.Int32
			withStreamWithToolsHookStub(t, func(_ context.Context, model string, request *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				requests.Add(1)
				if model != "current/new-model" {
					t.Errorf("stale model %q", model)
				}
				raw, _ := json.Marshal(request)
				for _, want := range []string{"fresh instructions after provider interruption", "once-only accepted policy", "current safe definition"} {
					if !strings.Contains(string(raw), want) {
						t.Errorf("fresh request lacks %q: %s", want, raw)
					}
				}
				for _, forbidden := range []string{"old prepared payload", "committed interrupted prefix", "interrupted private thought", "unadmitted argument", "tampered UI draft"} {
					if strings.Contains(string(raw), forbidden) {
						t.Errorf("unadopted/saved request replayed: %s", raw)
					}
				}
				if strings.Count(string(raw), "original interrupted prompt") != 1 || strings.Count(string(raw), "once-only accepted policy") != 1 {
					t.Error("prompt/policy replayed", string(raw))
				}
				if _, err := hooks.OnPayload(map[string]any{"messages": "fresh provider payload"}, nil); err != nil {
					return nil, err
				}
				var id string
				if err := s.DB().QueryRow(`select json_extract(payload_json,'$.request_id') from turn_events where event_type='inference.started' order by seq desc limit 1`).Scan(&id); err != nil {
					t.Fatal(err)
				}
				if id == ready["request"] {
					t.Error("old request occurrence reused")
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "fresh completed response"}}, Usage: &goai.Usage{Input: 11, Output: 3}}}, nil
			})
			e := NewWithRuntimeConfig(s, config.Load(root), "")
			defer e.Close()
			if err = e.RegisterTool(tools.RegisteredTool{Name: "must-not-run", Description: "current safe definition", Parameters: json.RawMessage(`{"type":"object"}`), ReplaySafe: true, Executor: func(context.Context, tools.ToolRuntime, goai.ToolCall) (string, error) {
				toolsRun.Add(1)
				return "forbidden", nil
			}}); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{HookBeforeAgentStart, HookAgentStart, HookToolCall, HookApproveTool} {
				if _, err = e.RegisterHook(name, "fixture", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = e.RegisterHook(HookBeforeProviderRequest, "fixture", func(context.Context, HookRequest) (HookResponse, error) {
				requestHooks.Add(1)
				return HookResponse{}, nil
			}); err != nil {
				t.Fatal(err)
			}
			time.Sleep(25 * time.Millisecond)
			if requests.Load() != 0 {
				t.Fatal("request dispatched before runtime ready")
			}
			if err = e.ResumeInterruptedWork(context.Background()); err != nil {
				t.Fatal(err)
			}
			waitForCondition(t, 5*time.Second, func() bool {
				row, err := s.GetTurn(context.Background(), ready["turn"])
				return err == nil && row.Status == "completed"
			}, "fresh provider generation after interrupted outcome")
			e.runs.Wait()
			if requests.Load() != 1 || toolsRun.Load() != 0 || starts.Load() != 0 || requestHooks.Load() != 2 {
				t.Fatal("work replayed or fresh hook skipped", requests.Load(), toolsRun.Load(), starts.Load(), requestHooks.Load())
			}
			accounting, err := s.TurnAccounting(context.Background(), ready["turn"])
			want := 11
			if partial {
				want += 17
			}
			if err != nil || accounting.Unknown != 1 || accounting.Usage.Input != want || accounting.Iterations != 1 {
				t.Fatal("interrupted accounting lost or double counted", accounting, err)
			}
			messages, err := s.ListMessages(context.Background(), "interrupted")
			if err != nil {
				t.Fatal(err)
			}
			prompts, prefixes, finals := 0, 0, 0
			for _, m := range messages {
				if m.Role == "user" && m.Content == "original interrupted prompt" {
					prompts++
				}
				if m.ID == store.ProviderMessageID(ready["request"]) {
					prefixes++
					if m.Content != "committed interrupted prefix" {
						t.Error("prefix lost or UI draft trusted", m.Content)
					}
				}
				if m.Role == "assistant" && m.Content == "fresh completed response" {
					finals++
				}
			}
			expectedPrefixes := 0
			if partial {
				expectedPrefixes = 1
			}
			if prompts != 1 || prefixes != expectedPrefixes || finals != 1 {
				t.Fatal("history duplicated/lost", prompts, prefixes, finals)
			}
		})
	}
}

func TestProviderInterruptionObserverCannotStealLiveHost(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var start, finish sync.Once
	unblock := func() { finish.Do(func() { close(release) }) }
	var requests atomic.Int32
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		requests.Add(1)
		start.Do(func() { close(entered) })
		<-release
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "live owner completed"}}, Usage: &goai.Usage{Input: 1}}}, nil
	})
	owner := New(s)
	defer owner.Close()
	defer unblock()
	run, err := owner.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "live owner", Model: "mock-stream"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("provider did not enter")
	}
	_, token, err := s.GetSessionActiveTurn(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z' where session_id='A'`); err != nil {
		t.Fatal(err)
	}
	observer := New(s)
	defer func() { unblock(); observer.Close() }()
	err = observer.ResumeInterruptedWork(ctx)
	if err != nil && !errors.Is(err, store.ErrQueueConflict) {
		t.Fatal(err)
	}
	closed := make(chan struct{})
	go func() { observer.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		unblock()
		<-closed
		t.Fatal("observer acquired live provider ownership")
	}
	_, after, err := s.GetSessionActiveTurn(ctx, "A")
	if err != nil || after != token {
		t.Fatal("live owner fenced", after, token, err)
	}
	var interrupted int
	if err = s.DB().QueryRow(`select count(*) from turn_events where turn_id=? and event_type='inference.interrupted'`, run.TurnID).Scan(&interrupted); err != nil || interrupted != 0 {
		t.Fatal("live request sealed", interrupted, err)
	}
	unblock()
	owner.runs.Wait()
	row, err := s.GetTurn(ctx, run.TurnID)
	if err != nil || row.Status != "completed" || requests.Load() != 1 {
		t.Fatal("observer closed/replayed live work", row, requests.Load(), err)
	}
}
