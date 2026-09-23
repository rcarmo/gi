package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/config"
	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
	"github.com/rcarmo/gi/internal/store"
)

// WriteFile shares filesystem/VFS semantics across the native tool, HTTP and
// script bridge. Filesystem notifications are durable but never schedule work.
func WriteFile(ctx context.Context, cfg config.RuntimeConfig, s *store.Store, path, content string) error {
	resolved, err := ResolveToolPath(cfg.WorkspaceRoot, path, true)
	if err != nil {
		return err
	}
	if resolved.IsVFS() {
		_, err = s.SaveVFSFile(ctx, resolved.VFSNamespace, resolved.VFSPath, inferContentTypeFromFilename(resolved.VFSPath), []byte(content), map[string]any{})
		return err
	}
	configs := make([]searchstore.ScopeConfig, 0, 3)
	for _, scope := range []string{"all", "notes", "skills"} {
		c, err := searchstore.ConfiguredScopeConfig(cfg.WorkspaceRoot, scope, cfg.WorkspaceIndex.ExtraRoots, cfg.WorkspaceIndex.ExtraExtensions, cfg.WorkspaceIndex.OptionalRoots, chunking.LineVersion)
		if err != nil {
			return fmt.Errorf("write not attempted: index configuration: %w", err)
		}
		configs = append(configs, c)
	}
	absolute, err := filepath.Abs(cfg.WorkspaceRoot)
	if err != nil {
		return err
	}
	target, err := filepath.Abs(resolved.WorkspacePath)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(absolute, target)
	if err != nil || !filepath.IsLocal(relative) {
		return fmt.Errorf("write path escapes workspace")
	}
	root, err := os.OpenRoot(configs[0].Workspace())
	if err != nil {
		return err
	}
	defer root.Close()
	// Alias writes cannot be attributed to the lexical path scanned by the index.
	// Reject symlink components; os.Root also prevents escape during path races.
	current := ""
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := root.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("write through symlink is not supported: %s", current)
		}
		if current == relative && !info.Mode().IsRegular() {
			return fmt.Errorf("write requires a regular file: %s", current)
		}
	}
	storage := searchstore.NewRefreshStore(s.DB())
	notify := func(ctx context.Context) error {
		op, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := storage.InvalidateScopes(op, configs, []string{filepath.ToSlash(relative)})
		return err
	}
	return writeWithInvalidation(ctx, notify, func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := root.MkdirAll(filepath.Dir(relative), 0755); err != nil {
			return err
		}
		file, err := openWorkspaceWrite(root, relative)
		if err != nil {
			return err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("write requires a regular file: %s", relative)
		}
		if err = file.Truncate(0); err != nil {
			return err
		}
		_, err = file.WriteString(content)
		return errors.Join(err, file.Close())
	})
}

// The filesystem and SQLite cannot share a transaction. A pre-notification
// failure prevents mutation. Always post-notify after an attempted write,
// including partial failure, with bounded cleanup independent of caller cancel.
func writeWithInvalidation(ctx context.Context, notify func(context.Context) error, write func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := notify(ctx); err != nil {
		return fmt.Errorf("write not attempted: index invalidation: %w", err)
	}
	writeErr := write()
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := notify(cleanup); err != nil {
		return errors.Join(writeErr, fmt.Errorf("file write attempted; bytes may have changed, but index invalidation failed; explicit reindex required: %w", err))
	}
	return writeErr
}
