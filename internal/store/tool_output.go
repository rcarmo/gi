package store

import (
	"strings"
	"unicode/utf8"
)

// ToolOutputPreview follows the Piclaw 3.2.5 status window: the last
// 100 source lines, then the final 12 KiB decoded as text. The decoded UTF-8
// can be up to six bytes longer when orphaned continuation bytes become U+FFFD. This is
// a status preview, not a successful model result. Recovery may retain a
// committed SDK-reported tail only inside an explicit interrupted error result.
func ToolOutputPreview(text string) map[string]any {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	total := len(lines)
	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}
	preview := strings.Join(lines, "\n")
	truncated := total > len(lines)
	if len(preview) > 12*1024 {
		// Classic decodes the final 12 KiB as UTF-8. A cutoff inside a
		// multibyte rune becomes U+FFFD in the displayed status text.
		window := preview[len(preview)-12*1024:]
		// The input is valid UTF-8; only the cut prefix can be invalid.
		// Node Buffer.toString replaces each orphaned continuation byte.
		prefix := 0
		for prefix < len(window) && !utf8.RuneStart(window[prefix]) {
			prefix++
		}
		preview = strings.Repeat("�", prefix) + window[prefix:]
		truncated = true
	}
	return map[string]any{"output_preview": preview, "output_total_lines": total, "output_preview_lines": strings.Count(preview, "\n") + 1, "output_truncated": truncated}
}
