// Package docs exposes Gi's shipped documentation assets.
package docs

import (
	"embed"
	"io/fs"
)

// internalReference contains the internal reference documents and topic trees.
//
//go:embed internal/*.md internal/hooks/*.md internal/scripting/*.md internal/search/*.md internal/skills/*.md internal/tools/*.md internal/vfs/*.md
var internalReference embed.FS

// InternalReferenceFS returns an FS rooted at docs/internal.
func InternalReferenceFS() fs.FS {
	sub, err := fs.Sub(internalReference, "internal")
	if err != nil {
		panic(err)
	}
	return sub
}
