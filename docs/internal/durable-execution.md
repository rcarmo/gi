# Durable execution boundaries

Gi commits effect intent before entering a provider or tool executor. Admission
atomically checks the active running-turn claim, changes its phase and appends
its start event. Cancellation or a rejected SQLite write prevents host entry.
Start notifications follow the commit.

Each provider attempt records the selected model, native context, thinking and
transport. After the final payload hook, a separate prepared-request boundary
pins the transport payload under the same request ID. Credentials such as
Authorization, API keys, cookies and access/refresh tokens are excluded from
stored provider intent. Intent JSON is bounded to 64 MiB; oversized or
unserialisable intent fails admission.

Tool start events retain the original model call separately from the final
hook-prepared call, current schema and replay-safety declaration. Go tool
registrations default `ReplaySafe` to false. A committed tool-batch checkpoint
pins the provider response, call order and prepared arguments. Recovery adopts
settled receipts without repeating their effects. A started call can execute
again only when both saved and current registrations permit replay and current
tool permissions allow it. As in go-ai v1.4.0's pi-durable execute phase, replay
uses the saved prepared arguments directly; it skips preparation and replacement
schema validation.

Interrupted preparation or unsafe execution produces an explicit interrupted
error result. Committed SDK output remains part of that error, with no invented
completion time. Gi records result-hook attempts before entry and never repeats
an attempted result hook. Each physical execution retains separate received usage
and interruption accounting. Unknown stages, malformed receipts and legacy
frontiers retain conservative whole-turn holds. Explicit retry requeues a proven
native root checkpoint on its original turn and resets only its crash budget;
it never resubmits the prompt. Parent-linked or terminal frontiers stay held.
For a proven native tool batch, `Close` reserves a new writer incarnation and
cancels the executor without publishing a terminal. Bounded SDK output and usage
recorders remain bound to that physical host while it joins. Only after all owned
work and the closing heartbeat join does the controller publish a queued tool
checkpoint. Recorder or commit failures return an error and leave the reservation
unpublished. Graceful shutdown neither spends nor resets the crash budget.

When no continuation is proven, `Close` reserves an `interruption-closing`
writer and joins the same owned hosts and recorders before parking a review hold.
Received SDK output and usage remain in their diagnostic snapshots. A returned
provider outcome settles as an uncertain error-boundary receipt; its terminal
message and committed prefix stay separate. The hold creates no root terminal
receipt or agent-end attempt. If shutdown itself is
interrupted, that worker identity forces a hold even for setup or pre-request
compaction; automatic prompt replay is disabled. An explicit user cancellation
or previously staged outcome keeps its own settlement path. Failed hold
publication rolls back and returns the shared Close error.

Unlaunched checkpoints support terminal-only cancellation through the engine:
writer fencing and the cancelled decision commit together, and interrupted
cleanup cannot dispatch the checkpoint. Outstanding calls receive error results,
including never-admitted slots, without replaying preparation or execution.
A proven recovery-held checkpoint also supports explicit skip through failed
terminal cleanup. Its original wire receipts, usage and skipped resolution survive
recovery, and required parent joins wait for its receipt and final hook.
Parent-linked retry remains conservatively blocked; it cannot reverse consumed
child outcomes. See [held-turn retry and skip](held-turn-retry.md).

Executed, rejected, denied, skipped and hook-supplied tool results commit their
terminal event and replay message in one transaction. Model continuation and
completion notifications require that commit. Tool-owned details and harness
output-policy diagnostics remain separate.

## Provider terminals

Gi validates assistant role, supported content, tool-call identity and arguments,
stop reason and non-negative finite usage before adopting a result or executing
tools. `stop` and `length` cannot carry tool calls. Empty `toolUse` is a final
answer. `pending` is invalid as a terminal; deferred provider execution is
explicitly unsupported and fails the turn without dispatching tools.

Streaming adapters require exactly one terminal, reject contradictory stop
metadata and drain the provider channel through closure before returning. The
OpenCode-compatible SSE path requires a final finish reason; `[DONE]` alone
cannot turn an incomplete response into success.

## Receipt settlement and accounting

Provider response adoption commits the assistant row, context measurement,
occurrence-keyed usage and successor checkpoint in one SQLite transaction.
Repeated settlement preserves the first receipt and counts its usage once.
Terminal settlement commits status, finish time, sub-turn/failure projections,
system message, accounting/finish events and idle session state together. Runtime
publication follows a successful commit.

Tool runtime `AddUsage` reports counters from nested model calls. `UnknownUsage`
marks a remote call whose counters were unavailable. Both affect the enclosing
tool receipt; known counters survive later tool/provider failure or cancellation.
`unknown_receipts` also counts interrupted admitted work without a receipt and
provider results with absent or all-zero counters. It does not estimate missing
billing. Raw late responses retain their remote terminal reason; `adopted: false`
excludes them from model context when cancellation wins.

See [ADR 0061](../adr/0061-receipt-accounting-and-terminal-settlement.md) for the
transaction boundaries and historical-accounting limitation.

Iteration-end completion hooks run before settlement so Escape/steering can still
win. Their completion proposal carries `settled: false`. The agent-end hook reads
the stored terminal status; failed settlement reports `reason: "unsettled"`.
