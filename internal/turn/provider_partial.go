package turn

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

const providerPartialInterval = 75 * time.Millisecond
const providerPartialStringBytes = 32 << 10

// Partial checkpoints deliberately omit oversized/deep values. They are history
// receipts with adopted=false and can never authorise tool execution or replay.
type partialBudget struct {
	bytes, nodes int
	truncated    bool
}

func (b *partialBudget) text(s string) string {
	n := min(len(s), b.bytes)
	for n > 0 && n < len(s) && !utf8.RuneStart(s[n]) {
		n--
	}
	if n < len(s) || !utf8.ValidString(s[:n]) {
		b.truncated = true
	}
	b.bytes -= n
	return strings.Clone(strings.ToValidUTF8(s[:n], "?"))
}
func (b *partialBudget) value(v any, depth int) any {
	if b.nodes <= 0 {
		b.truncated = true
		return nil
	}
	b.nodes--
	if depth > 8 {
		b.truncated = true
		return nil
	}
	switch x := v.(type) {
	case nil, bool, int, int64, int32, uint, uint64, uint32:
		return x
	case string:
		return b.text(x)
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			b.truncated = true
			return nil
		}
		return x
	case float32:
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
			b.truncated = true
			return nil
		}
		return x
	case json.Number:
		if len(x) > b.bytes {
			b.truncated = true
			return nil
		}
		raw, err := json.Marshal(x)
		if err != nil {
			b.truncated = true
			return nil
		}
		b.bytes -= len(raw)
		return json.Number(strings.Clone(string(raw)))
	case json.RawMessage:
		if len(x) > b.bytes {
			b.truncated = true
			return nil
		}
		var parsed any
		if json.Unmarshal(x, &parsed) != nil {
			b.truncated = true
			return nil
		}
		return b.value(parsed, depth+1)
	case map[string]any:
		out := map[string]any{}
		for k, v := range x {
			if b.nodes <= 0 || b.bytes <= 0 {
				b.truncated = true
				break
			}
			if len(k) > b.bytes {
				b.truncated = true
				break
			}
			key := b.text(k)
			out[key] = b.value(v, depth+1)
		}
		return out
	case []any:
		out := make([]any, 0, min(len(x), b.nodes))
		for _, v := range x {
			if b.nodes <= 0 || b.bytes <= 0 {
				b.truncated = true
				break
			}
			out = append(out, b.value(v, depth+1))
		}
		return out
	default:
		b.truncated = true
		return nil
	}
}

func partialDetailMap(v any) map[string]any { m, _ := v.(map[string]any); return m }

