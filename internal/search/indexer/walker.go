package indexer

import (
	"io/fs"
	"path/filepath"
)

// WalkFiles returns candidate workspace file paths under root.
func WalkFiles(root string, include func(path string, d fs.DirEntry) bool) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if include == nil || include(path, d) {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
