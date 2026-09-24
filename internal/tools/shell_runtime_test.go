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

func TestRunShellPromptCancellationReturnsAfterReadersClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	var output string
	var runErr error
	var cancelled bool
	go func() {
		defer close(done)
		output, runErr, cancelled = RunShellPrompt(ctx, strings.Repeat("cancel-output-", 2000), nil, func(string) { cancel() })
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation blocked on output reader or child wait")
	}
	if runErr != nil || !cancelled {
		t.Fatalf("cancelled=%v err=%v", cancelled, runErr)
	}
	if !strings.HasPrefix(output, "Gi received: ") {
		t.Fatalf("lost observed output: %q", output)
	}
}
