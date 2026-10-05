## `/btw` backend

The side-prompt API runs one inference request against a read-only snapshot of the selected session. The answer and thinking are returned to the caller and never written to conversation history. The Classic composer, side-answer panel and injection controls need an upstream frontend handoff before `/btw` is accepted as a user-visible feature; issue #40 and its shared-suite skips stay open.

## HTTP contract

Authenticated same-origin clients on HTTPS or localhost can send:

```http
POST /agent/side-prompt/stream
Content-Type: application/json

{"prompt":"side question","chat_jid":"gi:<session>","system_prompt":"optional instructions"}
```

`chat_jid` must name an existing session. An omitted or blank system prompt uses `SidePromptDefaultSystem`. Unknown JSON fields, trailing JSON and invalid inputs return 400; missing sessions return 404, unauthenticated requests return 401, unsafe transport/origin returns 403 and a full side-request pool returns 429. Methods other than POST return 405.

`POST /agent/side-prompt` accepts the same body and returns JSON:

```json
{"status":"success","result":"answer","thinking":"reasoning text","model":"provider/model"}
```

Its inference failures return 502 with `{status:"error",result:null,thinking:null,model:null,error:"..."}`. Responses use `Cache-Control: private, no-store`.

The streaming route sends SSE events in this order:

| Event | Data |
|---|---|
| `side_prompt_start` | `{chat_jid}` |
| `side_prompt_thinking_delta` | `{delta}` |
| `side_prompt_text_delta` | `{delta}` |
| `side_prompt_done` | The complete success JSON |
| `side_prompt_error` | The error JSON, instead of done |

Thinking/text deltas may interleave. Once SSE starts, inference errors retain HTTP 200 and use `side_prompt_error`. A disconnected caller cancels the provider request and releases its pool slot; no final event is guaranteed after disconnect. Timeout and engine shutdown errors are sent when the caller remains connected. Retry is a new POST. There is no retained side transcript, automatic retry or side-specific injection endpoint.

## Isolation and limits

The request captures the session's selected model and thinking level, then reads the committed context checkpoint and uncovered user/assistant history through `Store.ContextSnapshot`. Projection preserves compacted summaries and provider-safe session-owned media; another session's media is excluded. Later main-turn appends cannot change this request's copied context. The ordinary provider credential loader supplies authentication.

Side inference declares no tools and does not acquire the main-turn queue, create a turn, publish main-turn events, invoke extension hooks or update context/usage measurements. Main-turn Stop does not cancel it. Injection must submit the user's chosen question/answer through the normal session prompt route, whose asynchronous 202 admission and persistence rules still apply.

| Resource | Limit |
|---|---|
| Concurrent HTTP side requests | 4 per server, independent of main turns |
| Request lifetime | 2 minutes |
| SSE write deadline | 10 seconds per event |
| JSON body | 96 KiB |
| Prompt / system prompt | 64 KiB / 16 KiB |
| Projected input including base64 images | 2 MiB |
| Combined returned text and thinking | 512 KiB |
| Generated output | 1,024 tokens |

Media metadata is checked before projection loads/decompresses blobs. Context fitting uses the local text estimate, a conservative base64-byte reservation for images and 1,024 output tokens. Oversized context is rejected without modifying or compacting the session. Provider tokenisation may still reject a request that fits this estimate.

Cache retention is disabled for the one-off request. Engine shutdown fences new side admissions, cancels existing requests after main-turn finalisation, and waits up to five seconds for side inference before closing its services. A provider that ignores cancellation can exceed that grace; shutdown logs the timeout. A stream without a final assistant response returns an error, even if partial text arrived.

## Verification

`make test-side-prompt` runs the focused engine/HTTP checks uncached three times with CPU and heap profiles. They cover snapshot/session isolation, model/thinking selection, compacted history, input/output bounds, incomplete responses, main Stop isolation, joined engine shutdown, validation/authentication, pool exhaustion, HTTP disconnect, retry and ordinary injection.

`make test-ux-side-prompt-api` runs a disposable local provider/server and one Chromium desktop API journey. Browser fetch uses the real SSE route; SQLite history remains unchanged across side requests and abort/retry, then changes after normal injection. Server CPU/heap profiles cover initialisation through teardown; browser CPU/allocation sampling is retained beside them in `test-results/ux-parity/artifacts/`. This check does not exercise the Classic `/btw` panel.

The final focused run passed ten tests repeated three times, including completion alongside a running main turn. The isolated browser journey passed. The final full Go regression passed 2,151 tests across 37 packages. Vet and whitespace checks passed. No full fixtures-vibes run or frontend/capability change was made.

## Profile findings

The ten-test repeated workload took about 0.12 seconds in the turn test process and 0.41 seconds in the web process; its complete lifecycle took 5.6 seconds. An earlier media-bound rebuild raised lifecycle time from 5.1 to 8.3 seconds while test-process durations stayed similar, so that increase did not establish a request-performance regression.

Cumulative `alloc_space` and `alloc_objects` show model-catalogue initialisation as the largest allocation source: about 41.6 MiB and 97% of objects in the turn process, and 38.5 MiB and 72% of objects in the web process. This is one-time catalogue loading, charged to the first side request in the engine fixture. Snapshot projection and streaming do not appear among the leading cumulative allocation nodes. Fixtures retain independent stores/providers rather than sharing mutable state to suppress these costs.

The final browser server sampled 110 ms of CPU over 5.9 seconds, with SQLite/HTTP work and catalogue initialisation leading. Catalogue initialisation accounts for about 72% of its sampled allocation objects. Build invalidation raised outer lifecycle CPU/RSS; these are separate from the server profiles. Browser CPU sampling recorded about 5.1 seconds idle in a 5.8-second capture; its 5.0 MiB sampled allocations were concentrated in app/CodeMirror startup. The final full Go run was 5% faster in wall time and 16% lower in CPU than its preceding full baseline. Transcript rendering and model loading remain its largest allocation hotspots.

Earlier failed runs retained profiles and were reviewed. Two turn-test build failures produced no turn CPU/heap files and did not pass the profiling gate. Their web profiles were present. Delegated review attempts timed out; verification above comes from the native/browser checks and direct source review.