func boundedProviderPartial(message *goai.Message) (*goai.Message, bool) {
	if message == nil {
		return nil, false
	}
	b := partialBudget{bytes: providerPartialStringBytes, nodes: 256}
	out := &goai.Message{Role: goai.Role(b.text(string(message.Role))), Timestamp: message.Timestamp, Api: goai.Api(b.text(string(message.Api))), Provider: goai.Provider(b.text(string(message.Provider))), Model: b.text(message.Model), ResponseID: b.text(message.ResponseID), ResponseModel: b.text(message.ResponseModel), ProviderThinkingLevel: b.text(message.ProviderThinkingLevel), ThinkingLevel: goai.ModelThinkingLevel(b.text(string(message.ThinkingLevel))), StopReason: goai.StopReason(b.text(string(message.StopReason))), RawStopReason: b.text(message.RawStopReason), ErrorMessage: b.text(message.ErrorMessage)}
	if message.DurationMs != nil {
		n := *message.DurationMs
		out.DurationMs = &n
	}
	if message.EndTurn != nil {
		n := *message.EndTurn
		out.EndTurn = &n
	}
	if message.Usage != nil {
		if store.ValidateReceiptUsage(message.Usage) == nil {
			u := *message.Usage
			out.Usage = &u
		} else {
			b.truncated = true
		}
	}
	for _, block := range message.Content {
		if len(out.Content) >= 64 || b.bytes <= 0 {
			b.truncated = true
			break
		}
		c := goai.ContentBlock{Type: b.text(block.Type), Redacted: block.Redacted, TextSignaturePresent: block.TextSignaturePresent, ThinkingSignaturePresent: block.ThinkingSignaturePresent, RedactedPresent: block.RedactedPresent, ThoughtSignaturePresent: block.ThoughtSignaturePresent, NamespacePresent: block.NamespacePresent}
		c.Text = b.text(block.Text)
		c.Thinking = b.text(block.Thinking)
		c.TextSignature = b.text(block.TextSignature)
		c.ThinkingSignature = b.text(block.ThinkingSignature)
		c.ID = b.text(block.ID)
		c.Name = b.text(block.Name)
		c.ArgumentsError = b.text(block.ArgumentsError)
		c.ThoughtSignature = b.text(block.ThoughtSignature)
		c.Namespace = b.text(block.Namespace)
		if block.Arguments != nil {
			c.Arguments, _ = b.value(block.Arguments, 0).(map[string]any)
		}
		if block.Type == "image" {
			if len(block.Data) > b.bytes {
				b.truncated = true
				c = goai.ContentBlock{Type: "text", Text: "[image omitted from partial checkpoint]"}
			} else {
				c.Data = b.text(block.Data)
				c.MimeType = b.text(block.MimeType)
			}
		}
		out.Content = append(out.Content, c)
	}
	for _, d := range message.Diagnostics {
		if len(out.Diagnostics) >= 8 || b.bytes <= 0 {
			b.truncated = true
			break
		}
		out.Diagnostics = append(out.Diagnostics, goai.AssistantMessageDiagnostic{Type: b.text(d.Type), Timestamp: d.Timestamp, Error: goai.DiagnosticError{Name: b.text(d.Error.Name), Message: b.text(d.Error.Message), Stack: b.text(d.Error.Stack), Code: b.value(d.Error.Code, 0)}, Details: partialDetailMap(b.value(d.Details, 0))})
	}
	if message.Details != nil {
		out.Details = b.value(message.Details, 0)
	}
	if message.Deferred != nil {
		d := message.Deferred
		out.Deferred = &goai.DeferredHandle{Provider: b.text(d.Provider), ModelID: b.text(d.ModelID), Api: b.text(d.Api), ID: b.text(d.ID), ExpiresAt: d.ExpiresAt, PollAfterMs: d.PollAfterMs, Data: b.value(d.Data, 0)}
	}
	return out, b.truncated
}

// The provider callback is the writer: no progress goroutine outlives a request.
// Rate-limited snapshots replace one row. Seal forces the final owned snapshot
// after the producer closes, then fences subsequent callbacks under the mutex.
type providerPartialRecorder struct {
	mu                                     sync.Mutex
	store                                  *store.Store
	ctx                                    context.Context
	turnID, sessionID, occurrenceID, model string
	last                                   time.Time
	stopped                                bool
	err                                    error
}

func (p *providerPartialRecorder) write(message *goai.Message, force bool) error {
	if p.err != nil {
		return p.err
	}
	if message == nil {
		return nil
	}
	if len(message.Content) == 0 && len(message.Diagnostics) == 0 && providerUsageUnknown(message.Usage) {
		return nil
	}
	now := time.Now()
	if !force && !p.last.IsZero() && now.Sub(p.last) < providerPartialInterval {
		return nil
	}
	snapshot, truncated := boundedProviderPartial(message)
	if snapshot == nil {
		return nil
	}
	payload := map[string]any{"model": p.model, "provider_message": snapshot, "partial_checkpoint": map[string]any{"truncated": truncated, "string_bytes_limit": providerPartialStringBytes, "blocks_limit": 64}}
	var thinking []any
	for _, block := range snapshot.Content {
		if block.Type == "thinking" {
			thinking = append(thinking, map[string]any{"type": "thinking", "thinking": block.Thinking, "thinkingSignature": block.ThinkingSignature, "redacted": block.Redacted})
		}
	}
	if len(thinking) > 0 {
		payload["thinking_blocks"] = thinking
	}
	p.err = p.store.SaveProviderPartial(p.ctx, p.turnID, p.sessionID, p.occurrenceID, goai.GetTextContent(snapshot), payload)
	if p.err == nil {
		p.last = now
	}
	return p.err
}
func (p *providerPartialRecorder) Observe(message *goai.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return p.err
	}
	return p.write(message, false)
}
func (p *providerPartialRecorder) Seal(message *goai.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return p.err
	}
	err := p.write(message, true)
	p.stopped = true
	return err
}
