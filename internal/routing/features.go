package routing

import (
	"strings"
	"unicode/utf8"
)

const lookbackWindow = 6

type HistoryMessage struct {
	Payload map[string]any
}

type Features struct {
	TokenEstimate     int
	CodeBlockCount    int
	RecentToolCalls   int
	ConversationDepth int
	HasAttachments    bool
}

func ExtractFeatures(msg string, history []HistoryMessage) Features {
	return Features{
		TokenEstimate:     estimateTokens(msg),
		CodeBlockCount:    countCodeBlocks(msg),
		RecentToolCalls:   countRecentToolCalls(history),
		ConversationDepth: len(history),
		HasAttachments:    hasAttachments(msg),
	}
}

func estimateTokens(msg string) int {
	total := utf8.RuneCountInString(msg)
	if total == 0 {
		return 0
	}
	cjk := 0
	for _, r := range msg {
		if r >= 0x2E80 && r <= 0x9FFF || r >= 0xF900 && r <= 0xFAFF || r >= 0xAC00 && r <= 0xD7AF {
			cjk++
		}
	}
	return cjk + (total-cjk)/4
}

func countCodeBlocks(msg string) int { return strings.Count(msg, "```") / 2 }

func countRecentToolCalls(history []HistoryMessage) int {
	start := len(history) - lookbackWindow
	if start < 0 {
		start = 0
	}
	count := 0
	for _, msg := range history[start:] {
		if msg.Payload["source"] == "tool" || msg.Payload["kind"] == "tool" {
			count++
		}
	}
	return count
}

func hasAttachments(msg string) bool {
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "data:image/") || strings.Contains(lower, "data:audio/") || strings.Contains(lower, "data:video/") {
		return true
	}
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".mp3", ".wav", ".ogg", ".m4a", ".flac", ".mp4", ".avi", ".mov", ".webm"} {
		if strings.Contains(lower, ext) {
			return true
		}
	}
	return false
}
