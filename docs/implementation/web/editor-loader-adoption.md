# Lazy editor revision adoption

Gi adopts published fixtures-vibes `a67d8d0` (F1 implementation `005915c`).
The shared build adapter forwards the loaded snapshot revision through the
actual lazy proxy; vendored Piclaw source is unchanged.

Candidate-pin consumer checks pass 4/4, zero retries: Chromium/WebKit desktop
reviewed conflict/Save Copy and clean workspace-update SSE refresh with the
correct next-save expected_revision. A profiled repeat also passes4/4.
Upstream actual-proxy/revision checks pass16 tests/72assertions and mounted
adapter browser checks4/4, including real lazy loader and complete refresh.
Build and TDZ pass. Final upstream pin adds documentation only after the tested
implementation. Profiled final-pin consumer repeat passes4/4 with zero retries.

Unchanged shared workspace018 still fails after unapproved Review overwrite:
Unsaved changes persists, then cleanup exceeds120s. It is not aggregate shared
acceptance. Upstream unit suite has158pass/2fail/1error (meter overlay import and
upload cancellation full-suite interference); isolated meter baseline also
fails. #46 stays open. No skips/capabilities, reviewed-write policy, frontend
oracle or frozen results changed.

Pre-release profiles analysed: Bun focused CPU14ms, live heap1369KiB/10141objects;
mounted Chromium CPU168ms, sampled allocations7321KiB, runnerCPU3729ms/live
heap31873KiB. Bun live heap is not allocation history; WebKitfunctionalonly.
Consumer runtime/Node/Chromium capture conclusions are wrapper summaries;
no matched-workload speedup or allocation reduction measured. Dispose used
profiles/binaries/logs/fixtures after analysis; keep source and installed assets.

Rui also approved removing generic preview008 from VNC closure dependency.
Issue#47 now retains its historical request and adds explicit VNC-specific
closure scope. General-preview Chromium pass/WebKit upload limitation stays;
no VNC issue closure, capability/skip diff or frozen gate change.
