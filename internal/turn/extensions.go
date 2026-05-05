package turn

import (
	"context"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rcarmo/gi/internal/tools"
)

type extensionScript struct {
	Engine string
	Path   string
}

func discoverExtensionScripts(workspaceRoot string) []extensionScript {
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
	var out []extensionScript
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
			out = append(out, extensionScript{Engine: pattern.engine, Path: extPath})
		}
	}
	return out
}

func (e *Engine) loadWorkspaceExtensions(scriptTool *tools.ScriptTool) {
	for _, ext := range discoverExtensionScripts(e.runtimeCfg.WorkspaceRoot) {
		out := scriptTool.Execute(context.Background(), tools.ScriptInput{Engine: ext.Engine, Path: ext.Path})
		if out.Error != "" {
			log.Printf("extension load failed path=%s engine=%s: %s", ext.Path, ext.Engine, out.Error)
			e.recordExtension(ExtensionInfo{Engine: ext.Engine, Path: ext.Path, Status: "failed", Error: out.Error})
			continue
		}
		e.recordExtension(ExtensionInfo{Engine: ext.Engine, Path: ext.Path, Status: "loaded"})
		log.Printf("extension loaded path=%s engine=%s", ext.Path, ext.Engine)
	}
}
