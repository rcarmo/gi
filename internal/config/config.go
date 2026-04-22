package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type RuntimeConfig struct {
	WorkspaceRoot        string   `json:"workspace_root"`
	AssistantName        string   `json:"assistant_name"`
	AssistantAvatar      string   `json:"assistant_avatar"`
	UserName             string   `json:"user_name"`
	UserAvatar           string   `json:"user_avatar"`
	DefaultProvider      string   `json:"default_provider"`
	DefaultModel         string   `json:"default_model"`
	DefaultThinkingLevel string   `json:"default_thinking_level"`
	EnabledModels        []string `json:"enabled_models"`
}

type piclawConfig struct {
	Assistant struct {
		AssistantName   string `json:"assistantName"`
		AssistantAvatar string `json:"assistantAvatar"`
	} `json:"assistant"`
	User struct {
		UserName   string `json:"userName"`
		UserAvatar string `json:"userAvatar"`
	} `json:"user"`
}

type piSettings struct {
	DefaultProvider      string   `json:"defaultProvider"`
	DefaultModel         string   `json:"defaultModel"`
	DefaultThinkingLevel string   `json:"defaultThinkingLevel"`
	EnabledModels        []string `json:"enabledModels"`
}

func Load(workspaceRoot string) RuntimeConfig {
	cfg := RuntimeConfig{WorkspaceRoot: workspaceRoot}
	var pc piclawConfig
	if err := readJSON(filepath.Join(workspaceRoot, ".piclaw", "config.json"), &pc); err == nil {
		cfg.AssistantName = pc.Assistant.AssistantName
		cfg.AssistantAvatar = pc.Assistant.AssistantAvatar
		cfg.UserName = pc.User.UserName
		cfg.UserAvatar = pc.User.UserAvatar
	}
	var ps piSettings
	if err := readJSON(filepath.Join(workspaceRoot, ".pi", "settings.json"), &ps); err == nil {
		cfg.DefaultProvider = ps.DefaultProvider
		cfg.DefaultModel = ps.DefaultModel
		cfg.DefaultThinkingLevel = ps.DefaultThinkingLevel
		cfg.EnabledModels = append([]string(nil), ps.EnabledModels...)
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
