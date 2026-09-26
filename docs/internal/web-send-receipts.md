# Bounded web send receipts

The prompt HTTP handler now stores its exact successful native `SubmitResult`
under the source session and `client_request_id`. This happens after synchronous
submission returns, when subturn rollback is no longer possible, and before the
202 response is written. It covers local, routed, explicit-target and active-turn
steering results without searching other sessions.

## Storage and lookup

`web_send_receipts` is an additive table with an index on source session and
request ID. Every successful prompt request with a token appends a row. Reusing a token does not
coalesce submissions or select the newest result: multiple rows are ambiguous.
Receipt storage failure is logged but does not turn accepted work into an error.
Receipt rows are removed with their source session; no receipt expiry/GC is introduced.

`GET /api/sessions/<source>/send-receipt?client_request_id=<token>`:

- Accepts exactly one token query and GET only; rejects empty/control/invalid
  UTF-8 or over-128-byte tokens and extra/duplicate query keys.
- Uses the existing authenticated session-route boundary and an indexed lookup
  of at most two rows. It returns no prompt, message or attachment contents.
- Returns `confirmed:false` for missing/duplicate receipts or a missing/deleted
  target. It rechecks the stored turn's session and the result's source ownership.
- Returns the exact native result only for one valid receipt. Routed fields
  preserve target session/agent identity; steering preserves its active-turn ID.
- Sends `Cache-Control: no-store`. Results are capped at 4096 bytes when recorded.

The browser uses this endpoint for live-page and persisted-capture recovery.
Startup still checks at most six captures with two workers and a 3-second overall
network deadline, now using bounded receipt responses rather than whole turn
histories and event lists. No fallback history scan or text match remains.

## Limits

This is reply recovery, not POST idempotency. If two requests with one token are
admitted, lookup returns unknown. A crash after native admission but before
receipt storage also remains unknown. Old-version admissions without receipts,
commands returning 200 rather than prompt 202, peer-message endpoints outside the
prompt path, captures beyond the startup budget, and cross-tab IndexedDB races
are not inferred safe. No automatic replay, cross-chat scan or stronger delivery
guarantee is claimed. Unknown outcomes retain the existing warning and text.

## Verification

- `make test-web-send-receipts`: race detection, repeated three times, with store/web checks for local, directed,
  explicit-target and steered results; exact response equality; source isolation;
  invalid methods/queries/tokens; failed receipt writes after accepted work;
  upgrade/reopen; duplicate and two-connection concurrent receipts; deleted
  targets; `no-store`; and indexed query-plan use.
- `make test-web-basic-send`: 22 real HTTP/localhost Chromium/WebKit cases, including
  routed lost replies and routed close/reopen, one upstream POST, native target
  turn and unchanged source draft/session. Existing attachment, unknown-read,
  newer-draft and single-session close tests use the new endpoint.
- `make test-web-basic-controls`: 24 cases (four journeys × two engines × three
  sizes), including explicit Ctrl+Enter steering, lost acknowledgement and exact
  source receipt without new turns or restored follow-up text.
- `make test vet bun-checks test-web-http-helpers`: pass; 16 helpers/65 assertions.
- Full isolated functional suite: 129 passed, 11 existing skips.

Focused independent review found no blocking ownership, false-success, race or
security issue within this scope. Initial tests used the wrong server helper and
log import; routed source counts were changed to baseline comparisons because
sessions contain prior history. Destroying a live HTTP socket caused Chromium
to retry at transport level; that run remains evidence of why receipts are not
idempotency. The live-page lost-reply test uses the established failed-response
interceptor, while page-close tests use a real single-forward proxy. Steering
uses the actual Ctrl+Enter shortcut, not the initial incorrect Alt+Enter probe.
One WebKit startup-focus assertion failed then passed unchanged; no root-cause
fix is claimed. No timeout/test criteria weakened or supplied components edited.

This slice awaits whole product CI and exact-source deployment. Basic web
functionality remains the priority; terminal/autosave work stays paused.
