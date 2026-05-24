package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverSkillsAndToolManifests(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, ".gi", "skills", "demo")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("Name: demo\nDescription: Demo skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	toolDir := filepath.Join(root, ".gi", "tools")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "hello.json"), []byte(`{"name":"hello","description":"Hello tool","engine":"js","script":"'ok'"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := Discover(root)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(d.Skills) != 1 || d.Skills[0].Name != "demo" || d.Skills[0].Description != "Demo skill" {
		t.Fatalf("skills: %#v", d.Skills)
	}
	if len(d.Tools) != 1 || d.Tools[0].Name != "hello" {
		t.Fatalf("tools: %#v", d.Tools)
	}
	if summary := PromptSummary(d); !strings.Contains(summary, "demo") || !strings.Contains(summary, "hello") {
		t.Fatalf("summary: %s", summary)
	}
}

func TestDiscoverSkillsWarnsForMissingAgentSkillFields(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, ".gi", "skills", "legacy")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Legacy skill\n\nUse it."), 0o644); err != nil {
		t.Fatal(err)
	}
	skills, err := DiscoverSkills(root)
	if err != nil {
		t.Fatalf("discover skills: %v", err)
	}
	if len(skills) != 1 || skills[0].Name != "legacy" || skills[0].Description != "Legacy skill" {
		t.Fatalf("skills: %#v", skills)
	}
	joined := strings.Join(skills[0].Warnings, "\n")
	for _, want := range []string{"missing Name field", "missing Description field"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("warnings missing %q: %#v", want, skills[0].Warnings)
		}
	}
	out, err := ExecuteMeta(root, map[string]any{})
	if err != nil {
		t.Fatalf("execute meta: %v", err)
	}
	for _, want := range []string{"command: /skill:legacy [args]", "source:", "warning: missing Name field"} {
		if !strings.Contains(out, want) {
			t.Fatalf("skill list missing %q:\n%s", want, out)
		}
	}
}
