@derived-index @proposal
Feature: Scoped workspace indexing lifecycle
  This is a Gi implementation target derived from pinned source implementations.
  It is not part of the frozen Classic/shared corpus and has no runtime pass credit.
  Provenance and intentional differences are in docs/internal/search/indexing-lineage-20260922.md.

  @index-derived-001 @piclaw
  Scenario: Resolve configured workspace roots and supported file types
    Given roots and supported extensions are resolved for one workspace identity
    When a scope is indexed
    Then only regular eligible files under that scope's roots are candidates
    And excluded directories, unsupported types and oversized files are not indexed
    And notes and skills remain separate scopes even when all includes both

  @index-derived-002 @piclaw
  Scenario: Incrementally refresh changed files without replacing unchanged entries
    Given a committed index contains files in the selected scope
    When a refresh finds an unchanged file, a changed file and a new file
    Then the unchanged file retains its document and chunk identities
    And changed and new content becomes searchable with source locations
    And content from replaced chunks stops matching

  @index-derived-003 @piclaw @gi-strengthening
  Scenario: Delete only completed-scan memberships in the requested scope
    Given two scopes have overlapping roots and a third scope is unrelated
    When a complete refresh no longer finds a file in the first scope
    Then its first-scope membership is removed
    And its content remains indexed while another scope still owns it
    And documents with no memberships are removed with their chunks and FTS rows
    And the unrelated scope retains its documents and status

  @index-derived-004 @piclaw
  Scenario: Search does not await a background refresh by default
    Given the selected scope is missing or stale
    When a search without explicit refresh is requested
    Then it returns the currently committed matching hits within the query bounds
    And requests one bounded background refresh for that workspace
    And it does not label the stale snapshot as freshly indexed

  @index-derived-005 @piclaw
  Scenario: Explicit refresh waits for indexing to finish
    Given the selected scope has changed files
    When a search requests explicit refresh or the user chooses Reindex workspace
    Then the caller waits for the rebuild outcome
    And successful completion returns the committed index status
    And a failed rebuild reports an error instead of a false success

  @index-derived-006 @piclaw @gi-strengthening
  Scenario: Keep durable status separate from transient worker activity
    Given no committed index exists for a scope
    Then its status is never_indexed
    When indexing begins
    Then its status is indexing with the last committed count and timestamp
    When indexing commits successfully
    Then status becomes ready with a new count, timestamp and generation
    When a watched eligible path changes
    Then affected scopes become stale and request background refresh
    And unrelated scopes keep their status

  @index-derived-007 @piclaw @gi-strengthening
  Scenario: A failed or cancelled refresh preserves the committed snapshot
    Given a scope has committed searchable documents and durable status
    When scanning or writing fails, or the refresh is cancelled before commit
    Then no partial replacement or missing-file cleanup becomes visible
    And the last committed documents, count, timestamp and generation remain intact
    And failure detail is recorded without claiming a new successful refresh
    And reopening the database retains that failure and committed snapshot

  @index-derived-008 @piclaw @gi-strengthening
  Scenario: Bound traversal and treat incomplete scans as incomplete
    Given a refresh reaches an inventory bound or cannot read a required directory
    When the refresh cannot establish a complete inventory
    Then it does not delete unseen files as if the directory were empty
    And it reports the limit or scan error
    And it does not publish ready for an incomplete snapshot

  @index-derived-009 @gi-strengthening
  Scenario: Coordinate competing writers with a fenced workspace lease
    Given one worker owns an unexpired lease for a workspace
    When another worker requests a refresh for an overlapping scope
    Then at most one worker may commit workspace index changes
    When an expired lease is replaced with a new owner token
    Then the old owner cannot renew or commit after takeover
    And restart recovery can mark interrupted work stale without discarding committed data

  @index-derived-010 @piclaw
  Scenario: Rank and bound lexical hits with a safe fallback
    Given committed text matches the requested roots or scope
    When a lexical query is executed
    Then hits include path, source locations and a snippet with bounded limit and offset
    And lexical relevance orders the results
    And malformed FTS syntax uses an explicitly defined literal fallback
    And fallback search cannot escape the requested workspace and scope

  @index-derived-011 @tau @vibes @schema-candidate
  Scenario: Keep chunk identity and FTS rows consistent transactionally
    Given a document has stable chunk IDs and searchable content
    When a chunk is inserted, changed or deleted
    Then its FTS row has exactly the same row ID and current canonical content
    When the document path changes
    Then old path terms stop matching and new path terms match
    When the transaction rolls back
    Then both canonical rows and FTS matches return to the committed state

  @index-derived-012 @tau @vibes @schema-candidate
  Scenario: Repair an index from canonical content without resurrecting deleted rows
    Given documents and chunks are the canonical searchable content
    When FTS is rebuilt and its external-content integrity is checked
    Then the rebuilt matches correspond to those rows
    And removed documents or orphaned memberships do not reappear
    And workspace and scope isolation still applies

  @index-derived-013 @gi-strengthening
  Scenario: Change roots or index versions without reusing incompatible ready status
    Given persisted status records workspace identity and configuration and chunker versions
    When the configured root, eligible extensions or chunker version changes
    Then the previous index is not silently labelled ready for the new configuration
    And a successful migration or refresh records the new configuration explicitly
    And embedding model or dimension changes invalidate only their separate vector representation

  @index-derived-014 @gi-strengthening
  Scenario: Keep index operations outside chat submission
    Given the workspace pane has an unsent chat draft and file references
    When the user starts, retries or observes a failed index operation
    Then no prompt, steering item or queued chat message is created
    And draft content and references remain owned by their session
    And switching sessions does not move an index error into another draft

  @index-derived-015 @tui-design
  Scenario: Expose index status without permanent terminal chrome
    Given the terminal editor contains unsent text and the transcript is scrolled up
    When an explicit workspace index action is opened
    Then status and failure details use a temporary bounded surface
    And Escape restores the editor text, cursor and reading position
    And ordinary keys never submit the draft through that surface
    And idle editor, footer and transcript padding keep their existing row counts
    And separate acceptance is required at 60 by 18, 100 by 22 and 140 by 36

  @index-derived-016 @gi-strengthening
  Scenario: Load index root policy explicitly at startup
    Given workspaceIndex settings specify extra roots, extra extensions and optional roots
    When the server loads that workspace configuration
    Then extra roots widen only the all scope
    And notes and skills retain their own root ownership
    And invalid or covered optional roots are rejected instead of widening the scope
    And changing root policy requires configuration reload and changes its fingerprint
    And old results do not match queries under an incompatible fingerprint

  @index-derived-017 @gi-strengthening
  Scenario: Optional root absence cannot erase previously committed content
    Given a scope explicitly marks a distinct root optional
    When that root is absent during inventory and final verification
    Then it may be reported missing in a complete scan
    But publication checks missing-root memberships in its commit transaction
    And if that root owns committed documents in this scope publication fails
    And the previous content, count, generation and successful timestamp remain intact
    When the root is restored and an explicit retry succeeds
    Then the scope publishes one new generation

  @index-derived-018 @gi-strengthening
  Scenario: Invalidation during refresh cannot be acknowledged by an older scan
    Given a refresh captures the scope's requested invalidation revision before scanning
    When a new relevant path change is durably recorded before or during publication
    Then the scope's requested revision advances without altering committed results
    And successful publication acknowledges only the revision captured before scanning
    And a newer pending revision leaves the published scope stale
    And a subsequent successful refresh can acknowledge that newer revision
    When scanning, publication or failure cleanup fails
    Then no pending revision is acknowledged
    And reopening the database retains the pending revision and committed index

  @index-derived-019 @gi-strengthening
  Scenario: Coalesce bounded refresh work without hiding pending revisions
    Given a scheduler owns a fixed set of scopes for one workspace
    When concurrent callers request the same scope
    Then they share one bounded refresh batch and completion outcome
    And at most one local worker runs across those scopes at a time
    And a caller cancelling its wait does not cancel other callers' work
    When a successful scan leaves a newer invalidation pending
    Then the scheduler requests a follow-up within the same attempt budget
    And it reports success only after observing ready status with acknowledged revisions
    When another process owns the workspace lease
    Then retry delays are capped and attempts are bounded
    And a newer matching peer publication can satisfy the batch without a duplicate scan
    When a scan fails or the attempt budget is exhausted
    Then the batch returns an error without clearing durable failure or pending revisions

  @index-derived-020 @gi-strengthening
  Scenario: Stop scheduler work before closing the database
    Given a scheduler has active and queued refresh requests
    When its application owner cancels and closes it
    Then active scanning and retry waits are cancelled
    And close waits for worker renewal and failure cleanup to finish
    And every queued caller receives a terminal result
    And requests after cancellation are rejected
    And no scheduler database operation occurs after close returns

  @index-derived-021 @gi-strengthening
  Scenario: Explicit web refresh shares application-owned work
    Given application startup owns the configured scope scheduler
    Then startup and GET status or search do not scan or request refresh
    When concurrent authenticated Reindex requests wait for the same scope
    Then they share a bounded refresh batch
    And disconnecting one caller cancels only its wait
    And successful responses follow committed matching ready status
    When application shutdown or a listener failure stops the server
    Then new refresh requests are rejected
    And HTTP handlers and index cleanup are joined before closing the database

  @index-derived-022 @gi-strengthening
  Scenario: Native filesystem writes record invalidation without automatic refresh
    Given a native tool, HTTP tool or script bridge writes a regular workspace file
    Then affected configured scopes record an atomic invalidation before mutation
    And a failed pre-notification prevents the filesystem write
    And a second bounded notification follows an attempted write even after caller cancellation
    And later publication cannot acknowledge a revision captured before that second notification
    And unrelated scopes and VFS writes do not advance filesystem index revisions
    And GET queries retain the committed snapshot until explicit refresh
    When post-notification fails
    Then the error reports that bytes may have changed and explicit reindex is required
    And this protocol does not claim crash-atomic filesystem and database mutation
