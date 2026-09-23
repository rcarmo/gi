package store

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

const MaxSnapshotFiles = 10000
const MaxSnapshotText = 32 << 20
const MaxDocumentBytes = 1 << 20
const MaxSnapshotChunks = 20000

// ScopeConfig is constructed and normalised once, then carried by a refresh
// handle. Callers cannot mutate its roots/fingerprint while a scan is running.
type ScopeConfig struct {
	workspace, scope, fingerprint, chunker string
	roots, extensions                      []string
	optionalRoots                          []string
}

func NewScopeConfig(workspace, scope string, roots, extensions []string, chunker string) (ScopeConfig, error) {
	c := ScopeConfig{scope: scope, chunker: chunker}
	if workspace == "" || scope == "" || len(scope) > 128 || strings.TrimSpace(scope) != scope || chunker == "" || len(chunker) > 128 || !utf8.ValidString(scope+chunker) || strings.ContainsRune(scope+chunker, 0) {
		return c, fmt.Errorf("invalid scope or chunker")
	}
	absolute, err := filepath.Abs(workspace)
	if err != nil {
		return c, err
	}
	c.workspace, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return c, err
	}
	info, err := os.Stat(c.workspace)
	if err != nil || !info.IsDir() {
		return c, fmt.Errorf("workspace must be an existing directory")
	}
	if len(roots) > 128 || len(extensions) > 128 {
		return c, fmt.Errorf("scope configuration limit exceeded")
	}
	for _, root := range roots {
		if !localIndexPath(root) {
			return c, fmt.Errorf("invalid scope root %q", root)
		}
		c.roots = append(c.roots, path.Clean(root))
	}
	slices.Sort(c.roots)
	c.roots = slices.Compact(c.roots)
	var resolved []string
	for _, root := range c.roots {
		covered := false
		for _, parent := range resolved {
			if underIndexRoot(root, parent) {
				covered = true
				break
			}
		}
		if !covered {
			resolved = append(resolved, root)
		}
	}
	c.roots = resolved
	if len(c.roots) == 0 {
		return c, fmt.Errorf("scope requires roots")
	}
	for _, ext := range extensions {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if len(ext) < 2 || len(ext) > 64 || !utf8.ValidString(ext) || strings.ContainsAny(ext, "/\\\x00") {
			return c, fmt.Errorf("invalid extension")
		}
		c.extensions = append(c.extensions, ext)
	}
	slices.Sort(c.extensions)
	c.extensions = slices.Compact(c.extensions)
	if len(c.extensions) == 0 {
		return c, fmt.Errorf("scope requires supported extensions")
	}
	c.setFingerprint()
	return c, nil
}

// DefaultScopeConfig follows Piclaw's notes/skills/all ownership. Extra roots
// augment all, never widen notes/skills. Scanning/exclusion policy is separate.
func DefaultScopeConfig(workspace, scope string, extraRoots, extraExtensions []string, chunker string) (ScopeConfig, error) {
	roots := []string{"notes", ".pi/skills"}
	switch scope {
	case "notes":
		roots = []string{"notes"}
	case "skills":
		roots = []string{".pi/skills"}
	case "all":
		roots = append(roots, extraRoots...)
	default:
		return ScopeConfig{}, fmt.Errorf("unknown default scope %q", scope)
	}
	extensions := []string{".md", ".txt", ".ts", ".tsx", ".js", ".jsx", ".json", ".yaml", ".yml", ".sh", ".csv", ".xml", ".toml", ".env", ".py", ".go", ".rs"}
	return NewScopeConfig(workspace, scope, roots, append(extensions, extraExtensions...), chunker)
}

// ConfiguredScopeConfig validates global optional roots against the all-scope
// configuration, then applies only roots relevant to the requested scope. A
// subroot cannot be optional under a scanned required parent (ambiguous cleanup).
func ConfiguredScopeConfig(workspace, scope string, extraRoots, extraExtensions, optionalRoots []string, chunker string) (ScopeConfig, error) {
	all, err := DefaultScopeConfig(workspace, "all", extraRoots, extraExtensions, chunker)
	if err != nil {
		return ScopeConfig{}, err
	}
	if len(optionalRoots) > 128 {
		return ScopeConfig{}, fmt.Errorf("optional root limit exceeded")
	}
	for _, root := range optionalRoots {
		if !localIndexPath(root) || root == "." || !slices.Contains(all.roots, root) {
			return ScopeConfig{}, fmt.Errorf("optional root %q must be a distinct configured root", root)
		}
	}
	c, err := DefaultScopeConfig(workspace, scope, extraRoots, extraExtensions, chunker)
	if err != nil {
		return c, err
	}
	for _, root := range optionalRoots {
		if slices.Contains(c.roots, root) {
			c.optionalRoots = append(c.optionalRoots, root)
		}
	}
	slices.Sort(c.optionalRoots)
	c.optionalRoots = slices.Compact(c.optionalRoots)
	c.setFingerprint()
	return c, nil
}
func (c *ScopeConfig) setFingerprint() {
	raw, _ := json.Marshal(struct {
		Roots, Extensions []string
		Chunker           string
		MaxBytes          int
		OptionalRoots     []string `json:",omitempty"`
	}{c.roots, c.extensions, c.chunker, MaxDocumentBytes, c.optionalRoots})
	c.fingerprint = fmt.Sprintf("%x", sha256.Sum256(raw))
}
func (c ScopeConfig) OptionalRoots() []string         { return slices.Clone(c.optionalRoots) }
func (c ScopeConfig) IsOptionalRoot(root string) bool { return slices.Contains(c.optionalRoots, root) }

func localIndexPath(p string) bool {
	return p != "" && len(p) <= 4096 && utf8.ValidString(p) && !strings.ContainsAny(p, "\\\x00") && !strings.Contains(p, ":") && !path.IsAbs(p) && path.Clean(p) != ".." && !strings.HasPrefix(path.Clean(p), "../") && path.Clean(p) == p
}
func underIndexRoot(p, root string) bool {
	return root == "." || p == root || strings.HasPrefix(p, root+"/")
}
func (c ScopeConfig) eligible(p string) bool {
	if !localIndexPath(p) || !slices.Contains(c.extensions, strings.ToLower(path.Ext(p))) {
		return false
	}
	for _, root := range c.roots {
		if underIndexRoot(p, root) {
			return true
		}
	}
	return false
}
func (c ScopeConfig) Roots() []string        { return slices.Clone(c.roots) }
func (c ScopeConfig) Fingerprint() string    { return c.fingerprint }
func (c ScopeConfig) Workspace() string      { return c.workspace }
func (c ScopeConfig) Scope() string          { return c.scope }
func (c ScopeConfig) ChunkerVersion() string { return c.chunker }

// Eligible reports configured path/type membership, not filesystem permission.
func (c ScopeConfig) Eligible(p string) bool { return c.eligible(p) }
