# Codemode web payload shape

F3's renderer consumes Gi runtime envelopes for live nested rows and saved
result details for settled/reloaded rows. This shape records the implementation
at `2a5e247`; no web renderer or browser acceptance has been added.

| Stage | Fields | Source and handling |
|---|---|---|
| Parent call | tool=codemode, tool_call_id, turn_id, arguments.code | top-level executeToolCallsPhase; highlighted source, escaped as text |
| Nested start | tool, tool_call_id=`<parent>/<sequence>`, parent_tool_call_id, arguments, turn_id | runNestedTool; attach only to matching parent/turn/session |
| Nested finish/fail | same identity; optional duration_ms; error on failure | measured executor duration from N1; missing value stays unknown |
| Parent terminal | details.calls, optional details.fullOutputPath, output, duration_ms | toolExtras.withDetails on runtime.tool terminal event; replace live rows with settled snapshot |
| Saved tool_result | payload.tool_name, tool_call_id, turn_id, occurrence_id, duration_ms, details; content carries result text | Store message, real reload acceptance required |
| Settled calls | id, name, args string, status, durationMs number; optional error/cost | codemodeCallLog snapshot; status ok/error/cancelled; still-running calls become cancelled |
| Full output | vfs://codemode-output/<session>/<file>.txt | read through authenticated existing output/workspace contract; never concatenate into arbitrary filesystem URL |

`duration_ms` is an integral top-level/nested runtime measurement; settled
`durationMs` is a floating-point millisecond call-log measurement. They use
different boundaries: executor timing excludes surrounding hook/delivery work;
call-log timing includes the logged invocation. Do not replace one with the
other or require equality across these boundaries. Missing and zero have
separate semantics; live running duration is not a settled Took.

`internal/web/sse.go` forwards legacy subscribed broadcasts unchanged by payload
type; it separately forwards topic envelopes through the topic SSE endpoint.
`PublishRuntimeToolEvent` publishes on `runtime.tool`. The frontend must verify
which subscription actually carries nested details; it cannot infer that the
legacy stream exposes every runtime envelope. Mounted mocks are insufficient.
No new bridge/endpoint is justified until a live Gi run demonstrates a missing
transport field. Avoid subscribing both paths in a way that duplicates rows.

Parent result text begins with Script completed/failed, Wall time and Output.
The existing TUI strips that wrapper; web should preserve output literal text
while separating script/status/output. A failed script does not undo earlier
tool calls. Missing details degrade to generic output. Reload must work without
an active server/tool process and use persisted details only. Test truncation,
interruption, mismatched IDs/session switching, stale completions, hostile
markup/links and two parents with reused provider IDs.

Acceptance matrix: actual Gi live run, parent failure/cancellation, final snapshot
replacement, authenticated truncation lookup, offline reload and applicable
Chromium/WebKit desktop/tablet/phone rendering. Owner must publish upstream
fixtures-vibes changes before Gi adoption and independent consumer acceptance.
