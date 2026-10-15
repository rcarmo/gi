# Dependency upgrade — 2026-09-22

Gi now uses the go-ai `upstream-v0.87.0` release and current compatible Go/web dependencies. Frozen UX mappings remain 39/236 Classic and 2/42 shared; this upgrade adds no parity credit.

## Versions

| Dependency | Before | After |
|---|---|---|
| Go | 1.26.4 | 1.26.8 |
| go-ai | v0.80.10 | v0.80.11-0.20260921210723-c51fb076ad9f |
| goja | b07b74453ea9 | 793a2a65c13b |
| goldmark | v1.8.4 | v1.8.6 |
| x/crypto | v0.54.0 | v0.57.0 |
| x/net | v0.57.0 | v0.59.0 |
| modernc SQLite | v1.54.0 | v1.59.0 |
| Tailscale | v1.100.0 | v1.102.4 |
| Playwright / test runner | 1.60.0 | 1.63.0 |
| Preact | 10.29.2 | 10.29.8 |
| Marked | 18.0.4 | 18.0.14 |
| KaTeX | 0.17.0 | 0.18.7 |
| Babel runtime | 7.29.7 | 8.0.5 |

CodeMirror packages and applicable transitive dependencies are updated in `package.json`, `bun.lock`, `go.mod` and `go.sum`. `bun outdated` reports no remaining updates. Go 1.26.8 supplies the current patch toolchain; Tailscale requires at least 1.26.6. CI already reads the version from `go.mod`.

## go-ai release selection

`go get github.com/rcarmo/go-ai@latest` still resolves to July's v0.80.10 because newer releases use `upstream-*` tags. The explicit `upstream-v0.87.0` tag resolves to commit `c51fb076ad9f0207ba128af750d94fc40de9a121`, published on 21 September, and the pseudo-version above. No Gi provider adapter changes were required. This pins the release, not a moving branch head.

Local provider, inference, compaction, queue and session tests pass. Live external-provider calls were not made; upstream provider acceptance does not substitute for Gi's own integration evidence.

## Retained pins

- **go-tui v0.18.2:** v0.22.1 compiles after migrating off-screen `Render` calls to `RenderTo`, and fullscreen tests pass, but the regular-mode PTY test loses a completed response after a resize at 60×18. The failure repeats with and without Gi's resize marker and also reproduces on v0.19.0. The newer inline renderer clears rows above its relocated dock. Retaining v0.18.2 restores all three-size regular tests; the attempted API migrations were reverted. No terminal implementation changes are included.
- **gVisor v0.0.0-20260224225140-573d5e7127a8:** retain the version required by Tailscale v1.102.4. `go get -u ./...` selected gVisor HEAD `337254a73a61`, which fails Go compilation with mixed generated/template packages. Other selected dependencies remain upgraded.

## Asset and test corrections

`build.js` now copies KaTeX CSS, fonts and licence from the installed package, rewrites font URLs to `/fonts/katex/`, and embeds them with the JavaScript. The host HTML loads the matching stylesheet. Previously the stylesheet was stale, not loaded by the host, and referenced absent fonts. The new functional test checks the renderer version, stylesheet, every referenced font response and actual font loading. Supplied Piclaw component/style source files are unchanged.

The broader race run exposed two test synchronisation defects in `internal/web/server_test.go`:

- SSE fixtures read `httptest.ResponseRecorder.Body` while handlers wrote it. They now use a real HTTP test server and read complete stream events, waiting for connected readiness before publishing.
- The dispatcher fixture observed a message before its inbound item completed. It now waits for both observations, retaining the exact one-item assertion and existing deadline.

## Verification

- **444/444 browser executions:** 282 main, 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit, 12 meter. The final main run includes the KaTeX asset changes. An earlier final run was interrupted; its incomplete results were discarded.
- **71/71 functional tests**, **29/29 helper tests**, hook checks, full Go tests/vet and `go mod verify` pass.
- Race tests pass three times for TUI, turn, store and web. Web was rerun after the fixture repairs.
- Live session, compaction, reading, outcomes, regular, search and selection suites pass at **60×18, 100×22 and 140×36** on the retained go-tui version; TUI smoke and Gherkin pass. Idle layout and transcript padding remain unchanged.
- Pure-Go Linux and macOS builds pass for amd64 and arm64. Windows is outside the requested scope.

Logs are under `/workspace/tmp/gi-deps-*`; browser artifacts are under `test-results/ux-parity/`. The lightbox/media adapter work remains separately stashed as `WIP lightbox media adapter before dependency upgrade` and is not part of this release.
