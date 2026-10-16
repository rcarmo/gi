package turn

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

// codemodeTestEngine is the MCP test engine (whose "fake" server has
// codemode tools) plus a plain text tool "greet".
func codemodeTestEngine(t *testing.T) *Engine {
	t.Helper()
	e, _ := mcpTestEngine(t)
	if err := e.tools.Register(tools.RegisteredTool{Name: "greet", Description: "Greets someone",
		Parameters: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`),
		Executor: func(_ context.Context, _ tools.ToolRuntime, call goai.ToolCall) (string, error) {
			return "hello " + call.Arguments["name"].(string), nil
		}}); err != nil {
		t.Fatal(err)
	}
	return e
}

func runCodemode(t *testing.T, e *Engine, sessionID, code string) (string, error) {
	t.Helper()
	return e.executeCodemode(context.Background(), tools.ToolRuntime{Store: e.store, SessionID: sessionID, ToolCallID: "tc_cm", ToolOccurrenceID: "occurrence_cm"},
		goai.ToolCall{ID: "tc_cm", Name: codemodeToolName, Arguments: map[string]any{"code": code}})
}

// An MCP server with codemode tools enables codemode by default (Pi's
// autoEnableCodemode); codemode is model-only and not callable from scripts.
func TestCodemodeAutoEnabledByMCP(t *testing.T) {
	e := codemodeTestEngine(t)
	tool, ok := e.tools.GetRegistered(codemodeToolName)
	if !ok || tool.Deferred || !tool.ModelOnly {
		t.Fatalf("codemode registration: %+v %v", tool, ok)
	}
	for _, c := range e.codemodeCallable(nil) {
		if c.Name == codemodeToolName || c.Name == "tool_search" {
			t.Fatalf("model-only tool callable: %s", c.Name)
		}
	}
}

func TestCodemodeNestedCallsStoreAndDiscovery(t *testing.T) {
	e := codemodeTestEngine(t)
	ctx := context.Background()
	sess, err := e.store.CreateSession(ctx, "ses_codemode", "codemode", nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := runCodemode(t, e, sess.ID, `const r = await tools.mcp__fake__search_code({ text: "x" });
const g = await tools.greet({ name: "bob" });
console.log(g);
store("n", 41);
const found = await searchTools("search code");
const ns = await describeNamespace("fake");
return { mcp: r.content[0].text, found: found[0].name, nsTools: ns.tools.includes("mcp__fake__search_code") };`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Script completed\nWall time", "Output:\nhello bob\n", `"mcp":"found:x"`, `"found":"mcp__fake__search_code"`, `"nsTools":true`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	out, err = runCodemode(t, e, sess.ID, `return load("n") + 1;`)
	if err != nil || !strings.HasSuffix(out, "Output:\n42") {
		t.Fatalf("store not persisted: %q %v", out, err)
	}
}

// Nested calls go through the tool hooks; a blocked call rejects in the
// script, and a failed script reports Pi's error and call summary.
func TestCodemodeHooksAndFailure(t *testing.T) {
	e := codemodeTestEngine(t)
	var parents []any
	if _, err := e.RegisterHook(HookToolCall, "test", func(_ context.Context, req HookRequest) (HookResponse, error) {
		parents = append(parents, req.Payload["parent_tool_call_id"])
		if req.Payload["parent_occurrence_id"] != "occurrence_cm" {
			t.Errorf("missing parent occurrence: %#v", req.Payload)
		}
		if req.ToolCall != nil && req.ToolCall.Name == "greet" {
			return HookResponse{Block: true, Reason: "no greetings"}, nil
		}
		return HookResponse{}, nil
	}); err != nil {
		t.Fatal(err)
	}
	out, err := runCodemode(t, e, "", `try { await tools.greet({ name: "x" }); } catch (e) { text("caught: " + e.message); }
await tools.mcp__fake__search_code({ text: "y" });
throw new Error("boom");`)
	if err == nil {
		t.Fatalf("expected failure, got %q", out)
	}
	msg := err.Error()
	for _, want := range []string{"Script failed\n", "caught: blocked by hook: no greetings", "Script error:\nError: boom", "Tool calls made before the failure (they are not undone): greet (error), mcp__fake__search_code (ok)"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("missing %q in:\n%s", want, msg)
		}
	}
	if len(parents) != 2 || parents[0] != "tc_cm" {
		t.Fatalf("hook parents: %v", parents)
	}
}

// Output past max_output_tokens keeps its start and end; the full output is
// saved in vfs://codemode-output.
func TestCodemodeOutputTruncation(t *testing.T) {
	e := codemodeTestEngine(t)
	out, err := runCodemode(t, e, "", "// @options: {\"max_output_tokens\": 10}\nreturn 'a'.repeat(30) + 'b'.repeat(200) + 'c'.repeat(30);")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Warning: truncated output (original token count: 65)") || !strings.Contains(out, "aaaaaaaaaaaaaaaaaaaa…55 tokens truncated…cccccccccccccccccccc") {
		t.Fatalf("truncation: %s", out)
	}
	i := strings.Index(out, "vfs://codemode-output/")
	if i < 0 {
		t.Fatalf("no saved output: %s", out)
	}
	path := strings.TrimPrefix(strings.Fields(out[i:])[0], "vfs://codemode-output/")
	_, content, err := e.store.GetVFSFileContent(context.Background(), codemodeOutputNamespace, path)
	if err != nil || len(content) != 260 {
		t.Fatalf("saved output: %v %d", err, len(content))
	}
}

// Mode on lists codemode-exposed tools in the description and appends the
// codemode declaration to declared callable tools; mode only drops the
// declarations of direct tools.
func TestCodemodeLoadout(t *testing.T) {
	e := codemodeTestEngine(t)
	ctx := context.Background()
	sess, err := e.store.CreateSession(ctx, "ses_loadout", "loadout", nil)
	if err != nil {
		t.Fatal(err)
	}
	build := func() *goai.Context {
		c := &goai.Context{Tools: []goai.Tool{{Name: "greet", Description: "Greets someone"}, {Name: "mcp__fake__echo", Description: "Echo"}, {Name: codemodeToolName}}}
		e.applyCodemodeLoadout(ctx, c, sess.ID, nil)
		return c
	}
	c := build()
	desc := c.Tools[len(c.Tools)-1].Description
	if !strings.Contains(desc, "## mcp__fake\n") || !strings.Contains(desc, "### `mcp__fake__search_code`") || strings.Contains(desc, "### `greet`") {
		t.Fatalf("mode on description:\n%s", desc)
	}
	if c.Tools[0].Description != "Greets someone\n\nCodemode: `tools.greet(args)` resolves to a string." {
		t.Fatalf("direct tool lacks declaration: %q", c.Tools[0].Description)
	}
	if err := e.SetSessionCodemode(ctx, sess.ID, "only"); err != nil {
		t.Fatal(err)
	}
	c = build()
	if len(c.Tools) != 1 || c.Tools[0].Name != codemodeToolName || !strings.Contains(c.Tools[0].Description, "### `greet`") {
		t.Fatalf("mode only tools: %+v", c.Tools)
	}
	if got := e.applySessionCodemodeToggle(ctx, sess.ID, []string{"read"}); len(got) != 2 {
		t.Fatalf("session on: %v", got)
	}
	_ = e.SetSessionCodemode(ctx, sess.ID, "off")
	if got := e.applySessionCodemodeToggle(ctx, sess.ID, []string{"read", codemodeToolName}); len(got) != 1 {
		t.Fatalf("session off: %v", got)
	}
}

// Scripts report their nested calls as result details for renderers.
func TestCodemodeCallDetails(t *testing.T) {
	e := codemodeTestEngine(t)
	var details map[string]any
	rt := tools.ToolRuntime{Store: e.store, ToolCallID: "tc_cm", SetDetails: func(d map[string]any) { details = d }}
	_, err := e.executeCodemode(context.Background(), rt, goai.ToolCall{ID: "tc_cm", Name: codemodeToolName,
		Arguments: map[string]any{"code": `await tools.greet({ name: "a" }); await tools.mcp__fake__search_code({ text: "q" }); return 1;`}})
	if err != nil {
		t.Fatal(err)
	}
	calls, _ := details["calls"].([]codemodeCallRecord)
	if len(calls) != 2 || calls[0].Name != "greet" || calls[0].Status != "ok" || calls[0].Args != `{"name":"a"}` || calls[1].Name != "mcp__fake__search_code" || calls[1].Status != "ok" {
		t.Fatalf("calls: %+v", calls)
	}
}
