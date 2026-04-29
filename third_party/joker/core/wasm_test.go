package core

import "testing"

func TestWasmArithmeticLoopCorrectness(t *testing.T) {
	expr := compileBenchExpr(t, `(loop [i 0 s 0]
  (if (= i 100) s (recur (inc i) (+ s (rem (* i 7) 11)))))`)
	// Get expected result from IR
	expected := Eval(expr, nil)

	// Try WASM compilation
	loop := expr.(*LoopExpr)
	prog := irCompile(loop)
	if prog == nil {
		t.Skip("IR compilation failed")
	}
	if !isWasmEligible(prog) {
		t.Skip("IR not WASM-eligible")
	}
	wp := wasmCompile(prog)
	if wp == nil {
		t.Skip("WASM compilation failed")
	}
	result := wasmExec(wp, []Object{Int{I: 0}, Int{I: 0}})
	if result == nil {
		t.Fatal("WASM execution returned nil")
	}
	if !result.Equals(expected) {
		t.Fatalf("WASM result %s != IR result %s", result.ToString(false), expected.ToString(false))
	}
}

func TestWasmSimpleLoop(t *testing.T) {
	expr := compileBenchExpr(t, `(loop [i 0 s 0]
  (if (= i 10) s (recur (+ i 1) (+ s i))))`)
	expected := Eval(expr, nil)

	loop := expr.(*LoopExpr)
	prog := irCompile(loop)
	if prog == nil {
		t.Skip("IR failed")
	}
	wp := wasmCompile(prog)
	if wp == nil {
		t.Skip("WASM failed")
	}
	result := wasmExec(wp, []Object{Int{I: 0}, Int{I: 0}})
	if result == nil {
		t.Fatal("nil result")
	}
	if !result.Equals(expected) {
		t.Fatalf("got %s, want %s", result.ToString(false), expected.ToString(false))
	}
}

func BenchmarkWasmArithmeticLoop(b *testing.B) {
	expr := compileBenchExpr(b, `(loop [i 0 s 0]
  (if (= i 100000) s (recur (inc i) (+ s (rem (* i 7) 11)))))`)
	loop := expr.(*LoopExpr)
	prog := irCompile(loop)
	if prog == nil {
		b.Skip("IR failed")
	}
	wp := wasmCompile(prog)
	if wp == nil {
		b.Skip("WASM failed")
	}
	slots := []Object{Int{I: 0}, Int{I: 0}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wasmExec(wp, slots)
	}
}
