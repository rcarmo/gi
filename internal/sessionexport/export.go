// Package sessionexport writes a gi session as a Pi-compatible JSONL session
// file (format version 3, importable by Pi's /import) or a standalone HTML
// transcript, like Pi's /export.
package sessionexport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
	goai "github.com/rcarmo/go-ai"
)

// Entry is one Pi session line. Fields follow docs/session-format.md.
type Entry map[string]any

type toolCall struct {
	ID, Name       string
	ArgumentsError string
	Arguments      map[string]any
	Preview        bool // arguments reconstructed from a tool.started preview
}

// item is one exported conversation element, before serialization.
type item struct {
	role      string // user | assistant | toolResult | system | compaction | branchSummary | customMessage | bashExecution
	text      string
	calls     []toolCall
	callID    string
	toolName  string
	isError   bool
	model     string
	createdAt time.Time
	sourceID  string
	source    store.Message
	// firstKept is the first message a compaction keeps (its source ID).
	firstKept string
}

// Load reads a session's messages (and tool.started previews for sessions
// recorded before full tool calls were stored) into export items.
func load(ctx context.Context, s *store.Store, sessionID string) ([]item, error) {
	messages, err := s.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	previews := map[string]string{} // tool_call_id -> preview
	turns := map[string]bool{}
	for _, m := range messages {
		if id, _ := m.Payload["turn_id"].(string); id != "" {
			turns[id] = true
		}
	}
	for turnID := range turns {
		events, err := s.ListTurnEvents(ctx, turnID)
		if err != nil {
			continue
		}
		for _, ev := range events {
			if ev.Type != "tool.started" {
				continue
			}
			id, _ := ev.Payload["tool_call_id"].(string)
			preview, _ := ev.Payload["preview"].(string)
			if id != "" {
				previews[id] = preview
			}
		}
	}

	// A compaction keeps the context messages its checkpoint does not cover.
	firstKept := ""
	if snapshot, err := s.ContextSnapshot(ctx, sessionID); err == nil && snapshot.Summary != "" && len(snapshot.Messages) > 0 {
		firstKept = snapshot.Messages[0].ID
	}
	lastCompaction := -1
	for i, m := range messages {
		if kind, _ := m.Payload["kind"].(string); m.Role == "assistant" && kind == "compaction" {
			lastCompaction = i
		}
	}
	var items []item
	for i, m := range messages {
		it := item{text: m.Content, createdAt: parseTime(m.CreatedAt), sourceID: m.ID, source: m}
		it.model, _ = m.Payload["model"].(string)
		kind, _ := m.Payload["kind"].(string)
		switch {
		case m.Role == "user" && kind == "branch_summary":
			it.role = "branchSummary"
		case m.Role == "user" && kind == "custom_message":
			it.role = "customMessage"
		case m.Role == "user" && kind == "bash_execution":
			it.role = "bashExecution"
		case m.Role == "assistant" && kind == "compaction":
			it.role = "compaction"
			// Older compactions are superseded; they keep what follows them.
			if i == lastCompaction && firstKept != "" {
				it.firstKept = firstKept
			} else if i+1 < len(messages) {
				it.firstKept = messages[i+1].ID
			}
		case m.Role == "user":
			it.role = "user"
		case m.Role == "assistant":
			it.role = "assistant"
			if kind == "tool_calls" {
				it.text, _ = m.Payload["display_text"].(string)
				it.calls = recordedCalls(m.Payload)
				if len(it.calls) == 0 {
					it.calls = reconstructCalls(m.Content, messages[i+1:], previews)
				}
			}
		case m.Role == "tool_result":
			it.role = "toolResult"
			it.callID, _ = m.Payload["tool_call_id"].(string)
			it.isError, _ = m.Payload["is_error"].(bool)
		case m.Role == "system":
			it.role = "system"
		default:
			continue
		}
		items = append(items, it)
	}
	// Tool results carry names from their calls.
	names := map[string]string{}
	for _, it := range items {
		for _, c := range it.calls {
			names[c.ID] = c.Name
		}
	}
	for i := range items {
		if items[i].role == "toolResult" {
			items[i].toolName = names[items[i].callID]
		}
	}
	return items, nil
}

func recordedCalls(payload map[string]any) []toolCall {
	raw, ok := payload["tool_calls"].([]any)
	if !ok {
		return nil
	}
	var calls []toolCall
	for _, r := range raw {
		m, _ := r.(map[string]any)
		id, _ := m["id"].(string)
		name, _ := m["name"].(string)
		args, _ := m["arguments"].(map[string]any)
		if id == "" || name == "" {
			continue
		}
		if args == nil {
			args = map[string]any{}
		}
		marker, _ := m["argumentsError"].(string)
		calls = append(calls, toolCall{ID: id, Name: name, Arguments: args, ArgumentsError: marker})
	}
	return calls
}

// reconstructCalls rebuilds calls for sessions recorded before full calls were
// stored: names come from the "[tool_call: name]" summary, IDs from the tool
// results that follow, arguments from the tool.started preview.
func reconstructCalls(summary string, following []store.Message, previews map[string]string) []toolCall {
	var names []string
	for _, line := range strings.Split(summary, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[tool_call: ") && strings.HasSuffix(line, "]") {
			names = append(names, strings.TrimSuffix(strings.TrimPrefix(line, "[tool_call: "), "]"))
		}
	}
	var calls []toolCall
	for _, m := range following {
		if m.Role != "tool_result" || len(calls) == len(names) {
			break
		}
		id, _ := m.Payload["tool_call_id"].(string)
		if id == "" {
			continue
		}
		args := map[string]any{}
		if preview, ok := previews[id]; ok && preview != "" {
			args = previewArguments(names[len(calls)], preview)
		}
		calls = append(calls, toolCall{ID: id, Name: names[len(calls)], Arguments: args, Preview: true})
	}
	return calls
}

func parseTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func entryID(sourceID string) string {
	sum := sha256.Sum256([]byte(sourceID))
	return hex.EncodeToString(sum[:4])
}

// piToolName is a gi tool's name in Pi's sessions: gi's shell tool is Pi's
// bash (the same command argument), so Pi renders it as one.
func piToolName(name string) string {
	if name == "shell" {
		return "bash"
	}
	return name
}

// previewArguments are a legacy call's arguments from its tool.started
// preview (collapsed whitespace, at most 200 characters), under the key Pi's
// tool takes.
func previewArguments(name, preview string) map[string]any {
	switch name {
	case "shell", "bash":
		return map[string]any{"command": preview}
	case "read", "write", "edit", "ls":
		return map[string]any{"path": preview}
	}
	return map[string]any{"preview": preview}
}

func splitModel(label string) (provider, model string) {
	if i := strings.Index(label, "/"); i > 0 {
		return label[:i], label[i+1:]
	}
	return "", label
}

func isoTime(t time.Time) string {
	if t.IsZero() {
		t = time.Unix(0, 0).UTC()
	}
	return t.Format("2006-01-02T15:04:05.000Z")
}

// JSONL serializes a session as Pi session format version 3.
// Document is a session as Pi's session format: the header and the entries,
// with model_change and thinking_level_change entries where the model or
// thinking level of the assistant messages changes.
func Document(ctx context.Context, s *store.Store, sessionID, cwd string) (Entry, []Entry, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}
	items, err := load(ctx, s, sessionID)
	if err != nil {
		return nil, nil, err
	}
	header := Entry{"type": "session", "version": 3, "id": sessionID, "timestamp": isoTime(parseTime(sess.CreatedAt)), "cwd": cwd}
	var entries []Entry
	var parent any // null for the root entry
	emit := func(e Entry, sourceID string, at time.Time) {
		id := entryID(sourceID)
		e["id"], e["parentId"], e["timestamp"] = id, parent, isoTime(at)
		parent = id
		entries = append(entries, e)
	}
	if title := strings.TrimSpace(sess.Title); title != "" {
		emit(Entry{"type": "session_info", "name": title}, sessionID+"#name", parseTime(sess.CreatedAt))
	}
	model, thinking := "", ""
	for _, it := range items {
		e := it.entry()
		if e == nil {
			continue
		}
		if it.role == "assistant" {
			// Pi records model and thinking level changes as entries.
			if it.model != "" && it.model != model {
				provider, id := splitModel(it.model)
				emit(Entry{"type": "model_change", "provider": provider, "modelId": id}, it.sourceID+"#model", it.createdAt)
				model = it.model
			}
			if level, _ := it.source.Payload["thinking_level"].(string); level != "" && level != thinking {
				emit(Entry{"type": "thinking_level_change", "thinkingLevel": level}, it.sourceID+"#thinking", it.createdAt)
				thinking = level
			}
		}
		emit(e, it.sourceID, it.createdAt)
	}
	return header, entries, nil
}

// JSONL is a session as Pi's session file.
func JSONL(ctx context.Context, s *store.Store, sessionID, cwd string) ([]byte, error) {
	header, entries, err := Document(ctx, s, sessionID, cwd)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for _, e := range append([]Entry{header}, entries...) {
		if err := enc.Encode(e); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// entry is the item as a Pi session entry, without id, parentId and
// timestamp; nil for items Pi has no entry for.
func (it item) entry() Entry {
	ms := it.createdAt.UnixMilli()
	var e Entry
	switch it.role {
	case "user":
		e = Entry{"type": "message", "message": map[string]any{"role": "user", "content": it.text, "timestamp": ms}}
	case "assistant":
		content := []any{}
		// Thinking first, as the model produced it.
		if blocks, ok := it.source.Payload["thinking_blocks"].([]any); ok {
			content = append(content, blocks...)
		}
		if strings.TrimSpace(it.text) != "" {
			content = append(content, map[string]any{"type": "text", "text": it.text})
		}
		for _, c := range it.calls {
			block := map[string]any{"type": "toolCall", "id": c.ID, "name": piToolName(c.Name), "arguments": c.Arguments}
			if c.ArgumentsError != "" {
				block["argumentsError"] = c.ArgumentsError
			}
			content = append(content, block)
		}
		provider, model := splitModel(it.model)
		stop, _ := it.source.Payload["stop_reason"].(string)
		if stop == "" {
			stop = "stop"
			if len(it.calls) > 0 {
				stop = "toolUse"
			}
		}
		// Sessions recorded before responses kept their usage export zeros.
		usage, ok := it.source.Payload["usage"].(map[string]any)
		if !ok {
			usage = map[string]any{"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0, "totalTokens": 0, "cost": map[string]any{"input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0, "total": 0}}
		}
		msg := map[string]any{"role": "assistant", "content": content, "provider": provider, "model": model, "stopReason": stop, "timestamp": ms, "usage": usage}
		if text, _ := it.source.Payload["error_message"].(string); text != "" {
			msg["errorMessage"] = text
		}
		if m := goai.GetModel(goai.Provider(provider), model); m != nil {
			msg["api"] = string(m.Api)
		}
		e = Entry{"type": "message", "message": msg}
	case "toolResult":
		e = Entry{"type": "message", "message": map[string]any{"role": "toolResult", "toolCallId": it.callID, "toolName": piToolName(it.toolName),
			"content": []any{map[string]any{"type": "text", "text": it.text}}, "isError": it.isError, "timestamp": ms}}
	case "system":
		// gi notices are not model context: a Pi custom entry, not a message.
		e = Entry{"type": "custom", "customType": "gi.system", "data": map[string]any{"text": it.text}}
	case "compaction":
		tokens, _ := it.source.Payload["tokens_before"].(float64)
		if n, ok := it.source.Payload["tokens_before"].(int); ok {
			tokens = float64(n)
		}
		e = Entry{"type": "compaction", "summary": it.text, "firstKeptEntryId": entryID(it.firstKept), "tokensBefore": int(tokens)}
	case "branchSummary":
		from, _ := it.source.Payload["from_id"].(string)
		e = Entry{"type": "branch_summary", "fromId": entryID(from), "summary": it.text}
	case "customMessage":
		customType, _ := it.source.Payload["custom_type"].(string)
		display, _ := it.source.Payload["display"].(bool)
		e = Entry{"type": "custom_message", "customType": customType, "content": it.text, "display": display}
	case "bashExecution":
		p := it.source.Payload
		msg := map[string]any{"role": "bashExecution", "command": p["command"], "output": p["output"], "exitCode": p["exit_code"],
			"cancelled": p["cancelled"] == true, "truncated": p["truncated"] == true, "timestamp": ms}
		if path, _ := p["full_output_path"].(string); path != "" {
			msg["fullOutputPath"] = path
		}
		if p["exclude_from_context"] == true {
			msg["excludeFromContext"] = true
		}
		e = Entry{"type": "message", "message": msg}
	}
	return e
}

// SourcedEntry is a session message as a Pi session entry (without id,
// parentId and timestamp), with the gi message it came from.
type SourcedEntry struct {
	Entry   Entry
	Message store.Message
}

// Entries reads a session's messages as Pi session entries, in order.
func Entries(ctx context.Context, s *store.Store, sessionID string) ([]SourcedEntry, error) {
	items, err := load(ctx, s, sessionID)
	if err != nil {
		return nil, err
	}
	out := make([]SourcedEntry, 0, len(items))
	for _, it := range items {
		if e := it.entry(); e != nil {
			out = append(out, SourcedEntry{Entry: e, Message: it.source})
		}
	}
	return out, nil
}

// HTML is the session as Pi's HTML export, in the theme's colours.
func HTML(ctx context.Context, s *store.Store, sessionID, cwd string, theme Theme) ([]byte, error) {
	header, entries, err := Document(ctx, s, sessionID, cwd)
	if err != nil {
		return nil, err
	}
	return RenderHTML(header, entries, theme)
}

// Export writes the session to path (resolved against cwd) as JSONL when the
// path ends in .jsonl and as HTML otherwise, like Pi's /export. An empty path
// writes gi-session-<session>.html in cwd. It returns the written path.
func Export(ctx context.Context, s *store.Store, sessionID, cwd, path string, theme Theme) (string, error) {
	if strings.TrimSpace(path) == "" {
		path = "gi-session-" + sessionID + ".html"
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	var data []byte
	var err error
	if strings.HasSuffix(strings.ToLower(path), ".jsonl") {
		data, err = JSONL(ctx, s, sessionID, cwd)
	} else {
		data, err = HTML(ctx, s, sessionID, cwd, theme)
	}
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
