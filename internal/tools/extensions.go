package tools

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
)

type ExtensionScript struct {
	Engine string
	Path   string
}

type ExtensionLoadResult struct {
	Engine string
	Path   string
	Error  string
}

func DiscoverExtensionScripts(workspaceRoot string) []ExtensionScript {
	if strings.TrimSpace(workspaceRoot) == "" {
		return nil
	}
	patterns := []struct {
		root   string
		engine string
		glob   string
	}{
		{filepath.Join(workspaceRoot, ".gi", "extensions"), "js", "*.js"},
		{filepath.Join(workspaceRoot, ".gi", "extensions"), "joker", "*.joke"},
		{filepath.Join(workspaceRoot, ".pi", "extensions"), "js", "*.js"},
		{filepath.Join(workspaceRoot, ".pi", "extensions"), "joker", "*.joke"},
	}
	var out []ExtensionScript
	seen := map[string]bool{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(pattern.root, pattern.glob))
		if err != nil {
			continue
		}
		sort.Strings(matches)
		for _, match := range matches {
			if seen[match] {
				continue
			}
			seen[match] = true
			extPath := match
			if rel, err := filepath.Rel(workspaceRoot, match); err == nil && !strings.HasPrefix(rel, "..") {
				extPath = rel
			}
			out = append(out, ExtensionScript{Engine: pattern.engine, Path: extPath})
		}
	}
	return out
}

func LoadWorkspaceExtensions(ctx context.Context, workspaceRoot string, scriptTool *ScriptTool) []ExtensionLoadResult {
	var out []ExtensionLoadResult
	for _, ext := range DiscoverExtensionScripts(workspaceRoot) {
		res := ExtensionLoadResult{Engine: ext.Engine, Path: ext.Path}
		r := scriptTool.Execute(ctx, ScriptInput{Engine: ext.Engine, Path: ext.Path})
		if r.Error != "" {
			res.Error = r.Error
		}
		out = append(out, res)
	}
	return out
}
