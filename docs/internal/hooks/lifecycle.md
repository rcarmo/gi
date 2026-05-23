# Hook lifecycle contract

## Status

Implemented runtime contract, with lifecycle-management refinements still tracked in the core runtime plan.

## Purpose

This page defines the stable hook families Gi exposes to in-process Go hooks, script hooks, and persistent process hooks. Hook requests and responses are JSON-safe so the same logical contract can cross Go, JavaScript, Joker, and JSON-RPC process boundaries.

## Canonical hook families

### Session/state hooks

| Hook | When it runs | Mutations |
| --- | --- | --- |
| `session_state` | Whenever the runtime publishes a canonical session-state transition. | Observe-only in current runtime paths. |
| `turn_state` | Whenever the runtime publishes a canonical turn-state transition. | Observe-only in current runtime paths. |
| `before_agent_start` / `agent_start` / `agent_end` | Agent lifecycle boundaries used by launcher/runtime integration. | Observe or return control actions; active turn mutation is not the primary use. |
| `turn_start` / `turn_end` | Turn lifecycle boundaries. | Observe or return control actions; state transitions remain owned by the turn coordinator. |

### LLM/provider hooks

| Hook | Alias | When it runs | Mutations |
| --- | --- | --- | --- |
| `before_provider_request` | `before_llm` | Before provider streaming starts, and again at the send-time provider payload boundary. | May `modify` `system_prompt`, `messages`, `tools`, or replace the raw provider request through `payload.request` at send time. May `abort_turn` / `hard_abort`. |
| `after_provider_response` | `after_llm` | After a provider response or provider-path error is observed. | Observe provider status/headers and response metadata; may return control actions, but normal response content remains runtime-owned. |
| `model_select` | — | Model/provider selection boundary. | Reserved for model selection policy. |

### Tool hooks

| Hook | Alias | When it runs | Mutations |
| --- | --- | --- | --- |
| `tool_call` | `before_tool` | After the model proposes a tool call, before approval and execution. | May `modify` the tool call, `respond` with a direct tool result, or `abort_turn` / `hard_abort`. `respond` skips later `tool_result` hooks for that call. |
| `approve_tool` | — | After `tool_call`, before executing the tool. | Gate hook. May `deny` with a reason, causing a skipped-tool result, or abort the turn. |
| `tool_result` | `after_tool` | After an executed tool returns. Not emitted for `tool_call` direct `respond`. | May observe or modify the tool result shape where the execution path supports it; may abort. |
| `tool_execution_start` / `tool_execution_update` / `tool_execution_end` | — | Coarser tool-execution lifecycle notifications. | Primarily observe. |

### Approval hooks

`approve_tool` is the current explicit approval phase. A blocking response uses either:

- `action: "deny"`
- `block: true`

The runtime records a skipped-tool result with the denial reason and publishes `runtime.tool` lifecycle notices carrying `hook_phase: "approve_tool"`.

### Observer/event hooks

| Hook | When it runs | Mutations |
| --- | --- | --- |
| `message_start` / `message_update` / `message_end` | Streaming message lifecycle. | Observe-oriented. |
| `context` | Context assembly/injection boundary. | Reserved for context injection policy. |
| `user_bash` | User shell execution boundary. | Reserved for shell policy/audit. |
| Script-registered event hooks | Script/event bridge lifecycle. | Session-scoped observer callbacks; cleanup is session-scoped and idempotent. |

## Hook request envelope

All hooks receive the shared `HookRequest` shape:

- identity: `name`, `session_id`, `turn_id`, `agent_id`, `model`, `iteration`
- state: `session_status`, `turn_status`, `turn_phase`
- tracing: `trace.id`, `trace.emitted_at`
- generic payload: `payload`
- LLM fields: `system_prompt`, `messages`, `tools`
- tool fields: `tool_call`, `tool_result`, `tool_error`

The envelope is intentionally broad so the same DTO covers observation, gates, and mutations without introducing incompatible script/process variants.

## Hook response actions

| Action | Meaning |
| --- | --- |
| `continue` | Default. Continue without mutation. |
| `modify` | Merge supported response fields into the active request/result for mutation-capable phases. |
| `respond` | Return a direct response. Currently meaningful for `tool_call`; skips `tool_result` for that call. |
| `deny` | Block a gate, currently `approve_tool`, and synthesize a skipped-tool result. |
| `abort_turn` | Abort the active turn cleanly. |
| `hard_abort` | Abort the active turn as a hard-stop control decision. |

Legacy boolean fields (`cancel`, `block`, `handled`) are still interpreted as compatibility fields inside the shared response DTO, but the canonical contract should use explicit `action` values.

## Process-hook protocol

Process hooks use the same logical request/response contract over mounted JSON-RPC 2.0 subprocesses:

1. the runtime starts or reuses a persistent process-hook session,
2. sends `hook.hello`,
3. invokes `hook.<phase>` methods with the JSON-safe hook request,
4. parses the same hook response fields/actions used by in-process and script hooks.

Method aliases mirror the major runtime phases: `before_llm`, `after_llm`, `before_tool`, and `after_tool`.

## Timeout and failure policy

Runtime configuration controls hook execution behavior:

- `hooks.timeout_ms` (default `1500`)
- `hooks.on_error` (default `error`)
- `hooks.on_timeout` (default `continue`)

Hook failures surface as typed execution errors with hook name/source, failure kind (`handler_error`, `timeout`, or `panic`), trace metadata, and durable audit rows when a turn/session context is available.

## Audit and topics

Each handler invocation is persisted to `hook_invocations` with hook name, phase, source, action, request/response JSON, error text, duration, and timestamp. Higher-level hook decisions publish under the canonical `runtime.hook` topic family so web/TUI/script subscribers can observe deny/abort/modify/respond outcomes without reading the audit table.

## Ordering guarantees

- Hooks execute synchronously in registration order for a phase.
- Mutation hooks are chained in registration order.
- Gate hooks stop at the first blocking/denying response.
- Runtime state remains coordinated by the turn engine; hooks may request actions, but they do not own active-turn claims or session queue state.
