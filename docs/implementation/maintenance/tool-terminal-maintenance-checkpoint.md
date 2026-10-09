# Tool terminal maintenance checkpoint

Paused at the user's full UX re-audit request, then checkpointed for maintenance.
The Piclaw 3.2.4 active-contract alignment is now pushed separately on
`wip/maintenance-ux-reaudit-20260926` through `eb96dc0`; this branch remains
isolated and is not approved for deployment or full Shared40 acceptance.

Owned source: occurrence-bound tool starts/terminals, explicit cancelled/aborted
persistence, status labels and native/browser tests. Runtime remains deployed
05e287f; Shared38 mapping 8196355 and docs5dc6bf7 were already complete.

Checks passed again on the exact `500e02f` source: `make test-tool-activity vet
bun-checks`, `make test-ux-tool-terminal` (12/12 across six projects) and
`make check` (139 functional passes, 11 existing skips). A separate read-only
review of occurrence matching and stopped-event persistence found no blocker;
a broader delegated review timed out and earns no credit. The earlier synthetic
hook occurrence bug has its regression. No GitHub Actions run, PR approval, four
builds or deployment exists for this branch.
Do not deploy it. Generated web output is deliberately not committed; build from
source when this slice is resumed. Archive of the local generated diff remains at
`/workspace/tmp/gi-maintenance-20260926`.

The re-audit preserves historical feature hashes and keeps Gi-specific deviations
outside Piclaw parity. This branch is still based on `5dc6bf7`; do not merge the
separately pushed active-contract work merely to publish this prerequisite. The
Shared40 expandable tool-output pane, mounted browser UX, production verification
and physical/accessibility checks remain separate. No production restart was
requested.

The unrelated TUI autosave WIP2a87a79 in `/workspace/projects/gi` remains LOCAL ONLY,
unpublished and untouched, as explicitly required by the user.
