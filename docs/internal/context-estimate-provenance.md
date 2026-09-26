# Estimated compaction history and measured context

The active compaction tooltip now labels the native history count as
`estimated history: N tokens`. The latest measured provider request retains its
own count and provenance text. No estimate replaces missing measured usage, and
no new idle control or numerical meter is added.

`internal/compaction/runtime.go` already emits `tokens_before` with
`tokens_source: "estimate"`. The activity snapshot carries that native payload.
The Gi adapter preserves it only for the matching active turn, and formats a
label only for the exact provenance and a nonnegative safe integer. Missing,
foreign, inactive, fractional or malformed values add no estimate label. Zero
is shown only when explicitly supplied as a valid native estimate, never inferred
from absence. Backend arithmetic, compaction admission and the measured-only
`context_usage` contract are unchanged.

## Web Shared35 evidence

The canonical `@shared-35` journey in `session-thinking.spec.mjs` covers all clauses:

- Unsupported models have no thinking selector; the supported model advertises
  low/high, accepts explicit low, and the actual provider request uses low.
- Initial unknown tokens/percentage remain null in native state and unavailable
  in the UI.
- Native compaction starts unavailable with a reason/token/policy; clicks do not
  mutate. After native history exists it advertises availability, and the UI
  submits exactly that token and receives 202. It disables repeated activation.
- The native active event contains estimate provenance/count. Tooltip and
  accessible text label estimated history separately from measured provider usage.
- Removing only provenance from a read-only activity response and reloading omits
  the estimate label while preserving measured usage, active state and the draft.
- Completion removes the active estimate label without changing provider usage,
  draft text or attachments.

Independent final clause review accepted this web-scoped evidence. Exactly one
test carries the canonical tag; supporting tests do not double-count it. Source
mapping coverage becomes 32/42 shared cases and remains 101/236 Classic IDs. The
focused report has one shared pass, 31 mapped-but-not-run and ten unmapped cases;
it is not a full parity run. Reports reject incomplete/duplicate-project evidence
and cannot borrow Classic/context or Shared36 coverage for Shared35.

## Verification

- `make test-shared-capability-evidence`: 36 thinking/capability browser cases,
  eight report/ledger tests, 1963 assertions; six-project Shared35 pass report.
- `make test-ux-compaction`: 102 cases, including native event values, unchanged
  measured context, completion, legacy missing-provenance suppression and ownership.
- `make test-ux-context-meter`: 18 cases retaining exact measured/unknown labels.
- `make test-context-control-helpers`: six helpers, 62 assertions, including bad
  provenance/counts and foreign/inactive ownership.
- `make test vet bun-checks`: passed; full isolated functional suite 139 passes,
  11 existing skips. Supplied components and frozen feature files are unchanged.

One command-harness interruption stopped an initial compaction run; the separate
rerun passed. Thinking race repetition also exposed a fixture lifetime bug: its
shared-memory store could outlive cleanup and collide with the next test. The
fixture now uses a private file-backed store and waits under the runner lock for
cleanup to finish. No runtime behaviour, browser assertion or timeout changed for
that correction. Earlier review timeouts did not count as approval.

Initial CI `36258125176` failed the unrelated Shared36 fixture: it selected any
idle frame, so the earlier warm-up could satisfy the wait before the cancelled
turn's frame arrived. Exact session/turn matching exposed a second readiness race;
the native session-scoped `connected` subscription barrier now precedes Stop.
Eighteen repeated cases and the full 66-case reconnect suite passed, with all
replay assertions unchanged and no timeout increases. Scoped review approved;
frame identities are attached on each run for diagnosis. Runtime code is unchanged
by this fixture correction; the failed CI does not approve deployment.

Evidence is retained in `/workspace/tmp/gi-compaction-estimates`. Whole-product CI
and exact-source deployment are still required. Physical-device, Visual-skin,
exact-pixel and TUI acceptance are separate. Local-only TUI WIP `2a87a79` stays
isolated; its generic cancellation/thinking behaviours receive no new parity credit.
