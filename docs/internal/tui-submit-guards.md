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
