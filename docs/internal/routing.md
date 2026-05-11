# Routing and route-event introspection

This document tracks Gi's in-process routing layer and how routing decisions are persisted for observability.

It includes mechanism diagrams and concrete sequence flows for prompt, peer-routing, introspection, parent-turn steering, and the current direct/IPC ingress handoff into the same runtime paths.

## Current runtime shape

Route/session resolution now has an explicit orchestration surface inside `internal/turn/`:

- prompt route preparation (`preparePromptRouteResolution`)
- peer route preparation (`preparePeerRouteResolution`)
- target-session resolution (`resolveRoutedPromptTarget`)
- local-route metadata application (`applyLocalRouteMetadata`)
- session-allocation planning (`prepareRouteSessionPlan`)
- existing-session reuse lookup (`resolveExistingRouteSession`)
- clone-on-miss child-session creation (`cloneRouteSession`)

That keeps the submit path aligned with the newer explicit runtime-phase style used elsewhere in turn execution, instead of leaving route/session allocation buried inline inside `SubmitPromptRouted(...)` and `ResolveOrCreateRouteSession(...)`.

The routing path now also depends on store-backed session identity/allocation semantics rather than treating `sessions.scope_json` as the source of truth. Route/session resolution prefers:

1. canonical opaque session key lookup
2. channel-binding reuse when a bound inbound surface already maps to the same logical session
3. validated alias reuse
4. canonical scope-signature fallback
5. create-on-miss

Two guardrails matter here:

- alias matches are only accepted when the persisted session scope still matches the requested routed allocation, so broad compat aliases cannot accidentally collapse distinct routed sessions
- channel-binding reuse is agent-safe: a binding only qualifies if the bound session identity still belongs to the same agent as the incoming allocation

## Mechanism overview

```mermaid
flowchart TD
  subgraph Input[Input surfaces]
    WA["Web API: POST /api/sessions/{session}/prompt"]
    WPM["Web API: POST /api/sessions/{session}/peer-message"]
    TC["TUI command /send @agent ..."]
    HE["Inline @mention in user prompt"]
    DI["Direct/IPC envelope: ProcessDirect(...)"]
  end

  subgraph Routing[Routing + session resolution]
    SP["parseDirectedPrompt / parse mention"]
    RR["routing.ResolveRoute"]
    AR["Allocate/Resolve target session"]
    CL["session.AllocateRouteSession"]
    SESS["Create or reuse scoped/alias session"]
  end

  subgraph Engine[Turn engine]
    SUB["SubmitPrompt / SubmitPromptRouted / SubmitPeerMessage"]
    TURN["SubmitPrompt (queue + run path)"]
    MUX["Session runner: queued vs immediate"]
    RUN["runTurn"]
    MODEL["modelRouter.SelectModel"]
    INF["go-ai inference stream"]
    RD["Record route event"]
    EVT["turn events + SSE broadcast"]
    MSG["assistant/user messages"]
  end

  subgraph Persistence[SQLite persistence]
    S[sessions]
    T[turns]
    TE[turn_events]
    RE[routing_events]
    M[messages]
  end

  subgraph SSE[/SSE clients/stream/]
    RDEC["routing_decision"]
    RIN["routing_incoming"]
    DEL["agent_draft_delta / thinking_delta"]
    RESP["new_post / agent_response"]
  end

  WA --> SUB
  WPM --> SUB
  TC --> SUB
  HE --> SUB
  DI --> SUB

  SUB --> SP --> RR --> AR --> CL --> SESS
  SESS --> T

  SUB --> TURN --> MUX --> RUN
  RUN --> MODEL --> INF --> EVT --> DEL
  RUN --> MSG --> M
  TURN --> TE
  TURN --> T
  TURN --> RD --> RE
  RD --> EVT

  RE --> EVT --> SSE
  EVT --> SSE
  SSE --> RDEC
  SSE --> RIN
  SSE --> RESP

  SESS --> S
  MSG --> M
```

---

## Prompt routing sequence (with user prompt)

