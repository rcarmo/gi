package turn

import (
	"strings"
	"testing"
)

func TestExecuteToolsToolListAll(t *testing.T) {
	result, err := executeToolsTool(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "tool(s):") {
		t.Fatalf("expected tool count header, got: %q", result)
	}
	for _, name := range []string{"tools", "skills", "compact", "rtk", "read", "write", "shell"} {
		if !strings.Contains(result, "- "+name+":") {
			t.Fatalf("expected tool %q in listing, got: %q", name, result)
		}
	}
}

func TestExecuteToolsToolDetailByName(t *testing.T) {
	result, err := executeToolsTool(map[string]any{"name": "shell"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `"name": "shell"`) {
		t.Fatalf("expected shell detail, got: %q", result)
	}
	if !strings.Contains(result, `"command"`) {
		t.Fatalf("expected command parameter in schema, got: %q", result)
	}
}

func TestExecuteToolsToolNotFound(t *testing.T) {
	_, err := executeToolsTool(map[string]any{"name": "nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got: %v", err)
	}
}

func TestExecuteToolsToolQueryFilter(t *testing.T) {
	result, err := executeToolsTool(map[string]any{"query": "workspace"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "read") || !strings.Contains(result, "write") {
		t.Fatalf("expected read/write in filtered results, got: %q", result)
	}
	if !strings.Contains(result, "read-only") {
		t.Fatalf("expected metadata in filtered results, got: %q", result)
	}
	if strings.Contains(result, "- shell:") {
		t.Fatalf("shell should not match 'workspace' query, got: %q", result)
	}
}

func TestExecuteToolsToolActivation(t *testing.T) {
	e := New(nil)
	result, err := e.executeToolsTool(map[string]any{"activate": []any{"read"}})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !strings.Contains(result, "read") || !strings.Contains(result, "tools") {
		t.Fatalf("activation result: %q", result)
	}
	active := strings.Join(e.ActiveTools(), ",")
	if strings.Contains(active, "write") || !strings.Contains(active, "read") || !strings.Contains(active, "tools") {
		t.Fatalf("active=%s", active)
	}
	if _, err := e.executeToolsTool(map[string]any{"reset_active": true}); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if !strings.Contains(strings.Join(e.ActiveTools(), ","), "write") {
		t.Fatalf("expected reset to restore write: %v", e.ActiveTools())
	}
}

func TestExecuteToolsToolQueryNoMatch(t *testing.T) {
	result, err := executeToolsTool(map[string]any{"query": "zzzznonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "No tools matched") {
		t.Fatalf("expected no-match message, got: %q", result)
	}
}

func TestToolDefsForMetadataPreservesRegistryOrderForAllowedSubset(t *testing.T) {
	e := New(nil)
	defs := e.toolDefsForMetadata(map[string]any{"effective_tools": []string{"shell", "tools", "read"}})
	if len(defs) != 3 {
		t.Fatalf("expected 3 filtered tool defs, got %#v", defs)
	}
	got := []string{defs[0].Name, defs[1].Name, defs[2].Name}
	want := []string{"tools", "read", "shell"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("expected registry-order filtered tools %v, got %v", want, got)
	}
}
