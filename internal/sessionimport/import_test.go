package sessionimport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/sessionexport"
	"github.com/rcarmo/gi/internal/store"
)

type golden struct {
	JSONL            string
	Path, Context    []string
	Name             *string
	Model            *struct{ Provider, ModelID string } `json:"model"`
	ThinkingLevel    string
	HasThinkingEntry bool
	Labels           map[string]string
}

func loadGolden(t *testing.T) map[string]golden {
	t.Helper()
	raw, err := os.ReadFile("testdata/pi-sessions.json")
	if err != nil {
		t.Fatal(err)
	}
	var g map[string]golden
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "gi.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func importText(t *testing.T, s *store.Store, id, jsonl string) *Result {
	t.Helper()
	path := filepath.Join(t.TempDir(), id+".jsonl")
	if err := os.WriteFile(path, []byte(jsonl), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(context.Background(), id, "@agent", map[string]any{"status": "idle", "thinking_level": "medium"}); err != nil {
		t.Fatal(err)
	}
	r, err := ImportFile(context.Background(), s, path, id)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// Pi session files built by Pi's own SessionManager import with the model
// context Pi builds from them (scripts/golden-session-import.mjs): the active
// branch, the latest compaction's summary in place of the messages before
// its first kept entry, context edits, branch summaries, custom messages and
// ! commands, and Pi's name, model, thinking level and labels.
func TestImportMatchesPiContext(t *testing.T) {
	ctx := context.Background()
	for name, g := range loadGolden(t) {
		t.Run(name, func(t *testing.T) {
			s := openStore(t)
			r := importText(t, s, "s_"+name, g.JSONL)
			f, err := Parse(strings.NewReader(g.JSONL))
			if err != nil {
				t.Fatal(err)
			}
			entries := map[string]Entry{}
			for _, e := range f.Entries {
				entries[e.str("id")] = e
			}
			var gotPath []string
			for _, e := range f.Path() {
				gotPath = append(gotPath, e.str("id"))
			}
			if strings.Join(gotPath, ",") != strings.Join(g.Path, ",") {
				t.Fatalf("path %v, Pi %v", gotPath, g.Path)
			}
			// The context entries gi's model sees: Pi's, less those that project
			// to nothing (state entries, excluded ! commands).
			var want []string
			summary := ""
			for _, id := range g.Context {
				e := entries[id]
				msg, _ := e["message"].(map[string]any)
				switch e.str("type") {
				case "compaction":
					summary = e.str("summary")
				case "custom_message", "branch_summary":
					want = append(want, id)
				case "message":
					if role := msg["role"]; role == "user" || role == "assistant" || (role == "bashExecution" && msg["excludeFromContext"] != true) || role == "custom" {
						want = append(want, id)
					}
				}
			}
			snapshot, err := s.ContextSnapshot(ctx, r.Session.ID)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, m := range snapshot.Messages {
				got = append(got, fmt.Sprint(m.Payload["pi_entry_id"]))
			}
			if strings.Join(got, ",") != strings.Join(want, ",") || snapshot.Summary != summary {
				t.Fatalf("context %v summary %q\nPi %v summary %q", got, snapshot.Summary, want, summary)
			}
			if wantName := "@agent"; g.Name != nil && r.Session.Title != *g.Name || g.Name == nil && r.Session.Title != wantName {
				t.Fatalf("title %q, Pi %v", r.Session.Title, g.Name)
			}
			if model := g.Model.Provider + "/" + g.Model.ModelID; r.Session.State["model"] != model {
				t.Fatalf("model %v, Pi %s", r.Session.State["model"], model)
			}
			wantThinking := "medium" // the session's default without a change entry
			if g.HasThinkingEntry {
				wantThinking = g.ThinkingLevel
			}
			if r.Session.State["thinking_level"] != wantThinking {
				t.Fatalf("thinking %v, want %s", r.Session.State["thinking_level"], wantThinking)
			}
			messages, err := s.ListMessages(ctx, r.Session.ID)
			if err != nil {
				t.Fatal(err)
			}
			byEntry := map[string]store.Message{}
			for _, m := range messages {
				byEntry[fmt.Sprint(m.Payload["pi_entry_id"])] = m
			}
			labels := store.TreeLabels(*r.Session)
			if len(labels) != len(g.Labels) {
				t.Fatalf("labels %v, Pi %v", labels, g.Labels)
			}
			for entryID, label := range g.Labels {
				if labels[byEntry[entryID].ID].Label != label {
					t.Fatalf("label of %s: %v, Pi %q", entryID, labels[byEntry[entryID].ID], label)
				}
			}
		})
	}
}

// The branched golden's details: tool calls with full arguments, tool
// results, an image as media, a context edit, ! commands as Pi's text.
func TestImportMapsMessages(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	r := importText(t, s, "s1", loadGolden(t)["branched"].JSONL)
	messages, err := s.ListMessages(ctx, r.Session.ID)
	if err != nil {
		t.Fatal(err)
	}
	find := func(pred func(store.Message) bool) store.Message {
		t.Helper()
		for _, m := range messages {
			if pred(m) {
				return m
			}
		}
		t.Fatal("no such message")
		return store.Message{}
	}
	call := find(func(m store.Message) bool { return m.Payload["kind"] == "tool_calls" })
	calls, _ := call.Payload["tool_calls"].([]any)
	first, _ := calls[0].(map[string]any)
	if call.Content != "Reading.\n[tool_call: read]" || call.Payload["display_text"] != "Reading." || first["id"] != "t1" || fmt.Sprint(first["arguments"]) != "map[path:src/parse.go]" || call.Payload["model"] != "anthropic/claude-sonnet-4-5" {
		t.Fatalf("tool call %+v", call)
	}
	if blocks, _ := call.Payload["thinking_blocks"].([]any); len(blocks) != 1 || call.Payload["stop_reason"] != "toolUse" || call.Payload["usage"] == nil {
		t.Fatalf("response record %+v", call.Payload)
	}
	result := find(func(m store.Message) bool { return m.Role == "tool_result" })
	if result.Content != "package parse" || result.Payload["tool_call_id"] != "t1" || result.Payload["tool_name"] != "read" {
		t.Fatalf("tool result %+v", result)
	}
	image := find(func(m store.Message) bool { return m.Content == "Look at this" })
	if refs, err := store.NormalizeMediaReferences(image.Payload["media"]); err != nil || len(refs) != 1 {
		t.Fatalf("image %+v %v", image.Payload, err)
	}
	if find(func(m store.Message) bool {
		return strings.Contains(m.Content, "redacted") || strings.Contains(m.Content, "hunter2")
	}).Content != "[redacted]" {
		t.Fatal("context edit not applied")
	}
	bash := find(func(m store.Message) bool { return m.Payload["kind"] == "bash_execution" && m.Role == "user" })
	if bash.Content != "Ran `go test ./...`\n```\nok\n```" {
		t.Fatalf("bash %q", bash.Content)
	}
	if excluded := find(func(m store.Message) bool { return m.Payload["exclude_from_context"] == true }); excluded.Role != "system" {
		t.Fatalf("excluded command %+v", excluded)
	}
	if notice := find(func(m store.Message) bool { return m.Role == "system" && m.Content == "Queued prompt" }); notice.ID == "" {
		t.Fatal("gi notice lost")
	}
	for _, m := range messages {
		if strings.Contains(m.Content, "Try another way") || strings.Contains(m.Content, "count") {
			t.Fatalf("off-branch or state entry imported: %+v", m)
		}
	}
}

// normalized is an export with entry ids replaced by their position and the
// session's own id and creation time (header, name entry) dropped: an import
// is a new gi session.
func normalized(t *testing.T, jsonl []byte) string {
	t.Helper()
	ids := map[string]string{}
	var lines []string
	for i, line := range strings.Split(strings.TrimSpace(string(jsonl)), "\n") {
		var e map[string]any
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		if id, _ := e["id"].(string); id != "" && e["type"] != "session" {
			ids[id] = fmt.Sprintf("e%d", i)
		}
		lines = append(lines, line)
	}
	for i, line := range lines {
		var e map[string]any
		_ = json.Unmarshal([]byte(line), &e)
		if e["type"] == "session" {
			delete(e, "id")
		}
		if e["type"] == "session" || e["type"] == "session_info" {
			delete(e, "timestamp")
		}
		for _, key := range []string{"id", "parentId", "firstKeptEntryId", "fromId"} {
			if v, ok := e[key].(string); ok && ids[v] != "" {
				e[key] = ids[v]
			}
		}
		out, _ := json.Marshal(e)
		lines[i] = string(out)
	}
	return strings.Join(lines, "\n")
}

// gi's export imports back to the same session: export → import → export
// is stable (ids aside).
func TestExportImportRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	r := importText(t, s, "s1", loadGolden(t)["branched"].JSONL)
	first, err := sessionexport.JSONL(ctx, s, r.Session.ID, "/home/u/project")
	if err != nil {
		t.Fatal(err)
	}
	again := importText(t, s, "s2", string(first))
	second, err := sessionexport.JSONL(ctx, s, again.Session.ID, "/home/u/project")
	if err != nil {
		t.Fatal(err)
	}
	if a, b := normalized(t, first), normalized(t, second); a != b {
		t.Fatalf("round trip changed the export:\n%s\n---\n%s", a, b)
	}
	// The exported compaction keeps what Pi's did.
	if !strings.Contains(string(first), `"type":"compaction"`) || !strings.Contains(string(first), `"type":"branch_summary"`) {
		t.Fatalf("export lacks compaction or branch summary:\n%s", first)
	}
}

// Pi's bash is gi's shell, and back on export.
func TestImportMapsBashToShell(t *testing.T) {
	ctx := context.Background()
	s := openStore(t)
	jsonl := `{"type":"session","version":3,"id":"x","timestamp":"2026-10-03T10:00:00.000Z","cwd":"/w"}
{"type":"message","id":"a","parentId":null,"timestamp":"2026-10-03T10:00:01.000Z","message":{"role":"assistant","content":[{"type":"toolCall","id":"t","name":"bash","arguments":{"command":"ls"}}],"provider":"openai","model":"gpt-5","stopReason":"toolUse","timestamp":1}}
{"type":"message","id":"b","parentId":"a","timestamp":"2026-10-03T10:00:02.000Z","message":{"role":"toolResult","toolCallId":"t","toolName":"bash","content":[{"type":"text","text":"a"}],"isError":false,"timestamp":2}}
`
	r := importText(t, s, "s1", jsonl)
	messages, _ := s.ListMessages(ctx, r.Session.ID)
	calls, _ := messages[0].Payload["tool_calls"].([]any)
	if call, _ := calls[0].(map[string]any); call["name"] != "shell" || messages[0].Content != "[tool_call: shell]" || messages[1].Payload["tool_name"] != "shell" {
		t.Fatalf("messages %+v", messages)
	}
	out, err := sessionexport.JSONL(ctx, s, r.Session.ID, "/w")
	if err != nil || strings.Count(string(out), `"bash"`) != 2 || strings.Contains(string(out), `"shell"`) {
		t.Fatalf("export %s %v", out, err)
	}
}

func TestParseRejectsNonSessions(t *testing.T) {
	if _, err := Parse(strings.NewReader("{\"type\":\"message\"}\nnot json\n")); err != ErrNotSession {
		t.Fatalf("err %v", err)
	}
}

func TestAssistantPayloadRetainsArgumentsError(t *testing.T) {
	payload := assistantPayload(map[string]any{"content": []any{map[string]any{"type": "toolCall", "id": "bad", "name": "shell", "arguments": map[string]any{}, "argumentsError": "Expected object"}}})
	calls := payload["tool_calls"].([]any)
	if calls[0].(map[string]any)["argumentsError"] != "Expected object" {
		t.Fatalf("parse error lost: %+v", payload)
	}
}
