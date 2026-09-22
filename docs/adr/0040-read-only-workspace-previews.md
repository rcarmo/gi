# ADR-0040: Bounded read-only workspace previews

## Status

Accepted — 2026-09-22. Classic workspace-008 passes all six browser projects. Coverage is 44/236 Classic and 2/42 shared, with 192/40 unmapped. Editor, mutation and terminal preview contracts remain open.

## Host and native changes

The workspace file API returned only `{path, content}` and the host registered no preview extensions. The supplied explorer requires typed preview data and a registered renderer. `web/src/app.ts` now registers the existing Markdown and default read-only preview panes. Supplied components, panes and UI helpers are unchanged.

`internal/web/workspace_preview.go` provides path, kind, content type, size and modification time. Text previews include `text`, the legacy `content` alias and a truncation flag. Raster images link to native raw bytes; binary files receive metadata without a text body. Markdown extensions use `text/markdown`, HTML and SVG remain escaped source text. PNG/JPEG/GIF/WebP/BMP classification uses content sniffing rather than trusting an extension. Browser image evidence in this slice uses PNG.

`getWorkspaceFile` passes the requested byte limit. The native default is 20,000 bytes, capped at 256 KiB. Reads use a bounded sniff/UTF-8 window and preserve complete code points at the truncation boundary. Large files are not loaded in full for previews. The legacy missing-file status remains HTTP 500; invalid preview limits and rejected paths return 400. The endpoint rejects mutations with 405.

`/api/workspace/raw?path=…` supports GET/HEAD and byte ranges through a file handle. `download=1` forces attachment disposition; only sniffed raster types may be inline. Headers include sandbox/default-src-none CSP, nosniff and private/no-store. Both preview and raw routes use the existing instance authentication guard.

## Confinement

Filesystem access uses `os.OpenRoot`, so the actual open remains rooted across symlink resolution and concurrent renames. Lexical escapes, outside absolute paths, absolute-target symlinks, directories and known non-regular files are rejected. Relative in-root symlinks work. Absolute paths to ordinary files within the configured root are converted to local paths; absolute-target symlinks are rejected even if they point back inside the root. VFS/FTS previews are not provided here.

The browser tree and other tool APIs are unchanged; these rooted-open guarantees apply to the new preview/raw path only. The fixture initially wrote shell-generated files into the process working directory because the web shell helper does not set a workspace cwd. It now changes explicitly to the native runtime workspace. The three stray fixture files were removed; no shell-tool behaviour was changed.

## Acceptance and regressions

`tests/ux/workspace-preview.spec.mjs` creates files through native tools and selects actual explorer rows. It verifies:

- Markdown headings and emphasis through the supplied Markdown preview extension;
- literal HTML-like text in escaped code, with no injected image or script execution;
- PNG preview decode through the native raw URL;
- the binary download-oriented message;
- kind, extension, type, size, modified time and path metadata;
- SVG script source displayed as code;
- 20,000-byte truncation and retained composer text through selections/reload.

Selecting workspace files intentionally adds draft references under the existing host behaviour. This slice does not redefine selection or earn workspace-006/009 credit. Native tests cover UTF-8 limits, empty/large files, confined relative symlinks, outside paths/symlinks, directories, raw bytes/ranges/HEAD, download disposition, active-content headers, methods and post-enrollment authentication rejection. A functional browser test covers native metadata, raw download and Markdown preview rendering.

**480/480 browser executions:** 318 main (159 per browser family), 54 reconnect, 66 compaction, 18 Steer/admission, 12 context-fit and 12 context-meter. **73/73 functional**, **31/31 helpers**, full Go tests/vet/build/hook checks and web race tests ×3 pass. Frozen feature hashes are unchanged. Delegated read-only review timed out and supplied no evidence.

Desktop/tablet/phone screenshots were attached as `gi-workspace-preview-{desktop,tablet,phone}.png`; originals are under `/workspace/tmp/gi-ui-captures/september22/gi-workspace-text-*.png`. Binary/image test fixtures and browser artifacts are isolated under the test workspace and `test-results/ux-parity/`.

## Terminal adaptation

Use an explicit preview action opening a temporary, height-bounded read-only viewer. Show filename/type/size in that viewer, preserve literal text and existing Markdown projection, and represent binary/images with metadata plus an explicit capability-gated open/download action. Never allocate a permanent sidebar or image area. Escape must restore draft text/cursor and transcript reading position; viewer navigation must not submit or silently attach files. Retain Pi transcript padding and unchanged idle rows.

Independent acceptance at 60×18, 100×22 and 140×36 must cover long/Unicode text, unsupported formats, lookup failure, resize and dismiss restoration before terminal credit. No terminal code changed or PTY suite was rerun for this web-only slice.
