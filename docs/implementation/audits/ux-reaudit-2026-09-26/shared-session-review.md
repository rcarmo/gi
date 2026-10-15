# Shared23–26: picker focus, coherent selection and native actions

Gi's mounted `web/src/app.ts` switches the selected session synchronously,
invalidates timeline/model/queue generations, and sends pin, rename, archive
and restore through native mutation handlers. Copied
`ui/compose-session-switcher.ts` is byte-identical in the pinned Piclaw 3.2.4
source manifest. The current Piclaw UI was not exercised for these clauses.

| Cases | Native tagged assertion | Boundary |
|---|---|---|
| `23`–`24` | Pointer and keyboard opening give search focus on the first visible frame, keep the popup anchored, filter native identifiers and restore trigger focus on Escape. | `session.spec.mjs` six browser projects each. |
| `25` | Delayed main-session responses cannot replace research timeline, queue, model, context, draft or attachments after switching. | `context-fit.spec.mjs` six projects with `GI_UX_CONTEXT=1`. |
| `26` | Supported native session mutations succeed; rejected and idempotent actions keep the picker/selection recoverable without enabling unsupported delete. | `session.spec.mjs` six projects. |

Focused results: **18/18** for 23, 24, 26 and **6/6** for 25, across
Chromium/WebKit phone, tablet and desktop. A first attempt to run 25 without
`GI_UX_CONTEXT=1` found no tests and earned no pass; the corrected run above
passed. These fixture results do not establish Piclaw 3.2.4 browser parity,
physical keyboard behaviour or all session-tree actions. No frozen contract
or production code changed.
