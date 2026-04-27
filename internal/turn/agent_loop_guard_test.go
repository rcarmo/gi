package turn

import (
	"fmt"
	"testing"

	goai "github.com/rcarmo/go-ai"
)

func TestToolFailureSignatureStableForEqualArgs(t *testing.T) {
	callA := goai.ToolCall{Name: "shell", Arguments: map[string]any{"command": "echo hi", "timeout": 1}}
	callB := goai.ToolCall{Name: "shell", Arguments: map[string]any{"timeout": 1, "command": "echo hi"}}
	err := fmt.Errorf("shell: command is required")
	if toolFailureSignature(callA, err) != toolFailureSignature(callB, err) {
		t.Fatalf("expected stable signature for same args")
	}
}

func TestNextRepeatedToolFailureCount(t *testing.T) {
	call := goai.ToolCall{Name: "shell", Arguments: map[string]any{"command": ""}}
	err := fmt.Errorf("shell: command is required")
	sig, count := nextRepeatedToolFailureCount("", 0, call, err)
	if count != 1 || sig == "" {
		t.Fatalf("expected first failure count=1, got sig=%q count=%d", sig, count)
	}
	sig2, count2 := nextRepeatedToolFailureCount(sig, count, call, err)
	if sig2 != sig || count2 != 2 {
		t.Fatalf("expected repeated failure count=2, got sig=%q count=%d", sig2, count2)
	}
	other := goai.ToolCall{Name: "shell", Arguments: map[string]any{"command": "pwd"}}
	_, count3 := nextRepeatedToolFailureCount(sig2, count2, other, err)
	if count3 != 1 {
		t.Fatalf("expected different args to reset count, got %d", count3)
	}
}
