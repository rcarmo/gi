# Chat UX audit rejected

The 2026-09-26 Gherkin audit did not capture Piclaw's chat lifecycle accurately. The user's 2026-09-27 screenshot shows internal tool results and provider errors attributed to `Validation User`, literal tool-call markers in an assistant response, and a persistent `✓ Completed: skills 0s` footer. The previous fixture-backed pass totals cannot establish chat UX parity.

## Reproduced failures

Read-only Chromium and WebKit probes against the disposable validation instance at `192.168.1.153:19151` both found:

- Four non-user records rendered as `Validation User`: two system errors and two tool results.
- An assistant post containing literal `[tool_call: shell]` and `[tool_call: skills]` markers.
- Both raw tool results rendered as conversation posts.
- The `Completed: skills` footer after the turn had failed.

The probes block every non-GET and off-origin request. They made no inference calls, changed no live chat records and touched no production service.

Commands: `make capture-gi-chat-baseline` and `ORACLE_BROWSER=webkit make capture-gi-chat-baseline`. JSON, DOM and screenshots are under `test-results/ux-oracle/chat-baseline/{chromium,webkit}/`. New runs use timestamped subdirectories to retain failures.

## Piclaw evidence

`make test-piclaw-chat-lifecycle` runs the installed Piclaw 3.2.4 event translator and shipped Classic assets on an isolated fixture host. It records translated frames and rendered DOM at idle, streaming thought/draft, tool output, post-tool waiting, terminal error, idle reload and persisted assistant reply.

Chromium desktop passed. WebKit rendered the preview, output and waiting phases, but its run failed on network cancellation/access-control diagnostics during reload. That failed run is retained; there is no WebKit oracle pass or six-project acceptance.

The bounded oracle shows:

- Idle with no status data produces no activity pane.
- Draft and Thoughts are separate panes above the composer.
- Tool updates feed the Output pane. Its existing trimming/Markdown rules apply; the contract must not invent a plain-text-only renderer.
- The last tool finishing changes status to `Waiting for model...`, with the completed call retained as metadata. It does not produce Gi's persistent Completed footer.
- A terminal provider error produces a transient agent error presentation and clears previews. A later idle snapshot clears the activity pane. Durable error-message persistence was not tested by this probe.
- A persisted assistant response has the assistant identity and renders Markdown emphasis/list items.

This is event-translation and rendering evidence with synthetic events and timeline records. It does not verify production backend persistence, provider recovery, full chat behaviour or physical devices.

Source paths under `/opt/piclaw/current/app/runtime/`:

- `src/channels/web/sse/agent-events.ts`: active tools, output preview, post-tool waiting and terminal provider errors.
- `src/channels/web/agent/agent-status.ts`: authoritative active/idle snapshots.
- Classic source map `web/static/classic/dist/app.bundle.js.map`: `components/chat-surface.ts`, `components/status.ts`, `components/post.ts`, `ui/app-sse-events.ts`, `ui/app-agent-status-orchestration.ts`.

## Causes found in Gi

| Path | Finding |
|---|---|
| `web/src/api.ts`, timeline and search projection | Every non-assistant role becomes `user_message`, including `system` and `tool_result`. |
| `internal/turn/engine.go`, tool-call persistence | Persists `[tool_call: NAME]` in assistant content; timeline renders the storage representation directly. |
| `web/src/app.ts` | Mounts `ToolActivity` instead of the complete Piclaw status/output flow and suppresses `AgentStatus` when tool metadata exists. |
| `web/src/gi-tool-activity.ts` | Adds a Gi-specific terminal footer. This is a defect against the requested clone, not an accepted product deviation. |
| `web/src/components/status.ts` | Supplied component lacks the tool-output pane present in installed Piclaw 3.2.4. Component provenance/version reconciliation is required; no supplied component was edited. |
| `tests/ux/tool-activity.spec.mjs` | Asserted the Gi footer and credited `@ux-original-027`; it never exercised the Piclaw output-pane transition. |

The provider error also contains `tool_use.input: Input should be an object`. The provider rejected the tool-history payload; the cause and fix are not yet verified. Presentation changes alone cannot repair that rejected request.

## Contract changes and next gates

The historical `.gherkin` files and SHA manifests are unchanged. Active `@ux-original-027` now specifies the actual tool/status transition. Its old mapping is removed. The existing Gi footer tests remain as implementation regressions, without Piclaw tags or accepted-deviation status.

Five new active `@ux-chat-lifecycle-*` scenarios cover idle, streaming panes, output routing, assistant identity/Markdown and terminal error handling. They are unmapped. Current inventory is 241 Classic IDs / 262 expanded cases, of which 97 have mappings and 144 do not. These counts are inventory only.

Required before implementation acceptance:

- [ ] Verify all visible chat states against Piclaw, including Settings title, references, composer, model changes and reconnection.
- [ ] Reconcile the supplied component version without locally patching protected component files.
- [ ] Fix native message projection and event routing; keep provider history separate from visible conversation posts.
- [ ] Fix provider tool-input serialization and prove a complete tool-using turn with an isolated provider fixture.
- [ ] Run failing-then-passing native browser journeys for the full lifecycle, including reload, error, cancellation, cross-session routing and output disclosure.
- [ ] Resolve the WebKit oracle reload failure rather than ignoring its diagnostics.
- [ ] Diagnose the integration `make check` exit and review pre-existing generated assets.
- [ ] Independent review before integration promotion. No main push or production deployment is authorised by this checkpoint.

Independent read-only review initially found overbroad clauses and an error-cleanup test that could self-validate. The clauses were narrowed, and a DOM observer now proves zero Draft/Thought panes while the transient error is present, before reload and without synthetic cleanup. Follow-up review found no blocker in this bounded contract checkpoint. `make ux-parity-inventory` passed 195 tests / 7,693 assertions; this validates contract tooling, not Gi chat UX.

Nine staged deletions under `artifacts/tui-audit-20260525/` appeared while this audit was running. Their origin is unverified. They and pre-existing generated web changes are excluded from this checkpoint; no unrelated staged changes are accepted by it.

The LAN proxy's previous 10-second SSE idle timeout was a separate test-infrastructure failure. Removing it and receiving a heartbeat after 16 seconds did not validate Gi's reconnect lifecycle.
