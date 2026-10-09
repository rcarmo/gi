# Terminal pre-admission draft guards

When no model is selected, Return now rejects a model-bound submission before
clearing the editor, adding command history, claiming media or recording a
queued-draft shortcut. The full text, cursor, undo/yank state and existing queue
shortcuts remain available. The guard also applies while another turn is active.

The existing model-selection guidance uses transient command output; regular
mode prints it above the dock. No permanent row, new shortcut or input limit is
added. Alt-M can select a model without replacing the draft; the user must press
Return again to submit. Slash commands and local `!!` commands remain available.
Model-directed `!` and `@peer` input use the guard, as ordinary submissions do.

This fixes a pre-admission loss, not durable text recovery. Plaintext and
queued-draft restart persistence still need a journal/ownership contract that
cannot offer already-accepted work as unsent after a crash. Generic submission
failure recovery and unavailable-provider validation remain separate work.

## Deployment

Exact `c8450508f54fa452b1bc98baae9e1d66ed515cc7` passed product
[CI36225264763](https://github.com/rcarmo/gi/actions/runs/36225264763), including
terminal submit/retry PTYs and all four platform builds. Run36225411598 was an
Actions-cleanup workflow, not product CI; it supplies no deployment acceptance.

Detached-source Makefile build/restart runs on8090, PID1130830,
PGID/SID1130724. Six read-only Chromium/WebKit probes at390/820/1440 pass existing
composer/model/session/context/theme/focus/draft checks with zero HTTP writes,
errors or permission requests. DB62sessions/51turns/146messages, integrity/FKs,
normalised SQL excluding runtime leases, and absent auth remain unchanged.
Newer plaintext/paired journal code is excluded. Terminal submission acceptance
uses disposable PTYs, not live mutations or physical/Visual acceptance.

## Verification

- `make test-tui-submit-guards`: race checks, three repeats. Covers idle/busy,
  Unicode multiline text/cursor/undo/yank, repeated submission, unchanged
  history and queued shortcuts, zero turns/media-journal writes and available
  slash commands. Existing no-model attachment coverage remains enabled.
- `make test-tui-unselected-model BIN_DIR=/tmp/gi-tui-unselected-bin`: six
  fullscreen/regular PTYs at60×18,100×22,140×36. Repeated Return retains the draft
  and middle cursor. Alt-M persists a session model without submitting;
  subsequent insertion and explicit Return store exactly `unsent 中文🙂 Xtail`
  once. Global settings, regular terminal ownership and idle dock size remain
  unchanged. The isolated fixture uses a whitespace default model with a usable
  test catalogue to exercise the empty-after-trimming selection state.
- `make test vet bun-checks` passes.
- `make test-ux BIN_DIR=/tmp/gi-tui-unselected-functional-bin`:107passed,
  11existing skips.

The unit/race and PTY targets join required CI without removing tests or
increasing existing timeouts. Initial unit compilation used incorrect editor
field names; those were corrected before accepted runs. A delegated review
timed out and supplies no independent acceptance. No live data or terminal
mutation, physical-device, visual or new web feature-mapping credit is claimed.
