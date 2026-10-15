# PicoClaw parity status

## Status

Private/internal audit note for the Gi runtime refactor. This is not a public compatibility promise; it records which PicoClaw-inspired runtime ideas Gi has ported, adapted, or intentionally rejected while keeping SQLite/WAL as canonical state.

## Ported/adapted

- **Active-turn coordination:** Gi uses store-backed active-turn claims, durable turn phases, queued-turn handoff, cancellation-safe cleanup, and stale-claim recovery instead of goroutine-local busy flags.
- **Session identity:** Gi uses relational session identity, aliases, canonical allocation keys, main-session preference, channel bindings, and binding-aware continuation instead of runtime scans over `sessions.scope_json`.
- **Steering:** same-session input becomes durable steering rows, observed at explicit checkpoints, with continuation staging and skipped-tool persistence.
- **Hooks:** Gi has a canonical hook taxonomy, JSON-safe request/response DTOs, explicit actions, durable `hook_invocations`, mounted process hooks, and provider request/response interception via `internal/inference`.
- **Sub-turns:** parent/child linkage, depth/concurrency controls, restricted tool inheritance, async/sync delivery, orphan handling, cancellation propagation, and topic publication are implemented with durable audit rows.
- **Topics/events:** canonical bounded firehose topic bus with monotonic sequences, session/agent filtering, script publish/subscribe/read/unsubscribe APIs for JS and Joker, topic SSE, TUI topic-native rendering, connectivity bridging, and aggregate `runtime` topic publication.
- **Direct/IPC ingress:** `DirectInput` / `DirectOrigin`, explicit system/internal entrypoints, durable inbound queue, retry/backoff state, dispatcher lease ownership, and runtime topics reuse normal routing/session/steering/turn paths.
- **Multi-channel shape:** channel/account bindings, explicit cross-channel continuation, bound reuse, source channel/account metadata, and explicit no-auto-link/no-auto-fanout policy are documented.

## Intentionally different from PicoClaw

- **Storage:** Gi keeps SQLite/WAL as canonical state. It does not adopt JSONL session storage.
- **Topics:** the topic bus is a bounded live firehose with gap detection, not a persistent replay log. Replay/state reconstruction comes from SQLite APIs.
- **Multi-channel policy:** Gi does not automatically merge unrelated inbound identities and does not fan out responses to every bound channel without an explicit future policy layer.
- **Connectivity:** the connectivity registry remains the request/response route owner; the topic bus observes/bridges connectivity events rather than replacing route dispatch.

## Remaining work

- TUI stack evaluation against PicoClaw/Pi launcher ergonomics remains open and should not block runtime semantics.
- Broader full-row topic mirroring is still a future enhancement; current canonical topics cover runtime-critical lifecycle slices rather than every stored audit row.
- Dependency refresh / Joker compiled hook caching remains in the end-of-plan maintenance backlog.
