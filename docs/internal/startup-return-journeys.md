# Startup and first Return

Gi now focuses the composer after successful startup and new-chat creation, and
offers an explicit retry when initial chat loading fails. The fix is in
`web/src/app.ts`; supplied Piclaw components are unchanged.

## Reproduced failures

With an empty database and fresh browser storage, the composer appeared without
focus. Creating a new chat also left its textarea unfocused: the host keys the
composer by session, so the supplied component's callback targeted its removed
textarea. Corrected user-level focus tests failed in all six Chromium/WebKit
viewport projects before the host change.

A failed initial session POST left the screen at `Loading…` indefinitely. Six
project executions reproduced the missing error/retry path. Early test-authoring
failures involving locators, response field names and request base URLs are
excluded from this product evidence.

## Host behaviour

Successful startup and a current-view new-chat result each request one focus
operation after the new composer commits. The operation checks that the selected
session still matches, focus is on the document body and no modal is open. It
does not run for polling or ordinary session switches. A delayed new-chat result
cannot move focus out of Settings.

Bootstrap failures show a fixed error and **Retry opening chat**. Retry is
user-driven. If session creation already committed, the next attempt finds the
stored or existing main session rather than creating another chat. Effect cleanup
fences stale success, error and draft-recovery callbacks. No server request is
automatically replayed on failure.

The supplied composer still clears its text while admission is pending and
restores it when admission fails. This slice does not change that behaviour.

## Required browser gate

`make test-ux-journey` runs `tests/ux/startup-journey.spec.mjs` against one disposable
native process and empty database per case. Browser contexts are fresh. The
fixture neither creates sessions via API nor seeds `gi_session_id`; the page's
normal startup creates the session. Only inference uses a deterministic local
provider; production HTTP, SSE, turn admission and SQLite paths are exercised.

Seven journeys run in Chromium and WebKit at 390×844, 820×1180 and 1440×900:

- Clean boot focuses the composer; Return creates exactly one captured turn,
  with exact persisted user content and rendered user/assistant posts. Reload
  preserves the selected chat without another prompt or session.
- A rejected first prompt restores its draft. An explicit retry held before
  admission creates one turn after release; extra empty Returns send nothing.
- A held, failed initial session request hides the composer until explicit retry.
- A committed session whose response is lost is recovered by lookup without a
  second session POST.
- Malformed runtime configuration after session creation can be retried without
  another chat.
- New-chat failure preserves the parent draft and selection; a later delayed
  success leaves an open Settings dialog focused.
- New chat focuses an empty composer. Shift+Enter inserts exactly one newline
  without admission, Return sends once to the child, and selecting the parent
  restores its unsent draft.

The suite produces `test-results/ux-parity/journey-results.json`. Its dedicated
CI job installs both browsers, uploads evidence and gates all Linux/macOS builds
alongside native tests and passkey browser coverage. The suite has no frozen
scenario tags and adds no Classic/shared mappings.

## Verification and limits

The completed local runs passed 42 journey cases, 126 session cases, 216 Settings
cases, 168 draft cases, 110 functional cases (five existing skips), full Go/vet
checks and 107 helper tests with 2,975 assertions. A combined regression run
expired at the command timeout and another was interrupted; only the completed
split runs count. Focused review identified a stale draft-load callback, which
was guarded and re-reviewed without further blockers.

This covers the reproduced startup/new-chat focus and loading-failure paths. It
does not establish current Piclaw visual equivalence, conversation-level Return
or Escape semantics, physical mobile-keyboard behaviour, picker geometry, direct
slash-menu ownership or complete workspace-tab transitions. The
[dated audit](ux-test-audit-2026-09-24.md) retains those findings. No terminal
layout or interaction changes are made by this browser slice.
