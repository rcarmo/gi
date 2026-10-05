// Command testprofile runs gi's test runs with profiling and
// analysis, so the suites stay lean: every run reports wall and CPU time,
// peak memory, the slowest packages and tests, CPU and allocation hot spots,
// and what got slower since the previous run. Reports and a history log are
// kept under -dir (default ~/.cache/gi-test-profile, on disk).
//
//	testprofile go [-run regexp] [packages]   # Go suite, one package at a time
//	testprofile run -name NAME -- cmd args... # any suite: time, CPU, memory
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

type pkgResult struct {
	Package    string             `json:"package"`
	Status     string             `json:"status"`    // ok, fail, skip
	Elapsed    float64            `json:"elapsed_s"` // wall, build included
	TestTime   float64            `json:"test_s"`    // reported by go test
	CPU        float64            `json:"cpu_s"`     // user+sys of go and the test binary
	Tests      map[string]float64 `json:"tests"`
	Failed     []string           `json:"failed,omitempty"`
	CPUProfile string             `json:"cpu_profile,omitempty"`
	MemProfile string             `json:"mem_profile,omitempty"`
}

type report struct {
	Name     string       `json:"name"`
	Started  time.Time    `json:"started"`
	Wall     float64      `json:"wall_s"`
	CPU      float64      `json:"cpu_s"`
	PeakRSS  int64        `json:"peak_rss_kb"`
	PeakBy   string       `json:"peak_rss_by,omitempty"`
	ExitCode int          `json:"exit_code"`
	Packages []*pkgResult `json:"packages,omitempty"`
}

func childrenUsage() (cpu float64, maxRSS int64) {
	var ru syscall.Rusage
	if syscall.Getrusage(syscall.RUSAGE_CHILDREN, &ru) != nil {
		return 0, 0
	}
	tv := func(t syscall.Timeval) float64 { return float64(t.Sec) + float64(t.Usec)/1e6 }
	rss := int64(ru.Maxrss)
	if runtime.GOOS == "darwin" { // Darwin reports bytes; Linux reports KiB.
		rss /= 1024
	}
	return tv(ru.Utime) + tv(ru.Stime), rss
}

func defaultDir() string {
	if d := os.Getenv("GI_TEST_PROFILE_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "gi-test-profile")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: testprofile go|run ...")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "go":
		os.Exit(goMode(os.Args[2:]))
	case "run":
		os.Exit(runMode(os.Args[2:]))
	case "gotest":
		os.Exit(goTestMode(os.Args[2:]))
	}
	fmt.Fprintln(os.Stderr, "usage: testprofile go|run ...")
	os.Exit(2)
}

// ── run: any command ────────────────────────────────────────────────────

func runMode(args []string) int {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	name := fs.String("name", "suite", "suite name (history key)")
	key := fs.String("key", "", "selection key for a separate comparison baseline")
	dir := fs.String("dir", defaultDir(), "report directory")
	_ = fs.Parse(args)
	cmdArgs := fs.Args()
	if len(cmdArgs) == 0 {
		fmt.Fprintln(os.Stderr, "testprofile run: no command")
		return 2
	}
	reportName := *name
	if *key != "" {
		h := fnv.New32a()
		_, _ = h.Write([]byte(*key))
		reportName = fmt.Sprintf("%s-%08x", *name, h.Sum32())
	}
	rep := &report{Name: reportName, Started: time.Now()}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	rep.Wall = time.Since(rep.Started).Seconds()
	rep.CPU, rep.PeakRSS = childrenUsage()
	rep.ExitCode = exitCode(err)
	prev := lastReport(*dir, reportName)
	save(*dir, rep)
	fmt.Printf("\n── %s profile ── wall %s · cpu %s · peak RSS %s%s\n", *name, dur(rep.Wall), dur(rep.CPU), mb(rep.PeakRSS), compareTotals(prev, rep))
	return rep.ExitCode
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	return 1
}

