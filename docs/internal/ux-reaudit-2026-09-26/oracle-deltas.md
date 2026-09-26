# Current Piclaw oracle versus pinned Gherkin

Current oracle: installed `piclaw-3.2.4-linux-x64-baseline`
(`/opt/piclaw/current`), asset version `990f0c49a932`, source map SHA-256
`c53092cfb75415f30b4c6b16b758cab68fed9bbc546953e5916a91b382acb32b`.
The release manifest names no source commit. Source excerpts below come from
its own `app.bundle.js.map` `sourcesContent`, not the modified local Piclaw
checkout. The four relevant compose, Quick Actions, Markdown and SVG source
files have identical hashes to the earlier shipped 3.2.3 map. The 3.2.3 release
was removed after the initial probe; its versioned reference is retained only as
historical evidence. The older frozen Classic corpus was cut at `70d33bc93ab...`
(`tests/ux/upstream/PROVENANCE.txt`). The shipped 3.2.2 pixel reference was
`0afe5366...`, not an acceptance claim for 3.2.4.

`make test-piclaw-oracle-basic` serves byte-hash-checked 3.2.4 UI assets with
fixture-only `web:default` responses and no real chat/auth writes. It reproduces
the server's normal `sanitizeSvgFences=true` HTML substitution from
`runtime/src/channels/web/http/static.ts:107` and the default in
`runtime/src/core/config-web.ts:293`. Evidence:
`test-results/ux-oracle/piclaw-3.2.4/browser-probe.json` and screenshots.
It does not exercise Piclaw's live DB, production authorization, physical touch,
passkeys, keyboard assistive technology or Gi's runtime.

| Contract | Piclaw 3.2.4 observation and source | Existing Gherkin/test conflict | Resolution for this audit |
|---|---|---|---|
| Quick Actions slash prefill | Clicking `/model` replaces `ORACLE EXISTING DRAFT` with exactly `/model`, focuses textarea, sends no prompt. Shipped `components/timeline-quick-actions.ts:365`, `app.ts:126-130`, `components/compose-box.ts:1446-1465`. | Frozen Classic `@ux-original-007` requires replacement **with a trailing space**; `tests/ux/quick-actions.spec.mjs:40-51` asserts `/model `. Shared16/17 require **preserving** the previous draft. | Versioned oracle-only scenario records actual `/model` replacement. Do not flip Gi's durable draft handling just to earn parity: replacing user text is destructive and needs an explicit policy decision, a migration story and regression gates. No new Classic/Shared credit. |
| Return queued item | Clicking Edit in compose replaces `NEWER UNSENT TEXT` with `QUEUED ORACLE TEXT`, focuses textarea and POSTs `/agent/queue-remove` with `row_id:101,chat_jid:web:default`. Shipped `components/compose-box.ts:929-985`. | Classic `@ux-original-017` and `@ux-compose-004` replace/clear media, while Shared28 requires merging latest concurrent draft/media before DELETE; Gi's transactional recovery test `tests/ux/queue-return.spec.mjs:26` follows Shared28. | Keep both pinned reference and safety contract visible. Versioned oracle scenario describes replacement, but fixture does not establish media/failure semantics. Do not replace Gi merge with lossy behavior without explicit approval. |
| Fenced SVG | Safe SVG renders as `<img class="model-svg-image" src="data:image/svg+xml;base64,..." alt="Oracle SVG">`; source remains in disclosure for copy. A fence with script, `onload` and external image URL stays escaped code, with **no preview, script execution or external request**. Shipped `markdown.ts:684-688`, `utils/svg-images.ts:74-225`; config default noted above. Raw model geometry is not an inline privileged SVG DOM child. | Frozen Classic `@ux-original-029` and `tests/ux/rendering.spec.mjs:54-61` require source-only; Shared41 asks for inline SVG and stripping unsafe elements before rendering. At `70d33bc` `markdown.ts` had no `renderSvgFences`; `f8bfbd25e` later added the safe image path. | Versioned oracle scenario states safe-image plus unsafe-source fallback. Neither frozen scenario can be claimed from the other. A full SVG security suite needs malformed/oversize/external-resource/theme cases before Shared41 credit. |

## Other source-backed re-audit leads

- `tests/features/settings/gi-settings.feature:40` said thinking was read-only;
  `web/src/gi-settings-models.ts:117` and `tests/ux/session-thinking.spec.mjs:18-35`
  expose/apply native supported choices. The additive Gi feature is corrected with
  explicit choice and model-switch reset clauses; the six-project focused test
  for `@gi-settings-029` passed. This does **not** alter frozen Classic/shared files.
- `features/search/workspace-index.feature:33` asks a default background refresh,
  whereas lines `199,211` require GET to stay read-only. Gi
  `internal/web/workspace_index.go:138-141` rejects `GET ?refresh=...`; Piclaw
  `runtime/src/workspace-search.ts:101-120` requests background work by default.
  This is a policy conflict across native search and web contracts. Do not
  silently make GET mutating or weaken the transactional-index safeguards.
- A six-project Playwright matrix uses viewport sizes, not physical touch, iPad,
  hardware passkey or actual reduced-motion preference by default
  (`playwright.ux.config.mjs:3-10`). Presence of such a tag is not physical proof.

The full 358-definition/421-case audit is still underway. These findings are
verified **slices**, not an exhaustive result or a full parity report.
