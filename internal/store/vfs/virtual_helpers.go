package vfs

import (
	"encoding/json"
	"sort"
	"strings"
)

func RenderFrontmatterMarkdown(meta map[string]any, body string) []byte {
	if meta == nil {
		meta = map[string]any{}
	}
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString("---\n")
	for _, k := range keys {
		raw, _ := json.Marshal(meta[k])
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.Write(raw)
		sb.WriteString("\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

func SplitVirtualPath(p string) []string {
	trimmed := strings.Trim(strings.TrimSpace(p), "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}
