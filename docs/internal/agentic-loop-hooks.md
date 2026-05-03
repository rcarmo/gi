# Pi agentic loop — extension points assessment

**Date:** 2026-05-02  
**Scope:** Non-UX hooks only (tool registration, event lifecycle, context injection, provider/model control, session state, scripting exposure targets)

---

## 1. Pi's agentic loop — canonical flow

```
Extension factory (async ok)
  │
  ▼
session_start { reason: "startup" | "switch" | "fork" | "tree_restore" }
resources_discover                     ← extension can inject resource list
  │
  ▼ (user prompt)
input                                  ← can intercept, transform, suppress
  │
  ├── (extension command) → skip agent
  ├── (skill/template expansion)
  │
  ▼
before_agent_start                     ← inject messages, mutate system prompt
agent_start
  │
  ▼
  ┌──────── turn loop (repeats while LLM calls tools) ──────────────┐
  │ turn_start                                                       │
  │ context                            ← can rewrite message list    │
  │ before_provider_request            ← can replace full payload    │
  │                                                                  │
  │   LLM responds                                                   │
  │                                                                  │
  │ after_provider_response            ← status + headers            │
  │ message_start / message_update / message_end                     │
  │                                                                  │
  │   per tool call:                                                 │
  │     tool_execution_start                                         │
  │     tool_call                      ← can block, redirect         │
  │     tool_execution_update                                        │
  │     tool_result                    ← can modify output           │
  │     tool_execution_end                                           │
  │                                                                  │
  │ turn_end                                                         │
  └──────────────────────────────────────────────────────────────────┘
  │
  ▼
agent_end

session_before_compact / session_compact   ← custom compaction handler
session_before_switch / session_before_fork / session_before_tree
session_shutdown
```

---

## 2. Pi's extension points — non-UX catalogue

### 2.1 Tool registration

| Method | What it does |
|--------|-------------|
| `pi.registerTool(def)` | Register a tool callable by the LLM. Has `name`, `description`, `parameters` (TypeBox), `execute`, optional `prepareArguments` for migration |
| `pi.getActiveTools()` / `pi.getAllTools()` | Inspect current tool set with `sourceInfo` provenance |
| `pi.setActiveTools(names)` | Swap the live tool set without restart |
| Override built-in tools | Register tool with same name as a builtin (`read`, `bash`, `edit`, `write`, `grep`, `find`, `ls`) — pi keeps both, uses extension's |
| `createReadTool(cwd, { operations })` etc. | Factory for builtins with pluggable ops (SSH, containers, sandbox) |
| `createBashTool(cwd, { spawnHook })` | Bash tool with command/cwd/env rewrite before spawn |
| `withFileMutationQueue(path, fn)` | Participate in per-file write queue (avoid race with concurrent edits) |

**Tool execute contract:**
```typescript
async execute(toolCallId, params, signal, onUpdate, ctx): Promise<{
  content: ContentBlock[],   // sent to LLM
  details?: any,             // for rendering + state reconstruction
  terminate?: true,          // hint: skip follow-up LLM call after this batch
}>
// throw to signal error (sets isError:true on result)
// onUpdate({content, details}) for streaming progress
```

---

### 2.2 Event hooks (full set)

#### Process / session lifecycle

| Event | When | Can block/mutate? | Payload highlights |
|-------|------|-------------------|--------------------|
| `session_start` | Startup, switch, fork, tree restore | No | `reason`, `sessionManager`, `modelRegistry` |
| `session_shutdown` | Session ends | No | — |
| `session_before_switch` | Before switching sessions | **Yes — return `{cancel:true}`** | `fromId`, `toId` |
| `session_before_fork` | Before forking | **Yes — return `{cancel:true}`** | `fromId` |
| `session_before_tree` | Before tree restore | **Yes — return `{cancel:true}`** | `fromId`, `toId` |
| `session_before_compact` | Before compaction | **Yes — custom summary** | `getMessages()`, return `{summary}` |
| `session_compact` | After compaction | No | `summary` |
| `session_tree` | After tree restore | No | `fromId`, `toId` |
| `resources_discover` | On startup and `/reload` | **Yes — inject resources** | Return `{resources:[...]}` |

