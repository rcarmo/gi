package turn

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/rcarmo/gi/internal/codemode"
	gimcp "github.com/rcarmo/gi/internal/mcp"
	"github.com/rcarmo/gi/internal/tools"
	goai "github.com/rcarmo/go-ai"
)

// The codemode tool (#25 phase 5), ported from Pi's codemode extension. The
// model writes JavaScript that calls other tools; it runs in the QuickJS
// engine (internal/codemode). Nested calls go through the tool hooks
// (tool_call, approve_tool, tool_result) like direct calls, but only the
// script's output reaches the model.

const (
	codemodeToolName          = "codemode"
	codemodeOutputNamespace   = "codemode-output"
	codemodeStoreStateKey     = "codemode_store"
	codemodeModeStateKey      = "codemode_mode" // session toggle: on, off, only
	codemodeMemoryLimit       = 256 * 1024 * 1024
	codemodeDefaultMaxTokens  = 10_000
	codemodeCharsPerToken     = 4
	codemodeNoTimeout         = 24 * time.Hour // Pi: no deadline by default
	codemodeErrorPreviewChars = 500
)

// codemodeParameters is Pi's codemodeSchema.
var codemodeParameters = json.RawMessage(`{"type":"object","properties":{"code":{"type":"string","description":` +
	mustJSON(codemode.Texts.CodeDescription) + `}},"required":["code"]}`)

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// codemodeState tracks whether codemode is declared by default.
type codemodeState struct {
	mu        sync.Mutex
	defaultOn bool
	renders   sync.Map // declarations hash -> codemode.Rendered
}

// registerCodemodeTool registers codemode unless the built-in is disabled
// (extensions: -builtin:codemode). It is declared by default when
// defaultTools enables it; otherwise sessions (or MCP auto-activation)
// enable it.
func (e *Engine) registerCodemodeTool() {
	if e.runtimeCfg.BuiltinDisabled(codemodeToolName) {
		return
	}
	e.codemode = &codemodeState{defaultOn: e.runtimeCfg.ToolEnabledByDefault(codemodeToolName, false)}
	e.registerCodemode()
}

func (e *Engine) registerCodemode() {
	if err := e.tools.Register(tools.RegisteredTool{
		Name: codemodeToolName, Description: codemode.Description(nil, codemode.Rendered{}, codemode.DescriptionOptions{}),
		Parameters: codemodeParameters, Source: "builtin", Kind: "mixed", ModelOnly: true,
		Deferred: !e.codemode.defaultOn, Executor: e.executeCodemode,
	}); err != nil {
		log.Printf("codemode: register: %v", err)
	}
}

// autoEnableCodemode declares codemode by default when an MCP server has
// codemode exposure (Pi), unless autoEnableCodemode is false or defaultTools
// removes it.
func (e *Engine) autoEnableCodemode(cfg gimcp.Config) {
	if e.codemode == nil || !cfg.AutoEnableCodemode || !e.runtimeCfg.ToolEnabledByDefault(codemodeToolName, true) {
		if warning := e.CodemodeWarning(context.Background(), ""); warning != "" {
			log.Printf("codemode: %s (enable with /codemode on)", warning) // warned once, at startup
		}
		return
	}
	for _, name := range cfg.Names() {
		sc := cfg.Servers[name]
		if sc.Enabled && sc.HasCodemodeTools() {
			e.codemode.mu.Lock()
			changed := !e.codemode.defaultOn
			e.codemode.defaultOn = true
			e.codemode.mu.Unlock()
			if changed {
				e.registerCodemode()
			}
			return
		}
	}
}

// sessionCodemodeMode is the session toggle (/codemode): on, off, only or "".
func (e *Engine) sessionCodemodeMode(ctx context.Context, sessionID string) string {
	if sessionID == "" {
		return ""
	}
	sess, err := e.store.GetSession(ctx, sessionID)
	if err != nil {
		return ""
	}
	mode, _ := sess.State[codemodeModeStateKey].(string)
	return mode
}

// SetSessionCodemode sets the session toggle; "" clears it (settings apply).
func (e *Engine) SetSessionCodemode(ctx context.Context, sessionID, mode string) error {
	switch mode {
	case "", "on", "off", "only":
	default:
		return fmt.Errorf("codemode mode must be on, off or only")
	}
	if e.codemode == nil && mode != "off" && mode != "" {
		return errors.New("codemode is disabled (extensions: -builtin:codemode)")
	}
	var value any = mode
	if mode == "" {
		value = nil
	}
	return e.store.TouchSessionState(ctx, sessionID, map[string]any{codemodeModeStateKey: value})
}

