# go-ai 1.1.0 adoption

Gi upgrades `github.com/rcarmo/go-ai` from v1.0.4 to v1.1.0, the stable tag
at `f44bf375763046ebef6f4eb49fea70ea046fc3ed`. Direct upstream lookup confirms
this release; the Go proxy's cached latest response initially still named
v1.0.4. No other dependency versions or frontend pin changed.

The release ports pi-ai/durable 1.1.0 provider retries, recorded assistant
stream duration, snapshot events, classifier images and Decisions, catalogues,
OAuth cancellation and provider-specific corrections. Gi compiles against it
without API adaptations. No live-provider acceptance was run.

## Native compatibility

225 focused inference/codemode/turn/TUI checks pass three race repetitions,
covering auth/model/classifier, codemode, cancellation/abort/retry, tool execution
and MCP paths. Fresh full regression passes 2,201 tests across 35 packages with
`RACE=`. Explicit vet and CGO-disabled build pass on Go 1.27.1.

Repeated race checks found a test-store leak: `openTestStore` reused a shared
in-memory database named after the test without closing it. The helper now
uses a unique invocation ID and registered cleanup. Engines close before their
stores through LIFO test cleanup. This prevents repeated abort checks from
colliding on prior session IDs; runtime persistence is unchanged.

Gi's retry classifier now recognises Pi 1.1.0 `server_busy`, `servers are
currently busy` and `selected model is at capacity`. Permanent HTTP responses,
unsupported requests, context cancellation and hook aborts take precedence.
Existing retry bounds and the no-retry-after-streamed-progress guard remain.
These changes do not replay executed tools or earlier inference iterations.

## Concurrent issue lanes

Seven issues were open during review: #11, #16, #25, #30, #45, #46 and #47.
Recorded native tool timing is the closest follow-up to this dependency release
and the Pi TUI 1.1.0 alignment. The dependency's assistant duration alone does
not make Gi's resumed tool results duration-correct; that needs a separate
execution-boundary persistence/rendering slice.

#30's TUI codemode rendering already exists; its web portion needs the shared
frontend owner. #25's OAuth/provider/model subissues are closed, but #30 and
remaining acceptance keep the umbrella open. #16's missing `/changelog`,
`/trust` and `/bug` require explicit Gi contracts rather than copied labels.
#11 restructuring is independent and should stay separate from this upgrade.

#46 awaits lazy-loader revision forwarding upstream, with failing consumer
regressions retained. #45 retains the WebKit popout-restoration failure.
#47 has passing live policy/input checks but its historical preview reference
still lacks WebKit coverage because of the shared upload limitation. No issue
was closed, no skip/capability changed, and no agent contact or deployment ran.

## Profiling

Pre-release CPU, alloc_space and alloc_objects were analysed. Focused workloads
concentrate in inference setup, QuickJS bridge and transcript/model-menu test
fixtures. The full-suite allocation hotspots include existing codemode VM
calls (388 MiB cumulative in its package view) and TUI transcript layout
(267 MiB cumulative in its package view). No equivalent-workload improvement
was measured; those aggregate values cannot establish a speedup from this bump.
Used profiles, matching test binaries and failed/probe logs were disposed after
diagnosis. Source, installed toolchains, shared active caches and unrelated
Joker edits were preserved.
