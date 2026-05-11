package turn

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

func stringValue(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func boolValue(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		s = strings.ToLower(strings.TrimSpace(s))
		return s == "true" || s == "1" || s == "yes"
	}
	return false
}

func boolValueOr(v any, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	if v == nil {
		return fallback
	}
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			return fallback
		}
		return s == "true" || s == "1" || s == "yes"
	}
	switch num := v.(type) {
	case float64:
		return num != 0
	case int:
		return num != 0
	case int64:
		return num != 0
	case int32:
		return num != 0
	case uint:
		return num != 0
	case uint64:
		return num != 0
	case uint32:
		return num != 0
	default:
		return fallback
	}
}

func intValueOr(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case uint:
		return int(n)
	case uint64:
		return int(n)
	case uint32:
		return int(n)
	case float64:
		return int(n)
	case string:
		n = strings.TrimSpace(n)
		if n == "" {
			return fallback
		}
		if parsed, err := strconv.Atoi(n); err == nil {
			return parsed
		}
	}
	return fallback
}

func normalizeSubTurnDeliveryMode(mode string) (string, error) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return "sync", nil
	}
	switch mode {
	case "sync", "async":
		return mode, nil
	default:
		return "", fmt.Errorf("invalid subturn delivery mode: %s", mode)
	}
}

func SortQueuedTurns(turns []store.Turn) {
	sort.SliceStable(turns, func(i, j int) bool { return turns[i].CreatedAt < turns[j].CreatedAt })
}
