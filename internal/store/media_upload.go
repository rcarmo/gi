package store

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	storeobject "github.com/rcarmo/gi/internal/store/object"
)

// CreateOrReuseWebMedia gives browser retries stable native references without
// changing CreateMedia (tools/TUI/API callers may intentionally create copies).
// A hash narrows the candidates but never substitutes for exact original bytes.
func (s *Store) CreateOrReuseWebMedia(ctx context.Context, sessionID, filename, contentType string, raw []byte) (*Media, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("create web media: session id required")
	}
	if len(raw) > 10<<20 {
		return nil, fmt.Errorf("media exceeds 10 MiB limit")
	}
	if filename == "" {
		filename = "attachment"
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	hash := sha256.Sum256(raw)
	digest := hex.EncodeToString(hash[:])
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Read the response inside the transaction: once Commit succeeds, caller
	// cancellation must not turn that success into a second database read error.
	finish := func(id int64) (*Media, error) {
		var item Media
		var meta string
		err := tx.QueryRowContext(ctx, `SELECT id,session_id,filename,content_type,metadata_json,original_size,compressed_size,compressed,created_at,updated_at FROM media WHERE id=?`, id).Scan(&item.ID, &item.SessionID, &item.Filename, &item.ContentType, &meta, &item.OriginalSize, &item.CompressedSize, &item.Compressed, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		item.Metadata, err = unmarshalJSONMap(meta)
		if err != nil {
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &item, nil
	}
	// File databases use BEGIN IMMEDIATE. Take a write reservation explicitly for
	// deferred in-memory connections as well, before the read→insert decision.
	if _, err = tx.ExecContext(ctx, `UPDATE media SET id=id WHERE 0`); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,compressed,content FROM media
  WHERE session_id=? AND filename=? AND content_type=? AND original_size=?
  AND json_extract(metadata_json,'$.web_upload_sha256')=? ORDER BY id`, sessionID, filename, contentType, len(raw), digest)
	if err != nil {
		return nil, err
	}
	var existing int64
	for rows.Next() {
		var id int64
		var compressed bool
		var candidate []byte
		if err = rows.Scan(&id, &compressed, &candidate); err != nil {
			rows.Close()
			return nil, err
		}
		// Corrupt stored candidates fail closed. Do not silently create a new
		// identity or rewrite an existing message's bytes to repair corruption.
		if compressed {
			reader, e := gzip.NewReader(bytes.NewReader(candidate))
			if e != nil {
				rows.Close()
				return nil, fmt.Errorf("reuse web media: %w", e)
			}
			candidate, e = io.ReadAll(io.LimitReader(reader, int64(len(raw))+1))
			reader.Close()
			if e != nil {
				rows.Close()
				return nil, e
			}
		}
		if bytes.Equal(candidate, raw) {
			existing = id
			break
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if existing != 0 {
		return finish(existing)
	}
	stored, compressed, err := storeobject.MaybeCompressBlob(raw)
	if err != nil {
		return nil, err
	}
	metadata := map[string]any{"source": "web", "size": len(raw), "sha256": digest, "web_upload_sha256": digest}
	if len(raw) > 0 {
		metadata["detected_content_type"] = http.DetectContentType(raw)
	}
	encoded, err := marshalJSON(metadata)
	if err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO media(session_id,filename,content_type,metadata_json,original_size,compressed_size,compressed,content,created_at,updated_at)
 VALUES(?,?,?,?,?,?,?,?,`+defaultNow+`,`+defaultNow+`)`, sessionID, filename, contentType, encoded, len(raw), len(stored), boolToInt(compressed), stored)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return finish(id)
}
