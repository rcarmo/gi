# projects/gi Pi/PiClaw UX Convergence Plan

## Goals

- [ ] Make Gi's TUI steady-state layout match Pi's terminal layout exactly: no top chrome, transcript first, Pi-like bottom editor/status band
- [ ] Keep a single physical bottom notification/status line for transient state
- [ ] Preserve Gi's SQLite/WAL runtime truth and current API/SSE/topic contracts
- [ ] Add screenshot-backed regression coverage for layout-sensitive TUI states
- [ ] Keep changes small, tested, documented, and pushed

## Phase 1: lock Pi-identical TUI layout contract

- [ ] Capture current Pi reference panes at 60x18, 100x22, 140x36 for empty/startup, command, prompt, queued/error where practical
- [ ] Capture current Gi panes for the same states
- [x] Write an explicit TUI layout contract doc defining row order, bottom band rows, truncation, and notification behavior
- [ ] Add deterministic capture fixtures/scripts for Gi layout regression
- [x] Remove any remaining non-Pi top/mid chrome in steady-state render paths
- [x] Ensure empty input row matches Pi visually
- [x] Ensure the bottom band uses exactly: separator, editor row, separator, path/branch row, single status row
- [x] Add/update tests for render height budgeting and bottom-band content

## Phase 2: single-line status semantics

- [x] Inventory all current status/notification writers (`status`, transcript sys lines, topic renderers)
- [x] Define severity/priority rules for the single status line
- [x] Route transient notifications into the bottom status line when they should not become transcript history
- [x] Keep durable/history-worthy events in transcript only when useful
- [x] Add m/t/q/s metrics and model/thinking right-alignment without wrapping
- [ ] Add token/context/cost placeholders only if data exists; never add a second line
- [x] Add tests for truncation, priority, queued/running/error/tool states

## Phase 3: prompt/model interaction screenshots

- [ ] Capture real prompt-response flow with a working deterministic provider path
- [ ] Capture `/model`, model selection, Ctrl-L cycling, `/thinking`, queued prompt, and error states
- [ ] Render updated Gi-vs-Pi HTML/SVG comparison artifacts
- [ ] Attach artifacts for review

## Phase 4: PiClaw operator fit follow-ups

- [ ] Add web media/attachment cards and download affordances
- [ ] Design durable plan/sidebar equivalent for Gi
- [ ] Add release checksums and binary smoke tests to CI

## Validation

- [x] Run `go test ./internal/tui` after each TUI slice
- [ ] Run targeted web/turn tests for status/topic changes if touched
- [ ] Run `go test ./...` and `go vet ./...` before closing the convergence track
- [ ] Push committed slices to `origin/main`
