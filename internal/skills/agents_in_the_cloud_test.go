package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAgentsInTheCloudSkillNamespacesAndCompatibilityLinks(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "repo")
	t.Setenv("GI_CODING_AGENT_DIR", filepath.Join(root, "empty-gi"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(root, "empty-pi"))
	write := func(rel, name string) {
		t.Helper()
		p := filepath.Join(root, rel, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("---\nname: "+name+"\ndescription: Fixture skill\n---\nFixture.\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(".agents-in-the-cloud/skills/cloud", "cloud")
	write(".agents/skills/standard", "standard")
	write(".agents-in-the-cloud/skills/override", "override")
	write("repo/.gi/skills/override", "override")
	if err := os.MkdirAll(filepath.Join(workspace, ".pi"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, ".agents-in-the-cloud/skills"), filepath.Join(workspace, ".pi/skills")); err != nil {
		t.Fatal(err)
	}
	found, err := DiscoverSkills(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 3 {
		t.Fatalf("missing or duplicated skills: %+v", found)
	}
	paths := map[string]string{}
	for _, s := range found {
		paths[s.Name] = s.Path
	}
	for name, rel := range map[string]string{"cloud": ".agents-in-the-cloud/skills/cloud/SKILL.md", "standard": ".agents/skills/standard/SKILL.md", "override": "repo/.gi/skills/override/SKILL.md"} {
		if paths[name] != filepath.Join(root, rel) {
			t.Fatalf("%s resolved to %s", name, paths[name])
		}
	}
}
