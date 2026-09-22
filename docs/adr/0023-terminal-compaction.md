# ADR-0023: Terminal compaction in the existing footer

## Status

Accepted — 2026-09-22. Live-tested at 60×18, 100×22 and 140×36.

## Commands and ownership

`/compact` invokes the shared native idle-only maintenance operation. `/compact info` retains the previous settings/counts diagnostics. Alt-C invokes the same operation without consuming editor text, cursor, undo/yank state, history or queued drafts. Busy/empty/ineligible sessions report a brief rejection and create no work.

Escape requests run-bound cancellation while compaction is active; `/cancel` also recognises active compaction. The editor's focused Escape handler delegates to the host before ordinary blur, because go-tui dispatches focused stop bindings before parent preemptive handlers. Other editor/menu Escape behaviour is retained. Running state is not cleared optimistically.

Lifecycle events invalidate the current session's native activity snapshot. The existing subscription generation rejects late A→B→A deliveries; snapshot reads identify the actual turn/occurrence and original timestamp. A terminal outcome has a new event sequence, so duplicate notifications do not extend notice lifetime. Cancelled/suppressed/failed events cannot render a false success block.

## Terminal footprint

Active progress is an inline `Compacting m:ss` segment in the existing stats row. Cancellation uses `Cancelling compact`. It truncates with the existing footer; no top bar, panel, border or active-only row is added. Outcomes and rejection details reuse the optional transient notice and expire after four seconds. After expiry, idle separators and row count match the original screen.

Compaction hook audit events no longer create generic live transcript blocks, including the misleading running “Hook invoked” block. They refresh native activity instead. Durable audit/history remains in the store. Existing stored summary rendering and cancellation system messages still appear in the transcript on history load; there is no new permanent UI.

State reads/advisory admission use short contexts to limit synchronous database waits. The engine's durable cancellation semantics and mutex waits remain authoritative; fully asynchronous handling under pathological storage contention is not established. No animation/redraw changes are included.

## Verification

`make test-tui-compaction` compiles an opt-in Go test fixture, installs a native before-compact hook gate and runs the production terminal loop in tmux. Native shell turns build history. The harness never seeds checkpoint outcomes or paints fake progress.

At each terminal size it checks:

- Empty and busy rejection without extra turns.
- Alt-C admission, inline elapsed status and unchanged separator positions.
- Draft and cursor preservation through active work, resize, Escape cancellation and retry.
- Cancellation writes no checkpoint; success commits one without sending the draft.
- `/compact info` diagnostics and explicit `/compact` operation.
- Process reopen retains the checkpoint and accepts the next native prompt.
- Expired notices leave zero additional idle rows.

Unit/race tests cover editor undo/yank/history state, native occurrence sequences, suppression/failure detail and session-generation isolation. Validation: full Go tests/vet, targeted TUI races repeated three times, new three-size compaction PTY suite, existing three-size session/model PTY suite, seven-file TUI Gherkin suite, TUI smoke, hook checks and 70/70 web functional tests.

The first live run found that focused Escape bypassed the parent binding; this was fixed and rerun. Fixture checks now wait for native outcome notice expiry instead of assuming the event arrives before a fixed sleep. A cursor assertion strips the visible cursor glyph before comparing text. These are explicit assertion-boundary corrections, not test retries.

## Scope

Browser mapping stays Classic 26/236 and shared 2/42 (210/40 unmapped). The previous 294/294 browser matrix is unchanged evidence, not a rerun in this terminal slice. No pi runtime or dependency change; rendering remains Go/go-tui. Broader reconnect/crash, pending media drafts and terminal queue controls still need acceptance work.
