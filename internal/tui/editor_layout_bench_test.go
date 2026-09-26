package tui

import (
	"fmt"
	"strings"
	"testing"
)

var editorLayoutBenchmarkRows []renderedLine

func BenchmarkEditorLayout(b *testing.B) {
	for _, bytes := range []int{1024, 100 * 1024, 1024 * 1024} {
		draft := strings.Repeat("draft 中文🙂 e\u0301 data\n", bytes/len("draft 中文🙂 e\u0301 data\n")+1)
		for _, mode := range []string{"redraw", "cursor", "resize"} {
			b.Run(fmt.Sprintf("%d/%s", bytes, mode), func(b *testing.B) {
				m := newMultilineInput(80, "", nil, nil)
				m.maxLines = 6
				m.Focus()
				m.SetText(draft)
				m.renderLines()
				end := m.cursorPos
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					switch mode {
					case "cursor":
						m.cursorPos = end - i%2
					case "resize":
						m.width = 80 + i%2
					}
					editorLayoutBenchmarkRows = m.renderLines()
				}
			})
		}
	}
}
