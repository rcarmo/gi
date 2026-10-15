# Web conversation projection and idle activity

The screenshot3130 failures had two sources: raw internal history was projected as conversation posts, and idle snapshots retained tool/activity data that the app rendered as current work. This change fixes the bounded projection and idle-state paths. Running-tool Output presentation still needs the installed Piclaw component.

## Conversation view

The browser requests `view=conversation` on message-page and message-search endpoints. Filtering happens before pagination or search limits, so long runs of hidden tool results do not displace real conversation messages.

- User messages retain their original text even if metadata resembles tool records.
- Assistant tool-call summaries retain prose and omit synthetic `[tool_call: …]` suffixes. New summaries store explicit `display_text`; legacy `kind=tool_calls` rows use the known storage delimiter.
- Tool-result and unknown internal roles do not become timeline posts.
- System records remain visible with an explicit System identity. Both initial history/search and incoming `system_message` SSE frames use the same visual role; they cannot fall through to the user's name/avatar.
- Ordinary assistant Markdown, message IDs, media references, recovery content blocks and link previews remain intact.

The unpaged export API and default raw pagination/search are unchanged. No message, tool result, payload or model-history record is deleted or rewritten. Invalid view values fail with HTTP400; existing cursor scope checks remain in force.

The supplied Post component has two visual roles, so the adapter uses bot presentation with a reserved System identity. The exact final error styling is not yet full Piclaw parity. Protected component/UI/pane files and historical Gherkin snapshots were not edited.

## Activity state

Idle and terminal activity metadata remain available through the native API for diagnostics and controls, but no longer fabricate Working/Idle/completed-tool footers. A completed tool within an active run becomes Waiting for model. Retry and cancellation notices keep their existing ownership. Terminal SSE and idle refreshes clear transient Draft/Thought previews without changing the composer draft.

The installed Piclaw lifecycle probe established absence of idle activity and transition from tool execution to post-tool waiting. `@ux-original-027` and the chat-lifecycle scenarios remain unmapped: the complete running tool Output pane, disclosure, timing and multi-tool lifecycle are not implemented by this slice.

## Tests and corrected expectations

- Store tests cover stable paging across 78 hidden tool records, chronological cursors, cross-session rejection, display-text search, raw-history preservation and user-authored tool syntax.
- Native HTTP tests cover opt-in view, raw compatibility, search and invalid views.
- Chromium and WebKit browser tests use a disposable database with the screenshot's role pattern. They assert visible author identity, Markdown, hidden raw tools, idle reload, retained draft, live system SSE and terminal preview cleanup. Activity transitions and the live-system frame are bounded fixtures.
- `make test-ux-tool-terminal`: 12 native lifecycle cases passed across six browser/viewport projects. Persisted timing/cancellation metadata remains checked without expecting terminal footers to stay visible.
- Final `make check`: Go/vet/build/hooks and 142 functional tests passed; 11 functional tests skipped. This includes the user-metadata preservation guard and live-event preview cleanup.
- `make ux-parity-inventory`: 200 support tests / 7,723 assertions pass. The isolated asset host accepts the conversation query only for Gi message/search routes; malformed views still fail closed.
- Focused projection review found no blocker; an earlier broad review attempt timed out.

Two old functional tests encoded the defect: persistent completed-tool display and using an idle pane as a swipe surface. The former now verifies durable timing with no idle footer; the latter uses an explicitly active fixture while retaining session-swipe and draft checks. The old tool-lifecycle browser tests were updated consistently. No test was removed and no new parity credit was granted.

Evidence logs are under `/workspace/tmp/gi-conversation-*`. No production restart, migration, live-chat write or deployment was performed.
