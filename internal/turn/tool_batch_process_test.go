package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

const batchTestSchema = `{"type":"object","properties":{"value":{"type":"string"},"nonce":{"type":"integer"}},"required":["value","nonce"],"additionalProperties":false}`

func appendBatchTestLog(path, text string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text + "\n")
	return err
}

func TestToolBatchProcessHelper(t *testing.T) {
	path := os.Getenv("GI_TOOL_BATCH_DB")
	if path == "" {
		t.Skip("subprocess helper")
	}
	ctx := context.Background()
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	recovering := os.Getenv("GI_TOOL_BATCH_RECOVER") == "true"
	if recovering {
		if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
			t.Fatal(err)
		}
	}
	e := NewWithRuntimeConfig(s, config.Load(filepath.Dir(path)), "")
	defer e.Close()
	if !recovering {
		if _, err = s.CreateSession(ctx, "batch", "batch", nil); err != nil {
			t.Fatal(err)
		}
	}
	stage := os.Getenv("GI_TOOL_BATCH_STAGE")
	ready := func(turn string) {
		if err := os.WriteFile(path+".ready", []byte(turn), 0600); err != nil {
			t.Fatal(err)
		}
		select {}
	}
	if err = e.RegisterTool(tools.RegisteredTool{Name: "batch-effect", Parameters: json.RawMessage(batchTestSchema), ReplaySafe: os.Getenv("GI_TOOL_BATCH_SAFE") == "true", Executor: func(_ context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
		value := fmt.Sprint(call.Arguments["value"])
		prefix := "old:"
		if recovering {
			prefix = "retry:"
		}
		if err := appendBatchTestLog(path+".effects", prefix+value); err != nil {
			return "", err
		}
		if value == "prepared-two" {
			if err := rt.OnOutput("old-progress:prepared-two"); err != nil {
				return "", err
			}
			rt.AddUsage(goai.Usage{Input: 11})
			if stage == "execute" {
				ready(rt.TurnID)
			}
		}
		return "done:" + value, nil
	}}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{HookAgentStart, HookToolCall, HookApproveTool, HookToolResult} {
		name := name
		if _, err = e.RegisterHook(name, "fixture", func(_ context.Context, req HookRequest) (HookResponse, error) {
			id := ""
			if req.ToolCall != nil {
				id = req.ToolCall.ID
			}
			if err := appendBatchTestLog(path+".hooks", name+":"+id); err != nil {
				return HookResponse{}, err
			}
			if name == HookApproveTool && id == "prepared-two" && stage == "prepare" {
				ready(req.TurnID)
			}
			if name == HookToolCall {
				call := *req.ToolCall
				call.ID = "prepared-" + call.ID
				call.Arguments["value"] = "prepared-" + fmt.Sprint(call.Arguments["value"])
				return HookResponse{ToolCall: &call}, nil
			}
			if name == HookToolResult {
				if id == "prepared-two" && stage == "result" {
					ready(req.TurnID)
				}
				result := "reviewed:" + req.ToolResult
				return HookResponse{ToolResult: &result}, nil
			}
			return HookResponse{}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = e.RegisterHook(HookBeforeAgentStart, "prefix", func(context.Context, HookRequest) (HookResponse, error) {
		return HookResponse{Message: "once-only lifecycle prefix"}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if recovering && stage == "batch_end" {
		if _, err = e.RegisterHook(HookToolExecutionEnd, "settled-checkpoint", func(_ context.Context, req HookRequest) (HookResponse, error) {
			ready(req.TurnID)
			return HookResponse{}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		blocks := make([]goai.ContentBlock, 0, 3)
		for _, id := range []string{"one", "two", "three"} {
			blocks = append(blocks, goai.ContentBlock{Type: "toolCall", ID: id, Name: "batch-effect", Arguments: map[string]any{"value": id, "nonce": json.Number("9007199254740993")}})
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: blocks, Usage: &goai.Usage{Input: 7}}}, nil
	})
	if recovering {
		err = e.ResumeInterruptedWork(ctx)
	} else {
		_, err = e.SubmitPrompt(ctx, RunInput{SessionID: "batch", Prompt: "original prompt", Model: "mock-stream"})
	}
	if err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestToolBatchSurvivesKilledOwner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires process kill")
	}
	for _, tc := range []struct {
		name, stage, secondStage                            string
		oldSafe, currentSafe, schemaChanged, replay, parked bool
	}{
		{name: "safe", stage: "execute", oldSafe: true, currentSafe: true, replay: true},
		{name: "preparation_interrupted", stage: "prepare", oldSafe: true, currentSafe: true},
		{name: "default_unsafe", stage: "execute"},
		{name: "saved_unsafe", stage: "execute", currentSafe: true},
		{name: "currently_unsafe", stage: "execute", oldSafe: true},
		{name: "changed_schema_uses_saved_arguments", stage: "execute", oldSafe: true, currentSafe: true, schemaChanged: true, replay: true},
		{name: "result_policy_attempted", stage: "result", oldSafe: true, currentSafe: true},
		{name: "retry_admitted_then_interrupted", stage: "execute", secondStage: "execute", oldSafe: true, currentSafe: true, replay: true},
		{name: "retry_admitted_then_denied", stage: "execute", secondStage: "execute", oldSafe: true},
		{name: "settled_batch_successor", stage: "execute", secondStage: "batch_end", oldSafe: true, currentSafe: true},
		{name: "parked_safe", stage: "execute", oldSafe: true, currentSafe: true, replay: true, parked: true},
		{name: "parked_unsafe", stage: "execute", parked: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			secondSettled := tc.secondStage == "batch_end"
			root := t.TempDir()
			path := filepath.Join(root, "batch.db")
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
			stages := []string{tc.stage}
			if tc.secondStage != "" {
				stages = append(stages, tc.secondStage)
			}
			var turnID string
			for i, stage := range stages {
				if i > 0 {
					if err = os.Remove(path + ".ready"); err != nil {
						t.Fatal(err)
					}
				}
				child := exec.Command(binary, "-test.run=^TestToolBatchProcessHelper$")
				child.Env = append(os.Environ(), "GI_TOOL_BATCH_DB="+path, "GI_TOOL_BATCH_STAGE="+stage, fmt.Sprintf("GI_TOOL_BATCH_SAFE=%t", tc.oldSafe), fmt.Sprintf("GI_TOOL_BATCH_RECOVER=%t", i > 0))
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
				waitForCondition(t, 8*time.Second, func() bool { _, err := os.Stat(path + ".ready"); return err == nil }, "admitted tool interruption")
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
				if turnID != "" && turnID != string(turnBytes) {
					t.Fatal("recovery changed turn identity")
				}
				turnID = string(turnBytes)
			}
			s, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
				t.Fatal(err)
			}
			if tc.parked {
				for i := 0; i < 5; i++ {
					if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
						t.Fatal(err)
					}
					claims, listErr := s.ListStaleActiveTurnClaims(context.Background(), time.Now(), "batch")
					if listErr != nil || len(claims) != 1 {
						t.Fatal(claims, listErr)
					}
					if cp, acquireErr := s.AcquireToolBatchRecovery(context.Background(), claims[0], fmt.Sprintf("pre-crash-%d", i), time.Now()); acquireErr != nil || cp == nil {
						t.Fatal(cp, acquireErr)
					}
				}
				if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
					t.Fatal(err)
				}
			}
			var providers, executions, starts atomic.Int32
			withStreamWithToolsStub(t, func(_ context.Context, _ string, request *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
				providers.Add(1)
				raw, _ := json.Marshal(request)
				rawText := string(raw)
				third := "reviewed:new:prepared-three"
				if secondSettled {
					third = "reviewed:done:prepared-three"
				}
				for _, text := range []string{"original prompt", "once-only lifecycle prefix", "reviewed:done:prepared-one", third} {
					if strings.Count(rawText, text) != 1 {
						t.Errorf("lost/duplicated %q: %s", text, raw)
					}
				}
				ids := map[string]int{}
				results := map[string]goai.Message{}
				for _, msg := range request.Messages {
					if msg.Role == goai.RoleToolResult {
						results[msg.ToolCallID] = msg
					}
					for _, block := range msg.Content {
						if block.Type == "toolCall" {
							ids[block.ID]++
						}
					}
				}
				for _, id := range []string{"one", "two", "three"} {
					if ids[id] != 1 {
						t.Errorf("wire identity %s: %v", id, ids)
					}
					if _, ok := results[id]; !ok {
						t.Errorf("missing wire result %s: %s", id, raw)
					}
				}
				hasProgress := strings.Contains(rawText, "old-progress:prepared-two")
				if hasProgress != (!tc.replay && !secondSettled && tc.stage != "prepare") {
					t.Errorf("committed partial output lost or adopted into a successful replay: %s", raw)
				}
				if tc.stage == "result" && strings.Contains(rawText, "done:prepared-two") {
					t.Errorf("unapproved returned result adopted: %s", raw)
				}
				second := results["two"]
				if secondSettled {
					if second.IsError || !strings.Contains(rawText, "reviewed:done:prepared-two") {
						t.Errorf("settled replay result missing: %s", raw)
					}
				} else if tc.replay {
					if second.IsError || !strings.Contains(rawText, "reviewed:new:prepared-two") {
						t.Errorf("safe result missing: %s", raw)
					}
				} else if !second.IsError || !strings.Contains(strings.ToLower(rawText), "verify") {
					t.Errorf("uncertainty not disclosed: %s", raw)
				}
				return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Content: []goai.ContentBlock{{Type: "text", Text: "continued batch"}}, Usage: &goai.Usage{Input: 13}}}, nil
			})
			e := NewWithRuntimeConfig(s, config.Load(root), "")
			defer e.Close()
			schema := batchTestSchema
			if tc.schemaChanged {
				schema = `{"type":"object","required":["fresh"],"properties":{"fresh":{"type":"boolean"},"value":{"type":"string"},"nonce":{"type":"integer"}},"additionalProperties":false}`
			}
			if err = e.RegisterTool(tools.RegisteredTool{Name: "batch-effect", Parameters: json.RawMessage(schema), ReplaySafe: tc.currentSafe, Executor: func(_ context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
				executions.Add(1)
				if secondSettled {
					t.Error("settled batch executed again")
				}
				value := fmt.Sprint(call.Arguments["value"])
				if value == "prepared-one" || value == "prepared-two" && !tc.replay {
					t.Error("settled/unsafe callback replayed", call)
				}
				if call.ID != "prepared-"+strings.TrimPrefix(value, "prepared-") || fmt.Sprint(call.Arguments["nonce"]) != "9007199254740993" {
					t.Error("prepared identity/number changed", call)
				}
				if value == "prepared-two" {
					if tc.schemaChanged {
						if _, exists := call.Arguments["fresh"]; exists {
							t.Error("replacement schema changed prepared arguments", call)
						}
					}
					rt.AddUsage(goai.Usage{Input: 17})
				}
				if err := appendBatchTestLog(path+".effects", "new:"+value); err != nil {
					return "", err
				}
				return "new:" + value, nil
			}}); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{HookBeforeAgentStart, HookAgentStart} {
				if _, err = e.RegisterHook(name, "no-replay", func(context.Context, HookRequest) (HookResponse, error) { starts.Add(1); return HookResponse{}, nil }); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{HookToolCall, HookApproveTool, HookToolResult} {
				name := name
				if _, err = e.RegisterHook(name, "fixture", func(_ context.Context, req HookRequest) (HookResponse, error) {
					if secondSettled {
						t.Error("settled batch policy repeated", name)
					}
					if name == HookToolResult {
						if req.ToolCall.ID == "prepared-one" || req.ToolCall.ID == "prepared-two" && !tc.replay {
							t.Error("result policy replayed", req.ToolCall.ID)
						}
						result := "reviewed:" + req.ToolResult
						return HookResponse{ToolResult: &result}, nil
					}
					if req.ToolCall.ID != "three" && req.ToolCall.ID != "prepared-three" {
						t.Error("preparation replayed", name, req.ToolCall.ID)
					}
					if name == HookToolCall {
						call := *req.ToolCall
						call.ID = "prepared-three"
						call.Arguments["value"] = "prepared-three"
						if tc.schemaChanged {
							call.Arguments["fresh"] = true
						}
						return HookResponse{ToolCall: &call}, nil
					}
					return HookResponse{}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			if secondSettled {
				if _, err = e.RegisterHook(HookToolExecutionEnd, "no-repeat", func(context.Context, HookRequest) (HookResponse, error) {
					t.Error("execution-end policy repeated")
					return HookResponse{}, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			time.Sleep(20 * time.Millisecond)
			if providers.Load() != 0 || executions.Load() != 0 {
				t.Fatal("dispatched before runtime-ready barrier")
			}
			if err = e.ResumeInterruptedWork(context.Background()); err != nil {
				t.Fatal(err)
			}
			if tc.parked {
				row, getErr := s.GetTurn(context.Background(), turnID)
				if getErr != nil || row.Status != "failed" || row.Phase != "held_for_retry_or_skip" || providers.Load() != 0 || executions.Load() != 0 {
					t.Fatal("exhausted checkpoint not parked", row, getErr)
				}
				admitted, retryErr := e.RetryHeldTurn(context.Background(), turnID, "operator retry")
				if retryErr != nil || admitted.TurnID != turnID || !admitted.Queued {
					t.Fatal("checkpoint retry changed invocation", admitted, retryErr)
				}
				duplicate, retryErr := e.RetryHeldTurn(context.Background(), turnID, "duplicate")
				if retryErr != nil || duplicate.TurnID != turnID {
					t.Fatal("checkpoint retry not idempotent", duplicate, retryErr)
				}
			}
			waitForCondition(t, 5*time.Second, func() bool {
				row, err := s.GetTurn(context.Background(), turnID)
				return err == nil && row.Status == "completed"
			}, "tool batch continuation")
			e.runs.Wait()
			wantExecutions := int32(1)
			if tc.replay {
				wantExecutions++
			}
			if secondSettled {
				wantExecutions = 0
			}
			if providers.Load() != 1 || executions.Load() != wantExecutions || starts.Load() != 0 {
				t.Fatal("replayed/skipped work", providers.Load(), executions.Load(), starts.Load())
			}
			accounting, err := s.TurnAccounting(context.Background(), turnID)
			wantInput := int64(31)
			wantUnknown := 1
			if tc.stage == "result" {
				wantUnknown = 0
			}
			if tc.stage == "prepare" {
				wantInput, wantUnknown = 20, 0
			}
			if tc.replay {
				wantInput += 17
			}
			if tc.secondStage != "" {
				wantInput += 11
				if !secondSettled {
					wantUnknown++
				}
			}
			if err != nil || accounting.Iterations != 2 || int64(accounting.Usage.Input) != wantInput || accounting.Unknown != wantUnknown {
				t.Fatal("usage lost/duplicated", accounting, err)
			}
			effects, err := os.ReadFile(path + ".effects")
			if err != nil {
				t.Fatal(err)
			}
			thirdEffect := "new:prepared-three"
			if secondSettled {
				thirdEffect = "retry:prepared-three"
			}
			expectedEffects := []string{"old:prepared-one", thirdEffect}
			if tc.secondStage != "" {
				expectedEffects = append(expectedEffects, "retry:prepared-two")
			}
			if tc.replay {
				expectedEffects = append(expectedEffects, "new:prepared-two")
			}
			if tc.stage != "prepare" {
				expectedEffects = append(expectedEffects, "old:prepared-two")
			} else if strings.Contains(string(effects), "old:prepared-two") {
				t.Fatal("executor ran before preparation completed", string(effects))
			}
			for _, text := range expectedEffects {
				if strings.Count(string(effects), text) != 1 {
					t.Error("effect count changed", text, string(effects))
				}
			}
		})
	}
}
