package web

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"strconv"
)

// An opaque revision identifies the complete bytes plus metadata read for edit.
// It is a precondition, not a generated-output equivalence check.
func workspaceRevision(info fs.FileInfo, content []byte) string {
	h := sha256.New()
	h.Write([]byte(strconv.FormatInt(info.ModTime().UnixNano(), 10)))
	h.Write([]byte{0})
	h.Write([]byte(strconv.FormatUint(uint64(info.Mode()), 10)))
	h.Write([]byte{0})
	h.Write(content)
	return "file-v1-" + hex.EncodeToString(h.Sum(nil))
}