// ── go: the Go suite ────────────────────────────────────────────────────

type testEvent struct {
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}

func goMode(args []string) int {
	fs := flag.NewFlagSet("go", flag.ExitOnError)
	run := fs.String("run", "", "go test -run")
	dir := fs.String("dir", defaultDir(), "report directory")
	hot := fs.Int("hot", 3, "packages to show CPU/allocation hot spots for")
	_ = fs.Parse(args)
	return profileGo(fs.Args(), *run, nil, *dir, *hot)
}

// goTestMode accepts the go test flags used by focused Makefile targets.
// Profiles and uncached execution are added without dropping race/repeat/bench flags.
func goTestMode(args []string) int {
	patterns, run, flags, err := splitGoTestArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return profileGo(patterns, run, flags, defaultDir(), 3)
}

func splitGoTestArgs(args []string) (patterns []string, run string, flags []string, err error) {
	values := map[string]bool{"-run": true, "-count": true, "-bench": true, "-benchtime": true, "-timeout": true, "-parallel": true, "-tags": true}
	bools := map[string]bool{"-race": true, "-v": true, "-benchmem": true, "-short": true}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			patterns = append(patterns, a)
			continue
		}
		key, value, hasValue := strings.Cut(a, "=")
		if !values[key] && !bools[key] {
			return nil, "", nil, fmt.Errorf("testprofile gotest: unsupported flag %s", key)
		}
		if values[key] && !hasValue {
			i++
			if i == len(args) {
				return nil, "", nil, fmt.Errorf("%s needs a value", key)
			}
			value = args[i]
		}
		if key == "-run" {
			run = value
		} else if values[key] || hasValue {
			flags = append(flags, key+"="+value)
		} else {
			flags = append(flags, a)
		}
	}
	return
}

func profileGo(patterns []string, run string, flags []string, dir string, hot int) int {
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	gobin := os.Getenv("GO")
	if gobin == "" {
		gobin = "go"
	}
	list, err := exec.Command(gobin, append([]string{"list", "-f", "{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}"}, patterns...)...).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "go list:", err)
		return 1
	}
	// Filters/repeats/benchmarks need separate baselines from the full suite.
	name := "go"
	if strings.Join(patterns, " ") != "./..." || run != "" || len(flags) != 0 {
		h := fnv.New32a()
		_, _ = h.Write([]byte(strings.Join(patterns, " ") + "\x00" + run + "\x00" + strings.Join(flags, " ")))
		name = fmt.Sprintf("go-subset-%08x", h.Sum32())
	}
	rep := &report{Name: name, Started: time.Now()}
	runDir := filepath.Join(dir, "go-"+rep.Started.Format("20060102-150405"))
	_ = os.MkdirAll(runDir, 0o755)
	var lastCPU float64
	var lastRSS int64
	for _, pkg := range strings.Fields(string(list)) {
		res := &pkgResult{Package: pkg, Tests: map[string]float64{}}
		short := strings.TrimPrefix(pkg, "github.com/rcarmo/gi/")
		pdir := filepath.Join(runDir, strings.ReplaceAll(short, "/", "_"))
		_ = os.MkdirAll(pdir, 0o755)
		res.CPUProfile, res.MemProfile = filepath.Join(pdir, "cpu.pprof"), filepath.Join(pdir, "mem.pprof")
		bin := filepath.Join(pdir, "pkg.test")
		testArgs := []string{"test", "-json", "-count=1", "-cpuprofile", res.CPUProfile, "-memprofile", res.MemProfile, "-o", bin}
		testArgs = append(testArgs, flags...)
		if run != "" {
			testArgs = append(testArgs, "-run", run)
		}
		start := time.Now()
		cmd := exec.Command(gobin, append(testArgs, pkg)...)
		cmd.Stderr = os.Stderr
		out, _ := cmd.StdoutPipe()
		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		output := consume(out, res)
		_ = os.WriteFile(filepath.Join(pdir, "test.log"), []byte(output), 0o644)
		err := cmd.Wait()
		_ = os.Remove(bin) // profiles carry their symbols; binaries are large
		res.Elapsed = time.Since(start).Seconds()
		cpu, rss := childrenUsage()
		res.CPU, lastCPU = cpu-lastCPU, cpu
		if rss > lastRSS {
			lastRSS, rep.PeakBy = rss, short
		}
		if res.Status == "" {
			res.Status = "ok"
		}
		if err != nil && res.Status == "ok" {
			res.Status = "fail"
		}
		if res.Status == "fail" {
			rep.ExitCode = 1
			fmt.Print(filterOutput(output))
			fmt.Printf("FAIL\t%s\t%.3fs\n", pkg, res.TestTime)
		} else {
			fmt.Printf("ok  \t%s\t%.3fs\n", pkg, res.TestTime)
			for _, line := range strings.Split(output, "\n") {
				if strings.HasPrefix(line, "Benchmark") {
					fmt.Println(line)
				}
			}
		}
		for _, profile := range []string{res.CPUProfile, res.MemProfile} {
			if info, err := os.Stat(profile); err != nil || info.Size() == 0 {
				fmt.Fprintf(os.Stderr, "MISSING profile (not a profiling pass): %s\n", profile)
			}
		}
		rep.Packages = append(rep.Packages, res)
	}
	rep.Wall = time.Since(rep.Started).Seconds()
	rep.CPU, rep.PeakRSS = childrenUsage()
	prev := lastReport(dir, rep.Name)
	save(dir, rep)
	analyze(rep, prev, hot, runDir)
	pruneRuns(dir, "go-", 5)
	return rep.ExitCode
}

