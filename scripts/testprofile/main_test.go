package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSplitGoTestArgs(t *testing.T) {
	patterns, run, flags, err := splitGoTestArgs([]string{"-race", "-count=3", "./internal/web", "./internal/store", "-run", "Plan|Widget", "-benchmem", "-benchtime", "100ms"})
	if err != nil || run != "Plan|Widget" || !reflect.DeepEqual(patterns, []string{"./internal/web", "./internal/store"}) || !reflect.DeepEqual(flags, []string{"-race", "-count=3", "-benchmem", "-benchtime=100ms"}) {
		t.Fatalf("patterns=%v run=%q flags=%v err=%v", patterns, run, flags, err)
	}
}

func TestSplitGoTestArgsRejectsMissingAndUnsupportedFlags(t *testing.T) {
	for _, args := range [][]string{{"-run"}, {"-o", "elsewhere.test"}, {"-cpuprofile=x"}, {"-unknown"}} {
		if _, _, _, err := splitGoTestArgs(args); err == nil {
			t.Errorf("accepted %v", args)
		}
	}
}

func TestProfileModeAndOwnedCleanup(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			root := t.TempDir()
			log := filepath.Join(root, "calls")
			gobin := filepath.Join(root, "go")
			script := `#!/bin/sh
printf '%s\n' "$*" >> "$CALL_LOG"
case "$1" in
 list) echo github.com/rcarmo/gi/fixture ;;
 test)
  shift
  while [ "$#" -gt 0 ]; do
   case "$1" in -cpuprofile|-memprofile|-o) shift; printf fixture > "$1" ;; esac
   shift
  done
  printf '{"Action":"pass","Package":"github.com/rcarmo/gi/fixture","Elapsed":0.001}\n' ;;
 tool) printf 'flat flat%% sum%% cum cum%%\n1ms 100%% 100%% 1ms 100%% github.com/rcarmo/gi/fixture.Run\n' ;;
esac
`
			if err := os.WriteFile(gobin, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("GO", gobin)
			t.Setenv("CALL_LOG", log)
			t.Setenv("PROFILING", map[bool]string{false: "0", true: "1"}[enabled])
			t.Setenv("PROFILE_KEEP", "0")
			dir := filepath.Join(root, "profiles")
			if code := profileGo([]string{"./fixture"}, "TestFixture", []string{"-race", "-count=3"}, dir, 3); code != 0 {
				t.Fatal(code)
			}
			calls, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(calls), "-cpuprofile") != enabled {
				t.Fatal("profile mode ignored", string(calls))
			}
			if !strings.Contains(string(calls), "-race -count=3") {
				t.Fatal("test flags lost", string(calls))
			}
			entries, _ := os.ReadDir(dir)
			for _, entry := range entries {
				if entry.IsDir() {
					t.Fatal("raw captures retained", entry.Name())
				}
			}
			if !enabled && len(entries) != 0 {
				t.Fatal("ordinary tests created profile data")
			}
		})
	}
}

func TestDefaultProfileDirUsesProjectRoot(t *testing.T) {
	t.Setenv("GI_TEST_PROFILE_DIR", "")
	t.Setenv("PROJECT_TMP_ROOT", filepath.Join(t.TempDir(), "gi"))
	if got, want := defaultDir(), filepath.Join(os.Getenv("PROJECT_TMP_ROOT"), "runs", "profiling"); got != want {
		t.Fatal(got, want)
	}
}
