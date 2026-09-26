# Tool terminal provenance

A turn stopping does not identify when a particular tool stopped. The activity
projection must not invent a tool duration from a generic turn timestamp.

New provider-loop and bootstrap shell tool starts persist an `occurrence_id`.
Their matching `tool.finished` and `tool.failed` events carry the same ID.
Cancellation during execution now writes `tool.cancelled`; an abort from the
result hook writes `tool.aborted`. These two events use the runtime background
context because the execution context may already be cancelled. They precede the
existing turn finalization without changing queue, Stop, steering or TUI policy.

The activity projection matches turn, provider call ID and occurrence ID, taking
the first matching terminal event after the latest start. A later event for an old
occurrence cannot terminate a newer call, even if the provider reused its call ID.
Duplicate terminal events cannot replace the accepted terminal state or duration.
Legacy events without occurrence IDs retain their previous call-ID/sequence match;
strong reused-ID fencing is only available for newly written events.

The status adapter names `cancelled` and `aborted`, removes the running spinner and
uses persisted terminal duration. Reopening a session reads the same occurrence
and timing. A missing or failed terminal write remains `interrupted` with unknown
timing once the turn is inactive. Generic `turn.finished` does not fill that gap.
No synthetic completion event or duration is created on recovery.

## Scope and verification

This is a Shared40 prerequisite, not an expandable tool-output pane. It does not
map Shared40, change supplied components, restore historic output panes, or prove
focus/keyboard/reduced-motion acceptance for such a pane. Runtime tool-event topic
schemas and the provider's tool-result semantics are unchanged.

`make test-tool-activity` runs native store/web/turn tests with race detection three
times plus the timer helper. Tests cover completed, failed, cancellation with an
already-cancelled context, result-hook abort, old/reused/duplicate terminal events,
legacy fallback, and an injected terminal-write failure that must not block turn
cancellation or invent timing.

`make test-ux-tool-terminal` adds twelve browser cases across Chromium/WebKit and
three viewport sizes. Existing completion/failure/stale-read/session tests are
retained. The cancellation journey uses a real gated shell execution, captured
Stop, exact stored event identity, frozen duration, draft retention and reload.
All fixtures use disposable databases; no live chat or authentication writes.

Whole-product CI, four builds and exact-source deployment remain required.
