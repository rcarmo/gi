# Classic context and compaction 001–008: bounded clause review

The installed Piclaw 3.2.4 `components/compose-box.ts` **changed** since the
frozen 70d33bc tree; `ui/app-agent-turn-events.ts` is identical. Current
Piclaw source was inspected for the context pie, upload/submit and model
handling, but no Piclaw backend compaction or meter UI journey was run. The
Gi tests below use disposable native sessions and provider fixtures. Do not
infer measured token usage from a local estimate or from a passing tag alone.

| Frozen IDs | Gi assertion → code | Scope and limit |
|---|---|---|
| `@ux-compaction-001`, `002`, `004`, `@ux-context-004` | One parameterised `tests/ux/compaction.spec.mjs:30–52` case checks native automatic compaction status, elapsed label, accepted completion, refreshed usage, measured request kept distinct from estimated history, draft/media and session ownership. Gi `web/src/components/compose-box.ts:202+`, `web/src/app.ts:151–178` and native compaction state provide display/refresh. | Four tags share a single assertion body per project; no atomic SSE deadline or guarantee that usage shrinks. |
| `@ux-compaction-003` | `compaction.spec.mjs:53–62` uses the visible Stop control for a captured active run, checks cancellation without clearing draft or targeting another turn. Gi `web/src/app.ts:388–393` calls native run cancellation. | Not generic TUI cancellation or physical input. |
| `@ux-compaction-005` | `compaction.spec.mjs:63–67` checks suppression notice, native failure/retry detail and retained draft. | Only the fixture suppression path, not all provider backoff categories. |
| `@ux-compaction-006`, `007` | `tests/ux/context-fit.spec.mjs:27–61` checks blocked context-fit selection (including native 400), accepted model mutation, refreshed context-window/percentage, preserved measured token/turn provenance and draft/media across reload and session switch. Gi `web/src/gi-context-usage.ts`, model-picker guard and native session model route own the boundary. | No current Piclaw backend model mutation probe. |
| `@ux-compaction-008` | `tests/ux/models.spec.mjs:169+` submits supported `/model test/bootstrap` and rejects an unknown name without a turn or lost draft. | Configured provider catalogue in disposable Gi server only. |
| `@ux-context-001`, `005` | `tests/ux/context-meter.spec.mjs:38–63` checks measured usage label, tooltip, rounded percentage, reload/session preservation and green/amber/red boundaries (75% and 90%). `web/src/app.ts:151–178` synchronises `data-tooltip`/aria labels with `ContextPie`. | Browser CSS computation, not Piclaw pixel or physical contrast acceptance. |
| `@ux-context-002` | `tests/ux/context.spec.mjs:4+` checks unavailable token values remain unknown and no fabricated number appears after turn/model switch, while draft persists. Gi `web/src/gi-context-usage.ts` represents provenance. | It does not prove provider usage for every model. |
| `@ux-context-003` | `compaction.spec.mjs:124–153` checks manual Compact availability, explicit admission, one run, disabled duplicate, and retained draft/media. `ContextPie` disables its callback when absent. | Native policy/turn capability differs from Piclaw's supplied callback alone. |

Focused native runs: `@ux-compaction-00[1-5]` and `@ux-context-00[34]`
**42/42**; `make test-ux-context-meter` **18/18** (12 tagged);
`make test-ux-context-fit` **36/36** (12 tagged for compaction006/007);
`@ux-context-002|@ux-compaction-008` **12/12**. The broad
`make test-ux-compaction` run was interrupted at 91/102 and has no result.
All focused projects use Chromium/WebKit phone/tablet/desktop viewports.
Current Piclaw compaction UI/backend and physical-device acceptance remain
unprobed. No new deployed parity credit follows.
