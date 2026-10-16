package web

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	giui "github.com/rcarmo/gi/references/fixtures-vibes/ui/classic"
	"net/http"
	"regexp"
	"strings"
)

// /theme and /tint follow Piclaw 3.2.5's ui-theme-commands.ts: they change
// the UI only, never reach the model, and reply in the timeline. Gi keeps
// appearance in the browser, so the client applies the returned payload.

var themeCatalogueJSON = giui.ThemeCatalogue

type themePreset struct {
	Name  string            `json:"name"`
	Label string            `json:"label"`
	Mode  string            `json:"mode"`
	Light map[string]string `json:"light"`
	Dark  map[string]string `json:"dark"`
}

type themePayload struct {
	Theme string  `json:"theme"`
	Tint  *string `json:"tint"`
}

type themeCommandResult struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Payload *themePayload `json:"payload,omitempty"`
}

var themePresets = func() []themePreset {
	var presets []themePreset
	if err := json.Unmarshal(themeCatalogueJSON, &presets); err != nil {
		panic(fmt.Sprintf("theme catalogue: %v", err))
	}
	return presets
}()

// Piclaw's WEB_THEME_ALIASES, kept only where Gi has the target preset.
var piclawThemeAliases = map[string]string{
	"auto": "default", "drac": "dracula", "catpp": "catppuccin", "catppuccin-mocha": "catppuccin",
	"gruv": "gruvbox", "gruvbox-dark": "gruvbox", "tokyo-night": "tokyo", "catpuccin": "catppuccin",
}

var themeListColorKeys = []string{"bgPrimary", "bgSecondary", "textPrimary", "textSecondary", "borderColor", "accent", "danger", "success"}

var themeFallbackPalette = map[string]string{
	"bgPrimary": "#ffffff", "bgSecondary": "#f7f9fa", "textPrimary": "#0f1419", "textSecondary": "#536471",
	"borderColor": "#eff3f4", "accent": "#1d9bf0", "danger": "#f4212e", "success": "#00ba7c",
}

var themeClearValues = map[string]bool{"off": true, "clear": true, "none": true, "reset": true, "default": true}

var (
	hexTintPattern   = regexp.MustCompile(`^[0-9a-fA-F]{3}$|^[0-9a-fA-F]{6}$`)
	namedTintPattern = regexp.MustCompile(`^[a-z]+$`)
	swatchHexPattern = regexp.MustCompile(`^#[0-9a-fA-F]{3}$|^#[0-9a-fA-F]{6}$`)
)

func normalizeThemeName(input string) (string, bool) {
	raw := strings.ToLower(strings.TrimSpace(input))
	if raw == "" {
		return "", false
	}
	if target, ok := piclawThemeAliases[raw]; ok {
		raw = target
	}
	for _, p := range themePresets {
		if p.Name == raw {
			return raw, true
		}
	}
	return "", false
}

func themeLabel(name string) string {
	for _, p := range themePresets {
		if p.Name == name && p.Label != "" {
			return p.Label
		}
	}
	return name
}

func normalizeTint(input string) (string, bool) {
	raw := strings.TrimSpace(input)
	hex := strings.TrimPrefix(raw, "#")
	if hexTintPattern.MatchString(hex) {
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		return "#" + strings.ToLower(hex), true
	}
	named := strings.ToLower(raw)
	if namedTintPattern.MatchString(named) {
		return named, true
	}
	return "", false
}

func themeSwatch(color string) string {
	if !swatchHexPattern.MatchString(color) {
		color = "#000000"
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20" role="img" aria-hidden="true">
      <rect x="0.5" y="0.5" width="19" height="19" rx="3" ry="3" fill="%s" />
    </svg>`, strings.ToLower(color))
	return "![](data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg)) + ")"
}

func formatThemeList() string {
	lines := []string{"Available themes:", "| Theme | Mode | Swatches |", "| --- | --- | --- |"}
	for _, p := range themePresets {
		source := p.Light
		switch p.Mode {
		case "dark":
			source = p.Dark
		case "light":
		default:
			if source == nil {
				source = p.Dark
			}
		}
		mode := p.Mode
		if mode == "auto" {
			mode = "auto (light)"
		}
		swatches := make([]string, 0, len(themeListColorKeys))
		for _, key := range themeListColorKeys {
			color := themeFallbackPalette[key]
			if v := strings.TrimSpace(source[key]); v != "" {
				color = v
			}
			swatches = append(swatches, themeSwatch(color))
		}
		lines = append(lines, fmt.Sprintf("| %s (%s) | %s | %s |", p.Label, p.Name, mode, strings.Join(swatches, " ")))
	}
	return strings.Join(append(lines, "", "Usage: /theme <name>", "(omit name to show this list)"), "\n")
}

// themeCommand returns nil when prompt is not /theme or /tint.
func themeCommand(prompt string) *themeCommandResult {
	fields := strings.Fields(strings.TrimSpace(prompt))
	if len(fields) == 0 {
		return nil
	}
	args := strings.Join(fields[1:], " ")
	switch strings.ToLower(fields[0]) {
	case "/theme":
		if args == "" {
			return &themeCommandResult{Status: "success", Message: formatThemeList()}
		}
		name, ok := normalizeThemeName(args)
		if !ok {
			return &themeCommandResult{Status: "error", Message: fmt.Sprintf("Unknown theme: %s. Omit the name to show available themes.", args)}
		}
		return &themeCommandResult{Status: "success", Message: fmt.Sprintf("Theme set to %s.", themeLabel(name)), Payload: &themePayload{Theme: name}}
	case "/tint":
		if args == "" {
			return &themeCommandResult{Status: "error", Message: "Usage: /tint #hex (e.g. /tint #3b82f6), /tint orange, or /tint off"}
		}
		if themeClearValues[strings.ToLower(args)] {
			return &themeCommandResult{Status: "success", Message: "Tint cleared (default light/dark restored).", Payload: &themePayload{Theme: "default"}}
		}
		tint, ok := normalizeTint(args)
		if !ok {
			return &themeCommandResult{Status: "error", Message: fmt.Sprintf("Invalid tint value: %s. Use a hex color (e.g. #3b82f6), a named color (e.g. orange), or /tint off.", args)}
		}
		return &themeCommandResult{Status: "success", Message: fmt.Sprintf("Tint set to %s.", tint), Payload: &themePayload{Theme: "default", Tint: &tint}}
	}
	return nil
}

func (s *Server) handleThemeCommand(w http.ResponseWriter, r *http.Request, sessionID, prompt string) bool {
	result := themeCommand(prompt)
	if result == nil {
		return false
	}
	if _, err := s.store.GetSession(r.Context(), sessionID); err != nil {
		writeJSON(w, 404, map[string]any{"error": err.Error()})
		return true
	}
	if _, err := s.turns.PostSystemMessage(r.Context(), sessionID, result.Message, map[string]any{"kind": "ui_theme", "command": prompt}); err != nil {
		writeJSON(w, 500, map[string]any{"error": "Could not persist appearance command result"})
		return true
	}
	writeJSON(w, 200, map[string]any{"thread_id": nil, "ui_only": true, "source_committed": true, "command": result})
	return true
}