```mermaid
sequenceDiagram
  autonumber
  actor User
  participant UI as Web/TUI Client
  participant API as Web API Handler
  participant Engine as turn.Engine
  participant TR as RouteResolver
  participant SA as Session Allocator
  participant Store as SQLite Store
  participant SSE as SSE Stream
  participant AI as go-ai
  participant MR as Model Router

  User->>UI: Sends prompt
  UI->>API: POST /api/sessions/{session}/prompt
  API->>Engine: SubmitPromptRouted(RunInput)
  Engine->>Store: GetSession(source)
  Engine->>Engine: parseDirectedPrompt(prompt)
  Engine->>TR: ResolveRoute(InboundContext)
  TR-->>Engine: ResolvedRoute(agent, matched_by)
  Engine->>SA: ResolveOrCreateRouteSession(source, route, inbound)
  alt matching target session exists
    SA->>Store: ResolveSessionByAllocation(...)
    SA-->>Engine: target session
  else no target session yet
    SA->>Store: ResolveOrCreateSessionFromAllocation(...)
    SA->>Store: CreateSessionWithMetadata + relational identity + alias rows
    SA->>Store: Copy/fork prior history
    SA-->>Engine: new target session
  end

  alt same session as source
    Engine->>Engine: SubmitPrompt(RunInput)
  else routed to different session
    Engine->>Store: Add system message "↪ routed to @agent"
    Engine->>Engine: submitPeerRoutedPrompt(...)
  end

  Engine->>Store: CreateTurnWithStatus(target session)
  Engine->>Store: RecordRouteEvent(...)
  Engine->>Store: AppendTurnEvent(turn.submitted)
  Engine->>SSE: emit routing_decision (+ routing_incoming when target differs)

  Engine->>Engine: queue/run via session runner
  Engine->>Store: Append turn.started + user message payload
  Engine->>MR: SelectModel(prompt, history, agent model)
  MR-->>Engine: selectedModel, usedLightModel, score

  alt light model / stub model
    Engine->>Store: Add assistant bootstrap message
  else real provider
    Engine->>AI: CompleteWithBroadcast()
    AI-->>Engine: text_delta / thinking_delta
    AI-->>Engine: done + usage
  end

  Engine->>Store: Add assistant message + usage
  Engine->>Store: UpdateTurnStatus(completed)
  Engine->>Store: AppendTurnEvent(turn.completed)
  Engine->>SSE: new_post / agent_response
  SSE-->>UI: streaming + completion events
  UI-->>User: rendered assistant reply
```

---

## Peer message routing sequence (`/api/sessions/{session}/peer-message`)

```mermaid
sequenceDiagram
  autonumber
  actor User
  participant UI as Web Client
  participant API as Web API Handler
  participant Engine as turn.Engine
  participant TR as RouteResolver
  participant SA as Session Allocator
  participant Store as SQLite Store
  participant SSE as SSE Stream

  User->>UI: Sends manual peer message (target_agent_id + content)
  UI->>API: POST /api/sessions/{source}/peer-message
  API->>Engine: SubmitPeerMessage(..., targetAgentID, content, intent, model, parentTurnID)

  Engine->>Store: GetSession(source)
  Engine->>TR: ResolveRoute(inbound {mentioned=true, sender=source agent})
  TR-->>Engine: route target
  Engine->>SA: ResolveOrCreateRouteSession(source, route, inbound)

  alt target session exists
    SA->>Store: ResolveSessionByAllocation(...)
    SA-->>Engine: target session
  else create target session
    SA->>Store: ResolveOrCreateSessionFromAllocation(...)
    SA->>Store: CreateSessionWithMetadata + relational identity + aliases
    SA-->>Engine: new target session
  end

  Engine->>Store: Add system message "↪ routed to @target"
  Engine->>Engine: submitPeerRoutedPrompt(...)
  Engine->>Store: CreateTurnWithStatus(target)
  Engine->>Store: RecordRouteEvent(mode="peer-message")
  Engine->>SSE: emit routing_decision / routing_incoming

  Engine->>Engine: queue / run route turn
  Engine->>Store: Append turn.started + user+assistant messages
  Engine->>SSE: streaming + completion as normal turn events
  SSE-->>UI: response events
```

---

## Route-event introspection sequence (`GET /api/sessions/{session}/route-events`)

```mermaid
sequenceDiagram
  autonumber
  actor User
  participant UI as Client
  participant API as Web API Handler
  participant Store as SQLite Store

  User->>UI: Open route events panel / call introspection path
  UI->>API: GET /api/sessions/{session}/route-events
  API->>Store: ListRouteEvents(sessionID)
  Store-->>API: route-event list
  API-->>UI: { route_events: [...] }
  UI-->>User: show route history + routing reasons
```

If you need per-turn details, call normal introspection and inspect embedded metadata:
- `GET /api/sessions/{session}/introspect`
- includes `route_events` and `route_event_count`

---

## Parent-turn steering sequence (`parent_turn_id`)

Parent-turn threading allows UI callers to pass the parent relation when creating follow-up routed prompts.

