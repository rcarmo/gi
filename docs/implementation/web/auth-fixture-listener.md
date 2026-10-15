# Disposable auth fixture listener

The auth/passkey browser fixture now allocates its port inside the native test server and keeps that listener through serving. The previous helper reserved a port in JavaScript, released it, then asked Go to rebind it.

CI36186190456 failed one passkey case before test assertions: `/api/auth/status` never became ready within 10 seconds. The remaining 188 cases passed. That job did not retain its server log, so the underlying bind failure was not proven. The reserve/release race was confirmed by source inspection and removed without raising the readiness limit. Required CI now uploads the passkey fixture logs, traces and result report even on failure.

`tests/ux/server/main.go` binds `GI_UX_LISTEN` before loading fixture configuration, then publishes the origin through `GI_UX_READY_FILE`. `auth-environment.mjs` requests `127.0.0.1:0` for first startup. Passkey fixtures use the test-only `GI_UX_PASSKEY_ORIGIN=localhost` flag to derive a localhost RP/origin from that owned listener. This changes no production RP configuration.

Restart cases reuse the same origin and explicit port so existing browser cookies and ceremony configuration remain valid. If another process takes that port during a deliberate restart gap, readiness fails with native diagnostics; it never silently changes origins. Spawn errors, early exits and changed restart origins reject the fixture. Readiness-file discovery and HTTP polling share the existing 10-second budget. Cleanup flushes the server log before removing the disposable state directory.

Verification: `make test vet`, 189 passkey browser cases, 120 auth cases and 42 startup journeys pass. Native restart, current-RP inventory, insecure-host, duplicate, removal, logout and setup cases are included in those suites. Local validation used the current session-panel working tree; this commit changes only fixture startup and CI diagnostics. The build error from a duplicate `err` declaration during the edit was corrected before these runs. No live auth mutations, policy changes, feature mappings or device/Visual acceptance credit were added.
