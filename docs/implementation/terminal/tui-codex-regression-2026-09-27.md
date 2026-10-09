# Codex request and TUI system-message fixes

Screenshot 3137 shows `Codex error: Unsupported parameter: max_output_tokens`, an Error panel plus an assistant-labelled copy, and a stray `bytes).` fragment. The footer names the older `wip/maintenance-tui-autosave-20260926` worktree. Repeated user text alone does not establish duplicate admission.

## Request compatibility

The pinned go-ai Codex builder assigns `MaxOutputTokens` via `ClampStreamMaxTokensPtr` whenever stream options exist, including when Gi did not request a limit. The installed Piclaw 3.2.4 Codex request builder omits this field:

`/opt/piclaw/current/app/node_modules/@earendil-works/pi-ai/dist/api/openai-codex-responses.js`, `buildRequestBody`.

`internal/inference/codex_compat.go` composes the existing payload hook and removes only `max_output_tokens` for the Codex Responses API. Other Responses APIs are unchanged. Hook failures propagate; custom payload fields and JSON numeric precision are preserved without mutating hook-owned objects.

The native test uses synthetic local credentials and an HTTP server that rejects the unsupported field. It exercises go-ai's actual compressed SSE request after local WebSocket rejection, checks exactly one POST and receives streamed assistant text. No production credentials, provider calls or chat writes are involved.

## System-message identity

`chatTUI.handleEvent` previously sent all `new_post` content through assistant finalisation. The runtime emits terminal provider failures as `system_message`; treating those as assistant text bypassed the TUI's existing error deduplication.

System posts now use system-role projection, matching reload. Regression tests feed the streamed error followed by the durable system post and require one Error presentation with no assistant-labelled copy. Normal assistant replies retain their role.

## Evidence and limits

- `make test-codex-tui-regression`: both focused suites passed with race checking and three repetitions.
- `make check`: Go tests, vet, web build, hook checks and 139 functional tests passed; 11 functional tests skipped.
- Follow-up native PTYs passed: 24 Markdown, 3 scrollbar, 6 source-copy and the input-history Gherkin scenario.
- The new Gherkin files describe the covered behaviour. They are not Piclaw web parity mappings or claims of complete terminal acceptance.
- A focused independent review attempt timed out; no independent approval is recorded.
- The running TUI was not restarted or upgraded. The stray fragment and broader layout discrepancy have not been reproduced or fixed by this change.

A repeated-test fixture initially reused a registered model ID and retained a closed server URL. Each run now registers a unique fixture model. The first failed run and final gate logs are retained under `/workspace/tmp/gi-worktree-consolidation/`.
