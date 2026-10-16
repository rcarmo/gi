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

// pi-tui's Editor and CombinedAutocompleteProvider replayed with Pi's /mcp
// getArgumentCompletions: bun scripts/golden-slash-autocomplete.mjs. The
// golden's command catalogue is smaller than gi's, so command lists must
// contain Pi's matches in Pi's order; argument lists must match exactly.
func TestSlashArgumentCompletionMatchesPi(t *testing.T) {
	raw, err := os.ReadFile("testdata/pi-slash-autocomplete.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden map[string][]struct {
		Step         string
		Text         string
		Submitted    bool
		Open         bool
		Items        []string
		Descriptions []string
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatal(err)
	}
	oauth := mcptest.NewOAuth(t)
	local := mcptest.New("local")
	t.Cleanup(local.Close)
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "mcp.json")
	body := fmt.Sprintf(`{"mcpServers": {"github": {"url": %q}, "local": {"url": %q, "headers": {"Authorization": "Bearer t"}}}}`, oauth.URL+"/mcp", local.URL)
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	c := sessionTestChat(t)
	c.engine.EnableMCPConfig(gimcp.LoadConfig(cfgPath, "", false), gimcp.NewCredentialStore(filepath.Join(dir, "mcp-auth.json")), "")
	describe := map[string]string{}
	deadline := time.Now().Add(10 * time.Second)
	for {
		statuses, _ := c.engine.MCPStatus()
		for _, st := range statuses {
			describe[st.Name] = gimcp.DescribeState(st)
		}
		// Connected precedes the asynchronous tools/list refresh. Wait for
		// the fixture's four tools so descriptions cannot change mid-replay.
		if describe["github"] == "needs sign-in" && describe["local"] == "connected · 4 tools" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("servers not ready: %v", describe)
		}
		time.Sleep(20 * time.Millisecond)
	}
	for name, states := range golden {
		t.Run(name, func(t *testing.T) {
			c.slash = slashMenu{}
			c.input.SetText("")
			var submitted string
			c.input.onSubmit = func(s string) { submitted = s }
			for _, want := range states {
				submitted = ""
				switch want.Step {
				case "tab", "enter", "esc":
					key := map[string]gotui.Key{"tab": gotui.KeyTab, "enter": gotui.KeyEnter, "esc": gotui.KeyEscape}[want.Step]
					if !c.handleSlashKey(key) && key == gotui.KeyEnter {
						c.input.enter(gotui.KeyEvent{Key: gotui.KeyEnter})
					}
				case "backspace":
					c.input.backspace()
				default:
					for _, r := range want.Step {
						c.input.insertRune(gotui.KeyEvent{Key: gotui.KeyRune, Rune: r})
					}
				}
				text := c.input.Text()
				if want.Submitted {
					text = strings.TrimSpace(submitted)
				}
				if text != want.Text || (submitted != "") != want.Submitted || c.slash.active != want.Open {
					t.Fatalf("after %q: text %q submitted %v open %v; Pi: %q %v %v", want.Step, text, submitted != "", c.slash.active, want.Text, want.Submitted, want.Open)
				}
				if !want.Open {
					continue
				}
				var labels, descriptions []string
				for _, item := range c.slash.items {
					labels = append(labels, item.name)
					descriptions = append(descriptions, item.description)
				}
				if !c.slash.argument {
					if !isSubsequence(want.Items, labels) {
						t.Fatalf("after %q: commands %q lack Pi's %q in order", want.Step, labels, want.Items)
					}
					continue
				}
				expected := make([]string, len(want.Items))
				for i, label := range want.Items {
					if want.Descriptions[i] != "" {
						expected[i] = describe[label] // Pi's describeState of gi's server
					}
				}
				if strings.Join(labels, "|") != strings.Join(want.Items, "|") || strings.Join(descriptions, "|") != strings.Join(expected, "|") {
					t.Fatalf("after %q: arguments %q %q; Pi: %q %q", want.Step, labels, descriptions, want.Items, expected)
				}
			}
		})
	}
}

func isSubsequence(want, got []string) bool {
	i := 0
	for _, g := range got {
		if i < len(want) && g == want[i] {
			i++
		}
	}
	return i == len(want)
}
