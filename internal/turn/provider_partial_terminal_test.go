package turn

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// Providers may return a structured terminal error with diagnostics but without
// repeating the streamed content or usage. The committed prefix must survive
// the terminal update on the same timeline row, including after reopening.
func TestProviderPartialEmptyTerminalPreservesCommittedPrefixOnReopen(t *testing.T) {
	for _, mode := range []string{"error", "abort", "terminal-usage", "zero-terminal-usage", "invalid-terminal-usage", "different-terminal", "shorter-terminal", "longer-terminal"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "partial-terminal.db")
			s, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			e := New(s)
			defer e.Close()
			if _, err = s.CreateSession(ctx, "A", "A", nil); err != nil {
				t.Fatal(err)
			}
			withStreamWithToolsHookStub(t, func(requestCtx context.Context, _ string, _ *goai.Context, _ func(map[string]any), hooks *inference.StreamHooks) (*inference.StreamResult, error) {
				partial := testProviderPartial()
				if err := hooks.OnProgress(partial); err != nil {
					return nil, err
				}
				// UI text cannot become the authority for the retained provider prefix.
				if _, err := s.DB().Exec(`update messages set content='untrusted UI text' where role='assistant'`); err != nil {
					return nil, err
				}
				terminal := &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonError, ErrorMessage: "structured provider failure", ResponseID: "failed-response", Diagnostics: []goai.AssistantMessageDiagnostic{{Type: "provider", Error: goai.DiagnosticError{Name: "ProviderError", Message: "diagnostic retained", Code: "upstream_failure"}}}}
				switch mode {
				case "different-terminal":
					terminal.Content = []goai.ContentBlock{{Type: "text", Text: "terminal error fragment"}}
				case "shorter-terminal":
					terminal.Content = []goai.ContentBlock{{Type: "text", Text: "streamed"}}
				case "longer-terminal":
					terminal.Content = []goai.ContentBlock{{Type: "text", Text: "streamed answer completed"}}
				case "terminal-usage":
					terminal.Usage = &goai.Usage{Input: 29}
				case "zero-terminal-usage":
					terminal.Usage = &goai.Usage{}
				case "invalid-terminal-usage":
					terminal.Usage = &goai.Usage{Input: -1}
				}
				if mode == "abort" {
					if err := e.CancelActiveTurn(ctx, "A", e.LiveActiveTurn("A")); err != nil {
						return nil, err
					}
					terminal.StopReason = goai.StopReasonAborted
					return &inference.StreamResult{Message: terminal, TerminalReason: goai.StopReasonAborted, Incomplete: true}, requestCtx.Err()
				}
				return &inference.StreamResult{Message: terminal, TerminalReason: goai.StopReasonError, Usage: terminal.Usage}, errors.New("provider failed")
			})
			run, err := e.SubmitPrompt(ctx, RunInput{SessionID: "A", Prompt: "preserve streamed prefix", Model: "provider/model"})
			if err != nil {
				t.Fatal(err)
			}
			e.runs.Wait()
			if err = e.Close(); err != nil {
				t.Fatal(err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := store.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			row, err := reopened.GetTurn(ctx, run.TurnID)
			wantStatus := "failed"
			if mode == "abort" {
				wantStatus = "cancelled"
			}
			if err != nil || row.Status != wantStatus {
				t.Fatal(row, err)
			}
			messages, err := reopened.ListMessages(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			found := 0
			for _, message := range messages {
				if message.Role != "assistant" {
					continue
				}
				found++
				visible := "streamed answer"
				if mode == "different-terminal" {
					visible += "\n\nProvider response:\n\nterminal error fragment"
				} else if mode == "longer-terminal" {
					visible += " completed"
				}
				if message.Content != visible || message.Payload["adopted"] != false {
					t.Fatal("terminal erased committed prefix or used UI text", message)
				}
				terminal, ok := decodeAssistantReplay(message.Payload["provider_message"])
				contentCount := 0
				if mode == "different-terminal" || mode == "shorter-terminal" || mode == "longer-terminal" {
					contentCount = 1
				}
				if !ok || len(terminal.Content) != contentCount || terminal.ErrorMessage != "structured provider failure" || terminal.ResponseID != "failed-response" || len(terminal.Diagnostics) != 1 || terminal.Diagnostics[0].Error.Message != "diagnostic retained" {
					t.Fatal("structured terminal changed", terminal)
				}
				partial, ok := decodeAssistantReplay(message.Payload["provider_partial"])
				if !ok || len(partial.Content) != 3 || partial.Content[1].ThinkingSignature != "signature" || partial.Content[2].Arguments["prefix"] != "unfinished" || partial.Usage.Input != 17 {
					t.Fatal("committed partial structure lost", partial)
				}
				thinking, ok := message.Payload["thinking_blocks"].([]any)
				if !ok || len(thinking) != 1 {
					t.Fatal("partial thinking not visible in history", message.Payload)
				}
			}
			if found != 1 {
				t.Fatal("partial and terminal duplicated timeline row", found)
			}
			accounting, err := reopened.TurnAccounting(ctx, run.TurnID)
			input, unknown := 17, 1
			if mode == "terminal-usage" {
				input, unknown = 29, 0
			}
			if err != nil || accounting.Usage.Input != input || accounting.Unknown != unknown || accounting.Iterations != 1 {
				t.Fatal("received usage lost/double counted or falsely final", accounting, err)
			}
			snapshot, err := reopened.ContextSnapshot(ctx, "A")
			if err != nil {
				t.Fatal(err)
			}
			for _, message := range (&sessionRunner{store: reopened}).projectContextSnapshot(ctx, "A", snapshot) {
				if strings.Contains(goai.GetTextContent(&message), "streamed answer") || len(goai.GetToolCalls(&message)) != 0 {
					t.Fatal("failed prefix entered model context", message)
				}
			}
		})
	}
}