// consume reads go test -json events into res and returns the package's
// text output.
func consume(r io.Reader, res *pkgResult) string {
	var b strings.Builder
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 16<<20)
	for sc.Scan() {
		var ev testEvent
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			b.WriteString(sc.Text() + "\n")
			continue
		}
		b.WriteString(ev.Output)
		switch {
		case ev.Test != "" && (ev.Action == "pass" || ev.Action == "fail"):
			res.Tests[ev.Test] = ev.Elapsed
			if ev.Action == "fail" {
				res.Failed = append(res.Failed, ev.Test)
			}
		case ev.Test == "" && ev.Action == "pass":
			res.Status, res.TestTime = "ok", ev.Elapsed
		case ev.Test == "" && ev.Action == "fail":
			res.Status, res.TestTime = "fail", ev.Elapsed
		case ev.Test == "" && ev.Action == "skip":
			res.Status = "skip"
		}
	}
	return b.String()
}

// filterOutput drops go test -v progress lines from a failed package.
func filterOutput(s string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "=== RUN") || strings.HasPrefix(t, "=== PAUSE") || strings.HasPrefix(t, "=== CONT") || strings.HasPrefix(t, "--- PASS") || strings.HasPrefix(t, "PASS") || t == "" {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

// ── analysis ────────────────────────────────────────────────────────────

func analyze(rep, prev *report, hot int, runDir string) {
	tests, failed := 0, 0
	for _, p := range rep.Packages {
		tests += len(p.Tests)
		failed += len(p.Failed)
	}
	// Peak RSS is RUSAGE_CHILDREN's: the largest process go test ran,
	// including the compiler and linker (linking internal/web's test binary
	// peaks near 1.3 GB while the tests stay near 100 MB).
	fmt.Printf("\n── go test profile ── %d packages · %d tests (%d failed) · wall %s · cpu %s · peak RSS %s incl. build (%s)%s\n",
		len(rep.Packages), tests, failed, dur(rep.Wall), dur(rep.CPU), mb(rep.PeakRSS), rep.PeakBy, compareTotals(prev, rep))

	prevPkg := map[string]*pkgResult{}
	if prev != nil {
		for _, p := range prev.Packages {
			prevPkg[p.Package] = p
		}
	}
	pkgs := append([]*pkgResult(nil), rep.Packages...)
	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Elapsed > pkgs[j].Elapsed })
	fmt.Println("slowest packages (wall incl. build · test · cpu):")
	for _, p := range pkgs[:min(8, len(pkgs))] {
		fmt.Printf("  %-36s %7s · %7s · %7s%s\n", short(p.Package), dur(p.Elapsed), dur(p.TestTime), dur(p.CPU), delta(prevPkg[p.Package], p))
	}

	type testTime struct {
		name string
		s    float64
	}
	var all []testTime
	prevTest := map[string]float64{}
	for _, p := range rep.Packages {
		for name, s := range p.Tests {
			if !strings.Contains(name, "/") {
				all = append(all, testTime{short(p.Package) + "." + name, s})
			}
		}
	}
	if prev != nil {
		for _, p := range prev.Packages {
			for name, s := range p.Tests {
				prevTest[short(p.Package)+"."+name] = s
			}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].s > all[j].s })
	fmt.Println("slowest tests:")
	for _, t := range all[:min(10, len(all))] {
		note := ""
		if old, ok := prevTest[t.name]; ok && t.s > old*1.25 && t.s-old > 0.5 {
			note = fmt.Sprintf("  ▲ was %s", dur(old))
		}
		name := t.name
		if len(name) > 72 {
			name = name[:71] + "…"
		}
		fmt.Printf("  %-72s %7s%s\n", name, dur(t.s), note)
	}

	var regressions []string
	for _, p := range rep.Packages {
		if old := prevPkg[p.Package]; old != nil && p.TestTime > old.TestTime*1.25 && p.TestTime-old.TestTime > 2 {
			regressions = append(regressions, fmt.Sprintf("%s test time %s → %s", short(p.Package), dur(old.TestTime), dur(p.TestTime)))
		}
	}
	for _, t := range all {
		if old, ok := prevTest[t.name]; ok && t.s > old*1.5 && t.s-old > 1 {
			regressions = append(regressions, fmt.Sprintf("%s %s → %s", t.name, dur(old), dur(t.s)))
		}
	}
	if len(regressions) > 0 {
		fmt.Println("slower than the previous run:")
		for _, r := range regressions[:min(10, len(regressions))] {
			fmt.Println("  ▲ " + r)
		}
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].TestTime > pkgs[j].TestTime })
	fmt.Println("hot spots (flat CPU, then allocated bytes) in the slowest packages:")
	for _, p := range pkgs[:min(hot, len(pkgs))] {
		fmt.Printf("  %s\n", short(p.Package))
		for _, line := range pprofTop(p.CPUProfile, nil, 5) {
			fmt.Println("    cpu   " + line)
		}
		for _, line := range pprofTop(p.MemProfile, []string{"-sample_index=alloc_space"}, 3) {
			fmt.Println("    alloc " + line)
		}
	}
	fmt.Printf("disk: go build cache %s · go tmp %s · profiles %s\n", dirSize(os.Getenv("GOCACHE")), dirSize(os.Getenv("GOTMPDIR")), dirSize(filepath.Dir(runDir)))
	fmt.Printf("report: %s\n", filepath.Join(runDir, "report.json"))
}

