package turn

import (
	"context"
	"fmt"
	"log"

	"github.com/rcarmo/gi/internal/store"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func (e *Engine) RegisterTool(tool tools.RegisteredTool) error { return e.tools.Register(tool) }
func (e *Engine) SetActiveTools(names []string) error          { return e.tools.SetActive(names) }
func (e *Engine) ActiveTools() []string                        { return e.tools.ActiveNames() }
func (e *Engine) ResetActiveTools() {
	if err := e.tools.SetActive(nil); err != nil {
		log.Printf("reset active tools: %v", err)
	}
}
func (e *Engine) ToolEntries() []tools.RegisteredTool { return e.tools.AllEntries() }
func (e *Engine) toolDefs() []goai.Tool               { return e.tools.Definitions() }
func (e *Engine) ExecuteToolsMeta(args map[string]any) (string, error) {
	return e.executeToolsTool(args)
}
func (e *Engine) ExecuteToolByName(ctx context.Context, name, sessionID string, args map[string]any) (string, error) {
	tool, ok := e.tools.Get(name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Executor(ctx, tools.ToolRuntime{Store: e.store, SessionID: sessionID, WorkspaceRoot: e.runtimeCfg.WorkspaceRoot}, goai.ToolCall{Name: name, Arguments: args})
}
func (e *Engine) executeToolsTool(args map[string]any) (string, error) {
	return tools.ExecuteToolsTool(e.tools, args, e.SetActiveTools, e.ActiveTools, e.ResetActiveTools)
}

func (e *Engine) allRegisteredToolNames() []string {
	entries := e.tools.AllEntries()
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name)
	}
	return tools.NormalizeToolNames(out)
}
func (e *Engine) defaultEffectiveToolNames() []string {
	return tools.NormalizeToolNames(e.ActiveTools())
}
func (e *Engine) resolveEffectiveToolNames(parentTurn *store.Turn, inputMetadata map[string]any) ([]string, bool, error) {
	var parent map[string]any
	if parentTurn != nil {
		parent = parentTurn.Metadata
	}
	return tools.ResolveEffectiveToolNames(parent, inputMetadata, e.defaultEffectiveToolNames(), e.allRegisteredToolNames())
}
func toolAllowedByMetadata(metadata map[string]any, toolName string) bool {
	return tools.ToolAllowedByMetadata(metadata, toolName)
}
func (e *Engine) toolDefsForMetadata(metadata map[string]any) []goai.Tool {
	return tools.ToolDefsForMetadata(metadata, e.tools.AllEntries(), e.toolDefs())
}
