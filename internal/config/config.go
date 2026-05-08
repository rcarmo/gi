package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/routing"
	"github.com/rcarmo/gi/internal/skills"
)

type RuntimeConfig struct {
	WorkspaceRoot        string                     `json:"workspace_root"`
	AssistantName        string                     `json:"assistant_name"`
	AssistantAvatar      string                     `json:"assistant_avatar"`
	UserName             string                     `json:"user_name"`
	UserAvatar           string                     `json:"user_avatar"`
	UserAvatarBackground string                     `json:"user_avatar_background"`
	DefaultProvider      string                     `json:"default_provider"`
	DefaultModel         string                     `json:"default_model"`
	DefaultThinkingLevel string                     `json:"default_thinking_level"`
	EnabledModels        []string                   `json:"enabled_models"`
	Agents               routing.AgentsConfig       `json:"agents"`
	Session              routing.SessionConfig      `json:"session"`
	Routing              routing.ModelRoutingConfig `json:"routing"`
	MaxIterations        int                        `json:"max_iterations"`
	ScrollbackLimit      int                        `json:"scrollback_limit"`
	Compaction           CompactionSettings         `json:"compaction"`
	Peering              PeeringSettings            `json:"peering"`
	SystemPrompt         string                     `json:"-"`
	Discovery            skills.Discovery           `json:"-"`
}

type piclawConfig struct {
	Assistant struct {
		AssistantName   string `json:"assistantName"`
		AssistantAvatar string `json:"assistantAvatar"`
	} `json:"assistant"`
	User struct {
		UserName             string `json:"userName"`
		UserAvatar           string `json:"userAvatar"`
		UserAvatarBackground string `json:"userAvatarBackground"`
	} `json:"user"`
}

type CompactionSettings struct {
	Enabled          bool   `json:"enabled"`
	ContextWindow    int    `json:"context_window"`
	ReserveTokens    int    `json:"reserve_tokens"`
	KeepRecentTokens int    `json:"keep_recent_tokens"`
	ThresholdTokens  int    `json:"threshold_tokens"`
	Strategy         string `json:"strategy"`
}

type PeeringSettings struct {
	Enabled         bool   `json:"enabled"`
	Hostname        string `json:"hostname"`
	StateDir        string `json:"state_dir"`
	AuthKeyEnv      string `json:"auth_key_env"`
	AuthKeyKeychain string `json:"auth_key_keychain"`
}

type piSettings struct {
	DefaultProvider      string                     `json:"defaultProvider"`
	DefaultModel         string                     `json:"defaultModel"`
	DefaultThinkingLevel string                     `json:"defaultThinkingLevel"`
	EnabledModels        []string                   `json:"enabledModels"`
	MaxIterations        int                        `json:"maxIterations"`
	TUIScrollbackLimit   int                        `json:"tuiScrollbackLimit"`
	Compaction           CompactionSettings         `json:"compaction"`
	Peering              PeeringSettings            `json:"peering"`
	Agents               routing.AgentsConfig       `json:"agents"`
	Session              routing.SessionConfig      `json:"session"`
	Routing              routing.ModelRoutingConfig `json:"routing"`
}

