package store

import (
	"context"
	goai "github.com/rcarmo/go-ai"
	"strings"
	"testing"
)

func partialPayload(text string) map[string]any {
	return map[string]any{"provider_message": &goai.Message{Role: goai.RoleAssistant, StopReason: goai.StopReasonPending, Content: []goai.ContentBlock{{Type: "text", Text: text}, {Type: "thinking", Thinking: "reason", ThinkingSignature: "signature"}, {Type: "toolCall", ID: "call", Name: "test", Arguments: map[string]any{"prefix": "unfinished"}, ArgumentsError: "incomplete"}}, Usage: &goai.Usage{Input: 17}}, "partial_checkpoint": map[string]any{"truncated": true}}
}
func TestProviderPartialReopenAndSeal(t *testing.T) {
	s, path := settlementFixture(t)
	ctx := context.Background()
	if err := s.SaveProviderPartial(ctx, "t", "s", "request", "draft", partialPayload("draft")); err == nil {
		t.Fatal("unadmitted checkpoint accepted")
	}
	if err := s.AppendTurnEvent(ctx, "t", "s", "inference.started", map[string]any{"request_id": "request"}); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"first", "latest"} {
		if err := s.SaveProviderPartial(ctx, "t", "s", "request", text, partialPayload(text)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SaveProviderPartial(ctx, "t", "foreign", "request", "bad", partialPayload("bad")); err == nil {
		t.Fatal("foreign checkpoint accepted")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	record, err := s.ProviderPartial(ctx, "t", "s", "request")
	if err != nil || record == nil || record.Content != "latest" || record.Message.Usage.Input != 17 || len(record.Message.Content) != 3 {
		t.Fatal(record, err)
	}
	a, err := s.TurnAccounting(ctx, "t")
	if err != nil || a.Unknown != 1 || a.Usage.Input != 0 {
		t.Fatal("unsettled usage claimed as final", a, err)
	}
	sealed, err := s.SealProviderPartial(ctx, "t", "s", "request", "owner terminated before terminal response")
	if err != nil || !sealed.Applied || sealed.Accounting.Unknown != 1 || sealed.Accounting.Usage.Input != 17 {
		t.Fatal(sealed, err)
	}
	again, err := s.SealProviderPartial(ctx, "t", "s", "request", "repeat")
	if err != nil || again.Applied || again.Accounting.Usage.Input != 17 {
		t.Fatal("duplicate interrupted receipt", again, err)
	}
	if err = s.SaveProviderPartial(ctx, "t", "s", "request", "late", partialPayload("late")); err != nil {
		t.Fatal(err)
	}
	messages, err := s.ListMessages(ctx, "s")
	if err != nil || len(messages) != 1 || messages[0].Content != "latest" || messages[0].Payload["adopted"] != false {
		t.Fatal(messages, err)
	}
}
func TestProviderPartialTerminalUpdateRollbackAndBounds(t *testing.T) {
	s, _ := settlementFixture(t)
	ctx := context.Background()
	if err := s.AppendTurnEvent(ctx, "t", "s", "inference.started", map[string]any{"request_id": "request"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveProviderPartial(ctx, "t", "s", "request", "draft", partialPayload("draft")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`create trigger reject_final before insert on turn_receipts begin select raise(abort,'receipt fault'); end`); err != nil {
		t.Fatal(err)
	}
	receipt := sampleProviderReceipt()
	receipt.MessageID = ProviderMessageID("request")
	if _, err := s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err == nil {
		t.Fatal("fault hidden")
	}
	record, err := s.ProviderPartial(ctx, "t", "s", "request")
	if err != nil || record == nil || record.Content != "draft" {
		t.Fatal("checkpoint lost on failed terminal update", record, err)
	}
	if _, err = s.db.Exec(`drop trigger reject_final`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SettleProviderAttempt(ctx, "t", "s", "request", receipt); err != nil {
		t.Fatal(err)
	}
	if record, err = s.ProviderPartial(ctx, "t", "s", "request"); err != nil || record != nil {
		t.Fatal("successful terminal left partial", record, err)
	}
	messages, err := s.ListMessages(ctx, "s")
	if err != nil || len(messages) != 1 || messages[0].Content != "answer" {
		t.Fatal(messages, err)
	}
	if err = s.AppendTurnEvent(ctx, "t", "s", "inference.started", map[string]any{"request_id": "huge"}); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveProviderPartial(ctx, "t", "s", "huge", "huge", partialPayload(strings.Repeat("x", 513<<10))); err == nil {
		t.Fatal("unbounded partial accepted")
	}
}
