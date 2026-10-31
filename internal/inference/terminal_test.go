package inference

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	goai "github.com/rcarmo/go-ai"
)

func terminalFixture(stop goai.StopReason, calls bool) *StreamResult {
	message := &goai.Message{Role: goai.RoleAssistant, StopReason: stop, Content: []goai.ContentBlock{{Type: "text", Text: "answer"}}}
	if calls {
		message.Content = append(message.Content, goai.ContentBlock{Type: "toolCall", ID: "c", Name: "fixture", Arguments: map[string]any{}})
	}
	return &StreamResult{Message: message}
}

func TestTerminalValidationStopMatrix(t *testing.T) {
	for _, stop := range []goai.StopReason{goai.StopReasonStop, goai.StopReasonLength, goai.StopReasonToolUse, goai.StopReasonPending, goai.StopReasonDeferred, goai.StopReasonError, goai.StopReasonAborted, "", "unknown"} {
		for _, calls := range []bool{false, true} {
			t.Run(string(stop)+map[bool]string{false: "/text", true: "/calls"}[calls], func(t *testing.T) {
				result := terminalFixture(stop, calls)
				valid := stop == goai.StopReasonToolUse || (!calls && (stop == goai.StopReasonStop || stop == goai.StopReasonLength))
				if err := ValidateAssistantResult(result); (err == nil) != valid {
					t.Fatalf("valid=%v error=%v", valid, err)
				}
			})
		}
	}
}

func TestTerminalValidationRejectsInvalidRoleContentAndUsage(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*StreamResult)
	}{
		{"role", func(r *StreamResult) { r.Message.Role = goai.RoleUser }},
		{"content", func(r *StreamResult) { r.Message.Content[0].Type = "image" }},
		{"missing-id", func(r *StreamResult) { r.Message.Content[1].ID = "" }},
		{"duplicate-id", func(r *StreamResult) { r.Message.Content = append(r.Message.Content, r.Message.Content[1]) }},
		{"arguments", func(r *StreamResult) { r.Message.Content[1].Arguments["bad"] = make(chan int) }},
		{"negative-usage", func(r *StreamResult) { r.Usage = &goai.Usage{Input: -1} }},
		{"negative-reasoning", func(r *StreamResult) { r.Message.Usage = &goai.Usage{Reasoning: -1} }},
		{"nan-cost", func(r *StreamResult) { r.Message.Usage = &goai.Usage{Cost: goai.CostBreakdown{Total: math.NaN()}} }},
		{"infinite-cost", func(r *StreamResult) { r.Usage = &goai.Usage{Cost: goai.CostBreakdown{Input: math.Inf(1)}} }},
		{"deferred", func(r *StreamResult) {
			r.Message.StopReason = goai.StopReasonDeferred
			r.Message.Deferred = &goai.DeferredHandle{ID: "valid-but-unsupported"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := terminalFixture(goai.StopReasonToolUse, true)
			tc.edit(r)
			if ValidateAssistantResult(r) == nil {
				t.Fatal("invalid terminal accepted")
			}
		})
	}
}

func TestProviderStreamRejectsMissingDuplicateAndContradictoryTerminals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		events []goai.Event
	}{
		{"missing", []goai.Event{&goai.TextDeltaEvent{Delta: "partial"}}},
		{"duplicate", []goai.Event{&goai.DoneEvent{Message: terminalFixture(goai.StopReasonStop, false).Message}, &goai.DoneEvent{Message: terminalFixture(goai.StopReasonStop, false).Message}}},
		{"nil", []goai.Event{&goai.DoneEvent{}}},
		{"mismatch", []goai.Event{&goai.DoneEvent{Reason: goai.StopReasonLength, Message: terminalFixture(goai.StopReasonStop, false).Message}}},
		{"invalid-usage", []goai.Event{&goai.DoneEvent{Message: &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: -1}}}}},
		{"nil-error", []goai.Event{&goai.ErrorEvent{Reason: goai.StopReasonError}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := make(chan goai.Event, len(tc.events))
			for _, ev := range tc.events {
				events <- ev
			}
			close(events)
			done := false
			if _, err := collectProviderStream(context.Background(), "fixture", events, func(ev map[string]any) {
				if ev["type"] == "done" {
					done = true
				}
			}); err == nil || done {
				t.Fatal("invalid terminal adopted", err, done)
			}
		})
	}
}

func TestProviderStreamJoinsProducerBeforeReturningTerminal(t *testing.T) {
	events := make(chan goai.Event)
	sent, release := make(chan struct{}), make(chan struct{})
	returned := make(chan error, 1)
	go func() {
		events <- &goai.DoneEvent{Message: terminalFixture(goai.StopReasonStop, false).Message}
		close(sent)
		<-release
		close(events)
	}()
	go func() { _, err := collectProviderStream(context.Background(), "fixture", events, nil); returned <- err }()
	<-sent
	select {
	case err := <-returned:
		t.Fatal("returned before producer closure", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-returned:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("did not finish after producer joined")
	}
}

func TestDeferredCapabilityRejectedExplicitly(t *testing.T) {
	result := terminalFixture(goai.StopReasonDeferred, false)
	if err := ValidateAssistantResult(result); err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatal(err)
	}
}
