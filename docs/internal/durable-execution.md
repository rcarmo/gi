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
