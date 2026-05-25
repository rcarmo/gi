package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	storeobject "github.com/rcarmo/gi/internal/store/object"
	storevfs "github.com/rcarmo/gi/internal/store/vfs"
)

type Media struct {
	ID             int64          `json:"id"`
	SessionID      string         `json:"session_id"`
	Filename       string         `json:"filename"`
	ContentType    string         `json:"content_type"`
	Metadata       map[string]any `json:"metadata"`
	OriginalSize   int            `json:"original_size"`
	CompressedSize int            `json:"compressed_size"`
	Compressed     bool           `json:"compressed"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type VFSFile struct {
	Namespace      string         `json:"namespace"`
	Path           string         `json:"path"`
	ContentType    string         `json:"content_type"`
	Metadata       map[string]any `json:"metadata"`
	OriginalSize   int            `json:"original_size"`
	CompressedSize int            `json:"compressed_size"`
	Compressed     bool           `json:"compressed"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type VFSListEntry struct {
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	IsDir        bool           `json:"is_dir"`
	ContentType  string         `json:"content_type,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	OriginalSize int            `json:"original_size,omitempty"`
}

func (s *Store) CreateMedia(ctx context.Context, sessionID, filename, contentType string, raw []byte, metadata map[string]any) (*Media, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("create media: missing session id")
	}
	if filename == "" {
		filename = "attachment"
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	compressedBlob, compressed, err := storeobject.MaybeCompressBlob(raw)
	if err != nil {
		return nil, fmt.Errorf("create media: compress: %w", err)
	}
	meta := copyMetadata(metadata)
	meta["size"] = len(raw)
	if _, ok := meta["sha256"]; !ok {
		sum := sha256.Sum256(raw)
		meta["sha256"] = hex.EncodeToString(sum[:])
	}
	if _, ok := meta["detected_content_type"]; !ok && len(raw) > 0 {
		meta["detected_content_type"] = http.DetectContentType(raw)
	}
	metadataJSON, err := marshalJSON(meta)
	if err != nil {
		return nil, fmt.Errorf("create media: metadata: %w", err)
	}
	res, err := s.db.ExecContext(ctx, `
		insert into media (session_id, filename, content_type, metadata_json, original_size, compressed_size, compressed, content, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
	`, sessionID, filename, contentType, metadataJSON, len(raw), len(compressedBlob), boolToInt(compressed), compressedBlob)
	if err != nil {
		return nil, fmt.Errorf("create media: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create media: read id: %w", err)
	}
	media := &Media{
		ID:             id,
		SessionID:      sessionID,
		Filename:       filename,
		ContentType:    contentType,
		Metadata:       meta,
		OriginalSize:   len(raw),
		CompressedSize: len(compressedBlob),
		Compressed:     compressed,
	}
	return media, nil
}

func (s *Store) GetMedia(ctx context.Context, id int64) (*Media, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, session_id, filename, content_type, metadata_json, original_size, compressed_size, compressed, created_at, updated_at
		from media where id = ?
	`, id)
	var item Media
	var metadataJSON string
	var compressedInt int
	if err := row.Scan(&item.ID, &item.SessionID, &item.Filename, &item.ContentType, &metadataJSON, &item.OriginalSize, &item.CompressedSize, &compressedInt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get media: %w", err)
	}
	item.Compressed = compressedInt != 0
	meta, err := unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("get media: metadata: %w", err)
	}
	item.Metadata = meta
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return &item, nil
}

func (s *Store) GetMediaContent(ctx context.Context, id int64) (*Media, []byte, error) {
	row := s.db.QueryRowContext(ctx, `
		select id, session_id, filename, content_type, metadata_json, original_size, compressed_size, compressed, content, created_at, updated_at
		from media where id = ?
	`, id)
	var item Media
	var metadataJSON string
	var compressedInt int
	var storedBlob []byte
	if err := row.Scan(&item.ID, &item.SessionID, &item.Filename, &item.ContentType, &metadataJSON, &item.OriginalSize, &item.CompressedSize, &compressedInt, &storedBlob, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, nil, fmt.Errorf("get media content: %w", err)
	}
	item.Compressed = compressedInt != 0
	meta, err := unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("get media content: metadata: %w", err)
	}
	item.Metadata = meta
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	raw, err := storeobject.MaybeDecompressBlob(storedBlob, item.Compressed)
	if err != nil {
		return nil, nil, fmt.Errorf("get media content: decompress: %w", err)
	}
	if item.Metadata["size"] == nil {
		item.Metadata["size"] = item.OriginalSize
	}
	return &item, raw, nil
}

