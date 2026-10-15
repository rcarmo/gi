package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rcarmo/gi/internal/agentdir"
)

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
	// BaseDir is the skill's directory (relative paths in it resolve there);
	// Source is "user" or "project".
	BaseDir string `json:"base_dir,omitempty"`
	Source  string `json:"source,omitempty"`
	// DisableModelInvocation hides the skill from the system prompt; it can
	// still be invoked explicitly (/skill:name).
	DisableModelInvocation bool     `json:"disable_model_invocation,omitempty"`
	Warnings               []string `json:"warnings,omitempty"`
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

// DiscoverSkills ports Pi's loadSkills: user skill directories (gi's agent
// directory, then Pi's), then the project's (.gi/skills, then .pi/skills);
// the first skill of a name wins and a file reached twice (symlinks) loads
// once. See skill_scan.go for the directory rules.
func DiscoverSkills(workspaceRoot string) ([]Skill, error) {
	var out []Skill
	seenName := map[string]bool{}
	seenFile := map[string]bool{}
	add := func(found []Skill) {
		for _, skill := range found {
			real := skill.Path
			if r, err := filepath.EvalSymlinks(skill.Path); err == nil {
				real = r
			}
			if seenFile[real] {
				continue
			}
			if seenName[skill.Name] {
				continue // Pi reports a collision; the first one wins
			}
			seenName[skill.Name], seenFile[real] = true, true
			out = append(out, skill)
		}
	}
	for _, dir := range UserSkillDirs() {
		add(loadSkillsFromDir(dir, "user"))
	}
	add(loadSkillsFromDir(filepath.Join(workspaceRoot, ".gi", "skills"), "project"))
	// The host exposes additional skills in these namespaces, including when gi
	// is launched from a nested repository. Closest directory wins; linked .pi
	// skills are deduplicated by canonical file above.
	if cwd, err := filepath.Abs(workspaceRoot); err == nil {
		for dir := cwd; ; dir = filepath.Dir(dir) {
			for _, name := range []string{".agents-in-the-cloud", ".agents"} {
				add(loadSkillsFromDir(filepath.Join(dir, name, "skills"), "project"))
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	add(loadSkillsFromDir(filepath.Join(workspaceRoot, ".pi", "skills"), "project"))
	return out, nil
}

// UserSkillDirs are the user-level skill directories (<agent dir>/skills),
// gi's first.
func UserSkillDirs() []string {
	var out []string
	for _, dir := range agentdir.Dirs() {
		out = append(out, filepath.Join(dir, "skills"))
	}
	return out
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

func relOrBase(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(path)
}
