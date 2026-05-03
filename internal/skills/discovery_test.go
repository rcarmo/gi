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
