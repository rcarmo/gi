# Codemode engine (`internal/codemode`)

Status: the engine is in place (#25, phase 4). The model-facing `codemode`
tool, its declarations and its toggles are phase 5
([plan](../implementation/plans/mcp-codemode-plan.md)).

## What it runs

Model-written **JavaScript**, with Pi's codemode semantics.

- The script is the body of an async function, so `return` and top-level
  `await` work.
- Its only capabilities are what the host injects:
  - `tools.<name>(args)`, under the tool's name and its identifier form
    (`my-tool` is also `tools.my_tool`);
  - `ALL_TOOLS`;
  - `text()`, `image()` (base64 `data:` URLs, `{image_url}` or MCP image
    content; remote URLs are rejected) and `console.*`;
  - `exit()`, which ends successfully at once and keeps the output;
  - `store(key, value)` / `load(key)`, with values up to 256 Ki and 1 Mi
    characters in total;
  - configured globals, including `namespace.fn`.
- Scripts get no timers, `fetch`, modules or `WebAssembly`.

## How

- **Runtime:** QuickJS-ng compiled to WASI (`quickjs-wasi` 3.6.2, MIT) runs
  under wazero.
- **Guest side:** Pi's codemode prelude (`@earendil-works/pi-codemode`
  1.0.1, MIT), vendored verbatim, so script behaviour matches Pi by
  construction. `scripts/vendor-codemode.mjs` refreshes both from the
  installed Pi packages (`internal/codemode/vendor`, with licences and
  `VERSIONS`).
- **Host side:** `engine.go` mirrors Pi's worker and host.
  - **Isolation:** each execution gets a fresh module instance (a fresh VM).
  - **Bridge:** it implements `bridge(kind, a, b, c)` for `call`, `global`,
    `output` and `done`.
  - **Calls:** tool and global calls run concurrently in goroutines. Results
    are settled on the VM's goroutine as JSON (or as the error message), and
    each call is recorded with its status and duration.
  - **Jobs and stalls:** pending jobs are drained after every step, and the
    prelude's `stalled()` fails a script that waits on a promise nothing can
    settle.
- **Limits:**
  - default timeout 300 s, enforced through `host_interrupt` and the context;
  - caller cancellation is reported as `aborted`;
  - QuickJS stack 512 KiB, so deep recursion throws a catchable `RangeError`;
  - linear memory capped at 256 MiB, plus an optional QuickJS heap limit;
  - script output: past 16 Mi characters or 100000 `text()`, `image()` and
    `console` calls the prelude fails the script (pi-codemode 1.0.1); `done`
    ends the script at once, as Pi's host does, so catching the error does not
    keep it running;
  - WASI gets no filesystem, environment or arguments, and its stdout and
    stderr are discarded.
- **Errors** use Pi's kinds: `script` (with name, message and stack),
  `timeout`, `aborted` and `sandbox`.
- **Cost on this ARM64 laptop:** compiling QuickJS takes about 0.4 s, once per
  process (`codemode.Default()`); each execution takes about 10 ms, including
  VM start and prelude load.

## API

```go
e, _ := codemode.Default()
res := e.Execute(ctx, code, codemode.Options{Tools: []codemode.Tool{...}, Globals: ..., Store: ..., Timeout: ..., MemoryLimitBytes: ...})
// res.OK, res.Value (JSON), res.Error{Kind,Name,Message,Stack}, res.Output, res.Calls, res.StoreWrites
```

## The `codemode` tool (`internal/turn/codemode_tool.go`)

The tool the model calls is a port of Pi's codemode extension. Its input is
`{ "code": "<JavaScript>" }`. The code may start with a
`// @options: {"max_output_tokens": N, "timeout_ms": N}` line, which
`codemode.ParseSource` reads with Pi's rules and messages.