#### Agent / turn lifecycle

| Event | When | Can block/mutate? | Payload highlights |
|-------|------|-------------------|--------------------|
| `input` | User submits prompt | **Yes — transform/suppress** | `content`, can return `{handled:true}` or modified content |
| `before_agent_start` | Before first LLM call | **Yes — inject message, mutate system prompt** | Can return `{message, systemPromptOptions}` |
| `agent_start` | Agent loop begins | No | `turnId`, `model` |
| `agent_end` | Agent loop ends | No | `turnId`, `reason` |
| `turn_start` | Each iteration starts | No | `iteration`, `model` |
| `turn_end` | Each iteration ends | No | `iteration` |
| `context` | Before each LLM call | **Yes — rewrite message list** | `messages[]` — can return mutated list |
| `before_provider_request` | Before HTTP to LLM | **Yes — replace payload** | `payload`, `provider`, `model`; return `{payload}` |
| `after_provider_response` | After HTTP response headers | No (observe only) | `statusCode`, `headers` |
| `message_start` / `message_update` / `message_end` | Streaming LLM text | No (observe) | `content`, `delta` |
| `model_select` | Model changes | No (observe) | `model`, `provider` |

#### Tool lifecycle

| Event | When | Can block/mutate? | Payload highlights |
|-------|------|-------------------|--------------------|
| `tool_execution_start` | Before a tool batch | No | `toolCalls[]` |
| `tool_call` | Each tool call before execute | **Yes — block, redirect** | `toolName`, `input`; return `{block:true, reason}` or `{redirect: newInput}` |
| `tool_execution_update` | During streaming tool | No | `toolCallId`, `partial` |
| `tool_result` | After tool executes | **Yes — modify output** | `toolCallId`, `toolName`, `result`; return `{result: modified}` |
| `tool_execution_end` | After a tool batch | No | `toolCalls[]`, `results[]` |
| `user_bash` | `bash` tool calls | **Yes — override execution** | Replaces the shell backend entirely |

---

### 2.3 System prompt / context injection

- `before_agent_start` → return `{ message: string | ContentBlock[] }` to prepend a synthetic user message
- `before_agent_start` → return `{ systemPromptOptions: BuildSystemPromptOptions }` to append/replace tool guidelines, sections
- `context` → rewrite/filter/append to `messages[]` on every turn
- `resources_discover` → inject a resource list that gets embedded in the system prompt's resources section

---

### 2.4 Session state persistence

- `pi.appendEntry(customType, data)` — persist opaque state into the session journal (survives restart, branches correctly)
- `ctx.sessionManager.getEntries()` / `ctx.sessionManager.getBranch()` — read back entries for state reconstruction on `session_start`
- `pi.setSessionName(name)` / `pi.getSessionName()` — session metadata
- `pi.setLabel(entryId, label)` — bookmark entries for tree navigation

---

### 2.5 Provider / model control

- `pi.registerProvider(name, config)` — register or override a provider with custom baseUrl, apiKey, headers, models, OAuth
- `pi.unregisterProvider(name)` — remove a provider (builtins restored)
- `pi.setModel(model)` — switch model programmatically
- `pi.getThinkingLevel()` / `pi.setThinkingLevel(level)` — `"off" | "minimal" | "low" | "medium" | "high" | "xhigh"`

---

### 2.6 Commands, shortcuts, flags

- `pi.registerCommand(name, { description, handler, getArgumentCompletions })` — slash command visible in the LLM prompt pipeline
- `pi.registerShortcut(shortcut, { description, handler })` — keybinding (TUI only, irrelevant for scripting targets)
- `pi.registerFlag(name, { type, default })` — CLI flag; readable via `pi.getFlag(name)`

---

### 2.7 Inter-extension / process

