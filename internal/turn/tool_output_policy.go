package turn

import (
	"strings"
	"unicode/utf8"
)

const toolOutputMaxBytes = 50 * 1024
const toolOutputMaxLines = 2000

// Tool details (including spill-file references) belong to the tool. Harness
// truncation diagnostics are separate metadata, never appended to its text.
func boundToolOutput(text string) (string, map[string]any) {
	text = strings.ToValidUTF8(text, "�")
	end, lines := len(text), 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines++
			if lines == toolOutputMaxLines {
				end = i + 1
				break
			}
		}
	}
	if end > toolOutputMaxBytes {
		if newline := strings.LastIndexByte(text[:toolOutputMaxBytes], '\n'); newline >= 0 {
			end = newline + 1
		} else {
			end = toolOutputMaxBytes
			for end > 0 && !utf8.RuneStart(text[end]) {
				end--
			}
		}
	}
	bounded := text[:end]
	diagnostics := map[string]any{"total_bytes": len(text), "retained_bytes": len(bounded), "total_lines": toolTextLines(text), "retained_lines": toolTextLines(bounded), "truncated": end < len(text), "max_bytes": toolOutputMaxBytes, "max_lines": toolOutputMaxLines, "retain": "head"}
	return bounded, diagnostics
}

func toolTextLines(text string) int {
	if text == "" {
		return 0
	}
	lines := strings.Count(text, "\n")
	if text[len(text)-1] != '\n' {
		lines++
	}
	return lines
}

func withOutputDiagnostics(payload, diagnostics map[string]any) map[string]any {
	payload["output_policy"] = diagnostics
	return payload
}
