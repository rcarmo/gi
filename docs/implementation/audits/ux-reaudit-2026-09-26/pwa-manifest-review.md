# Classic PWA icon clauses 001–006: audit disposition

The shipped Piclaw 3.2.4 oracle is
`/opt/piclaw/current/app/runtime/src/channels/web/manifest.ts` and
`http/dispatch-shell.ts`. These server files are readable in the installed
release; they are not part of the Classic UI source-map diff. A disposable
installed-helper probe exercises manifest avatar/cache and shell-icon
PNG/fallback decisions. No Piclaw avatar mutation or home-screen
installation was run. Gi tests use a disposable native web server, not the live
chat/auth database.

| Frozen ID | Current Piclaw source and Gi evidence | Disposition |
|---|---|---|
| `@ux-pwa-001` | Piclaw's default manifest has static 192/512 PNG icons (`manifest.ts`). Gi `tests/ux/pwa.spec.mjs:4–36` gets `/manifest.json`, checks name, icons, sizes, type/purpose, PNG bytes, `HEAD` and the HTML manifest link. Gi `internal/web/server.go:1127–1166` serves the static manifest. | **Bounded default-static pass.** The fixture has no configured avatar; it does not establish the avatar branch. Six Gi browser projects passed. |
| `@ux-pwa-002` | Installed Piclaw 3.2.4 `handleManifestRequest` with a disposable avatar cache returns four 192/512 PNG `/avatar/agent` icon records; `any` and `maskable` variants encode `format=png`, size and version, and `piclaw_avatar` references the first icon. Gi `serveManifest` hard-codes `/static/icon-192.png` and `/static/icon-512.png`; there is no agent-avatar manifest branch. | **Gi capability gap**, untagged and not implemented. Callback evidence does not establish avatar storage, generated PNG bytes or an installed home-screen icon. |
| `@ux-pwa-003` | Installed helper returns the same four static icons for an absent avatar, a null cache result or a thrown cache-preparation error. Gi's static-only manifest and `001` native test show the no-avatar branch. | **Bounded default-static evidence.** No live avatar-configuration transition or cache failure was exercised. |
| `@ux-pwa-004` | Installed `handleShellRoutes` with disposable callbacks tested all four Apple paths: each requests agent `format=png` at 180/167/152/180 respectively; HTTP 200 PNG wins, while HTTP 200 WebP or 404 PNG uses the path-matched static asset. Gi's `internal/web/static/` has four static routes and an untagged browser test checks 200 PNG signatures. | **Avatar/fallback-decision gap in Gi:** no avatar-first handler or native per-size branch. The installed probe checks helper calls and synthetic response selection, not image conversion or physical home-screen results. |
| `@ux-pwa-005` | Installed helper requests an agent `format=png&size=48` favicon; disposable 200 PNG wins, 200 WebP or 404 PNG selects static `favicon.ico`. The untagged Gi browser test checks only a nonempty static favicon response. | **Avatar/fallback-decision gap in Gi.** The frozen note does not require a PNG-decoded fallback; this probe does not validate image bytes. |
| `@ux-pwa-006` | Installed helper with cache revisions `version A` and `version B` changes all four avatar icon `v` values; GET/HEAD content lengths also matched under identical metadata. Gi's static manifest URLs have no avatar version. | **Gi capability gap**, untagged and not implemented. No actual cache mutation or OS icon refresh was observed. |

`make test-ux-parity UX_PARITY_PORT=19134
UX_PARITY_ARGS='tests/ux/pwa.spec.mjs'` passed **12/12** across Chromium/WebKit
phone/tablet/desktop: six tagged `001`, six untagged static Apple/favicon route
checks. The test and source trace qualify `001` and the static half of `003`.
`make test-piclaw-pwa-icons` passed the installed 3.2.4 manifest helper's
static/avatar/cache/HEAD branches and **15/15** shell-icon path/response cases
(5 paths × PNG/WebP/404). The helper probe uses disposable callbacks. Neither
suite tests Piclaw avatar storage, Gi dynamic avatars, generated PNG content,
physical home-screen selection or cache invalidation. Gi needs native avatar
handling and adversarial PNG/non-PNG/fallback tests before mapping `002`,
`004`, `005` or `006`.