// CodemodeStatus reports whether codemode is declared for a session and its
// presentation mode.
func (e *Engine) CodemodeStatus(ctx context.Context, sessionID string) (enabled bool, mode, source string) {
	if e.codemode == nil {
		return false, "off", "extensions: -builtin:codemode"
	}
	mode = e.codemodePresentation(ctx, sessionID)
	switch e.sessionCodemodeMode(ctx, sessionID) {
	case "off":
		return false, mode, "session"
	case "on", "only":
		return true, mode, "session"
	}
	e.codemode.mu.Lock()
	on := e.codemode.defaultOn
	e.codemode.mu.Unlock()
	return on, mode, "settings"
}

// CodemodeWarning explains why MCP codemode tools are unreachable: they
// are only callable from codemode scripts.
func (e *Engine) CodemodeWarning(ctx context.Context, sessionID string) string {
	if e.mcp == nil {
		return ""
	}
	if enabled, _, _ := e.CodemodeStatus(ctx, sessionID); enabled {
		return ""
	}
	var servers []string
	cfg := e.mcp.manager.Config()
	for _, name := range cfg.Names() {
		if sc := cfg.Servers[name]; sc.Enabled && sc.HasCodemodeTools() {
			servers = append(servers, name)
		}
	}
	if len(servers) == 0 {
		return ""
	}
	return "warning: codemode is off, so the codemode tools of MCP server(s) " + strings.Join(servers, ", ") + " cannot be called"
}

// codemodePresentation is codemode.mode: the session's "only" wins, else
// the settings value.
func (e *Engine) codemodePresentation(ctx context.Context, sessionID string) string {
	if e.sessionCodemodeMode(ctx, sessionID) == "only" || e.runtimeCfg.Codemode.Mode == "only" {
		return "only"
	}
	return "on"
}

// applySessionCodemodeToggle adjusts a turn's admitted tool set for the
// session toggle.
func (e *Engine) applySessionCodemodeToggle(ctx context.Context, sessionID string, names []string) []string {
	if e.codemode == nil {
		return names
	}
	switch e.sessionCodemodeMode(ctx, sessionID) {
	case "off":
		out := names[:0:0]
		for _, n := range names {
			if n != codemodeToolName {
				out = append(out, n)
			}
		}
		return out
	case "on", "only":
		for _, n := range names {
			if n == codemodeToolName {
				return names
			}
		}
		return append(append([]string(nil), names...), codemodeToolName)
	}
	return names
}

// codemodeCallable lists the tools a script may call: the turn's active
// tools plus every deferred tool (codemode/deferred MCP tools), except
// model-only ones (codemode, tool_search).
func (e *Engine) codemodeCallable(metadata map[string]any) []tools.RegisteredTool {
	allowed := tools.EffectiveToolNameSetFromMetadata(metadata)
	var out []tools.RegisteredTool
	for _, t := range e.tools.AllEntries() {
		if t.ModelOnly || t.Name == codemodeToolName {
			continue
		}
		if t.Deferred || len(allowed) == 0 || allowed[t.Name] {
			out = append(out, t)
		}
	}
	return out
}

func toDeclaration(t tools.RegisteredTool) codemode.Declaration {
	out := t.OutputSchema
	if len(out) == 0 {
		out = codemode.TextOutputSchema
	}
	return codemode.Declaration{Name: t.Name, Description: t.Description, InputSchema: t.Parameters, OutputSchema: out}
}

// renderDeclarations renders (and caches) Pi's tool samples.
func (e *Engine) renderDeclarations(ctx context.Context, decls []codemode.Declaration) (codemode.Rendered, error) {
	raw, _ := json.Marshal(decls)
	sum := sha256.Sum256(raw)
	key := hex.EncodeToString(sum[:])
	if v, ok := e.codemode.renders.Load(key); ok {
		return v.(codemode.Rendered), nil
	}
	eng, err := codemode.Default()
	if err != nil {
		return codemode.Rendered{}, err
	}
	r, err := eng.RenderDeclarations(ctx, decls)
	if err != nil {
		return r, err
	}
	e.codemode.renders.Store(key, r)
	return r, nil
}

