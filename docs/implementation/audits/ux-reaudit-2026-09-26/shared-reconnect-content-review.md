# Shared36–39 and 42: scoped native journeys

Gi has tagged tests for these five shared definitions, with separate fixture
servers for reconnect, copy/delete and message retrieval. All five focused
runs passed six Chromium/WebKit viewport projects each. The earlier deployed
Shared36–38 web mappings remain documented separately; this audit adds no
live-service verification.

| ID | Native evidence and scope limit |
|---|---|
| `36` | `reconnect.spec.mjs` holds a captured turn across a real SSE proxy outage, refreshes authoritative status and cancels only that turn while preserving queue/draft and ignoring stale terminal frames. 6/6. Clean-state version-drift policy is separate. |
| `37` | `shared-copy-delete.spec.mjs` checks accessible message/code copy, clipboard source text, denied-copy feedback, busy-turn deletion rejection and later single-message removal by ID with draft/session isolation. 6/6. Reply-cascade Classic024 is unsupported. |
| `38` | `message-retrieval.spec.mjs` runs a real provider→native `messages` tool loop with durable row IDs, bounded context/windows, exact pagination, foreign/missing ID isolation and quoted data. Race×3 store/tools/turn tests and six browser cases passed. Classic025 all-chat/family scope is absent. |
| `39` | `drafts.spec.mjs` tagged attach path cancels/retries native media delivery once, persists attachment bytes through reload and source removal. 6/6. Drop/paste variants are untagged and were not in this focused run. |
| `42` | `speech-contract.spec.mjs` checks native code text copy, supported speech visibility, owner transfer and stale callback fencing. 6/6 with stubbed synthesis. Audible/physical acceptance is absent. |

These results do not establish current Piclaw 3.2.4 backend/UI runtime,
physical-device acceptance or a whole shared suite pass. No production code
or frozen contract changed.
