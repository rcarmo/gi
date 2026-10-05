package web

import (
	"strings"
	"testing"
)

func BenchmarkWebTerminalReplayRing(b *testing.B) {
	ts := &terminalSession{clients: map[*terminalClient]struct{}{}}
	ts.output(strings.Repeat("x", terminalReplayLimit))
	chunk := strings.Repeat("x", 8192)
	b.ReportAllocs()
	b.SetBytes(int64(len(chunk)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.output(chunk)
	}
}
