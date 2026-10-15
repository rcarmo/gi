# Web terminal acceptance (#45)

Gi passes all 20 shared terminal scenarios applicable across Chromium/WebKit desktop, phone and tablet projects at fixtures-vibes `92425adf2772d06221e85d124dcf3830b649a442`. Independent runs return **110 passes, 10 unchanged source skips, zero retries**. Twenty obsolete #45 entries are removed from Gi's skips file. No shared assertion, selector or frozen oracle record changed.

## Scope and results

The selection covers terminal001–017, workspace014 and shell004/007: standalone output and commands, IME-free input, click/tap close, pop-out/reattach, theme, dock keyboard/button toggles, splitter, concurrent editing, zen controls and reconnect/exit states. The shell cases verify both terminal and VNC menu/tab callbacks; they do not prove a working VNC connection.

| Selection | Passed | Source skips | Gate |
|---|---:|---:|---|
| Chromium/WebKit desktop | 38 | 2 | 19 accepted scenarios, one source-skipped, 290 outside selection |
| Chromium/WebKit phone and tablet | 72 | 8 | 18 accepted scenarios, two source-skipped, 290 outside selection |

Source skips are terminal005 on both desktop projects, and terminal004/006 on all four touch projects. They are the shared click/tap/pop-out applicability rules. The catalogue has retired `@cap-terminal`; these scenarios are core, so no obsolete capability is added to the profile.

Run each batch separately through Make, with host agent-directory overrides unset:

```sh
env -u PI_CODING_AGENT_DIR -u GI_CODING_AGENT_DIR \
  make fixtures-vibes-focused PROFILING=1 RACE= \
  PLAYWRIGHT_BROWSERS_PATH=/tmp/gi/cache/ms-playwright \
  FIXTURES_SPEC_ARGS='terminal.spec.ts shell-menu.spec.ts --project chromium-desktop --project webkit-desktop --grep "@ux-terminal-|@ux-workspace-014 |@ux-shell-00[47] "'
```

For the touch batch, replace the project flags with chromium-phone, chromium-tablet, webkit-phone and webkit-tablet. Focused report gates pass after skip reconciliation. These results are separate from a complete current-pin compliance matrix; no full browser-suite pass is claimed. The earlier broader Gi-only functional run remains unresolved (100 passes, 47 failures, 13 skips; two failures reproduced on the prior backend). No fresh remote Piclaw oracle run was possible.

## Teardown race and fix

The initial desktop matrix returned 37 passes, two source skips and one failure in terminal008 cleanup. The trace showed an uncaught xterm viewport callback reading a missing renderer's `dimensions` after dock teardown. Three isolated repeats passed, establishing that a rerun alone would miss the race.

The shared build adaptation now retains successfully loaded addon disposables and drains them in reverse activation order before disposing the terminal core. This covers ligatures/image and other addons as well as renderer/fit; names remain available for diagnostics. Individual cleanup failures cannot prevent core cleanup. Vendor trees stay byte-identical, source anchors fail closed and global browser errors remain visible.

Classic units pass **169/169**, including unsafe-order reproduction, reverse/once-only cleanup, failed activation/disposal and drift rejection. The final unchanged desktop matrix passes 38/38 applicable executions. A separate Gi consumer passes three repetitions on both desktop browsers (**6/6**, 60 real dock teardown cycles), checking actual command output and no late page errors. The existing pop-out continuity consumer also passes **6/6**: same PTY PID, environment and unsent shell input survive reattachment; stale same-origin senders cannot claim the pane, the composer draft survives and only one client remains attached.

## Native safety and profiling

`make test-web-terminal PROFILING=1 RACE=-race` passes three repetitions of native input/resize, origin/client/owner/cookie fences, bounded replay, reconnect/handoff, wrong-owner/replayed/expired tokens, grace expiry, no-spawn metadata and foreground process-group shutdown. The Make target now forwards the race flag and the reattach build uses the configured Go toolchain.

CPU, allocated bytes and allocated objects were inspected in the browser/runtime/runner and native checks. Registry initialisation samples about 24–33 MiB per browser worker; static delivery and session listing are other visible sites. The native repeated oversized-output workload attributes about 6.1 MiB each to output and replay and 12 MiB to fixture input. Chromium captures cover browser CPU and sampled live allocations; WebKit evidence is functional only. Addon ordering prevents avoidable post-disposal work, but no matched timing or memory improvement is claimed. Raw captures, traces, matching test binaries, completed fixture roots and disposable logs are removed after analysis.
