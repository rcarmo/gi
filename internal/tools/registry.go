package tools

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

type toolEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters,omitempty"`
	Source      string `json:"source,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Weight      string `json:"weight,omitempty"`
	Activation  string `json:"activation,omitempty"`
	Active      bool   `json:"active"`
}

type ToolRuntime struct {
	Store         *store.Store
	SessionID     string
	TurnID        string
	WorkspaceRoot string
	// OnOutput receives cumulative command output; errors fail execution closed.
	OnOutput func(string) error
	// AttachImage adds an image to the tool result sent to the model (nil
	// when the caller cannot attach images; tools then describe them in text).
	AttachImage func(mimeType string, data []byte)
	// AddTools reports tools this call loaded for the session (tool_search);
	// they are declared from the next model call (nil outside turns).
	AddTools func(names []string)
	// ToolOccurrenceID identifies this execution independently of reused model IDs.
	ToolOccurrenceID string
	// ToolCallID is the model's ID for this call (the parent of codemode's
	// nested calls).
	ToolCallID string
	// PublishMessage broadcasts a message only after its durable write succeeds.
	PublishMessage func(store.Message)
	// PublishPlan announces a committed session plan mutation.
	PublishPlan func(store.SessionPlan, string, string)
	// SetDetails attaches structured details to the result for renderers
	// (nil outside turns), e.g. codemode's nested calls.
	SetDetails func(map[string]any)
	// AddUsage adds model usage made by the tool (codemode's models.*) to
	// the turn's usage and cost (nil outside turns).
	AddUsage func(goai.Usage)
	// UnknownUsage marks a completed remote model call whose counters were lost.
	UnknownUsage func()
}

type ToolExecutor func(context.Context, ToolRuntime, goai.ToolCall) (string, error)

type RegisteredTool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Source      string
	Kind        string
	Weight      string
	Activation  string
	Executor    ToolExecutor
	// Deferred tools are registered (executable) but not declared by default:
	// they reach the model only when loaded (tool_search) or listed explicitly
	// in the active set (Pi's deferred/codemode exposure).
	Deferred bool
	// ReplaySafe permits host replay only when both the committed intent and
	// current registration opt in. Tools are unsafe by default.
	ReplaySafe bool
	// ModelOnly tools (codemode, tool_search) cannot be called from codemode
	// scripts.
	ModelOnly bool
	// OutputSchema is the JSON Schema of the value codemode scripts receive;
	// empty means text (Pi's default). MCP tools declare a CallToolResult.
	OutputSchema json.RawMessage
	// StructuredExecutor, when set, returns the value codemode scripts
	// receive (e.g. an MCP CallToolResult as JSON) and whether it reports an
	// error; scripts still resolve to it, like Pi's structuredContent.
	StructuredExecutor func(context.Context, ToolRuntime, goai.ToolCall) (json.RawMessage, bool, error)
	// PromptSnippet and PromptGuidelines are the tool's contribution to the
	// system prompt's tools and rules sections (Pi's promptSnippet and
	// promptGuidelines), used while the tool is declared.
	// CompletesTurn marks tools that publish the assistant response themselves.
	// The runner still drains the admitted tool-call batch before completion.
	CompletesTurn    bool
	PromptSnippet    string
	PromptGuidelines []string
}

func (t RegisteredTool) Definition() goai.Tool {
	return goai.Tool{Name: t.Name, Description: t.Description, Parameters: t.Parameters}
}

type ToolRegistry struct {
	mu     sync.RWMutex
	tools  map[string]RegisteredTool
	active map[string]bool
	order  []string
}

func NewToolRegistry() *ToolRegistry { return &ToolRegistry{tools: make(map[string]RegisteredTool)} }

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
	if strings.TrimSpace(tool.Source) == "" {
		tool.Source = "runtime"
	}
	if strings.TrimSpace(tool.Kind) == "" {
		tool.Kind = "mixed"
	}
	if strings.TrimSpace(tool.Weight) == "" {
		tool.Weight = "standard"
	}
	if strings.TrimSpace(tool.Activation) == "" {
		tool.Activation = "default"
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[tool.Name]; !exists {
		r.order = append(r.order, tool.Name)
	}
	r.tools[tool.Name] = tool
	return nil
}

// Unregister removes a tool (e.g. an MCP tool its server withdrew).
func (r *ToolRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tools[name]; !ok {
		return
	}
	delete(r.tools, name)
	delete(r.active, name)
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
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

func (r *ToolRegistry) GetRegistered(name string) (RegisteredTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
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
func (r *ToolRegistry) Entries() []RegisteredTool    { return r.entries(false) }
func (r *ToolRegistry) AllEntries() []RegisteredTool { return r.entries(true) }
func (r *ToolRegistry) entries(includeInactive bool) []RegisteredTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := make([]RegisteredTool, 0, len(r.tools))
	for _, name := range r.order {
		if !includeInactive && !r.isActiveLocked(name) {
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
	if len(r.active) == 0 {
		return !r.tools[name].Deferred
	}
	return r.active[name]
}

// DeferredEntries lists registered deferred tools in registration order.
func (r *ToolRegistry) DeferredEntries() []RegisteredTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []RegisteredTool
	for _, name := range r.order {
		if t := r.tools[name]; t.Deferred {
			out = append(out, t)
		}
	}
	return out
}

func ExecuteToolsTool(reg *ToolRegistry, args map[string]any, setActive func([]string) error, activeNames func() []string, reset func()) (string, error) {
	name, _ := args["name"].(string)
	query, _ := args["query"].(string)
	intent, _ := args["intent"].(string)
	includeInactive, _ := args["include_inactive"].(bool)
	includeParameters, _ := args["include_parameters"].(bool)
	if raw, ok := args["activate"].([]any); ok {
		names := make([]string, 0, len(raw))
		for _, v := range raw {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				names = append(names, strings.TrimSpace(s))
			}
		}
		if err := setActive(names); err != nil {
			return "", err
		}
		return fmt.Sprintf("Activated tools: %s", strings.Join(activeNames(), ", ")), nil
	}
	if resetActive, _ := args["reset_active"].(bool); resetActive {
		reset()
		return "Reset active tools to default registry set.", nil
	}
	entries := reg.Entries()
	if includeInactive {
		entries = reg.AllEntries()
	}
	active := map[string]bool{}
	for _, n := range activeNames() {
		active[n] = true
	}
	if name != "" {
		for _, t := range entries {
			if t.Name == name {
				var params any
				if err := json.Unmarshal(t.Parameters, &params); err != nil {
					params = map[string]any{"error": "invalid parameter schema", "raw": string(t.Parameters)}
				}
				entry := toolEntry{Name: t.Name, Description: t.Description, Parameters: params, Source: t.Source, Kind: t.Kind, Weight: t.Weight, Activation: t.Activation, Active: active[t.Name]}
				b, err := json.MarshalIndent(entry, "", "  ")
				if err != nil {
					return "", err
				}
				return string(b), nil
			}
		}
		return "", fmt.Errorf("tool not found: %s", name)
	}
	var rows []toolEntry
	needle := strings.ToLower(strings.TrimSpace(FirstNonEmpty(query, intent)))
	for _, t := range entries {
		if needle != "" {
			body := strings.ToLower(strings.Join([]string{t.Name, t.Description, t.Source, t.Kind, t.Weight, t.Activation}, " "))
			if !strings.Contains(body, needle) {
				continue
			}
		}
		row := toolEntry{Name: t.Name, Description: t.Description, Source: t.Source, Kind: t.Kind, Weight: t.Weight, Activation: t.Activation, Active: active[t.Name]}
		if includeParameters {
			if err := json.Unmarshal(t.Parameters, &row.Parameters); err != nil {
				row.Parameters = map[string]any{"error": "invalid parameter schema", "raw": string(t.Parameters)}
			}
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	if len(rows) == 0 {
		return "No tools matched the query.", nil
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d tool(s):\n", len(rows))
	for _, e := range rows {
		status := "inactive"
		if e.Active {
			status = "active"
		}
		fmt.Fprintf(&sb, "- %s: %s [%s/%s/%s, %s, source=%s]\n", e.Name, e.Description, e.Kind, e.Weight, e.Activation, status, e.Source)
	}
	sb.WriteString("\nUse tools({name: \"<tool>\"}) for full metadata/schema. Use tools({activate:[...]}) or tools({reset_active:true}) to change active tools.")
	return sb.String(), nil
}
