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
- real `queue_count` synchronization from queued turn rows
- compaction checkpoints now mark a durable `compacting` phase
- stale active-turn recovery on engine startup and before same-session submission
- advisory `turn_failures` rows for durable failure/recovery postmortems
- regression coverage for same-session concurrency, active-turn row lifecycle, stale-turn recovery, and non-blocking failure markers

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

- fuller lifecycle separation for setup / provider / tool / finalize / recovery

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
- store/unit coverage for dequeue mode behavior and turn/engine coverage for same-session steering continuation
- steering lifecycle publication on the topic bus via `session.steering` notices for enqueue/dequeue/stage/continue/inject checkpoints
- steering queue overflow coverage (cap remains enforced at 10)
- different-session concurrent submit coverage (sessions execute independently)
- explicit skipped-tool persistence assertions (`tool.skipped` events plus skipped `tool_result` rows)
- steering media payloads are now preserved in persisted chat history payloads during injection/continuation

Current steering semantics:

- Gi now follows PicoClaw's core rule that same-session inbound messages during an active turn become steering
- steering does not interrupt the currently executing tool; it is observed at explicit checkpoints
- when steering is found after a tool, remaining tool calls in that batch are skipped and recorded as skipped results
- when steering arrives after a direct non-tool LLM answer, the loop continues instead of finalizing that answer immediately
- a final pre-finalization checkpoint stages a queued continuation turn before the active turn is released when late steering is already waiting
- idle sessions can be resumed explicitly through `ContinueSession(...)` / the web continuation endpoint, and runtime fallback continuation still runs after turn end when appropriate
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

Still pending in this area:

- maximum depth / concurrency guardrails at runtime
- explicit async delivery/orphan-result handling modes
- restricted tool inheritance policies per subturn

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
- script hook responses now accept canonical hook actions and map them into runtime semantics:
  - `continue`
  - `modify`
  - `respond`
  - `deny`
  - `abort_turn`
  - `hard_abort`

Still pending in this area:

- durable `hook_invocations` audit table + store APIs
- process-hook/IPC handshake for external hook executors sharing the same logical contract as in-process hooks
- timeout/failure policy and richer tracing metadata for hook replay

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

### Add: `inbound_work_queue`

```sql
create table inbound_work_queue (
  id integer primary key autoincrement,
  source_kind text not null,
  session_id text references sessions(id) on delete set null,
  explicit_session_key text not null default '',
  envelope_json text not null,
  status text not null default 'queued',
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
- `docs/internal/topic-system.md`
- future ADRs/docs for session identity, turn coordination, steering, sub-turns, and hooks
