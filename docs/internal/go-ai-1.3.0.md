# go-ai 1.3.0 alignment

Target release: `github.com/rcarmo/go-ai v1.3.0`, tag commit `d9449bf4b102eb4303b8c60e9b0649b4bf99d08f`, upstream pi revision `71023aec1546d6d677a65669c5665c0927ce0dd0`. This branch builds on the verified, unmerged 1.2.0 candidate from Gi PR #55; frontend PR #2 also remains unmerged.

## First contract slice

1.3 adds `ArgumentsError` to tool calls/content blocks and preserves provider parse failures through `GetToolCalls`. Gi now rejects marked calls before tool-call/approval/result hooks and executor effects. Rejection commits an error result and fresh terminal occurrence atomically, without an execution start or invented duration. Assistant call summaries and fallback replay retain the marker.

Development verification: Go1.27.1, project-owned isolated test environment with inherited agent-directory overrides unset. Focused malformed-call/replay tests pass race ×3; inference, turn and store suites pass. No 1.3 pre-release profiling, full suite or browser matrix has run yet.

## Remaining audit

- Ensure batch execution start/end hooks and direct/nested dispatch paths honour the upstream malformed-argument boundary; current test covers call/approval/result hooks only.
- Direct dispatch rejects the marker before lookup; import/export preserves argumentsError. Focused direct/import/export/malformed-call tests pass race ×3. Persistence/reopen and batch-hook coverage remain.
- Compatibility SSE now accumulates argument fragments by index and validates completed object JSON, preserving ArgumentsError into the emitted content block. HTTP fixture tests cover split/interleaved valid objects, array/null/truncated failures and empty arguments; race ×3 plus full inference/turn suites pass. Deliberate current limitation: go-ai's internal completed-call parser can repair malformed escapes/control characters; Gi's compatibility parser rejects them conservatively. Equivalent repair remains an audit item.
- The compatibility request serializer also needs typed call/result replay review; parser completion alone does not prove provider transcript parity.
- Audit remaining v1.2→v1.3 provider/catalogue/durable delta against tagged release docs, rather than assuming the dependency upgrade is sufficient.
- Run full native/race and web acceptance, pre-release CPU/allocation analysis and disposal, then publish a review candidate. No merge or deployment authorised.
