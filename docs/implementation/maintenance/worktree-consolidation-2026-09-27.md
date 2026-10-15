# Local worktree consolidation

All local branch tips were already included in `ffff1ed`. This change adds the uncommitted worktree source captured on 2026-09-27 at 22:19:37 UTC, after the user explicitly requested consolidation into main and publication.

## Included work

- TUI Markdown spacing, progressive rendering, label removal, plain subprocess output, transcript layout and selection/search row alignment.
- Session activity restoration and concurrent tool/thinking presentation tests.
- Transcript scrollbar track/drag handling and its three-size PTY test.
- Persistent, session-scoped input history, legacy-prompt fallback, retention, draft/cursor restoration, and restart/session-switch coverage.
- Expanded Markdown and input-history Gherkin and Make targets.
- Nine staged removals of old `artifacts/tui-audit-20260525/` captures from the integration worktree.
- Web assets rebuilt from the combined main source. Old worktrees' generated chunks and dependency links were not copied over newer source output.

`/workspace/tmp/gi-worktree-consolidation/` retains inventories, staged/unstaged/combined patches, untracked files, the source cutoff, failing captures and gate logs. The original worktrees were not reset, removed or committed by this consolidation. Edits arriving after the cutoff remain there for a later pass.

## Merge resolutions

The older TUI worktree predates main's web receipt, queue-hold and numeric-message migrations. The input-history table was merged additively into the current schema; those web tables and migrations are retained. A first combined run exposed the incompatible schema replacement and was retained as a failed run.

Main's UX-oracle Make targets and checklist entries are retained alongside the TUI additions. The input-history target passes `FEATURE_FILE`, `ARTIFACT_DIR` and `TEST_DIR` as environment variables to the runner.

Markdown PTY assertions now compare visible words across intervening foreground/bold ANSI sequences, normalise nonbreaking layout spaces and tolerate table-cell alignment padding. Assertions still enforce user/assistant background boundaries, code indentation, inline styling/reset, link targets, source copying, viewport limits and draft preservation. Earlier failures remain in the local evidence directory.

## Validation

- `make check`: successful after schema reconciliation; Go tests, vet, web build, hook checks and 139 functional tests passed; 11 functional tests skipped.
- `make test-tui-markdown`: 24 real tmux cases passed across fullscreen/regular modes and three sizes.
- `make test-tui-scrollbar`: 3 real tmux cases passed.
- `make test-tui-source-copy`: 6 real tmux cases passed.
- `make test-tui-input-history`: isolated Gherkin scenario passed, including draft/cursor recovery, session switching and restart persistence.
- `git diff HEAD --check`: passed.

The final combined invocation completed `make check`, Markdown, scrollbar and source-copy tests before exposing incorrect environment forwarding in the input-history target. That target was corrected and passed independently. An independent review attempt timed out; there is no independent approval claim.

## Follow-up and cleanup

`41f87fc` published the initial consolidation. `93fe374` restricts recall history to accepted prompts; `9d6ba09` merges the source worktree's later `098fe8a` commit. `d0470da` fixes the screenshot's Codex request rejection and TUI system-message attribution, with separate evidence in [TUI Codex regression](../terminal/tui-codex-regression-2026-09-27.md).

After the user's deletion request, 23 redundant temporary worktrees were removed, two stale registrations pruned, and five fully merged local branches deleted. Before removal, every candidate was checked for merged ancestry, archived changes and active processes. No remote branches were deleted.

Recovery files under `/workspace/tmp/gi-worktree-consolidation/cleanup/`:

- `all-refs-final.bundle`: verified Git bundle containing all refs before deletion.
- `retired-worktrees.tar.zst`: verified 276 MB archive with 51,841 entries, including changed sources, generated assets and test evidence. Build binaries, node_modules and .git pointers were excluded.
- `archive-list.txt`, `inventory.json`, `removed.json` and per-worktree patches.

The first gzip archive was interrupted and failed verification. It was replaced by the verified zstd archive before any worktree deletion, then removed.

Two worktrees remain: `/workspace/projects/gi-main` on main and `/workspace/projects/gi`, the Git common-directory owner and working directory of the existing live TUI. The latter and its checked-out branch were retained to avoid disrupting the running session. Their committed source is included in main.

This consolidation does not fix the outstanding web chat UX or establish Piclaw parity. No production database migration, restart or deployment was performed.
