# Copilot response-header timeout recovery

Screenshot 3143 shows `Post "https://api.enterprise.githubcopilot.com/responses": http2: timeout awaiting response headers`, followed by a duplicate assistant-labelled inference error. The system-role duplication was fixed in `d0470da`; the previous runtime still ended a turn on its first timeout.

## Oracle and implemented scope

`make test-provider-retry-oracle` invokes installed Pi/Piclaw functions. Pi's `isRetryableAssistantError` returns true for the exact error; `retryDelayMs` yields 2000, 4000 and 8000 ms. Piclaw's `classifyOpaqueAgentFailure` returns `timeout`, and its default recovery decision permits a transient retry when no output or tool execution has occurred.

The implementation retries only the current provider request, within the same turn and iteration. It never replays earlier tools or resubmits the user message. An attempt that streams text, thinking or tool-call events is excluded from this retry path. Authentication, request/schema, cancellation, hook and unclassified failures are also excluded. This is a pre-response retry fix, not full Pi/Piclaw automatic-recovery parity.

Workspace `.pi/settings.json` accepts Pi-compatible `retry.enabled`, `maxRetries`, `baseDelayMs` and `maxAgentDelayMs`. Missing values use enabled/3/2000/60000. Explicit false and zero retry count are respected; nonpositive delays fall back to defaults. Counts are bounded at ten retries and delays at 60 seconds; no API/UI setting editor is added.

The runtime records `inference.retry_scheduled` and `inference.retry_resumed`, and keeps the turn claimed/running during backoff. A cancellable timer releases promptly on Stop/Escape. The activity endpoint restores retry metadata while the authoritative phase is `retry_wait`; terminal or newer activity cannot revive it. TUI topic events show a retry status rather than an Error block. The browser adapter projects the same notice and prevents the last tool footer from hiding it. New composer text is untouched.

Transient error events are not broadcast as terminal failures. On exhaustion, the existing turn owner records one durable system error and finalises failure. Raw diagnostics remain in the terminal failure record; no credential-bearing request body is added to retry events.

## Verification

- Installed classifier/backoff/decision probe for the exact screenshot error passed.
- Focused race×3 tests cover same-turn recovery, exhaustion, cancellation, disabled retry, partial-progress suppression, nonretryable failures, configuration, reload snapshot and TUI presentation.
- Final `make check` passed: Go tests, vet, web build, hook checks and 140 functional tests, with 11 skips. The initial full gate failed because the new browser spec imported a nonexistent helper; that fixture was corrected without relaxing the assertions.
- A final cancellation-race guard adjustment was followed by another successful focused race×3 run.
- The browser regression is an isolated activity-route fixture: notice, reload, newer draft retention and suppression of stale completed-tool UI. It does not establish a live provider's browser recovery journey.

Independent review prompted additional guards: attempt-scoped deadlines retry only while the owning context is live; only selected transient 5xx codes retry; nonpositive delays cannot create immediate retry bursts; scheduling failures preserve the original provider error. Claim or store failures still stop recovery rather than proceeding without authoritative ownership.

No live provider request, production database change, executable replacement or restart was performed. Partial-stream recovery, provider-specific retry hints, broader compaction/recovery policy and the full Pi/Piclaw error UX remain open.
