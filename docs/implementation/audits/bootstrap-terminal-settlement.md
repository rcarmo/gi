# Bootstrap terminal settlement

Bootstrap shell turns now commit their response and tool receipt together, then
use the common child-before-parent terminal coordinator. The shell process and
coalesced progress writer finish before that coordinator starts. Required child
hosts join before parent terminal publication, claim release or queued work.

`Store.SettleToolResponse` shares the tool-result transaction and stores an
assistant history row instead of an orphan tool result. Empty output adds no
history row. Failed and cancelled output retains bounded observed stdout and
native stop metadata; context projection excludes it. The transaction checks
cancellation independently of the host context, so an already committed cancel
cannot adopt a late response. Caller metadata is not mutated.

Receipt or terminal persistence failures retain the active claim and block the
queue. Recovery remains conservative; this change does not replay shell effects.
#86/#87 still require graceful checkpoint-preserving shutdown. #92 still tracks
restart continuation. #90 is already closed.

## Verification

Regressions cover completed/failed/cancelled bootstrap parents with a
cancellation-ignoring child, output retention and actual context exclusion,
child-before-parent receipts, queue blocking, response/terminal write failures,
transaction rollback/reopen/idempotency and cancellation winning adoption.
`make test-bootstrap-settlement` repeats the focused race checks three times.
The equivalent real-shell workload is `make test-bootstrap-settlement-profile`.

Final verification: full Go suite, focused race ×3, `make vet`, `make build-web`
and `make bun-checks` pass. A fresh isolated `make test-ux` run completes with
147 passed and 13 skipped. No release or deployment was performed.

## Profiling

Go 1.27.3, two CPUs, 200 real-shell turns, with a fresh WAL store and engine per
test. The legacy shell implementation and the common path used identical
visible-response and exactly-once boundary assertions. CPU, `alloc_space` and
`alloc_objects` were inspected. SQLite preparation/execution and fixture store
creation dominate; receipt and staged-terminal transactions add allocation work.

At the default heap sampling rate, legacy/common sampled allocations were
60.61/67.57 MiB and 768,975/983,264 objects. After the cancellation-adoption check,
the common path sampled 69.38 MiB and 1,111,081 objects: about 45 KiB and 1,711
additional sampled objects per turn versus legacy. Common-path CPU samples varied
2.04–2.88 s and test time 5.30–8.71 s; legacy sampled 2.43 s and took 6.66 s.
These short, fixture-heavy runs do not establish a throughput improvement.
The extra durable boundaries and cancellation/child ordering justify the small
absolute allocation cost; no acceptance or isolation checks were removed.

A measured SQL/accounting follow-up combined two snapshot queries into one.
An equivalent 200-turn comparison at a 64 KiB heap sample rate increased sampled
CPU 2.62→3.34 s, allocations 62.09→64.64 MiB and objects 962,095→1,007,014.
That experiment was reverted. Broader query/cache changes are outside this
bounded correctness repair. Raw captures, matching binaries, temporary source
copies and disposable comparison logs were removed after analysis.

The first browser acceptance run was interrupted at case 152/160. Its stopped
Node launcher sampled 639 ms CPU and 60,488 KiB live heap, dominated by module
loading; it did not capture application allocation history. Those probe
captures were analysed and disposed. It is not a completed browser-suite result.
