# Bounded current-session message retrieval

The native `messages` tool reads historical messages from the runtime's current
session. It does not accept a session identifier, all-chat scope, SQL, or write
actions. The combined Shared38 web journey covers the current-session contract.
Classic025 remains unmapped: all-chat and family-owned authorization are separate.

## Identity and migration

Existing message UUIDs, HTTP history responses and UUID cursors are unchanged.
`message_rows` maps each UUID to an explicit SQLite `INTEGER PRIMARY KEY
AUTOINCREMENT` identity. Legacy rows are assigned in `(created_at,id)` order inside
the schema-upgrade transaction. An insert trigger assigns identities to every new
message, including transactional compaction writes. Trigger failure rolls back the
message insert. Foreign-key cascades remove mappings when messages are deleted.

Committed numeric IDs survive reopen and VACUUM and are not reused after deletion.
New IDs reflect allocation order, not necessarily chronological order. They are
local to this database, not portable identifiers across independent databases or
restores. Retain `message_rows` and `sqlite_sequence` in backups. Do not drop the
mapping table as a rollback strategy after numeric IDs have been exposed.

## Queries and bounds

An empty query returns the earliest messages. To retrieve specific anchors, pass
`row_ids` (1–100 distinct positive safe JSON integers) and optionally
`context_before`/`context_after` (0–10 each). Neighbours are selected within the
current session in `(created_at,id)` order and deduplicated. Missing and foreign
anchors both appear only in `missing_row_ids`; neither expands context.

Alternatively, use exclusive `after_row` and/or `before_row` numeric bounds.
Numeric bounds filter allocation IDs; returned rows still use chronological order.
Do not combine explicit IDs with numeric bounds or use context without anchors.

`limit` is 1–100 (default 50). `content_bytes` is 1–2048 per message (default 2048).
SQL clips content before transferring it to Go. UTF-8 clipping does not split a
code point. Each row reports the original byte count and `content_truncated`.
The encoded message-array budget is 80 KB, accounting for JSON escaping; rows may
fill that budget before `limit`. Metadata too large to fit fails rather than
silently dropping a row. Total output stays below the engine's 100 KB display cap.
Payloads and attachments are not included.

`returned` is the actual row count. `has_more` means another matching row was
observed beyond the count or byte budget. When true, `next_cursor` continues with
the same query. Cursors bind the session and normalized query to the last returned
chronological tuple; a deleted cursor row does not prevent continuation. Changing
the query or session rejects the cursor. Cursors are not authorization tokens:
the SQL session predicate applies even to a forged cursor.

Each read uses one database snapshot. Multiple pages are not a frozen export:
concurrent inserts/deletes may change later pages, and newly inserted historical
rows before the cursor will not appear on subsequent pages. Explicit missing IDs
are calculated against each read's snapshot, independently of output pagination.

Unknown arguments, nulls, fractional IDs, duplicate IDs, overflow, invalid limits,
oversized cursors and ambiguous modes fail without fallback to another scope.

## Quoted data, not execution

Responses are JSON tool results with an explicit quoted-data notice. Historical
roles are row fields, not new provider instructions. Reading instruction-like text
or stored tool metadata does not execute it or replay a tool call. This boundary
does not claim to make a language model immune to prompt injection.

## Verification

`make test-message-retrieval` runs migration, store, argument and native-dispatch
tests with the race detector three times. `make test-ux-message-retrieval` adds six
browser projects against disposable databases and a deterministic local provider.
The provider emits real `messages` calls through the production parser and engine,
then returns the received tool JSON. Browser assertions cover ID/context and window
queries, pagination, truncation, foreign isolation, native error lifecycle, quoted
rendering, reload and unsent draft retention. Live chats and auth are not test data.

## Shared38 web mapping

The canonical six-project journey now verifies 100 owned anchors with context:
exact rows 0–99, followed by exact rows 100–109 with no duplicates and no further
page. A reversed two-anchor request with context, a foreign ID and a missing ID
returns the exact deduplicated timeline across two pages. A numeric window verifies
only its three owned rows and the requested byte cap. The fixture seeds 120 owned
rows with tie-free chronology so newly submitted prompts cannot enter the expected
context union.

Native tests establish reopen/VACUUM identity durability and encoded-byte budgeting;
the browser exercises real provider calls, tool dispatch, store output and the next
provider request. Quoted hostile content is not dispatched or executed in the page.
The exact-clause review initially withheld approval for incomplete browser order/tail
assertions, then approved the strengthened source after all six projects passed.

`make test-shared-message-evidence` also verifies incomplete/duplicate matrices cannot
pass and Shared38 cannot earn Classic025 credit. Only `@shared-38` is newly mapped:
33/42 shared mappings and 101/236 Classic mappings. This is not a full-suite pass,
all-chat/family authorization, TUI, physical-device or pixel acceptance. Frozen
feature sources and supplied components are unchanged. Mapping CI is pending.

## Deployment

Exact source `05e287fb6e76f0090f2ac66ea303e6450efb4e25` passed whole-product
[CI 36263310102](https://github.com/rcarmo/gi/actions/runs/36263310102) on its first
attempt, including the new retrieval gate and all four builds. Local evidence:
race×3, six browser projects plus 12 repeat cases, 139 functional passes with 11
existing skips, core tests, vet and hook checks. Scoped source review found no
blocker. Early browser harness failures were corrected to use isolated fixture
origins, actual selectors, the events envelope and native `tool.failed` lifecycle;
no assertion/timeouts were relaxed. A core test's chronological-position assumption
was replaced by identifying the tool-result role.

The approved detached source replaced `d415541` on port 8090, PID `776843`, process
group/session `776835`. Release binary SHA-256:
`6702d8eaf3f3f7be0671261c496cea07b2fb73af6563a9859c08674d37813e47`.

All 146 messages were mapped with the expected chronological backfill; no unmapped
or mismatched rows remain. Every pre-existing table's data and schema match, apart
from the dispatcher runtime lease. Existing sequence entries match; the new sequence
is 146. Database integrity and foreign keys are clean. Counts remain 62 sessions,
51 turns, 146 messages, zero active turns. Auth hashes and the isolated local-only
TUI WIP HEAD/diff match. Four blocked-send and six read-only UI probes passed with
no live chat/auth writes. The additive migration is the only intended data change.

Evidence: `/workspace/tmp/gi-message-retrieval` and
`/workspace/tmp/gi-retrieval-deploy-05e287f`. Database backups, dumps and auth hashes
stay local and are excluded from the downloadable archive. This deployment does
not establish physical-device, pixel or TUI acceptance.