func Load(workspaceRoot string) RuntimeConfig {
	cfg := RuntimeConfig{WorkspaceRoot: workspaceRoot, Compaction: CompactionSettings{Enabled: true}}
	var pc piclawConfig
	if err := readJSON(filepath.Join(workspaceRoot, ".piclaw", "config.json"), &pc); err == nil {
		cfg.AssistantName = pc.Assistant.AssistantName
		cfg.AssistantAvatar = pc.Assistant.AssistantAvatar
		cfg.UserName = pc.User.UserName
		cfg.UserAvatar = pc.User.UserAvatar
		cfg.UserAvatarBackground = pc.User.UserAvatarBackground
	}
	var ps piSettings
	if err := readJSON(filepath.Join(workspaceRoot, ".pi", "settings.json"), &ps); err == nil {
		cfg.DefaultProvider = ps.DefaultProvider
		cfg.DefaultModel = ps.DefaultModel
		cfg.DefaultThinkingLevel = ps.DefaultThinkingLevel
		cfg.EnabledModels = append([]string(nil), ps.EnabledModels...)
		cfg.MaxIterations = ps.MaxIterations
		cfg.ScrollbackLimit = ps.TUIScrollbackLimit
		cfg.Compaction = ps.Compaction
		cfg.Peering = ps.Peering
		cfg.Agents = ps.Agents
		cfg.Session = ps.Session
		cfg.Routing = ps.Routing
	}
	if discovery, err := skills.Discover(workspaceRoot); err == nil {
		cfg.Discovery = discovery
	}
	if cfg.AssistantName == "" {
		cfg.AssistantName = "Gi"
	}
	if cfg.UserName == "" {
		cfg.UserName = "User"
	}
	if strings.TrimSpace(cfg.DefaultProvider) == "" {
		cfg.DefaultProvider = "opencode-zen"
	}
	if len(cfg.EnabledModels) == 0 {
		cfg.EnabledModels = []string{"opencode-zen/minimax-m2.5-free"}
	}
	if cfg.DefaultModel == "" && len(cfg.EnabledModels) > 0 {
		cfg.DefaultModel = cfg.EnabledModels[0]
	}
	if strings.TrimSpace(cfg.DefaultThinkingLevel) == "" {
		cfg.DefaultThinkingLevel = "low"
	}
	if len(cfg.Session.Dimensions) == 0 {
		cfg.Session.Dimensions = []string{"chat"}
	}
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 64
	}
	if cfg.ScrollbackLimit <= 0 {
		cfg.ScrollbackLimit = 1000
	}
	applyCompactionDefaults(&cfg.Compaction)
	if len(cfg.Agents.List) == 0 {
		cfg.Agents.List = []routing.AgentConfig{{ID: "agent", Name: cfg.AssistantName, Default: true, Model: cfg.DefaultModel}}
	}
	// Load workspace instructions from AGENTS.md and wrap them in gi's runtime prompt.
	workspaceInstructions := ""
	agentsPath := filepath.Join(workspaceRoot, "AGENTS.md")
	if data, err := os.ReadFile(agentsPath); err == nil && len(data) > 0 {
		workspaceInstructions = string(data)
	}
	cfg.SystemPrompt = buildSystemPrompt(cfg, workspaceInstructions)
	return cfg
}

func PersistModelSelection(workspaceRoot, provider, model, thinking string, enabledModels []string) error {
	if strings.TrimSpace(workspaceRoot) == "" {
		return errors.New("workspace root is required")
	}
	piDir := filepath.Join(workspaceRoot, ".pi")
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		return err
	}
	settingsPath := filepath.Join(piDir, "settings.json")
	settings := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("decode settings.json: %w", err)
		}
	}
	if strings.TrimSpace(provider) != "" {
		settings["defaultProvider"] = provider
	}
	if strings.TrimSpace(model) != "" {
		settings["defaultModel"] = model
	}
	if strings.TrimSpace(thinking) != "" {
		settings["defaultThinkingLevel"] = thinking
	}
	models := append([]string(nil), enabledModels...)
	if strings.TrimSpace(model) != "" && !contains(models, model) {
		models = append(models, model)
	}
	if len(models) > 0 {
		settings["enabledModels"] = models
	}
	blob, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	blob = append(blob, '\n')
	return os.WriteFile(settingsPath, blob, 0o644)
}

func PersistScrollbackLimit(workspaceRoot string, limit int) error {
	if strings.TrimSpace(workspaceRoot) == "" {
		return errors.New("workspace root is required")
	}
	if limit <= 0 {
		return errors.New("scrollback limit must be > 0")
	}
	piDir := filepath.Join(workspaceRoot, ".pi")
	if err := os.MkdirAll(piDir, 0o755); err != nil {
		return err
	}
	settingsPath := filepath.Join(piDir, "settings.json")
	settings := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("decode settings.json: %w", err)
		}
	}
	settings["tuiScrollbackLimit"] = limit
	blob, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	blob = append(blob, '\n')
	return os.WriteFile(settingsPath, blob, 0o644)
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty file")
	}
	return json.Unmarshal(data, target)
}