// toolExposure is a tool's exposure: MCP tools have their configured
// exposure; every other tool counts as direct.
func (e *Engine) toolExposure(name string) string {
	if e.mcp == nil {
		return gimcp.ExposureDirect
	}
	e.mcp.mu.Lock()
	defer e.mcp.mu.Unlock()
	if server, ok := e.mcp.owners[name]; ok {
		for _, mt := range e.mcp.byServer[server] {
			if mt.Name == name {
				return mt.Exposure
			}
		}
	}
	return gimcp.ExposureDirect
}

func (e *Engine) codemodeNamespace(name string) *codemode.Namespace {
	ns := e.mcpNamespaceOf(name)
	if ns == nil {
		return nil
	}
	return &codemode.Namespace{Name: ns.Name, Description: ns.Description, Instructions: ns.Instructions}
}

// applyCodemodeLoadout ports Pi's prepareCodemodeLoadout for a turn whose
// tools declare codemode: the codemode description lists callable tools
// (mode on: those without direct exposure; only: all), mode on appends the
// codemode declaration to declared callable tools, and mode only leaves out
// declared direct tools.
func (e *Engine) applyCodemodeLoadout(ctx context.Context, convCtx *goai.Context, sessionID string, metadata map[string]any) {
	if e.codemode == nil || convCtx == nil {
		return
	}
	idx := -1
	for i, t := range convCtx.Tools {
		if t.Name == codemodeToolName {
			idx = i
		}
	}
	if idx < 0 {
		return
	}
	mode := e.codemodePresentation(ctx, sessionID)
	callable := e.codemodeCallable(metadata)
	decls := make([]codemode.Declaration, len(callable))
	for i, t := range callable {
		decls[i] = toDeclaration(t)
	}
	rendered, err := e.renderDeclarations(ctx, decls)
	if err != nil {
		log.Printf("codemode: %v", err)
		return
	}
	var listed []codemode.Declaration
	namespaces := map[string]codemode.Namespace{}
	deferred := map[string]bool{}
	callableSet := map[string]bool{}
	for i, t := range callable {
		callableSet[t.Name] = true
		exposure := e.toolExposure(t.Name)
		if mode != "only" && exposure == gimcp.ExposureDirect {
			continue
		}
		listed = append(listed, decls[i])
		if ns := e.codemodeNamespace(t.Name); ns != nil {
			namespaces[t.Name] = *ns
		}
		if exposure == gimcp.ExposureDeferred {
			deferred[t.Name] = true
		}
	}
	budget := codemode.DefaultInlineBudget
	if b := e.runtimeCfg.Codemode.InlineBudget; b != nil && *b >= 0 {
		budget = *b
	}
	convCtx.Tools[idx].Description = codemode.Description(listed, rendered, codemode.DescriptionOptions{Namespaces: namespaces, Deferred: deferred, InlineBudget: &budget,
		Models: true, DocsPath: codemode.ReferencePath})
	declByName := map[string]codemode.Declaration{}
	for _, d := range decls {
		declByName[d.Name] = d
	}
	out := convCtx.Tools[:0]
	for _, t := range convCtx.Tools {
		if callableSet[t.Name] {
			if mode == "only" && e.toolExposure(t.Name) == gimcp.ExposureDirect {
				continue // Pi: requests leave out declarations of direct tools
			}
			if mode == "on" {
				// Pi 1.0: one line on how scripts call it, not the declaration.
				t.Description = codemode.ScriptCallDescription(declByName[t.Name], rendered)
			}
		}
		out = append(out, t)
	}
	convCtx.Tools = out
}

