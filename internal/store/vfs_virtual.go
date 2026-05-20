package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	storevfs "github.com/rcarmo/gi/internal/store/vfs"
)

func (s *Store) getVirtualVFSFileContent(ctx context.Context, namespace, filePath string) (*VFSFile, []byte, error) {
	switch strings.TrimSpace(namespace) {
	case storevfs.NamespaceChat:
		return s.getChatVFSFileContent(ctx, filePath)
	default:
		return nil, nil, virtualNamespaceNotFoundErr(namespace)
	}
}

func (s *Store) listVirtualVFSChildren(ctx context.Context, namespace, dir string) ([]VFSListEntry, error) {
	switch strings.TrimSpace(namespace) {
	case storevfs.NamespaceChat:
		return s.listChatVFSChildren(ctx, dir)
	default:
		return nil, virtualNamespaceNotFoundErr(namespace)
	}
}

func virtualNamespaceNotFoundErr(namespace string) error {
	return fmt.Errorf("virtual vfs namespace not found: %s", namespace)
}

func (s *Store) listChatVFSChildren(ctx context.Context, dir string) ([]VFSListEntry, error) {
	switch dir {
	case "":
		return []VFSListEntry{
			{Name: "README.md", Path: "README.md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_readme"}},
			{Name: "sessions", Path: "sessions", IsDir: true, Metadata: map[string]any{"virtual": true, "kind": "chat_sessions"}},
		}, nil
	case "sessions":
		sessions, err := s.ListSessions(ctx)
		if err != nil {
			return nil, err
		}
		entries := []VFSListEntry{{Name: "index.md", Path: "sessions/index.md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_sessions_index"}}}
		for _, session := range sessions {
			entries = append(entries, VFSListEntry{Name: session.ID, Path: "sessions/" + session.ID, IsDir: true, Metadata: map[string]any{"virtual": true, "title": session.Title}})
		}
		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		})
		return entries, nil
	}
	parts := storevfs.SplitVirtualPath(dir)
	if len(parts) >= 2 && parts[0] == "sessions" {
		sessionID := parts[1]
		if _, err := s.GetSession(ctx, sessionID); err != nil {
			return nil, err
		}
		if len(parts) == 2 {
			return []VFSListEntry{
				{Name: "session.md", Path: "sessions/" + sessionID + "/session.md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_session"}},
				{Name: "messages", Path: "sessions/" + sessionID + "/messages", IsDir: true, Metadata: map[string]any{"virtual": true, "kind": "chat_messages"}},
				{Name: "turns", Path: "sessions/" + sessionID + "/turns", IsDir: true, Metadata: map[string]any{"virtual": true, "kind": "chat_turns"}},
			}, nil
		}
		if len(parts) == 3 && parts[2] == "messages" {
			messages, err := s.ListMessages(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			entries := []VFSListEntry{{Name: "index.md", Path: "sessions/" + sessionID + "/messages/index.md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_messages_index"}}}
			for _, msg := range messages {
				entries = append(entries, VFSListEntry{Name: msg.ID + ".md", Path: "sessions/" + sessionID + "/messages/" + msg.ID + ".md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_message", "role": msg.Role}, OriginalSize: len(msg.Content)})
			}
			return entries, nil
		}
		if len(parts) == 3 && parts[2] == "turns" {
			turns, err := s.ListTurns(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			entries := []VFSListEntry{{Name: "index.md", Path: "sessions/" + sessionID + "/turns/index.md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_turns_index"}}}
			for _, tr := range turns {
				entries = append(entries, VFSListEntry{Name: tr.ID + ".md", Path: "sessions/" + sessionID + "/turns/" + tr.ID + ".md", IsDir: false, ContentType: "text/markdown", Metadata: map[string]any{"virtual": true, "kind": "chat_turn", "status": tr.Status}, OriginalSize: len(tr.Prompt)})
			}
			return entries, nil
		}
	}
	return []VFSListEntry{}, nil
}

