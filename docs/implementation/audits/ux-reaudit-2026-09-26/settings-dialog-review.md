# Classic settings dialog 001–005: bounded native shell

`tests/ux/features/classic/settings/settings-dialog.feature` is a frozen
Piclaw Classic contract. Gi mounts `web/src/gi-settings.ts` with a modal
`BodyPortal`, a General snapshot cache, a cold loading shell, and pane modules
loaded on visit by `gi-settings-lazy.ts`. The installed 3.2.4 Classic Settings
shell was probed separately with disposable responses. The new shortcut probe
checks one target-sensitive opening rule; these tests do not establish parity
for every setting or subsection.

| ID | Tagged native assertion | Boundary |
|---|---|---|
| `001` | Three rapid Control-comma presses from a focused composer leave one Gi Settings dialog and portal; Escape restores composer focus. `GiSettings` opens in a capture-phase window listener without an editable-target guard. Installed Classic 3.2.4 blocks all three presses from its focused composer textarea, but opens one dialog after three presses from noneditable body focus. | **Target-sensitive Gi divergence.** The frozen scenario does not specify where focus starts; the Gi test does. Installed and Gi assertions cannot be merged into a whole-clause pass. Browser key events do not establish OS/physical shortcuts. |
| `002` | Gi opens via hamburger, visits panes, then reopens while `/api/runtime/config` is held: General shows a cached snapshot in under one second, one portal, no draft/media loss or writes. Installed Piclaw 3.2.4 mounted Classic independently reopens with a cached General compose-upload value `64` while a second `/agent/settings-data` read is held; it renders within one second, then updates to `96` after release with one dialog and an intact draft. | Both times come from disposable browser fixtures, not a global latency SLO. Piclaw performs a fresh read on reopen while showing cached data; Gi's route and cached values differ. No production config mutation or whole-clause credit for every pane. |
| `003` | Gi holds its first `/api/runtime/config` read, shows `Loading settings…` with General selected, then renders General values within two seconds. The existing installed Piclaw 3.2.4 shell probe instead opens after the Classic dialog module is already loaded: General remains interactive while `/agent/settings-data` is held, with no import-time loading shell in that sequence. | **Different first-open boundaries.** Gi's held-data shell is not Piclaw's lazy-import shell. Neither run observes an installed cold import-time shell or compares like-for-like loading; the frozen clause lacks whole-scenario parity. Timings are fixture-only. |
| `004` | Gi Compaction's **Saved context window** spinbutton accepts typed `128000` without a PATCH. Installed Piclaw 3.2.4 has a different Compaction numeric text stepper: **semantic summary input limit** accepts `128000` within its 500–200000 range without a settings write; out-of-range `300000` is marked invalid while editing and clamps to `200000` on blur. | **Typability subset, different field and role.** The installed probe does not establish a Saved context window control, equivalent budget semantics, valid policy, settings save or live persistence. Do not merge the two field checks into whole-clause parity. |
| `005` | Gi shows General without other pane chunks; Models, Appearance, Compaction, Providers and Authentication each fetch a chunk when visited and reuse it on revisit. Installed Piclaw 3.2.4 source pre-caches General and dynamically resolves other built-in section components. Mounted Classic probe sees General first, Models content only after its click, and no extra script request on revisit. | **Different build graph.** Installed `app.bundle.js` emitted no per-pane network chunk in this fixture (four script requests already present at open); source-level component loaders still run on section selection. Only Models was mounted in the Piclaw probe. No whole-clause proof for every unopened built-in or cold import. |

Focused `make test-ux-parity` filter for all five tags previously passed
**30/30** across Chromium/WebKit phone, tablet and desktop projects. The
current `@ux-settings-dialog-001` Gi run passed **6/6**. `make
test-piclaw-settings-shortcut` passed **6/6** installed mounted Classic cases:
focused textarea blocks Control-comma, noneditable body opens one dialog on
three presses and retains a draft. The separate installed shell probe covers
General/Escape/backdrop, not this shortcut boundary. No settings writes,
physical input, deployed Gi or whole-settings acceptance was tested. The Gi
handler and frozen Gherkin were left unchanged pending a decision on the
composer-focus divergence.

`make test-piclaw-settings-reopen` passed **6/6** installed Classic cases
across Chromium/WebKit phone/tablet/desktop. The focused Gi `002` journey
passed **6/6** separately; it includes draft/media ownership and pane visits that the
installed probe did not exercise. The installed fixture held a second read;
it did not verify live settings persistence or latency under network load.

The existing `make test-piclaw-settings-shell` probe passed **6/6** again:
General was present while its first data request was held, and the ordinary
open showed no `.settings-dialog-loading-shell`. The focused Gi
`@ux-settings-dialog-003` run passed **6/6** with its own held runtime-config
read and loading status. The Piclaw probe opens after its module has loaded;
this comparison does not test Piclaw's cold dynamic import, so `003` stays
unmapped as a complete scenario.

`make test-piclaw-settings-stepper` passed **6/6** installed mounted Classic
cases across Chromium/WebKit phone/tablet/desktop; the focused Gi
`@ux-settings-dialog-004` run passed **6/6** separately. Both show typed
`128000`. Their controls have different labels, roles and settings meanings;
neither run saves a policy. The installed probe also checks the local invalid
value and blur normalization without claiming downstream equivalence.

`make test-piclaw-settings-pane-load` passed **6/6** mounted Classic cases:
General appears first, Models content appears after selection and returning to
Models requests no new script. The installed release bundles that pane code
into its shipped app graph, despite source-level section loaders. Focused Gi
`@ux-settings-dialog-005` passed **6/6** with five actual on-demand chunk
requests and cached revisits. These are different module graphs; the installed
probe does not test all unopened panes or a cold import.