```mermaid
sequenceDiagram
  autonumber
  actor User
  participant UI as Client
  participant API as Web API Handler
  participant Engine as turn.Engine
  participant Store as SQLite Store
  participant SSE as SSE Stream

  User->>UI: Clicks/constructs turn-threaded action
  UI->>API: POST /api/sessions/{session}/prompt with
  UI-->>API: { prompt, parent_turn_id }
  API->>Engine: SubmitPromptRouted(RunInput{ParentTurnID})
  Engine->>Engine: SubmitPrompt + metadata["parent_turn_id"]
  Engine->>Store: CreateTurnWithStatus(turn metadata includes parent_turn_id)
  Engine->>Store: RecordRouteEvent(route metadata + parent key)
  Engine->>Store: Append turn.submitted
  Engine->>Engine: runTurn() as usual (queue/busyness path)
  Engine->>SSE: stream events + completion
  SSE-->>UI: turn-thread aware UI can use metadata
  Store-->>Engine: (and UI) complete status updates
```

## API surface

- `GET /api/sessions/{session_id}/route-events`
  - returns `{ "route_events": [ ... ] }`
- `GET /api/sessions/{session_id}/introspect`
  - now includes:
    - `route_events`
    - `route_event_count`

## Config input

Runtime config consumed by routing:
- `agents`: known agents + per-agent model override
- `session.dimensions`: session namespace dimensions used by scope allocation
- `routing`: model-routing thresholds + route selector settings

These values are loaded from runtime configuration:
- `/.pi/settings.json`
- `config.RuntimeConfig`

## SSE events

The engine publishes non-fatal, best-effort routing observability events:
- `routing_decision`
- `routing_incoming`

Both events are sent as SSE payloads with routing context including:
- `chat_jid` (`gi:{session_id}`)
- `turn_id`
- `source_session`
- `target_session`
- `source_agent_id`
- `target_agent_id`
- `mode`
- `matched_by` (decision-only event)

## Route-event persistence (`routing_events`)

Columns tracked:
- `id`
- `turn_id` (nullable)
- `source_session_id`
- `target_session_id` (nullable)
- `source_agent_id`
- `target_agent_id`
- `mode` (`prompt`, `peer-message`)
- `matched_by`
- `routing_policy`
- `requested_agent_id`
- `metadata_json`
- `created_at`

## Direct/IPC ingress handoff

The first direct-processing slice now uses a normalized engine-facing envelope:

- `DirectInput`
- `DirectOrigin`
- `Engine.ProcessDirect(...)`
- `Engine.ProcessSystemDirect(...)`
- `Engine.ProcessInternalDirect(...)`

Current behavior:

- direct prompt ingress reuses `SubmitPromptRouted(...)`
- direct peer-message ingress reuses the routed peer-message path
- direct continue ingress reuses `ContinueSession(...)`
- direct ingress may target a session by explicit `SessionID` or by explicit `SessionKey`, with session-key resolution flowing through the same canonical store lookup path used elsewhere; if both are supplied they must agree, otherwise the request is rejected as ambiguous
- explicit system/internal entrypoints default the origin kind/role instead of relying on callers to pass those values ad hoc
- same-session direct/system follow-up input while a turn is active reuses the steering path instead of spawning a competing queued turn
- routed direct prompts still flow through the normal route/session-resolution path and carry ingress metadata forward onto the resulting target turn
- direct-origin turns stamp normalized ingress audit metadata onto turn metadata, persisted user-message payloads, and `turn.started` event payloads (`ingress_kind`, `ingress_source_kind`, `ingress_source_id`, `ingress_role`, `ingress_label`, `ingress_session_key` when provided)

This is intentionally a runtime entrypoint and envelope first, not yet a durable inbound work queue.

## Session allocation semantics used by routing

The current routing/session allocation stack assumes:

- routed/default allocation produces a canonical scope plus opaque session key and compat aliases
- `identity_links` canonicalization happens again at the store boundary, not only at route-preparation time
- explicit caller-supplied opaque allocation keys are preserved when they are already canonical/valid
- default/main-session flows can prefer a stored main session for the `(agent, channel, account)` tuple instead of relying only on recency
- explicit cross-channel continuation can reuse an existing session via `ContinueSessionID`, after which future allocations on that bound inbound surface may resolve back through `session_channel_bindings`

What routing does **not** assume yet:

- automatic cross-channel linking between unrelated inbound identities
- automatic outbound fan-out to every bound channel
- richer binding policy beyond explicit continuation plus later bound reuse

## Recovery / DB-first semantics

Routing never bypasses the normal queue/cancel/recovery path. Route decisions are persisted before execution starts so each routing action is durable even if inference is cancelled or fails.
