# Pi environment, startup logo and hook visibility

Gi now reads Pi's transport preference, loads AgentsInTheCloud instructions and skills from ancestor directories, shows a compact G/i block logo, and hides routine hook audit blocks unless started with `-debug`.

## WebSocket error diagnosis

The reported `websocket: message too big: read limited at 32769 bytes` comes from coder/websocket's default 32 KiB read limit. The boundary reader consumes one extra byte to detect overflow; 32769 does not establish the full event size. It is unrelated to hooks, debugging or model context limits.

The managed Pi user settings specify `transport: "sse"`. Gi previously omitted that field when loading settings and constructing stream options, so go-ai's Codex auto transport opened a WebSocket. The preference now reaches normal turns, side prompts, compaction and branch summaries. Project settings override user settings; legacy `websockets` booleans remain readable. No credential or managed settings file was changed.

A deterministic native-provider test loads the SSE preference from a private Pi agent directory, serves a 40 KiB event, and checks the complete result, exactly one POST and zero WebSocket upgrades. The follow-up go-ai fix (`6244052`, adopted through an exact module pseudo-version) sets a 4 MiB limit on every Codex WebSocket dial, including auto and cached connections. It retains a bounded reader and leaves the general-purpose transport default unchanged. Gi tests now load both explicit WebSocket and auto preferences from private Pi settings, preserve a 1 MiB event and verify one upgrade with no SSE fallback. Upstream production-path tests cover 40 KiB/1 MiB delta and final envelopes, two-turn cached reuse, exact 4 MiB acceptance and rejection of one excess byte.

## Host resources and terminal output

Secondary `.agents-in-the-cloud/AGENTS.md` files follow each ancestor's normal instructions. Skill discovery includes `.agents-in-the-cloud/skills` and `.agents/skills`, with closest-directory precedence and canonical-file deduplication of compatibility symlinks. Gi-specific and user-level overrides remain supported. Tests use synthetic private directories and never rewrite the managed files.

Successful hook invocation, modification and response notices require `gi -debug`. Errors and deny/abort decisions remain visible. Runtime execution and durable audit publication are unchanged. Completed debug invocation rows no longer appear as running work.

The logo packs four pixel rows into two terminal rows and uses the existing blue/theme-text colours. Apple Terminal on macOS retains the text fallback. Quiet startup still suppresses it; regular mode prints it once into native scrollback.

## Verification

`make test-pi-environment-settings PROFILING=1 RACE=-race` passes three repetitions across config, skills, inference, turn and TUI. It covers settings precedence, unchanged files, native SSE/WebSocket/auto, main/side/summary transport propagation, ancestor resources, hook visibility and narrow logo rendering. The real-PTY harness passes fullscreen, regular, debug and quiet-startup cases; it checks successful audits are hidden without debugging and denials still appear.

The broader Gi-only `make test-ux` run returned 100 passes, 47 failures and 13 skips. The appearance preset and rapid reverse session-navigation failures also reproduce with the prior `b4572b6b` backend and the current shared UI; only the editor bundle changed in that UI commit. The remaining 45 failures were not individually baseline-tested. This is an unresolved broader regression-suite result, separate from the focused editor acceptance and native checks; no full browser-suite pass is claimed.

Native CPU, allocated bytes and allocated objects were analysed. One-time model-registry initialisation remains the main inference/turn allocation site (about 31–33 MiB in the focused run). Narrow transcript projection contributes about 10 MiB across repeated TUI checks. The follow-up run including 1 MiB WebSocket events sampled about 25–41 MiB in registry initialisation per inference/turn worker; event reads and JSON encoding account for the added large-payload work. Both repositories' full Go suites pass, with Gi vet/build and go-ai vet/staticcheck also passing. These are endpoint workloads with no matched performance comparison; no speedup is claimed. Raw captures and completed scratch are disposed after use.
