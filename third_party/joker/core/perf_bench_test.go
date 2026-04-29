package core

import (
	"io"
	"strings"
	"testing"
)

func compileBenchExpr(tb testing.TB, script string) Expr {
	tb.Helper()
	reader := NewReader(strings.NewReader(script), "<bench>")
	obj, err := TryRead(reader)
	if err != nil {
		tb.Fatalf("read script: %v", err)
	}
	if _, err := TryRead(reader); err != io.EOF {
		tb.Fatalf("benchmark script must contain exactly one form")
	}
	expr, err := TryParse(obj, &ParseContext{GlobalEnv: GLOBAL_ENV})
	if err != nil {
		tb.Fatalf("parse script: %v", err)
	}
	return expr
}

func BenchmarkEvalArithmeticLoop(b *testing.B) {
	expr := compileBenchExpr(b, `(loop [i 0 s 0]
  (if (= i 100000)
    s
    (recur (inc i) (+ s (rem (* i 7) 11)))))`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

func BenchmarkEvalRecursiveFib(b *testing.B) {
	expr := compileBenchExpr(b, `(letfn [(fib [n] (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))]
  (loop [i 3 s 0]
    (if (zero? i)
      s
      (recur (dec i) (+ s (fib 24))))))`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

func BenchmarkEvalWordFrequency(b *testing.B) {
	words := make([]string, 0, 4000)
	base := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"}
	for i := 0; i < 4000; i++ {
		words = append(words, base[i%len(base)])
	}
	text := strings.Join(words, " ")
	expr := compileBenchExpr(b, `(let [text "`+text+`"
      words (re-seq #"\S+" text)]
  (loop [i 0 counts {}]
    (if (= i (count words))
      (+ (get counts "theta" 0) (get counts "alpha" 0))
      (let [w (nth words i)]
        (recur (inc i) (assoc counts w (inc (get counts w 0))))))))`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

func BenchmarkEvalMapUpdateLoop(b *testing.B) {
	expr := compileBenchExpr(b, `(let [ks [:k0 :k1 :k2 :k3 :k4 :k5 :k6 :k7 :k8 :k9 :k10 :k11 :k12 :k13 :k14 :k15]]
  (loop [i 0 m {}]
    (if (= i 5000)
      (+ (get m :k0 0) (+ (get m :k7 0) (get m :k15 0)))
      (let [k (nth ks (rem i 16))]
        (recur (inc i) (assoc m k (+ 1 (get m k 0))))))))`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}

func BenchmarkEvalJSONParseSum(b *testing.B) {
	b.Skip("requires additional benchmark harness setup for classpath-aware std namespace loading")
}

func BenchmarkEvalTailRecursiveSum(b *testing.B) {
	expr := compileBenchExpr(b, `(letfn [(sum [n acc]
    (if (zero? n)
      acc
      (sum (dec n) (+ acc n))))]
  (sum 100000 0))`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Eval(expr, nil)
	}
}
