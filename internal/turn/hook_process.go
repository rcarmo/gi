package turn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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

type mountedProcessHook struct {
	workspaceRoot string
	spec          scripting.EventHookSpec

	mu        sync.Mutex
	started   bool
	nextID    int
	stderrGen uint64
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	enc       *json.Encoder
	dec       *json.Decoder
	stderrMu  sync.Mutex
	stderr    bytes.Buffer
}

func newProcessHookHandler(workspaceRoot string, spec scripting.EventHookSpec) HookHandler {
	mounted := &mountedProcessHook{workspaceRoot: workspaceRoot, spec: spec}
	return func(ctx context.Context, req HookRequest) (HookResponse, error) {
		return mounted.invoke(ctx, req)
	}
}

func invokeProcessHook(ctx context.Context, workspaceRoot string, spec scripting.EventHookSpec, req HookRequest) (HookResponse, error) {
	mounted := &mountedProcessHook{workspaceRoot: workspaceRoot, spec: spec}
	return mounted.invoke(ctx, req)
}

func (m *mountedProcessHook) invoke(ctx context.Context, req HookRequest) (HookResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ensureStartedLocked(ctx, req); err != nil {
		return HookResponse{}, err
	}
	m.nextID++
	invokeReq := processHookRPCRequest{
		JSONRPC: "2.0",
		ID:      m.nextID,
		Method:  processHookMethodName(req.Name),
		Params:  processHookPayload(req),
	}
	if err := m.enc.Encode(invokeReq); err != nil {
		stderrOut := m.stderrStringLocked()
		m.closeLocked()
		if stderrOut != "" {
			return HookResponse{}, fmt.Errorf("write process hook request: %w (%s)", err, stderrOut)
		}
		return HookResponse{}, fmt.Errorf("write process hook request: %w", err)
	}
	invokeResp, err := m.decodeResponseLocked(ctx)
	if err != nil {
		stderrOut := m.stderrStringLocked()
		m.closeLocked()
		if stderrOut != "" {
			return HookResponse{}, fmt.Errorf("read process hook response: %w (%s)", err, stderrOut)
		}
		return HookResponse{}, fmt.Errorf("read process hook response: %w", err)
	}
	if invokeResp.Error != nil {
		return HookResponse{}, fmt.Errorf("process hook %s failed: %s", req.Name, invokeResp.Error.Message)
	}
	resp, err := hookResponseFromScript(strings.TrimSpace(string(invokeResp.Result)))
	if err != nil {
		return HookResponse{}, err
	}
	return resp, nil
}

func (m *mountedProcessHook) ensureStartedLocked(ctx context.Context, req HookRequest) error {
	transport := strings.TrimSpace(m.spec.Transport)
	if transport == "" {
		transport = "stdio"
	}
	if transport != "stdio" {
		return fmt.Errorf("process hook transport not supported: %s", transport)
	}
	protocol := strings.TrimSpace(m.spec.Protocol)
	if protocol == "" {
		protocol = processHookProtocol
	}
	if protocol != processHookProtocol {
		return fmt.Errorf("process hook protocol not supported: %s", protocol)
	}
	if m.started && m.cmd != nil && m.cmd.Process != nil {
		return nil
	}
	command, err := resolveProcessHookCommand(m.workspaceRoot, m.spec.Command)
	if err != nil {
		return err
	}
	cmd := exec.Command(command, m.spec.Args...)
	cmd.Dir = resolveProcessHookDir(m.workspaceRoot, m.spec.CWD)
	cmd.Env = appendProcessHookEnv(os.Environ(), m.spec, req, m.workspaceRoot)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("process hook stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("process hook stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("process hook stderr: %w", err)
	}
	m.stderrMu.Lock()
	m.stderr.Reset()
	m.stderrGen++
	stderrGen := m.stderrGen
	m.stderrMu.Unlock()
	go func(gen uint64) {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderr)
		m.stderrMu.Lock()
		defer m.stderrMu.Unlock()
		if gen != m.stderrGen {
			return
		}
		_, _ = m.stderr.Write(buf.Bytes())
	}(stderrGen)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start process hook: %w", err)
	}
	m.cmd = cmd
	m.stdin = stdin
	m.stdout = stdout
	m.enc = json.NewEncoder(stdin)
	m.enc.SetEscapeHTML(false)
	m.dec = json.NewDecoder(stdout)
	m.started = true
	m.nextID = 1
	helloReq := processHookRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "hook.hello",
		Params: map[string]any{
			"name":     firstNonEmpty(strings.TrimSpace(m.spec.Source), strings.TrimSpace(m.spec.Name), filepath.Base(command)),
			"version":  1,
			"modes":    processHookModes(req.Name),
			"protocol": processHookProtocol,
		},
	}
	if err := m.enc.Encode(helloReq); err != nil {
		m.closeLocked()
		return fmt.Errorf("write process hook hello: %w", err)
	}
	helloResp, err := m.decodeResponseLocked(ctx)
	if err != nil {
		m.closeLocked()
		return fmt.Errorf("read process hook hello: %w", err)
	}
	if helloResp.Error != nil {
		m.closeLocked()
		return fmt.Errorf("process hook hello failed: %s", helloResp.Error.Message)
	}
	var helloResult processHookHelloResult
	if err := json.Unmarshal(helloResp.Result, &helloResult); err != nil {
		m.closeLocked()
		return fmt.Errorf("decode process hook hello: %w", err)
	}
	if !helloResult.OK {
		m.closeLocked()
		return fmt.Errorf("process hook hello rejected hook %s", req.Name)
	}
	return nil
}

func (m *mountedProcessHook) decodeResponseLocked(ctx context.Context) (processHookRPCResponse, error) {
	type decoded struct {
		resp processHookRPCResponse
		err  error
	}
	dec := m.dec
	if dec == nil {
		return processHookRPCResponse{}, fmt.Errorf("process hook decoder not initialized")
	}
	ch := make(chan decoded, 1)
	go func(decoder *json.Decoder) {
		var resp processHookRPCResponse
		err := decoder.Decode(&resp)
		ch <- decoded{resp: resp, err: err}
	}(dec)
	select {
	case <-ctx.Done():
		m.closeLocked()
		return processHookRPCResponse{}, ctx.Err()
	case res := <-ch:
		return res.resp, res.err
	}
}

func (m *mountedProcessHook) stderrStringLocked() string {
	m.stderrMu.Lock()
	defer m.stderrMu.Unlock()
	return strings.TrimSpace(m.stderr.String())
}

func (m *mountedProcessHook) closeLocked() {
	m.stderrMu.Lock()
	m.stderrGen++
	m.stderrMu.Unlock()
	if m.stdin != nil {
		_ = m.stdin.Close()
	}
	if m.stdout != nil {
		_ = m.stdout.Close()
	}
	if m.cmd != nil {
		if m.cmd.Process != nil {
			_ = m.cmd.Process.Kill()
		}
		_ = m.cmd.Wait()
	}
	m.started = false
	m.cmd = nil
	m.stdin = nil
	m.stdout = nil
	m.enc = nil
	m.dec = nil
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