- `pi.events` — shared in-process event bus (`on`, `emit`) for extension-to-extension communication
- `pi.exec(cmd, args, opts)` — spawn a process; returns `{stdout, stderr, code, killed}`
- `pi.sendMessage(msg)` — inject a non-LLM message into the session timeline
- `pi.sendUserMessage(content, opts)` — inject a real user message (triggers a turn)

---

## 3. Gi's current agentic loop — state of play

Gi's loop lives in `internal/turn/agent_loop.go`. It is simpler and has **zero hook surface** today.

### 3.1 Current loop structure

```
SubmitPrompt(input)
  │
  └─► runAgentLoop(ctx, store, turnID, sessionID, model, agentID)
        │
        ├─ load history from DB
        ├─ build goai.Context{SystemPrompt, Tools, Messages}
        │
        └── for iter := 0..maxIter:
              │
              ├─ inference.StreamWithTools(ctx, model, convCtx, streamCb)
              │     └── broadcasts: agent_draft_delta, agent_thought_delta, tool_call_start, error
              │
              ├─ if no tool calls → persist + broadcastPost + done
              │
              └─ for each call ∈ toolCalls:
                    ├─ executeTool(call)         ← read / write / shell / tools (meta)
                    └─ goai.AppendToolResult(convCtx, ...)
```

**Broadcast events already emitted to SSE clients:**

| SSE type | When |
|----------|------|
| `agent_status` | Start, each iteration, each tool call, idle |
| `agent_draft_delta` | Streaming LLM text |
| `agent_thought_delta` | Streaming reasoning/thinking |
| `tool_call_start` | Before tool execution |
| `new_post` | Final assistant message |
| `agent_response` | Final assistant message (secondary) |
| `error` | Inference error |

**Store events written:**

| Store event | When |
|-------------|------|
| `inference.started` | Each iteration |
| `inference.failed` | Inference error |
| `inference.finished` | Final iteration (with usage) |
| `tool.started` | Before each tool |
| `tool.failed` | Tool error |
| `tool.finished` | Tool success |
| `turn.finished` | Turn end (with status) |

---

### 3.2 Hook gaps — what gi has vs. what pi exposes

| Pi hook | Gi equivalent | Gap |
|---------|--------------|-----|
| `session_start` | — | ❌ No session lifecycle callbacks |
| `session_shutdown` | — | ❌ |
| `before_agent_start` | — | ❌ No system prompt injection point |
| `input` | — | ❌ No input transform/intercept |
| `context` | — | ❌ No per-turn message list rewrite |
| `before_provider_request` | — | ❌ No payload inspection/replacement |
| `after_provider_response` | — | ❌ No response header observation |
| `turn_start` / `turn_end` | Store events | ✅ Emitted (store), ❌ Not callable |
| `tool_call` (gate) | — | ❌ No pre-execution gate |
| `tool_result` (mutate) | — | ❌ No result modification |
| `tool_execution_start/end` | Store events | ✅ Emitted (store), ❌ Not callable |
| `session_before_compact` | — | ❌ No compaction hook (no compaction yet) |
| `resources_discover` | — | ❌ No resource injection |
| `registerTool` | `toolDefs()` hardcoded | ❌ Tools are hardcoded, not registerable |
| `setActiveTools` | — | ❌ No dynamic tool set |
| `registerProvider` | config only | ❌ No runtime provider registration |
| `setModel` | config only | ❌ No runtime model switch |
| `setThinkingLevel` | — | ❌ |
| `pi.events` | — | ❌ No internal event bus |
| `pi.exec` | shell tool | ✅ Approximated via shell tool |
| `sendMessage` | broadcast | ✅ Partial (SSE only, not session-persisted) |

---

## 4. Target scripting exposure — what to expose in gi

For gi's scripting bridge (Goja/Joker), the hooks that make sense to expose are those that:
- **Are pure data** (no TUI, no process management)
- **Have clear Go ↔ script call boundaries**
- **Don't require async callback chains** (or can be wrapped in synchronous host calls)

### Tier 1 — Expose first (low complexity, high value)

