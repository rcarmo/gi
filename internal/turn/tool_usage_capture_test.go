package turn

import (
	"context"
	"testing"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

func TestToolUsageCaptureRevokesEscapedCallbacks(t *testing.T) {
	images := &toolExtras{}
	var inputs []int
	var returnedFlags, unknownFlags []bool
	images.usageRecorder = func(usage goai.Usage, unknown, returned bool) error {
		inputs = append(inputs, usage.Input)
		returnedFlags = append(returnedFlags, returned)
		unknownFlags = append(unknownFlags, unknown)
		return nil
	}
	add, unknown := images.addUsageFunc(), images.unknownUsageFunc()
	add(goai.Usage{Input: 5})
	unknown()
	if err := images.finishUsageCapture(); err != nil {
		t.Fatal(err)
	}
	add(goai.Usage{Input: 100})
	unknown()
	if err := images.finishUsageCapture(); err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 3 || inputs[0] != 5 || inputs[2] != 5 || returnedFlags[0] || returnedFlags[1] || !returnedFlags[2] || unknownFlags[0] || !unknownFlags[1] || !unknownFlags[2] {
		t.Fatal(inputs, returnedFlags, unknownFlags)
	}
	payload := images.withReplay(map[string]any{}, "bounded result", false)
	if payload["usage"].(goai.Usage).Input != 5 || payload["usage_unknown"] != true || images.usageRecorder != nil || images.usageCaptureCancel != nil {
		t.Fatal("escaped callback changed receipt or retained authority", payload)
	}
}

func TestToolUsageCaptureFailureCancelsCapturedOwnerAndKeepsKnownUsage(t *testing.T) {
	owner, cancel := context.WithCancel(context.Background())
	defer cancel()
	images := &toolExtras{usageCaptureCancel: cancel, usageRecorder: func(goai.Usage, bool, bool) error { return store.ErrQueueConflict }}
	add := images.addUsageFunc()
	add(goai.Usage{Input: 7})
	if owner.Err() != context.Canceled {
		t.Fatal("failed recorder did not cancel its owner")
	}
	if err := images.finishUsageCapture(); err != store.ErrQueueConflict {
		t.Fatal(err)
	}
	payload := images.withReplay(map[string]any{}, "", true)
	if payload["usage"].(goai.Usage).Input != 7 || payload["usage_unknown"] != false {
		t.Fatal("storage failure discarded received usage or fabricated billing uncertainty", payload)
	}
	add(goai.Usage{Input: 9})
	if images.usage.Input != 7 {
		t.Fatal("revoked callback changed known usage", images.usage)
	}
}
