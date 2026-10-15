package turn

import (
	"context"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// /tree's branch summary reads the branch as Pi's summarizer does: tool
// calls with their arguments, no tool results, summaries as context; the
// model gets no tools and a capped response.
func TestSummarizeBranchReadsBranchLikePi(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := NewWithRuntimeConfig(s, config.RuntimeConfig{Transport: "sse"}, "")
	defer e.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	add := func(role, content string, payload map[string]any) store.Message {
		id := store.NowID("msg")
		if err := s.AddMessage(ctx, id, "A", role, content, payload); err != nil {
			t.Fatal(err)
		}
		return store.Message{ID: id, SessionID: "A", Role: role, Content: content, Payload: payload}
	}
	branch := []store.Message{
		add("user", "Earlier branch.\n\n<read-files>\nREADME.md\n</read-files>", map[string]any{"kind": BranchSummaryKind}),
		add("user", "Fix src/parse.go", nil),
		add("assistant", "[tool_call: edit]", map[string]any{"kind": "tool_calls", "display_text": "Editing.", "tool_calls": []any{map[string]any{"id": "t1", "name": "edit", "arguments": map[string]any{"path": "src/parse.go"}}}}),
		add("tool_result", "ok", map[string]any{"kind": "tool_result", "tool_call_id": "t1"}),
		add("assistant", "Done.", map[string]any{"kind": "chat"}),
	}
	var req *goai.Context
	var hooks *inference.StreamHooks
	withStreamWithToolsHookStub(t, func(_ context.Context, _ string, conv *goai.Context, _ func(map[string]any), h *inference.StreamHooks) (*inference.StreamResult, error) {
		req, hooks = conv, h
		return &inference.StreamResult{Text: "SUMMARY", Message: &goai.Message{StopReason: goai.StopReasonStop}}, nil
	})
	summary, err := e.SummarizeBranch(ctx, "A", "p/m", branch, "the tests")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(summary, "The user explored a different conversation branch before returning here.") || !strings.Contains(summary, "SUMMARY\n\n<read-files>\nREADME.md\n</read-files>\n\n<modified-files>\nsrc/parse.go\n</modified-files>") {
		t.Fatalf("summary %q", summary)
	}
	if req == nil || len(req.Tools) != 0 || req.SystemPrompt != compaction.SummarizationSystemPrompt || hooks.Transport != goai.TransportSSE || hooks.MaxTokens != 4096 {
		t.Fatalf("request %+v hooks %+v", req, hooks)
	}
	prompt := req.Messages[0].Content[0].Text
	for _, want := range []string{"[User]: " + compaction.BranchSummaryPrefix + "Earlier branch.", "[User]: Fix src/parse.go", "[Assistant]: Editing.", `[Assistant tool calls]: edit(path="src/parse.go")`, "[Assistant]: Done.", "Additional focus: the tests"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt lacks %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "[Tool result]") {
		t.Fatalf("prompt has a tool result:\n%s", prompt)
	}
}

// A branch summary is model context as Pi's convertToLlm makes it.
func TestBranchSummaryContextLikePi(t *testing.T) {
	s := openTestStore(t)
	defer s.Close()
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "A", "A", nil); err != nil {
		t.Fatal(err)
	}
	if err := s.AddMessage(ctx, store.NowID("msg"), "A", "user", "Gone that way.", map[string]any{"kind": BranchSummaryKind}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := s.ContextSnapshot(ctx, "A")
	if err != nil {
		t.Fatal(err)
	}
	r := &sessionRunner{engine: e, store: s}
	messages := r.projectContextSnapshot(ctx, "A", snapshot)
	want := "The following is a summary of a branch that this conversation came back from:\n\n<summary>\nGone that way.</summary>"
	if len(messages) != 1 || messages[0].Role != goai.RoleUser || messages[0].Content[0].Text != want {
		t.Fatalf("context %+v", messages)
	}
}