| Hook / API | Scripting form | Notes |
|------------|---------------|-------|
| `registerTool` | `gi.registerTool(name, description, paramsSchema, fn)` | fn called sync from Go host; returns `{content, error}` |
| `before_agent_start` | `gi.on("before_agent_start", fn)` | fn receives `{sessionID, model}`, returns `{message?, systemPrompt?}` |
| `tool_call` gate | `gi.on("tool_call", fn)` | fn receives `{name, args}`, returns `{block?, reason?, args?}` |
| `tool_result` mutate | `gi.on("tool_result", fn)` | fn receives `{name, output}`, returns `{output?}` |
| `turn_start` / `turn_end` | `gi.on("turn_start/end", fn)` | observe-only, no return value |
| `session_start/shutdown` | `gi.on("session_start/shutdown", fn)` | observe-only |
| `context` inject | `gi.on("context", fn)` | fn receives messages array, returns modified array |

### Tier 2 — Expose after Tier 1 is stable

| Hook / API | Scripting form | Notes |
|------------|---------------|-------|
| `input` transform | `gi.on("input", fn)` | fn returns `{content?, handled?}` |
| `before_provider_request` | `gi.on("before_request", fn)` | fn receives raw payload, returns `{payload?}` |
| `after_provider_response` | `gi.on("after_response", fn)` | observe only, headers + status |
| `message_start/update/end` | `gi.on("message_*", fn)` | streaming text observation |
| `setActiveTools(names)` | `gi.setActiveTools([...])` | runtime tool set swap |
| `setModel(id)` | `gi.setModel("provider/model")` | runtime model switch |
| `appendEntry(type, data)` | `gi.appendEntry(type, data)` | persist script state |
| `getEntries()` | `gi.getEntries()` | restore script state on start |

### Tier 3 — Future / advanced

| Hook / API | Scripting form | Notes |
|------------|---------------|-------|
| `resources_discover` | `gi.on("resources_discover", fn)` | inject resource list |
| `session_before_compact` | `gi.on("before_compact", fn)` | custom summary fn |
| `registerProvider` | `gi.registerProvider(name, cfg)` | dynamic provider addition |
| `pi.events` bus | `gi.events.on/emit` | inter-script messaging |
| `session_before_fork` etc. | `gi.on("before_fork", fn)` | session guard hooks |

---

## 5. Design constraints for gi scripting bridge

### Synchronous host call model

Both Goja (JS) and Joker (Clojure) run on Go goroutines. The host→script call model is **synchronous**: the Go loop blocks on the script call and resumes when it returns. This means:

- Hook `fn` must return **within the script's own execution**, not schedule async callbacks
- Long-running script hooks will stall the agent loop — document the responsibility
- For Joker, async is irrelevant (no event loop); for Goja, Promises must be `.then()` unwrapped in the host before returning to Go

### Tool execute boundary

Script-registered tools look like any other `executeTool` case. The Go host:
1. Receives `goai.ToolCall{Name: "my_script_tool", Arguments: {...}}`
2. Looks up script-registered handlers by name
3. Calls script fn with `(args map)`, receives `(result string, error string)`
4. Returns result to the loop like any built-in

Schema is passed to `toolDefs()` as raw JSON — script registers `name + description + jsonSchemaString`.

### Hook fan-out order

When multiple scripts register the same hook:
- Observe-only hooks (`turn_start`, `session_start`, etc.): all called in registration order
- Gate hooks (`tool_call`): first non-nil blocking response wins, rest skipped
- Mutate hooks (`tool_result`, `context`, `before_agent_start`): chained — each fn receives the output of the previous

### System prompt injection

`before_agent_start` is the correct point to inject per-turn context (date, workspace state, tool guidance). Scripts returning a `message` string will have it prepended as a synthetic user message before the first LLM call in that turn.

---

## 6. What gi needs to build first (prerequisite to scripting hooks)

Before any scripting hook surface can be wired:

