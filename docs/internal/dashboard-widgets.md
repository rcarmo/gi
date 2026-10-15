# Dashboard widget backend

`send_dashboard_widget` persists a self-contained HTML artifact as an assistant timeline message and publishes a `new_post` event after the transaction commits. The front-end host and bridge are owned by the fixtures-vibes agent. Gi does not yet claim `@cap-widgets`.

## Tool

Required: `html`, non-blank and at most 256 KiB.

Optional fields:

| Field | Default | Limit |
|---|---|---|
| `title` | `Generated widget` | 160 bytes |
| `content` | `Widget ready — open to interact.` | 4096 bytes |
| `open_label` | `Open widget` | 80 bytes |
| `interactive` | `true` | Boolean |
| `widget_id` | Generated ID | 128 bytes; letters, digits, `_`, `.`, `:`, `-` |
| `chat_jid` | Current `gi:{session}` | Must name the current session |

Unknown arguments and invalid types fail before persistence. An explicit widget ID cannot replace an existing widget in the same session. The same ID may exist independently in another session.

The response contains `status: "posted"`, `tool`, `widget_id`, `chat_jid` and `post_id`. After a successful post, the runner finishes the admitted tool-call batch and completes the turn without another model request. Queued work follows the normal turn-completion path.

## Persisted message and event

The message has role `assistant`, fallback `content`, and this payload:

```json
{
  "kind": "dashboard_widget",
  "source": "send_dashboard_widget",
  "turn_id": "turn-id",
  "content_blocks": [{
    "type": "generated_widget",
    "widget_id": "widget-id",
    "title": "Generated widget",
    "open_label": "Open widget",
    "interactive": true,
    "capabilities": ["interactive"],
    "artifact": {"kind": "html", "html": "<p>Example</p>"}
  }]
}
```

Non-interactive blocks have an empty capabilities array. Conversation pages return the persisted blocks through the normal message payload. The `new_post` SSE event includes `id`, `chat_jid`, `turn_id`, `content`, `timestamp`, `sender: "agent"`, `is_bot_message: true` and `data.content_blocks`. `data.thread_id` names the owning turn's prompt when available.

The backend stores HTML as data and never executes it. Model-history projection includes fallback text; it does not turn the artifact payload into model content. The model's original tool arguments remain in its tool-call history.

## Artifact lookup and bridge

Authenticated `GET /api/sessions/{session}/widgets/{widget_id}` returns:

```json
{
  "widget_id": "widget-id",
  "post_id": "message-id",
  "chat_jid": "gi:session-id",
  "content_block": {"type": "generated_widget"}
}
```

`content_block` is the complete stored block. Missing artifacts and lookup under another session return 404. Other methods return 405. Reads neither change queues nor admit turns.

Front-end bridge contract:

- `widget.submit`: send validated non-empty text to the origin session's existing prompt endpoint, with its usual client request ID and queue admission rules.
- `widget.request_refresh`: rebuild the dashboard snapshot through host session routes and deliver a host update. The authenticated artifact lookup is host-only, not fetched by the opaque-origin frame.
- `widget.close`: close the pane locally. No HTTP request and no queue mutation.

The shared frontend applies Rui's strict sandbox: interactive HTML frames allow scripts but never `allow-same-origin`. The host checks iframe source and session key for every bridge message. Public embedded JavaScript grants `Access-Control-Allow-Origin: *`, including Brotli/gzip responses, so opaque-origin frames can import modules. Authenticated APIs grant no CORS.

## Verification

`make test-dashboard-widgets` covers validation, transaction rollback, persistence across reopen, duplicate IDs, session isolation, HTTP authentication, SSE publication after commit and completion after the tool batch. Independent Plan/widget/extra014 acceptance passes 66/66 across Chromium/WebKit phone, tablet and desktop at fixtures4259e82; the profile claims `@cap-widgets`. CORS/auth/Plan/widget handler checks pass 15/15; full Go tests pass 2,122, and functional tests pass 142 with 11 skips. See [acceptance and profiles](../implementation/maintenance/classic-ui-acceptance.md).