// executeCodemode runs one script (Pi's executeCodemode).
func (e *Engine) executeCodemode(ctx context.Context, rt tools.ToolRuntime, call goai.ToolCall) (string, error) {
	started := time.Now()
	input, _ := call.Arguments["code"].(string)
	code, opts, err := codemode.ParseSource(input)
	if err != nil {
		return "", err
	}
	var metadata map[string]any
	if rt.TurnID != "" {
		if turnRec, err := e.store.GetTurn(ctx, rt.TurnID); err == nil {
			metadata = turnRec.Metadata
		}
	}
	callable := e.codemodeCallable(metadata)
	decls := make([]codemode.Declaration, len(callable))
	for i, t := range callable {
		decls[i] = toDeclaration(t)
	}
	rendered, err := e.renderDeclarations(ctx, decls)
	if err != nil {
		return "", err
	}
	var seq int32
	var seqMu sync.Mutex
	nextID := func() string {
		seqMu.Lock()
		defer seqMu.Unlock()
		seq++
		return fmt.Sprintf("%s/%d", rt.ToolCallID, seq)
	}
	calls := &codemodeCallLog{}
	scriptTools := make([]codemode.Tool, len(callable))
	for i, t := range callable {
		t := t
		scriptTools[i] = codemode.Tool{Name: t.Name, Description: rendered.Samples[t.Name],
			Execute: func(callCtx context.Context, args json.RawMessage) (json.RawMessage, error) {
				id := nextID()
				rec := calls.start(id, t.Name, args)
				value, err := e.runNestedTool(callCtx, rt, id, t, args)
				calls.finish(rec, err, callCtx.Err() != nil)
				return value, err
			}}
	}
	store := e.codemodeStore(ctx, rt.SessionID)
	timeout := codemodeNoTimeout
	if opts.TimeoutMs != nil {
		timeout = time.Duration(*opts.TimeoutMs) * time.Millisecond
	}
	eng, err := codemode.Default()
	if err != nil {
		return "", err
	}
	generatedImages := 0
	globals := append(e.discoveryGlobals(callable, rendered), e.codemodeModelGlobals(rt, calls, &generatedImages)...)
	res := eng.Execute(ctx, code, codemode.Options{Tools: scriptTools, Globals: globals,
		Store: store, Timeout: timeout, MemoryLimitBytes: codemodeMemoryLimit})

	var texts []string
	var images []codemode.OutputItem
	for _, item := range res.Output {
		if item.Type == "image" {
			images = append(images, item)
		} else {
			texts = append(texts, item.Text)
		}
	}
	if res.OK {
		if len(res.StoreWrites.Set) > 0 || len(res.StoreWrites.Delete) > 0 {
			e.applyCodemodeStore(ctx, rt.SessionID, store, res.StoreWrites)
		}
		if res.Value != nil {
			texts = append(texts, valueText(res.Value))
		}
	} else {
		texts = append(texts, "Script error:\n"+formatCodemodeError(res.Error, res.Calls))
	}
	if generatedImages > 0 && len(images) == 0 {
		plural := "s"
		if generatedImages == 1 {
			plural = ""
		}
		texts = append(texts, fmt.Sprintf("Note: models.generateImages() returned %d image%s that the script did not show. Show each image block of result.output with image(block).", generatedImages, plural))
	}
	maxTokens := codemodeDefaultMaxTokens
	if opts.MaxOutputTokens != nil {
		maxTokens = *opts.MaxOutputTokens
	}
	body, fullOutputPath := e.truncateCodemodeOutput(ctx, rt.SessionID, texts, maxTokens)
	if rt.SetDetails != nil {
		details := map[string]any{"calls": calls.snapshot()}
		if fullOutputPath != "" {
			details["fullOutputPath"] = fullOutputPath
		}
		rt.SetDetails(details)
	}
	header := "Script failed"
	if res.OK {
		header = "Script completed"
	}
	header += fmt.Sprintf("\nWall time %.1f seconds\nOutput:\n", time.Since(started).Seconds())
	for _, img := range images {
		if rt.AttachImage != nil {
			if data, err := decodeBase64(img.Data); err == nil {
				rt.AttachImage(img.MimeType, data)
			}
		}
	}
	text := header + body
	if !res.OK {
		return "", errors.New(text)
	}
	return text, nil
}

