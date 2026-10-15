# Classic model selection 020 and sparse metadata 022

The pinned Piclaw 3.2.4 source manifest marks the copied
`components/model-picker.ts` identical. Gi's mounted `ComposeBox` obtains the
selected session's catalogue and routes model changes through its native
session API; `web/src/app.ts` uses model revisions, mutation guards and
selection scope before applying refreshed model/context state. The installed
Piclaw model-switch backend was not exercised here.

| ID | Native tagged assertion | Boundary |
|---|---|---|
| `020` | `tests/ux/models.spec.mjs` accepts a session-scoped model PATCH, checks label/context response, unchanged draft/attachment and other session, reload persistence, keyboard selection and explicit rejected-model alert without a submitted turn. | Six Chromium/WebKit viewport projects passed under disposable Gi. |
| `022` | Sparse option omits context/reasoning badges; unknown usage appears as `?` rather than a measured count. A delayed catalogue response for a superseded session cannot replace current model or draft. | Six projects passed; the fixture's null usage is not a Piclaw measurement. |

The focused two-tag run passed **12/12**. Classic021 keyboard corrections and
its separate Piclaw oracle probe are recorded in `picker-keyboard-gap.md`;
they do not settle session-picker mutation or physical keyboard acceptance.
No production code or frozen text changed in this review.
