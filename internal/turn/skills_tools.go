package turn

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	giskills "github.com/rcarmo/gi/internal/skills"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

func ExecuteSkillsMeta(workspaceRoot string, args map[string]any) (string, error) {
	return executeSkillsTool(workspaceRoot, args)
}

func executeSkillsTool(workspaceRoot string, args map[string]any) (string, error) {
	d, err := giskills.Discover(workspaceRoot)
	if err != nil {
		return "", err
	}
	name, _ := args["name"].(string)
	query, _ := args["query"].(string)
	name = strings.TrimSpace(name)
	query = strings.ToLower(strings.TrimSpace(query))
	if name != "" {
		for _, skill := range d.Skills {
			if strings.EqualFold(skill.Name, name) {
				data, err := os.ReadFile(skill.Path)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
		return "", fmt.Errorf("unknown skill: %s", name)
	}
	var rows []giskills.Skill
	for _, skill := range d.Skills {
		if query == "" || strings.Contains(strings.ToLower(skill.Name), query) || strings.Contains(strings.ToLower(skill.Description), query) {
			rows = append(rows, skill)
		}
	}
	if len(rows) == 0 {
		return "No skills matched the query.", nil
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d skill(s) available:\n", len(rows))
	for _, skill := range rows {
		fmt.Fprintf(&sb, "- %s: %s\n", skill.Name, skill.Description)
	}
	sb.WriteString("\nUse skills({name: \"<skill>\"}) to read SKILL.md before applying it.")
	return sb.String(), nil
}

func registerDiscoveredTools(e *Engine, scriptTool *tools.ScriptTool) {
	for _, manifest := range e.runtimeCfg.Discovery.Tools {
		tool := manifest
		if strings.TrimSpace(tool.Name) == "" || strings.TrimSpace(tool.Script) == "" && strings.TrimSpace(tool.Path) == "" {
			continue
		}
		params := tool.Parameters
		if len(params) == 0 {
			params = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		if err := e.RegisterTool(RegisteredTool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  params,
			Source:      "workspace-manifest",
			Kind:        "mixed",
			Weight:      "standard",
			Activation:  "on-demand",
			Executor: func(ctx context.Context, rt ToolRuntime, call goai.ToolCall) (string, error) {
				payload := map[string]any{"tool_call_id": call.ID, "name": call.Name, "arguments": call.Arguments, "session_id": rt.SessionID}
				input := tools.ScriptInput{Engine: tool.Engine, Path: tool.Path, SessionID: rt.SessionID, Script: scriptWithPayload(tool.Engine, "tool", payload, tool.Script)}
				out := scriptTool.Execute(ctx, input)
				if out.Error != "" {
					return out.Result, fmt.Errorf("%s", out.Error)
				}
				return out.Result, nil
			},
		}); err != nil {
			log.Printf("register discovered tool %q: %v", tool.Name, err)
		}
	}
}
