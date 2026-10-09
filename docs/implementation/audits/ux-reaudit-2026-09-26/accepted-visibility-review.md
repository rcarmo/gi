# Classic accepted-message visibility 007–011: native send/refresh trace

`tests/ux/features/classic/compose/instant-visibility.feature` scopes the
Piclaw contract to accepted messages and explicitly avoids a one-second
network-delivery promise. The pinned 3.2.4 manifest marks
`runtime/web/src/components/compose-box.ts` changed in Gi, so source is
compared at behaviour/branch level, not claimed byte-identical. Gi captures a
draft before clearing the composer, uploads media in sequence, builds text and
reference blocks, sends one native prompt, and calls `onPost` on an accepted
response. `web/src/app.ts` refreshes the selected view and merges persisted
posts with generation/scroll guards. Current Piclaw backend acceptance and
physical-device behaviour were not exercised for these cases.

| ID | Focused Gi browser assertion and code | Limit |
|---|---|---|
| `007` | Held accepted POST; stored user row appears once, acknowledgement separately triggers timeline GET, remains after reload. `compose-box.ts` calls `onPost`; `app.ts:809` refreshes current view. | Accepted fixture response only; not a delivery-time guarantee. |
| `008` | Multiline trimmed text + file/folder/message references become ordered native prompt blocks. A second reference-only send proves empty text can submit. | Gi's native prompt wire format is not Piclaw's `/agent/default/message` wire format. |
| `009` | Three media uploads keep exact byte, Unicode filename, returned ID and outgoing media/prompt association. | Native upload endpoint; Piclaw upload backend not run. |
| `010` | Held first acknowledgement while typing and attaching a new draft; response leaves new text, cursor and attachment intact through reload with one original user message. | Does not imply all unknown-delivery/crash recovery is solved. |
| `011` | Real persisted history plus later arrivals reconcile unique IDs, preserve reading anchor, and scroll only when near bottom. `app.ts` uses `mergeMessagePages`, revisions and scroll restore. | Emulated viewport/browser behaviour, not physical scroll parity. |

Focused command:
`make test-ux-parity UX_PARITY_PORT=19134 UX_PARITY_ARGS='tests/ux/drafts.spec.mjs --grep "@ux-compose-00[7-9]|@ux-compose-01[01]"'`
passed **30/30** across six Chromium/WebKit viewport projects. An earlier
attempt with an unquoted shell `|` did not run these tests (command-not-found /
EPIPE); only the corrected run counts. No frozen requirements or production
code were changed in this review. The queued-return `@ux-compose-004` policy
conflict remains separate.
