package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// Custom themes load, resolve and fail as Pi's own theme code does
// (scripts/golden-custom-theme.mjs): vars (also chained), hex, okhsl, oklch,
// 256-colour indexes, the terminal default, fallbacks and Pi's errors.
func TestCustomThemesMatchPi(t *testing.T) {
	raw, err := os.ReadFile("testdata/pi-custom-theme.json")
	if err != nil {
		t.Fatal(err)
	}
	var g struct {
		Files     map[string]json.RawMessage
		Available []string
		Themes    map[string]struct {
			Colors map[string]json.RawMessage
			Error  string
		}
	}
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for name, doc := range g.Files {
		content := []byte(doc)
		if name == "broken" {
			var s string
			_ = json.Unmarshal(doc, &s)
			content = []byte(s)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".json"), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0o644)
	previous := customThemeDirs
	customThemeDirs = func() []string { return []string{filepath.Join(t.TempDir(), "missing"), dir} }
	defer func() { customThemeDirs = previous }()

	if got := availablePiThemes(); !slices.Equal(got, g.Available) {
		t.Errorf("available %v, Pi %v", got, g.Available)
	}
	for name, want := range g.Themes {
		colors, err := loadCustomTheme(name)
		if want.Error != "" {
			got := ""
			if err != nil {
				got = err.Error()
			}
			// Pi's parse error ends with the JavaScript engine's message.
			if name == "broken" {
				want.Error, _, _ = strings.Cut(want.Error, "SyntaxError")
				if strings.HasPrefix(got, want.Error) {
					got = want.Error
				}
			}
			if got != want.Error {
				t.Errorf("%s: error\n%s\nPi:\n%s", name, got, want.Error)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for key, value := range want.Colors {
			c, ok := colors[key]
			var got any
			switch {
			case !ok:
				got = "missing"
			case c.isDef:
				got = "default"
			case c.isIndex:
				got = c.index
			default:
				got = []int{int(c.rgb.r), int(c.rgb.g), int(c.rgb.b)}
			}
			if gotJSON, _ := json.Marshal(got); string(gotJSON) != strings.Join(strings.Fields(string(value)), "") {
				t.Errorf("%s.%s = %s, Pi %s", name, key, gotJSON, value)
			}
		}
	}
}

// Editing the active custom theme's file applies it live (Pi's theme
// watcher); an invalid edit keeps the last good theme.
func TestActiveCustomThemeReloads(t *testing.T) {
	raw, err := os.ReadFile("testdata/pi-custom-theme.json")
	if err != nil {
		t.Fatal(err)
	}
	var g struct{ Files map[string]json.RawMessage }
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatal(err)
	}
	var theme map[string]any
	if err := json.Unmarshal(g.Files["ocean"], &theme); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "ocean.json")
	write := func(accent any) {
		theme["colors"].(map[string]any)["accent"] = accent
		data, _ := json.Marshal(theme)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("#0af")
	previous := customThemeDirs
	customThemeDirs = func() []string { return []string{dir} }
	c := sessionTestChat(t)
	c.uiQueue = make(chan func(), 16)
	t.Cleanup(func() {
		c.themeWatcher.stop()
		customThemeDirs = previous
		applyPiTheme("dark")
	})
	if err := setPiTheme("ocean"); err != nil {
		t.Fatal(err)
	}
	c.watchActiveTheme()
	if c.themeWatcher == nil {
		t.Fatal("no watcher for a custom theme")
	}
	waitAccent := func(want any) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for piAccent != want {
			select {
			case fn := <-c.uiQueue:
				fn()
			case <-time.After(20 * time.Millisecond):
			}
			if time.Now().After(deadline) {
				t.Fatalf("accent %v, want %v", piAccent, want)
			}
		}
	}
	write("#f00")
	waitAccent(piRGB(255, 0, 0))
	generation := piThemeGeneration
	write("not a colour")
	time.Sleep(3 * themeReloadDelay)
	for len(c.uiQueue) > 0 {
		(<-c.uiQueue)()
	}
	if piAccent != piRGB(255, 0, 0) || piThemeGeneration != generation {
		t.Fatalf("an invalid edit changed the theme")
	}
	// Built-in themes are not watched.
	_ = setPiTheme("light")
	c.watchActiveTheme()
	if c.themeWatcher != nil {
		t.Fatal("watching a built-in theme")
	}
}

// A queued reload must remain harmless after stop, even if the UI drains it later.
func TestThemeWatcherStopFencesQueuedReload(t *testing.T) {
	settingsThemeFixture(t)
	path := filepath.Join(customThemeDirs()[0], "ocean.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := setPiTheme("ocean"); err != nil {
		t.Fatal(err)
	}
	c := &chatTUI{}
	c.uiQueue = make(chan func(), 4)
	c.watchActiveTheme()
	t.Cleanup(func() { c.themeWatcher.stop() })
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	var reload func()
	select {
	case reload = <-c.uiQueue:
	case <-time.After(3 * time.Second):
		t.Fatal("theme reload was not queued")
	}
	c.themeWatcher.stop()
	c.themeWatcher.stop() // shutdown is idempotent
	applyPiTheme("dark")
	generation := piThemeGeneration
	reload()
	if piThemeGeneration != generation || piActiveTheme != "dark" {
		t.Fatal("stopped watcher applied a queued reload")
	}
}
