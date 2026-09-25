# Composer and picker geometry

Gi uses a pinned Piclaw Classic reference for composer side padding and the outer
bounds of session/model pickers. This slice changes host CSS and inserts two
mobile Close controls through a guarded build adapter. Supplied component and
stylesheet files remain unchanged.

## Reference

`tests/ux/fixtures/picker-geometry-reference.json` records Piclaw commit
`0afe5366ced9bca8246abd99a0feb1875a6ffbcc`, installed release
`piclaw-3.2.2-linux-x64-baseline` and the live Classic asset identifier
`15958f2c3dc9`. The installed Classic chat/responsive CSS matches the inspected
source reference. The Piclaw working checkout was dirty on another branch and
was neither used as the build source nor changed.

- `chat.css` SHA-256: `b6d3da5e2b0aff389efe095a20fe7388c37fe679bfa1ba04bc12ca4f4680597b`
- `responsive.css` SHA-256: `752164b9406576add5fd160569259414a46f2551f429ffbf5dd59fddf25e1b82`

Read-only Chromium measurements compared the running Classic UI and Gi at widths
390, 639, 640, 820 and 1440, height900. Mutating requests were blocked. Inventories
and histories differed, so their row counts and content heights are not pixel
comparison oracles.

| Rule | Reference | Gi before | Gi after |
|---|---|---|---|
| Composer side padding | 0 | 12px mobile / 16px desktop | 0 |
| Desktop session width | Composer anchor | At most440px | Composer anchor |
| Desktop popup gap | 6px above anchor | 6px | 6px |
| Desktop model width | min(680px, viewport−24px) | Anchor width | Reference width, capped to anchor |
| Mobile breakpoint | At most639px | Absolute anchored picker | Fixed panel |
| Mobile bounds | 8px inset, safe-area-aware top/bottom | Small panel above input | Reference bounds |

At640px, the reference model popup measured x80+width616=696, overflowing the
viewport. Gi deliberately caps that intermediate-width popup to the available
composer anchor. The fixture records this containment difference separately.
At820/1440, model width matches680px; mobile widths match viewport−16px.

## Interaction

`internal/web/static/css/gi-picker-geometry.css` loads after the app bundle. Lists
shrink and scroll within the mobile panel; search and action controls have at
least44px height. Search text and focus survive resizing through639/640px.

`patch-picker-geometry.mjs` only adds mobile Close controls. It rejects missing,
duplicated or already-patched anchors. Session dismissal invokes the existing
`closeSessionPopup(true)` path; model dismissal closes the popup and restores its
trigger if focus returned to the body. Escape and existing keyboard selection
remain in the supplied handlers.

Fixed mobile pickers cover footer controls. Model regression tests now dismiss
the model picker with its real Close button before opening the session picker.
Held responses, mutation ownership and stale-result assertions remain unchanged.
The old shared23/24 absolute-mobile geometry assertion now checks fixed mobile
bounds and retains the existing desktop anchor, insertion/first-frame focus,
query/resize, Escape and draft checks.

## Validation

`make test-ux-picker-geometry` runs four tests across six Chromium/WebKit projects:
24 executions. Twelve executions create explicit390px touch contexts, so those
are repeated mobile coverage, not additional tablet/desktop coverage. The other
twelve start at each configured project size and resize across639/640/390/820.
Tests cover relative desktop bounds, fixed-mobile edges, overflow, long session
lists, trusted touch dismissal, keyboard Close, Escape, focus/query/draft
preservation and no post-fixture mutations. The functional suite also checks
mobile bounds and dismissal.

The geometry target runs after startup/Return in the required browser CI job.
Local validation passed24geometry,42startup,126session,36model,30shell/workspace
motion,6model-filter and111functional cases (five existing skips), plus full
Go/vet/hooks and109helpers/2,988assertions. The initial combined session/model run
timed out after exposing obsolete mobile click-through assumptions; completed
split runs provide regression evidence. A focused review found no blocker.

This is outer-geometry acceptance only. Current Piclaw's session strip, catalogue
header/footer/grouping, row styles, desktop content heights, font/theme details
and full visual equivalence still differ or lack verification. Physical safe
areas, mobile keyboard behaviour and Visual skin require separate checks.
Frozen feature files and mapping counts are unchanged. No terminal controls or
idle rows are added.
