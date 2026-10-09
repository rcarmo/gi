# Workspace editor acceptance (#46)

On 9 October 2026, Gi `881d67c6` with the unchanged fixtures-vibes `bb7786f7b1a65508d763e4e919e49899793e354f` UI passes 19 shared editor scenarios. Issue #46 stays open for workspace018's shared conflict-test mismatch. No frontend source, shared assertion, oracle record or production capability declaration was changed.

## Refinement and scope

The issue and shared scenarios already specify the user, workflow and acceptance criteria: workspace users edit and save complete files through the adopted CodeMirror editor, including Vim, dirty tabs, rename/MRU/pinning, preview, docking/popout and agent-open requests. The existing native editor and owner/session/request-bound `open_workspace_file` protocol provide the implementation; this slice verifies the current handoff and removes stale acceptance entries.

Runtime constraints remain pure Go, bounded UTF-8 edit reads/writes, rooted filesystem access and conditional revisions. Errors retain unsaved edits; unchanged content avoids writes; newer edits survive an in-flight save. Browser/editor preferences remain frontend-owned. No WYSIWYG editor, unconditional overwrite, test relaxation, new frontend fork or Windows device claim is in scope. Success means the exact shared scenarios pass on their applicable browser/viewports. Closure additionally requires independent shared conflict acceptance.

## Shared checks

The production profile was used for the profiled matrix, with zero retries:

```sh
make fixtures-vibes-focused PROFILING=1 \
  FIXTURES_SPEC_ARGS='editor.spec.ts editor-open-file.spec.ts workspace-flows.spec.ts shell-menu.spec.ts --grep "@ux-editor-|@ux-workspace-(009|010|011|012|013|016|017|019|020)|@ux-shell-009"'
```

Results: **107 passed, seven source skips, zero failures**, 114 executions, 1,197.7 seconds. Source skips are editor003 on the four touch projects and workspace009 on all three WebKit projects. These are unchanged shared-suite exclusions, not Gi-added skips.

Accepted IDs: editor001–009; workspace009–013,016–017,019–020; shell009. Agent-open includes actual tool-driven SSE, invalid/outside-path rejection and cross-chat isolation. Vim, save failure/retry, unchanged writes, held-save typing, clean external updates, popout/reattach, tab state and monospaced preview all have the shared checks applicable to their projects.

Eighteen stale entries were removed from Gi's skips file; editor007 already had no entry. The targeted report gate passes for all 19 scenarios. This is not a full-suite compliance result; the frozen October run remains historical. The current catalogue treats editor behaviour as core or `@cap-workspace`; it no longer defines `@cap-editor`, so no such capability was added.

The consumer target now selects only the two editor consumer files, rather than also running unrelated terminal tests. Both checks pass on Chromium and WebKit desktop: **four passes**. They exercise Reload, cancelling an overwrite review without writing, exact reviewed-revision PUT, create-only Save Copy and preservation of the next-save revision after a clean external SSE refresh.

## Remaining conflict mismatch

An unchanged Chromium-desktop run of shared workspace018 reproduces the failure at `suite/specs/editor.spec.ts:265`: after clicking Overwrite, it expects a clean tab without approving the adopted UI's **Review overwrite** dialog. The dialog remains open, no write is authorised and the tab correctly stays dirty. Cleanup then times out behind that unanswered dialog.

Gi's separately scoped consumer approves the reviewed revision explicitly and passes on both desktop browsers. The shared scenario remains a `known-defect` entry linked to #46. Closing the issue requires upstream oracle-backed alignment of this scenario with the reviewed-overwrite workflow, followed by unchanged shared acceptance. Gi does not remove the review, add an unconditional-write fallback or award shared acceptance from its consumer test.

## Profiling and disposal

Chromium supplied 114 browser CPU/heap captures across 57 executions. Aggregate sampled CPU was 435.5 seconds, about 81% idle; active sites include shared app handlers and CodeMirror. Browser heap samples total 438,145 KiB across test endpoints, dominated by V8, CodeMirror, app handling and KaTeX. That sum is not peak memory or allocation history. WebKit has functional evidence only.

Six native worker profiles were reviewed for CPU, allocated bytes and objects. Existing go-ai model-catalogue cloning accounts for 24–34 MiB per worker; SQLite/JSON/HTTP handling supplies much of the remaining allocation work. No isolated runtime latency or memory regression, old/new throughput comparison or performance improvement is established by this acceptance run.

The tests and JSON reporter finished successfully just before the outer 1,200-second command deadline. The deadline interrupted the wrapper's repeated Node-summary pass, not the tests. Remaining browser and native captures were reviewed separately. Raw profiles, disposable logs, copied specs, fixtures and traces were removed after analysis; only these conclusions remain.
