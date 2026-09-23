package indexer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/rcarmo/gi/internal/search/chunking"
	searchstore "github.com/rcarmo/gi/internal/search/store"
)

var ErrScanChanged = errors.New("workspace changed during index scan")
var ErrScanLimit = errors.New("workspace index scan limit exceeded")

// ScanScope requires roots to exist unless explicitly configured as optional.
// Missing optional roots are rechecked and reported, never silently treated as
// deletions: scoped publication checks their previous memberships transactionally.
func ScanScope(ctx context.Context, config searchstore.ScopeConfig) (searchstore.CompleteSnapshot, error) {
	return scanScope(ctx, config, defaultScanLimits(), nil)
}

type scanLimits struct{ entries, depth, totalText, files, chunks int }

func defaultScanLimits() scanLimits {
	return scanLimits{10000, 64, searchstore.MaxSnapshotText, searchstore.MaxSnapshotFiles, searchstore.MaxSnapshotChunks}
}

type observation struct {
	path string
	info os.FileInfo
	hash *[32]byte
}

// afterInventory is an unexported test seam for deterministic external changes.
func scanScope(ctx context.Context, config searchstore.ScopeConfig, limits scanLimits, afterInventory func()) (searchstore.CompleteSnapshot, error) {
	fail := func(err error) (searchstore.CompleteSnapshot, error) { return searchstore.CompleteSnapshot{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if config.Fingerprint() == "" || config.ChunkerVersion() != chunking.LineVersion {
		return fail(fmt.Errorf("scanner requires %s configuration", chunking.LineVersion))
	}
	root, err := os.OpenRoot(config.Workspace())
	if err != nil {
		return fail(err)
	}
	defer root.Close()
	var docs []searchstore.RefreshDocument
	var missingRoots []string
	var observed []observation
	entries, readBytes, chunkCount := 0, 0, 0
	check := func() error { return ctx.Err() }
	observe := func(name string, info os.FileInfo, hash *[32]byte) {
		observed = append(observed, observation{name, info, hash})
	}
	var walk func(string, int) error
	walk = func(name string, depth int) error {
		if err := check(); err != nil {
			return err
		}
		if depth > limits.depth {
			return fmt.Errorf("%w: depth", ErrScanLimit)
		}
		before, err := root.Lstat(name)
		if err != nil {
			return err
		}
		if !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("scope directory unavailable: %s", name)
		}
		dir, err := openScanEntry(root, name)
		if err != nil {
			return err
		}
		actual, err := dir.Stat()
		if err != nil {
			dir.Close()
			return err
		}
		if !sameObserved(before, actual) {
			dir.Close()
			return ErrScanChanged
		}
		children, readErr := dir.ReadDir(limits.entries - entries + 1)
		dir.Close()
		if readErr != nil && readErr != io.EOF {
			return readErr
		}
		entries += len(children)
		if entries > limits.entries {
			return fmt.Errorf("%w: entries", ErrScanLimit)
		}
		observe(name, before, nil)
		sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })
		for _, child := range children {
			if err := check(); err != nil {
				return err
			}
			childPath := path.Join(name, child.Name())
			if len(childPath) > 4096 || !utf8.ValidString(childPath) {
				return fmt.Errorf("unsupported index path")
			}
			info, err := root.Lstat(childPath)
			if err != nil {
				return err
			}
			// Ignore special entries and symlinks without opening them. .pi is not
			// excluded: Piclaw's skills scope lives beneath it.
			if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
				observe(childPath, info, nil)
				continue
			}
			if info.IsDir() {
				if excludedIndexDir(child.Name()) {
					observe(childPath, info, nil)
					continue
				}
				if err := walk(childPath, depth+1); err != nil {
					return err
				}
				continue
			}
			if !config.Eligible(childPath) || info.Size() > searchstore.MaxDocumentBytes {
				observe(childPath, info, nil)
				continue
			}
			// Count all candidate bytes, including binary content later excluded.
			// Final verification rereads at most this same bounded byte set.
			if info.Size() > int64(limits.totalText-readBytes) {
				return fmt.Errorf("%w: candidate bytes", ErrScanLimit)
			}
			raw, err := readScanFile(ctx, root, childPath, info)
			if err != nil {
				return err
			}
			readBytes += len(raw)
			hash := sha256.Sum256(raw)
			observe(childPath, info, &hash)
			if !utf8.Valid(raw) || bytes.IndexByte(raw, 0) >= 0 {
				continue
			} // explicit binary policy
			if len(docs) >= limits.files {
				return fmt.Errorf("%w: text/files", ErrScanLimit)
			}
			chunks, err := (chunking.Lines{}).Chunk(childPath, raw)
			if err != nil {
				return err
			}
			chunkCount += len(chunks)
			if chunkCount > limits.chunks {
				return fmt.Errorf("%w: chunks", ErrScanLimit)
			}
			docs = append(docs, searchstore.RefreshDocument{Path: childPath, Content: string(raw), MtimeNS: info.ModTime().UnixNano(), Chunks: chunks})
		}
		return nil
	}
	// Inspect every ancestor of an explicitly selected root, so an in-workspace
	// symlink cannot silently alias another configured subtree.
	ancestors := map[string]bool{}
