# Installed Piclaw Output/status slice

Gi now uses the installed Piclaw 3.2.4 status renderer through a build adapter instead of the older supplied renderer and separate Gi tool footer. Protected component/UI/pane files and historical contract snapshots are unchanged. See [ADR 0059](../../adr/0059-pinned-piclaw-status-adapter.md).

## Observed contract

The independent probe runs `/opt/piclaw/current`'s real `createStreamingEventHandler` and shipped Classic assets, not Gi's vendor copy. The bounded Output journey matches in Chromium and WebKit at 390×844, 820×1180 and 1440×900:

- Output is separate from conversation posts and command preview.
- Collapsed Output selects the newest six lines and clips long lines at 132 characters. Expansion restores the retained preview in its scrollable body.
- Markdown line breaks, disclosure DOM/text and selected computed styles match. Embedded script text remains sanitised.
- Tool completion removes Output and shows `Waiting for model...`; it does not imply turn completion.
- Authoritative idle removes transient status and retains the unsent composer draft.

Piclaw's Thoughts/Draft panels show nine-line tail windows with generic `more…`/`less` controls. Expanded bodies remain scrollable; they are not the older unbounded bodies with omitted-line counts. Existing tests now assert these observed semantics, retained earlier text, keyboard scope and independent expansion. No test was removed. Gi's existing resize/fallback measurement guard remains a narrow adapter because installed Piclaw does not remeasure container-only resizing.

## Native data path

`internal/turn/tool_output.go` records bounded `tool.output` snapshots before broadcasting activity invalidation. The store matches turn/call/occurrence identity and restores the current preview after reopen. Late output for an older occurrence cannot replace a newer one.

`ToolRuntime.OnOutput` is optional. Shell commands report cumulative stdout/stderr; other tools report their returned output. Intermediate persistence is limited to four snapshots per second, with a final flush. This is not an output log or a replacement for raw model history.

A review found that returning a persistence error directly from the pipe writer could stop draining and block a noisy child. The writer now kills the process group, continues draining and returns the original persistence error. The regression uses an endless writer under a deadline. Parent cancellation still terminates the process group. These process changes apply only to the streaming path.

## Evidence

- `make test-piclaw-output-oracle`: six installed-Piclaw/Gi browser comparisons passed. Artifact directory: `test-results/ux-oracle/output-contract/`.
- `make test-ux-tool-terminal`: 12 native lifecycle cases passed across six projects, including Output reload, timing, cancellation, stale reads, session isolation and draft/media retention.
- `make test-ux-thoughts`: 48 cases passed across six projects, including wrapped-content resize and the no-ResizeObserver fallback.
- `make test-piclaw-tool-output`: race×3 store, tool and reporter checks plus adapter/projection tests.
- Final `make check`: Go/vet/build/hooks and 142 functional cases passed; 11 functional cases skipped.
- Final `make ux-parity-inventory`: 204 support tests / 7,747 assertions passed. Classic remains 97 mapped / 144 unmapped; Shared remains 30 mapped / 12 unmapped.
- Focused follow-up review confirmed the noisy-writer blocker resolved and found no new blocker in that scope. An initial broader runtime investigation timed out.

The full `test-piclaw-chat-lifecycle` probe now passes **6/6** Chromium/WebKit phone/tablet/desktop cases. It checks synthetic idle, Thoughts/Draft previews, Output, Waiting after tool completion, terminal provider-error cleanup, idle after reload and a fixture assistant Markdown post after another reload. WebKit still reports presence/SSE cancellations and a presence access-control page error at unload. A minimal page with no Piclaw or Gi assets reproduced the same errors and connected a replacement stream. The fixture now admits only the old scoped stream and one presence unload cancellation/access-control pair inside each explicit WebKit reload window; it requires the old stream to close and a replacement to connect, records all allowances, and rejects unrelated errors. Four new support tests cover that fence, including repeated failures. `ORACLE_TRANSIENT_ONLY=1` also passed **6/6** separately with no unload allowances. This is synthetic UI and fixture-history evidence, not a live provider, persisted backend reply or general reload-parity claim. Native Gi reload is covered separately.

The asset fixture still recognises WebKit's cancellation wording for the single unscoped Gi bootstrap stream, as it already did for Chromium. Outside the explicit installed-Classic reload window, scoped, repeated, topic and Piclaw cancellations remain failures. A native generated manifest route is explicitly supplied by the comparison harness.

No production process, live chat, configuration, database or running TUI was changed. No new lifecycle mapping or blanket parity acceptance is granted. Settings, IDs/references, composer/model, queue/steer and malformed tool-input work remain open.
