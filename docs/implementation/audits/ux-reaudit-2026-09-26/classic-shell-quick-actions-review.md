# Classic canonical 001–006: menu and Quick Actions

The pinned Piclaw 3.2.4 source manifest marks
`runtime/web/src/components/timeline-quick-actions.ts` byte-identical to Gi's
copy. Its key guard excludes editable controls, popup regions, modified,
repeated, composing and consumed keys. `web/src/app.ts` mounts that component
with Gi session/workspace commands. The Classic menu is mounted through
`TimelineMenu` and its shell callbacks. The Piclaw 3.2.4 UI was not replayed
for these six cases; installed source and the Gi fixture are separate evidence.

| ID | Tagged Gi browser path | Boundary |
|---|---|---|
| `001` | Real hamburger pointer/keyboard opening, Escape and outside-pointer dismissal, no composer submission. | 6/6 viewport/browser projects. |
| `002` | Workspace visibility toggles out and back with the unsent draft and empty turn history. | 6/6; no workspace CRUD credit. |
| `003` | Timeline typing opens focused query; native Agents/Workspace/Slash groups, matching, arrow wrap and Enter activation run. | 6/6; selected `/model` prefill is Gi-specific, while frozen `007` has a separate 3.2.4 oracle delta. |
| `004` | An untagged Gi test checks real composer, menu and session search plus synthetic DOM for absent Monaco/terminal/contenteditable selectors. | **Incomplete outline:** six target examples are not all exercised as real supplied controls. No tagged case credit or physical IME proof. |
| `005` | Tagged test rejects repeated, composing, whitespace, modifier, default-prevented and registered-shortcut keys, then opens with an eligible real key. | 6/6; synthetic metadata are not physical IME acceptance. |
| `006` | Escape/outside pointer dismisses, resets query and executes no action; a later opening starts fresh. | 6/6. |

Focused `make test-ux-parity` filter for `001`, `002`, `003`, `005`, `006`
passed **30/30** across Chromium/WebKit phone, tablet and desktop. The
separate `@ux-original-004` six-example outline remains a behavioural test
gap; selector logic and artificial targets cannot fill it. This review changes
no production code or frozen Gherkin.
