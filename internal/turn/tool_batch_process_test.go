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
	e := NewWithRuntimeConfig(s, config.Load(filepath.Dir(path)), "")
	defer e.Close()
	if _, err = s.CreateSession(ctx, "batch", "batch", nil); err != nil {
		t.Fatal(err)
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
		if err := appendBatchTestLog(path+".effects", "old:"+value); err != nil {
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
	withStreamWithToolsStub(t, func(context.Context, string, *goai.Context, func(map[string]any)) (*inference.StreamResult, error) {
		blocks := make([]goai.ContentBlock, 0, 3)
		for _, id := range []string{"one", "two", "three"} {
			blocks = append(blocks, goai.ContentBlock{Type: "toolCall", ID: id, Name: "batch-effect", Arguments: map[string]any{"value": id, "nonce": json.Number("9007199254740993")}})
		}
		return &inference.StreamResult{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonToolUse, Content: blocks, Usage: &goai.Usage{Input: 7}}}, nil
	})
	if _, err = e.SubmitPrompt(ctx, RunInput{SessionID: "batch", Prompt: "original prompt", Model: "mock-stream"}); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestToolBatchSurvivesKilledOwner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires process kill")
	}
	for _, tc := range []struct {
		name, stage                                 string
		oldSafe, currentSafe, schemaChanged, replay bool
	}{
		{"safe", "execute", true, true, false, true},
		{"preparation_interrupted", "prepare", true, true, false, false},
		{"default_unsafe", "execute", false, false, false, false},
		{"saved_unsafe", "execute", false, true, false, false},
		{"currently_unsafe", "execute", true, false, false, false},
		{"changed_schema_uses_saved_arguments", "execute", true, true, true, true},
		{"result_policy_attempted", "result", true, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
			child := exec.Command(binary, "-test.run=^TestToolBatchProcessHelper$")
			child.Env = append(os.Environ(), "GI_TOOL_BATCH_DB="+path, "GI_TOOL_BATCH_STAGE="+tc.stage, fmt.Sprintf("GI_TOOL_BATCH_SAFE=%t", tc.oldSafe))
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
			turnID := string(turnBytes)
			s, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if _, err = s.DB().Exec(`update session_active_turns set updated_at='2000-01-01T00:00:00Z'`); err != nil {
				t.Fatal(err)
			}
			var providers, executions, starts atomic.Int32
			withStreamWithToolsStub(t, func(_ context.Context, _ string, request *goai.Context, _ func(map[string]any)) (*inference.StreamResult, error) {
				providers.Add(1)
				raw, _ := json.Marshal(request)
				rawText := string(raw)
				for _, text := range []string{"original prompt", "once-only lifecycle prefix", "reviewed:done:prepared-one", "reviewed:new:prepared-three"} {
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
				if hasProgress != (!tc.replay && tc.stage != "prepare") {
					t.Errorf("committed partial output lost or adopted into a successful replay: %s", raw)
				}
				if tc.stage == "result" && strings.Contains(rawText, "done:prepared-two") {
					t.Errorf("unapproved returned result adopted: %s", raw)
				}
				second := results["two"]
				if tc.replay {
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
			time.Sleep(20 * time.Millisecond)
			if providers.Load() != 0 || executions.Load() != 0 {
				t.Fatal("dispatched before runtime-ready barrier")
			}
			if err = e.ResumeInterruptedWork(context.Background()); err != nil {
				t.Fatal(err)
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
			if err != nil || accounting.Iterations != 2 || int64(accounting.Usage.Input) != wantInput || accounting.Unknown != wantUnknown {
				t.Fatal("usage lost/duplicated", accounting, err)
			}
			effects, err := os.ReadFile(path + ".effects")
			if err != nil {
				t.Fatal(err)
			}
			expectedEffects := []string{"old:prepared-one", "new:prepared-three"}
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
