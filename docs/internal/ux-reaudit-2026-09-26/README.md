# UX re-audit: maintenance checkpoint, not a completed audit

User direction: audit all Gherkin for correct basic UX interactions, then cascade
corrections to tests and code. Use Piclaw's actual behaviour and source as the oracle.
Do not make the requirements merely describe current Gi code. Record the Piclaw
revision and environment for each comparison, observed interaction and source
provenance, expected test assertion, Gi handler and any deliberate safety deviation.
Conflicting frozen contracts require explicit resolution; hashes remain unchanged.

## Baseline and inventory

Initial clean source baseline: `5dc6bf77968a1c6da1c14292d2a894085012f7b1`.
Its 37 tracked feature files, 358 scenario definitions and 421 expanded cases parsed
successfully with Cucumber. This baseline is immutable; the separate oracle-only
feature and corrected Gi settings scenarios are additive work after the inventory. Classic/shared frozen-source hashes verified. Includes
native TUI/search, additive Gi settings/session specs, passkey additions and all
Classic/shared contracts—not only the web mapping catalogue.

`scenario-inventory.json` preserves every definition, Background, step, outline row
and expanded count. Test candidates are lexical links, NOT verified coverage.
The test index cannot resolve every dynamic tag/title and test factory; a missing
candidate is not proof that no test exists. All rows remain review-not-complete and
oracle-not-yet-compared. These artifacts do not supersede prior accepted evidence.

Four broad and two smaller delegate attempts timed out without usable completed
reviews. An empty partial JSON was excluded. Their work earns no review credit.
The lead read the whole compact scenario corpus; clause-by-clause test/code/oracle
verification remains unfinished. No new runtime pass or parity credit is claimed.

## Work after maintenance clearance

Maintenance hold was explicitly removed; no restart was requested. Piclaw's
installed release changed during the audit from 3.2.3 to **3.2.4**; the current
oracle is asset version `990f0c49a932` with the source-map hash pinned in
`tests/ux/oracle/piclaw-3.2.4-reference.json`. Relevant source-map modules are
byte-identical to the removed 3.2.3 release. `make test-piclaw-oracle-basic`
passed against 3.2.4 with hash-checked UI assets, the production-default SVG
flag substitution and isolated API fixtures; see `oracle-deltas.md` and
`tests/ux/features/oracle/piclaw-3.2.4-basic-interactions.feature`. The older
3.2.3 reference is historical. These probes are **not Gi parity tests**.

`scenario-status.csv`/`scenario-inventory.json` contain all 358 baseline definitions.
Only 12 are currently annotated with qualified oracle/conflict/correction findings;
346 still await clause audit. `expanded-crosswalk.json` enumerates all 421 cases,
including the 56-case pinned passkey criteria ledger. It finds 117 direct-tagged
browser-test candidates; 248 have neither such a candidate nor passkey-ledger
entry. **A direct tag is not behavioural acceptance**, and zero direct candidates
does not establish absence of another test. The crosswalk is a
triage aid, not a completed UX review.

The additive `@gi-settings-004` text was corrected to reflect advertised thinking
support, with `@gi-settings-028` and `@gi-settings-029` scenario clauses. A new
six-project native browser test for model-switch reset passed; the earlier supported
thinking provider-request test is tagged. The broader thinking run was interrupted
by the command harness after progress through 32/42 tests; its result is unknown.

The basic HTTP run had 31/32 passes, but Chromium initially stayed on Loading Gi
because static requests failed with `net::ERR_NETWORK_CHANGED` before app boot.
An unchanged rerun was interrupted at 21/32 by `make: wait: No child processes`.
Neither run counts as a green gate. Traces and logs are under the local audit
workspace; no assertion, timeout or production code was changed for that failure.

## Initial concrete issues to verify against Piclaw

- `tests/features/settings/gi-settings.feature:40` still says thinking is read-only;
  `web/src/gi-settings-models.ts:270-305` exposes supported Apply-thinking controls.
  The tagged Settings tests use an unsupported model, so that contradiction can hide.
- `features/search/workspace-index.feature:33` requests automatic background refresh;
  later clauses at199/211 require read-only GET and explicit refresh. Production
  `internal/web/workspace_index.go:138-141` rejects GET refresh. Choose one explicit
  contract, informed by the user's intended Piclaw parity and safety requirements.
- Classic queue-return `@ux-original-017`/`@ux-compose-004` replaces draft and clears
  media; Shared28 preserves/merges latest draft/media before deletion. These are
  conflicting policies, not interchangeable test mappings.
- Classic007 replaces Quick Actions prefill; Shared16/17 preserve the existing draft.
- Classic029 keeps fenced SVG as source; Shared41 renders sanitized inline SVG.
- `features/tui/keyboard_behavior.feature:4` bundles focus, history, scrolling,
  resizing and quit, but scroll assertions only require unchanged text to remain
  present. It cannot by itself prove scrolling or cursor/viewport preservation.
- Many Classic requirements describe callbacks, supplied props or “can display”
  rather than a complete human journey. Strengthen with actual trigger, exact
  native identity/result, focus/caret, failure/retry and preservation assertions.

These are source-backed audit leads, not an exhaustive defect list. No frozen
spec, test or production source was changed in this audit worktree.

## Resume

1. Establish the Piclaw oracle revision/source and isolated runnable environment.
2. Compare each of all358 scenario definitions and421 examples against actual
   pointer/keyboard/touch, focus, draft/media, session, async-error/reload behaviour.
3. Trace each clause to test assertions (not tags), identify mock/DOM/source-only
   substitutions and exact native-code paths. Keep absent evidence distinct from bugs.
4. Produce per-scenario dispositions and a prioritized requirement→test→code plan.
5. Implement approved corrections with Make gates and fresh evidence; no deployed
   changes without exact-revision whole-product CI/all four builds.

Shared40 WIP is separately checkpointed at500e02f on
`wip/maintenance-tool-terminal-20260926`; it must not be deployed automatically.
Live runtime remains05e287f on8090. TUI WIP2a87a79 remains local-only and untouched.
Maintenance hold is cleared. Continue from this branch without redoing completed
Shared35/36/38 deployment or touching live data.

Inventory scripts currently retain absolute local audit paths; they are checkpoint
reproduction aids, not shipped product tooling. Generated candidates need manual
validation, including dynamically generated test names and non-web harnesses.
