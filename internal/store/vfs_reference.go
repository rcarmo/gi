package store

import (
	"database/sql"
	"io/fs"
	"mime"
	"path"
	"sort"
	"strings"

	referencedocs "github.com/rcarmo/gi/docs"
	storevfs "github.com/rcarmo/gi/internal/store/vfs"
)

func (s *Store) getReferenceVFSFileContent(filePath string) (*VFSFile, []byte, error) {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		filePath = "README.md"
	}
	info, err := fs.Stat(referencedocs.InternalReferenceFS(), filePath)
	if err != nil || info.IsDir() {
		return nil, nil, sql.ErrNoRows
	}
	raw, err := fs.ReadFile(referencedocs.InternalReferenceFS(), filePath)
	if err != nil {
		return nil, nil, err
	}
	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	item := &VFSFile{
		Namespace:    storevfs.NamespaceReference,
		Path:         filePath,
		ContentType:  contentType,
		Metadata:     map[string]any{"virtual": true, "kind": "reference/document", "source": "docs/internal", "size": len(raw)},
		OriginalSize: len(raw),
		Compressed:   false,
	}
	return item, raw, nil
}

func (s *Store) listReferenceVFSChildren(dir string) ([]VFSListEntry, error) {
	dir = strings.TrimSpace(dir)
	fsDir := dir
	if fsDir == "" {
		fsDir = "."
	}
	children, err := fs.ReadDir(referencedocs.InternalReferenceFS(), fsDir)
	if err != nil {
		return nil, err
	}
	entries := make([]VFSListEntry, 0, len(children))
	for _, child := range children {
		entryPath := child.Name()
		if dir != "" {
			entryPath = path.Join(dir, child.Name())
		}
		entry := VFSListEntry{Name: child.Name(), Path: entryPath, IsDir: child.IsDir(), Metadata: map[string]any{"virtual": true, "source": "docs/internal"}}
		if !child.IsDir() {
			info, infoErr := child.Info()
			if infoErr != nil {
				return nil, infoErr
			}
			entry.OriginalSize = int(info.Size())
			entry.ContentType = mime.TypeByExtension(path.Ext(child.Name()))
			if entry.ContentType == "" {
				entry.ContentType = "application/octet-stream"
			}
			entry.Metadata["kind"] = "reference/document"
			entry.Metadata["size"] = entry.OriginalSize
		} else {
			entry.Metadata["kind"] = "reference/directory"
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name) })
	return entries, nil
}
