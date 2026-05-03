package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
}

type ToolManifest struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Engine      string          `json:"engine,omitempty"`
	Script      string          `json:"script,omitempty"`
	Path        string          `json:"path,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type Discovery struct {
	Skills []Skill        `json:"skills"`
	Tools  []ToolManifest `json:"tools"`
}

func Discover(workspaceRoot string) (Discovery, error) {
	var out Discovery
	if workspaceRoot == "" {
		return out, nil
	}
	skills, err := DiscoverSkills(workspaceRoot)
	if err != nil {
		return out, err
	}
	tools, err := DiscoverToolManifests(workspaceRoot)
	if err != nil {
		return out, err
	}
	out.Skills = skills
	out.Tools = tools
	return out, nil
}

func DiscoverSkills(workspaceRoot string) ([]Skill, error) {
	roots := []string{filepath.Join(workspaceRoot, ".gi", "skills"), filepath.Join(workspaceRoot, ".pi", "skills")}
	seen := map[string]bool{}
	var out []Skill
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(root, entry.Name(), "SKILL.md")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			skill := parseSkillMarkdown(entry.Name(), path, string(data))
			key := strings.ToLower(skill.Name)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, skill)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func DiscoverToolManifests(workspaceRoot string) ([]ToolManifest, error) {
	roots := []string{filepath.Join(workspaceRoot, ".gi", "tools"), filepath.Join(workspaceRoot, ".pi", "tools")}
	seen := map[string]bool{}
	var out []ToolManifest
	for _, root := range roots {
		matches, err := filepath.Glob(filepath.Join(root, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, path := range matches {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var tool ToolManifest
			if err := json.Unmarshal(data, &tool); err != nil {
				continue
			}
			tool.Name = strings.TrimSpace(tool.Name)
			if tool.Name == "" || seen[strings.ToLower(tool.Name)] {
				continue
			}
			if tool.Path != "" && !filepath.IsAbs(tool.Path) {
				tool.Path = filepath.Join(filepath.Dir(path), tool.Path)
			}
			seen[strings.ToLower(tool.Name)] = true
			out = append(out, tool)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func PromptSummary(d Discovery) string {
	if len(d.Skills) == 0 && len(d.Tools) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\n## Workspace-discovered capabilities\n")
	if len(d.Skills) > 0 {
		sb.WriteString("\nSkills are discovered from `.gi/skills/*/SKILL.md` and `.pi/skills/*/SKILL.md`. Load a skill before using it when the task matches.\n")
		for _, skill := range d.Skills {
			fmt.Fprintf(&sb, "- %s: %s (%s)\n", skill.Name, skill.Description, relOrBase(skill.Path))
		}
	}
	if len(d.Tools) > 0 {
		sb.WriteString("\nScript tool manifests are discovered from `.gi/tools/*.json` and `.pi/tools/*.json` and registered in the tool registry.\n")
		for _, tool := range d.Tools {
			fmt.Fprintf(&sb, "- %s: %s\n", tool.Name, tool.Description)
		}
	}
	return sb.String()
}

var mdFieldRE = regexp.MustCompile(`(?m)^([A-Za-z][A-Za-z0-9_-]*):\s*(.+)$`)

func parseSkillMarkdown(fallbackName, path, body string) Skill {
	s := Skill{Name: fallbackName, Path: path}
	for _, match := range mdFieldRE.FindAllStringSubmatch(body, -1) {
		switch strings.ToLower(match[1]) {
		case "name":
			s.Name = strings.TrimSpace(match[2])
		case "description":
			s.Description = strings.TrimSpace(match[2])
		}
	}
	if s.Description == "" {
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if line != "" && !strings.Contains(line, ":") {
				s.Description = line
				break
			}
		}
	}
	return s
}

func relOrBase(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(path)
}
