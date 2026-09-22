# ADR-0011: Browser-local drafts and captured-send recovery

## Status

Accepted — 2026-09-22

## Context

The browser stored drafts in a page-local Map. Reload lost text and attachments. Failed background sends replaced newly typed text and references; completion after a session switch could not restore the unmounted origin. The media upload adapter returned null.

Frozen cases `@ux-compose-001`, `002`, `003` and `006` require captured submissions, immediate clearing, merged recovery, empty-send rejection and an immutable upload destination. The separate shared queue-return contract also needs durable recovery, but its queue DELETE/retry requirements are outside this slice.

## Decision

Use IndexedDB database `gi-session-drafts`, store `drafts`, keyed by native session ID. Each record contains editable text/media/file/message references, pending captured sends and the last recovery error. Records are browser-profile-local. No server copy, cross-device synchronisation or encryption is added.

Persist attachment bytes as ArrayBuffers with name/type/last-modified metadata and reconstruct File objects on load. WebKit rejected File values during commit despite accepting their structured clone; explicit byte storage passed real reload/upload tests. The 10 MiB per-file cap remains in selection, upload and backend handling. Encoded immutable files are cached in a WeakMap to avoid rereading their bytes on every keystroke.

Serialize writes in one page. Ordinary edits update memory immediately and enqueue persistence. Storage errors are visible; the editor remains usable. Draft persistence is asynchronous, so a browser/process crash before the transaction commits can lose the last edit. Concurrent editing of one session in multiple tabs is not reconciled: the last committed record wins.

Submission captures its origin session, mode, text, files and references synchronously. Before network I/O, write a recovery record and the cleared displayed draft. If this write fails, do not send; merge the capture back into the origin draft. On successful acknowledgement, remove only that capture, preserving subsequent edits. A later UI refresh or cleanup-storage failure must never convert an acknowledged send into an automatic retry.

Failed sends merge captured text ahead of newer origin text and deduplicate captured/current references and files. The host repository owns recovery across unmounts and A→B→A revisits. A different selected chat keeps its editor and focus. Errors after dispatch that lack an HTTP acknowledgement report unknown delivery, because the server may have accepted the prompt.

On reload, pending captures merge ahead of the current draft in submission order and show an uncertainty warning. Recovery never automatically submits. Users must check the timeline before resending. If acknowledgement cleanup could not be persisted, recovery may include text already delivered; its warning states that limitation.

Upload uses `/api/sessions/{capturedID}/media` and the subsequent prompt carries native media references. File and message pills use the same draft repository. A small workspace adapter change wraps the native tree as `{root}` so native file selection can attach references; deep lazy loading and full workspace parity remain open. Timestamp links reserve space for post-action buttons so message-reference clicks remain reachable.

## Evidence

- `tests/ux/support/drafts.test.ts`: merge/deduplication, origin ownership, persisted recovery ordering, acknowledgement isolation, storage-write failure and repeat-reload behaviour.
- `tests/ux/drafts.spec.mjs`: real file/message references, reload and session switches, byte-for-byte restored attachment upload, delayed upload response, failed origin send with newer text/media/refs, A→B→A recovery, empty submission, unknown-delivery reload and storage/acknowledgement-cleanup faults.
- The six-project Chromium/WebKit × phone/tablet/desktop matrix passes 102 executions: nine mapped Classic IDs plus eight Gi regressions. The existing functional suite passes 70/70. Go tests/vet, Bun hook checks and 14 source/helper tests pass.

## Terminal adaptation

The terminal session cache from ADR-0010 remains process-local. A durable terminal implementation should reuse these ownership and recovery rules with a local native store, not IndexedDB. Pending media refs need explicit staging and per-session recovery; durable queue return must persist before deleting a backend item. Show failure or unknown-delivery notices in the existing footer/on-demand queue view. No extra idle rows, persistent recovery panel or sidebar are required. These terminal additions are not implemented here.