roots:
	for _, scopeRoot := range config.Roots() {
		parent := "."
		if scopeRoot != "." {
			for _, segment := range strings.Split(scopeRoot, "/") {
				parent = path.Join(parent, segment)
				if ancestors[parent] {
					continue
				}
				info, err := root.Lstat(parent)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) && config.IsOptionalRoot(scopeRoot) {
						missingRoots = append(missingRoots, scopeRoot)
						continue roots
					}
					return fail(err)
				}
				if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || excludedIndexDir(segment) {
					return fail(fmt.Errorf("invalid index root ancestor: %s", parent))
				}
				ancestors[parent] = true
				observe(parent, info, nil)
			}
		}
		if err := walk(scopeRoot, 0); err != nil {
			return fail(err)
		}
	}
	if afterInventory != nil {
		afterInventory()
	}
	// Verify identity/size/mtime/mode for all visited entries, then rehash eligible
	// reads. This catches same-size edits with restored timestamps and file swaps.
	// It is change detection, not a filesystem-wide atomic snapshot: a later edit
	// still requires invalidation/retry by the future worker.
	for _, item := range observed {
		if err := check(); err != nil {
			return fail(err)
		}
		now, err := root.Lstat(item.path)
		if err != nil {
			return fail(fmt.Errorf("%w: %s", ErrScanChanged, item.path))
		}
		if !sameObserved(item.info, now) {
			return fail(fmt.Errorf("%w: %s", ErrScanChanged, item.path))
		}
		if item.hash != nil {
			raw, err := readScanFile(ctx, root, item.path, now)
			if err != nil {
				return fail(err)
			}
			if hash := sha256.Sum256(raw); hash != *item.hash {
				return fail(fmt.Errorf("%w: %s", ErrScanChanged, item.path))
			}
		}
	}
	for _, missing := range missingRoots {
		absent, err := absentScopeRoot(root, missing)
		if err != nil {
			return fail(err)
		}
		if !absent {
			return fail(fmt.Errorf("%w: optional root appeared: %s", ErrScanChanged, missing))
		}
	}
	// Check the workspace name still resolves to the opened directory.
	current, err := os.OpenRoot(config.Workspace())
	if err != nil {
		return fail(ErrScanChanged)
	}
	defer current.Close()
	original, err := root.Stat(".")
	if err != nil {
		return fail(err)
	}
	now, err := current.Stat(".")
	if err != nil || !os.SameFile(original, now) {
		return fail(ErrScanChanged)
	}
	if err := check(); err != nil {
		return fail(err)
	}
	return searchstore.CompleteSnapshot{Complete: true, Documents: docs, MissingRoots: missingRoots}, nil
}

func absentScopeRoot(root *os.Root, name string) (bool, error) {
	parent := "."
	for _, part := range strings.Split(name, "/") {
		parent = path.Join(parent, part)
		info, err := root.Lstat(parent)
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || excludedIndexDir(part) {
			return false, fmt.Errorf("invalid optional index root ancestor: %s", parent)
		}
	}
	return false, nil
}

func excludedIndexDir(name string) bool {
	switch name {
	case ".git", "node_modules", ".cache", "generated":
		return true
	}
	return false
}
func sameObserved(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Mode() == b.Mode() && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

func readScanFile(ctx context.Context, root *os.Root, name string, expected os.FileInfo) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := openScanEntry(root, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || !sameObserved(expected, info) {
		return nil, ErrScanChanged
	}
	var result bytes.Buffer
	buffer := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, err := file.Read(buffer)
		if n > 0 {
			if result.Len()+n > searchstore.MaxDocumentBytes {
				return nil, ErrScanChanged
			}
			result.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	after, err := file.Stat()
	if err != nil {
		return nil, err
	}
	latest, err := root.Lstat(name)
	if err != nil || !sameObserved(expected, after) || !sameObserved(after, latest) || int64(result.Len()) != after.Size() {
		return nil, ErrScanChanged
	}
	return result.Bytes(), nil
}
