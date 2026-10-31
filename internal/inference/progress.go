package inference

import (
	"strings"

	goai "github.com/rcarmo/go-ai"
)

// Prefer provider-owned structured partials. The fallback is reconstructed from
// provider events, never UI draft text, and is always non-authoritative history.
func streamEventPartial(event goai.Event) *goai.Message {
	switch e := event.(type) {
	case *goai.StartEvent:
		return e.Partial
	case *goai.TextStartEvent:
		return e.Partial
	case *goai.TextDeltaEvent:
		return e.Partial
	case *goai.TextEndEvent:
		return e.Partial
	case *goai.ThinkingStartEvent:
		return e.Partial
	case *goai.ThinkingDeltaEvent:
		return e.Partial
	case *goai.ThinkingEndEvent:
		return e.Partial
	case *goai.ToolCallStartEvent:
		return e.Partial
	case *goai.ToolCallDeltaEvent:
		return e.Partial
	case *goai.ToolCallEndEvent:
		return e.Partial
	}
	return nil
}

type progressSlot struct {
	index int
	kind  string
}
type fallbackProgress struct {
	message  goai.Message
	bytes    int
	slots    map[progressSlot]int
	builders map[progressSlot]*strings.Builder
}

func (f *fallbackProgress) truncated() {
	f.message.Details = map[string]any{"partial_fallback_truncated": true}
}
func (f *fallbackProgress) append(index int, kind string, dst *string, s string) {
	n := min(len(s), max(0, (64<<10)-f.bytes))
	if n < len(s) {
		f.truncated()
	}
	if n == 0 {
		return
	}
	if f.builders == nil {
		f.builders = map[progressSlot]*strings.Builder{}
	}
	key := progressSlot{index, kind}
	builder := f.builders[key]
	if builder == nil {
		builder = &strings.Builder{}
		f.builders[key] = builder
	}
	builder.WriteString(s[:n])
	*dst = builder.String()
	f.bytes += n
}
func (f *fallbackProgress) block(index int, kind string) *goai.ContentBlock {
	if index < 0 || index >= 64 {
		f.truncated()
		return nil
	}
	key := progressSlot{index, kind}
	if at, ok := f.slots[key]; ok {
		return &f.message.Content[at]
	}
	if len(f.message.Content) >= 64 {
		f.truncated()
		return nil
	}
	if f.slots == nil {
		f.slots = map[progressSlot]int{}
	}
	f.slots[key] = len(f.message.Content)
	f.message.Content = append(f.message.Content, goai.ContentBlock{Type: kind})
	return &f.message.Content[len(f.message.Content)-1]
}
func (f *fallbackProgress) observe(event goai.Event) *goai.Message {
	f.message.Role = goai.RoleAssistant
	f.message.StopReason = goai.StopReasonPending
	switch e := event.(type) {
	case *goai.TextDeltaEvent:
		if b := f.block(e.ContentIndex, "text"); b != nil {
			f.append(e.ContentIndex, "text", &b.Text, e.Delta)
		}
	case *goai.ThinkingDeltaEvent:
		if b := f.block(e.ContentIndex, "thinking"); b != nil {
			f.append(e.ContentIndex, "thinking", &b.Thinking, e.Delta)
		}
	case *goai.ToolCallDeltaEvent:
		if b := f.block(e.ContentIndex, "toolCall"); b != nil {
			previous, _ := b.Arguments["_partial_json"].(string)
			f.append(e.ContentIndex, "toolCall", &previous, e.Delta)
			if b.Arguments == nil {
				b.Arguments = map[string]any{}
			}
			b.Arguments["_partial_json"] = previous
			b.ArgumentsError = "incomplete streamed arguments"
		}
	case *goai.ToolCallEndEvent:
		if b := f.block(e.ContentIndex, "toolCall"); b != nil {
			*b = goai.ContentBlock{Type: "toolCall", ID: e.ToolCall.ID, Name: e.ToolCall.Name, Arguments: e.ToolCall.Arguments, ArgumentsError: e.ToolCall.ArgumentsError, ThoughtSignature: e.ToolCall.ThoughtSignature, ThoughtSignaturePresent: e.ToolCall.ThoughtSignaturePresent, Namespace: e.ToolCall.Namespace, NamespacePresent: e.ToolCall.NamespacePresent}
		}
	}
	if len(f.message.Content) == 0 {
		return nil
	}
	return &f.message
}

func compatProviderPartial(model *goai.Model, text, thinking string, usage *goai.Usage, calls map[int]*goai.ToolCall, fragments map[int]*strings.Builder) *goai.Message {
	message := &goai.Message{Role: goai.RoleAssistant, Model: model.ID, Provider: model.Provider, Api: model.Api, StopReason: goai.StopReasonPending}
	if usage != nil {
		u := *usage
		message.Usage = &u
	}
	if thinking != "" {
		message.Content = append(message.Content, goai.ContentBlock{Type: "thinking", Thinking: thinking})
	}
	if text != "" {
		message.Content = append(message.Content, goai.ContentBlock{Type: "text", Text: text})
	}
	limited := make(map[int]*goai.ToolCall, min(len(calls), 64))
	for idx, call := range calls {
		if len(limited) == 64 {
			message.Details = map[string]any{"partial_calls_truncated": true}
			break
		}
		limited[idx] = call
	}
	for _, idx := range sortedToolCallIndexes(limited) {
		call := limited[idx]
		block := goai.ContentBlock{Type: "toolCall", ID: call.ID, Name: call.Name, ArgumentsError: "incomplete streamed arguments"}
		if value := fragments[idx]; value != nil {
			block.Arguments = map[string]any{"_partial_json": value.String()}
		}
		message.Content = append(message.Content, block)
	}
	return message
}
