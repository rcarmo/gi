# Model picker keyboard gap (current oracle, frozen Classic021)

Piclaw 3.2.4 shipped source `components/model-picker.ts:75-103` handles ArrowUp,
ArrowDown, PageUp, PageDown, and Control/Meta+Home/End while the model search
field is focused. Plain Home/End stays with text editing; Enter activates the
highlighted entry and Escape closes the picker. The file is byte-identical to
the frozen `70d33bc` version (`frozen-to-oracle-sources.json`). The frozen
`@ux-original-021` clause at `tests/ux/features/classic/canonical/canonical-ux.feature:205-218`
correctly describes that keyboard contract.

Before the audit-branch correction, `tests/ux/models.spec.mjs:46-72` tagged
a held model response after switching sessions as `@ux-original-021`. It did
not test those keys. Gi's supplied `web/src/components/compose-box.ts:1491-1535`
handles only ArrowUp/Down and Enter; its existing build adapter used
`web/src/gi-model-picker.ts`, which rejected Control/Meta. The correction below
changes that Gi-owned helper. Session picker has separate Page/Home/End evidence at
`tests/ux/session.spec.mjs:1031-1063`, which cannot be borrowed for the model
picker clause.

**Correction on the audit branch (`f5ad830`):** `web/src/gi-model-picker.ts` now handles
Control/Meta+Home/End only for a focused search input. Plain Home/End
retain text-editing behaviour; other modified keys remain unhandled.
Model Arrow navigation clamps at both ends and Page keys move seven enabled
results, matching shipped Piclaw's `ui/model-catalogue.ts:697-715`. The old
helper reused the wrapping eight-row session-picker rule. The
existing guarded `scripts/patch-model-picker.mjs` adapter wires this helper into
the supplied component at build time without editing `compose-box.ts`.
`tests/ux/models.spec.mjs` moved the misleading Classic021 tag from the stale
response test to a real picker journey. Six Chromium/WebKit phone/tablet/desktop
projects passed search filtering, plain caret movement, modified Home/End,
clamped Arrow/seven-row Page, exact Enter mutation, Escape focus, one PATCH,
retained draft and no prompt submission. Shared34's existing six-project native fixture checks
navigation over disabled entries and selection; its expected order and an
obsolete option text locator were corrected to reflect the supplied sorted
catalogue followed by Gi's Current group. Six Shared34 projects and 42 model
panel projects passed. `make test-model-panel-helpers` passed eight tests.
The canonical disposable `make test-ux-parity` fixture passed Classic020–022
in all six projects (18 browser cases). An initial run against the unrelated
`test-ux-steer` fixture could not provide Classic020/022 models, so the tagged
journey now selects the fixture's actual current model after its keyboard
checks. The canonical rerun passed without weakening a key assertion.

The audit-branch commit fixes the demonstrable Gi model-picker code/test gap.
Frozen Classic021 also covers the **session** picker, which has separate tagged tests;
a dedicated isolated 3.2.4 UI probe also passed Page and Control/Meta+Home/End,
search focus and Escape focus in all six browser projects. Its first PageDown
can land at the first filtered entry when the previous current model remains
active outside the search; it is not used to assert an exact seven-row jump. It used twelve listed
fixture models, no live model mutation, blocked entries or physical keyboard.
No full Classic021, physical keyboard or deployed parity credit follows from
this correction. Scoped CI/review remains to do.
