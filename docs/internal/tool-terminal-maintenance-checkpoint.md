# Tool terminal maintenance checkpoint

Paused at the user's full UX re-audit request, then checkpointed for maintenance.
This branch is not approved for deployment or full Shared40 acceptance.

Owned source: occurrence-bound tool starts/terminals, explicit cancelled/aborted
persistence, status labels and native/browser tests. Runtime remains deployed
05e287f; Shared38 mapping 8196355 and docs5dc6bf7 were already complete.

Checks passed before checkpoint: `make test-tool-activity test vet bun-checks`;
earlier 12 tool-activity browser cases and 139 functional passes (11 existing skips).
The last review found synthetic hook completion lacked occurrence identity; this
was fixed and a regression added. Core/race/helper gates passed after that fix,
but no final independent approval or whole-product CI exists for this branch.
Do not deploy it. Generated web output is deliberately not committed; build from
source when this slice is resumed. Archive of the local generated diff remains at
`/workspace/tmp/gi-maintenance-20260926`.

The requested re-audit now takes precedence. Use Piclaw's actual UI behaviour and
source as the oracle when correcting Gherkin, then tests and Gi code. Record oracle
revision, source/test evidence and explicit safety deviations; do not silently
reconcile conflicting frozen specs. Clean audit baseline:5dc6bf7 in the separate
`gi-ux-reaudit-source` worktree. No production restart was requested.

The unrelated TUI autosave WIP2a87a79 in `/workspace/projects/gi` remains LOCAL ONLY,
unpublished and untouched, as explicitly required by the user.
