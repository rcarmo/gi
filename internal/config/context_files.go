package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rcarmo/gi/internal/agentdir"
)

// Project context files, as Pi's loadProjectContextFiles
// (core/resource-loader.js): the agent directory's file (gi's, else Pi's),
// then one file per directory from the filesystem root down to the
// workspace. A linked git worktree's own file shadows the main repository's.

// ContextFile is one instructions file in the system prompt.
type ContextFile struct {
	Path    string
	Content string
}

var contextFileNames = []string{"AGENTS.override.md", "AGENTS.md", "AGENTS.MD", "CLAUDE.md", "CLAUDE.MD"}

// loadContextFileFromDir is the first context file in dir.
func loadContextFileFromDir(dir string) (ContextFile, bool) {
	for _, name := range contextFileNames {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return ContextFile{Path: path, Content: strings.TrimPrefix(string(data), "\uFEFF")}, true
	}
	return ContextFile{}, false
}

func canonicalize(path string) string {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		return real
	}
	return path
}

// findGitPaths is Pi's: the nearest .git (directory, or a "gitdir:" file of
// a linked worktree) with its common git directory.
func findGitPaths(cwd string) (repoDir, commonGitDir string, ok bool) {
	for dir := cwd; ; {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.Mode().IsRegular() {
				data, err := os.ReadFile(gitPath)
				content := strings.TrimSpace(string(data))
				if err != nil || !strings.HasPrefix(content, "gitdir: ") {
					return "", "", false
				}
				gitDir := strings.TrimSpace(content[len("gitdir: "):])
				if !filepath.IsAbs(gitDir) {
					gitDir = filepath.Join(dir, gitDir)
				}
				if _, err := os.Stat(filepath.Join(gitDir, "HEAD")); err != nil {
					return "", "", false
				}
				common := gitDir
				if raw, err := os.ReadFile(filepath.Join(gitDir, "commondir")); err == nil {
					common = strings.TrimSpace(string(raw))
					if !filepath.IsAbs(common) {
						common = filepath.Join(gitDir, common)
					}
				}
				return dir, filepath.Clean(common), true
			}
			if info.IsDir() {
				if _, err := os.Stat(filepath.Join(gitPath, "HEAD")); err != nil {
					return "", "", false
				}
				return dir, gitPath, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", false
		}
		dir = parent
	}
}

// findShadowedContextFile is Pi's: in a linked worktree nested under its main
// repository, the main repository's context file of the same name.
func findShadowedContextFile(cwd string) string {
	repo, common, ok := findGitPaths(cwd)
	if !ok {
		return ""
	}
	commonGitDir, worktreeRoot := canonicalize(common), canonicalize(repo)
	mainRepoRoot := filepath.Dir(commonGitDir)
	if !strings.HasPrefix(worktreeRoot, mainRepoRoot+string(filepath.Separator)) {
		return ""
	}
	if canonicalize(filepath.Join(mainRepoRoot, ".git")) != commonGitDir {
		return ""
	}
	if f, ok := loadContextFileFromDir(worktreeRoot); ok {
		return filepath.Join(mainRepoRoot, filepath.Base(f.Path))
	}
	return ""
}

// LoadContextFiles is Pi's loadProjectContextFiles for a workspace.
func LoadContextFiles(workspace string) []ContextFile {
	cwd, err := filepath.Abs(workspace)
	if err != nil {
		return nil
	}
	var out []ContextFile
	seen := map[string]bool{}
	for _, dir := range agentdir.Dirs() {
		if f, ok := loadContextFileFromDir(dir); ok {
			out = append(out, f)
			seen[canonicalize(f.Path)] = true
			break
		}
	}
	shadowed := findShadowedContextFile(cwd)
	var ancestors []ContextFile
	for dir := cwd; ; {
		var local []ContextFile
		// AgentsInTheCloud applies secondary instructions after the directory's normal context.
		for _, contextDir := range []string{dir, filepath.Join(dir, ".agents-in-the-cloud")} {
			if f, ok := loadContextFileFromDir(contextDir); ok && !(shadowed != "" && canonicalize(f.Path) == shadowed) && !seen[canonicalize(f.Path)] {
				local = append(local, f)
				seen[canonicalize(f.Path)] = true
			}
		}
		ancestors = append(local, ancestors...)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return append(out, ancestors...)
}
