//go:build fixtures_vibes

package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	goai "github.com/rcarmo/go-ai"
)

var fixtureCPU *os.File
var fixtureProfileDir string

// These models exist only in the disposable compliance build. Production Gi
// never registers a provider from an environment-controlled URL.
func init() {
	// Per-worker profiles include fixture/runtime setup and graceful teardown.
	if root := os.Getenv("GI_FIXTURE_PROFILE_DIR"); root != "" {
		var err error
		fixtureProfileDir = filepath.Join(root, fmt.Sprintf("runtime-%d", os.Getpid()))
		if err = os.MkdirAll(fixtureProfileDir, 0o700); err != nil {
			panic(err)
		}
		fixtureCPU, err = os.Create(filepath.Join(fixtureProfileDir, "cpu.pprof"))
		if err != nil {
			panic(err)
		}
		if err = pprof.StartCPUProfile(fixtureCPU); err != nil {
			panic(err)
		}
	}
	base := strings.TrimRight(os.Getenv("FIXTURE_MODEL_URL"), "/")
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.Path != "/v1" {
		panic("fixtures_vibes requires a loopback FIXTURE_MODEL_URL ending in /v1")
	}
	for _, id := range []string{"fixture-1", "fixture-2"} {
		goai.RegisterModel(&goai.Model{
			ID: id, Name: "Fixture model " + id, Provider: "fixture-vibes",
			Api: goai.ApiOpenAICompletions, BaseURL: base,
			// Shared fixture contract: both advertised models have a 128K window.
			Input: []string{"text"}, ContextWindow: 128000, MaxTokens: 1024,
		})
	}
}

func finishFixtureProfiles() {
	if fixtureCPU == nil {
		return
	}
	pprof.StopCPUProfile()
	fixtureCPU.Close()
	runtime.GC()
	heap, err := os.Create(filepath.Join(fixtureProfileDir, "mem.pprof"))
	if err != nil {
		panic(err)
	}
	defer heap.Close()
	if err = pprof.WriteHeapProfile(heap); err != nil {
		panic(err)
	}
}
