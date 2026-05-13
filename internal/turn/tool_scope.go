package turn

import (
	"fmt"
	"strings"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func appendToolNames(out []string, names ...string) []string {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func toolNamesFromValue(v any) []string {
	switch raw := v.(type) {
	case []string:
		return normalizeToolNames(appendToolNames(nil, raw...))
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if name, ok := item.(string); ok {
				out = appendToolNames(out, name)
			}
		}
		return normalizeToolNames(out)
	case string:
		return normalizeToolNames(appendToolNames(nil, strings.Split(raw, ",")...))
	default:
		return nil
	}
}

func normalizeToolNames(names []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func explicitSubTurnToolNames(metadata map[string]any) ([]string, bool) {
	if metadata == nil {
		return nil, false
	}
	for _, key := range []string{"subturn_tools", "subturn_allowed_tools"} {
		value, ok := metadata[key]
		if !ok {
			continue
		}
		return toolNamesFromValue(value), true
	}
	return nil, false
}

func toolNameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func effectiveToolNamesFromMetadata(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	return toolNamesFromValue(metadata["effective_tools"])
}

func effectiveToolNameSetFromMetadata(metadata map[string]any) map[string]bool {
	names := effectiveToolNamesFromMetadata(metadata)
	if len(names) == 0 {
		return nil
	}
	return toolNameSet(names)
}

func (e *Engine) allRegisteredToolNames() []string {
	entries := e.tools.AllEntries()
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name)
	}
	return normalizeToolNames(out)
}

func (e *Engine) defaultEffectiveToolNames() []string {
	return normalizeToolNames(e.ActiveTools())
}

func (e *Engine) resolveEffectiveToolNames(parentTurn *store.Turn, inputMetadata map[string]any) ([]string, bool, error) {
	if parentTurn == nil {
		return e.defaultEffectiveToolNames(), false, nil
	}
	parentEffective := effectiveToolNamesFromMetadata(parentTurn.Metadata)
	if len(parentEffective) == 0 {
		parentEffective = e.defaultEffectiveToolNames()
	}
	explicit, restricted := explicitSubTurnToolNames(inputMetadata)
	if !restricted {
		return parentEffective, false, nil
	}
	parentSet := toolNameSet(parentEffective)
	registeredSet := toolNameSet(e.allRegisteredToolNames())
	for _, name := range explicit {
		if !registeredSet[name] {
			return nil, false, fmt.Errorf("unknown subturn tool: %s", name)
		}
		if !parentSet[name] {
			return nil, false, fmt.Errorf("subturn tools must be a subset of parent effective tools: %s", name)
		}
	}
	return explicit, true, nil
}

func toolAllowedByMetadata(metadata map[string]any, toolName string) bool {
	allowedSet := effectiveToolNameSetFromMetadata(metadata)
	if len(allowedSet) == 0 {
		return true
	}
	return allowedSet[toolName]
}

func (e *Engine) toolDefsForMetadata(metadata map[string]any) []goai.Tool {
	allowedSet := effectiveToolNameSetFromMetadata(metadata)
	if len(allowedSet) == 0 {
		return e.toolDefs()
	}
	entries := e.tools.AllEntries()
	defs := make([]goai.Tool, 0, len(allowedSet))
	for _, entry := range entries {
		if !allowedSet[entry.Name] {
			continue
		}
		defs = append(defs, entry.Definition())
	}
	return defs
}
