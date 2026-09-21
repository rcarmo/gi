# ADR-0010: Isolate terminal selections without adding idle chrome

## Status

Accepted — 2026-09-22

## Context

The terminal used long-lived event channels so go-tui watchers survived a session switch. Unsubscribing left already-buffered events in those channels. Their handlers had no selection-generation check. Switching also retained the previous editor/history state while clearing only part of the streamed response state.

The web session work established generation-based ownership and per-session drafts. Terminal adaptation must preserve the existing transcript, editor, separators and footer.

## Decision

Wrap each forwarded engine/topic event with the session ID and generation captured when its subscription was created. Keep the watcher channels stable, increment generation on rebinding, and discard deliveries with a different scope. A return from B to A does not accept events buffered during the earlier visit to A. Validate explicit session IDs in payloads as an additional check.

Give each subscription pair a cancellation context. Forwarders select on cancellation while both receiving and sending; a full downstream channel cannot strand an old forwarder. Cleanup cancels before unsubscribing.

Capture session/model inputs before launching asynchronous prompt, follow-up or peer-message work. Apply resulting UI updates only if the original scope still owns the editor when the queued callback runs. Native work continues in its original session; switching does not cancel the turn or redirect its result.

Cache terminal editor state per session: text, rune cursor, single-step undo, yank buffer, command-history/search position and the existing local queued-draft list. Keep this state separate from the streamed assistant draft. Validate the target before touching the origin. On success, reload persisted transcript/model state and clear origin-owned tool/thought blocks, extension slots and usage values. Missing model metadata falls back to the initial terminal configuration, not the preceding session.

A session-bound extension question is cancelled on switch. The editor draft that preceded that question is restored before saving the origin. Questions and their partial answers do not follow the new selection.

Extend the existing shared selector, without adding another pane:

- `Alt-S` opens the session selector without replacing the unsent input. `/sessions` remains the command fallback.
- At most six result rows appear above the editor, reduced to fit the existing editor/footer and transcript reservation.
- Title and search use two additional temporary rows. No surrounding box, top header or idle picker row is added.
- Up/Down wrap; page/Home/End navigation remains bounded. Enter selects; Escape restores editor focus and draft.
- Session filtering uses full IDs to avoid collisions after compact-ID truncation. Display rows are single-line, UTF-8-safe and terminal-cell bounded.
- Muted help and a single accented selection marker follow the pi-tui SelectList interaction pattern. Gi retains go-tui and its existing terminal palette.

## Limits

Draft caches last only for the process lifetime. `/attach` and `/paste-image` still store media immediately or submit it with their command prompt; the TUI has no pending-media composer collection. Typed attachment paths survive as draft text, but this slice does not implement a media-ref draft queue.

The local queued-draft list remains distinct from durable queue actions. Restoring a local queued draft does not prove backend dequeue/retry semantics. Session mutation submenus, archived filtering controls and durable browser drafts need separate acceptance work.

## Verification

`internal/tui/session_state_test.go` covers editor/history/undo round trips, failed-switch retention, A→B→A event rejection, a real topic-bus buffered event, late UI completion rejection, blocked-forwarder cancellation, model defaults, extension-question cancellation, assistant/editor draft separation, selector wrapping and Unicode widths.

`make test-tui-sessions` builds Gi and drives a real tmux session with isolated database/settings. It creates sessions through native commands, then verifies 60×18, 100×22 and 140×36: at most six results, exact before/after cancel screenshots, two existing separators, visible editor, both restored drafts, zero submitted turns and selected-row visibility during resize. Artifacts live in `test-results/tui-sessions/`.

Existing Go/vet, TUI smoke/Gherkin and browser suites remain required. Terminal evidence does not increase the number of mapped Classic browser scenarios.
