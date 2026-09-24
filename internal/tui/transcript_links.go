package tui

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"

	gotui "github.com/grindlemire/go-tui"
)

// The Markdown projector displays explicit links as label (URL). Decorate only
// intact targets, never guess a destination from a clipped/wrapped URL fragment.
// Visible text and wrapping remain unchanged; the terminal owns activation.
var transcriptURL = regexp.MustCompile(`\((https?://[^()\s]+)\)`)

func safeTranscriptURL(raw string) bool {
	if len(raw) > 2048 || strings.Contains(raw, "\\") || strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return false
	}
	decoded, err := url.PathUnescape(raw)
	return err == nil && strings.IndexFunc(decoded, unicode.IsControl) < 0
}

func transcriptLinkSpans(text string, style gotui.Style) []gotui.TextSpan {
	var spans []gotui.TextSpan
	start := 0
	for _, match := range transcriptURL.FindAllStringSubmatchIndex(text, -1) {
		a, b := match[2], match[3]
		target := text[a:b]
		if !safeTranscriptURL(target) {
			continue
		}
		if a > start {
			spans = append(spans, gotui.TextSpan{Text: text[start:a], Style: style})
		}
		spans = append(spans, gotui.TextSpan{Text: target, Style: style, Link: target})
		start = b
	}
	if start < len(text) || len(spans) == 0 {
		spans = append(spans, gotui.TextSpan{Text: text[start:], Style: style})
	}
	return spans
}
