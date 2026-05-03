package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

type piSettings struct {
	DefaultProvider      string                     `json:"defaultProvider"`
	DefaultModel         string                     `json:"defaultModel"`
	DefaultThinkingLevel string                     `json:"defaultThinkingLevel"`
	EnabledModels        []string                   `json:"enabledModels"`
	MaxIterations        int                        `json:"maxIterations"`
	Agents               routing.AgentsConfig       `json:"agents"`
	Session              routing.SessionConfig      `json:"session"`
	Routing              routing.ModelRoutingConfig `json:"routing"`
}

func Load(workspaceRoot string) RuntimeConfig {
	cfg := RuntimeConfig{WorkspaceRoot: workspaceRoot}
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
	if cfg.DefaultModel == "" && len(cfg.EnabledModels) > 0 {
		cfg.DefaultModel = cfg.EnabledModels[0]
	}
	if len(cfg.Session.Dimensions) == 0 {
		cfg.Session.Dimensions = []string{"chat"}
	}
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 64
	}
	if len(cfg.Agents.List) == 0 {
		cfg.Agents.List = []routing.AgentConfig{{ID: "agent", Name: cfg.AssistantName, Default: true, Model: cfg.DefaultModel}}
	}
	// Load system prompt from AGENTS.md
	agentsPath := filepath.Join(workspaceRoot, "AGENTS.md")
	if data, err := os.ReadFile(agentsPath); err == nil && len(data) > 0 {
		cfg.SystemPrompt = string(data)
	} else {
		cfg.SystemPrompt = fmt.Sprintf("You are %s, a helpful coding assistant.", cfg.AssistantName)
	}
	cfg.SystemPrompt += skills.PromptSummary(cfg.Discovery)
	return cfg
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
