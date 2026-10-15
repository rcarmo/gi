# Classic025: installed message retrieval and Gi scope

The installed Piclaw 3.2.4 `messages` tool reads owned row IDs across sessions in family-shared mode. Gi's native `messages` tool reads only the current runtime session, so Classic025 stays unmapped. Shared38 retains its narrower current-session mapping.

## Disposable oracle

Run `make test-piclaw-message-retrieval`. `tests/ux/oracle/piclaw-message-retrieval-probe.mjs` imports the installed Piclaw runtime, sets `PICLAW_DB_IN_MEMORY=1`, creates a temporary workspace, checks `PRAGMA database_list` for an empty main filename, and removes the workspace afterwards. It seeds an owner with a home, child and second root; a different owner has a fourth root. The probe creates a fixture login and server-side execution identity, then calls `runMessagesTool`. It does not call live HTTP routes or read or write the live store.

| Scope | Observed `get` and row window behaviour |
| --- | --- |
| Single-user | Unqualified `get` of row IDs `[5,3,3,-1,0,999999]` returned `[5,3]` and reported `[999999]` missing. An explicit home chat returned `[3]`, reporting `[5]` missing. A cross-chat `diff` with `before_row`, `chat_jid: 'all'` and `limit: 2` returned `[4,5]`. |
| Bounded content | `get` returned the selected lines `second`, `third` and same-chat before/after rows. Requested context sizes of 100 were capped at 20. Invalid `content_lines` returned `invalid_content_lines`. |
| Family-shared | Without a trusted identity, `get` returned `access_denied`. Under the owner's live identity, unqualified `get` returned `[2,4,3]` across the child, second root and home, reporting the other owner's `[5]` missing. `chat_jid: 'all'` yielded the same owned rows. Explicit second-root lookup returned `[4]`; explicit foreign chat was denied. |
| Family windows | Default `diff` after row 1 returned home rows `[3,6]`. With `chat_jid: 'all'`, it returned owned rows `[2,3,4,6]`; a limited `before_row` window returned `[3,4]`. An explicit foreign-chat window was denied. Expiring the fixture web login caused the same cached run identity to be denied. |

The probe uses synthetic IDs, so these numbers are fixture row IDs, not live message references. Piclaw's `fetchByRowId` calls `appendSenderFilter`, which applies the owned-chat SQL condition even without a sender filter. Context queries also constrain `chat_jid` to the selected row's chat. The previously suspected row-ID ownership leak was not reproduced.

## Gi decision before implementation

`internal/tools/messages.go` requires `rt.SessionID` and calls `internal/store/message_retrieval.go` with that session. It exposes neither another-session selection nor an all-chat selector. This is the intended Shared38 boundary. Extending Classic025 needs a server-owned principal attached to tool execution, a durable Gi session-owner/root relation, a fresh permission check for selected sessions, and row/window queries that apply that permission to every row and context row. A model-supplied session ID alone must not grant access. Define single-user versus family/all-chat scope and archived-session policy before changing the API; preserve indistinguishable missing versus foreign row IDs and bounded output. The current Gi store/session model supplies no equivalent family-owned authorisation contract.

This is runtime-method evidence in an in-memory fixture. Production Piclaw authentication, live routing, Gi family-mode policy, deployment and whole-web parity have not been accepted by this probe.
