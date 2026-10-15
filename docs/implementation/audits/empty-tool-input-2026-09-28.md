# Empty tool-input round trip

A disposable two-request Anthropic exchange reproduced the reported error class: `tool_use.input: Input should be an object`.

The pinned go-ai Anthropic decoder creates a tool-call block without an argument map on `content_block_start`. If a no-argument tool emits no JSON deltas, the map stays nil. Its encoder then marshals that map as `input: null` on the next model request. The local fixture rejects that request with the same HTTP400 message.

This establishes a concrete failure mode, not the provider/binary identity of the earlier screenshot. OpenCode's independent argument-chunk assembly is still a suspected separate defect.

## Fix

`internal/inference/tool_input.go` gives tool-call blocks with absent argument maps an empty object. It copies only changed message/content slices, preserving caller history, nonempty argument maps and non-tool content. The inference boundary applies this to incoming context and final assistant messages. Payload hooks remain in their existing order and their output is not rewritten by this helper.

Installed Pi 0.87.1 initialises Anthropic tool arguments with `input ?? {}`. The new oracle invokes its real stream/encoder using an injected SSE client: both returned arguments and the next request's tool input are `{}`.

The adapter cannot reconstruct argument content discarded upstream. It does not claim repair of malformed JSON, partial argument chunks, unknown tool schemas or every provider path.

## Verification

- Baseline local wire test: first run failed with HTTP400 and `Input should be an object`. The subsequent repeated baseline attempts used a cached model endpoint and failed with connection refused; they are not counted as independent reproductions. The fixture now uses unique model IDs for race repetitions.
- `make test-tool-input-contract`: installed Pi0.87.1 oracle passes; native HTTP round trip and immutable normalisation checks pass with race×3. No external provider or credentials are used.
- `make check`: Go/vet/build/hooks and 144 functional tests pass; 11 skipped.
- `make ux-parity-inventory`: 206 support tests / 7,760 assertions pass. No new mappings.
- Focused copy/normalisation review found no concrete mutation or data-loss blocker. Unchanged nested maps remain shallow-shared, as before; the helper does not mutate them.

Evidence logs: `/workspace/tmp/gi-tool-input-{baseline,oracle,final}.log`. No deployment, restart or live-chat writes.
