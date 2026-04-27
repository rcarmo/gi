//go:build ignore

package main

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/scripting"
)

func fibGo(n int) int {
	if n < 2 {
		return n
	}
	return fibGo(n-1) + fibGo(n-2)
}

type result struct {
	Name    string
	Value   string
	Runs    []time.Duration
	Median  time.Duration
	Average time.Duration
}

func median(ds []time.Duration) time.Duration {
	cp := append([]time.Duration(nil), ds...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	return cp[len(cp)/2]
}

func average(ds []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range ds {
		total += d
	}
	return total / time.Duration(len(ds))
}

func bench(name string, runs int, fn func() (string, error)) result {
	out := result{Name: name}
	for i := 0; i < runs; i++ {
		start := time.Now()
		v, err := fn()
		d := time.Since(start)
		if err != nil {
			out.Value = "ERR: " + err.Error()
			out.Runs = append(out.Runs, d)
			break
		}
		out.Value = v
		out.Runs = append(out.Runs, d)
	}
	out.Median = median(out.Runs)
	out.Average = average(out.Runs)
	return out
}

func main() {
	ctx := context.Background()
	const n = 30
	const runs = 3

	_, _ = scripting.NewGojaRunner().Execute(ctx, "1+1", &scripting.Bridge{})
	_, _ = scripting.NewJokerRunner().Execute(ctx, "(+ 1 1)", &scripting.Bridge{})

	results := []result{
		bench("go", runs, func() (string, error) {
			return strconv.Itoa(fibGo(n)), nil
		}),
		bench("python", runs, func() (string, error) {
			code := fmt.Sprintf(`def fib(n):
    return n if n < 2 else fib(n-1) + fib(n-2)
print(fib(%d))`, n)
			cmd := exec.CommandContext(ctx, "python3", "-c", code)
			b, err := cmd.CombinedOutput()
			return strings.TrimSpace(string(b)), err
		}),
		bench("goja", runs, func() (string, error) {
			script := fmt.Sprintf(`function fib(n){ return n < 2 ? n : fib(n-1) + fib(n-2); }
fib(%d).toString();`, n)
			return scripting.NewGojaRunner().Execute(ctx, script, &scripting.Bridge{})
		}),
		bench("joker", runs, func() (string, error) {
			script := fmt.Sprintf(`(defn fib [n] (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))
(str (fib %d))`, n)
			return scripting.NewJokerRunner().Execute(ctx, script, &scripting.Bridge{})
		}),
	}

	fmt.Printf("Benchmark: recursive fib(%d), %d runs per runtime\n\n", n, runs)
	fmt.Printf("| Runtime | Result | Median | Average | Runs | Slowdown vs Go (median) |\n")
	fmt.Printf("|---|---:|---:|---:|---|---:|\n")
	goMedian := results[0].Median
	for _, r := range results {
		runParts := make([]string, 0, len(r.Runs))
		for _, d := range r.Runs {
			runParts = append(runParts, d.String())
		}
		slowdown := float64(r.Median) / float64(goMedian)
		fmt.Printf("| %s | %s | %s | %s | %s | %.1fx |\n", r.Name, r.Value, r.Median, r.Average, strings.Join(runParts, ", "), slowdown)
	}
}
