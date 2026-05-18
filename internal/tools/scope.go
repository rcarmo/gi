package tools

import (
	"fmt"
	"strings"

	goai "github.com/rcarmo/go-ai"
)

func AppendToolNames(out []string, names ...string) []string {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}
func ToolNamesFromValue(v any) []string {
	switch raw := v.(type) {
	case []string:
		return NormalizeToolNames(AppendToolNames(nil, raw...))
	case []any:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			if name, ok := item.(string); ok {
				out = AppendToolNames(out, name)
			}
		}
		return NormalizeToolNames(out)
	case string:
		return NormalizeToolNames(AppendToolNames(nil, strings.Split(raw, ",")...))
	default:
		return nil
	}
}
func NormalizeToolNames(names []string) []string {
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
func ExplicitSubTurnToolNames(metadata map[string]any) ([]string, bool) {
	if metadata == nil {
		return nil, false
	}
	for _, key := range []string{"subturn_tools", "subturn_allowed_tools"} {
		if value, ok := metadata[key]; ok {
			return ToolNamesFromValue(value), true
		}
	}
	return nil, false
}
func EffectiveToolNamesFromMetadata(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	return ToolNamesFromValue(metadata["effective_tools"])
}
func EffectiveToolNameSetFromMetadata(metadata map[string]any) map[string]bool {
	names := EffectiveToolNamesFromMetadata(metadata)
	if len(names) == 0 {
		return nil
	}
	set := map[string]bool{}
	for _, n := range names {
		set[n] = true
	}
	return set
}
func ResolveEffectiveToolNames(parentMetadata, inputMetadata map[string]any, defaultEffective, allRegistered []string) ([]string, bool, error) {
	if parentMetadata == nil {
		return NormalizeToolNames(defaultEffective), false, nil
	}
	parentEffective := EffectiveToolNamesFromMetadata(parentMetadata)
	if len(parentEffective) == 0 {
		parentEffective = NormalizeToolNames(defaultEffective)
	}
	explicit, restricted := ExplicitSubTurnToolNames(inputMetadata)
	if !restricted {
		return parentEffective, false, nil
	}
	parentSet, registeredSet := map[string]bool{}, map[string]bool{}
	for _, n := range parentEffective {
		parentSet[n] = true
	}
	for _, n := range allRegistered {
		registeredSet[n] = true
	}
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
func ToolAllowedByMetadata(metadata map[string]any, toolName string) bool {
	allowedSet := EffectiveToolNameSetFromMetadata(metadata)
	if len(allowedSet) == 0 {
		return true
	}
	return allowedSet[toolName]
}
func ToolDefsForMetadata(metadata map[string]any, allEntries []RegisteredTool, defaultDefs []goai.Tool) []goai.Tool {
	allowedSet := EffectiveToolNameSetFromMetadata(metadata)
	if len(allowedSet) == 0 {
		return defaultDefs
	}
	defs := make([]goai.Tool, 0, len(allowedSet))
	for _, entry := range allEntries {
		if !allowedSet[entry.Name] {
			continue
		}
		defs = append(defs, entry.Definition())
	}
	return defs
}
