# Shared31–35: authoritative picker and capability state

Gi's mounted model selector, `web/src/gi-model-picker.ts` keyboard adapter and
session-scoped model/thinking APIs supply the native paths for these clauses.
The pinned Piclaw 3.2.4 model-picker source is byte-identical; the Gi-owned
keyboard correction is recorded in `picker-keyboard-gap.md`. This batch does
not rerun Piclaw model mutations or a physical keyboard.

| Shared cases | Focused native assertion | Result |
|---|---|---|
| `31`–`32` | Pointer and keyboard choose an advertised second model for the selected session, preserve unsent draft/references and persist the accepted model through reload. | 12/12 across six browser projects with `GI_UX_CONTEXT=1`. |
| `33` | Native session picker searches identifier/name/capabilities and uses incremental typeahead and enabled-entry navigation without unsupported mutation. | 6/6 in `session.spec.mjs`. |
| `34` | Model filtering, paging and modified Home/End respect text editing and enabled results, with one activation. | 6/6 with `GI_UX_CONTEXT=1 GI_UX_MODEL_PICKER=1`. |
| `35` | Thinking level and context controls follow reported provider capabilities and unknown/provisional usage labels; native provider request receives selected reasoning. | 6/6 with `GI_UX_THINKING=1`. Earlier deployed web-scope Shared35 evidence is separate. |

The four scoped commands are independent, not a whole-suite gate. Gi's native
fixture passed **30/30** cases in total. No current Piclaw backend inference,
physical input, TUI thinking parity, or production change follows from these
runs.
