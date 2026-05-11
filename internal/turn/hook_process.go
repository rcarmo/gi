package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/scripting"
)

const processHookProtocol = "hook-jsonrpc-v1"

type processHookRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type processHookRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type processHookRPCResponse struct {
	JSONRPC string               `json:"jsonrpc"`
	ID      int                  `json:"id"`
	Result  json.RawMessage      `json:"result,omitempty"`
	Error   *processHookRPCError `json:"error,omitempty"`
}

type processHookHelloResult struct {
	OK   bool   `json:"ok"`
	Name string `json:"name,omitempty"`
}

func isProcessHookSpec(spec scripting.EventHookSpec) bool {
	if strings.EqualFold(strings.TrimSpace(spec.Engine), "process") {
		return true
	}
	return strings.TrimSpace(spec.Command) != ""
}

func newProcessHookHandler(workspaceRoot string, spec scripting.EventHookSpec) HookHandler {
	return func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return invokeProcessHook(ctx, workspaceRoot, spec, req)
	}
}

func invokeProcessHook(ctx context.Context, workspaceRoot string, spec scripting.EventHookSpec, req HookRequest) (HookResponse, error) {
	transport := strings.TrimSpace(spec.Transport)
	if transport == "" {
		transport = "stdio"
	}
	if transport != "stdio" {
		return HookResponse{}, fmt.Errorf("process hook transport not supported: %s", transport)
	}
	protocol := strings.TrimSpace(spec.Protocol)
	if protocol == "" {
		protocol = processHookProtocol
	}
	if protocol != processHookProtocol {
		return HookResponse{}, fmt.Errorf("process hook protocol not supported: %s", protocol)
	}
	command, err := resolveProcessHookCommand(workspaceRoot, spec.Command)
	if err != nil {
		return HookResponse{}, err
	}
	cmd := exec.CommandContext(ctx, command, spec.Args...)
	cmd.Dir = resolveProcessHookDir(workspaceRoot, spec.CWD)
	cmd.Env = appendProcessHookEnv(os.Environ(), spec, req, workspaceRoot)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return HookResponse{}, fmt.Errorf("process hook stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return HookResponse{}, fmt.Errorf("process hook stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return HookResponse{}, fmt.Errorf("process hook stderr: %w", err)
	}
	stderrBytes := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(stderr)
		stderrBytes <- data
	}()
	if err := cmd.Start(); err != nil {
		return HookResponse{}, fmt.Errorf("start process hook: %w", err)
	}
	enc := json.NewEncoder(stdin)
	enc.SetEscapeHTML(false)
	dec := json.NewDecoder(stdout)
	helloReq := processHookRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "hook.hello",
		Params: map[string]any{
			"name":     firstNonEmpty(strings.TrimSpace(spec.Source), strings.TrimSpace(spec.Name), filepath.Base(command)),
			"version":  1,
			"modes":    processHookModes(req.Name),
			"protocol": processHookProtocol,
		},
	}
	if err := enc.Encode(helloReq); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("write process hook hello: %w", err)
	}
	var helloResp processHookRPCResponse
	if err := dec.Decode(&helloResp); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("read process hook hello: %w", err)
	}
	if helloResp.Error != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("process hook hello failed: %s", helloResp.Error.Message)
	}
	var helloResult processHookHelloResult
	if err := json.Unmarshal(helloResp.Result, &helloResult); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("decode process hook hello: %w", err)
	}
	if !helloResult.OK {
		_ = stdin.Close()
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("process hook hello rejected hook %s", req.Name)
	}
	invokeReq := processHookRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  processHookMethodName(req.Name),
		Params:  processHookPayload(req),
	}
	if err := enc.Encode(invokeReq); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("write process hook request: %w", err)
	}
	_ = stdin.Close()
	var invokeResp processHookRPCResponse
	if err := dec.Decode(&invokeResp); err != nil {
		_ = cmd.Wait()
		return HookResponse{}, fmt.Errorf("read process hook response: %w", err)
	}
	waitErr := cmd.Wait()
	stderrOut := strings.TrimSpace(string(<-stderrBytes))
	if invokeResp.Error != nil {
		if stderrOut != "" {
			return HookResponse{}, fmt.Errorf("process hook %s failed: %s (%s)", req.Name, invokeResp.Error.Message, stderrOut)
		}
		return HookResponse{}, fmt.Errorf("process hook %s failed: %s", req.Name, invokeResp.Error.Message)
	}
	if waitErr != nil {
		if stderrOut != "" {
			return HookResponse{}, fmt.Errorf("process hook %s exited: %v (%s)", req.Name, waitErr, stderrOut)
		}
		return HookResponse{}, fmt.Errorf("process hook %s exited: %w", req.Name, waitErr)
	}
	resp, err := hookResponseFromScript(strings.TrimSpace(string(invokeResp.Result)))
	if err != nil {
		return HookResponse{}, err
	}
	return resp, nil
}

