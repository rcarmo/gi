package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gotui "github.com/grindlemire/go-tui"
	gimcp "github.com/rcarmo/gi/internal/mcp"
	"github.com/rcarmo/gi/internal/mcp/mcptest"
)

// Pi's McpManagerView rendered by Pi: bun scripts/golden-mcp-manager.mjs
func TestMCPManagerRenderMatchesPi(t *testing.T) {
	raw, err := os.ReadFile("testdata/pi-mcp-manager.json")
	if err != nil {
		t.Fatal(err)
	}
	type piMenu struct {
		Title, Details, Error, Empty, Selected string
		ConfirmLabel, CancelLabel              string
		Items                                  []struct{ Value, Label, Description string }
	}
	var golden struct {
		Menus map[string]map[string]struct {
			Menu piMenu
			Rows []string
		}
		Status map[string][]string
		Signin map[string]struct {
			URL          string
			Empty, Typed []string
		}
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	for name, widths := range golden.Menus {
		for w, g := range widths {
			var width int
			fmt.Sscan(w, &width)
			menu := mcpMenu{title: g.Menu.Title, details: g.Menu.Details, errorText: g.Menu.Error, empty: g.Menu.Empty,
				confirmLabel: g.Menu.ConfirmLabel, cancelLabel: g.Menu.CancelLabel}
			selected := 0
			for i, it := range g.Menu.Items {
				menu.items = append(menu.items, slashItem{name: it.Label, value: it.Value, description: it.Description})
				if it.Value == g.Menu.Selected {
					selected = i
				}
			}
			if got := spanRowsText(mcpMenuRows(width, menu, selected)); strings.Join(got, "\n") != strings.Join(g.Rows, "\n") {
				t.Errorf("%s@%d:\n%s\nPi:\n%s", name, width, strings.Join(got, "\n"), strings.Join(g.Rows, "\n"))
			}
		}
	}
	for w, rows := range golden.Status {
		var width int
		fmt.Sscan(w, &width)
		if got := spanRowsText(mcpStatusRows(width, "MCP server github", "Reconnecting…")); strings.Join(got, "\n") != strings.Join(rows, "\n") {
			t.Errorf("status@%d:\n%s\nPi:\n%s", width, strings.Join(got, "\n"), strings.Join(rows, "\n"))
		}
	}
	for w, g := range golden.Signin {
		var width int
		fmt.Sscan(w, &width)
		s := &mcpSignInScreen{title: "Sign in to github", url: newAuthURL(g.URL)}
		for _, step := range []struct {
			input string
			rows  []string
		}{{"", g.Empty}, {"http://x", g.Typed}} {
			s.input = step.input
			if got := spanRowsText(s.rows(width)); strings.Join(got, "\n") != strings.Join(step.rows, "\n") {
				t.Errorf("signin %q@%d:\n%s\nPi:\n%s", step.input, width, strings.Join(got, "\n"), strings.Join(step.rows, "\n"))
			}
		}
	}
}

// The manager drives real servers: list (attention first), a server's
// actions, its tools, exposure saved to mcp.json, and disable/enable.
func TestMCPManagerManagesServers(t *testing.T) {
	fake := mcptest.New("ok")
	t.Cleanup(fake.Close)
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "mcp.json")
	body := fmt.Sprintf("{\n    \"note\": \"kept\",\n    \"mcpServers\": {\n        \"zeta\": {\"url\": %q, \"headers\": {\"Authorization\": \"Bearer t\"}},\n        \"broken\": {\"url\": \"http://127.0.0.1:1/mcp\", \"headers\": {\"Authorization\": \"Bearer t\"}}\n    }\n}\n", fake.URL)
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	c := sessionTestChat(t)
	c.uiQueue = make(chan func(), 64)
	c.engine.EnableMCPConfig(gimcp.LoadConfig(cfgPath, "", false), nil, "")
	drain := func(what string, cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(15 * time.Second)
		for !cond() {
			select {
			case fn := <-c.uiQueue:
				fn()
				continue
			default:
			}
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for %s:\n%s", what, strings.Join(spanRowsText(c.piMCPManagerRows(80)), "\n"))
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	state := func(name string) string {
		statuses, _ := c.engine.MCPStatus()
		st, _ := mcpStatusOf(statuses, name)
		return st.State
	}
	drain("startup", func() bool { return state("zeta") == gimcp.StateConnected && state("broken") == gimcp.StateFailed })
	if out := c.mcpCommand([]string{"/mcp"}); out != nil || !c.modelMenuOpen || c.modelMenuKind != "mcp-manager" {
		t.Fatal("/mcp did not open the manager")
	}
	key := func(k gotui.Key) {
		t.Helper()
		if !dispatchKey(c.KeyMap(), gotui.KeyEvent{Key: k}) {
			t.Fatalf("key %v not handled", k)
		}
	}
	rows := func() string { return strings.Join(spanRowsText(c.piMCPManagerRows(100)), "\n") }
	// Failed servers come first.
	if menu := c.mcpManagerMenu(); menu.items[0].value != "broken" || !strings.HasPrefix(menu.items[1].description, "connected · ") ||
		!strings.HasSuffix(menu.items[1].description, " · codemode · global") {
		t.Fatalf("servers: %+v", menu.items)
	}
	key(gotui.KeyDown)
	key(gotui.KeyEnter)
	// The project-owned scratch path may hard-wrap within a path component.
	// Compare the displayed path without layout whitespace, keeping the full path.
	compact := func(s string) string { return strings.Join(strings.Fields(s), "") }
	if !strings.Contains(rows(), "MCP server zeta") || !strings.Contains(rows(), fake.URL) || !strings.Contains(compact(rows()), compact("global: "+cfgPath)) {
		t.Fatalf("server menu:\n%s", rows())
	}
	labels := func() []string {
		var out []string
		for _, it := range c.mcpManagerMenu().items {
			out = append(out, it.value)
		}
		return out
	}
	if got := strings.Join(labels(), ","); got != "tools,reconnect,exposure,disable" {
		t.Fatalf("actions %s", got)
	}
	key(gotui.KeyEnter) // Tools
	if c.mcpManager.screen != "tools" || len(c.mcpManagerMenu().items) == 0 {
		t.Fatalf("tools:\n%s", rows())
	}
	key(gotui.KeyEscape)
	// Exposure: direct, saved to mcp.json with its other content.
	c.mcpManager.selected["server"] = "exposure"
	key(gotui.KeyEnter)
	if c.mcpManager.screen != "exposure" || c.mcpManagerMenu().items[0].name != "✓ codemode" {
		t.Fatalf("exposure:\n%s", rows())
	}
	key(gotui.KeyDown)
	key(gotui.KeyDown)
	key(gotui.KeyEnter)
	drain("exposure saved", func() bool {
		cfg, _ := c.engine.MCPServerConfig("zeta")
		return cfg.Exposure == gimcp.ExposureDirect && c.mcpManager.status[0] == ""
	})
	saved, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(saved), `"note": "kept"`) || !strings.Contains(string(saved), `"exposure": "direct"`) || !strings.Contains(string(saved), "\n    \"mcpServers\"") {
		t.Fatalf("mcp.json:\n%s", saved)
	}
	if tools := c.engine.MCPServerTools("zeta"); len(tools) == 0 || tools[0].Exposure != gimcp.ExposureDirect {
		t.Fatalf("tools not re-registered: %+v", tools)
	}
	// Disable, then enable again.
	c.mcpManager.selected["server"] = "disable"
	key(gotui.KeyEnter)
	drain("disabled", func() bool { return state("zeta") == gimcp.StateDisabled && c.mcpManager.status[0] == "" })
	if got := strings.Join(labels(), ","); got != "enable" {
		t.Fatalf("disabled actions %s", got)
	}
	if saved, _ := os.ReadFile(cfgPath); !strings.Contains(string(saved), `"enabled": false`) {
		t.Fatalf("mcp.json:\n%s", saved)
	}
	key(gotui.KeyEnter)
	drain("enabled", func() bool { return state("zeta") == gimcp.StateConnected && c.mcpManager.status[0] == "" })
	if saved, _ := os.ReadFile(cfgPath); strings.Contains(string(saved), `"enabled"`) {
		t.Fatalf("mcp.json:\n%s", saved)
	}
	key(gotui.KeyEscape)
	key(gotui.KeyEscape)
	if c.modelMenuOpen || c.mcpManager != nil {
		t.Fatal("manager not closed")
	}
}

// In a trusted project a global server can be turned off for the project
// alone (Pi 1.0.1): the manager saves a project override, lists the server
// as overridden and saves later changes to the override.
func TestMCPManagerProjectOverride(t *testing.T) {
	fake := mcptest.New("ok")
	t.Cleanup(fake.Close)
	dir := t.TempDir()
	user := filepath.Join(dir, "mcp.json")
	if err := os.WriteFile(user, []byte(fmt.Sprintf(`{"mcpServers": {"zeta": {"url": %q, "headers": {"Authorization": "Bearer t"}}}}`, fake.URL)), 0o600); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(dir, "work", ".pi", "mcp.json") // not created yet
	c := sessionTestChat(t)
	c.uiQueue = make(chan func(), 64)
	c.engine.EnableMCPConfig(gimcp.LoadConfig(user, project, true), nil, "")
	drain := func(what string, cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(15 * time.Second)
		for !cond() {
			select {
			case fn := <-c.uiQueue:
				fn()
				continue
			default:
			}
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for %s:\n%s", what, strings.Join(spanRowsText(c.piMCPManagerRows(80)), "\n"))
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	state := func() string {
		statuses, _ := c.engine.MCPStatus()
		st, _ := mcpStatusOf(statuses, "zeta")
		return st.State
	}
	drain("startup", func() bool { return state() == gimcp.StateConnected })
	c.mcpCommand([]string{"/mcp"})
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter})
	items := func() string {
		var out []string
		for _, it := range c.mcpManagerMenu().items {
			out = append(out, it.value+"="+it.description)
		}
		return strings.Join(out, ",")
	}
	if got := items(); !strings.HasSuffix(got, ",disable=saved to the global mcp.json,disable-project=saved to the project mcp.json") {
		t.Fatalf("actions %s", got)
	}
	c.mcpManager.selected["server"] = "disable-project"
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter})
	drain("disabled", func() bool { return state() == gimcp.StateDisabled && c.mcpManager.status[0] == "" })
	if saved, _ := os.ReadFile(project); string(saved) != "{\n  \"mcpServers\": {\n    \"zeta\": {\n      \"enabled\": false\n    }\n  }\n}\n" {
		t.Fatalf("project mcp.json:\n%s", saved)
	}
	if saved, _ := os.ReadFile(user); strings.Contains(string(saved), "enabled") {
		t.Fatalf("user mcp.json changed:\n%s", saved)
	}
	if got := items(); got != "enable=saved to the project mcp.json" {
		t.Fatalf("overridden actions %s", got)
	}
	if menu := c.mcpManagerMenu(); !strings.Contains(menu.details, "project override: "+project) {
		t.Fatalf("details %q", menu.details)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEnter}) // Enable, into the override
	drain("enabled", func() bool { return state() == gimcp.StateConnected && c.mcpManager.status[0] == "" })
	if saved, _ := os.ReadFile(project); !strings.Contains(string(saved), `"enabled": true`) {
		t.Fatalf("project mcp.json:\n%s", saved)
	}
	pressMenuKey(t, c.KeyMap(), gotui.KeyEvent{Key: gotui.KeyEscape})
	if menu := c.mcpManagerMenu(); !strings.HasSuffix(menu.items[0].description, " · codemode · global, project override") {
		t.Fatalf("servers %+v", menu.items)
	}
}
