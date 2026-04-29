package core

import "fmt"

// ir_exported.go — temporary test helpers for IR debugging.

func IrCompileExported(loop *LoopExpr) *IRProgram { return irCompile(loop) }

func IrDebugCompile(loop *LoopExpr) {
	loopLet := (*LetExpr)(loop)
	fmt.Printf("loop body has %d exprs\n", len(loopLet.body))
	for i, expr := range loopLet.body {
		fmt.Printf("  body[%d]: %T\n", i, expr)
	}
	frame := guessLoopFrame(loopLet.body)
	fmt.Printf("guessed loop frame: %d\n", frame)
	c := &irCompiler{
		bindingMap: make(map[bindingKey]int),
		loopFrame:  frame,
	}
	for i := range loop.names {
		c.bindingMap[bindingKey{frame: frame, index: i}] = i
	}
	c.numSlots = len(loop.names)
	if frame < 0 {
		c.loopFrame = 1
	}
	for i, expr := range loopLet.body {
		if !c.compileExpr(expr, i == len(loopLet.body)-1) {
			fmt.Printf("  FAILED to compile body[%d]: %T\n", i, expr)
			break
		}
	}
}

func (p *IRProgram) CodeLen() int                       { return len(p.code) }
func (p *IRProgram) ConstLen() int                      { return len(p.constants) }
func (p *IRProgram) NumSlots() int                      { return p.numSlots }
func (p *IRProgram) CaptureKeys() []bindingKey          { return p.captureKeys }
func (e *LetExpr) Body() []Expr                         { return e.body }
func IrExecExported(prog *IRProgram, s []Object) Object { return irExec(prog, s) }
func (e *LetExpr) Values() []Expr                       { return e.values }

func TestInlineInfo(expr Expr) string {
	letExpr, ok := expr.(*LetExpr)
	if !ok {
		return "not a let"
	}
	loopExpr, ok := letExpr.Body()[0].(*LoopExpr)
	if !ok {
		return "body[0] not a loop"
	}
	loopLet := (*LetExpr)(loopExpr)
	// Find the if body
	ifExpr, ok := loopLet.body[0].(*IfExpr)
	if !ok {
		return "loop body not if"
	}
	// The negative branch should have the recur with fn call
	recurExpr, ok := ifExpr.negative.(*RecurExpr)
	if !ok {
		return fmt.Sprintf("negative not recur, is %T", ifExpr.negative)
	}
	// The second recur arg should be (+ s (A i i))
	addCall, ok := recurExpr.args[1].(*CallExpr)
	if !ok {
		return fmt.Sprintf("recur arg[1] not call, is %T", recurExpr.args[1])
	}
	// The second arg of + should be (A i i)
	fnCall, ok := addCall.args[1].(*CallExpr)
	if !ok {
		return fmt.Sprintf("add arg[1] not call, is %T", addCall.args[1])
	}
	bindExpr, ok := fnCall.callable.(*BindingExpr)
	if !ok {
		return fmt.Sprintf("fn callable not binding, is %T", fnCall.callable)
	}
	if bindExpr.binding.value == nil {
		return fmt.Sprintf("binding.value is nil (frame=%d, index=%d)", bindExpr.binding.frame, bindExpr.binding.index)
	}
	return fmt.Sprintf("binding.value type: %T", bindExpr.binding.value)
}

func (fn *Fn) FnExpr() *FnExpr        { return fn.fnExpr }
func (e *FnExpr) TailRewritten() bool { return e.tailRewritten }
func (e *FnExpr) Self() *string       { return e.self.name }
func (e *LoopExpr) Values() []Expr    { return (*LetExpr)(e).values }

func WasmCompileExported(prog *IRProgram) *WasmProgram        { return wasmCompile(prog) }
func WasmExecExported(wp *WasmProgram, slots []Object) Object { return wasmExec(wp, slots) }

func IsWasmEligibleExported(prog *IRProgram) bool { return isWasmEligible(prog) }
func IrToWasmExported(prog *IRProgram) []byte     { return irToWasm(prog) }
