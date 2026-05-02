package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// toolEntry is a compact representation of a tool for the registry.
type toolEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters,omitempty"`
	Source      string `json:"source,omitempty"`
}

// ToolRuntime is supplied to registered tools at execution time.
type ToolRuntime struct {
	Engine        *Engine
	Runner        *sessionRunner
	Store         *store.Store
	SessionID     string
	TurnID        string
	WorkspaceRoot string
}

// ToolExecutor executes a model tool call and returns the text sent back to the LLM.
type ToolExecutor func(context.Context, ToolRuntime, goai.ToolCall) (string, error)

// RegisteredTool combines model-visible metadata with the executor.
type RegisteredTool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Source      string
	Executor    ToolExecutor
}

func (t RegisteredTool) Definition() goai.Tool {
	return goai.Tool{Name: t.Name, Description: t.Description, Parameters: t.Parameters}
}

type ToolRegistry struct {
	mu     sync.RWMutex
	tools  map[string]RegisteredTool
	active map[string]bool // nil/empty means all tools active
	order  []string
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]RegisteredTool)}
}

func (r *ToolRegistry) Register(tool RegisteredTool) error {
	if strings.TrimSpace(tool.Name) == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Executor == nil {
		return fmt.Errorf("tool %s executor is required", tool.Name)
	}
	if len(tool.Parameters) == 0 {
		tool.Parameters = json.RawMessage(`{"type":"object","properties":{}}`)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[tool.Name]; !exists {
		r.order = append(r.order, tool.Name)
	}
	r.tools[tool.Name] = tool
	return nil
}

func (r *ToolRegistry) Get(name string) (RegisteredTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	if !ok || !r.isActiveLocked(name) {
		return RegisteredTool{}, false
	}
	return tool, true
}

func (r *ToolRegistry) Definitions() []goai.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]goai.Tool, 0, len(r.tools))
	for _, name := range r.order {
		if !r.isActiveLocked(name) {
			continue
		}
		defs = append(defs, r.tools[name].Definition())
	}
	return defs
}

func (r *ToolRegistry) Entries() []RegisteredTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := make([]RegisteredTool, 0, len(r.tools))
	for _, name := range r.order {
		if !r.isActiveLocked(name) {
			continue
		}
		entries = append(entries, r.tools[name])
	}
	return entries
}

func (r *ToolRegistry) SetActive(names []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(names) == 0 {
		r.active = nil
		return nil
	}
	active := make(map[string]bool, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := r.tools[name]; !ok {
			return fmt.Errorf("unknown tool: %s", name)
		}
		active[name] = true
	}
	// Keep the registry exploration tool available; otherwise models can strand themselves.
	if _, ok := r.tools["tools"]; ok {
		active["tools"] = true
	}
	r.active = active
	return nil
}

func (r *ToolRegistry) ActiveNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for _, name := range r.order {
		if r.isActiveLocked(name) {
			names = append(names, name)
		}
	}
	return names
}

func (r *ToolRegistry) isActiveLocked(name string) bool {
	return len(r.active) == 0 || r.active[name]
}

func (e *Engine) RegisterTool(tool RegisteredTool) error { return e.tools.Register(tool) }
func (e *Engine) SetActiveTools(names []string) error    { return e.tools.SetActive(names) }
func (e *Engine) ActiveTools() []string                  { return e.tools.ActiveNames() }

func (e *Engine) toolDefs() []goai.Tool { return e.tools.Definitions() }

// executeToolsTool handles the "tools" meta-tool: list, search, and inspect.
func (e *Engine) executeToolsTool(args map[string]any) (string, error) {
	name, _ := args["name"].(string)
	query, _ := args["query"].(string)
	entries := e.tools.Entries()

	if name != "" {
		for _, t := range entries {
			if t.Name == name {
				var params any
				_ = json.Unmarshal(t.Parameters, &params)
				entry := toolEntry{Name: t.Name, Description: t.Description, Parameters: params, Source: t.Source}
				b, _ := json.MarshalIndent(entry, "", "  ")
				return string(b), nil
			}
		}
		return "", fmt.Errorf("tool not found: %s", name)
	}

	var rows []toolEntry
	for _, t := range entries {
		if query != "" {
			lq := strings.ToLower(query)
			if !strings.Contains(strings.ToLower(t.Name), lq) && !strings.Contains(strings.ToLower(t.Description), lq) && !strings.Contains(strings.ToLower(t.Source), lq) {
				continue
			}
		}
		rows = append(rows, toolEntry{Name: t.Name, Description: t.Description, Source: t.Source})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	if len(rows) == 0 {
		return "No tools matched the query.", nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d tool(s):\n", len(rows))
	for _, e := range rows {
		fmt.Fprintf(&sb, "- %s: %s\n", e.Name, e.Description)
	}
	sb.WriteString("\nUse tools({name: \"<tool>\"}) for full parameter schema.")
	return sb.String(), nil
}
