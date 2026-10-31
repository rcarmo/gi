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
registrations default `ReplaySafe` to false. This metadata alone does not enable
automatic replay; interrupted work remains held until checkpoint recovery can
establish that resumption is safe.

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
