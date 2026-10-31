package inference

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Native credential tests must never inherit the running agent's auth path.
// The Makefile supplies project-owned TMPDIR. Individual path-override tests
// can still opt into fixtures; an unconfigured runner fails closed.
func TestMain(m *testing.M) {
	if os.Getenv("GI_CREDENTIAL_TEST_MODE") != "" {
		// Cross-process lock/reopen probes inherit an already isolated fixture.
		root, home := os.Getenv("GI_TEST_RUN_ROOT"), os.Getenv("HOME")
		relative, err := filepath.Rel(root, home)
		if root == "" || !filepath.IsAbs(root) || err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			fmt.Fprintln(os.Stderr, "credential helper requires a project-owned fixture home")
			os.Exit(1)
		}
		os.Unsetenv("PI_CODING_AGENT_DIR")
		os.Unsetenv("GI_CODING_AGENT_DIR")
		os.Exit(m.Run())
	}
	root := os.Getenv("GI_TEST_RUN_ROOT")
	relative, err := filepath.Rel(root, os.TempDir())
	if root == "" || !filepath.IsAbs(root) || err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		fmt.Fprintln(os.Stderr, "run inference tests through make with project-owned scratch")
		os.Exit(1)
	}
	home, err := os.MkdirTemp("", "inference-home-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, key := range []string{"PI_CODING_AGENT_DIR", "GI_CODING_AGENT_DIR"} {
		if err = os.Unsetenv(key); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.RemoveAll(home)
			os.Exit(1)
		}
	}
	if err = os.Setenv("HOME", home); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.RemoveAll(home)
		os.Exit(1)
	}
	code := m.Run()
	if err = os.RemoveAll(home); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}
