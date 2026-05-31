package config

import (
	"os"
	"path/filepath"
	"strings"
)

func DefaultWorkspaceRoot() string {
	if cwd, err := os.Getwd(); err == nil && strings.TrimSpace(cwd) != "" {
		return cwd
	}
	return "/workspace"
}

func DefaultStateDir(appName string) string {
	appName = strings.TrimSpace(appName)
	if appName == "" {
		appName = "gi"
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_STATE_HOME")); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.TrimSpace(home) != "" {
		return filepath.Join(home, ".local", "state", appName)
	}
	return filepath.Join(".", ".gi-run")
}

func DefaultTUIDBPath() string {
	return filepath.Join(DefaultStateDir("gi"), "gi.db")
}
