# Gi TUI single-line status semantics

Status: active contract for Pi/PiClaw UX convergence.

## Rule

Gi has exactly one physical notification/status line in the steady-state TUI: the final row of the Pi-like bottom band.

Transient runtime notices must update that line instead of adding transcript rows unless the event is durable conversation/history.

## Bottom status content

The final row is composed as:

```text
<left status/notification>                                      <model> • <thinking>
```

Left side priority:

1. Current transient status if present and not just the default model label, e.g. `Running: read`, `Tool skipped: shell`, `inbound work queued (ipc) [queued]`.
2. Otherwise compact counters: `m<messages>/t<turns>`.
3. Add `q<queued>/s<steering>` only when either value is non-zero.

Right side:

- current model;
- thinking level when known;
- never wraps; truncates if required.

## Transcript vs status

Prefer status line only for:

- tool started/finished/skipped;
- hook invocation/modify/respond/deny notices;
- inbound work enqueue/retry/requeue/discard/dispatcher lease/drain notices;
- short-lived running/queued/waiting notifications.

Keep transcript lines for:

- user messages;
- assistant responses;
- durable system terminal outcomes such as cancellation/failure summaries;
- compaction summaries;
- routing/session switch messages initiated by user commands;
- tool failures when the failure text is useful history.

## No wrapping

Status text must be truncated to preserve a single physical row. If future token/context/cost metrics are added, they must fit into this row or be omitted.
