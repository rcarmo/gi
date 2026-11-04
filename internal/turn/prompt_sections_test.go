package turn

import (
	"context"
	"strings"
	"testing"

	goai "github.com/rcarmo/go-ai"

	"github.com/rcarmo/gi/internal/prompt"
)

func sectionMessages(n int) []goai.Message {
	out := make([]goai.Message, n)
	for i := range out {
		out[i] = goai.UserMessage(string(rune('a' + i)))
	}
	return out
}

func sections(pairs ...string) []prompt.Section {
	var out []prompt.Section
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, prompt.Section{Name: pairs[i], Text: pairs[i+1]})
	}
	return out
}

// The sections at session start lead; a change is inserted once, as Pi's
// section update, before the prompt that introduced it and stays there;
// models without mid-conversation system messages get it folded in; a
// compacted-away anchor folds it into the leading prompt.
func TestPromptSectionPlacementLikePi(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	defer e.Close()
	ctx := context.Background()
	if _, err := s.CreateSession(ctx, "sec", "Sections", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	a := sections("preamble", "P", "tools", "<tools>\nA\n</tools>")
	b := sections("preamble", "P", "tools", "<tools>\nB\n</tools>")
	plan := func(current []prompt.Section, ids []string, offset int, mid bool) (string, []goai.Message) {
		t.Helper()
		msgs := sectionMessages(len(ids) + offset)
		p := e.planPromptSections(ctx, "sec", current, ids, offset, msgs)
		return p.apply(p.systemPrompt(), msgs, mid)
	}
	sys, msgs := plan(a, []string{"m1"}, 0, true)
	if sys != "P\n\n<tools>\nA\n</tools>" || len(msgs) != 1 {
		t.Fatalf("new session: %q %d", sys, len(msgs))
	}
	if sys, msgs = plan(a, []string{"m1", "m2", "m3"}, 0, true); len(msgs) != 3 || sys != "P\n\n<tools>\nA\n</tools>" {
		t.Fatal("unchanged sections must not add messages or change the prompt")
	}
	sys, msgs = plan(b, []string{"m1", "m2", "m3"}, 0, true)
	update := "Updated system prompt section \"tools\":\n\n<tools>\nB\n</tools>"
	if sys != "P\n\n<tools>\nA\n</tools>" || len(msgs) != 4 || msgs[2].Role != goai.RoleSystem || msgs[2].Content[0].Text != update {
		t.Fatalf("change: prompt %q messages %+v", sys, msgs)
	}
	// Without mid-conversation system messages the update is folded in.
	if sys, msgs = plan(b, []string{"m1", "m2", "m3"}, 0, false); sys != "P\n\n<tools>\nB\n</tools>" || len(msgs) != 3 {
		t.Fatalf("folded: %q %d", sys, len(msgs))
	}
	// Next turn: the update keeps its position, so the prefix stays cached.
	if _, msgs = plan(b, []string{"m1", "m2", "m3", "m4", "m5"}, 0, true); len(msgs) != 6 || msgs[2].Role != goai.RoleSystem {
		t.Fatalf("update moved: %+v", msgs)
	}
	// Removal.
	if _, msgs = plan(sections("preamble", "P"), []string{"m1", "m2", "m3", "m4", "m5"}, 0, true); len(msgs) != 7 || msgs[5].Content[0].Text != `Removed system prompt section "tools".` {
		t.Fatalf("removal: %+v", msgs)
	}
	// Compaction removed m1..m3 (summary at index 0): the B update leads; the
	// removal keeps its place before m5.
	sys, msgs = plan(sections("preamble", "P"), []string{"m4", "m5"}, 1, true)
	if sys != "P\n\n<tools>\nB\n</tools>" || len(msgs) != 4 || msgs[2].Role != goai.RoleSystem {
		t.Fatalf("compaction: %q %+v", sys, msgs)
	}
	// A hook-replaced prompt is used as is.
	p := e.planPromptSections(ctx, "sec", sections("preamble", "P"), []string{"m4", "m5"}, 1, sectionMessages(3))
	if sys, msgs = p.apply("forced", sectionMessages(3), true); sys != "forced" || len(msgs) != 3 {
		t.Fatalf("forced: %q %d", sys, len(msgs))
	}
	// A mid-turn compaction (anchors moved): the latest sections lead.
	if sys, msgs = p.apply(p.systemPrompt(), []goai.Message{goai.UserMessage("summary-only")}, true); sys != "P" || len(msgs) != 1 {
		t.Fatalf("mid-turn compaction: %q %d", sys, len(msgs))
	}
}

// The default prompt is Pi's structure with gi's preamble; declared tools
// bring their snippets and guidelines (codemode's only while declared).
func TestPromptSectionsFollowDeclaredTools(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	defer e.Close()
	text := prompt.Render(e.promptSections([]goai.Tool{{Name: "read"}, {Name: "edit"}, {Name: "shell"}}))
	for _, want := range []string{
		"You are Gi, an expert coding assistant operating inside gi",
		"<tools>\n- read: Read file contents\n- edit: Make precise file edits",
		"- shell: Execute shell commands",
		"<rules>\n- Use shell for file operations like ls, rg, find\n- Use read to examine files instead of cat or sed.\n- Use edit for precise changes",
		"- Be concise in your responses\n- Show file paths clearly when working with files\n</rules>",
		"<docs>\ngi documentation",
		"<cwd>\n",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "- codemode:") || strings.Contains(text, "Use codemode to batch") {
		t.Fatal("codemode guidance without codemode")
	}
	with := prompt.Render(e.promptSections([]goai.Tool{{Name: "read"}, {Name: codemodeToolName}}))
	if !strings.Contains(with, "- codemode: Run JavaScript that calls other tools") || !strings.Contains(with, "- Use codemode to batch independent tool calls") {
		t.Fatalf("codemode contribution missing:\n%s", with)
	}
	// A custom prompt replaces the preamble, tools, rules and docs (Pi).
	e.systemPrompt = "Custom."
	custom := e.promptSections([]goai.Tool{{Name: "read"}})
	if custom[0].Text != "Custom." || custom[1].Name != "cwd" {
		t.Fatalf("custom: %+v", custom)
	}
}

func TestPromptGuidesScriptAndFailureDiagnosis(t *testing.T) {
	s := openTestStore(t)
	e := New(s)
	defer e.Close()
	text := prompt.Render(e.promptSections([]goai.Tool{{Name: "script"}, {Name: "read"}}))
	for _, want := range []string{"gi.listDir", "gi.listTurnEvents", "console.log", "Await bridge calls in QuickJS", "not OS filesystem paths", "inference.failed/turn.failure_marked", "not claim a runtime fix", "vfs://reference/scoped-models.md", "vfs://reference/tui-selectors.md", "ordered shortlist for model cycling"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing guidance %q", want)
		}
	}
}
