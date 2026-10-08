# Recorded tool execution duration

N1 records `duration_ms` from Go's monotonic execution clock. Normal tool
success/failure payloads persist the measurement and occurrence ID in runtime
events, terminal audit events and saved tool results. Cancellation and result
abort persist their measured boundary even when the caller context is cancelled.
Output flushing and result-hook delivery time are outside the measurement.
Nested codemode runtime events carry their own executor measurement; existing
settled call details retain their independent call-log durations.

The TUI uses the optional non-negative integral millisecond field for final
`Took`. Zero is valid. Missing, fractional, negative, non-finite and overflowing
values have unknown final timing. Running shell `Elapsed` uses its live clock;
terminal results never fall back to receipt timestamps. Local `!!` blocks retain
the monotonic start/end measurement within their existing execution boundary.

Stored result rendering consumes the same field. Recorded terminal occurrences
without a saved result (including bootstrap shell, cancellation and abort) are
projected for TUI display by `ListTerminalToolResults`. This projection does not
insert messages or enter provider history. The first matching terminal boundary
wins; matching requires occurrence and call identity. Existing result messages
suppress duplicate event projections. Legacy events without measured duration
are not synthesised. A stopped projection currently carries a status message,
not a reconstruction of full tool output or arguments.

## Verification checkpoint

On Go 1.27.1, focused timing/lifecycle/store/TUI checks pass three race
repetitions. A red regression first rendered `Took 9.0s` despite a recorded
1.5-second execution; live/restored tests now agree, including recorded zero,
invalid data, reused same-turn IDs, cancellation and late terminal events.
A persisted-result assertion compares occurrence and measured duration to the
terminal audit event. Nested rendering tests use delayed delivery to verify
that the event measurement wins.

Final pre-release Make verification passes 2,233 tests across 35 packages,
explicit vet and CGO-disabled build. Focused races pass three repetitions.
`make test-tool-duration-pty` runs the production fullscreen loop in tmux,
asserts recorded output and `Took 1.5s` from a restored result, and checks clean
Ctrl+D shutdown. Its fixture skips during ordinary native tests. The earlier
generic tmux smoke supplies additional navigation/lifecycle coverage.
No live-provider, physical-terminal or frontend acceptance ran.

The first full verification failed 12 inference cases because the harness's
`PI_CODING_AGENT_DIR` override bypassed tests' isolated HOME fixtures. The passing
run used `env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR make ...`. An earlier
unprofiled TUI run also failed a follow-up-admission TempDir cleanup race;
subsequent full runs passed. These failed runs are not passing evidence. The
unsafe override run touched the service-user auth file; no credentials were
printed or restored from an obsolete backup. No precise pre-run snapshot exists
to establish which credential metadata changes belonged to that run.

CPU, alloc_space and alloc_objects were inspected for focused and full runs.
Full allocation hotspots remain codemode VM calls (about 416 MiB cumulative)
and transcript layout (about 278 MiB cumulative in this run). There is no
matched-workload speedup claim; wrapper comparisons against older suite sizes
are not equivalent workloads. Raw captures and matching test binaries were
automatically disposed after analysis; owned diagnostic logs/run scratch are
removed at a safe boundary. Active/shared caches, toolchains and unrelated
Joker edits are preserved.

## Projection limits

Bootstrap projection tests cover measured zero, mismatched call IDs, first
terminal selection and suppression by a saved result. Stored messages retain
their stable order; projected events merge by persisted timestamp. Equal
millisecond timestamps do not establish exact cross-table ordering. Projection
uses session/turn event indexes and occurrence checks but scans session history;
large-history cost has not been benchmarked independently. It runs on transcript
reload, not each rendered frame. Interrupted executions without a recorded end
remain absent from this projection. Nested zero durations retain the existing
compact row convention (no timing label). Frontend acceptance and other Pi
1.1.0 lanes are separate.

Code: `internal/turn/engine.go` records top-level/bootstrap execution;
`tool_terminal.go` persists cancellation/abort; `codemode_tool.go` measures
nested executor events. `internal/store/tool_activity.go` projects status;
`tool_transcript.go` supplies display-only reload rows. TUI validation and
rendering are in `tool_duration.go`, `tool_render.go`, `chat.go` and
`transcript_window.go`. `test-tool-duration-pty.sh` owns the PTY fixture lifecycle.
