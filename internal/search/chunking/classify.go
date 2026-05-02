package chunking

import "strings"

func ClassifyPath(path string) (kind, language string) {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown"):
		return "markdown", "markdown"
	case strings.HasSuffix(lower, ".go"):
		return "code", "go"
	case strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".ts"):
		return "code", "javascript"
	case strings.HasSuffix(lower, ".json"):
		return "json", "json"
	case strings.HasSuffix(lower, ".txt"):
		return "text", "text"
	default:
		return "text", ""
	}
}
