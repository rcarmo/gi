package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ExecuteMeta(workspaceRoot string, args map[string]any) (string, error) {
	return ExecuteTool(workspaceRoot, args)
}

func ExecuteTool(workspaceRoot string, args map[string]any) (string, error) {
	d, err := Discover(workspaceRoot)
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
	var rows []Skill
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
		fmt.Fprintf(&sb, "  command: /skill:%s [args]\n", skill.Name)
		fmt.Fprintf(&sb, "  source: %s\n", filepath.ToSlash(skill.Path))
		for _, warning := range skill.Warnings {
			fmt.Fprintf(&sb, "  warning: %s\n", warning)
		}
	}
	sb.WriteString("\nUse /skill:<name> [args] in the TUI, or skills({name: \"<skill>\"}) from tools, to read SKILL.md before applying it.")
	return sb.String(), nil
}
