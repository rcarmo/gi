# Shared 1–42: source/test audit map

The shared corpus has 42 expanded cases from 29 scenario definitions. The
historical report marks 33 mapped and nine unmapped (`8`, `16`, `18–22`, `40`,
`41`). A mapping means a tagged Gi test exists; it does **not** establish current
Piclaw 3.2.4 behaviour, physical-device acceptance or a full suite pass. The
current oracle was directly probed only for first send, Quick Actions slash and
fixture-listed skill insertion, model-picker keys, queue return and SVG (six
isolated browser projects). All other rows
need current-release comparison before being called oracle-aligned.

| IDs | Gi evidence / source | Current audit boundary |
|---|---|---|
| 1–3 | `tests/ux/classic.spec.mjs`, `workspace-preview.spec.mjs` | Menu focus, dismissal/no underlying send and workspace toggle have native web tests; no current-oracle pixel/physical touch claim. |
| 4–7, 9–15 | `tests/ux/quick-actions.spec.mjs` | Native timeline typing, grouping, key ownership, dismissal and focus are covered in combinations. Trustworthy target typing is distinct from synthetic modifier/composition guard testing. |
| **8** | `quick-actions.spec.mjs:65-84`, `web/src/ui/app-browser-events.ts:6-16` | **Unmapped:** artificial contenteditable DOM proves a selector guard, not a real editor receiving text. Add a native editable surface journey. |
| **16** | `quick-actions.spec.mjs`, mounted `web/src/app.ts`/composer | **Unmapped/contract conflict:** Shared16 requires supported-command insertion to preserve an existing draft. Classic007 explicitly requires replacement, and mounted `/model` follows Classic/Piclaw; its focused test starts empty and cannot settle that conflict. The skill-only adapter satisfies a narrower Shared17 path. Choose a priority policy before changing general prefill; neither incompatible clause gets borrowed credit. |
| 17 | `tests/ux/skills.spec.mjs`, `build.js` → `scripts/patch-skill-prefill.mjs`, `tests/ux/oracle/piclaw-basic-probe.mjs` | The built canonical `/skill:<name>` adapter prepends command plus existing text. Focused Chromium-phone recheck passed 3/3 with loaded fixture, captured-session execution and retained draft/media. Piclaw 3.2.4's fixture-listed `/skill:proof` replaces text; that oracle probe establishes only UI prefill, not loaded-skill execution. Gi's skill path is a deliberate no-loss difference, not general-command parity. |
| **18–22** | `plan-oracle-gap.md`; installed Plan-sidebar add-on 0.1.25 | **Unmapped:** Gi has no Plan store/tool/sidebar. Gi's target keeps revision CAS and Shared20 dirty-refresh confirmation as a deliberate no-loss deviation from the installed add-on's Markdown/`updated_at` and unconfirmed discard. No native implementation or parity credit. |
| 23–26 | `tests/ux/session.spec.mjs` | Picker focus, session coherence and guarded mutations have native tests; seeded sessions and viewport projects do not prove current Piclaw pixel layout. |
| 27–30 | `tests/ux/queue.spec.mjs`, `queue-return.spec.mjs`, `queue-steer.spec.mjs` | Gi queue IDs, return CAS, reorder, removal and exact-run Steer are tested. ADR-0017 accepts Shared28's durable recovery and no-loss merge; it deliberately differs from Piclaw's queued Return replacement. Classic replacement remains unmapped. |
| 31–35 | `tests/ux/models.spec.mjs`, `model-panel.spec.mjs`, `session-thinking.spec.mjs`, `context-meter.spec.mjs`, `compaction.spec.mjs` | Native model/estimate/thinking tests are bounded web evidence; Classic021's model-picker Page/Home/End gap has a separately tested local audit-branch correction; this shared group alone does not earn the frozen Classic ID. An isolated six-project current Piclaw picker-key probe covers only search navigation and Escape. |
| 36 | `tests/ux/reconnect.spec.mjs`, `docs/internal/web-stop-queue.md` | Captured web Stop and explicit Resume preserve queued work; generic/TUI cancellation differs. |
| 37 | `tests/ux/shared-copy-delete.spec.mjs` | Bounded single-message copy/delete; does not earn Classic cascade-delete replies. |
| 38 | `tests/ux/message-retrieval.spec.mjs`, `docs/internal/message-retrieval.md` | Durable numeric IDs and session-only model tool evidenced in six projects; Classic all-chat/family authorization remains unmapped. |
| 39 | `tests/ux/drafts.spec.mjs` and attachment tests | Native upload/cancel/retry/reload evidence is scoped; physical file-picker acceptance separate. |
| **40** | `tests/ux/tool-activity.spec.mjs`, separate WIP500e02f | **Unmapped:** Gi status row is not an accessible expandable persisted output pane. Terminal-provenance WIP is not deployed or approved. |
| **41** | `oracle-deltas.md`, `tests/ux/rendering.spec.mjs` | **Unmapped/version conflict:** Piclaw 3.2.4 safe SVG is a sanitized data-URL image with inert source fallback; frozen Shared41 asks for inline SVG and stripping unsafe content before rendering. Gi still renders fenced source only. |
| 42 | `tests/ux/speech-contract.spec.mjs` | Browser speech/copy controls are bounded. Physical audio and assistive-technology checks separate. |

## Priority corrections

1. Do not merge draft-destructive Piclaw behaviour into Gi without explicit
   no-loss policy approval. Keep versioned oracle Gherkin separate from historical
   and strengthened safety contracts.
2. Review/CI the local Classic021 model-picker correction and check the frozen
   session-picker half and native mutation separately. The isolated Piclaw
   keyboard probe has no live backend.
3. Implement revisioned Plan as a real native store/tool/UI subsystem only if
   Shared18–22 remain required; current Plan add-on behaviour is not a CAS oracle.
4. Complete real contenteditable, accessible tool-pane and adversarial SVG
   journeys before adding IDs 8, 40 or 41. Document physical-device exclusions.