// runNestedTool runs one script tool call through the tool hooks.
func (e *Engine) runNestedTool(ctx context.Context, rt tools.ToolRuntime, id string, tool tools.RegisteredTool, rawArgs json.RawMessage) (json.RawMessage, error) {
	args := map[string]any{}
	if len(rawArgs) > 0 && string(rawArgs) != "null" {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return nil, fmt.Errorf("tool %s expects an object argument", tool.Name)
		}
	}
	call := goai.ToolCall{Type: "toolCall", ID: id, Name: tool.Name, Arguments: args}
	payload := func(extra map[string]any) map[string]any {
		p := map[string]any{"tool": call.Name, "tool_call_id": call.ID, "arguments": call.Arguments, "parent_tool_call_id": rt.ToolCallID}
		for k, v := range extra {
			p[k] = v
		}
		return p
	}
	for _, hook := range []string{HookToolCall, HookApproveTool} {
		resp, err := e.emitHook(ctx, HookRequest{Name: hook, SessionID: rt.SessionID, TurnID: rt.TurnID, ToolCall: &call, Payload: payload(nil)})
		if err != nil {
			log.Printf("hook %s error: %v", hook, err)
			continue
		}
		if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s aborted by hook", call.Name)); abortErr != nil {
			return nil, abortErr
		}
		if resp.Block {
			reason := resp.Reason
			if reason == "" {
				reason = "tool call blocked"
			}
			return nil, fmt.Errorf("blocked by hook: %s", reason)
		}
		if resp.ToolCall != nil && resp.ToolCall.Name == call.Name {
			call.Arguments = resp.ToolCall.Arguments
		}
		if hook == HookToolCall {
			if injected, ok := directToolResultFromHook(resp); ok {
				return json.RawMessage(mustJSON(injected)), nil
			}
		}
	}
	e.PublishRuntimeToolEvent("tool_started", rt.SessionID, rt.TurnID, "", call.Name, call.ID, 0, nil, payload(map[string]any{"phase": "tool"}))
	nestedRT := tools.ToolRuntime{Store: rt.Store, SessionID: rt.SessionID, TurnID: rt.TurnID, WorkspaceRoot: rt.WorkspaceRoot, ToolCallID: call.ID}
	executionStart := time.Now()
	var value json.RawMessage
	var text string
	var toolErr error
	if tool.StructuredExecutor != nil {
		value, _, toolErr = tool.StructuredExecutor(ctx, nestedRT, call)
	} else {
		text, toolErr = tool.Executor(ctx, nestedRT, call)
	}
	durationMS := time.Since(executionStart).Milliseconds()
	if toolErr != nil {
		e.PublishRuntimeToolEvent("tool_failed", rt.SessionID, rt.TurnID, "", call.Name, call.ID, 0, toolErr, payload(map[string]any{"phase": "tool", "duration_ms": durationMS}))
		msg := toolErr.Error()
		if msg == "" {
			msg = fmt.Sprintf("Tool %q failed", call.Name)
		}
		return nil, errors.New(msg)
	}
	if tool.StructuredExecutor == nil {
		resp, err := e.emitHook(ctx, HookRequest{Name: HookToolResult, SessionID: rt.SessionID, TurnID: rt.TurnID, ToolCall: &call, ToolResult: text, Payload: payload(map[string]any{"is_error": false})})
		if err == nil {
			if abortErr := hookAbortFromResponse(resp, fmt.Sprintf("tool %s result aborted by hook", call.Name)); abortErr != nil {
				return nil, abortErr
			}
			if resp.ToolResult != nil {
				text = *resp.ToolResult
			}
		}
		value = json.RawMessage(mustJSON(text))
	}
	e.PublishRuntimeToolEvent("tool_finished", rt.SessionID, rt.TurnID, "", call.Name, call.ID, 0, nil, payload(map[string]any{"phase": "tool", "output_length": len(value), "duration_ms": durationMS}))
	return value, nil
}

