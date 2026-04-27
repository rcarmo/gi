//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rcarmo/gi/internal/scripting"
)

type runtimeResult struct {
	Name    string
	Value   string
	Runs    []time.Duration
	Median  time.Duration
	Average time.Duration
	Err     error
}

type task struct {
	Name        string
	Description string
	GoFn        func() string
	PythonCode  string
	GojaCode    string
	JokerCode   string
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

func bench(name string, runs int, fn func() (string, error)) runtimeResult {
	res := runtimeResult{Name: name}
	for i := 0; i < runs; i++ {
		start := time.Now()
		value, err := fn()
		d := time.Since(start)
		res.Runs = append(res.Runs, d)
		if err != nil {
			res.Err = err
			res.Value = value
			break
		}
		res.Value = value
	}
	res.Median = median(res.Runs)
	res.Average = average(res.Runs)
	return res
}

func fibGo(n int) int {
	if n < 2 {
		return n
	}
	return fibGo(n-1) + fibGo(n-2)
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func main() {
	ctx := context.Background()
	const runs = 3

	_, _ = scripting.NewGojaRunner().Execute(ctx, "1+1", &scripting.Bridge{})
	_, _ = scripting.NewJokerRunner().Execute(ctx, "(+ 1 1)", &scripting.Bridge{})

	jsonRecords := make([]map[string]any, 0, 500)
	for i := 0; i < 500; i++ {
		jsonRecords = append(jsonRecords, map[string]any{
			"id":    i,
			"group": fmt.Sprintf("g%d", i%10),
			"value": (i % 17) + 1,
			"name":  fmt.Sprintf("item-%03d", i),
		})
	}
	jsonPayload := mustJSON(jsonRecords)

	textWords := make([]string, 0, 12000)
	base := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"}
	for i := 0; i < 12000; i++ {
		textWords = append(textWords, base[i%len(base)])
	}
	textPayload := strings.Join(textWords, " ")

	tasks := []task{
		{
			Name:        "recursive-fib",
			Description: "CPU-bound recursion: fib(28) repeated 5 times inside one invocation.",
			GoFn: func() string {
				sum := 0
				for i := 0; i < 5; i++ {
					sum += fibGo(28)
				}
				return strconv.Itoa(sum)
			},
			PythonCode: `def fib(n):
    return n if n < 2 else fib(n-1) + fib(n-2)
s = 0
for _ in range(5):
    s += fib(28)
print(s)`,
			GojaCode: `function fib(n){ return n < 2 ? n : fib(n-1) + fib(n-2); }
let s = 0;
for (let i = 0; i < 5; i++) s += fib(28);
s.toString();`,
			JokerCode: `(defn fib [n] (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))
(loop [i 5 s 0]
  (if (zero? i)
    (str s)
    (recur (dec i) (+ s (fib 28)))))`,
		},
		{
			Name:        "arith-loop",
			Description: "Tight arithmetic loop: 1,000,000 iterations accumulating a modular sum.",
			GoFn: func() string {
				sum := 0
				for i := 0; i < 1000000; i++ {
					sum += (i * 7) % 11
				}
				return strconv.Itoa(sum)
			},
			PythonCode: `s = 0
for i in range(1000000):
    s += (i * 7) % 11
print(s)`,
			GojaCode: `let s = 0;
for (let i = 0; i < 1000000; i++) s += (i * 7) % 11;
s.toString();`,
			JokerCode: `(loop [i 0 s 0]
  (if (= i 1000000)
    (str s)
    (recur (inc i) (+ s (mod (* i 7) 11)))))`,
		},
		{
			Name:        "json-parse-sum",
			Description: "JSON workload: parse a 500-record JSON array 20 times and sum the \"value\" fields.",
			GoFn: func() string {
				var rows []struct {
					Value int `json:"value"`
				}
				total := 0
				for i := 0; i < 20; i++ {
					rows = rows[:0]
					if err := json.Unmarshal([]byte(jsonPayload), &rows); err != nil {
						panic(err)
					}
					for _, row := range rows {
						total += row.Value
					}
				}
				return strconv.Itoa(total)
			},
			PythonCode: fmt.Sprintf(`import json
payload = %q
total = 0
for _ in range(20):
    rows = json.loads(payload)
    for row in rows:
        total += row["value"]
print(total)`, jsonPayload),
			GojaCode: fmt.Sprintf(`const payload = %q;
let total = 0;
for (let i = 0; i < 20; i++) {
  const rows = JSON.parse(payload);
  for (const row of rows) total += row.value;
}
total.toString();`, jsonPayload),
			JokerCode: fmt.Sprintf(`(require '[joker.json :as json])
(def payload %q)
(loop [i 20 total 0]
  (if (zero? i)
    (str total)
    (let [rows (json/read-string payload)
          subtotal (reduce (fn [acc row] (+ acc (get row "value"))) 0 rows)]
      (recur (dec i) (+ total subtotal)))))`, jsonPayload),
		},
		{
			Name:        "word-frequency",
			Description: "Text processing: split a 12k-word string and build a frequency map.",
			GoFn: func() string {
				counts := map[string]int{}
				for _, word := range strings.Fields(textPayload) {
					counts[word]++
				}
				return strconv.Itoa(counts["theta"] + counts["alpha"])
			},
			PythonCode: fmt.Sprintf(`text = %q
counts = {}
for word in text.split():
    counts[word] = counts.get(word, 0) + 1
print(counts["theta"] + counts["alpha"])`, textPayload),
			GojaCode: fmt.Sprintf(`const text = %q;
const counts = Object.create(null);
for (const word of text.split(/\s+/)) counts[word] = (counts[word] || 0) + 1;
(counts["theta"] + counts["alpha"]).toString();`, textPayload),
			JokerCode: fmt.Sprintf(`(def text %q)
(def words (re-seq #"\S+" text))
(loop [i 0 counts {}]
  (if (= i (count words))
    (str (+ (get counts "theta" 0) (get counts "alpha" 0)))
    (let [w (nth words i)]
      (recur (inc i) (assoc counts w (inc (get counts w 0)))))))`, textPayload),
		},
	}

	var report strings.Builder
	report.WriteString("# Scripting benchmark suite\n\n")
	report.WriteString("This report compares **Go**, **Python**, **Goja**, and **Joker** on a small set of common scripting-style workloads.\n\n")
	report.WriteString("Methodology:\n")
	report.WriteString("- 3 runs per runtime/task\n")
	report.WriteString("- median reported as the main comparison metric\n")
	report.WriteString("- each runtime executes the full task in a single invocation\n")
	report.WriteString("- Goja/Joker use gi's embedded scripting runners\n")
	report.WriteString("- Python uses `python3 -c ...`\n\n")

	report.WriteString("## Summary\n\n")
	report.WriteString("Across these tasks, the rough ranking is consistently **Go > Python > Goja > Joker**. Python is often an order of magnitude slower than Go but still practical; Goja is generally slower than Python for these workloads; Joker is far slower for CPU-heavy tasks.\n\n")

	for _, task := range tasks {
		results := []runtimeResult{
			bench("go", runs, func() (string, error) { return task.GoFn(), nil }),
			bench("python", runs, func() (string, error) {
				cmd := exec.CommandContext(ctx, "python3", "-c", task.PythonCode)
				b, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(b)), err
			}),
			bench("goja", runs, func() (string, error) {
				return scripting.NewGojaRunner().Execute(ctx, task.GojaCode, &scripting.Bridge{})
			}),
			bench("joker", runs, func() (string, error) {
				return scripting.NewJokerRunner().Execute(ctx, task.JokerCode, &scripting.Bridge{})
			}),
		}

		report.WriteString("## " + task.Name + "\n\n")
		report.WriteString(task.Description + "\n\n")
		report.WriteString("| Runtime | Result | Median | Average | Runs | Slowdown vs Go |\n")
		report.WriteString("|---|---:|---:|---:|---|---:|\n")
		goMedian := results[0].Median
		for _, r := range results {
			runParts := make([]string, 0, len(r.Runs))
			for _, d := range r.Runs {
				runParts = append(runParts, d.String())
			}
			slowdown := float64(r.Median) / float64(goMedian)
			value := r.Value
			if r.Err != nil {
				value = "ERR: " + strings.ReplaceAll(r.Err.Error(), "|", "/")
			}
			report.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %.1fx |\n", r.Name, value, r.Median, r.Average, strings.Join(runParts, ", "), slowdown))
		}
		report.WriteString("\n")
	}

	report.WriteString("## Interpretation\n\n")
	report.WriteString("- **Go** is the clear performance baseline and best fit for compute-heavy or latency-sensitive logic.\n")
	report.WriteString("- **Python** is much slower than Go, but still reasonable for many orchestration and data-processing tasks.\n")
	report.WriteString("- **Goja** is viable for embedded scripting and control logic, but not competitive with Python on these heavier microbenchmarks.\n")
	report.WriteString("- **Joker** is expressive and useful for scripting ergonomics, but should not be treated as a high-performance compute runtime.\n")

	path := "tmp/scripting-benchmark-report.md"
	if err := os.WriteFile(path, []byte(report.String()), 0644); err != nil {
		panic(err)
	}
	fmt.Print(report.String())
}
