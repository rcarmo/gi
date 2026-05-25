# Deferred TUI parity closure

Date: 2026-05-25

Status: closure summary for the deferred Pi-like TUI parity plan after opt-in clipboard support.

## Summary

Gi completed the remaining high-value deferred TUI parity slices without changing its runtime architecture:

- SQLite/WAL remains the durable source of truth for sessions, turns, messages, steering, media, extension command state, and audit/event projections.
- Existing API/SSE/topic/tool contracts remain compatible; media support is additive.
- The current Go TUI stack remains in place.
- New behavior is covered by targeted tests and documented in `docs/internal/`.

## Implemented

### Extension command dispatch

Implemented in commit `b48b5aa`.

- Added an extension command registry with builtin-command precedence.
- Added JS and Joker command registration bridge APIs.
- Routed unknown TUI slash commands through registered extension commands.
- Surfaced extension commands and conflicts in `/plugins` and `/commands`.
- Published extension command registration/invocation/failure notices.
- Added tests and updated scripting/extension docs.

### Shared media ingestion contract

Implemented across commits `26abdc2`, `55f146a`, and `c3113b6`.

- Added `docs/internal/media-ingestion-contract.md`.
- Added normalized `media:<id>` references and helper normalization tests.
- Added SHA-256 and detected content-type metadata at media creation.
- Added session-scoped media endpoints:
  - `GET /api/sessions/{session_id}/media`
  - `POST /api/sessions/{session_id}/media`
  - `GET /api/sessions/{session_id}/media/{media_id}`
- Added optional media refs to web prompt submissions.
- Persisted normalized refs into turn metadata and user-message payloads.
- Added provider-safe projection for supported image types through `go-ai` image blocks.
- Preserved transcript-safe placeholder behavior for unsupported media/providers.

### TUI media attachment fallback

Implemented in commit `4c4d505`.

- Reassessed `go-tui` direct image paste support and kept direct image paste deferred because the current key/parser surface does not expose clipboard image payloads safely.
- Added `/attach <path> [prompt]` as the terminal-safe fallback.
- `/attach` stores files through the shared session media store with `source=tui`.
- `/attach` with a prompt submits normalized media refs through the same turn metadata path as web/API media.
- Ordinary text paste remains unchanged.
- Updated `docs/internal/tui-clipboard-media.md` and tests.

### Picker/collapse widget audit

Recorded in commit `e209d67`.

- Added `docs/internal/tui-picker-collapse-widgets.md`.
- Kept numbered textual `/model` and `/commands [query]` pickers as the deterministic, keyboard-friendly prototypes.
- Deferred modal overlays and collapse toggles until key conflicts and transcript grouping are solved.
- Preserved textual fallbacks for every interactive surface.

### Screenshot/layout report

Recorded in commit `68a1c5f`.

- Captured tmux pane artifacts under `artifacts/tui-audit-20260525/`:
  - `gi-narrow.*` at 60x18;
  - `gi-wide.*` at 100x22;
  - `gi-desktop.*` at 140x36;
  - `gi-markdown.*` with seeded Markdown/table content;
  - best-effort `pi-wide.ansi.txt` reference.
- Updated `docs/internal/tui-ux-report.md` with layout notes.
- Updated capture/render scripts to include the desktop size and tolerate missing Pi reference capture.
- PDF rendering remains optional; the current environment lacked a Playwright Chromium binary, so pane captures are the retained review artifact.

## Adapted rather than cloned

- Direct Pi/Claude-style modal picker parity was not cloned. Gi's current text-first picker behavior is more reliable under tmux/script and has stable command fallbacks.
- Direct image paste was not implemented as a parser hack. Gi now has a shared media ingestion layer plus `/attach`, so future direct paste can plug into the same contract.
- Tool/thinking collapse toggles were not added because the current transcript model needs stable grouping helpers before UI-local collapse state can be conflict-free and non-destructive.

## Validation

Targeted validation was run after each slice. Final closure validation:

```text
go test ./...
go vet ./...
```

Both passed on 2026-05-25.

## Remaining future work

These are no longer blockers for the deferred parity plan:

- direct terminal image paste if `go-tui` or a replacement parser exposes safe paste/media events;
- richer modal/list widgets if they wrap existing textual state helpers without new key conflicts;
- UI-local collapse toggles after transcript grouping metadata exists;
- optional PDF rendering of the layout report when Playwright browsers are installed in the environment.
