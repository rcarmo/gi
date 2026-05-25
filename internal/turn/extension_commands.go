package turn

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/rcarmo/gi/internal/topics"
)

type ExtensionCommandSpec struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Usage       string `json:"usage,omitempty"`
	Source      string `json:"source,omitempty"`
	Engine      string `json:"engine,omitempty"`
}

type ExtensionCommandContext struct {
	Name      string   `json:"name"`
	Args      string   `json:"args"`
	Argv      []string `json:"argv,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
	AgentID   string   `json:"agent_id,omitempty"`
}

type ExtensionCommandResult struct {
	Type   string   `json:"type,omitempty"`
	Lines  []string `json:"lines,omitempty"`
	Prompt string   `json:"prompt,omitempty"`
	Status string   `json:"status,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type ExtensionCommandHandler func(context.Context, ExtensionCommandContext) (ExtensionCommandResult, error)

type registeredExtensionCommand struct {
	spec    ExtensionCommandSpec
	handler ExtensionCommandHandler
}

type ExtensionCommandRegistry struct {
	mu        sync.RWMutex
	commands  map[string]registeredExtensionCommand
	conflicts []ExtensionCommandSpec
}

var extensionCommandNameRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func NewExtensionCommandRegistry() *ExtensionCommandRegistry {
	return &ExtensionCommandRegistry{commands: map[string]registeredExtensionCommand{}}
}

func normalizeExtensionCommandName(name string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, "/")))
}

func (r *ExtensionCommandRegistry) Register(spec ExtensionCommandSpec, handler ExtensionCommandHandler) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("extension command registry is nil")
	}
	name := normalizeExtensionCommandName(spec.Name)
	if name == "" || !extensionCommandNameRE.MatchString(name) {
		return false, fmt.Errorf("invalid extension command name: %q", spec.Name)
	}
	if handler == nil {
		return false, fmt.Errorf("extension command %s has no handler", name)
	}
	spec.Name = name
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.commands == nil {
		r.commands = map[string]registeredExtensionCommand{}
	}
	if _, exists := r.commands[name]; exists {
		r.conflicts = append(r.conflicts, spec)
		return false, nil
	}
	r.commands[name] = registeredExtensionCommand{spec: spec, handler: handler}
	return true, nil
}

func (r *ExtensionCommandRegistry) List() []ExtensionCommandSpec {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ExtensionCommandSpec, 0, len(r.commands))
	for _, cmd := range r.commands {
		out = append(out, cmd.spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *ExtensionCommandRegistry) Conflicts() []ExtensionCommandSpec {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := append([]ExtensionCommandSpec(nil), r.conflicts...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *ExtensionCommandRegistry) Lookup(name string) (registeredExtensionCommand, bool) {
	if r == nil {
		return registeredExtensionCommand{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.commands[normalizeExtensionCommandName(name)]
	return cmd, ok
}

func (e *Engine) RegisterExtensionCommand(spec ExtensionCommandSpec, handler ExtensionCommandHandler) (bool, error) {
	if e.extensionCommands == nil {
		e.extensionCommands = NewExtensionCommandRegistry()
	}
	registered, err := e.extensionCommands.Register(spec, handler)
	if err != nil {
		return false, err
	}
	e.publishExtensionCommandTopic(map[string]any{"type": map[bool]string{true: "registered", false: "conflict"}[registered], "command": normalizeExtensionCommandName(spec.Name), "engine": spec.Engine, "source": spec.Source})
	return registered, nil
}

func (e *Engine) ExtensionCommandInfos() []ExtensionCommandSpec {
	if e.extensionCommands == nil {
		return nil
	}
	return e.extensionCommands.List()
}

func (e *Engine) ExtensionCommandConflicts() []ExtensionCommandSpec {
	if e.extensionCommands == nil {
		return nil
	}
	return e.extensionCommands.Conflicts()
}

func (e *Engine) InvokeExtensionCommand(ctx context.Context, name, rawArgs, sessionID, agentID string) (ExtensionCommandResult, bool, error) {
	if e.extensionCommands == nil {
		return ExtensionCommandResult{}, false, nil
	}
	cmd, ok := e.extensionCommands.Lookup(name)
	if !ok {
		return ExtensionCommandResult{}, false, nil
	}
	argv := []string{}
	if strings.TrimSpace(rawArgs) != "" {
		argv = strings.Fields(rawArgs)
	}
	e.publishExtensionCommandTopic(map[string]any{"type": "invoked", "command": cmd.spec.Name, "engine": cmd.spec.Engine, "source": cmd.spec.Source, "session_id": sessionID, "agent_id": agentID})
	res, err := cmd.handler(ctx, ExtensionCommandContext{Name: cmd.spec.Name, Args: rawArgs, Argv: argv, SessionID: sessionID, AgentID: agentID})
	if err != nil {
		e.publishExtensionCommandTopic(map[string]any{"type": "failed", "command": cmd.spec.Name, "engine": cmd.spec.Engine, "source": cmd.spec.Source, "session_id": sessionID, "agent_id": agentID, "error": err.Error()})
		return res, true, err
	}
	if strings.TrimSpace(res.Type) == "" {
		res.Type = "message"
	}
	return res, true, nil
}

func (e *Engine) publishExtensionCommandTopic(payload map[string]any) {
	if e == nil {
		return
	}
	e.publishTopicEvent(topics.Envelope{Topic: "extension.command", Source: "extension", Type: "notice", Payload: payload})
}
