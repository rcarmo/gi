package scripting

import (
	"context"
	"fmt"
	"strings"
	"testing"

	core "github.com/rcarmo/go-joker/v42/core"
)

const jokerBenchmarkScript = `(+ (:counter (gi.state/session-state)) (count (gi.state/messages)))`

func jokerBenchmarkBridge() *Bridge {
	messages := make([]map[string]any, 32)
	for i := range messages {
		messages[i] = map[string]any{"id": fmt.Sprint(i), "role": "user", "content": strings.Repeat("x", 512)}
	}
	return NewBridge("fixture-session", BridgeFuncs{
		GetSessionState: func(context.Context) (map[string]any, error) {
			return map[string]any{"counter": 7}, nil
		},
		GetConfig: func(context.Context) (map[string]any, error) {
			return map[string]any{"model": "fixture-model"}, nil
		},
		ListMessages: func(context.Context, int) ([]map[string]any, error) { return messages, nil },
	})
}

func benchmarkJokerCall(b *testing.B, bridge *Bridge) {
	b.Helper()
	result, err := ExecuteEmbeddedJoker(context.Background(), jokerBenchmarkScript, bridge)
	if err != nil || result != "39" {
		b.Fatalf("bridge/script contract: result=%q err=%v", result, err)
	}
}

// bench-joker-first-call runs this with -benchtime=1x in three fresh test
// processes, not -count=3 in one process (which would share loaded libraries).
func BenchmarkEmbeddedJokerFirstCall(b *testing.B) {
	bridge := jokerBenchmarkBridge()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		benchmarkJokerCall(b, bridge)
	}
}

func BenchmarkEmbeddedJokerSteady(b *testing.B) {
	for _, miss := range []bool{false, true} {
		name := "parsed-hit"
		if miss {
			name = "parsed-miss"
		}
		b.Run(name, func(b *testing.B) {
			embeddedJokerProgramMu.Lock()
			embeddedJokerPrograms = make(map[string][]core.Expr)
			embeddedJokerProgramMu.Unlock()
			bridge := jokerBenchmarkBridge()
			ids := make([]string, b.N)
			for i := range ids {
				ids[i] = fmt.Sprintf("fixture-session-%d", i)
			}
			// Load libraries and prime the script before timing. Profiling totals
			// still include this one-time priming; per-operation counters do not.
			benchmarkJokerCall(b, bridge)
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				if miss {
					// Dynamic bridge metadata changes the existing full-script key.
					bridge.SessionID = ids[i]
				}
				benchmarkJokerCall(b, bridge)
			}
		})
	}
}