func (s *Store) SaveVFSFile(ctx context.Context, namespace, filePath, contentType string, raw []byte, metadata map[string]any) (*VFSFile, error) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil, fmt.Errorf("save vfs file: missing namespace")
	}
	if storevfs.IsReadOnlyNamespace(namespace) {
		return nil, fmt.Errorf("save vfs file: namespace is read-only: %s", namespace)
	}
	normalizedPath, err := storevfs.NormalizePath(filePath)
	if err != nil {
		return nil, fmt.Errorf("save vfs file: path: %w", err)
	}
	if normalizedPath == "" {
		return nil, fmt.Errorf("save vfs file: missing path")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	compressedBlob, compressed, err := storeobject.MaybeCompressBlob(raw)
	if err != nil {
		return nil, fmt.Errorf("save vfs file: compress: %w", err)
	}
	meta := copyMetadata(metadata)
	meta["size"] = len(raw)
	metadataJSON, err := marshalJSON(meta)
	if err != nil {
		return nil, fmt.Errorf("save vfs file: metadata: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		insert into vfs_files (namespace, path, content_type, metadata_json, original_size, compressed_size, compressed, content, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?, `+defaultNow+`, `+defaultNow+`)
		on conflict(namespace, path) do update set
			content_type = excluded.content_type,
			metadata_json = excluded.metadata_json,
			original_size = excluded.original_size,
			compressed_size = excluded.compressed_size,
			compressed = excluded.compressed,
			content = excluded.content,
			updated_at = excluded.updated_at
	`, namespace, normalizedPath, contentType, metadataJSON, len(raw), len(compressedBlob), boolToInt(compressed), compressedBlob)
	if err != nil {
		return nil, fmt.Errorf("save vfs file: upsert: %w", err)
	}
	return &VFSFile{
		Namespace:      namespace,
		Path:           normalizedPath,
		ContentType:    contentType,
		Metadata:       meta,
		OriginalSize:   len(raw),
		CompressedSize: len(compressedBlob),
		Compressed:     compressed,
	}, nil
}

func (s *Store) GetVFSFile(ctx context.Context, namespace, filePath string) (*VFSFile, error) {
	normalizedPath, err := storevfs.NormalizePath(filePath)
	if err != nil {
		return nil, fmt.Errorf("get vfs file: path: %w", err)
	}
	if storevfs.IsVirtualNamespace(namespace) {
		item, _, err := s.getVirtualVFSFileContent(ctx, namespace, normalizedPath)
		if err != nil {
			return nil, fmt.Errorf("get vfs file: %w", err)
		}
		return item, nil
	}
	if normalizedPath == "" {
		return nil, fmt.Errorf("get vfs file: missing path")
	}
	row := s.db.QueryRowContext(ctx, `
		select namespace, path, content_type, metadata_json, original_size, compressed_size, compressed, created_at, updated_at
		from vfs_files where namespace = ? and path = ?
	`, namespace, normalizedPath)
	var item VFSFile
	var metadataJSON string
	var compressedInt int
	if err := row.Scan(&item.Namespace, &item.Path, &item.ContentType, &metadataJSON, &item.OriginalSize, &item.CompressedSize, &compressedInt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get vfs file: %w", err)
	}
	item.Compressed = compressedInt != 0
	meta, err := unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("get vfs file: metadata: %w", err)
	}
	item.Metadata = meta
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	if item.Metadata["size"] == nil {
		item.Metadata["size"] = item.OriginalSize
	}
	return &item, nil
}

func (s *Store) GetVFSFileContent(ctx context.Context, namespace, filePath string) (*VFSFile, []byte, error) {
	normalizedPath, err := storevfs.NormalizePath(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("get vfs file content: path: %w", err)
	}
	if storevfs.IsVirtualNamespace(namespace) {
		item, raw, err := s.getVirtualVFSFileContent(ctx, namespace, normalizedPath)
		if err != nil {
			return nil, nil, fmt.Errorf("get vfs file content: %w", err)
		}
		return item, raw, nil
	}
	if normalizedPath == "" {
		return nil, nil, fmt.Errorf("get vfs file content: missing path")
	}
	row := s.db.QueryRowContext(ctx, `
		select namespace, path, content_type, metadata_json, original_size, compressed_size, compressed, content, created_at, updated_at
		from vfs_files where namespace = ? and path = ?
	`, namespace, normalizedPath)
	var item VFSFile
	var metadataJSON string
	var compressedInt int
	var storedBlob []byte
	if err := row.Scan(&item.Namespace, &item.Path, &item.ContentType, &metadataJSON, &item.OriginalSize, &item.CompressedSize, &compressedInt, &storedBlob, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, nil, fmt.Errorf("get vfs file content: %w", err)
	}
	item.Compressed = compressedInt != 0
	item.Metadata, err = unmarshalJSONMap(metadataJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("get vfs file content: metadata: %w", err)
	}
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	raw, err := storeobject.MaybeDecompressBlob(storedBlob, item.Compressed)
	if err != nil {
		return nil, nil, fmt.Errorf("get vfs file content: decompress: %w", err)
	}
	if item.Metadata["size"] == nil {
		item.Metadata["size"] = item.OriginalSize
	}
	return &item, raw, nil
}

func (s *Store) ListVFSChildren(ctx context.Context, namespace, dir string) ([]VFSListEntry, error) {
	ns := strings.TrimSpace(namespace)
	if ns == "" {
		return nil, fmt.Errorf("list vfs children: missing namespace")
	}
	normalizedDir, err := storevfs.NormalizePath(dir)
	if err != nil {
		return nil, fmt.Errorf("list vfs children: path: %w", err)
	}
	if storevfs.IsVirtualNamespace(ns) {
		return s.listVirtualVFSChildren(ctx, ns, normalizedDir)
	}
	prefix := ""
	if normalizedDir != "" {
		prefix = normalizedDir + "/"
	}
	pattern := "%"
	if prefix != "" {
		pattern = prefix + "%"
	}
	rows, err := s.db.QueryContext(ctx, `
		select path, content_type, metadata_json, original_size from vfs_files
		where namespace = ? and path like ?
		order by path asc
	`, ns, pattern)
	if err != nil {
		return nil, fmt.Errorf("list vfs children: %w", err)
	}
	defer rows.Close()

	entries := map[string]VFSListEntry{}
	for rows.Next() {
		var filePath string
		var contentType string
		var metadataJSON string
		var originalSize int
		if err := rows.Scan(&filePath, &contentType, &metadataJSON, &originalSize); err != nil {
			return nil, fmt.Errorf("list vfs children: scan: %w", err)
		}
		if filePath == normalizedDir {
			continue
		}
		relative := filePath
		if prefix != "" {
			if !strings.HasPrefix(filePath, prefix) {
				continue
			}
			relative = strings.TrimPrefix(filePath, prefix)
		} else {
			relative = strings.TrimPrefix(filePath, "")
		}
		relative = strings.TrimPrefix(relative, "/")
		if relative == "" {
			continue
		}
		parts := strings.SplitN(relative, "/", 2)
		name := parts[0]
		entry, exists := entries[name]
		if !exists {
			entry = VFSListEntry{Name: name, Path: joinVFSPath(prefix, name)}
		}
		if len(parts) > 1 {
			entry.IsDir = true
			entry.Metadata = nil
			entry.ContentType = ""
			entry.OriginalSize = 0
			entries[name] = entry
			continue
		}
		if !entry.IsDir {
			meta, err := unmarshalJSONMap(metadataJSON)
			if err != nil {
				meta = map[string]any{}
			}
			if meta == nil {
				meta = map[string]any{}
			}
			entry.ContentType = contentType
			entry.Metadata = meta
			entry.OriginalSize = originalSize
			if entry.OriginalSize == 0 {
				entry.OriginalSize = toInt(meta, "size", 0)
			}
			entries[name] = entry
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list vfs children: rows: %w", err)
	}
	list := make([]VFSListEntry, 0, len(entries))
	for _, entry := range entries {
		list = append(list, entry)
	}
	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list, nil
}

func (s *Store) DeleteVFSFile(ctx context.Context, namespace, filePath string) error {
	namespace = strings.TrimSpace(namespace)
	if storevfs.IsReadOnlyNamespace(namespace) {
		return fmt.Errorf("delete vfs file: namespace is read-only: %s", namespace)
	}
	normalizedPath, err := storevfs.NormalizePath(filePath)
	if err != nil {
		return fmt.Errorf("delete vfs file: path: %w", err)
	}
	if normalizedPath == "" {
		return fmt.Errorf("delete vfs file: missing path")
	}
	if _, err := s.db.ExecContext(ctx, `delete from vfs_files where namespace = ? and path = ?`, namespace, normalizedPath); err != nil {
		return fmt.Errorf("delete vfs file: %w", err)
	}
	return nil
}

func copyMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return map[string]any{}
	}
	copied := make(map[string]any, len(metadata))
	for key, value := range metadata {
		if strings.TrimSpace(key) == "" {
			continue
		}
		copied[key] = value
	}
	return copied
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func joinVFSPath(prefix, leaf string) string {
	if prefix == "" {
		return leaf
	}
	if prefix[len(prefix)-1] == '/' {
		return prefix + leaf
	}
	return prefix + "/" + leaf
}

func toInt(metadata map[string]any, key string, fallback int) int {
	if metadata == nil {
		return fallback
	}
	raw, ok := metadata[key]
	if !ok {
		return fallback
	}
	switch value := raw.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case float32:
		return int(value)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fallback
		}
		return parsed
	default:
		return fallback
	}
}
