package compaction

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/config"
	goai "github.com/rcarmo/go-ai"
)

func TestPrepareCompactionKeepsRecentMessages(t *testing.T) {
	messages := []goai.Message{
		goai.UserMessage(strings.Repeat("a", 400)),
		goai.UserMessage(strings.Repeat("b", 400)),
		goai.UserMessage(strings.Repeat("c", 400)),
		goai.UserMessage(strings.Repeat("d", 400)),
	}
	prep := Prepare(messages, EstimateMessagesTokens(messages), 120, 20, 100, "default")
	if prep.MessagesToSummarize == 0 || prep.RecentMessages == 0 {
		t.Fatalf("bad preparation: %#v", prep)
	}
	if !strings.Contains(prep.Transcript, "user") {
		t.Fatalf("expected transcript: %#v", prep)
	}
}

func TestMaybeCompactContextUsesHookSummary(t *testing.T) {
	conv := &goai.Context{Messages: []goai.Message{
		goai.UserMessage(strings.Repeat("older1 ", 80)),
		goai.UserMessage(strings.Repeat("older2 ", 80)),
		goai.UserMessage(strings.Repeat("older3 ", 80)),
		goai.UserMessage(strings.Repeat("older4 ", 80)),
		goai.UserMessage("recent question"),
		goai.UserMessage("recent answer"),
	}}
	err := MaybeCompactContext(context.Background(), RuntimeRequest{SessionID: "s", TurnID: "t", AgentID: "agent", Model: "bootstrap", Settings: config.CompactionSettings{Enabled: true, ContextWindow: 1000, ThresholdTokens: 50, KeepRecentTokens: 20, ReserveTokens: 10}}, conv, RuntimeOps{BeforeCompact: func(context.Context, map[string]any, []goai.Message) (HookDecision, error) {
		return HookDecision{Payload: map[string]any{"summary": "smart joker summary"}}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(conv.Messages) >= 6 {
		t.Fatalf("expected compacted context, got %d", len(conv.Messages))
	}
	if !strings.Contains(goai.GetTextContent(&conv.Messages[0]), "smart joker summary") {
		t.Fatalf("missing hook summary: %q", goai.GetTextContent(&conv.Messages[0]))
	}
}

func runtimeFixture() (*goai.Context, RuntimeRequest) {
	conv := &goai.Context{}
	for i := 0; i < 6; i++ {
		conv.Messages = append(conv.Messages, goai.UserMessage(strings.Repeat("history ", 60)))
	}
	return conv, RuntimeRequest{SessionID: "s", TurnID: "t", Settings: config.CompactionSettings{Enabled: true, ThresholdTokens: 50, KeepRecentTokens: 20}}
}
func TestCompactionLifecycleOutcomesKeepOriginalContext(t *testing.T) {
	for _, kind := range []string{"cancel", "suppress", "hook-error", "persistence", "begin"} {
		t.Run(kind, func(t *testing.T) {
			conv, req := runtimeFixture()
			before := append([]goai.Message(nil), conv.Messages...)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var events []string
			var outcomes []string
			ops := RuntimeOps{BackgroundContext: context.Background, Begin: func(context.Context, string, string, map[string]any) (int, error) {
				if kind == "begin" {
					return 0, errors.New("begin failed")
				}
				return 4, nil
			},
				BeforeCompact: func(context.Context, map[string]any, []goai.Message) (HookDecision, error) {
					switch kind {
					case "cancel":
						cancel()
					case "suppress":
						return HookDecision{Block: true}, nil
					case "hook-error":
						return HookDecision{}, errors.New("hook failed")
					}
					return HookDecision{Payload: map[string]any{"summary": "summary"}}, nil
				},
				Finish: func(ctx context.Context, _, _ string, seq int, outcome, summary string, _ map[string]any) (string, error) {
					if ctx.Err() != nil || seq != 4 {
						t.Fatal("terminal persistence must have live durable context", ctx.Err(), seq)
					}
					outcomes = append(outcomes, outcome)
					if kind == "persistence" && outcome == "completed" {
						return "", errors.New("write failed")
					}
					return outcome, nil
				},
				Broadcast: func(_ string, ev map[string]any) { events = append(events, ev["type"].(string)) },
			}
			err := MaybeCompactContext(ctx, req, conv, ops)
			if (kind == "cancel" || kind == "persistence" || kind == "begin") && err == nil {
				t.Fatal("missing failure")
			}
			if !reflect.DeepEqual(conv.Messages, before) {
				t.Fatal("failed compaction changed context")
			}
			for _, event := range events {
				if event == "compaction_completed" || event == "compaction" {
					t.Fatal("false success", events)
				}
			}
			switch kind {
			case "cancel":
				if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(outcomes, []string{"cancelled"}) {
					t.Fatal(err, outcomes)
				}
			case "suppress":
				if !reflect.DeepEqual(outcomes, []string{"suppressed"}) {
					t.Fatal(outcomes)
				}
			case "hook-error":
				if !reflect.DeepEqual(outcomes, []string{"failed"}) {
					t.Fatal(outcomes)
				}
			case "persistence":
				if !reflect.DeepEqual(outcomes, []string{"completed", "failed"}) {
					t.Fatal(outcomes)
				}
			case "begin":
				if len(events) > 0 || len(outcomes) > 0 {
					t.Fatal(events, outcomes)
				}
			}
		})
	}
}
func TestCompactionPersistsBeforeContextAndSuccessBroadcast(t *testing.T) {
	conv, req := runtimeFixture()
	persisted := false
	var events []string
	err := MaybeCompactContext(context.Background(), req, conv, RuntimeOps{
		Begin: func(context.Context, string, string, map[string]any) (int, error) { return 1, nil },
		BeforeCompact: func(context.Context, map[string]any, []goai.Message) (HookDecision, error) {
			return HookDecision{Payload: map[string]any{"summary": "small"}}, nil
		},
		Finish: func(_ context.Context, _, _ string, _ int, outcome, summary string, _ map[string]any) (string, error) {
			if len(conv.Messages) != 6 || summary != "small" {
				t.Fatal("mutated before durable commit")
			}
			persisted = true
			return outcome, nil
		},
		Broadcast: func(_ string, ev map[string]any) {
			kind := ev["type"].(string)
			for _, key := range []string{"summary", "content", "text", "message"} {
				if _, ok := ev[key]; ok {
					t.Fatalf("internal summary exposed to live SSE: %+v", ev)
				}
			}
			if kind != "compaction_started" && !persisted {
				t.Fatal("broadcast before persistence")
			}
			events = append(events, kind)
		},
	})
	if err != nil || len(conv.Messages) != 2 || !reflect.DeepEqual(events, []string{"compaction_started", "compaction_completed", "compaction"}) {
		t.Fatal(err, len(conv.Messages), events)
	}
}

// Pi's /compact [instructions]: the hook receives custom_instructions and the
// default summary carries "Additional focus" into the compacted context.
func TestMaybeCompactContextCarriesCustomInstructions(t *testing.T) {
	conv, req := runtimeFixture()
	req.Force = true
	req.Instructions = "keep the database migration plan"
	var hookSaw any
	err := MaybeCompactContext(context.Background(), req, conv, RuntimeOps{BeforeCompact: func(_ context.Context, payload map[string]any, _ []goai.Message) (HookDecision, error) {
		hookSaw = payload["custom_instructions"]
		return HookDecision{}, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if hookSaw != req.Instructions {
		t.Fatalf("hook custom_instructions = %v", hookSaw)
	}
	if got := goai.GetTextContent(&conv.Messages[0]); !strings.Contains(got, "Additional focus: keep the database migration plan") {
		t.Fatalf("summary lacks the focus: %q", got)
	}
}

// The session model writes the summary (Pi); a failed summary falls back to
// gi's heuristic and says why; a hook summary still wins.
func TestCompactionSummarisesWithTheModel(t *testing.T) {
	conv, req := runtimeFixture()
	req.Settings.ReserveTokens = 1000
	var prompts []SummaryRequest
	ops := RuntimeOps{Summarize: func(_ context.Context, r SummaryRequest) (SummaryResponse, error) {
		prompts = append(prompts, r)
		return SummaryResponse{Text: "## Goal\nmodel summary", StopReason: goai.StopReasonStop, Usage: &goai.Usage{Input: 7, Output: 3}}, nil
	}}
	var finished map[string]any
	ops.Finish = func(_ context.Context, _, _ string, _ int, outcome, summary string, payload map[string]any) (string, error) {
		finished = payload
		return outcome, nil
	}
	if err := MaybeCompactContext(context.Background(), req, conv, ops); err != nil {
		t.Fatal(err)
	}
	if len(prompts) != 1 || prompts[0].SystemPrompt != SummarizationSystemPrompt || prompts[0].MaxTokens != 800 || !strings.Contains(prompts[0].Prompt, "<conversation>\n[User]: history") {
		t.Fatalf("prompts %+v", prompts)
	}
	if !strings.Contains(goai.GetTextContent(&conv.Messages[0]), "model summary") || finished["summary_source"] != "model" || finished["usage"] == nil {
		t.Fatalf("summary %q payload %v", goai.GetTextContent(&conv.Messages[0]), finished)
	}
	// The next compaction updates that summary.
	for i := 0; i < 6; i++ {
		conv.Messages = append(conv.Messages, goai.UserMessage(strings.Repeat("more ", 60)))
	}
	prompts = nil
	if err := MaybeCompactContext(context.Background(), req, conv, ops); err != nil {
		t.Fatal(err)
	}
	if len(prompts) != 1 || !strings.Contains(prompts[0].Prompt, "<previous-summary>\n## Goal\nmodel summary\n</previous-summary>") {
		t.Fatalf("update prompt %q", prompts[0].Prompt)
	}

	// A truncated summary is not used.
	conv, _ = runtimeFixture()
	ops.Summarize = func(context.Context, SummaryRequest) (SummaryResponse, error) {
		return SummaryResponse{Text: "partial", StopReason: goai.StopReasonLength}, nil
	}
	if err := MaybeCompactContext(context.Background(), req, conv, ops); err != nil {
		t.Fatal(err)
	}
	if finished["summary_source"] != "heuristic" || !strings.Contains(finished["summary_error"].(string), "token cap") {
		t.Fatalf("fallback payload %v", finished)
	}
	// A hook summary wins without calling the model.
	conv, _ = runtimeFixture()
	called := false
	ops.Summarize = func(context.Context, SummaryRequest) (SummaryResponse, error) { called = true; return SummaryResponse{}, nil }
	ops.BeforeCompact = func(context.Context, map[string]any, []goai.Message) (HookDecision, error) {
		return HookDecision{Payload: map[string]any{"summary": "hook summary"}}, nil
	}
	if err := MaybeCompactContext(context.Background(), req, conv, ops); err != nil || called || finished["summary_source"] != "hook" {
		t.Fatalf("hook: %v called=%v %v", err, called, finished)
	}
}
