# Open-tab local notifications

Gi can opt in to local browser notifications for fresh assistant replies in the selected chat. It does not implement Web Push, service-worker delivery, closed-tab delivery or mobile-background delivery. Native OS permission prompts and presentation are unverified.

## Browser-scoped opt-in

The existing composer bell is enabled only with a secure context, a Notification constructor and Web Locks. Permission is requested synchronously from an explicit user gesture, never at startup. The browser-local `gi_notifications_v1` preference persists across reloads and applies to sibling tabs of the same origin. The UI states that matching chat tabs must remain open. Denied/dismissed permission leaves delivery off; a failed request or storage operation requires explicit retry. The request wait is bounded to15seconds.

A per-document random client ID avoids copied sessionStorage IDs when a tab is duplicated. The shared browser/device ID is reread before publishing presence so simultaneous first-tab initialisation converges. Presence uses the supplied coordinator's120-second TTL and15-second refresh interval. A visible same-chat candidate suppresses delivery; otherwise the lexicographically first hidden client owns it. Disabled clients still publish presence, matching the frozen visibility/ownership rules. No backend presence or push subscription request is made.

## Event and delivery bounds

`gi-notifications-state.ts` accepts only current-chat `new_post` frames with a fresh assistant `agent_response`, nonempty content and a stable message ID. It rejects history older than controller activation or two minutes, future timestamps beyond five seconds, system/user/control/recovery/compaction posts, content-block posts and explicit suppression flags. Ordinary history fetches never notify. Content-block-bearing assistant replies are conservatively excluded in this slice.

A browser lock serialises the final ownership check and message claim across tabs. The verified localStorage ledger retains at most256 recent IDs; saturation suppresses more delivery until entries age out rather than evicting replay protection. Visible suppression also records the ID, preventing later hidden reconnect replay. Replies are claimed before invoking the native constructor; a failed constructor is not retried by a sibling tab. Lock acquisition is bounded to five seconds and cancelled on retirement.

Notification text is generic: title `Gi`, body `An assistant reply is ready.` No assistant text is stored in presence/deduplication state or passed to the OS. Click focuses the window and selects only the captured session; retired callbacks are ignored. At most eight native notification objects are retained; disable/session change/pagehide/unmount closes them and retires pending callbacks.

## Logout and failure handling

Settings retires the local controller before logout, disables the browser preference and waits for a barrier on the same delivery lock. An already-entered sibling callback completes before the browser-authority revocation request starts; queued callbacks recheck the disabled preference. If lock cleanup fails/times out, no logout POST is sent and the existing explicit reconciliation UI reports failure. A failed logout does not automatically restore opt-in.

Revoked storage access cannot throw out of cleanup or unmount. A pagehide/reload without logout retains the saved browser preference; explicit browser opt-in is not a per-tab permission. OS permission remains under browser control. Family-client logout and browser-policy diagnostics in Settings are separate unsupported acceptance areas.

## Verification and limits

- Ten helper tests cover candidate exclusions, visible suppression, hidden leader selection, concurrent duplicate frames, browser-ID changes, visible-to-hidden replay, permission retirement/timeouts, storage denial, ledger saturation, native constructor failure and logout lock ordering.
- `make test-ux-notifications`:24 Chromium/WebKit × viewport cases pass with native reply SSE/auth state and controlled browser Notification/visibility APIs. They prove explicit opt-in, same-browser coordination, generic content, disable/cleanup, no Web Push traffic, and a held lock delaying the real logout POST.
- Auth120, compose30, slash72 and functional107 cases pass; the functional suite retains11fixture-dependent skips. Go/vet/hooks and14pixel helpers pass.
- Passkey regression was interrupted after completing the phone/tablet projects and entering desktop; no whole-run pass is claimed. The resumed desktop project passed63/63, including logout/reconciliation. Required CI still runs the full189-case matrix.
- A focused review prompted the logout barrier and clearer browser-wide persistence wording. The claimed same-tab mid-callback race does not apply to the synchronous delivery/claim body; cross-tab lifecycle changes are serialized at logout and rechecked before construction.

Initial browser test failures used an incorrect cross-tab permission stub and stale sign-in/confirmation button labels. Those fixture mistakes were corrected before the completed24-case run. Frozen feature files and mappings remain unchanged: local ownership implements part of `@ux-extra-011`; single-user cleanup is not family logout parity.

Final pixel run `run-1790383705798-514248`:72images captured with no fixture/page errors; all18cross-host full frames differ and14/36repeat pairs are unstable. Compose differences (light/dark): phone248/764, tablet242/818, desktop308/1048. Exact RGBA, reference assets and renderer flags are unchanged. Native OS notifications, physical devices, Web Push and Visual acceptance remain open. TUI gains no bell, background poller or idle row; notification integration there requires a separate explicit terminal contract.

Voice-input commit `7f5f26f` passed replacement CI36202895941 after CI36201925678 was cancelled. Deployment is held while this notification slice awaits CI; live Gi remains on the verified contrast build during development. Run36206444743 passed notification, passkey and core jobs but the expanded browser-journey job exceeded its15-minute limit (check annotation108303890554). The four compose/panel/voice targets now run in a separate required15-minute job; the original journey, geometry, slash and workspace targets remain in their existing required job. All targets and artifact uploads are retained, with no timeout increase or test skip.
