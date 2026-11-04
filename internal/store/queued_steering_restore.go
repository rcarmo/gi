package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Called with the SQLite writer slot held. A runner either dequeues first (we
// reject a claimed row) or sees our committed cancellation, never both. Direct
// steering becomes dequeued atomically before leaving the store; those entries
// are already delivered and are deliberately outside the pending queue.
func pendingTextSteeringForRestore(ctx context.Context, tx *sql.Tx, sessionID string) ([]SteeringMessage, error) {
	var claimed bool
	if err := tx.QueryRowContext(ctx, `select exists(select 1 from steering_queue where session_id=? and status='claimed')`, sessionID).Scan(&claimed); err != nil {
		return nil, err
	}
	if claimed {
		return nil, ErrQueueConflict
	}
	rows, err := tx.QueryContext(ctx, `select id, role, content, payload_json, media_json, coalesce(source_queue_id,'') from steering_queue where session_id=? and status='queued' order by id`, sessionID)
	if err != nil {
		return nil, err
	}
	var out []SteeringMessage
	for rows.Next() {
		var msg SteeringMessage
		var payload, media string
		if err = rows.Scan(&msg.ID, &msg.Role, &msg.Content, &payload, &media, &msg.SourceQueueID); err != nil {
			break
		}
		msg.Payload, err = unmarshalJSONMap(payload)
		if err != nil {
			break
		}
		msg.Media, err = unmarshalJSONStringArray(media)
		if err != nil {
			break
		}
		if msg.Role != "user" || strings.TrimSpace(msg.Content) == "" || !utf8.ValidString(msg.Content) || len(msg.Media) > 0 || msg.Payload["media"] != nil || msg.Payload["tui_media_claim"] != nil || msg.Payload["kind"] == "context_reset" || msg.Payload["operation"] == "manual_compaction" {
			err = fmt.Errorf("%w: pending media or non-text steering requires /queue", ErrQueueConflict)
			break
		}
		out = append(out, msg)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return nil, err
	}
	for _, msg := range out {
		if msg.SourceQueueID == "" {
			continue
		}
		var valid bool
		if err := tx.QueryRowContext(ctx, `select exists(select 1 from turns where id=? and session_id=? and status='steering' and prompt=?)`, msg.SourceQueueID, sessionID, msg.Content).Scan(&valid); err != nil {
			return nil, err
		}
		if !valid {
			return nil, ErrQueueConflict
		}
	}
	return out, nil
}