func resolveProcessHookCommand(workspaceRoot, command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("process hook command is required")
	}
	if filepath.IsAbs(command) {
		return command, nil
	}
	if strings.ContainsRune(command, filepath.Separator) {
		return filepath.Join(workspaceRoot, command), nil
	}
	return command, nil
}

func resolveProcessHookDir(workspaceRoot, cwd string) string {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return workspaceRoot
	}
	if filepath.IsAbs(cwd) {
		return cwd
	}
	return filepath.Join(workspaceRoot, cwd)
}

func appendProcessHookEnv(base []string, spec scripting.EventHookSpec, req HookRequest, workspaceRoot string) []string {
	env := append([]string{}, base...)
	env = append(env,
		"GI_HOOK_NAME="+req.Name,
		"GI_SESSION_ID="+req.SessionID,
		"GI_TURN_ID="+req.TurnID,
		"GI_AGENT_ID="+req.AgentID,
		"GI_MODEL="+req.Model,
		"GI_WORKSPACE_ROOT="+workspaceRoot,
	)
	for k, v := range spec.Env {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		env = append(env, key+"="+v)
	}
	return env
}

func processHookMethodName(name string) string {
	switch normalizeHookName(name) {
	case HookBeforeProviderRequest:
		return "hook.before_llm"
	case HookAfterProviderResponse:
		return "hook.after_llm"
	case HookToolCall:
		return "hook.before_tool"
	case HookToolResult:
		return "hook.after_tool"
	default:
		return "hook." + normalizeHookName(name)
	}
}

func processHookModes(name string) []string {
	switch normalizeHookName(name) {
	case HookToolCall, HookToolResult:
		return []string{"tool"}
	case HookApproveTool:
		return []string{"approve"}
	case HookBeforeProviderRequest, HookAfterProviderResponse:
		return []string{"llm"}
	default:
		return []string{"observe"}
	}
}

func processHookPayload(req HookRequest) map[string]any {
	payload := hookScriptPayload(req)
	payload["hook_phase"] = req.Name
	payload["hook_method"] = processHookMethodName(req.Name)
	if req.ToolCall != nil {
		payload["tool"] = req.ToolCall.Name
		payload["arguments"] = req.ToolCall.Arguments
	}
	if strings.TrimSpace(req.ToolResult) != "" || req.ToolError {
		payload["result"] = map[string]any{
			"for_llm":  req.ToolResult,
			"is_error": req.ToolError,
		}
	}
	if req.Name == HookBeforeProviderRequest || req.Name == normalizeHookName("before_llm") {
		if req.Payload != nil {
			if request, ok := req.Payload["request"]; ok && request != nil {
				payload["request"] = request
			} else {
				payload["request"] = map[string]any{
					"model":         req.Model,
					"system_prompt": req.SystemPrompt,
					"messages":      req.Messages,
					"tools":         req.Tools,
				}
			}
		} else {
			payload["request"] = map[string]any{
				"model":         req.Model,
				"system_prompt": req.SystemPrompt,
				"messages":      req.Messages,
				"tools":         req.Tools,
			}
		}
	}
	return payload
}
