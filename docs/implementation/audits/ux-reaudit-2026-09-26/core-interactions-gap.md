# Classic core interactions: five remaining clause audits

`tests/ux/features/classic/canonical/core-interactions.feature` contains
three previously reviewed recovery/card-rejection clauses (`003`, `012`,
`013`), which are not rescored here. The pinned Piclaw 3.2.4 source manifest
marks `btw-panel.ts`, `floating-widget-pane.ts`,
`adaptive-card-submission.ts` and `notification-delivery-coordinator.ts`
identical; `generated-widget.ts` changed in Gi. Source presence is not a
mounted Gi capability or current Piclaw runtime acceptance.

| ID | Frozen clause versus mounted Gi | Bounded finding |
|---|---|---|
| `@ux-extra-001` | `btw-panel.ts` has the expected question/thinking/answer/Retry/Inject conditions, and `app-main-shell-render.ts` supplies callbacks. The mounted Gi entry `web/src/app.ts` does **not** render that shell's BTW panel; `api.ts` stubs `streamSidePrompt`. | Installed 3.2.4 BTW UI probe 6/6 saw error, Retry, streaming answer and intercepted Inject. Gi mounted 6/6 submitted `/btw` as an ordinary prompt with no panel. See [BTW audit](classic-btw-panel-review.md). The clause remains unmapped. |
| `@ux-extra-002` | `adaptive-card-submission.ts` now enforces the installed Classic block identity bounds and normalises omitted `action_type`; `post.ts` accepts `Action.Submit` events but `api.ts` deliberately rejects submission until identity/authorisation/persistence exist. | Installed 3.2.4 persisted-block filter 6/6, Gi mounted HTTP/SQLite receipt filter 6/6 and native helper bounds pass. See [card identity audit](classic-card-identity-review.md). No card-action submission or whole-clause acceptance. |
| `@ux-extra-004` | `generated-widget.ts` builds persisted payloads only with usable artifact content and distinguishes live `loading`/`streaming`/`final`/`error` status. `post.ts` and `FloatingWidgetPane` mount in Gi, but `gi-sse-client.ts` does not bind the widget events and `app.ts` has no mounted live-widget event branch. | Installed persisted artifact probe 6/6 and Gi mounted HTTP/SQLite persisted slice 6/6; installed widget SSE probe 6/6 reports raw widget frames delivered without a pane. See [widget audit](classic-core-widget-review.md). No live Gi journey or whole-clause acceptance. |
| `@ux-extra-005` | `web/src/app.ts` closes the mounted floating pane with `setFloatingWidget(null)` and has no explicit dismissed-session-key update; the separate, unused composition chain implements such a dismissal. Close does not call queue mutation in the mounted callback. | Partial native behaviour: no live pane mounts from SSE in either shipped Piclaw or Gi, so the frozen visible-live-widget dismissal precondition was not exercised. No tagged browser journey or clause acceptance. |
| `@ux-extra-011` | `notification-delivery-coordinator.ts` computes same-device live presence, suppresses for visible candidates, otherwise chooses lowest client ID; withdraw removes stored presence. `gi-notifications-state.ts` uses the coordinator with locking and cleanup. | Installed Piclaw selected-chat presence publish/withdraw **6/6** and mounted final-reply ownership with synthetic sibling records/mocked Notification **6/6**; helpers **10/10**, focused Gi browser **24/24**. See [notification audit](classic-notification-coordination-review.md). No two-mounted-client delivery or OS acceptance. |

The earlier `@ux-extra-003`, `012`, `013` bounded review remains in
`recovery-card-review.md`; none of these five findings upgrades those clauses
or the full Classic shell. No production code or frozen Gherkin changed.
