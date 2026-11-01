# ADR 0062: provider partial and failed-response receipts

Status: accepted
Date: 2026-10-31
Issue: #87

## Storage and ownership

Gi retains one bounded assistant snapshot per locally admitted provider request. The snapshot comes from structured provider events. Adapters without structured partials reconstruct a bounded prefix from provider deltas; browser draft text is never an input.

The request occurrence determines the message ID. Checkpoints replace that row at most once per 75 ms. The first useful snapshot commits immediately. After the provider channel closes, the owner forces the final observed snapshot and fences further callbacks. The checkpoint writer runs synchronously inside the provider observer, so it creates no additional goroutine to join. Native stream collection drains through channel closure before returning. PR #113 adds engine-wide shutdown joins, including providers, tools, hooks, children and closing leases. General checkpoint continuation remains in #86/#92.

Snapshots detach strings, arguments, signatures, diagnostics, usage and deferred metadata. Limits are 32 KiB of strings, 64 content blocks, eight diagnostics, 256 nested values and eight nesting levels. Oversized images and unsupported values are omitted with truncation metadata. Storage additionally rejects checkpoints over 512 KiB of encoded metadata or 64 KiB of display content. Checkpoints do not append raw deltas to the turn journal.

A failed terminal retains its structured provider message, error text, event stop reason, signatures, diagnostics and received usage. Unencodable provider values use a bounded structured fallback with explicit omission diagnostics; invalid usage remains uncertain. Failure handling retains the bounded committed prefix separately as `provider_partial`, even when the terminal message exists but omits streamed content or usage. Conversation messages and search remove both raw snapshots. History text comes from those structured snapshots; missing or shorter terminal text cannot erase the committed prefix. A successful receipt atomically replaces the same row with the complete response and its successor boundary. Failed terminal commits roll back without deleting the checkpoint.

## History, context and accounting

Partial and failed rows remain visible in raw history and conversation reads. Thinking blocks remain available to the UI. Their `adopted=false` flag excludes them from model context, including malformed provider errors carrying a user role. This projection does not rewrite the stored provider message or authorise tool execution.

An unsettled request remains an unknown accounting occurrence. On failure, known partial counters contribute once while the receipt remains explicitly uncertain. `SealProviderPartial` supports conversion after the coordinator has joined the owner or established its death. It does not acquire/release claims, run tools or resend requests. Checkpoint-based recovery is tracked separately in #92. No local receipt proves remote exactly-once billing.

## Initial verification (2026-10-31)

`make test-provider-partials` passed three repetitions with race detection in each affected package. Makefile runs passed sequentially for `internal/store`, `internal/turn`, `internal/inference` and `internal/web`, followed by `make vet RACE=` and `git diff --check`. Focused regressions cover structured text/thinking/tool prefixes, error and abort retention, incomplete native/compatibility streams, detached buffers, bounded/coalesced updates, rejected writes, terminal rollback, context exclusion and received usage. A subprocess is killed after committing a checkpoint; the parent waits for its death, reopens SQLite, converts the committed partial and verifies known counters plus uncertainty without replay.

The full turn suite caught a sortable-ID regression in steering order. Request-derived message IDs now retain the existing `msg_<nanoseconds>` shape; the unchanged steering regressions and full suite pass. These were development checks.

## Follow-up verification (2026-11-01)

The [terminal-retention audit](../implementation/audits/provider-partial-terminal-retention.md) records the empty-terminal regressions, passing race checks, 2,728 Go tests, Chromium/WebKit acceptance (149 passed, 13 skipped), and CPU/allocation analysis with prompt disposal. Valid terminal usage takes precedence; received-only usage remains explicitly uncertain. No release, deployment or live-provider billing validation was performed.