// discoveryGlobals ports Pi's searchTools, describeTool and describeNamespace.
func (e *Engine) discoveryGlobals(callable []tools.RegisteredTool, rendered codemode.Rendered) []codemode.Global {
	entry := func(name string) map[string]any {
		return map[string]any{"name": codemode.Identifier(name), "description": rendered.Samples[name]}
	}
	args := func(raw json.RawMessage) []any {
		var a []any
		_ = json.Unmarshal(raw, &a)
		return a
	}
	return []codemode.Global{
		{Name: "searchTools", Spread: true, Execute: func(_ context.Context, raw json.RawMessage) (json.RawMessage, error) {
			a := args(raw)
			query, ok := argAt(a, 0).(string)
			if !ok {
				return nil, errors.New("searchTools() expects a query string")
			}
			limit, namespace := tools.DefaultToolSearchLimit, ""
			if o, ok := argAt(a, 1).(map[string]any); ok {
				if v, has := o["limit"]; has && v != nil {
					f, ok := v.(float64)
					if !ok || f != math.Trunc(f) || f <= 0 {
						return nil, errors.New("searchTools() limit must be a positive integer")
					}
					limit = int(f)
				}
				if v, has := o["namespace"]; has && v != nil {
					s, ok := v.(string)
					if !ok {
						return nil, errors.New("searchTools() namespace must be a string")
					}
					namespace = s
				}
			}
			var docs []tools.SearchDocument
			for _, t := range callable {
				ns := e.mcpNamespaceOf(t.Name)
				if namespace != "" && (ns == nil || !isNamespaceName(ns.Name, namespace)) {
					continue
				}
				var params any
				_ = json.Unmarshal(t.Parameters, &params)
				docs = append(docs, tools.NewSearchDocument(t.Name, t.Description, params, ns))
			}
			out := []map[string]any{}
			for _, m := range tools.RankBM25(query, docs, limit) {
				out = append(out, entry(m.Name))
			}
			b, err := json.Marshal(out)
			return b, err
		}},
		{Name: "describeTool", Spread: true, Execute: func(_ context.Context, raw json.RawMessage) (json.RawMessage, error) {
			name, ok := argAt(args(raw), 0).(string)
			if !ok {
				return nil, errors.New("describeTool() expects a tool name")
			}
			for _, t := range callable {
				if t.Name == name || codemode.Identifier(t.Name) == name {
					return json.RawMessage(mustJSON(rendered.Samples[t.Name])), nil
				}
			}
			return nil, nil
		}},
		{Name: "describeNamespace", Spread: true, Execute: func(_ context.Context, raw json.RawMessage) (json.RawMessage, error) {
			name, ok := argAt(args(raw), 0).(string)
			if !ok {
				return nil, errors.New("describeNamespace() expects a namespace name")
			}
			var found *tools.SearchNamespace
			names := []string{}
			for _, t := range callable {
				ns := e.mcpNamespaceOf(t.Name)
				if ns == nil || !isNamespaceName(ns.Name, name) {
					continue
				}
				if found == nil {
					found = ns
				}
				names = append(names, codemode.Identifier(t.Name))
			}
			if found == nil {
				return nil, nil
			}
			out := map[string]any{"name": found.Name, "tools": names}
			if found.Description != "" {
				out["description"] = found.Description
			}
			if found.Instructions != "" {
				out["instructions"] = found.Instructions
			}
			b, err := json.Marshal(out)
			return b, err
		}},
	}
}

func argAt(a []any, i int) any {
	if i < len(a) {
		return a[i]
	}
	return nil
}

// isNamespaceName ports Pi's: the name, its identifier form, or the part
// after its last "__" in either form.
func isNamespaceName(namespace, query string) bool {
	id, queryID := codemode.Identifier(namespace), codemode.Identifier(query)
	suffix := func(name string) (string, bool) {
		if i := strings.LastIndex(name, "__"); i >= 0 {
			return name[i+2:], true
		}
		return "", false
	}
	if namespace == query || id == queryID {
		return true
	}
	if s, ok := suffix(namespace); ok && s == query {
		return true
	}
	if s, ok := suffix(id); ok && s == queryID {
		return true
	}
	return false
}

func (e *Engine) codemodeStore(ctx context.Context, sessionID string) map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	if sessionID == "" {
		return out
	}
	sess, err := e.store.GetSession(ctx, sessionID)
	if err != nil {
		return out
	}
	if m, ok := sess.State[codemodeStoreStateKey].(map[string]any); ok {
		for k, v := range m {
			if b, err := json.Marshal(v); err == nil {
				out[k] = b
			}
		}
	}
	return out
}

// applyCodemodeStore persists a successful script's store writes in the
// session state (Pi keeps them as session entries; forks copy the state).
func (e *Engine) applyCodemodeStore(ctx context.Context, sessionID string, store map[string]json.RawMessage, writes codemode.StoreWrites) {
	if sessionID == "" {
		return
	}
	next := map[string]any{}
	for k, v := range store {
		next[k] = v
	}
	for _, k := range writes.Delete {
		delete(next, k)
	}
	for k, v := range writes.Set {
		next[k] = v
	}
	if err := e.store.TouchSessionState(ctx, sessionID, map[string]any{codemodeStoreStateKey: next}); err != nil {
		log.Printf("codemode: store: %v", err)
	}
}

// valueText is Pi's: strings as is, other values as compact JSON.
func valueText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}

