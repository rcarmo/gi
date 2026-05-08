package turn

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func toolNamesFromValue(v any) []string {
	switch raw := v.(type) {
	case []string:
		out := make([]string, 0, len(raw))
		for _, name := range raw {
			name = strings.TrimSpace(name)
			if name != "" {
				out = append(out, name)
			}
		}
		return normalizeToolNames(out)
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if name, ok := item.(string); ok {
				name = strings.TrimSpace(name)
				if name != "" {
					out = append(out, name)
				}
			}
		}
		return normalizeToolNames(out)
	case string:
		parts := strings.Split(raw, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			name := strings.TrimSpace(part)
			if name != "" {
				out = append(out, name)
			}
		}
		return normalizeToolNames(out)
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
	parentEffective := toolNamesFromValue(parentTurn.Metadata["effective_tools"])
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
	allowed := toolNamesFromValue(metadata["effective_tools"])
	if len(allowed) == 0 {
		return true
	}
	return toolNameSet(allowed)[toolName]
}

func (e *Engine) toolDefsForMetadata(metadata map[string]any) []goai.Tool {
	allowed := toolNamesFromValue(metadata["effective_tools"])
	if len(allowed) == 0 {
		return e.toolDefs()
	}
	allowedSet := toolNameSet(allowed)
	entries := e.tools.AllEntries()
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	defs := make([]goai.Tool, 0, len(allowed))
	for _, entry := range entries {
		if !allowedSet[entry.Name] {
			continue
		}
		defs = append(defs, entry.Definition())
	}
	return defs
}
