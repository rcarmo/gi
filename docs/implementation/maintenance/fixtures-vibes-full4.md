# Fixtures-vibes full4

Gi: `5a68f4005a9fdab60d72808579c0c483536f07d1`. Fixtures: `6e49ae1dfb0c6ca900a9173ecc80817c12fd77cb`.

The six-project run started at 2026-10-04T20:46:28.351Z and took 388.5 minutes. Playwright: 1233 passed, 280 failed, 71 skipped, 0 flaky.

| Report | Scenarios | Passed | Unlisted failures | Skipped | Listed failing | No suite test |
|---|---:|---:|---:|---:|---:|---:|
| Original | 316 | 185 | 1 | 3 | 49 | 78 |
| Reconciled skips, 2026-10-05 | 316 | 185 | 0 | 3 | 50 | 78 |

The original gate rejected core capability-absence skips, stale entries for extra004/005 and editor007, and the unlisted editor009 failure. Only the skips file changed before rerunning the report against the existing raw results. The reconciled report has zero problems and `Gate: OK`; the 280 failing browser tests still fail.

The report was regenerated with:

```sh
env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR make -C references/fixtures-vibes report PROFILE=/workspace/projects/gi-main/tests/fixtures-vibes/profile.json SHELL="/usr/bin/env nice -n 10 taskset -c 0-1 bash"
```

## Issue-backed failures

* [gi#40](https://github.com/rcarmo/gi/issues/40): `@ux-extra-001`.
* [gi#45](https://github.com/rcarmo/gi/issues/45): `@ux-shell-004`, `@ux-shell-007`, `@ux-workspace-014`, `@ux-terminal-001`, `@ux-terminal-002`, `@ux-terminal-003`, `@ux-terminal-004`, `@ux-terminal-005`, `@ux-terminal-006`, `@ux-terminal-007`, `@ux-terminal-008`, `@ux-terminal-009`, `@ux-terminal-010`, `@ux-terminal-011`, `@ux-terminal-012`, `@ux-terminal-013`, `@ux-terminal-014`, `@ux-terminal-015`, `@ux-terminal-016`, `@ux-terminal-017`.
* [gi#46](https://github.com/rcarmo/gi/issues/46): `@ux-shell-009`, `@ux-workspace-010`, `@ux-workspace-011`, `@ux-workspace-012`, `@ux-workspace-013`, `@ux-workspace-009`, `@ux-workspace-016`, `@ux-workspace-017`, `@ux-workspace-019`, `@ux-workspace-020`, `@ux-workspace-018`, `@ux-editor-001`, `@ux-editor-002`, `@ux-editor-003`, `@ux-editor-004`, `@ux-editor-005`, `@ux-editor-006`, `@ux-editor-008`, `@ux-editor-009`.
* [gi#47](https://github.com/rcarmo/gi/issues/47): `@ux-workspace-008`, `@ux-workspace-015`.
* [gi#48](https://github.com/rcarmo/gi/issues/48): `@ux-shared-009`, `@ux-shared-010`, `@ux-shared-011`, `@ux-shared-012`, `@ux-original-009`, `@ux-original-010`, `@ux-original-011`, `@ux-original-012`.
* [gi#50](https://github.com/rcarmo/gi/issues/50): `@ux-shell-env-002`, `@ux-shell-env-007`.

Windows shell-env002/007 are skipped because Windows builds have no host acceptance. Shell008 is suite-skipped on all six projects. The other entries cover mandatory core workflows and use `not-implemented`. Passing widget extra004/005 tests do not cover the newer strict sandbox/bridge lifecycle; #49 stays open for that acceptance.

## Local raw artifacts

* `references/fixtures-vibes/test-results/compliance.json`
* `references/fixtures-vibes/test-results/compliance-report-gi.json` and `.md`
* `references/fixtures-vibes/test-results/evidence-gi.json`
* `/workspace/tmp/fv-full4-original-report.json` and `.md`
* `/workspace/tmp/fv-full4.log` and `fv-full4-reconciled.log`
