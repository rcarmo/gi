package rtk

import (
	"bufio"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Result struct {
	Command string `json:"command"`
	Mode    string `json:"mode"`
	RawLen  int    `json:"raw_len"`
	OutLen  int    `json:"out_len"`
	Output  string `json:"output"`
}

func Filter(command, output string) Result {
	mode := detect(command)
	filtered := output
	switch mode {
	case "git-status":
		filtered = filterGitStatus(output)
	case "git-log":
		filtered = firstLines(output, 40)
	case "go-test", "pytest", "npm-test", "cargo-test":
		filtered = filterTests(output)
	case "grep", "rg":
		filtered = filterSearch(output)
	case "ls", "tree", "find":
		filtered = filterListing(output)
	default:
		filtered = truncateMiddle(output, 12000)
	}
	return Result{Command: command, Mode: mode, RawLen: len(output), OutLen: len(filtered), Output: filtered}
}

func detect(command string) string {
	c := strings.TrimSpace(command)
	fields := strings.Fields(c)
	if len(fields) == 0 {
		return "generic"
	}
	if fields[0] == "git" && len(fields) > 1 {
		switch fields[1] {
		case "status":
			return "git-status"
		case "log":
			return "git-log"
		case "diff":
			return "generic"
		}
	}
	if fields[0] == "go" && len(fields) > 1 && fields[1] == "test" {
		return "go-test"
	}
	if fields[0] == "pytest" {
		return "pytest"
	}
	if fields[0] == "cargo" && len(fields) > 1 && fields[1] == "test" {
		return "cargo-test"
	}
	if (fields[0] == "npm" || fields[0] == "pnpm" || fields[0] == "yarn" || fields[0] == "bun") && strings.Contains(c, "test") {
		return "npm-test"
	}
	if fields[0] == "grep" {
		return "grep"
	}
	if fields[0] == "rg" {
		return "rg"
	}
	if fields[0] == "ls" {
		return "ls"
	}
	if fields[0] == "tree" {
		return "tree"
	}
	if fields[0] == "find" {
		return "find"
	}
	return "generic"
}

func filterGitStatus(out string) string {
	counts := map[string]int{}
	var branch string
	for _, line := range lines(out) {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "On branch ") || strings.HasPrefix(s, "## ") {
			branch = s
			continue
		}
		switch {
		case strings.HasPrefix(s, "modified:") || strings.HasPrefix(s, "M ") || strings.HasPrefix(s, " M"):
			counts["modified"]++
		case strings.HasPrefix(s, "new file:") || strings.HasPrefix(s, "A ") || strings.HasPrefix(s, "??") || isLikelyIndentedPath(line):
			counts["added/untracked"]++
		case strings.HasPrefix(s, "deleted:") || strings.HasPrefix(s, "D ") || strings.HasPrefix(s, " D"):
			counts["deleted"]++
		case strings.HasPrefix(s, "renamed:") || strings.HasPrefix(s, "R "):
			counts["renamed"]++
		}
	}
	if len(counts) == 0 {
		return firstLines(out, 20)
	}
	keys := sortedKeys(counts)
	var b strings.Builder
	if branch != "" {
		fmt.Fprintf(&b, "%s\n", branch)
	}
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %d\n", k, counts[k])
	}
	return strings.TrimSpace(b.String())
}

func filterTests(out string) string {
	var keep []string
	for _, line := range lines(out) {
		s := strings.TrimSpace(line)
		lower := strings.ToLower(s)
		if strings.Contains(lower, "fail") || strings.Contains(lower, "error") || strings.Contains(lower, "panic") || strings.HasPrefix(s, "--- FAIL") || strings.HasPrefix(s, "FAIL") {
			keep = append(keep, line)
		}
	}
	if len(keep) == 0 {
		return lastInteresting(out, "ok / pass; no failures detected")
	}
	return truncateMiddle(strings.Join(keep, "\n"), 12000)
}

func filterSearch(out string) string {
	byFile := map[string]int{}
	var samples []string
	for _, line := range lines(out) {
		if strings.TrimSpace(line) == "" {
			continue
		}
		file := line
		if i := strings.IndexAny(line, ":"); i > 0 {
			file = line[:i]
		}
		byFile[file]++
		if len(samples) < 30 {
			samples = append(samples, line)
		}
	}
	if len(byFile) == 0 {
		return "no matches"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d matching file(s)\n", len(byFile))
	for _, f := range sortedKeys(byFile) {
		fmt.Fprintf(&b, "- %s: %d\n", f, byFile[f])
	}
	b.WriteString("\nSamples:\n")
	b.WriteString(strings.Join(samples, "\n"))
	return truncateMiddle(b.String(), 12000)
}

func filterListing(out string) string {
	byExt := map[string]int{}
	count := 0
	for _, line := range lines(out) {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "total ") {
			continue
		}
		count++
		ext := filepath.Ext(s)
		if ext == "" {
			ext = "[no ext/dir]"
		}
		byExt[ext]++
	}
	if count < 80 {
		return truncateMiddle(out, 12000)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d entries\n", count)
	for _, k := range sortedKeys(byExt) {
		fmt.Fprintf(&b, "- %s: %d\n", k, byExt[k])
	}
	return b.String()
}

func isLikelyIndentedPath(line string) bool {
	return strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "  ") && strings.Contains(strings.TrimSpace(line), ".")
}

func lines(s string) []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	return out
}
func sortedKeys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
func firstLines(s string, n int) string {
	ls := lines(s)
	if len(ls) <= n {
		return s
	}
	return strings.Join(ls[:n], "\n") + fmt.Sprintf("\n... (%d more lines)", len(ls)-n)
}
func truncateMiddle(s string, max int) string {
	if len(s) <= max {
		return s
	}
	half := max / 2
	return s[:half] + fmt.Sprintf("\n... (%d bytes omitted) ...\n", len(s)-max) + s[len(s)-half:]
}
func lastInteresting(s, ok string) string {
	ls := lines(s)
	if len(ls) == 0 {
		return ok
	}
	start := len(ls) - 20
	if start < 0 {
		start = 0
	}
	return ok + "\n" + strings.Join(ls[start:], "\n")
}
