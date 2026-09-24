## Piclaw read-only pane host

Gi's workspace tabs support a **Piclaw-compatible read-only preview subset**. They do not load arbitrary extensions, provide an editor host or promise compatibility with every built-in Piclaw pane.

The comparison target is Piclaw `bfc34e4ebfe9b0ce0deefa6202a9d9a4780a5322`, under `runtime/web/src/panes/`. Gi's `pane-types.ts`, `pane-registry.ts` and `workspace-preview-pane.ts` match that revision byte-for-byte; the support tests pin their SHA-256 hashes. This is a separate target from the frozen UX baseline `70d33bc93ab540845bbcf5f80503ca8125c71594`, whose `dirty` and `retainOnTabSwitch` fields are absent here. Neither baseline nor supplied pane source was changed for this slice.

### Supported boundary

| Operation | Gi behaviour |
|---|---|
| Resolve | Existing Piclaw registry priority ordering; highest selected pane must declare `tabs` placement and only `readonly`/`preview` capabilities. Empty, unknown, edit, terminal and mixed capability lists fail before mount; no silent fallback. |
| Context | Native file GET with `max_bytes=20000`; `path`, `mode: 'view'`, original `preview`, `mtime` and full file `size`. Text previews also populate `content` with the bounded UTF-8 text. Binary/image previews do not invent text. |
| Truncation | `context.content` may be partial; `context.preview.truncated` is authoritative. File size is not the length of the returned text. |
| Resize | Initial `resize()` after close registration, then container `ResizeObserver` notifications. Without that API, window resize is the fallback. |
| Pane close | `onClose` fences callbacks and disposes before requesting closure of the captured tab. Duplicate and stale callbacks do nothing. |
| Refresh / tab switch | Dispose and remount, with late reads fenced. Refresh deliberately does not use `setContent`, which cannot carry the complete image/binary/truncation metadata. |
| Cleanup | Disconnect observation, dispose once, clear host. A throwing extension dispose cannot prevent host clearing; close-registration or resize failures dispose and expose Retry. |
| Focus | Asynchronous file completion never calls `focus()`: it must not take focus from newer composer or Settings interaction. Existing preview DOM remains focusable; final tab closure uses the shell's guarded composer restoration. |

The only production registrations are the supplied workspace Markdown and default preview panes. Capability declarations are a compatibility check, not a security sandbox: registered pane JavaScript is trusted. A pane that throws before returning an instance must clean up any non-DOM resources it allocated; the host can only clear its container.

`mountWorkspaceTab` is an internal host helper, not a public extension API. `WorkspaceTab` owns cleanup-before-remount and supplies its Preact state setter and the shell's synchronous tab-close handler. Those host callbacks must not throw or synchronously mount another owner in the same container. Extension lifecycle exceptions are isolated; arbitrary host callback failure is not promised. The independent review raised that distinction, including the toolbar's host-owned close path; broad callback sandboxing was not added.

Dirty/save callbacks, automatic focus handoff, retained hidden instances, dock placement, detach/pop-out, transfer hooks, editor/terminal/VNC backends and generic extension loading are unsupported. Copying their source interfaces does not implement them.

### Verification

`tests/ux/support/workspace-tab.test.ts` covers capability rejection, text/binary context, late responses/errors, synchronous close during registration, throwing hooks/disposal, exact listener cleanup and pinned source bytes. The suite first failed the three new lifecycle/capability cases, then passed after the host changes.

`tests/ux/pane-host.spec.mjs` bundles the real `WorkspaceTab` with a test-only Piclaw extension. Only the test bundle is intercepted; file reads and refresh use the isolated native Go server. Chromium and WebKit at phone/tablet/desktop sizes check exact bounded text/metadata, resize delivery, disposal ordering, pane-requested close, stale close suppression, unsupported capability errors, no asynchronous focus stealing, no browser API writes and unchanged draft/session turns. The fixture is never embedded in production assets. Existing workspace-tab tests separately check the real supplied renderers, Settings focus races, failure/retry and draft/media preservation.

Validation on 2026-09-24: 24 focused browser cases, 174 workspace/shell/settings regression cases, 92 functional cases, 80 support tests (963 assertions), Go tests/vet and hook checks passed. A broader run interrupted at case 67 was discarded and rerun from a fresh instance. Diagnostics found no issues but reported the optional oxlint validator unavailable.

No new frozen criterion is mapped: **87/236 Classic and 27/42 shared**, with **149/15** unmapped.

### Terminal adaptation

A web pane is not a terminal interaction model. Read-only file inspection should use a temporary bounded preview reached through an existing compact picker/action, with Escape restoring the draft, logical cursor and reader position. Do not add a persistent tab strip, dock or idle status row. Binary content should offer a truthful path/type/size summary rather than pretending to render a web viewer.

This is design only. Before terminal credit, independently test regular/fullscreen at 60x18, 100x22 and 140x36, including truncation, failure/retry, resize, stale occurrence rejection, Unicode/multiline drafts and unchanged history/settings/session ownership. This web-only slice changes no terminal code or footprint.