// pprofTop is the top n rows of go tool pprof -top, with samples charged to
// the nearest gi function (-show), so hot spots name gi code rather than the
// runtime: "flat flat% function".
func pprofTop(profile string, extra []string, n int) []string {
	if info, err := os.Stat(profile); err != nil || info.Size() == 0 {
		return nil
	}
	gobin := os.Getenv("GO")
	if gobin == "" {
		gobin = "go"
	}
	args := append([]string{"tool", "pprof", "-top", fmt.Sprintf("-nodecount=%d", n), "-show=github.com/rcarmo/gi"}, extra...)
	out, err := exec.Command(gobin, append(args, profile)...).Output()
	if err != nil {
		return nil
	}
	var rows []string
	header := false
	for _, line := range strings.Split(string(out), "\n") {
		f := strings.Fields(line)
		if len(f) >= 5 && f[0] == "flat" {
			header = true
			continue
		}
		if header && len(f) >= 6 && f[0] != "0" {
			rows = append(rows, fmt.Sprintf("%8s %6s  %s", f[0], f[1], short(strings.Join(f[5:], " "))))
		}
	}
	return rows
}

// ── history ─────────────────────────────────────────────────────────────

func save(dir string, rep *report) {
	_ = os.MkdirAll(dir, 0o755)
	if b, err := json.MarshalIndent(rep, "", " "); err == nil {
		if strings.HasPrefix(rep.Name, "go") {
			_ = os.WriteFile(filepath.Join(dir, "go-"+rep.Started.Format("20060102-150405"), "report.json"), b, 0o644)
		}
		if rep.ExitCode == 0 { // the baseline for the next run
			_ = os.WriteFile(filepath.Join(dir, "latest-"+rep.Name+".json"), b, 0o644)
		}
	}
	totals := map[string]any{"name": rep.Name, "started": rep.Started, "wall_s": rep.Wall, "cpu_s": rep.CPU, "peak_rss_kb": rep.PeakRSS, "exit_code": rep.ExitCode, "packages": len(rep.Packages)}
	if f, err := os.OpenFile(filepath.Join(dir, "history.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
		b, _ := json.Marshal(totals)
		_, _ = f.Write(append(b, '\n'))
		_ = f.Close()
	}
}

// lastReport is the previous complete run of a suite (failed runs are kept
// in the history but not used as a baseline).
func lastReport(dir, name string) *report {
	b, err := os.ReadFile(filepath.Join(dir, "latest-"+name+".json"))
	if err != nil {
		return nil
	}
	var r report
	if json.Unmarshal(b, &r) != nil {
		return nil
	}
	return &r
}

func pruneRuns(dir, prefix string, keep int) {
	entries, _ := os.ReadDir(dir)
	var runs []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			runs = append(runs, e.Name())
		}
	}
	sort.Strings(runs)
	for len(runs) > keep {
		_ = os.RemoveAll(filepath.Join(dir, runs[0]))
		runs = runs[1:]
	}
}