- **Description** (Pi 1.0's leaner form):
  - **Generation:** `codemode.Description` ports `createCodemodeDescription`. It starts from Pi's base description (intro plus one line per global, without the `models` API), which `scripts/vendor-codemode.mjs` takes from Pi's own `createCodemodeDescription([])` into `vendor/codemode-texts.json`, together with the `code` parameter text and Pi's system-prompt snippet and guideline.
  - **Tool samples:** rendered by Pi's own `declarations.js`, run in QuickJS (`Engine.RenderDeclarations`). Results are cached by a hash of the declarations.
  - **Listed tools:** in mode `on`, tools without direct exposure (MCP codemode and deferred tools); in mode `only`, every callable tool.
  - **Inline budget:** sections are chosen per namespace, cheapest first, until `codemode.inlineBudget` is spent (default 3000 tokens). Deferred-exposure tools are never listed.
  - **Golden values:** `scripts/golden-codemode-description.mjs` regenerates `internal/codemode/testdata/pi-description.json` from Pi.
  - **Prompt contribution:** while codemode is declared, its snippet and guideline (Pi's text, `vendor/codemode-texts.json`) are in the system prompt's tools and rules sections ([system-prompt.md](system-prompt.md)).
- **Loadout (each request):**
  - **Mode `on`:** declared callable tools get one more line saying how scripts call them and what the call resolves to (`codemode.ScriptCallDescription`, a port of Pi's `describeScriptCall`), for example ``Codemode: `tools.bash(args)` resolves to `{ output, exit_code }`.``
  - **Mode `only`:** declarations of direct tools are left out. `codemode` and `tool_search` stay.
- **Callable tools:**
  - **Included:** the turn's active tools, plus every deferred registry entry (MCP codemode/deferred tools).
  - **Excluded:** `ModelOnly` tools (`codemode`, `tool_search`).
- **What scripts receive:**
  - **MCP tools:** their `CallToolResult` (without `_meta`), via `RegisteredTool.StructuredExecutor`, with `OutputSchema` from `codemode.MCPResultSchema`.
  - **Other tools:** their text.
- **Nested calls:**
  - **Hooks:** they run `tool_call`, `approve_tool` and `tool_result` (text tools). Payloads carry `parent_tool_call_id` and IDs look like `<parent>/<n>`.
  - **Events:** `tool_started`/`tool_finished`/`tool_failed` runtime events are published.
  - **Results:** they are not added to the transcript. Only the script's output reaches the model, as in Pi.
  - **Errors:** a blocked or failed call rejects with an Error.
- **Globals:**
  - **Missing members:** reading a `tools` member that does not exist throws an error naming close matches (Pi 1.0's prelude). Probe with `"name" in tools`, not `typeof`.
  - **Discovery:** `searchTools(query, {limit, namespace})` (BM25 from `tool_search`), `describeTool(name)` and `describeNamespace(name)` (MCP servers).
  - **Store:** `store`/`load` persist in session state (`codemode_store`), applied only when the script succeeds.
- **Output:**
  - **Format:** `Script completed|Script failed\nWall time X seconds\nOutput:\n` followed by the text items. A returned value is appended like `text()`.
  - **Failures:** they add `Script error:\n<stack>` plus Pi's tool-call summary, and the tool result is an error.
  - **Long output:** past `max_output_tokens` (default 10000, 4 characters per token) the text keeps its start and end. The full text is saved at `vfs://codemode-output/<session>/<id>.txt`, which is pruned after 7 days with `mcp-output`.
  - **Images:** attached to the tool result.
- **Limits:** 256 MiB of memory; no timeout unless `timeout_ms` is set. Aborting the turn cancels the script.
- **Enabling codemode:**
  - **Built-in:** codemode is registered unless `extensions` contains `-builtin:codemode`.
  - **Declared by default:** when `defaultTools` enables it (layered user→project: `+codemode`/`-codemode` edit the selection, plain names replace it), or when an enabled MCP server has codemode exposure and `autoEnableCodemode` is not false (an explicit `-codemode` wins).
  - **Session toggle:** `codemode_mode` (`on`/`off`/`only`, set by `Engine.SetSessionCodemode` and the TUI's `/codemode [on|off|only|default|status]`) overrides this at admission. `/tools activate codemode` is the same as `/codemode on`, and `/tools reset` also returns codemode to its settings default. When MCP servers have codemode tools but codemode is off, gi logs a warning at startup and `/codemode status` says so. `only` also forces the `only` presentation, as does `codemode.mode: "only"` in settings.

## Rendering (TUI)

Codemode calls render like Pi's codemode renderer (`internal/tui/codemode_render.go`):

- **Call:** the `codemode` title and the script, highlighted as JavaScript. Collapsed, it shows the first 10 lines.
- **Nested calls:** they are rows of the codemode block, never separate tool blocks. Each row shows a status icon (… ✓ ✗ ⊘), the name, an argument preview (80 characters collapsed), the duration and the cost. Collapsed, the block shows the last 8 rows; expanded, failed rows also show their error.
  - **While running:** rows come from the nested calls' runtime events (`parent_tool_call_id`).
  - **When done:** the tool's result details (`ToolRuntime.SetDetails`: `calls`, `fullOutputPath`) replace them. The details are also stored with the `tool_result` message, so reloaded sessions render the same.
- **Output:** the script output without the `Script completed|failed / Wall time / Output:` header. Collapsed, it shows 5 lines, plus `Full output: vfs://…` when the output was truncated.

The web UI shows tool calls only while they run (its status panel), so it has no codemode result view.

## Models API (`internal/turn/codemode_models.go`)

`models.*` is a port of Pi's `createModelGlobals`. The typed catalogs come from go-ai v1.0.1 (`internal/inference/models_api.go`).

- **`getModelsOfType(type, provider?)`, `getModelOfType(type, provider, id)`:** chat, image and classifier entries, without `headers`.
- **`getAvailableOfType(type, provider?)`:** keeps only models whose provider has credentials. For chat that is the model picker's authenticated models; for the others, an auth.json entry or the provider's environment variable.
- **`classify(model, context)` and `generateImages(model, context)`:**
  - **Model:** resolved by provider and id only, so a script-supplied baseUrl or headers never receive credentials.
  - **Context:** checked with Pi's messages.
  - **Running:** at most 4 at once per script; provider errors come back as results (`stopReason`, `errorMessage`).
  - **Call rows:** each call is a `models.classify` / `models.generateImages` row with the call's cost.
  - **Usage:** added to the turn's usage and cost (`ToolRuntime.AddUsage`).
  - **Unshown images:** a script that generates images without showing them gets Pi's note.
- **Description:** the codemode description includes Pi's models line, pointing at `vfs://reference/codemode-scripts.md`.
