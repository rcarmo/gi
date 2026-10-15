# WebKit terminal reattach adoption

Gi adopts published fixtures-vibes `bb7786f` (implementation19f946a,
source-window fence928eeca). Piclaw's Safari blanket terminal reattach/manual
close-recovery guards are adapted for Gi's server-owned PTY. Delayed recovery,
handoff/instance identity and browser authentication stay intact; stale-window
reattach messages are fenced. Vendored source bytes are unchanged.

Unchanged shared terminal006: baselineWebKit3/3fail; corrected candidate desktop
Chromium/WebKit6/6pass, three repetitions each/zero retries. Affected shared
terminal001/004/007/workspace0148/8pass. Independent final-published-pin Gi
consumer6/6pass (profiled): same PID/environment, unsent shell input, composer
draft, one connected client and rejection of forged stale sender. Native7
WebTerminal tests/profile plus racecount3pass. Guard/proxy units5/16assertions,
build/TDZ pass. No full shared-suite claim; known meter/import+upload suite
failures from F1 are unchanged. #45 stays open for broader acceptance.

Harness probes with wrong imports/selectors or dock actions failed; they do
not measure original-tab continuity. One adapter-anchor build probe failed and
was corrected/rebuilt; tests against stale assets excluded. Initial native
filter ran0tests, excluded. Owned failed/probe logs/fixtures disposed after use.

CPU/alloc_space/alloc_objects profiles inspected: replay/output allocations
and process shutdown dominate bounded native workload; build/catalogue/runtime
and browser rendering dominate consumer lifecycle. Node/Chromium profiles were
summarised, Bun live heap is not allocation history, WebKitfunctionalonly.
No matched-workload speedup measured. Raw captures/binaries disposed by wrapper;
owned inactive diagnosis scratch removed. Frozen gate/skips/capabilities and
backend protocol unchanged. Test command: `make test-terminal-reattach-consumer`
with installed Playwright browser path and agent-dir overrides unset.
