# ADR-0014: Session-local model selection

## Status

Accepted — 2026-09-22

## Context

The model picker sent `/model …` through ordinary prompt submission. The HTTP handler created agent turns for model commands, and later prompts defaulted to the global model. Session-scoped display reads alone did not provide a working mutation path.

## Decision

Provide `GET/PATCH /api/sessions/{id}/model`. GET returns the native configured/provider catalogue, authoritative current selection, supported thinking flag and catalogue context-window metadata. PATCH validates the requested label or unambiguous ID, rejects unknown/unavailable/uncredentialled entries, persists selection only in that session and returns the authoritative state. A selection does not contact the provider, create a turn or write global settings.

The prompt handler recognises `/model` and `/model <provider/model>` before turn admission. It uses the same validator and returns command state/errors. The configured `test/test-model` and `test/bootstrap` entries are explicit deterministic-shell models in the isolated test catalogue; arbitrary synthetic test entries remain unusable. Real providers require credentials and a registered context-window entry. Credential availability does not guarantee that a remote service will accept the next request.

Persist `selected_model` and `selected_provider` separately from runtime `model`/`provider` state. An old turn can finish with its captured model and update runtime metadata without replacing the user's selected model. New prompts without an explicit override prefer the selected model. Non-reasoning selections clear the session thinking level. No in-flight turn is retroactively switched.

The composer picker calls PATCH directly, retains draft/media/reference state, shows errors inline and closes only after acceptance. A mutation lock prevents overlapping activation. Host revisions reject model polls captured before or during a mutation; component mount/revision guards reject old picker catalogue results and late replies after A→B→A.

Native Enter on a focused model button activates that button. Tab traverses controls instead of selecting the highlighted entry. Initial popup focus uses a layout effect; subsequent selection/list changes scroll without reclaiming focus. The Next model button uses the same validated path. Removed the obsolete prompt-based picker helper.

The SSE hook now disconnects on `pagehide` and reconnects on visible return. A WebKit reload trace showed document load but uncompleted asset requests while the page remained at Loading Gi; it did not establish a model-state failure. After explicit navigation cleanup, the complete matrix passed. No retries or alternate navigation were added to mask that failure.

## Evidence

- `@ux-original-021`: held real accepted model response, switch to another session, unchanged target draft/model, then authoritative origin state on return.
- `@ux-compaction-008`: native model command selects configured model; unknown command reports an error and creates no turn.
- Gi regressions: pointer/keyboard selection, reload persistence isolated to the origin, draft/media retention, unusable-model rejection, A→B→A delayed catalogue and late failed mutation isolation.
- `TestSessionModelCommandsAreValidatedLocalAndDoNotCreateTurns`: native API/command validation, session/global isolation, old runtime-model overwrite protection and actual next-turn model selection. Three race-detector runs.
- Full matrix: 168/168 executions, 14/236 Classic IDs passing, 222 unmapped. Existing functional suite: 70/70. All 42 shared-contract cases remain unmapped.

## Limits and terminal adaptation

Measured context usage is unavailable (`context_usage: null`). No fake usage or context-fit compatibility is supplied. Therefore `020`, compaction `006/007`, and shared model/context cases remain unmapped. Full picker search/grouping and thinking mutations also need separate acceptance work.

The terminal should use the same validated session-local selection policy through its existing bounded `/model` selector and footer model segment. It must retain drafts, reject unknown choices and avoid writing global configuration for a session-local action. The subsequent terminal adaptation in [ADR-0015](0015-terminal-session-model-selection.md) implements that shared policy and closes the global-default-write gap for `/model`, picker and cycle actions. No sidebar, title bar, context row or other idle chrome is needed.
