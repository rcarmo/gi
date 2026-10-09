package turn

import (
	"context"
	"strings"

	"github.com/rcarmo/gi/internal/compaction"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// BranchSummaryKind marks a branch summary message: a user-role message the
// model sees wrapped in Pi's branch summary prefix and suffix.
const BranchSummaryKind = "branch_summary"

// branchSummaryContext is a branch summary message as model context (Pi's
// convertToLlm for branchSummary).
func branchSummaryContext(summary string) goai.Message {
	return goai.UserMessage(compaction.BranchSummaryPrefix + summary + compaction.BranchSummarySuffix)
}

// branchMessages turns a branch's session messages into the messages Pi's
// branch summarizer reads: tool results are left out (the calls carry the
// context) and summaries become their context messages.
func branchMessages(messages []store.Message) []compaction.BranchMessage {
	var out []compaction.BranchMessage
	for _, m := range messages {
		kind, _ := m.Payload["kind"].(string)
		switch {
		case m.Role == "user" && kind == BranchSummaryKind:
			out = append(out, compaction.BranchMessage{Message: branchSummaryContext(m.Content), Summary: m.Content, FromBranch: true})
		case m.Role == "user":
			out = append(out, compaction.BranchMessage{Message: goai.UserMessage(m.Content)})
		case m.Role == "assistant" && kind == "compaction":
			out = append(out, compaction.BranchMessage{Message: goai.UserMessage(compaction.SummaryPrefix + m.Content + compaction.SummarySuffix), Summary: m.Content})
		case m.Role == "assistant" && kind == "tool_calls":
			msg := goai.Message{Role: goai.RoleAssistant}
			if text, _ := m.Payload["display_text"].(string); strings.TrimSpace(text) != "" {
				msg.Content = append(msg.Content, goai.ContentBlock{Type: "text", Text: text})
			}
			calls, _ := m.Payload["tool_calls"].([]any)
			for _, raw := range calls {
				call, _ := raw.(map[string]any)
				id, _ := call["id"].(string)
				name, _ := call["name"].(string)
				args, _ := call["arguments"].(map[string]any)
				msg.Content = append(msg.Content, goai.ContentBlock{Type: "toolCall", ID: id, Name: name, Arguments: args})
			}
			out = append(out, compaction.BranchMessage{Message: msg})
		case m.Role == "assistant":
			out = append(out, compaction.BranchMessage{Message: goai.Message{Role: goai.RoleAssistant, Content: []goai.ContentBlock{{Type: "text", Text: m.Content}}}})
		}
	}
	return out
}

// SummarizeBranch is Pi's generateBranchSummary for /tree: the session's
// model, without tools, summarizes the messages of the branch being left.
// instructions add a focus to Pi's prompt.
func (e *Engine) SummarizeBranch(ctx context.Context, sessionID, model string, messages []store.Message, instructions string) (string, error) {
	thinking := ""
	if session, err := e.store.GetSession(ctx, sessionID); err == nil {
		thinking = e.admissionThinking(session, model)
	}
	summarize := func(ctx context.Context, req compaction.SummaryRequest) (compaction.SummaryResponse, error) {
		return summarizeWith(ctx, model, thinking, req, goai.Transport(e.runtimeCfg.Transport))
	}
	reserve := e.runtimeCfg.BranchSummaryReserveTokens
	if reserve <= 0 {
		reserve = compaction.DefaultBranchSummaryReserveTokens
	}
	summary, _, err := compaction.GenerateBranchSummary(ctx, summarize, branchMessages(messages), inference.ResolveModelContextWindow(e.runtimeCfg.DefaultProvider, model), reserve, inference.ModelMaxTokens(model), strings.TrimSpace(instructions), false)
	return summary, err
}
