# Session thinking selection

Settings → Models can apply a thinking level to future turns in the selected
session. Available choices come from the current model's go-ai catalogue. The
control is absent for unsupported models. Provider default leaves the reasoning
option unset; it does not mean a fabricated zero or an explicit off value.

## HTTP and storage

`GET /api/sessions/<id>/model` adds `thinking_levels`, `thinking_configurable` and
`thinking_token`. The displayed level is effective only when it is validated and
bound to the current model. Raw legacy strings are not shown as effective provider
configuration. Model identity uses the same configured fallback for GET, PATCH
and omitted-model prompt admission; native test-model/bootstrap sentinels remain
bare rather than becoming provider calls.

`PATCH /api/sessions/<id>/model` accepts either the existing model selection or:

```json
{"model":"provider/model","thinking_level":"high","thinking_token":"snapshot token"}
```

Thinking requires an exact canonical catalogue model, a currently supported level
(or empty string for provider default), and the current session-scoped token.
Invalid, null, unavailable and unsupported choices fail before mutation. Changed
model/thinking snapshots return 409. No prompt, provider call, global settings
write or other-session mutation occurs when applying a choice.

The token hashes session identity and relevant model/thinking fields, including
`selection_revision`. A SQLite writer transaction reads/checks the token and stores
`thinking_model`/`thinking_level` with an incremented revision. Selection changes
away and back invalidate old tokens. Unrelated status/queue-count writes do not.
Model Apply atomically resets thinking to provider default. A legacy writer that
sets thinking without a model binding clears the web binding, preventing a TUI
pass-through string from silently becoming a new effective provider configuration.

The Settings pane shares the existing model settlement invalidation and session
remount guards. Explicit Apply preserves draft/media; failure and late replies
cannot update another selected session. Lost acknowledgements are not replayed;
refresh reads native state. The supplied composer/model-panel source is unchanged;
its footer continues to display thinking read-only, with editing in Settings.

## Inference ownership

Admission snapshots a validated bound level into engine-owned
`selected_thinking_model` and `selected_thinking_level` metadata. Caller metadata
cannot override those fields. Queued turns keep their admitted choice even when
Settings changes later. Active steering cannot change the active turn's choice;
a separately staged continuation carries its admitted steering choice.

Each provider iteration reads the immutable turn fields and passes the level via
`StreamHooks.Thinking` to `goai.StreamOptions.Reasoning`. Validation is repeated
before transport use. The installed go-ai dispatches to the provider's simple
stream mapping when reasoning is specified. The special OpenCode Zen transport
has no thinking configuration and is not offered as configurable. Provider/API
errors remain errors; no fallback silently retries with another level.

Legacy/TUI-only unbound settings keep their previous inference-default behaviour.
This slice does not publish or deploy local-only TUI WIP `2a87a79`.

## Verification

- `make test-session-thinking`: race ×3 for store/inference/turn/HTTP. Covers
  session isolation, stale/foreign tokens, two-store CAS, rollback/reopen, ABA
  selection changes, legacy binding removal, exact catalogue levels, malformed
  requests, fallback/legacy identities and queued immutable turn capture despite
  forged caller metadata.
- `make test-ux-thinking`: 30 Chromium/WebKit cases across three sizes. Actual
  local OpenAI-compatible request bodies contain `reasoning_effort: high/low`;
  provider-default requests omit it. Tests cover explicit Apply, reload, no prompt
  on selection, draft/media retention, unsupported and stale writes, foreign
  sessions, bare default model, lost acknowledgement and delayed old-session reply.
- `make test-ux-model-panel`: 42 passes.
- Canonical parity fixture with `tests/ux/gi-settings.spec.mjs`: 108 passes.
- `make test-ux`: 139 passes, 11 existing skips.
- `make test-web-basic-controls`: 36 passes.
- `make test vet bun-checks test-web-http-helpers`: passed (24 browser helpers,
  119 assertions). Supplied components/ui/panes are unchanged.

Review corrected GET/PATCH fallback disagreement and bare-model capture mismatch;
final scoped review approved. Initial fixture API/type mistakes and a Settings
assertion variable typo were corrected, with failures retained in
`/workspace/tmp/gi-session-thinking`. A timed-out review is not approval. No browser
timeout or assertion tolerance increased. A separate required ten-minute CI job
runs thinking acceptance rather than extending an existing job's budget.

CI `36254114983` for initial source `715dfb0` passed the thinking ownership job
but failed the report helper: one test with 102 CLI subprocess calls exceeded
Bun's default five-second limit. The helper is now split into five independent
cases. All assertion lines and subprocess calls are preserved; seven report/ledger
tests pass with 1954 assertions. No runtime code or timeout changed for this fix.

Whole-product CI and exact-source deployment are still required. Shared35 remains
unmapped: this fixes the thinking prerequisite but does not resolve its positive
local-estimate labelling clause. Context stays measured-provider or unavailable;
no synthetic estimate, physical-device or exact-pixel acceptance is added.
