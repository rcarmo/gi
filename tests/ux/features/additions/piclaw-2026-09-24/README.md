# Single-user passkey Settings — additional contract

Copied verbatim from Piclaw commit `b531ea3a8cb38ba8f1e0729f8b6d0439f63cb5f1` (2026-09-24), `tests/e2e/features/shared/single-user-passkey-settings.feature`.

SHA-256: `bd48cab9126778763cab3ddfd8ee04a89bd8c29da62e3bc4188cf24dc3a55b82`.

The 26 scenarios/outlines describe **upstream** implementation and browser evidence, not Gi support. Gi has no WebAuthn credential store, challenges, registration/assertion verification or credential-management API. All 26 remain unimplemented here. The upstream `@implemented` and `@browser-verified` tags are retained as source text, not Gi credit.

This additive contract does not replace or expand the frozen 236 Classic / 42 shared inventory. Track it separately until a runner and per-case Gi evidence exist. Upstream design and evidence are at the same commit:

- `docs/design/single-user-passkey-settings.md`
- `docs/reviews/single-user-passkey-settings.md`

Implementation prerequisites: persistent per-credential IDs/public keys/counters/names; RP/origin-bound, expiring one-use registration/assertion challenges; five-minute recent-auth proof; atomic lockout-safe remove/recheck; owned rename/delete; duplicate/uncertain-registration recovery; policy-aware login; accessible browser-owned prompt cancellation/focus; no secret-bearing browser storage. Removal must block future credential assertions without silently revoking existing sessions. Physical/synced authenticators remain manual-device evidence.

The current TOTP-only browser gate must not present passkey controls before these native capabilities exist. The local TUI does not use HTTP authentication: do not add idle login rows or passkey-management chrome. Any future terminal action should open the authenticated browser Settings, not emulate browser WebAuthn or bypass recent proof.
