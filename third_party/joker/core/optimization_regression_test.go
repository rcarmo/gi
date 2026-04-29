package core

import (
	"io"
	"strings"
	"testing"
)

func compileTestExpr(tb testing.TB, script string) Expr {
	tb.Helper()
	reader := NewReader(strings.NewReader(script), "<test>")
	obj, err := TryRead(reader)
	if err != nil {
		tb.Fatalf("read script: %v", err)
	}
	if _, err := TryRead(reader); err != io.EOF {
		tb.Fatalf("test script must contain exactly one form")
	}
	expr, err := TryParse(obj, &ParseContext{GlobalEnv: GLOBAL_ENV})
	if err != nil {
		tb.Fatalf("parse script: %v", err)
	}
	return expr
}

func evalTestScript(tb testing.TB, script string) Object {
	tb.Helper()
	return Eval(compileTestExpr(tb, script), nil)
}

func requireInt(tb testing.TB, obj Object, want int) {
	tb.Helper()
	got, ok := obj.(Int)
	if !ok {
		tb.Fatalf("expected Int(%d), got %T (%s)", want, obj, obj.ToString(false))
	}
	if got.I != want {
		tb.Fatalf("expected Int(%d), got Int(%d)", want, got.I)
	}
}

func requireDouble(tb testing.TB, obj Object, want float64) {
	tb.Helper()
	got, ok := obj.(Double)
	if !ok {
		tb.Fatalf("expected Double(%v), got %T (%s)", want, obj, obj.ToString(false))
	}
	if got.D != want {
		tb.Fatalf("expected Double(%v), got Double(%v)", want, got.D)
	}
}

func requireBool(tb testing.TB, obj Object, want bool) {
	tb.Helper()
	got, ok := obj.(Boolean)
	if !ok {
		tb.Fatalf("expected Boolean(%v), got %T (%s)", want, obj, obj.ToString(false))
	}
	if got.B != want {
		tb.Fatalf("expected Boolean(%v), got Boolean(%v)", want, got.B)
	}
}

func requireString(tb testing.TB, obj Object, want string) {
	tb.Helper()
	got, ok := obj.(String)
	if !ok {
		tb.Fatalf("expected String(%q), got %T (%s)", want, obj, obj.ToString(false))
	}
	if got.S != want {
		tb.Fatalf("expected String(%q), got String(%q)", want, got.S)
	}
}

func TestOptimizationRegressionArithmeticFastPaths(t *testing.T) {
	t.Run("int arithmetic", func(t *testing.T) {
		requireInt(t, evalTestScript(t, `(+ (* 6 7) (- 10 3))`), 49)
	})

	t.Run("mixed int double arithmetic", func(t *testing.T) {
		requireDouble(t, evalTestScript(t, `(+ (* 2.5 4) (- 7 2.5))`), 14.5)
	})

	t.Run("unary operations and predicates", func(t *testing.T) {
		requireBool(t, evalTestScript(t, `(zero? (- (inc 0) (dec 2)))`), true)
		requireInt(t, evalTestScript(t, `(- 7)`), -7)
	})

	t.Run("remainder", func(t *testing.T) {
		requireInt(t, evalTestScript(t, `(rem 29 5)`), 4)
	})
}

func TestOptimizationRegressionLoopAndBindingResolution(t *testing.T) {
	requireInt(t, evalTestScript(t, `(let [x 10]
  (loop [i 5 acc 0]
    (if (zero? i)
      (+ acc x)
      (recur (dec i) (+ acc x)))))`), 60)
}

func TestOptimizationRegressionRecursiveFib(t *testing.T) {
	requireInt(t, evalTestScript(t, `(letfn [(fib [n] (if (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))]
  (fib 10))`), 55)
}

func TestOptimizationRegressionSeqNthAndTryNth(t *testing.T) {
	requireString(t, evalTestScript(t, `(nth (seq ["a" "b" "c" "d"]) 2)`), "c")
	requireString(t, evalTestScript(t, `(nth (seq ["a" "b"]) 5 "missing")`), "missing")
}

func TestOptimizationRegressionArraySeqDirectIndexing(t *testing.T) {
	requireString(t, evalTestScript(t, `(let [s (seq ["zero" "one" "two" "three"])]
  (str (nth s 0) "/" (nth s 2)))`), "zero/two")
}

func TestOptimizationRegressionMapGetAssoc(t *testing.T) {
	requireInt(t, evalTestScript(t, `(let [m (assoc (assoc {} :a 1) :b 2)]
  (+ (get m :a 0) (get m :b 0) (get m :missing 3)))`), 6)
}
