# ADR-0029: Keep terminal editing independent of transcript navigation

## Status

Accepted — 2026-09-22. Native terminal verification at 60×18, 100×22 and 140×36. No additional frozen browser mapping.

## Reading policy

Typing, programmatic editor updates and ordinary prompt submission now retain the transcript's existing follow mode. `onInputChanged` and both running/idle submit paths scroll only when `stickToBottom` is already true. PageUp or mouse navigation away from the newest edge disables following; explicit navigation back to the bottom resumes it.

The previous editor callback enabled following on every change, including `SetText("")` during submission. This moved a history reader even when subsequent native progress handlers correctly respected `stickToBottom`.

The change adds no widget, selector, status, keyboard binding or idle row. It uses the existing transcript/editor interaction and scroll controls. The captured editor clears once; progress and completion leave newer input and its cursor intact.

## Evidence

`make test-tui-reading` runs the production terminal in tmux at all three sizes with its status bar disabled. History comes from 22 real native shell turns. A shell-process gate delays the next provider completion while the user scrolls to history and types another draft. The test verifies:

- the same history text remains at the same physical row during editing and after completion;
- a cursor insertion still lands at the selected position;
- resize to 80×24 and back preserves the anchor, draft, cursor and separator rows;
- submitting while reading does not move the reader or duplicate the user message;
- PageDown to the newest edge resumes following for another native turn;
- newer unsent text never becomes a stored message;
- idle separator positions remain identical before and after the operation.

The test fails against the previous implementation with `typing moved the history reader`. Unit tests cover editor follow-mode preservation and current-session progress/completion; existing generation tests cover stale A→B→A callbacks. All terminal package race tests pass three repeated runs.

Validation: full Go tests/vet, hook checks, 29/29 browser helpers, three-size reading/session/model/compaction acceptance, terminal smoke/Gherkin and 70/70 functional browser tests. Web code and frozen features are unchanged; the preceding browser matrix remains 384/384 at commit `4071c01` with 34/236 Classic and 2/42 shared mapped. This slice did not rerun that complete matrix.

Fresh screenshots show actual XTerm windows at the verified sizes and browser desktop/search/phone views. Captures are under `/workspace/tmp/gi-ui-captures/september22/` and were attached to the conversation. Terminal screen text and test evidence are under `test-results/tui-reading/`.

## Remaining terminal work

The gate delays native provider completion, not the return of `SubmitPromptRouted`. Delayed cross-session routed acceptance still switches to its destination when its captured origin generation is current; that path needs a separate user-intent policy and native acceptance test. Error/uncertain-delivery recovery, persistent pending-media references, queue controls and durable message-ID reconciliation also remain open.

The reader anchor verified here covers append-only history within the existing scrollback limit and a resize round trip using short lines. Wrapped-content reflow, edits/removals above the viewport and scrollback eviction need stronger block/line anchoring. No new terminal parity credit is assigned to those paths. The read-only review delegate could not launch an approved executable model and supplied no review evidence.