func formatCodemodeError(err *codemode.Error, calls []codemode.Call) string {
	var head string
	switch err.Kind {
	case codemode.ErrorScript:
		head = err.Stack
		if head == "" {
			name := err.Name
			if name == "" {
				name = "Error"
			}
			head = name + ": " + err.Message
		}
	case codemode.ErrorTimeout:
		head = "Script timed out: " + err.Message
	case codemode.ErrorAborted:
		head = "Script aborted: " + err.Message
	default:
		head = "Script sandbox failed: " + err.Message
	}
	summary := "No tool calls were made."
	if len(calls) > 0 {
		parts := make([]string, len(calls))
		for i, c := range calls {
			parts[i] = fmt.Sprintf("%s (%s)", c.Name, c.Status)
		}
		summary = "Tool calls made before the failure (they are not undone): " + strings.Join(parts, ", ")
	}
	return head + "\n\n" + summary
}

// truncateCodemodeOutput ports Pi's truncateOutput: past the token budget
// (characters / 4) the text keeps its start and end and the full text is
// saved for reading (vfs://codemode-output; Pi uses a temp file).
func (e *Engine) truncateCodemodeOutput(ctx context.Context, sessionID string, texts []string, maxTokens int) (string, string) {
	combined := strings.Join(texts, "\n")
	budget := maxTokens * codemodeCharsPerToken
	runes := []rune(combined)
	if len(texts) == 0 || len(runes) <= budget {
		return combined, ""
	}
	head := budget / 2
	tail := budget - head
	removed := len(runes) - head - tail
	text := fmt.Sprintf("Warning: truncated output (original token count: %d)\nTotal output lines: %d\n\n%s…%d tokens truncated…%s",
		int(math.Ceil(float64(len(runes))/codemodeCharsPerToken)), strings.Count(combined, "\n")+1,
		string(runes[:head]), int(math.Ceil(float64(removed)/codemodeCharsPerToken)), string(runes[len(runes)-tail:]))
	var id [8]byte
	_, _ = rand.Read(id[:])
	session := sessionID
	if session == "" {
		session = "shared"
	}
	path := session + "/" + hex.EncodeToString(id[:]) + ".txt"
	if _, err := e.store.SaveVFSFile(ctx, codemodeOutputNamespace, path, "text/plain; charset=utf-8", []byte(combined), map[string]any{"source": "codemode"}); err != nil {
		return text + "\n\n[Could not save the full output: " + err.Error() + "]", ""
	}
	full := "vfs://" + codemodeOutputNamespace + "/" + path
	return text + "\n\n[Full output: " + full + " (read with offset/limit)]", full
}

// codemodeCallLog records a script's nested calls for the renderer (Pi's
// call rows): name, argument preview, status, duration and error preview.
type codemodeCallLog struct {
	mu    sync.Mutex
	calls []*codemodeCallRecord
}

type codemodeCallRecord struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Args       string  `json:"args"`
	Status     string  `json:"status"` // running, ok, error, cancelled
	DurationMs float64 `json:"durationMs,omitempty"`
	Error      string  `json:"error,omitempty"`
	Cost       float64 `json:"cost,omitempty"` // model calls (USD)
	started    time.Time
}

// truncatePreview is Pi's truncateText: "..." replaces the end.
func truncatePreview(text string, maxChars int) string {
	r := []rune(text)
	if len(r) <= maxChars {
		return text
	}
	return string(r[:maxChars-3]) + "..."
}

func (l *codemodeCallLog) start(id, name string, args json.RawMessage) *codemodeCallRecord {
	preview := ""
	if len(args) > 0 && string(args) != "null" {
		preview = truncatePreview(string(args), 200)
	}
	rec := &codemodeCallRecord{ID: id, Name: name, Args: preview, Status: "running", started: time.Now()}
	l.mu.Lock()
	l.calls = append(l.calls, rec)
	l.mu.Unlock()
	return rec
}

func (l *codemodeCallLog) finish(rec *codemodeCallRecord, err error, cancelled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	rec.DurationMs = float64(time.Since(rec.started).Microseconds()) / 1000
	switch {
	case err == nil:
		rec.Status = "ok"
	case cancelled:
		rec.Status = "cancelled"
	default:
		rec.Status = "error"
		rec.Error = truncatePreview(err.Error(), codemodeErrorPreviewChars)
	}
}

// snapshot returns the calls; those still running were cut off by the
// script ending, a timeout or an abort (Pi marks them cancelled).
func (l *codemodeCallLog) snapshot() []codemodeCallRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]codemodeCallRecord, len(l.calls))
	for i, c := range l.calls {
		out[i] = *c
		if out[i].Status == "running" {
			out[i].Status = "cancelled"
		}
	}
	return out
}

func decodeBase64(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }
