# Runtime target state

Date: 2026-05-05

This document captures the intended **database-backed target state** for Gi's core runtime refactor.

It is derived from the target-state schema discussion recorded in message `5175` and should be treated as the current canonical design reference for the runtime surfaces that need to move from implicit/in-memory behavior to explicit SQLite-backed coordination.

## Goal

Refactor Gi's core runtime around these first-class surfaces:

1. session identity / allocation
2. turn coordination
3. steering semantics
4. sub-turn runtime
5. hook protocol shape
6. event bus / topic system
7. multi-channel session bindings
8. IPC / direct-processing mode

## Design bias

The design intent is:

- copy the **runtime semantics** we want from PicoClaw
- keep **SQLite/WAL** as the canonical shared state
- avoid inheriting PicoClaw's JSONL storage assumptions
- treat JSON blobs as **compatibility or metadata escape hatches**, not the core coordination model

## Guiding rule

Do **not** stuff the critical runtime model into:

- `sessions.state_json`
- `turns.metadata_json`

Those fields remain useful for:

- incidental UI state
- compatibility snapshots
- extra metadata
- sparse optional payloads

But they should not be the primary storage model for:

- canonical session identity
- active turn coordination
- steering queues
- sub-turn lineage
- channel bindings

Those need first-class relational tables.

---

## 1. Session identity / allocation

### Keep

- `sessions`
- `scope_json` as a compatibility/debug snapshot

### Add: `session_identities`

Canonical routed identity, one row per logical session identity.

```sql
create table session_identities (
  session_id text primary key references sessions(id) on delete cascade,
  agent_id text not null,
  channel text not null,
  account text not null,
  scope_version integer not null default 1,
  canonical_scope_signature text not null,
  opaque_session_key text not null unique,
  is_main_session integer not null default 0,
  created_at text not null,
  updated_at text not null
);
```

### Add: `session_identity_dimensions`

Normalized dimensions instead of burying them in JSON.

```sql
create table session_identity_dimensions (
  session_id text not null references sessions(id) on delete cascade,
  dimension_name text not null,
  dimension_value text not null,
  ordinal integer not null,
  primary key (session_id, dimension_name)
);
create index idx_sid_dim_lookup
  on session_identity_dimensions(dimension_name, dimension_value);
```

### Add: `session_aliases`

Stop resolving aliases by scanning all sessions.

```sql
create table session_aliases (
  alias text primary key,
  session_id text not null references sessions(id) on delete cascade,
  alias_kind text not null,
  created_at text not null,
  updated_at text not null
);
create index idx_session_aliases_session on session_aliases(session_id);
```

### Why

This gives us:

- canonical lookup by opaque key
- alias resolution without full-table scans
- stable identity independent of current channel
- a clean base for later multi-channel continuation

### Current implementation status

Implemented so far:

- `session_identities`, `session_identity_dimensions`, and `session_aliases` are now populated transactionally with session creation/cloning paths
- store now exposes first-class identity APIs (`GetSessionIdentity(...)`, `ListSessionIdentities(...)`, canonical-key/alias resolution, alias list/update, resolve-or-create from allocation, and main-session resolve/promote semantics) backed by the relational identity tables instead of requiring runtime callers to infer identity from `sessions.scope_json`
- single-session identity reads now hydrate dimensions/aliases with targeted per-session queries instead of falling back to list-style full-table detail scans
- `FindSessionByAllocation(...)` now resolves by opaque key, validated alias matches, and canonical scope signature fallback
- allocation-backed store resolution/create paths now normalize `identity_links` at the store boundary, so linked sender identities collapse to the canonical session key/signature even when the incoming allocation was not already pre-canonicalized upstream
- `session_channel_bindings` now provide a first DB-backed multi-channel attachment path: main-session create paths register their primary chat binding, explicit alternate channel/account chat bindings can be attached to the same session, and allocation resolution consults those bindings before creating a duplicate session (while still requiring agent-id agreement)
- allocation-backed resolve-or-create now also supports explicit cross-channel continuation into an existing session via `ContinueSessionID`; when used, the new channel/account chat binding is attached to that session so future allocations on that inbound surface resolve back to the same history without another explicit handoff
- this means the current runtime now supports both "one session, many channel identities" (via bindings) and "temporary continuation from another channel" (via explicit continuation + bound reuse), while leaving broader automatic cross-channel policy for a later slice
- default session allocation now uses the same `direct:<chat>` chat-dimension encoding and alias format as routed allocation
- route-preparation and same-agent fast-path checks now read canonical agent/channel/account/dimension identity from the store first, with `scope_json` only as compatibility fallback
- web/TUI fork-agent allocation and TUI agent/session resolution now prefer canonical identity rows instead of trusting `sessions.scope_json` snapshots
- explicit root-session creation paths now promote the created session to the stored main session for its `(agent, channel, account)` tuple, and TUI startup prefers that stored main session over incidental recency ordering

Still pending in this area:

- replacing remaining runtime reads that still infer agent/channel/account from compatibility `scope_json` helpers instead of a first-class identity read path
- main-session semantics still only cover one preferred session per `(agent, channel, account)` tuple; they do not yet express richer multi-channel binding / continuation policy
- broader automatic cross-channel continuation policy and outbound fan-out semantics still need to be defined above the new binding table; current continuation support is explicit store-level continuation plus binding-aware reuse

---

## 2. Turn coordination model

The current `turns` table is close, but not explicit enough for coordinated runtime semantics.

### Extend `turns`

Add explicit phase/runtime columns, not just `status`.

```sql
alter table turns add column phase text not null default 'queued';
alter table turns add column claimed_by text;
alter table turns add column claimed_at text;
alter table turns add column started_at text;
alter table turns add column finished_at text;
alter table turns add column parent_turn_id text references turns(id);
alter table turns add column root_turn_id text;
alter table turns add column retry_of_turn_id text references turns(id);
```

### Add: `session_active_turns`

This is the concurrency primitive.

```sql
create table session_active_turns (
  session_id text primary key references sessions(id) on delete cascade,
  turn_id text not null references turns(id) on delete cascade,
  worker_id text,
  claim_token text not null,
  claimed_at text not null,
  updated_at text not null
);
```

### Optional: `turn_failures`

For held-failure / retry / skip semantics.

```sql
create table turn_failures (
  turn_id text primary key references turns(id) on delete cascade,
  session_id text not null references sessions(id) on delete cascade,
  failure_kind text not null,
  hold_state text not null,
  summary text not null default '',
  created_at text not null,
  updated_at text not null
);
```

Safety rule: this table is **advisory**, not authoritative. It must never decide whether a session is runnable, queued, or stuck. Active-turn claims and queued-turn rows remain the only coordination truth.

### Why

This gives us:

- exactly one active coordinating worker per session
- durable claim/release semantics
- explicit turn lifecycle
- crash recovery without guessing from `state_json`

### Current implementation status

Implemented so far:

- `turns.phase`, `claimed_by`, `claimed_at`, `started_at`, and `finished_at`
- `session_active_turns` claim/release store APIs plus active-turn heartbeat refreshes
- store-backed run-vs-queue decisions in `turn.Engine.SubmitPrompt(...)`
- queued-turn handoff through the same launch/claim path as immediately-started turns
- same-session cross-engine launch claim conflicts now collapse back into steering instead of leaving a transient competing queued turn behind when another worker wins the active-turn claim first
- if that steering fallback cannot enqueue because the steering queue is already full, the already-persisted queued turn is now kept as a safe fallback instead of dropping the prompt
- once a turn row is durably created, submit-time launch/orchestration writes now continue under a background coordination context so caller/request cancellation cannot make a persisted turn look like submission failed while leaving partially scheduled work behind
- submit-time post-persist work now distinguishes rollback-worthy vs best-effort follow-up: integral subturn-link creation rolls the child turn back on failure, while non-integral queue/system-message and `turn.submitted` bookkeeping warn instead of surfacing a false submission failure after the turn already exists
- runner-owned active-turn cancellation handles are now installed before the run goroutine starts, and turn cleanup only clears `runner.current` when it still points at the same finishing turn so an older cleanup path cannot wipe a newer active turn's cancel handle
- non-cancellation setup failures now finalize the turn as terminal `failed` / `setup_error` instead of leaving a turn stranded in `running` / `setup` until stale-claim recovery notices it
- same-store engine instances now share the same in-memory session coordination lock/current-turn handle for a given session, so cross-engine submit/cancel paths coordinate on one live ownership record instead of each engine keeping a disconnected local `runner.current`
- `ContinueSession(...)` now runs its idle no-active check, queued-turn launch attempt, and steering-continuation staging/launch under that same shared session coordination lock, so manual continue behaves like one atomic same-session ownership decision instead of a sequence of loosely related checks
- turn cleanup now makes its release-active-turn / clear-current / launch-next-queued-or-continued-work decision under the same shared session coordination lock as submit/continue/cancel, so a new same-session submit cannot slip in between active-turn release and continuation scheduling
- real `queue_count` synchronization from queued turn rows
- compaction checkpoints now mark a durable `compacting` phase
- stale active-turn recovery on engine startup and before same-session submission
- advisory `turn_failures` rows for durable failure/recovery postmortems
- regression coverage for same-session concurrency, active-turn row lifecycle, stale-turn recovery, and non-blocking failure markers
- runner/orchestration now uses explicit phase helpers for setup, context assembly, provider iteration, tool execution, final steering/finalize, and runner cleanup instead of keeping the whole lifecycle implicit inside one long goroutine body
- routed submission now uses explicit route/session-resolution helpers for prompt preparation, peer-route preparation, target-session resolution, local-route metadata application, allocation lookup/reuse, and clone-on-miss session creation
- active cancellation now resolves through terminal `cancelled` state for live provider-stream, setup-phase, and live tool-execution turns instead of being misclassified as generic provider/tool failures, and parent-turn cancellation propagates into running child subturns through the normal terminal path
- queued turns are launched oldest-first via durable `created_at`/`id` ordering, and queued ordering is now covered by runtime tests rather than only implied by store queries

Current recovery semantics:

- `compacting`, `setup`, and general `running` interruptions are re-queued
- `waiting_on_tools` interruptions are moved into `held_for_retry_or_skip` with `hold_state='review'` to avoid silently replaying side effects
- `cancelling` interruptions are finalized as aborted
- stale active claims are released and session state is normalized back to `idle` or `queued`

Current failure-marker semantics:

- failure markers are written for provider, shell, repeated-tool-failure, and recovery-conservative failure cases
- failure markers can now carry explicit hold states such as `review`
- failure markers are cleared automatically when a turn goes back to `queued` or `running`, or reaches `completed`
- resolved held failures retain advisory audit metadata (`resolution_state`, `resolution_summary`, `resolved_at`, `resolved_turn_id`) without affecting liveness
- failure markers do not participate in queue decisions, claim decisions, or session liveness

Current hold / retry / skip semantics:

- held failures use turn phase `held_for_retry_or_skip` while the turn status remains terminal (`failed`/`aborted`)
- hold state is explicit review metadata, not a coordination lock
- retry resolution creates a **new** turn through normal `SubmitPrompt(...)` queue/claim logic instead of mutating the failed turn back into runnable state
- skip resolution clears the hold and returns the original turn to its normal terminal phase
- later fresh submissions continue normally even when held/resolved failure rows exist

Still pending in this area:

- keep tightening direct-processing/IPC and future multi-channel inputs so they reuse the same route/session-resolution phase helpers instead of growing side paths

---

## 3. Steering semantics

This should not live only in memory if we want reconnect correctness or crash recovery.

### Add: `steering_queue`

```sql
create table steering_queue (
  id integer primary key autoincrement,
  session_id text not null references sessions(id) on delete cascade,
  turn_id text references turns(id) on delete set null,
  role text not null default 'user',
  content text not null default '',
  payload_json text not null default '{}',
  media_json text not null default '[]',
  queue_mode text not null default 'one-at-a-time',
  status text not null default 'queued',
  created_at text not null,
  updated_at text not null
);
create index idx_steering_queue_session_status
  on steering_queue(session_id, status, id);
```

### Why

This gives us:

- same-session serialization
- idle `Continue` semantics
- durable queued steering
- postmortem visibility into skipped/injected messages

### Current implementation status

Implemented so far:

- session-scoped durable steering rows in `steering_queue`
- dequeue modes `one-at-a-time` and `all`
- same-session busy submits now steer the active turn instead of immediately creating competing queued turns
- steering polling at loop start, after each tool, after direct LLM responses, and at a final pre-finalization checkpoint
- skipped remaining tool calls emit durable skipped tool results with `"Skipped due to queued user message."`
- end-of-turn idle continuation drains queued steering into a follow-on turn when no normal queued turn is ahead of it
- explicit `ContinueSession(...)` support plus a web continuation endpoint for idle-session steering
- idle continuation now stages a durable queued continuation turn before launch, so continued steering gets an ordered queue slot before execution starts
- store/unit coverage for dequeue mode behavior and turn/engine coverage for same-session steering continuation
- steering lifecycle publication on the topic bus via `session.steering` notices for enqueue/dequeue/stage/continue/inject checkpoints
- steering queue overflow coverage (cap remains enforced at 10)
- steering enqueue failures (for example queue-full rejection) now publish explicit `steering_rejected` runtime events / topic-bus notices instead of only returning a synchronous tool/API error
- different-session concurrent submit coverage (sessions execute independently)
- explicit skipped-tool persistence assertions (`tool.skipped` events plus skipped `tool_result` rows)
- steering media payloads are now preserved in persisted chat history payloads during injection/continuation

Current steering semantics:

- Gi now follows PicoClaw's core rule that same-session inbound messages during an active turn become steering
- steering does not interrupt the currently executing tool; it is observed at explicit checkpoints
- when steering is found after a tool, remaining tool calls in that batch are skipped and recorded as skipped results
- when steering arrives after a direct non-tool LLM answer, the loop continues instead of finalizing that answer immediately
- a final pre-finalization checkpoint stages a queued continuation turn before the active turn is released when late steering is already waiting
- steering continuation staging is now transactional at the store layer: "no queued turn exists" + dequeue steering + create queued continuation turn happen in one SQLite transaction so continuation ordering does not race against other queued work
- idle sessions can be resumed explicitly through `ContinueSession(...)` / the web continuation endpoint, and runtime fallback continuation still runs after turn end when appropriate
- continuation no longer drains steering directly into a fresh submit call; it first stages a queued continuation turn and then launches through normal queue/claim logic, reducing the race window against concurrent same-session submits
- bootstrap/test shell turns now preserve queued steering messages in history before running the continuation shell step

Still pending in this area:

- media-bearing steering injection through the exact same multimodal message-content block path as normal inbound web messages (currently media is preserved in steering payload/history and surfaced to the model as attachment hints)

---

## 4. Sub-turn runtime

This should be explicit rather than implicitly encoded inside generic turns.

### Add: `subturns`

```sql
create table subturns (
  id integer primary key autoincrement,
  parent_turn_id text not null references turns(id) on delete cascade,
  parent_session_id text not null references sessions(id) on delete cascade,
  child_turn_id text not null unique references turns(id) on delete cascade,
  child_session_id text not null references sessions(id) on delete cascade,
  delivery_mode text not null default 'sync',
  status text not null default 'running',
  depth integer not null default 1,
  metadata_json text not null default '{}',
  created_at text not null,
  updated_at text not null,
  finished_at text,
  unique(parent_turn_id, child_turn_id)
);
create index idx_subturns_parent on subturns(parent_turn_id, created_at asc);
create index idx_subturns_child on subturns(child_turn_id);
create index idx_subturns_parent_session on subturns(parent_session_id, created_at asc);
create index idx_subturns_child_session on subturns(child_session_id, created_at asc);
```

### Why

This gives us:

- parent/child lineage between turns with explicit DB-backed correlation
- durable audit for child-turn lifecycle progression
- per-parent listing and child-turn lookup without scanning generic turn metadata

### Current implementation status

Implemented so far:

- schema and store APIs for creating/listing/updating subturn records
- automatic subturn-link creation when a turn is submitted with `parent_turn_id`
- subturn status synchronization from child turn status transitions (`running`/`completed`/`failed`/`aborted`/`cancelled`)
- runtime depth guardrails for parent/child submission chains (default max depth = `8`)
- runtime per-parent concurrency guardrails for running child turns (default max concurrency = `4`)
- explicit store API support for counting running child turns per parent for guardrail enforcement
- synchronous and asynchronous result delivery modes (`sync` default, `async` opt-in) with explicit lifecycle/result topic events
- async orphan-result handling when parent turns are already terminal (durable metadata marker + parent notice + `subturn_orphaned` event)
- parent-terminal cancellation propagation for child subturns: graceful completion cancels non-critical children, hard abort/timeout cancels descendants regardless of criticality
- sub-turn tool inheritance/restriction: child turns persist `effective_tools`, inherit from parent by default, and may restrict to explicit subsets via subturn metadata
- sub-turn lifecycle publication on topic bus (`subturn_created`, `subturn_status`, `subturn_result_ready`, `subturn_result_delivered`, `subturn_orphaned`, `subturn_cancel_requested` → `turn.subturn`)

Still pending in this area:

- richer tool policy classes/presets per subturn if we later need more than explicit tool-name subsets

---

## 5. Hook protocol shape

Hook configuration may remain file/config-driven, but runtime activity benefits from persistence.

### Add: `hook_invocations`

```sql
create table hook_invocations (
  id integer primary key autoincrement,
  turn_id text references turns(id) on delete cascade,
  session_id text references sessions(id) on delete cascade,
  hook_name text not null,
  hook_phase text not null,
  hook_source text not null,
  action text not null default 'continue',
  request_json text not null default '{}',
  response_json text not null default '{}',
  error_text text not null default '',
  duration_ms integer not null default 0,
  created_at text not null
);
create index idx_hook_invocations_turn on hook_invocations(turn_id, id);
create index idx_hook_invocations_phase on hook_invocations(hook_phase, created_at);
```

### Why

This supports:

- replay/debugging
- approval audits
- process-hook protocol validation
- tracing hook-induced mutations

### Current implementation status

Implemented so far:

- hook registry/runtime supports canonical alias names for script-facing phases:
  - `before_llm` → `before_provider_request`
  - `after_llm` → `after_provider_response`
  - `before_tool` → `tool_call`
  - `after_tool` → `tool_result`
- explicit `approve_tool` phase is now emitted in the tool execution path (after `tool_call`, before tool execution)
- explicit state-oriented hook phases are now emitted for runtime state transitions:
  - `turn_state`
  - `session_state`
- hook request/response DTOs now carry JSON-safe/script-safe structured fields instead of dropping them from JSON:
  - `messages`
  - `tools`
  - `tool_call`
  - `tool_result`
  - `session_status`
  - `turn_status`
  - `turn_phase`
- hook requests now include trace metadata scaffolding for debugging/replay correlation:
  - `trace.id`
  - `trace.emitted_at`
- hook execution now has explicit timeout/failure policy with runtime-config defaults:
  - `hooks.timeout_ms` (default `1500`)
  - `hooks.on_error` (default `error`)
  - `hooks.on_timeout` (default `continue`)
- hook failures now surface as typed execution errors carrying:
  - hook name
  - hook source
  - failure kind (`handler_error` / `timeout` / `panic`)
  - trace metadata
- durable `hook_invocations` persistence is now implemented with schema + store APIs and per-handler audit rows carrying:
  - hook name / phase / source
  - action
  - request / response JSON
  - error text
  - duration_ms
  - created_at
- `before_provider_request` now has two effective checkpoints in the runtime: an early context-mutation stage (`system_prompt`, `messages`, `tools`) before streaming begins, plus a send-time payload stage where hook responses may replace the raw serialized provider request via `payload.request`
- `internal/inference` now exposes a hook-aware stream path that wires through `go-ai` request/response interception (`OnPayload` / `OnResponse`) and mirrors the same response-hook behavior for the custom OpenCodeZen streaming path
- `after_provider_response` now receives real observed provider response metadata (`status`, `headers`, provider/api identifiers) when the provider path exposes it, instead of only a synthetic success/failure marker
- `tool_call` now supports direct hook-handled responses (`respond`) that inject a tool result without executing a registered tool implementation, enabling plugin-style hook tools and cached/mock direct responses
- `respond` on `tool_call` now skips later `tool_result` hooks for that tool call, matching PicoClaw's documented behavior
- `abort_turn` / `hard_abort` semantics are now honored in active interception paths (`before_provider_request`, `tool_call`, `approve_tool`, `tool_result`) and finalize the current turn as `aborted`
- external process hooks are now supported through `EventHookSpec{engine:"process", command, args, env, cwd}` with:
  - stdio transport
  - JSON-RPC 2.0 request/response framing
  - a mounted persistent subprocess per registered process-hook handler instead of spawn-per-invocation startup
  - `hook.hello` handshake followed by `hook.<phase>` invocation over the mounted session
  - PicoClaw-style method aliases for major interceptor phases (`before_llm`, `after_llm`, `before_tool`, `after_tool`)
- process hooks and in-process/script hooks now share the same logical JSON payload/response contract:
  - request params reuse the JSON-safe `HookRequest` shape (plus convenience aliases like `tool`, `arguments`, `request`, `result`)
  - response result reuses the same `HookResponse` JSON fields/action semantics interpreted by `hookResponseFromScript(...)`
- script hook responses now accept canonical hook actions and map them into runtime semantics:
  - `continue`
  - `modify`
  - `respond`
  - `deny`
  - `abort_turn`
  - `hard_abort`

Still pending in this area:

- cleaner dedicated runtime-facing documentation/DTO ergonomics for the send-time raw provider request stage, instead of today’s `before_provider_request` + `payload.request` extension of the existing hook contract
- fuller mounted-process lifecycle management semantics (for example explicit teardown/reload cleanup across hook-registry resets) beyond the current restart-on-error mounted session behavior
- clearer replay/audit persistence around hook decisions beyond the current per-handler audit rows

---

## 6. Event bus / topic system

We do not need persistence for every event, but we probably want optional audit rows.

### Add: `runtime_events`

```sql
create table runtime_events (
  id integer primary key autoincrement,
  topic text not null,
  session_id text references sessions(id) on delete cascade,
  turn_id text references turns(id) on delete cascade,
  agent_id text,
  source text not null,
  type text not null,
  payload_json text not null default '{}',
  created_at text not null
);
create index idx_runtime_events_topic on runtime_events(topic, created_at);
create index idx_runtime_events_session on runtime_events(session_id, created_at);
create index idx_runtime_events_turn on runtime_events(turn_id, created_at);
```

### Constraint

This should be **selective**, not a firehose dump of every in-memory event.

---

## 7. Multi-channel sessions

This should be separate from canonical session identity.

### Add: `session_channel_bindings`

```sql
create table session_channel_bindings (
  id integer primary key autoincrement,
  session_id text not null references sessions(id) on delete cascade,
  channel text not null,
  account text not null,
  binding_type text not null,
  remote_identity text not null,
  metadata_json text not null default '{}',
  created_at text not null,
  updated_at text not null,
  unique(channel, account, remote_identity)
);
create index idx_session_channel_bindings_session
  on session_channel_bindings(session_id, channel, account);
```

### Why

This enables:

- one logical session across multiple channels
- continuation from another channel
- explicit or automatic linking later

---

## 8. IPC / direct-processing mode

If we want PicoClaw-style direct mode, we need an inbound work representation.

Current implementation status:

- `internal/turn` now exposes a normalized direct-ingress envelope via `DirectInput` / `DirectOrigin`
- `DirectInput` now supports both `SessionID` and explicit `SessionKey`, with direct processing resolving explicit session keys through the canonical store lookup path before entering runtime submission; when both are supplied they must resolve to the same session or the direct request is rejected as ambiguous
- `Engine.ProcessDirect(...)` routes direct-origin prompt, peer-message, and continue requests through the same existing queued/runtime submit paths (`SubmitPromptRouted(...)`, routed peer submit, `ContinueSession(...)`) instead of inventing a separate execution path
- `Engine.ProcessSystemDirect(...)` and `Engine.ProcessInternalDirect(...)` now make system/internal-origin processing explicit on top of the same direct envelope, defaulting the origin kind/role instead of relying on callers to smuggle those values through ad hoc metadata
- routed direct prompts now reuse the normal prompt-routing path as well, including routed target session creation/reuse and ingress metadata propagation onto the resulting target turn
- same-session direct/system ingress while a turn is already active now reuses the existing steering path rather than spawning a competing turn, so IPC/system-origin follow-up messages serialize the same way as web/TUI same-session input; system/internal-origin follow-ups also preserve their origin role on the queued steering row instead of degrading back to a generic user steering role
- direct-origin turns now stamp normalized ingress audit metadata onto the same persisted audit surfaces used by normal chat-origin turns: turn metadata, persisted user-message payloads, and `turn.started` event payloads (`ingress_kind`, `ingress_source_kind`, `ingress_source_id`, `ingress_role`, `ingress_label`)
- `inbound_work_queue` now provides a first durable queue surface for direct/IPC/system work, with store-backed enqueue/list/get/atomic-claim/status APIs
- `Engine.EnqueueDirectInbound(...)`, `Engine.ProcessNextInboundWork(...)`, and `Engine.ProcessQueuedInboundWork(...)` now let callers enqueue normalized direct envelopes durably and drain them back through the same `ProcessDirect(...)` runtime path instead of creating a separate queue-only execution flow
- the guarded web runtime surface now exposes that narrow queue path via `/api/runtime/inbound-work` (enqueue/list with optional `eligible=true|false` filtering and queue counts), `/api/runtime/inbound-work/drain` (bounded drain), `/api/runtime/inbound-work/requeue` (manual recovery for `failed`/`retry` rows), and `/api/runtime/inbound-work/discard` (terminal discard without deleting audit history), so the inbound queue is reachable through a real runtime caller instead of only tests/internal code
- the web server now also starts a small configurable background inbound-work dispatcher in the main web runtime, which periodically drains bounded batches through the same engine path rather than inventing a separate worker execution model
- dispatcher ownership is now coordinated through a small store-backed lease, so multiple web runtimes sharing the same SQLite store do not all drain the inbound queue concurrently; only the current lease holder runs the bounded drain loop at a given moment
- inbound queue lifecycle and dispatcher ownership/drain activity now also publish onto the canonical topic bus under `runtime.inbound_work` and `runtime.dispatcher`, so this newer direct/IPC runtime surface is observable through the same topic layer as steering/subturn/routing events
- per-handler hook invocation lifecycle now publishes onto the canonical topic bus under `runtime.hook`, carrying hook name/source/action/error/duration metadata alongside the durable `hook_invocations` audit rows; higher-level hook decision notices around tool-call/approval/result abort/deny/modify/respond paths now also publish there so consumers do not have to infer every important hook outcome from raw invocation rows alone, and the bridge/audit/finalization/recovery bookkeeping around this newer topic/runtime surface now consistently prefers explicit caller context or engine-owned background context over raw `context.Background()` so detached work survives request cancellation without silently becoming process-lifetime work
- core runtime turn/session lifecycle checkpoints now also publish onto the canonical topic bus under `runtime.turn` and `runtime.session`; this now includes generic `turn_state` / `session_state` notices emitted from the shared state-hook path plus explicit setup and terminal/idle checkpoints, with the explicit checkpoint path kept non-duplicated, completed edge paths normalized onto `turn_completed` / `turn_completed`-reasoned idle semantics, completed exits now carrying explicit completion metadata such as `iterations` / `completion_kind` across both turn and session runtime notices, interrupted-turn recovery publishing explicit `turn_recovered` checkpoints plus generic recovery `turn_state` / `session_state` notices while also restarting queued follow-up work automatically when stale-claim recovery leaves runnable queued work behind, active cancellation publishing explicit `turn_cancelling` checkpoints plus richer session-state metadata, queued cancellation publishing explicit terminal turn/session checkpoints when it leaves the session idle while also continuing later queued work when queued follow-up remains, and setup failure/cancel terminal finalization now preserving resolved model/agent identity before publishing terminal runtime/session notices, while still stopping short of a full mirror of every stored turn event row
- core runtime tool lifecycle checkpoints now also publish onto the canonical topic bus under `runtime.tool`; this currently covers started/finished/failed/skipped checkpoints at the real execution sites, not every approval or hook sub-phase around tool execution
- persisted routing/allocation decision notices now also publish onto the canonical topic bus under `runtime.routing`; this currently covers route-decision plus incoming-route notices emitted from the same normalized decision object that is stored in `routing_events`, rather than every pre-persistence route-resolution branch
- both script bridges now expose canonical topic-bus helpers backed by the same internal bus: JS via `gi.topics.publish` / `subscribe` / `read` / `unsubscribe`, and embedded Joker via `gi-topic-publish` / `gi-topic-subscribe` / `gi-topic-read` / `gi-topic-unsubscribe`
- web consumers now also have a dedicated topic-stream SSE endpoint (`/sse/topics`) that subscribes directly to the canonical topic bus by topic pattern and optional session/agent scoping, instead of requiring every runtime family to be remapped through the legacy chat-turn SSE feed first
- the TUI now also has a first topic-native consumption slice for the active session, subscribing to canonical topic families with session scoping and using `turn.status` / `turn.response` / `turn.draft` / `turn.thought` plus `runtime.tool` / `runtime.hook` / `runtime.turn` / `runtime.session` / `runtime.routing` / `runtime.inbound_work` and `session.compaction` / `session.steering` / `turn.subturn` notices directly for status and transcript updates; `runtime.hook` rendering now covers invocation errors/timeouts plus abort/deny/modify/respond decisions, `runtime.inbound_work` rendering now covers queue/retry/failure/completion plus manual requeue/discard notices including retry attempt counts, canonical running-state entry is aligned across `turn.draft`, `turn.thought`, `turn.status` running, `runtime.tool` started, `runtime.turn` waiting-on-tools, and `runtime.session` running, `runtime.session` queued notices now render a visible queued state, canonical idle/completed/terminal cleanup is aligned across `turn.status` idle, `turn.response`, `runtime.turn`, and `runtime.session`, the newer explicit cancel/recovery checkpoints now arrive through the same canonical runtime topic families instead of being visible only through stored turn events, terminal system messages now broadcast live for all terminal outcomes and therefore arrive through the same `turn.response` / `system_message` bridge the TUI already consumes, and those running-entry / cleanup paths are now centralized in helpers to reduce drift; routing, tool, hook, inbound-work, compaction, steering, and sub-turn transcript/status rendering are now centralized too, and when that topic subscription is live, overlapping legacy session-broadcast handling serves as fallback compatibility rather than the primary rendering source, while the remaining legacy fallback status/draft/thought plus completion/error cleanup paths still mirror the same entry/reset semantics when topic-native mode is inactive and legacy routing broadcasts still render through fallback
- inbound queue rows now carry first-class retry state (`attempt_count`, `last_error`, `next_attempt_at`), and queue processing applies bounded backoff/retry before a row becomes terminal `failed`; retrying items are skipped until eligible again instead of being hot-looped immediately
- the background dispatcher now continues past retrying/failed items within the same sweep, so one bad queued envelope does not prevent later eligible work from being processed
- the current slice is still intentionally narrow: it provides durable queue rows plus engine-side processing/drain helpers, a small guarded runtime caller, and a bounded background dispatcher with simple retry semantics, but not yet a broader inbound bus abstraction or richer multi-worker policy

### Add: `inbound_work_queue`

```sql
create table inbound_work_queue (
  id integer primary key autoincrement,
  source_kind text not null,
  session_id text references sessions(id) on delete set null,
  explicit_session_key text not null default '',
  envelope_json text not null,
  status text not null default 'queued',
  attempt_count integer not null default 0,
  last_error text not null default '',
  next_attempt_at text,
  claimed_by text,
  claimed_at text,
  created_at text not null,
  updated_at text not null
);
create index idx_inbound_work_queue_status on inbound_work_queue(status, id);
```

### Why

This decouples:

- ingestion
- routing
- turn claiming
- replay/recovery

without making web/TUI special forever.

Current implementation direction:

- Gi is choosing the **narrower DB-backed inbound queue first** rather than a broader shared message-bus abstraction for inbound work
- that queue slice has now expanded from store/schema primitives into a bounded but real runtime surface: engine enqueue/claim/process/drain, guarded web runtime controls, retry/backoff and manual recovery state, lease-coordinated dispatcher ownership, canonical topic-bus publication, and dedicated topic SSE delivery
- the wider shared inbound-bus abstraction decision is still deferred, but new direct/IPC/system ingress is now converging on one durable queue-backed path rather than growing parallel execution flows

---

## Recommended minimum schema slice first

If we want the smallest useful first move, implement only these first:

1. `session_identities`
2. `session_identity_dimensions`
3. `session_aliases`
4. `session_active_turns`
5. `steering_queue`

That is enough to support:

- DB-backed canonical session resolution
- alias lookup without scans
- explicit active turn claims
- real queue accounting
- future steering semantics

---

## What to avoid

### Avoid overusing JSON blobs for core coordination

Fine for:

- UI state
- extra metadata
- compatibility payloads

Bad for:

- claims
- canonical identities
- queue state
- binding lookup
- alias resolution

### Avoid persisting every event by default

Keep the bus in-memory first.
Persist only:

- selected checkpoints
- failures
- audit-worthy hook/tool/turn events

---

## Practical reading of this target state

This target state does **not** mean every table needs to land at once.

It means:

- the runtime semantics should move toward explicit DB-backed coordination
- the schema should grow in the direction of first-class runtime concepts
- JSON stays available for compatibility, but should stop being the primary home of critical coordination state

## Relationship to the plan

This document is the schema/state target for the runtime refactor plan. Use it together with:

- the active runtime refactor checklist in the shared plan
- `docs/internal/repo-structure-refactor.md`
- `docs/internal/topic-system.md`
- future ADRs/docs for session identity, turn coordination, steering, sub-turns, and hooks