func (s *Store) getChatVFSFileContent(ctx context.Context, filePath string) (*VFSFile, []byte, error) {
	if filePath == "" || filePath == "README.md" {
		body := strings.TrimSpace(`# Chat VFS

This namespace projects chat/session runtime state as read-only markdown documents.

## Paths

- vfs://chat/README.md
- vfs://chat/sessions/index.md
- vfs://chat/sessions/<session-id>/session.md
- vfs://chat/sessions/<session-id>/messages/index.md
- vfs://chat/sessions/<session-id>/messages/<message-id>.md
- vfs://chat/sessions/<session-id>/turns/index.md
- vfs://chat/sessions/<session-id>/turns/<turn-id>.md

All files contain markdown with JSON-compatible frontmatter so models can inspect state without SQL.
`) + "\n"
		return newVirtualVFSFile(storevfs.NamespaceChat, "README.md", "chat_readme", body, "", ""), []byte(body), nil
	}
	if filePath == "sessions/index.md" {
		sessions, err := s.ListSessions(ctx)
		if err != nil {
			return nil, nil, err
		}
		lines := make([]string, 0, len(sessions)+2)
		lines = append(lines, "# Sessions", "")
		for _, sess := range sessions {
			lines = append(lines, fmt.Sprintf("- `%s` — %s", sess.ID, sess.Title))
			lines = append(lines, fmt.Sprintf("  - session: `vfs://chat/sessions/%s/session.md`", sess.ID))
			lines = append(lines, fmt.Sprintf("  - messages: `vfs://chat/sessions/%s/messages/index.md`", sess.ID))
			lines = append(lines, fmt.Sprintf("  - turns: `vfs://chat/sessions/%s/turns/index.md`", sess.ID))
		}
		body := strings.Join(lines, "\n") + "\n"
		return newVirtualVFSFile(storevfs.NamespaceChat, filePath, "chat_sessions_index", body, "", ""), []byte(body), nil
	}

	parts := storevfs.SplitVirtualPath(filePath)
	if len(parts) < 2 || parts[0] != "sessions" {
		return nil, nil, sql.ErrNoRows
	}
	sessionID := parts[1]
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	if len(parts) == 3 && parts[2] == "session.md" {
		meta := map[string]any{
			"kind":              "chat/session",
			"session_id":        session.ID,
			"title":             session.Title,
			"parent_session_id": session.ParentSessionID,
			"created_at":        session.CreatedAt,
			"updated_at":        session.UpdatedAt,
			"aliases":           session.Aliases,
			"scope":             session.Scope,
			"state":             session.State,
		}
		body := fmt.Sprintf("# Session %s\n\nTitle: %s\n", session.ID, session.Title)
		raw := storevfs.RenderFrontmatterMarkdown(meta, body)
		return newVirtualVFSFile(storevfs.NamespaceChat, filePath, "chat_session", string(raw), session.CreatedAt, session.UpdatedAt), raw, nil
	}

	if len(parts) == 4 && parts[2] == "messages" && parts[3] == "index.md" {
		messages, err := s.ListMessages(ctx, sessionID)
		if err != nil {
			return nil, nil, err
		}
		lines := []string{"# Messages", ""}
		for _, msg := range messages {
			lines = append(lines, fmt.Sprintf("- `%s` [%s] `%s`", msg.ID, msg.Role, msg.CreatedAt))
			lines = append(lines, fmt.Sprintf("  - `vfs://chat/sessions/%s/messages/%s.md`", sessionID, msg.ID))
		}
		body := strings.Join(lines, "\n") + "\n"
		return newVirtualVFSFile(storevfs.NamespaceChat, filePath, "chat_messages_index", body, session.CreatedAt, session.UpdatedAt), []byte(body), nil
	}

	if len(parts) == 4 && parts[2] == "messages" && strings.HasSuffix(parts[3], ".md") {
		messageID := strings.TrimSuffix(parts[3], ".md")
		msg, err := s.getMessageByID(ctx, messageID)
		if err != nil {
			return nil, nil, err
		}
		if msg.SessionID != sessionID {
			return nil, nil, sql.ErrNoRows
		}
		meta := map[string]any{
			"kind":       "chat/message",
			"session_id": msg.SessionID,
			"message_id": msg.ID,
			"role":       msg.Role,
			"created_at": msg.CreatedAt,
			"payload":    msg.Payload,
		}
		body := msg.Content + "\n"
		raw := storevfs.RenderFrontmatterMarkdown(meta, body)
		return newVirtualVFSFile(storevfs.NamespaceChat, filePath, "chat_message", string(raw), msg.CreatedAt, msg.CreatedAt), raw, nil
	}

	if len(parts) == 4 && parts[2] == "turns" && parts[3] == "index.md" {
		turns, err := s.ListTurns(ctx, sessionID)
		if err != nil {
			return nil, nil, err
		}
		lines := []string{"# Turns", ""}
		for _, tr := range turns {
			lines = append(lines, fmt.Sprintf("- `%s` [%s/%s] `%s`", tr.ID, tr.Status, tr.Phase, tr.CreatedAt))
			lines = append(lines, fmt.Sprintf("  - `vfs://chat/sessions/%s/turns/%s.md`", sessionID, tr.ID))
		}
		body := strings.Join(lines, "\n") + "\n"
		return newVirtualVFSFile(storevfs.NamespaceChat, filePath, "chat_turns_index", body, session.CreatedAt, session.UpdatedAt), []byte(body), nil
	}

	if len(parts) == 4 && parts[2] == "turns" && strings.HasSuffix(parts[3], ".md") {
		turnID := strings.TrimSuffix(parts[3], ".md")
		tr, err := s.GetTurn(ctx, turnID)
		if err != nil {
			return nil, nil, err
		}
		if tr.SessionID != sessionID {
			return nil, nil, sql.ErrNoRows
		}
		meta := map[string]any{
			"kind":        "chat/turn",
			"session_id":  tr.SessionID,
			"turn_id":     tr.ID,
			"status":      tr.Status,
			"phase":       tr.Phase,
			"created_at":  tr.CreatedAt,
			"updated_at":  tr.UpdatedAt,
			"claimed_by":  tr.ClaimedBy,
			"claimed_at":  tr.ClaimedAt,
			"started_at":  tr.StartedAt,
			"finished_at": tr.FinishedAt,
			"metadata":    tr.Metadata,
		}
		body := fmt.Sprintf("# Turn %s\n\n## Prompt\n\n%s\n", tr.ID, tr.Prompt)
		raw := storevfs.RenderFrontmatterMarkdown(meta, body)
		return newVirtualVFSFile(storevfs.NamespaceChat, filePath, "chat_turn", string(raw), tr.CreatedAt, tr.UpdatedAt), raw, nil
	}

	return nil, nil, sql.ErrNoRows
}

func newVirtualVFSFile(namespace, filePath, kind, body, createdAt, updatedAt string) *VFSFile {
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(createdAt) == "" {
		createdAt = now
	}
	if strings.TrimSpace(updatedAt) == "" {
		updatedAt = createdAt
	}
	size := len(body)
	return &VFSFile{
		Namespace:      namespace,
		Path:           filePath,
		ContentType:    "text/markdown",
		Metadata:       map[string]any{"virtual": true, "kind": kind, "size": size},
		OriginalSize:   size,
		CompressedSize: size,
		Compressed:     false,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}

func (s *Store) getMessageByID(ctx context.Context, id string) (*Message, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, session_id, role, content, payload_json, created_at
		from messages
		where id = ?
	`, id)
	var item Message
	var payloadJSON string
	if err := row.Scan(&item.ID, &item.SessionID, &item.Role, &item.Content, &payloadJSON, &item.CreatedAt); err != nil {
		return nil, err
	}
	payload, err := unmarshalJSONMap(payloadJSON)
	if err != nil {
		return nil, err
	}
	item.Payload = payload
	return &item, nil
}
