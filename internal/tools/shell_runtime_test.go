package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunShellPromptDrainsOutputBeforeWait(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	prompt := strings.Repeat("drain-output-", 2000)
	var streamed strings.Builder
	out, err, cancelled := RunShellPrompt(ctx, prompt, nil, func(chunk string) {
		streamed.WriteString(chunk)
		// Let the child exit while the reader still has buffered output to drain.
		time.Sleep(time.Millisecond)
	})
	want := "Gi received: " + prompt
	if err != nil || cancelled {
		t.Fatalf("shell result: err=%v cancelled=%v", err, cancelled)
	}
	if out != want || streamed.String() != want {
		t.Fatalf("output truncated: returned=%d streamed=%d expected=%d", len(out), streamed.Len(), len(want))
	}
}
