# Web terminal backend

Gi implements Piclaw 3.2.5's terminal session, WebSocket and handoff protocol with a native PTY on Linux and macOS. Windows and other platforms return 503; their existing binary builds are retained. The shared Classic app at fixtures4259e82 still has no terminal host. Issue #45 stays open and `@cap-terminal` is unclaimed until the frontend owner's dock/tab/popout/zen integration passes independent acceptance.

## HTTP and WebSocket protocol

* `GET /terminal/session` returns `{enabled,transport:"websocket",ws_path:"/terminal/ws",cwd,shell,font_family,active,connected_clients}`. Metadata does not spawn a shell.
* `GET /terminal/ws` upgrades to a WebSocket. The first server message is `{type:"session",session_id,created_at,process_pid,cwd,cols,rows,font_family}`. Retained output follows as `{type:"output",data}`; process completion sends `{type:"exit",code,signal:null}`.
* Client frames are `{type:"input",data}` or `{type:"resize",cols,rows}`. Raw non-JSON frames retain Piclaw's input fallback. Frames are limited to 64 KiB. Dimensions default to 120x30 and clamp to 20..400 columns and 5..200 rows; omitted dimensions retain their current values.
* `POST /terminal/handoff` returns `{handoff:{token,expires_at}}` for an attached shell, or 409 when no attached terminal exists. The next socket supplies `?handoff=token`; the token is single-use, expires after five minutes and binds to the original owner and session. Successful transfer closes the previous client without replacing the PTY. Invalid, expired, replayed or cross-owner tokens are rejected.

The shell resolves through Gi's configured shell path/detection and starts interactively in the workspace root. Persisted non-secret environment overrides apply; keychain entries are not implicitly injected. A terminal executes with the Gi process's OS permissions and is not a filesystem or command sandbox.

## Access and ownership

Every terminal route requires same-origin HTTPS or a loopback peer with a loopback Host, including unenrolled instances. Enrolled instances require a valid `gi_session` browser cookie. Bearer/query credentials do not grant terminal access. Cookie validity is checked before each input frame and output delivery; revocation fences subsequent I/O. Auth errors fail closed.

Without an enrolled owner, a validated 8..128-character client identifier is required in `x-piclaw-terminal-client` or `?client=`. This distinguishes trusted local browser clients; it is not authentication. Browser-cookie sessions use a hash of the cookie for their internal key, and separate owners never share PTYs. No terminal route grants CORS.

## Resource lifecycle

The manager permits at most 16 current sessions. Each PTY retains up to 2 MiB of output bytes in a ring and each client has a bounded 64-frame queue. Slow clients close instead of extending queues. PTY reads preserve split UTF-8 characters; replay drops a partial leading code point. A five-second PTY write deadline prevents a blocked input handler from hanging forever.

A disconnected shell survives three seconds for reconnect; a pending handoff extends this only until the token expires. Reconnecting within grace retains session identity and output. A fresh attach to an already-attached owner replaces that owner's shell. `CloseTerminals` refuses new sessions, closes sockets/PTYs, terminates the shell and discovered descendants/process groups, and joins workers including replaced and grace-expired sessions. The main web process calls it on shutdown. Grace timer callbacks check timer identity so an expired callback cannot terminate a newer detach generation. Process inspection has a two-second timeout.

PTY startup creates a controlling terminal and process session. The master is registered with Go's poller; resize uses `SyscallConn.Control` so it does not change the descriptor to blocking mode. Process cleanup enumerates children with `ps`; independently daemonised/reparented processes are outside the descendant contract. Terminal commands must not be treated as a safe way to run untrusted code.

## Verification, 2026-10-05

* Seven native tests pass, including input, workspace cwd, resize, owner isolation, replay/reconnect, transfer, handoff rejection, cookie revocation, no-spawn metadata, grace expiry, polling preservation and foreground-job cleanup. The final `make test-web-terminal` repeats them three times; tests took 0.745s, with 971MB peak RSS including compilation. No race-detector result is claimed: this host's probe supplied no race flag.
* Full Go suite: 2,134 tests pass across 37 packages; the last full profile reports 110.9s wall, 62.7s CPU and 743MB peak RSS. Final resize-poller, timer-generation and bounded process-inspection changes are additionally covered by the repeated focused tests.
* Full functional suite: 146 passed, 11 skipped. Browser tests use a real WebSocket/PTY and verify resize, workspace cwd, transfer state, origin/client checks and grace cleanup. A resumed selected run passes 2/2 in 5.3s; its complete lifecycle took 172.3s wall, 193.5s CPU and 954MB peak RSS. Final PTY resize is covered natively.
* Windows amd64 and macOS arm64 binaries build with `CGO_ENABLED=0`. Runtime tests ran on Linux only. Vet passes.

The interrupted functional run has no completed profile/report and is not a passing gate. An ensuing startup failed on a malformed disposable SQLite database. Test lifecycle cleanup now waits for the identified fixture process before deleting its workspace/database and removes leftover WAL/SHM files during startup. A later clean full functional run passed; no user database was modified.

## Performance analysis

The output ring replaces whole-window copying on every frame. `BenchmarkWebTerminalReplayRing` measures an already-filled ring with no attached client: 196ns per 8KiB append, 0 B/op and 0 allocs/op. This is the ring operation only, not WebSocket or end-to-end terminal throughput. Initial ring growth, replay materialisation and JSON delivery still allocate.

The initial oversized-output test attributed about 12MB to the output handler. Avoiding oversized temporary copies and JSON encoding without attached clients reduced it to about 2MiB per test; three-repeat profiles report about 6.1MiB for output and 6.1MiB for replay. The synthetic input itself accounts for another 12MiB across three repeats. CPU in the benchmark is mainly the ring append/copy operation. Cumulative CPU and `alloc_space`/`alloc_objects` were inspected after the final repeated run: about 44% of sampled objects come from server fixture construction and HTTP route registration, while replay buffers dominate terminal allocated bytes. Reusing fixtures might reduce test allocation counts but would weaken owner/shutdown isolation; no fixture reuse was added. No changed-output hash is used for acceptance.

Go profile flags for tools/web/turn and OAuth tests were investigated with focused profiled runs. Earlier 1.5s OAuth and 2.4s composer fault-test spikes did not recur (29ms MCP package, 222ms turn package). Other slow web tests spend their time in index/turn waits. Cold builds and cache trimming inflated lifecycle wall times; the seven native tests themselves stayed below one second over three repeats. The long full functional lifecycle included concurrent cross-build compilation and cache eviction, so its aggregate +115% wall time is not a browser latency measurement. The resumed selected run was CPU +3%, RSS unchanged against its preceding selected run.

The profiler retains per-package `test.log`, prints benchmark results and reports missing CPU/heap artifacts. Script lifecycles collect CPU/wall/RSS, not browser allocation samples; Go allocation profiles cover native tests and benchmarks.

Local logs are `/workspace/tmp/gi-terminal-publication-tests.log`, `gi-terminal-profile-analysis.log`, `gi-terminal-go-final.log`, `gi-terminal-functional-all-final.log`, `gi-terminal-browser-resumed.log`, and `gi-terminal-publication-vet.log`. Reports/history are in `~/.cache/gi-test-profile`.
