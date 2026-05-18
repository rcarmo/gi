package turn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/connectivity"
	"github.com/rcarmo/gi/internal/scripting"
	giskills "github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/tools"
	"github.com/rcarmo/gi/internal/topics"
	goai "github.com/rcarmo/go-ai"
)

func (e *Engine) registerDefaultTools() {
	scriptTool := tools.NewScriptTool(e.store, e.runtimeCfg)
	scriptTool.SetConnectivityCallbacks(
		func(ctx context.Context, sessionID string, spec connectivity.RouteSpec) (connectivity.RouteInfo, error) {
			if spec.SessionID == "" {
				spec.SessionID = sessionID
			}
			return e.connectivity.Register(ctx, spec, func(ctx context.Context, event connectivity.EventEnvelope) (connectivity.RouteResponse, error) {
				payload := map[string]any{"event": event, "route": spec, "payload": event.Payload}
				input := tools.ScriptInput{Engine: spec.Engine, Path: spec.Path, SessionID: event.SessionID, Script: tools.ScriptWithPayload(spec.Engine, "event", payload, spec.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return connectivity.RouteResponse{Status: 500, Body: out.Result}, errors.New(out.Error)
				}
				resp := connectivity.RouteResponse{Status: 200, Body: out.Result}
				if strings.TrimSpace(out.Result) != "" {
					if err := json.Unmarshal([]byte(out.Result), &resp); err != nil {
						resp = connectivity.RouteResponse{Status: 200, Body: out.Result}
					} else if resp.Status == 0 && resp.Body == "" {
						resp.Body = out.Result
					}
				}
				if resp.Status == 0 {
					resp.Status = 200
				}
				return resp, nil
			})
		},
		func(ctx context.Context, sessionID, id string) error { return e.connectivity.Unregister(ctx, id) },
		func(ctx context.Context, sessionID string, filter map[string]any) ([]connectivity.RouteInfo, error) {
			return e.connectivity.List(ctx, filter)
		},
		func(ctx context.Context, sessionID, topic string, payload map[string]any) error {
			if payload == nil {
				payload = map[string]any{}
			}
			payload["session_id"] = sessionID
			return e.connectivity.Emit(ctx, topic, payload)
		},
		func(ctx context.Context, sessionID string, envelope map[string]any) error {
			if e == nil || e.Topics() == nil {
				return fmt.Errorf("publish topic: topic bus is not available")
			}
			topicName, _ := envelope["topic"].(string)
			topicName = strings.TrimSpace(topicName)
			if topicName == "" {
				return fmt.Errorf("publish topic: topic is required")
			}
			payload, _ := envelope["payload"].(map[string]any)
			if payload == nil {
				payload = map[string]any{}
			}
			agentID, _ := envelope["agent_id"].(string)
			source, _ := envelope["source"].(string)
			if strings.TrimSpace(source) == "" {
				source = "script"
			}
			typ, _ := envelope["type"].(string)
			sessionValue, _ := envelope["session_id"].(string)
			if strings.TrimSpace(sessionValue) == "" {
				sessionValue = sessionID
			}
			e.PublishTopicEvent(topics.Envelope{Topic: topicName, SessionID: sessionValue, AgentID: strings.TrimSpace(agentID), Source: source, Type: strings.TrimSpace(typ), Payload: payload})
			return nil
		},
		func(ctx context.Context, sessionID string, pattern string, opts scripting.TopicSubscribeOptions) (<-chan topics.Envelope, func(), error) {
			if e == nil || e.Topics() == nil {
				return nil, nil, fmt.Errorf("subscribe topic: topic bus is not available")
			}
			subOpts := topics.SubscribeOptions{Buffer: opts.Buffer, SessionID: strings.TrimSpace(opts.SessionID), AgentID: strings.TrimSpace(opts.AgentID)}
			if subOpts.SessionID == "" {
				subOpts.SessionID = sessionID
			}
			ch, unsubscribe := e.Topics().Subscribe(ctx, pattern, subOpts)
			return ch, unsubscribe, nil
		},
	)
	scriptTool.SetAgenticCallbacks(
		func(ctx context.Context, sessionID string, hook scripting.EventHookSpec) (func(), error) {
			if isProcessHookSpec(hook) {
				return e.RegisterHook(hook.Name, tools.FirstNonEmpty(hook.Source, "process"), newProcessHookHandler(e.runtimeCfg.WorkspaceRoot, hook))
			}
			return e.RegisterHook(hook.Name, tools.FirstNonEmpty(hook.Source, "script"), func(ctx context.Context, req HookRequest) (HookResponse, error) {
				payload := hookScriptPayload(req)
				input := tools.ScriptInput{Engine: hook.Engine, Path: hook.Path, SessionID: req.SessionID, Script: tools.ScriptWithPayload(hook.Engine, "hook", payload, hook.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return HookResponse{}, errors.New(out.Error)
				}
				return hookResponseFromScript(out.Result)
			})
		},
		func(ctx context.Context, sessionID string, spec scripting.ToolSpec) error {
			params := spec.Parameters
			if len(params) == 0 {
				params = json.RawMessage(`{"type":"object","properties":{}}`)
			}
			return e.RegisterTool(RegisteredTool{Name: spec.Name, Description: spec.Description, Parameters: params, Source: tools.FirstNonEmpty(spec.Source, "script"), Kind: "mixed", Weight: "standard", Activation: "on-demand", Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
				payload := map[string]any{"tool_call_id": call.ID, "name": call.Name, "arguments": call.Arguments, "session_id": rt.SessionID}
				input := tools.ScriptInput{Engine: spec.Engine, Path: spec.Path, SessionID: rt.SessionID, Script: tools.ScriptWithPayload(spec.Engine, "tool", payload, spec.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return out.Result, errors.New(out.Error)
				}
				return out.Result, nil
			}})
		},
		func(ctx context.Context, sessionID string, names []string) error { return e.SetActiveTools(names) },
		func(ctx context.Context, sessionID string) ([]string, error) { return e.ActiveTools(), nil },
	)
	e.loadWorkspaceExtensions(scriptTool)
	registerDiscoveredTools := func() {
		giskills.RegisterDiscoveredTools(e.runtimeCfg.Discovery.Tools, func(tool giskills.ToolManifest, params json.RawMessage) error {
			return e.RegisterTool(RegisteredTool{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  params,
				Source:      "workspace-manifest",
				Kind:        "mixed",
				Weight:      "standard",
				Activation:  "on-demand",
				Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
					payload := map[string]any{"tool_call_id": call.ID, "name": call.Name, "arguments": call.Arguments, "session_id": rt.SessionID}
					input := tools.ScriptInput{Engine: tool.Engine, Path: tool.Path, SessionID: rt.SessionID, Script: tools.ScriptWithPayload(tool.Engine, "tool", payload, tool.Script)}
					out := scriptTool.Execute(ctx, input)
					if out.Error != "" {
						return out.Result, errors.New(out.Error)
					}
					return out.Result, nil
				},
			})
		}, func(name string, err error) {
			log.Printf("register discovered tool %q: %v", name, err)
		})
	}
	must := func(t RegisteredTool) {
		if err := e.RegisterTool(t); err != nil {
			panic(err)
		}
	}
	must(RegisteredTool{
		Name:        "tools",
		Description: "List available tools or get details about a specific tool. Use with no arguments to list all tools (names + short descriptions). Pass a tool name via the `name` argument to get its full schema and usage. Use `query` to filter tools by keyword.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Exact tool name to get full details for"},"query":{"type":"string","description":"Filter tools by keyword in name/description/metadata"},"intent":{"type":"string","description":"Natural-language goal for staged discovery"},"include_parameters":{"type":"boolean","description":"Include parameter schemas in list results"},"include_inactive":{"type":"boolean","description":"Include inactive tools in discovery results"},"activate":{"type":"array","items":{"type":"string"},"description":"Set active tools by name; tools remains active"},"reset_active":{"type":"boolean","description":"Reset active tools to all default registry tools"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			return rt.Engine.executeToolsTool(call.Arguments)
		},
	})
	must(RegisteredTool{
		Name:        "skills",
		Description: "List workspace-discovered skills or read a skill's SKILL.md. Skills are discovered from .gi/skills and .pi/skills.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Exact skill name to read; omit to list skills"},"query":{"type":"string","description":"Filter listed skills by name or description"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			return giskills.ExecuteTool(rt.WorkspaceRoot, call.Arguments)
		},
	})
	registerDiscoveredTools()
	must(RegisteredTool{
		Name:        "read",
		Description: "Read text content from a workspace file. Supports workspace-relative paths and vfs:// paths.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Workspace-relative path or vfs://namespace/path"}},"required":["path"]}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteRead(ctx, rt.WorkspaceRoot, rt.Store, call)
		},
	})
	must(RegisteredTool{
		Name:        "write",
		Description: "Write text content to a workspace file. Creates parent directories for workspace paths.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Workspace-relative path or vfs://namespace/path"},"content":{"type":"string","description":"File content to write"}},"required":["path","content"]}`),
		Source:      "builtin",
		Kind:        "mutating",
		Weight:      "lightweight",
		Activation:  "default",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteWrite(ctx, rt.WorkspaceRoot, rt.Store, call)
		},
	})
	if def := scriptTool.Definition(); def != nil {
		params, _ := json.Marshal(def["parameters"])
		must(RegisteredTool{
			Name:        "script",
			Description: fmt.Sprint(def["description"]),
			Parameters:  params,
			Source:      "builtin",
			Kind:        "mixed",
			Weight:      "standard",
			Activation:  "default",
			Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
				input := tools.ScriptInput{SessionID: rt.SessionID}
				b, _ := json.Marshal(call.Arguments)
				if err := json.Unmarshal(b, &input); err != nil {
					return "", err
				}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return out.Result, errors.New(out.Error)
				}
				return out.Result, nil
			},
		})
	}
	must(RegisteredTool{
		Name:        "compact",
		Description: "Inspect compaction thresholds or estimate whether the current session should compact. Supports smart compaction plugins via session_before_compact hooks.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"session_id":{"type":"string","description":"Session to inspect; defaults to current session"},"dry_run":{"type":"boolean","description":"Return estimate/preparation without changing context"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "standard",
		Activation:  "on-demand",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			sessionID, _ := call.Arguments["session_id"].(string)
			if sessionID == "" {
				sessionID = rt.SessionID
			}
			msgs, err := rt.Store.ListMessages(ctx, sessionID)
			if err != nil {
				return "", err
			}
			var aiMsgs []goai.Message
			for _, m := range msgs {
				switch m.Role {
				case "user":
					aiMsgs = append(aiMsgs, goai.UserMessage(m.Content))
				case "assistant":
					aiMsgs = append(aiMsgs, goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}})
				}
			}
			settings := rt.Engine.runtimeCfg.Compaction
			tokens := compaction.EstimateMessagesTokens(aiMsgs)
			prep := compaction.Prepare(aiMsgs, tokens, settings.KeepRecentTokens, settings.ReserveTokens, settings.ThresholdTokens, settings.Strategy)
			b, _ := json.MarshalIndent(map[string]any{"enabled": settings.Enabled, "context_tokens": tokens, "threshold_tokens": settings.ThresholdTokens, "should_compact": settings.Enabled && tokens > settings.ThresholdTokens, "preparation": prep}, "", "  ")
			return string(b), nil
		},
	})
	must(RegisteredTool{
		Name:        "peering",
		Description: "Inspect gi peer-discovery status. The backend is tsnet/Tailscale and is disabled by default until configured.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"action":{"type":"string","description":"status (default); future: start/stop/discover"}}}`),
		Source:      "builtin",
		Kind:        "read-only",
		Weight:      "lightweight",
		Activation:  "on-demand",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			b, _ := json.MarshalIndent(rt.Engine.PeeringStatus(), "", "  ")
			return string(b), nil
		},
	})
	must(RegisteredTool{
		Name:        "rtk",
		Description: "Run a shell command and return RTK-style compact output using gi's native Go filters for git/search/listing/test output.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute and compact"},"filter_only":{"type":"boolean","description":"Filter the supplied output instead of executing"},"output":{"type":"string","description":"Raw output to filter when filter_only is true"}},"required":["command"]}`),
		Source:      "builtin",
		Kind:        "mixed",
		Weight:      "standard",
		Activation:  "on-demand",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteRTK(ctx, rt.WorkspaceRoot, call)
		},
	})
	must(RegisteredTool{
		Name:        "shell",
		Description: "Execute a shell command and return stdout/stderr. Use for running tests, installing packages, searching files, etc.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"}},"required":["command"]}`),
		Source:      "builtin",
		Kind:        "mixed",
		Weight:      "heavy",
		Activation:  "default",
		Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
			return tools.ExecuteShell(ctx, rt.WorkspaceRoot, call)
		},
	})
}
