# Feature contracts

All tracked Gherkin files live here. The folders separate ownership and evidence scope:

- `ux/classic/` and `ux/shared-canonical-ux.feature`: Gi's local web UX contracts, aligned with installed Piclaw 3.2.4. The browser scenarios Gi is tested against now come from the shared fixtures-vibes suite (`references/fixtures-vibes`); see [the browser test guide](../tests/ux/README.md) and [the validation ledger](VALIDATION.md).
- `ux/upstream/`: byte-preserved historical Classic and Shared snapshots. Source paths in `SHA256SUMS` remain those of the upstream checkout. The catalogue checks the bytes; these are provenance fixtures, not active tests.
- `ux/oracle/`: isolated reference-only browser scenarios. They do not award Gi parity.
- `ux/additions/`: Piclaw additions outside the frozen Classic/Shared corpus; their criteria require separate acceptance.
- `gi/`: Gi-specific session, settings, TUI and web-deviation criteria. Passing them cannot establish Piclaw parity.
- `tui/`: terminal acceptance, including Markdown under `tui/markdown/`.
- `search/`: derived workspace-index proposals, reported separately from Classic/Shared.

Historical audit artefacts under `docs/implementation/audits/ux-reaudit-2026-09-26/` retain their original pre-move paths; they are not current file locations.

`tests/ux/support/feature-tree.test.ts` checks that every tracked Gherkin file is in this tree and parses. `tests/ux/support/provenance.mjs` verifies the frozen snapshot hashes. The file-by-file status in [VALIDATION.md](VALIDATION.md) stays open until an independent installed reference supports each contract; a preserved or parsable file is not validation.
