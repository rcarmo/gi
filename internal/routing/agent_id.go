package routing

import (
	"regexp"
	"strings"
)

const (
	DefaultAgentID   = "main"
	DefaultAccountID = "default"
	MaxAgentIDLength = 64
)

var (
	validIDRe      = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
	invalidCharsRe = regexp.MustCompile(`[^a-z0-9_-]+`)
	leadingDashRe  = regexp.MustCompile(`^-+`)
	trailingDashRe = regexp.MustCompile(`-+$`)
)

func NormalizeAgentID(id string) string {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return DefaultAgentID
	}
	lower := strings.ToLower(trimmed)
	if validIDRe.MatchString(lower) {
		return lower
	}
	result := invalidCharsRe.ReplaceAllString(lower, "-")
	result = leadingDashRe.ReplaceAllString(result, "")
	result = trailingDashRe.ReplaceAllString(result, "")
	if len(result) > MaxAgentIDLength {
		result = result[:MaxAgentIDLength]
	}
	if result == "" {
		return DefaultAgentID
	}
	return result
}

func NormalizeAccountID(id string) string {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return DefaultAccountID
	}
	lower := strings.ToLower(trimmed)
	if validIDRe.MatchString(lower) {
		return lower
	}
	result := invalidCharsRe.ReplaceAllString(lower, "-")
	result = leadingDashRe.ReplaceAllString(result, "")
	result = trailingDashRe.ReplaceAllString(result, "")
	if len(result) > MaxAgentIDLength {
		result = result[:MaxAgentIDLength]
	}
	if result == "" {
		return DefaultAccountID
	}
	return result
}