1. **Hook registry** in `internal/turn/engine.go` — a typed map of `hookName → []HandlerFn`
2. **Hook call sites** in `runAgentLoop` — at each relevant point, fan-out to registered handlers
3. **Tool registry** — replace `toolDefs()` hardcoded slice with a runtime-registerable `ToolRegistry`
4. **Script→Go registration API** — in `internal/scripting/bridge.go`: `RegisterHook`, `RegisterTool`, `CallHook`
5. **Session context pass-through** — `sessionID`, `turnID`, `model`, current message list available to hook callees

---

## 7. Implemented slice (2026-05-02)

The initial implementation now exists in the turn engine:

- `internal/turn/hooks.go` — typed hook registry and pi-compatible non-UX hook constants.
- `internal/turn/tool_registry.go` — runtime tool registry, active-tool set, metadata-enriched entries (`kind`, `weight`, `activation`, `source`, `active`), staged discovery, activation/reset, and registry-backed `tools` meta-tool.
- `internal/turn/default_tools.go` — built-in tools registered through the runtime registry (`tools`, `skills`, `read`, `write`, `script`, `shell`).
- `internal/skills/discovery.go` — Pi-style workspace discovery for `.gi/skills/*/SKILL.md`, `.pi/skills/*/SKILL.md`, `.gi/tools/*.json`, and `.pi/tools/*.json`.
- `internal/turn/skills_tools.go` — `skills` meta-tool and auto-registration of manifest-declared script tools into the same runtime registry as built-ins and script-registered tools.
- `internal/turn/agent_loop.go` — hook call sites wired through the agent loop:
  - `before_agent_start`
  - `agent_start` / `agent_end`
  - `turn_start` / `turn_end`
  - `context`
  - `before_provider_request` / `after_provider_response` (metadata-level until provider payload interception lands in `go-ai`)
  - `message_update` / `message_end`
  - `tool_execution_start` / `tool_execution_end`
  - `tool_call` gate / rewrite
  - `tool_result` mutation
- `internal/scripting/bridge.go` and adapters expose script-facing registration/control methods:
  - `gi.registerTool(spec)`
  - `gi.registerEventHook(spec)` / `gi.on(spec)` in JS
  - `gi.setActiveTools(names)` / `gi.getActiveTools()`
  - `gi.setModel(model)`
  - `gi.appendEntry(type, data)` / `gi.getEntries(type)`
- `internal/tools/script.go` connects the script tool to the host engine through callbacks so scripts can register new tools and hooks at runtime.
- `config.Load` builds a Pi-like gi runtime system prompt: operating model, built-in tools, skill-loading guidance, shared path policy, runtime hook/connectivity notes, workspace `AGENTS.md`, and a compact discovered-capabilities section.

### JS examples

```js
gi.registerTool({
  name: "hello_script",
  description: "Say hello from a script-registered tool",
  parameters: { type: "object", properties: { name: { type: "string" } } },
  engine: "js",
  script: `"hello " + (gi.toolArgs.name || "world")`,
});

gi.on({
  name: "tool_call",
  source: "example",
  engine: "js",
  script: `
    if (gi.hook.tool_call && gi.hook.tool_call.name === "shell") {
      JSON.stringify({ block: true, reason: "shell disabled by script hook" })
    }
  `,
});
```

### Joker examples

```clojure
(gi-register-tool
  {:name "hello_joker"
   :description "Say hello from Joker"
   :parameters {:type "object" :properties {}}
   :engine "joker"
   :script "\"hello from joker\""})

(gi-set-active-tools ["tools" "read" "hello_joker"])
```

## Summary

Pi has **28 named event hooks**, a fully runtime-registerable tool system, dynamic provider/model control, and a well-defined state persistence model. Gi now has the same core non-UX shape at the engine level: a hook registry, live tool registry, active-tool controls, script-facing registration APIs, and agent-loop call sites. The remaining gap is deep provider payload replacement: `before_provider_request` is wired as an engine hook today, but true raw payload replacement requires lowering the hook into `go-ai` provider calls.
