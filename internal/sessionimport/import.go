// Package sessionimport reads a Pi session file (JSONL, format versions 1-3,
// Pi's docs/session-format.md) into a new gi session, like Pi's /import.
// It is the reverse of internal/sessionexport.
package sessionimport

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/store"
)

// Entry is one parsed session line.
type Entry map[string]any

func (e Entry) str(key string) string {
	s, _ := e[key].(string)
	return s
}

// File is a parsed Pi session file.
type File struct {
	Header  Entry
	Entries []Entry // in file order, with ids (older versions get them)
}

// ErrNotSession is a file without a Pi session header.
var ErrNotSession = errors.New("not a Pi session file (no session header)")

// Parse reads a session file as Pi's loadEntriesFromFile does (blank and
// malformed lines are skipped) and migrates versions 1 and 2 as Pi's
// migrateToCurrentVersion does: ids for linear entries, compaction
// firstKeptEntryIndex to firstKeptEntryId, hookMessage to custom.
func Parse(r io.Reader) (*File, error) {
	reader := bufio.NewReader(r)
	f := &File{}
	for {
		line, err := reader.ReadString('\n')
		if strings.TrimSpace(line) != "" {
			var e Entry
			if json.Unmarshal([]byte(line), &e) == nil && e != nil {
				if f.Header == nil && e.str("type") == "session" {
					f.Header = e
				} else if f.Header != nil {
					f.Entries = append(f.Entries, e)
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	if f.Header == nil {
		return nil, ErrNotSession
	}
	version, _ := f.Header["version"].(float64)
	if version < 2 {
		prev := ""
		for i, e := range f.Entries {
			if e.str("id") == "" {
				e["id"] = fmt.Sprintf("%08x", i+1)
			}
			if _, ok := e["parentId"]; !ok {
				if prev == "" {
					e["parentId"] = nil
				} else {
					e["parentId"] = prev
				}
			}
			if index, ok := e["firstKeptEntryIndex"].(float64); ok && e.str("type") == "compaction" {
				if int(index) >= 0 && int(index) < len(f.Entries) {
					e["firstKeptEntryId"] = f.Entries[int(index)].str("id")
				}
				delete(e, "firstKeptEntryIndex")
			}
			prev = e.str("id")
		}
	}
	for _, e := range f.Entries {
		if msg, ok := e["message"].(map[string]any); ok && msg["role"] == "hookMessage" {
			msg["role"] = "custom"
		}
	}
	return f, nil
}

// Path is the active branch: from the root to the last entry (Pi's leaf),
// following parentId.
func (f *File) Path() []Entry {
	if len(f.Entries) == 0 {
		return nil
	}
	byID := map[string]Entry{}
	for _, e := range f.Entries {
		byID[e.str("id")] = e
	}
	var path []Entry
	seen := map[string]bool{}
	for e := f.Entries[len(f.Entries)-1]; e != nil && !seen[e.str("id")]; {
		seen[e.str("id")] = true
		path = append([]Entry{e}, path...)
		parent, _ := e["parentId"].(string)
		e = byID[parent]
	}
	return path
}

// Name is the session name: the latest session_info entry in the file (Pi's
// getSessionName; an empty name clears it).
func (f *File) Name() string {
	for i := len(f.Entries) - 1; i >= 0; i-- {
		if e := f.Entries[i]; e.str("type") == "session_info" {
			return strings.TrimSpace(e.str("name"))
		}
	}
	return ""
}

// Settings are the model and thinking level of the active branch (Pi's
// getSessionContextSettings): model changes and assistant messages set the
// model, thinking level changes the level. Without one, thinking is "" and
// the default applies (as Pi's createAgentSession does).
func Settings(path []Entry) (provider, model, thinking string) {
	for _, e := range path {
		switch e.str("type") {
		case "thinking_level_change":
			thinking = e.str("thinkingLevel")
		case "model_change":
			provider, model = e.str("provider"), e.str("modelId")
		case "message":
			if msg, _ := e["message"].(map[string]any); msg["role"] == "assistant" {
				provider, _ = msg["provider"].(string)
				model, _ = msg["model"].(string)
			}
		}
	}
	return provider, model, thinking
}

// contentText joins a message's text blocks (or its string content).
func contentText(content any) string {
	switch c := content.(type) {
	case string:
		return c
	case []any:
		var parts []string
		for _, raw := range c {
			if b, _ := raw.(map[string]any); b["type"] == "text" {
				text, _ := b["text"].(string)
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// images are a message's image blocks.
func images(content any) []map[string]any {
	blocks, _ := content.([]any)
	var out []map[string]any
	for _, raw := range blocks {
		if b, _ := raw.(map[string]any); b["type"] == "image" {
			out = append(out, b)
		}
	}
	return out
}

// bashExecutionText is Pi's bashExecutionToText: how a ! command reaches the
// model.
func bashExecutionText(msg map[string]any) string {
	command, _ := msg["command"].(string)
	output, _ := msg["output"].(string)
	text := "Ran `" + command + "`\n"
	if output != "" {
		text += "```\n" + output + "\n```"
	} else {
		text += "(no output)"
	}
	if msg["cancelled"] == true {
		text += "\n\n(command cancelled)"
	} else if code, ok := msg["exitCode"].(float64); ok && code != 0 {
		text += fmt.Sprintf("\n\nCommand exited with code %d", int(code))
	}
	if path, _ := msg["fullOutputPath"].(string); msg["truncated"] == true && path != "" {
		text += "\n\n[Output truncated. Full output: " + path + "]"
	}
	return text
}

// Result is what an import created.
type Result struct {
	Session  *store.Session
	Messages int
}

// Import fills sessionID, a new empty gi session, from the file's active
// branch (source is the imported path, recorded in the session state):
//   - user, assistant and toolResult messages become gi messages (full tool
//     calls in the tool_calls payload, images as media);
//   - bashExecution messages, custom_message and branch_summary entries become
//     user messages the model sees as Pi's would;
//   - the latest compaction becomes gi's compaction message and context
//     checkpoint (its summary in place of the messages before
//     firstKeptEntryId);
//   - context_edit entries replace (or remove) the content of their targets;
//   - gi notices (custom gi.system entries) become system messages, other
//     custom entries are extension state and are skipped;
//   - labels become /tree labels; the name, model and thinking level are the
//     session's.
func Import(ctx context.Context, s *store.Store, f *File, sessionID, source string) (*Result, error) {
	path := f.Path()
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	state := map[string]any{}
	for k, v := range sess.State {
		state[k] = v
	}
	if provider, model, thinking := Settings(path); model != "" || thinking != "" {
		if model != "" {
			state["model"], state["provider"] = model, provider
			if provider != "" {
				state["model"] = provider + "/" + model
			}
		}
		if thinking != "" {
			state["thinking_level"] = thinking
		}
	}
	state["imported_from"] = source
	if id := f.Header.str("id"); id != "" {
		state["pi_session_id"] = id
	}
	if err := s.SetSessionState(ctx, sessionID, state); err != nil {
		return nil, err
	}
	if name := f.Name(); name != "" {
		if err := s.UpdateSessionTitle(ctx, sessionID, name); err != nil {
			return nil, err
		}
	}

	edits := map[string]Entry{}
	latestCompaction := -1
	for i, e := range path {
		switch e.str("type") {
		case "context_edit":
			edits[e.str("targetId")] = e
		case "compaction":
			latestCompaction = i
		}
	}
	firstKept := -1
	if latestCompaction >= 0 {
		keep := path[latestCompaction].str("firstKeptEntryId")
		for i := 0; i < latestCompaction; i++ {
			if path[i].str("id") == keep {
				firstKept = i
				break
			}
		}
		if firstKept < 0 {
			firstKept = latestCompaction // nothing kept
		}
	}

	var messages []store.ImportedMessage
	var covered []string
	ids := map[string]string{} // Pi entry id -> gi message id
	base := time.Now().UnixNano()
	last := ""
	add := func(e Entry, role, content string, payload map[string]any, inContext bool, index int) {
		if payload == nil {
			payload = map[string]any{}
		}
		payload["pi_entry_id"] = e.str("id")
		id := fmt.Sprintf("msg_%d", base+int64(len(messages)))
		at := isoTime(e)
		if at < last {
			at = last
		}
		last = at
		messages = append(messages, store.ImportedMessage{ID: id, Role: role, Content: content, Payload: payload, CreatedAt: at})
		ids[e.str("id")] = id
		if inContext && index < firstKept {
			covered = append(covered, id)
		}
	}
	thinkingLevel := "" // the branch's level so far, recorded on responses
	for i, e := range path {
		if e.str("type") == "thinking_level_change" {
			thinkingLevel = e.str("thinkingLevel")
		}
		inContext := true
		if edit, ok := edits[e.str("id")]; ok {
			replacement, _ := edit["replacement"].(map[string]any)
			if replacement == nil {
				continue // removed from context; gi keeps one copy, so it goes
			}
			e = applyEdit(e, replacement["content"])
		}
		switch e.str("type") {
		case "message":
			msg, _ := e["message"].(map[string]any)
			switch role, _ := msg["role"].(string); role {
			case "user":
				payload := map[string]any{"kind": "chat", "source": "import"}
				if refs := storeImages(ctx, s, sessionID, msg["content"]); len(refs) > 0 {
					payload["media"] = refs
				}
				add(e, "user", contentText(msg["content"]), payload, inContext, i)
			case "assistant":
				payload := assistantPayload(msg)
				if thinkingLevel != "" {
					payload["thinking_level"] = thinkingLevel
				}
				add(e, "assistant", "", payload, inContext, i)
				last := &messages[len(messages)-1]
				last.Content = assistantContent(msg, last.Payload)
			case "toolResult":
				callID, _ := msg["toolCallId"].(string)
				name, _ := msg["toolName"].(string)
				add(e, "tool_result", contentText(msg["content"]), map[string]any{"kind": "tool_result", "tool_call_id": callID, "tool_name": giToolName(name), "is_error": msg["isError"] == true}, inContext, i)
			case "bashExecution":
				payload := map[string]any{"kind": "bash_execution", "command": msg["command"], "output": msg["output"], "exit_code": msg["exitCode"],
					"cancelled": msg["cancelled"] == true, "truncated": msg["truncated"] == true}
				if path, _ := msg["fullOutputPath"].(string); path != "" {
					payload["full_output_path"] = path
				}
				role := "user"
				if msg["excludeFromContext"] == true { // Pi's !! commands stay out of context
					payload["exclude_from_context"] = true
					role = "system"
				}
				add(e, role, bashExecutionText(msg), payload, role == "user", i)
			case "custom":
				customType, _ := msg["customType"].(string)
				add(e, "user", contentText(msg["content"]), map[string]any{"kind": "custom_message", "custom_type": customType, "display": msg["display"] == true}, inContext, i)
			}
		case "custom_message":
			add(e, "user", contentText(e["content"]), map[string]any{"kind": "custom_message", "custom_type": e.str("customType"), "display": e["display"] == true}, inContext, i)
		case "branch_summary":
			if e.str("summary") == "" {
				continue
			}
			add(e, "user", e.str("summary"), map[string]any{"kind": "branch_summary", "from_id": ids[e.str("fromId")]}, inContext, i)
		case "compaction":
			tokens, _ := e["tokensBefore"].(float64)
			add(e, "assistant", e.str("summary"), map[string]any{"kind": "compaction", "tokens_before": int(tokens), "source": "import"}, false, i)
		case "custom":
			if e.str("customType") == "gi.system" {
				data, _ := e["data"].(map[string]any)
				text, _ := data["text"].(string)
				add(e, "system", text, map[string]any{"kind": "notice"}, false, i)
			}
		}
	}
	if err := s.AddImportedMessages(ctx, sessionID, messages); err != nil {
		return nil, err
	}
	if latestCompaction >= 0 && len(covered) > 0 {
		if err := s.SetContextCheckpoint(ctx, sessionID, path[latestCompaction].str("summary"), covered); err != nil {
			return nil, fmt.Errorf("import compaction: %w", err)
		}
	}
	for _, e := range path {
		if e.str("type") != "label" {
			continue
		}
		if target := ids[e.str("targetId")]; target != "" {
			at, _ := time.Parse(time.RFC3339Nano, e.str("timestamp"))
			if err := s.SetTreeLabel(ctx, sessionID, target, strings.TrimSpace(e.str("label")), at); err != nil {
				return nil, err
			}
		}
	}
	if sess, err = s.GetSession(ctx, sessionID); err != nil {
		return nil, err
	}
	return &Result{Session: sess, Messages: len(messages)}, nil
}

// ImportFile parses and imports a session file into sessionID.
func ImportFile(ctx context.Context, s *store.Store, path, sessionID string) (*Result, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	f, err := Parse(fh)
	if err != nil {
		return nil, err
	}
	return Import(ctx, s, f, sessionID, path)
}

func applyEdit(e Entry, content any) Entry {
	out := Entry{}
	for k, v := range e {
		out[k] = v
	}
	if msg, ok := e["message"].(map[string]any); ok {
		copied := map[string]any{}
		for k, v := range msg {
			copied[k] = v
		}
		copied["content"] = content
		out["message"] = copied
	} else {
		out["content"] = content
	}
	return out
}

func assistantPayload(msg map[string]any) map[string]any {
	provider, _ := msg["provider"].(string)
	model, _ := msg["model"].(string)
	label := model
	if provider != "" {
		label = provider + "/" + model
	}
	payload := map[string]any{"kind": "chat", "source": "import", "model": label}
	var calls []any
	blocks, _ := msg["content"].([]any)
	for _, raw := range blocks {
		b, _ := raw.(map[string]any)
		if b["type"] == "toolCall" {
			args, _ := b["arguments"].(map[string]any)
			if args == nil {
				args = map[string]any{}
			}
			calls = append(calls, map[string]any{"id": b["id"], "name": giToolName(b["name"]), "arguments": args, "argumentsError": b["argumentsError"]})
		}
	}
	if len(calls) > 0 {
		payload["kind"] = "tool_calls"
		payload["tool_calls"] = calls
		payload["display_text"] = strings.TrimSpace(contentText(msg["content"]))
	}
	if stop, _ := msg["stopReason"].(string); stop != "" {
		payload["stop_reason"] = stop
	}
	if text, _ := msg["errorMessage"].(string); text != "" {
		payload["error_message"] = text
	}
	if usage, ok := msg["usage"].(map[string]any); ok {
		payload["usage"] = usage
	}
	var thinking []any
	for _, raw := range blocks {
		if b, _ := raw.(map[string]any); b["type"] == "thinking" {
			thinking = append(thinking, b)
		}
	}
	if len(thinking) > 0 {
		payload["thinking_blocks"] = thinking
	}
	return payload
}

// giToolName is a Pi tool's name in gi: Pi's bash is gi's shell.
func giToolName(name any) any {
	if name == "bash" {
		return "shell"
	}
	return name
}

// assistantContent is the message as gi records it: the text, with a
// "[tool_call: name]" line per call.
func assistantContent(msg map[string]any, payload map[string]any) string {
	text := contentText(msg["content"])
	calls, _ := payload["tool_calls"].([]any)
	if len(calls) == 0 {
		return text
	}
	summary := strings.TrimSpace(text)
	for _, raw := range calls {
		call, _ := raw.(map[string]any)
		if summary != "" {
			summary += "\n"
		}
		summary += fmt.Sprintf("[tool_call: %v]", call["name"])
	}
	return summary
}

// storeImages saves a message's images as session media.
func storeImages(ctx context.Context, s *store.Store, sessionID string, content any) []string {
	var refs []string
	for i, img := range images(content) {
		data, _ := img["data"].(string)
		mime, _ := img["mimeType"].(string)
		raw, err := base64.StdEncoding.DecodeString(data)
		if err != nil || len(raw) == 0 {
			continue
		}
		ext := strings.TrimPrefix(mime, "image/")
		media, err := s.CreateMedia(ctx, sessionID, fmt.Sprintf("imported-%d.%s", i+1, ext), mime, raw, map[string]any{"source": "import"})
		if err != nil {
			continue
		}
		refs = append(refs, store.MediaRefID(media.ID))
	}
	return refs
}

// isoTime is an entry's time as gi stores it.
func isoTime(e Entry) string {
	t, err := time.Parse(time.RFC3339Nano, e.str("timestamp"))
	if err != nil {
		t = time.Now()
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
