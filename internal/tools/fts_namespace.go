package tools

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rcarmo/gi/internal/store"
)

type ftsMessageHit struct {
	ID        string
	SessionID string
	Role      string
	Content   string
	CreatedAt string
}

type ftsTurnHit struct {
	ID        string
	SessionID string
	Status    string
	Phase     string
	Prompt    string
	CreatedAt string
	UpdatedAt string
}

type ftsWorkspaceHit struct {
	Path    string
	Excerpt string
}

// ReadFTSQuery resolves a read-only fts:// locator into a markdown document.
// Expected locator forms include:
//
//	messages?q=...&limit=20
//	turns?q=...&limit=20
//	workspace?q=...&limit=20&glob=*.go
//	all?q=...&limit=20
func ReadFTSQuery(ctx context.Context, workspaceRoot string, s *store.Store, locator string) (string, error) {
	locator = strings.TrimLeft(strings.TrimSpace(locator), "/")
	if locator == "" {
		locator = "help"
	}
	u, err := url.Parse("fts://" + locator)
	if err != nil {
		return "", fmt.Errorf("fts parse: %w", err)
	}
	target := strings.Trim(strings.TrimSpace(u.Host+u.Path), "/")
	if target == "" {
		target = "help"
	}
	params := u.Query()
	query := strings.TrimSpace(firstNonEmpty(params.Get("q"), params.Get("query")))
	limit := parsePositiveInt(params.Get("limit"), 20, 1, 100)
	sessionID := strings.TrimSpace(params.Get("session"))
	glob := strings.TrimSpace(params.Get("glob"))

	switch target {
	case "help", "", "index":
		return strings.TrimSpace(`# FTS namespace

Use `+"`read`"+` with `+"`fts://...`"+` locators.

Examples:

- `+"`fts://messages?q=queue+overflow&limit=20`"+`
- `+"`fts://messages?q=steering&session=session_1`"+`
- `+"`fts://turns?q=compaction&limit=20`"+`
- `+"`fts://workspace?q=HookResponse&glob=internal/**/*.go&limit=20`"+`
- `+"`fts://all?q=subturn&limit=20`"+`

Results include pointers back into `+"`vfs://chat/...`"+` for raw chat artifacts.
`) + "\n", nil
	case "messages":
		if query == "" {
			return "", fmt.Errorf("fts://messages requires q=... or query=...")
		}
		hits, err := searchMessages(ctx, s, query, sessionID, limit)
		if err != nil {
			return "", err
		}
		var b strings.Builder
		writeSimpleFrontmatter(&b, map[string]any{"kind": "fts/messages", "query": query, "session": sessionID, "limit": limit, "count": len(hits)})
		b.WriteString("# Message search\n\n")
		if len(hits) == 0 {
			b.WriteString("No results.\n")
			return b.String(), nil
		}
		for _, hit := range hits {
			b.WriteString(fmt.Sprintf("- `%s` [%s] `%s`\n", hit.ID, hit.Role, hit.CreatedAt))
			b.WriteString(fmt.Sprintf("  - session: `%s`\n", hit.SessionID))
			b.WriteString(fmt.Sprintf("  - source: `vfs://chat/sessions/%s/messages/%s.md`\n", hit.SessionID, hit.ID))
			b.WriteString(fmt.Sprintf("  - excerpt: %q\n", excerptText(hit.Content, query, 180)))
		}
		return b.String(), nil
	case "turns":
		if query == "" {
			return "", fmt.Errorf("fts://turns requires q=... or query=...")
		}
		hits, err := searchTurns(ctx, s, query, sessionID, limit)
		if err != nil {
			return "", err
		}
		var b strings.Builder
		writeSimpleFrontmatter(&b, map[string]any{"kind": "fts/turns", "query": query, "session": sessionID, "limit": limit, "count": len(hits)})
		b.WriteString("# Turn search\n\n")
		if len(hits) == 0 {
			b.WriteString("No results.\n")
			return b.String(), nil
		}
		for _, hit := range hits {
			b.WriteString(fmt.Sprintf("- `%s` [%s/%s] `%s`\n", hit.ID, hit.Status, hit.Phase, hit.CreatedAt))
			b.WriteString(fmt.Sprintf("  - session: `%s`\n", hit.SessionID))
			b.WriteString(fmt.Sprintf("  - source: `vfs://chat/sessions/%s/turns/%s.md`\n", hit.SessionID, hit.ID))
			b.WriteString(fmt.Sprintf("  - excerpt: %q\n", excerptText(hit.Prompt, query, 180)))
		}
		return b.String(), nil
	case "workspace":
		if query == "" {
			return "", fmt.Errorf("fts://workspace requires q=... or query=...")
		}
		hits, err := searchWorkspaceFiles(workspaceRoot, query, glob, limit)
		if err != nil {
			return "", err
		}
		var b strings.Builder
		writeSimpleFrontmatter(&b, map[string]any{"kind": "fts/workspace", "query": query, "glob": glob, "limit": limit, "count": len(hits)})
		b.WriteString("# Workspace search\n\n")
		if len(hits) == 0 {
			b.WriteString("No results.\n")
			return b.String(), nil
		}
		for _, hit := range hits {
			b.WriteString(fmt.Sprintf("- `%s`\n", hit.Path))
			b.WriteString(fmt.Sprintf("  - excerpt: %q\n", hit.Excerpt))
		}
		return b.String(), nil
	case "all":
		if query == "" {
			return "", fmt.Errorf("fts://all requires q=... or query=...")
		}
		msgHits, err := searchMessages(ctx, s, query, sessionID, limit)
		if err != nil {
			return "", err
		}
		turnHits, err := searchTurns(ctx, s, query, sessionID, limit)
		if err != nil {
			return "", err
		}
		wsHits, err := searchWorkspaceFiles(workspaceRoot, query, glob, limit)
		if err != nil {
			return "", err
		}
		var b strings.Builder
		writeSimpleFrontmatter(&b, map[string]any{"kind": "fts/all", "query": query, "session": sessionID, "glob": glob, "limit": limit, "messages": len(msgHits), "turns": len(turnHits), "workspace": len(wsHits)})
		b.WriteString("# Unified search\n\n")
		b.WriteString("## Messages\n\n")
		if len(msgHits) == 0 {
			b.WriteString("- none\n")
		} else {
			for _, hit := range msgHits {
				b.WriteString(fmt.Sprintf("- `%s` in `%s` → `vfs://chat/sessions/%s/messages/%s.md`\n", hit.ID, hit.SessionID, hit.SessionID, hit.ID))
			}
		}
		b.WriteString("\n## Turns\n\n")
		if len(turnHits) == 0 {
			b.WriteString("- none\n")
		} else {
			for _, hit := range turnHits {
				b.WriteString(fmt.Sprintf("- `%s` in `%s` → `vfs://chat/sessions/%s/turns/%s.md`\n", hit.ID, hit.SessionID, hit.SessionID, hit.ID))
			}
		}
		b.WriteString("\n## Workspace\n\n")
		if len(wsHits) == 0 {
			b.WriteString("- none\n")
		} else {
			for _, hit := range wsHits {
				b.WriteString(fmt.Sprintf("- `%s`\n", hit.Path))
			}
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("unknown fts target: %s", target)
	}
}

func searchMessages(ctx context.Context, s *store.Store, query, sessionID string, limit int) ([]ftsMessageHit, error) {
	like := "%" + strings.ToLower(query) + "%"
	q := `
		select id, session_id, role, content, created_at
		from messages
		where lower(content) like ?
	`
	args := []any{like}
	if sessionID != "" {
		q += ` and session_id = ?`
		args = append(args, sessionID)
	}
	q += ` order by created_at desc, id desc limit ?`
	args = append(args, limit)
	rows, err := s.DB().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("fts messages query: %w", err)
	}
	defer rows.Close()
	out := []ftsMessageHit{}
	for rows.Next() {
		var item ftsMessageHit
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Role, &item.Content, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func searchTurns(ctx context.Context, s *store.Store, query, sessionID string, limit int) ([]ftsTurnHit, error) {
	like := "%" + strings.ToLower(query) + "%"
	q := `
		select id, session_id, status, phase, prompt, created_at, updated_at
		from turns
		where (lower(prompt) like ? or lower(metadata_json) like ?)
	`
	args := []any{like, like}
	if sessionID != "" {
		q += ` and session_id = ?`
		args = append(args, sessionID)
	}
	q += ` order by created_at desc, id desc limit ?`
	args = append(args, limit)
	rows, err := s.DB().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("fts turns query: %w", err)
	}
	defer rows.Close()
	out := []ftsTurnHit{}
	for rows.Next() {
		var item ftsTurnHit
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Status, &item.Phase, &item.Prompt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func searchWorkspaceFiles(workspaceRoot, query, glob string, limit int) ([]ftsWorkspaceHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	qLower := strings.ToLower(query)
	root := filepath.Clean(workspaceRoot)
	if root == "" {
		root = "."
	}
	out := []ftsWorkspaceHit{}
	stop := fmt.Errorf("stop walk")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			switch name {
			case ".git", "node_modules", ".cache", "artifacts", "tmp", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		if len(out) >= limit {
			return stop
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if glob != "" {
			matched, mErr := filepath.Match(glob, rel)
			if mErr == nil && !matched {
				return nil
			}
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 1024*1024 {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if !looksLikeText(raw) {
			return nil
		}
		content := string(raw)
		idx := strings.Index(strings.ToLower(content), qLower)
		if idx < 0 {
			return nil
		}
		out = append(out, ftsWorkspaceHit{Path: rel, Excerpt: excerptAt(content, idx, len(query), 200)})
		return nil
	})
	if err != nil && err != stop {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Path) < strings.ToLower(out[j].Path) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parsePositiveInt(raw string, fallback, min, max int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func writeSimpleFrontmatter(b *strings.Builder, meta map[string]any) {
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString("---\n")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(fmt.Sprintf("%q", fmt.Sprint(meta[k])))
		b.WriteString("\n")
	}
	b.WriteString("---\n\n")
}

func excerptText(content, query string, maxLen int) string {
	idx := strings.Index(strings.ToLower(content), strings.ToLower(strings.TrimSpace(query)))
	if idx < 0 {
		if len(content) <= maxLen {
			return content
		}
		return content[:maxLen] + "..."
	}
	return excerptAt(content, idx, len(query), maxLen)
}

func excerptAt(content string, idx, queryLen, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 120
	}
	start := idx - maxLen/2
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(content) {
		end = len(content)
	}
	snippet := content[start:end]
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	snippet = strings.TrimSpace(snippet)
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet += "..."
	}
	return snippet
}

func looksLikeText(raw []byte) bool {
	if len(raw) == 0 {
		return true
	}
	sample := raw
	if len(sample) > 4096 {
		sample = sample[:4096]
	}
	for _, b := range sample {
		if b == 0 {
			return false
		}
	}
	return true
}