// ── formatting ──────────────────────────────────────────────────────────

func compareTotals(prev, cur *report) string {
	if prev == nil || prev.Wall == 0 {
		return ""
	}
	return fmt.Sprintf("\n   vs previous (%s): wall %s · cpu %s · peak RSS %s",
		prev.Started.Format("Jan 2 15:04"), pct(prev.Wall, cur.Wall), pct(prev.CPU, cur.CPU), pct(float64(prev.PeakRSS), float64(cur.PeakRSS)))
}

func delta(prev, cur *pkgResult) string {
	if prev == nil || prev.TestTime == 0 {
		return ""
	}
	return "  " + pct(prev.TestTime, cur.TestTime)
}

func pct(old, cur float64) string {
	if old == 0 {
		return "n/a"
	}
	p := (cur - old) / old * 100
	mark := ""
	if p > 25 {
		mark = " ▲"
	}
	return fmt.Sprintf("%+.0f%%%s", p, mark)
}

func dur(s float64) string {
	d := time.Duration(s * float64(time.Second))
	if d >= time.Minute {
		return d.Round(time.Second).String()
	}
	return fmt.Sprintf("%.1fs", s)
}

func mb(kb int64) string { return fmt.Sprintf("%d MB", kb/1024) }

func short(pkg string) string { return strings.TrimPrefix(pkg, "github.com/rcarmo/gi/") }

func dirSize(dir string) string {
	if dir == "" {
		return "?"
	}
	var total int64
	_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return fmt.Sprintf("%d MB", total>>20)
}
