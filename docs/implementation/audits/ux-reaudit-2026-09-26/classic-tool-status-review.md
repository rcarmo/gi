# Classic027: tool execution status

Gi's mounted `web/src/app.ts` renders `ToolActivity` only for fresh selected
session activity outside compaction. `web/src/gi-tool-activity.ts` labels tool
name/state/preview and increments elapsed display on a one-second interval
while running; completed and failed entries use terminal timing. The native
activity route supplies a tool-call and turn identity rather than relying on
the tool's display name alone.

`tests/ux/tool-activity.spec.mjs` uses a gated real native turn. It checks the
running preview, advancing elapsed label, reload identity, completed duration,
absence of a spinner, stable terminal duration, a later failed occurrence
with a different tool-call ID, a held stale activity response, and session
switch isolation. Focused `test-ux-steer` run: **6/6** across Chromium/WebKit
phone, tablet and desktop.

`make test-piclaw-output-oracle` also passed **6/6** using the installed
Piclaw 3.2.4 event translator and shipped browser assets alongside Gi's
adapter. Its synthetic tool start/update/end shows call arguments, a bounded
Output preview, and Waiting for model after completion. A separate thought
and draft preview survives that intra-turn transition, with no tool-result
conversation post; an idle event clears all three panes. A follow-up installed
translator and mounted-UI probe passed **6/6**: a successful tool shows
`Waiting for model...`, while a failed tool shows `Reviewing failed tool
result...`. The mounted Gi Output fixture separately passed **6/6** with the
same two titles, no tool-result post and no revived Output pane. Gi's status
adapter now distinguishes the failed state; a Go store regression verifies
that the failed occurrence remains visible while its turn is still running.
The fixture does not test an authoritative reload or a live failed-tool wait.
A separate installed lifecycle probe now
passes **6/6** synthetic Chromium/WebKit viewport cases including idle reload
and a fixture assistant post: WebKit unload presence/SSE errors match a minimal
control and are recorded only in explicit reload windows with a required
replacement connection. This does not validate production routing or a stored
backend reply.

`make test-piclaw-tool-output-window` exercised twelve controlled installed
translator cases, including empty/whitespace/trailing-newline text, a 120-line
window, 2/3/4-byte UTF-8 cutoffs, multiple text blocks and a reported
truncation flag. The installed translator preserves source whitespace and
trailing blank lines. Each orphaned continuation byte at a 12 KiB cutoff
renders U+FFFD. Gi's Go preview now preserves the text and matches the bounded
byte-window decoding; focused race tests passed three runs, the mounted
Output oracle passed **6/6**, and native tool lifecycle passed **12/12**
separately. Multiple blocks and result-supplied truncation metadata are not
Go-string acceptance.

`make test-piclaw-concurrent-tools` passed **6/6** installed translator/UI
fixtures. With two synthetic overlapping calls, completion of the first
emits a status for the second, retaining its Output and carrying the first
as `last_completed_tool`. Waiting for model appears only after the second
ends. Gi's persisted `latestToolActivity` reads the most recent `tool.started`
occurrence; an earlier unfinished call is not represented after a later call
starts. The engine currently executes tool calls in order, so no native
concurrent-call acceptance follows from the fixture.

This covers the visible status path and stale-response guard in disposable Gi.
Full tool-pane lifecycle reconstruction, concurrent-call behaviour in Gi,
reduced-motion presentation, production Piclaw routing and deployed Gi remain
unverified.
Classic027 stays unmapped. The separate WIP tool-terminal provenance branch
has no CI/deployment credit from this review. No production code or frozen
contract changed.
