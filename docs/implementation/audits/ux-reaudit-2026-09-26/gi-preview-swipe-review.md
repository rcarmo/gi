# Gi-native preview and swipe 001–003: clause review

These six scenarios in `tests/features/sessions/gi-preview.feature` and
`gi-swipe.feature` are **Gi additions**, not frozen Classic or Shared Piclaw
contracts. Review compares their native assertions and handlers. Piclaw 3.2.4
web behaviour and physical touch are outside this Gi-specific evidence.

| IDs | Native assertion → code | Limit |
|---|---|---|
| `@gi-preview-001` | `tests/ux/thoughts.spec.mjs:62–90` expands a thought preview, switches sessions, starts another turn and reloads, checking that stale disclosure/buffer state does not leak and unsent drafts remain owned. `web/src/ui/app-agent-previews.ts` derives preview state per captured turn. | Does not prove provider thought content semantics. |
| `@gi-preview-002` | `thoughts.spec.mjs:20–61` tests streamed thought and draft disclosure, text retention, scroll/resize, Escape scope and composer preservation under both preview variants. | Test combines several Classic thought tags, but those tags are not Gi-preview oracle credit. |
| `@gi-preview-003` | `thoughts.spec.mjs:91+` uses wrapped paragraphs, resizes, streaming and a variant with `ResizeObserver` removed. `web/src/gi-preview-overflow.ts` measures actual DOM clipping and rechecks on content/size/font events. | Viewport emulation is not physical-device acceptance. |
| `@gi-swipe-001` and `002` | One `tests/ux/session.spec.mjs:847–909` case reverses quickly while catalogue and old timeline responses are held, then checks committed selection, draft/media ownership, no late timeline leak and turn counts. `web/src/ui/chat-swipe-navigation.ts` supplies candidate/gesture selection; chat-scoped readers fence late responses. | One combined test serves both IDs. It does not prove physical OS gesture handling. |
| `@gi-swipe-003` | `tests/ux/status-swipes.spec.mjs:16+` tests thought-panel link passthrough plus text-selection, direction, input and Settings exclusions; native turn completion and draft retention are checked. | Same factory also runs `@ux-mobile-003` on a separate draft-panel variant; neither grants real-device touch acceptance. |

`make test-ux-thoughts` first passed 47/48: Chromium tablet `@gi-preview-001`
failed during startup with only `Loading Gi…` visible, before the preview
assertions. An attempted direct Playwright rerun lacked its server and failed
with `ECONNREFUSED`; it provides no product evidence. A focused disposable
server run of all three Gi preview tags passed **18/18** across six projects.
The combined rapid-swipe case passed **6/6** via `make test-ux-parity`; the
status-swipe target passed **12/12** across two variants, including six
`@gi-swipe-003` cases. No assertions or timeout were weakened. The startup
failure cause is unknown and the full thoughts target has no green result.
