package core

import (
	"math"
	"sync"
)

// ir.go — tiny lowered IR for hot loop/arithmetic subsets.
//
// The IR represents a small subset of Joker expressions as a flat
// instruction sequence with slot-resolved locals. It is interpreted
// by a simple switch loop that avoids the overhead of tree-walking
// Eval, interface dispatch, defer, and frame allocation.
//
// The IR is lowered lazily from LoopExpr bodies when all contained
// expressions fall within the supported subset. Compiled programs
// are cached so the compile cost is only paid once per loop site.

// Opcodes
const (
	irLiteral      byte = iota // operand: index into constants pool
	irLoadSlot                 // operand: slot index in locals
	irStoreSlot                // operand: slot index in locals
	irAdd                      // pop 2, push sum (Int fast path)
	irSub                      // pop 2, push difference
	irMul                      // pop 2, push product
	irRem                      // pop 2, push remainder
	irDiv                      // pop 2, push quotient (Double)
	irInc                      // pop 1, push +1
	irDec                      // pop 1, push -1
	irLt                       // pop 2, push Boolean
	irEq                       // pop 2, push Boolean
	irIsZero                   // pop 1, push Boolean
	irJumpIfNot                // operand: target PC (uint16 big-endian in next 2 bytes)
	irJump                     // operand: target PC
	irRecur                    // operand: nargs (2 bytes) + target PC (2 bytes)
	irReturn                   // pop 1, return it
	irGet                      // pop 2 (coll, key), push result or NIL
	irGet3                     // pop 3 (coll, key, default), push result
	irAssoc                    // pop 3 (coll, key, val), push new map
	irNth                      // pop 2 (coll, index), push element
	irConj                     // pop 2 (coll, val), push conj'd
	irSqrt                     // pop 1, push sqrt
	irCallSlot                 // operand1: slot (2 bytes), operand2: nargs (2 bytes)
	irCallSelf                 // operand: nargs (2 bytes)
	irFirst                    // pop 1, push first element
	irBuildVec                 // operand: n elements; pop n, push new vector
	irStr2                     // pop 2, push string concatenation
	irCount                    // pop 1, push count
	irToTransient              // pop 1 (ArrayVector), push TransientVector
	irAssocBang                // pop 3 (tv, key, val), mutate in place, push tv
	irToPersistent             // pop 1 (TransientVector), push ArrayVector
	irFallback                 // cannot execute in IR; fall back to tree Eval
)

// ---------- Cache ----------

var irCache sync.Map   // map[*LoopExpr]*IRProgram
var irFnCache sync.Map // map[*FnArityExpr]*IRProgram

var irCompileFailed = &IRProgram{} // sentinel

func irGetCached(loop *LoopExpr) *IRProgram {
	if cached, ok := irCache.Load(loop); ok {
		prog := cached.(*IRProgram)
		if prog == irCompileFailed {
			return nil
		}
		return prog
	}
	prog := irCompile(loop)
	if prog == nil {
		irCache.Store(loop, irCompileFailed)
		return nil
	}
	irCache.Store(loop, prog)
	return prog
}

// ---------- Program ----------

type IRProgram struct {
	code        []byte
	constants   []Object
	numSlots    int
	captureKeys []bindingKey
	hasSelf     bool
}

// ---------- Fn compilation ----------

// irCompileFn attempts to compile a single-arity Fn body into an IRProgram.
// The args become slots 0..n-1. Returns nil if the body can't be compiled.
// selfBinding optionally identifies the binding key for self-recursive calls.
func irCompileFn(fn *Fn) *IRProgram {
	if len(fn.fnExpr.arities) != 1 {
		return nil
	}
	arity := fn.fnExpr.arities[0]

	if cached, ok := irFnCache.Load(&arity); ok {
		prog := cached.(*IRProgram)
		if prog == irCompileFailed {
			return nil
		}
		return prog
	}

	c := &irCompiler{
		bindingMap: make(map[bindingKey]int),
		loopFrame:  -1,
	}
	c.numSlots = len(arity.args)

	// Determine the frame from the body
	fnFrame := guessLoopFrame(arity.body)
	if fnFrame < 0 {
		fnFrame = 1
	}
	c.loopFrame = fnFrame
	for i := range arity.args {
		c.bindingMap[bindingKey{frame: fnFrame, index: i}] = i
	}

	// If the fn has a self-binding, mark it for self-recursive IR dispatch
	if fn.fnExpr.self.name != nil {
		// The self-binding is typically at frame fnFrame-1, index 0
		// (the letfn/fn frame that holds the fn itself)
		c.selfSlot = 0 // will use special irCallSelf opcode
		c.hasSelf = true
		c.selfNArgs = len(arity.args)
	}

	// Compile body
	for i, expr := range arity.body {
		if !c.compileExpr(expr, i == len(arity.body)-1) {
			irFnCache.Store(&arity, irCompileFailed)
			return nil
		}
	}
	if len(c.code) == 0 {
		irFnCache.Store(&arity, irCompileFailed)
		return nil
	}
	if c.code[len(c.code)-1] != irReturn {
		c.emit(irReturn)
	}
	// Allow self-captures but no other captures
	if len(c.captureKeys) > 0 {
		irFnCache.Store(&arity, irCompileFailed)
		return nil
	}
	prog := &IRProgram{
		code:      c.code,
		constants: c.constants,
		numSlots:  c.numSlots,
		hasSelf:   c.hasSelf,
	}
	irFnCache.Store(&arity, prog)
	return prog
}

// ---------- Compiler ----------

type bindingKey struct {
	frame int
	index int
}

type irCompiler struct {
	code         []byte
	constants    []Object
	bindingMap   map[bindingKey]int
	captureKeys  []bindingKey
	numSlots     int
	loopFrame    int
	depth        int
	hasSelf      bool
	selfSlot     int
	selfNArgs    int
	recurTargets []recurTarget // stack of recur targets for nested loops
}

type recurTarget struct {
	pc       int // bytecode offset of loop start
	baseSlot int // first slot of this loop's bindings
	nBinds   int // number of loop bindings
}

func irCompile(loop *LoopExpr) *IRProgram {
	c := &irCompiler{
		bindingMap: make(map[bindingKey]int),
		loopFrame:  -1,
	}
	c.numSlots = len(loop.names)

	loopLet := (*LetExpr)(loop)
	c.loopFrame = guessLoopFrame(loopLet.body)
	if c.loopFrame < 0 {
		c.loopFrame = 1
	}
	for i := range loop.names {
		c.bindingMap[bindingKey{frame: c.loopFrame, index: i}] = i
	}

	// Push the top-level recur target (PC=0, slots 0..n-1)
	c.recurTargets = []recurTarget{{pc: 0, baseSlot: 0, nBinds: len(loop.names)}}

	for i, expr := range loopLet.body {
		if !c.compileExpr(expr, i == len(loopLet.body)-1) {
			return nil
		}
	}
	if len(c.code) == 0 {
		return nil
	}
	if c.code[len(c.code)-1] != irReturn && c.code[len(c.code)-1] != irJump {
		c.emit(irReturn)
	}
	// Safety limit: too many captures indicates complex nested scoping
	if len(c.captureKeys) > 12 {
		return nil
	}
	// Validate: ensure no slot is assigned twice
	slotUsed := make(map[int]bool, c.numSlots)
	for _, slot := range c.bindingMap {
		if slotUsed[slot] {
			return nil // slot collision detected
		}
		slotUsed[slot] = true
	}
	return &IRProgram{
		code:        c.code,
		constants:   c.constants,
		numSlots:    c.numSlots,
		captureKeys: c.captureKeys,
	}
}

func guessLoopFrame(body []Expr) int {
	for _, expr := range body {
		if f := findBindingFrame(expr); f >= 0 {
			return f
		}
	}
	return -1
}

func findBindingFrame(expr Expr) int {
	switch e := expr.(type) {
	case *BindingExpr:
		return e.binding.frame
	case *IfExpr:
		if f := findBindingFrame(e.cond); f >= 0 {
			return f
		}
		if f := findBindingFrame(e.positive); f >= 0 {
			return f
		}
		return findBindingFrame(e.negative)
	case *CallExpr:
		for _, arg := range e.args {
			if f := findBindingFrame(arg); f >= 0 {
				return f
			}
		}
	case *RecurExpr:
		for _, arg := range e.args {
			if f := findBindingFrame(arg); f >= 0 {
				return f
			}
		}
	case *LetExpr:
		for _, v := range e.values {
			if f := findBindingFrame(v); f >= 0 {
				return f
			}
		}
	}
	return -1
}

func (c *irCompiler) emit(op byte) {
	c.code = append(c.code, op)
}

func (c *irCompiler) emitWithOperand(op byte, operand int) {
	c.code = append(c.code, op, byte(operand>>8), byte(operand))
}

func (c *irCompiler) patchJump(pos int, target int) {
	c.code[pos+1] = byte(target >> 8)
	c.code[pos+2] = byte(target)
}

func (c *irCompiler) addConstant(obj Object) int {
	for i, existing := range c.constants {
		if existing.Equals(obj) {
			return i
		}
	}
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *irCompiler) compileExpr(expr Expr, isLast bool) bool {
	switch e := expr.(type) {
	case *LiteralExpr:
		idx := c.addConstant(e.obj)
		c.emitWithOperand(irLiteral, idx)
		if isLast {
			c.emit(irReturn)
		}
		return true

	case *VectorExpr:
		// Try constant vector first (all elements are literals)
		allLiteral := true
		for _, elem := range e.v {
			if _, ok := elem.(*LiteralExpr); !ok {
				allLiteral = false
				break
			}
		}
		if allLiteral {
			arr := make([]Object, len(e.v))
			for i, elem := range e.v {
				arr[i] = elem.(*LiteralExpr).obj
			}
			idx := c.addConstant(&ArrayVector{arr: arr})
			c.emitWithOperand(irLiteral, idx)
		} else {
			// Compile each element, then emit a vector-build opcode
			for _, elem := range e.v {
				if !c.compileExpr(elem, false) {
					return false
				}
			}
			c.emitWithOperand(irBuildVec, len(e.v))
		}
		if isLast {
			c.emit(irReturn)
		}
		return true

	case *BindingExpr:
		key := bindingKey{frame: e.binding.frame, index: e.binding.index}
		slot, ok := c.bindingMap[key]
		if !ok {
			if e.binding.frame < c.loopFrame {
				slot = c.numSlots
				c.bindingMap[key] = slot
				c.captureKeys = append(c.captureKeys, key)
				c.numSlots++
			} else {
				return false
			}
		}
		c.emitWithOperand(irLoadSlot, slot)
		if isLast {
			c.emit(irReturn)
		}
		return true

	case *IfExpr:
		if !c.compileExpr(e.cond, false) {
			return false
		}
		jumpPos := len(c.code)
		c.emitWithOperand(irJumpIfNot, 0)
		if !c.compileExpr(e.positive, isLast) {
			return false
		}
		if !isLast {
			skipPos := len(c.code)
			c.emitWithOperand(irJump, 0)
			c.patchJump(jumpPos, len(c.code))
			if !c.compileExpr(e.negative, isLast) {
				return false
			}
			c.patchJump(skipPos, len(c.code))
		} else {
			c.patchJump(jumpPos, len(c.code))
			if !c.compileExpr(e.negative, isLast) {
				return false
			}
		}
		return true

	case *CallExpr:
		return c.compileCall(e, isLast)

	case *RecurExpr:
		if len(c.recurTargets) == 0 {
			return false
		}
		target := c.recurTargets[len(c.recurTargets)-1]
		for _, arg := range e.args {
			if !c.compileExpr(arg, false) {
				return false
			}
		}
		// Emit: nargs (2) + targetPC (2) [+ baseSlot (2) if nested]
		c.code = append(c.code, irRecur,
			byte(len(e.args)>>8), byte(len(e.args)),
			byte(target.pc>>8), byte(target.pc))
		if target.pc != 0 {
			// Nested loop: also emit baseSlot
			c.code = append(c.code, byte(target.baseSlot>>8), byte(target.baseSlot))
		}
		return true

	case *LetExpr:
		if c.depth > 4 {
			return false
		}
		c.depth++
		return c.compileLetBody(e, isLast)

	case *LoopExpr:
		if c.depth > 4 {
			return false
		}
		c.depth++
		return c.compileNestedLoop(e, isLast)

	default:
		return false
	}
}

func (c *irCompiler) compileLetBody(e *LetExpr, isLast bool) bool {
	baseSlot := c.numSlots
	letFrame := -1
	for _, bodyExpr := range e.body {
		if f := findBindingFrame(bodyExpr); f > c.loopFrame {
			letFrame = f
			break
		}
	}
	if letFrame < 0 {
		letFrame = c.loopFrame + c.depth
	}
	for i, bindExpr := range e.values {
		if !c.compileExpr(bindExpr, false) {
			return false
		}
		slot := baseSlot + i
		c.bindingMap[bindingKey{frame: letFrame, index: i}] = slot
		c.emitWithOperand(irStoreSlot, slot)
	}
	c.numSlots = baseSlot + len(e.values)
	for i, bodyExpr := range e.body {
		if !c.compileExpr(bodyExpr, isLast && i == len(e.body)-1) {
			return false
		}
	}
	return true
}

func (c *irCompiler) compileNestedLoop(loop *LoopExpr, isLast bool) bool {
	loopLet := (*LetExpr)(loop)
	baseSlot := c.numSlots

	loopFrame := -1
	for _, bodyExpr := range loopLet.body {
		if f := findBindingFrame(bodyExpr); f > c.loopFrame {
			loopFrame = f
			break
		}
	}
	if loopFrame < 0 {
		loopFrame = c.loopFrame + c.depth
	}

	for i, bindExpr := range loopLet.values {
		if !c.compileExpr(bindExpr, false) {
			return false
		}
		slot := baseSlot + i
		c.bindingMap[bindingKey{frame: loopFrame, index: i}] = slot
		c.emitWithOperand(irStoreSlot, slot)
	}
	c.numSlots = baseSlot + len(loopLet.names)

	loopStartPC := len(c.code)
	c.recurTargets = append(c.recurTargets, recurTarget{
		pc:       loopStartPC,
		baseSlot: baseSlot,
		nBinds:   len(loopLet.names),
	})

	for i, expr := range loopLet.body {
		if !c.compileExpr(expr, isLast && i == len(loopLet.body)-1) {
			c.recurTargets = c.recurTargets[:len(c.recurTargets)-1]
			return false
		}
	}

	c.recurTargets = c.recurTargets[:len(c.recurTargets)-1]
	return true
}

func (c *irCompiler) tryInlineCall(fnSlot int, expr *CallExpr, isLast bool) bool {
	// Inlining disabled for now — needs smarter slot reuse to avoid
	// bloating slot counts in compiled fn programs.
	return false
}

// findFnExprForBinding tries to find the FnExpr for a callable binding.
func findFnExprForBinding(callable Expr) *FnExpr {
	bindExpr, ok := callable.(*BindingExpr)
	if !ok {
		return nil
	}
	if bindExpr.binding.value == nil {
		return nil
	}
	if fnExpr, ok := bindExpr.binding.value.(*FnExpr); ok {
		return fnExpr
	}
	return nil
}

func (c *irCompiler) compileCall(expr *CallExpr, isLast bool) bool {
	// Check if callable is a binding (local/captured function)
	if bindExpr, ok := expr.callable.(*BindingExpr); ok {
		// Check for self-recursive call
		if c.hasSelf && bindExpr.binding.frame < c.loopFrame && len(expr.args) == c.selfNArgs {
			for _, arg := range expr.args {
				if !c.compileExpr(arg, false) {
					return false
				}
			}
			c.emitWithOperand(irCallSelf, len(expr.args))
			if isLast {
				c.emit(irReturn)
			}
			return true
		}

		key := bindingKey{frame: bindExpr.binding.frame, index: bindExpr.binding.index}
		slot, ok := c.bindingMap[key]
		if !ok {
			if bindExpr.binding.frame < c.loopFrame {
				slot = c.numSlots
				c.bindingMap[key] = slot
				c.captureKeys = append(c.captureKeys, key)
				c.numSlots++
			} else {
				return false
			}
		}

		// Try to inline the function call
		if c.tryInlineCall(slot, expr, isLast) {
			return true
		}

		for _, arg := range expr.args {
			if !c.compileExpr(arg, false) {
				return false
			}
		}
		c.code = append(c.code, irCallSlot,
			byte(slot>>8), byte(slot),
			byte(len(expr.args)>>8), byte(len(expr.args)))
		if isLast {
			c.emit(irReturn)
		}
		return true
	}

	vref, ok := expr.callable.(*VarRefExpr)
	if !ok {
		return false
	}
	procName := ""
	switch v := vref.vr.Value.(type) {
	case Proc:
		procName = v.Name
	case *Fn:
		procName = coreVarToProcName(vref.vr)
	}
	if procName == "" {
		return false
	}

	switch procName {
	case "procAdd":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irAdd)
	case "procSubtract":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irSub)
	case "procMultiply":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irMul)
	case "procRem":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irRem)
	case "procDivide":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irDiv)
	case "procInc":
		if len(expr.args) != 1 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) {
			return false
		}
		c.emit(irInc)
	case "procDec":
		if len(expr.args) != 1 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) {
			return false
		}
		c.emit(irDec)
	case "procLt":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irLt)
	case "procEq":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irEq)
	case "procIsZero":
		if len(expr.args) != 1 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) {
			return false
		}
		c.emit(irIsZero)
	case "procGet":
		if len(expr.args) == 2 {
			if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
				return false
			}
			c.emit(irGet)
		} else if len(expr.args) == 3 {
			if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) || !c.compileExpr(expr.args[2], false) {
				return false
			}
			c.emit(irGet3)
		} else {
			return false
		}
	case "procAssoc":
		if len(expr.args) != 3 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) || !c.compileExpr(expr.args[2], false) {
			return false
		}
		c.emit(irAssoc)
	case "procNth":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irNth)
	case "procConj":
		if len(expr.args) != 2 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
			return false
		}
		c.emit(irConj)
	case "procSqrt":
		if len(expr.args) != 1 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) {
			return false
		}
		c.emit(irSqrt)
	case "procFirst":
		if len(expr.args) != 1 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) {
			return false
		}
		c.emit(irFirst)
	case "procStr":
		if len(expr.args) == 2 {
			if !c.compileExpr(expr.args[0], false) || !c.compileExpr(expr.args[1], false) {
				return false
			}
			c.emit(irStr2)
		} else {
			return false
		}
	case "procCount":
		if len(expr.args) != 1 {
			return false
		}
		if !c.compileExpr(expr.args[0], false) {
			return false
		}
		c.emit(irCount)
	default:
		return false
	}
	if isLast {
		c.emit(irReturn)
	}
	return true
}

// coreVarToProcName maps well-known core Vars to internal proc names.
func coreVarToProcName(vr *Var) string {
	if vr.ns == nil || vr.ns != GLOBAL_ENV.CoreNamespace {
		return ""
	}
	switch vr.name.ToString(false) {
	case "+":
		return "procAdd"
	case "-":
		return "procSubtract"
	case "*":
		return "procMultiply"
	case "rem":
		return "procRem"
	case "inc":
		return "procInc"
	case "dec":
		return "procDec"
	case "<":
		return "procLt"
	case "=":
		return "procEq"
	case "zero?":
		return "procIsZero"
	case "/":
		return "procDivide"
	case "get":
		return "procGet"
	case "assoc":
		return "procAssoc"
	case "conj":
		return "procConj"
	case "sqrt":
		return "procSqrt"
	case "first":
		return "procFirst"
	case "str":
		return "procStr"
	case "count":
		return "procCount"
	case "nth":
		return "procNth"
	default:
		return ""
	}
}

// ---------- Interpreter ----------

func irExec(prog *IRProgram, initSlots []Object) Object {
	var slots []Object
	if prog.numSlots <= 16 {
		var buf [16]Object
		slots = buf[:prog.numSlots]
	} else {
		slots = make([]Object, prog.numSlots)
	}
	copy(slots, initSlots)

	// Convert ArrayVector slots to transient for in-place mutation
	for i, s := range slots {
		if av, ok := s.(*ArrayVector); ok {
			slots[i] = ToTransient(av)
		}
	}

	var stack []Object
	var stackBuf [16]Object
	stack = stackBuf[:0]
	code := prog.code
	pc := 0

loop:
	for pc < len(code) {
		op := code[pc]
		pc++
		switch op {
		case irLiteral:
			idx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			stack = append(stack, prog.constants[idx])

		case irLoadSlot:
			idx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			stack = append(stack, slots[idx])

		case irStoreSlot:
			idx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			slots[idx] = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

		case irAdd:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch av := a.(type) {
			case Int:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Int{I: av.I + bv.I})
					continue
				case Double:
					stack = append(stack, Double{D: float64(av.I) + bv.D})
					continue
				}
			case Double:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Double{D: av.D + float64(bv.I)})
					continue
				case Double:
					stack = append(stack, Double{D: av.D + bv.D})
					continue
				}
			}
			return nil

		case irSub:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch av := a.(type) {
			case Int:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Int{I: av.I - bv.I})
					continue
				case Double:
					stack = append(stack, Double{D: float64(av.I) - bv.D})
					continue
				}
			case Double:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Double{D: av.D - float64(bv.I)})
					continue
				case Double:
					stack = append(stack, Double{D: av.D - bv.D})
					continue
				}
			}
			return nil

		case irMul:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch av := a.(type) {
			case Int:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Int{I: av.I * bv.I})
					continue
				case Double:
					stack = append(stack, Double{D: float64(av.I) * bv.D})
					continue
				}
			case Double:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Double{D: av.D * float64(bv.I)})
					continue
				case Double:
					stack = append(stack, Double{D: av.D * bv.D})
					continue
				}
			}
			return nil

		case irRem:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			if ai, aok := a.(Int); aok {
				if bi, bok := b.(Int); bok {
					if bi.I == 0 {
						return nil
					}
					stack = append(stack, Int{I: ai.I % bi.I})
					continue
				}
			}
			return nil

		case irDiv:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			var av, bv float64
			switch x := a.(type) {
			case Int:
				av = float64(x.I)
			case Double:
				av = x.D
			default:
				return nil
			}
			switch x := b.(type) {
			case Int:
				bv = float64(x.I)
			case Double:
				bv = x.D
			default:
				return nil
			}
			if bv == 0 {
				return nil
			}
			stack = append(stack, Double{D: av / bv})

		case irInc:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch av := a.(type) {
			case Int:
				stack = append(stack, Int{I: av.I + 1})
			case Double:
				stack = append(stack, Double{D: av.D + 1})
			default:
				return nil
			}

		case irDec:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch av := a.(type) {
			case Int:
				stack = append(stack, Int{I: av.I - 1})
			case Double:
				stack = append(stack, Double{D: av.D - 1})
			default:
				return nil
			}

		case irLt:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch av := a.(type) {
			case Int:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Boolean{B: av.I < bv.I})
					continue
				case Double:
					stack = append(stack, Boolean{B: float64(av.I) < bv.D})
					continue
				}
			case Double:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Boolean{B: av.D < float64(bv.I)})
					continue
				case Double:
					stack = append(stack, Boolean{B: av.D < bv.D})
					continue
				}
			}
			return nil

		case irEq:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch av := a.(type) {
			case Int:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Boolean{B: av.I == bv.I})
					continue
				case Double:
					stack = append(stack, Boolean{B: float64(av.I) == bv.D})
					continue
				}
			case Double:
				switch bv := b.(type) {
				case Int:
					stack = append(stack, Boolean{B: av.D == float64(bv.I)})
					continue
				case Double:
					stack = append(stack, Boolean{B: av.D == bv.D})
					continue
				}
			}
			return nil

		case irIsZero:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch av := a.(type) {
			case Int:
				stack = append(stack, Boolean{B: av.I == 0})
			case Double:
				stack = append(stack, Boolean{B: av.D == 0})
			default:
				return nil
			}

		case irJumpIfNot:
			target := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			val := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch v := val.(type) {
			case Nil:
				pc = target
			case Boolean:
				if !v.B {
					pc = target
				}
			}

		case irJump:
			target := int(code[pc])<<8 | int(code[pc+1])
			pc = target

		case irRecur:
			n := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			targetPC := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			// For nested loops, the recur target baseSlot might not be 0
			// We need to figure out which slots to write to.
			// Convention: recur writes to slots starting from the baseSlot
			// of the loop that recur targets. For the top-level loop, baseSlot=0.
			// For nested loops, we determine baseSlot from the target PC.
			// Simple approach: if targetPC==0, write to slots 0..n-1 (backward compat).
			// Otherwise, we need the baseSlot encoded somewhere.
			// For now, recur always writes to the slots at the end of stack in order.
			if targetPC == 0 {
				// Top-level loop recur
				for i := n - 1; i >= 0; i-- {
					slots[i] = stack[len(stack)-1]
					stack = stack[:len(stack)-1]
				}
			} else {
				// Nested loop recur — find the base slot from the compile info
				// We store base slot in the bytecode too: nargs(2) + targetPC(2) + baseSlot(2)
				// ... but we didn't emit baseSlot yet. Let's add it.
				// For now, infer: the slots for this loop start at (numSlots - n) or
				// we need to extend the encoding.
				// Quick fix: also encode baseSlot
				baseSlot := int(code[pc])<<8 | int(code[pc+1])
				pc += 2
				for i := n - 1; i >= 0; i-- {
					slots[baseSlot+i] = stack[len(stack)-1]
					stack = stack[:len(stack)-1]
				}
			}
			pc = targetPC
			stack = stack[:0]
			goto loop

		case irReturn:
			if len(stack) == 0 {
				return NIL
			}
			result := stack[len(stack)-1]
			// Freeze any transient vectors before returning
			if tv, ok := result.(*TransientVector); ok {
				return tv.ToPersistent()
			}
			return result

		case irGet:
			key := stack[len(stack)-1]
			coll := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			if g, ok := coll.(Gettable); ok {
				ok, v := g.Get(key)
				if ok {
					stack = append(stack, v)
				} else {
					stack = append(stack, NIL)
				}
			} else {
				return nil
			}

		case irGet3:
			def := stack[len(stack)-1]
			key := stack[len(stack)-2]
			coll := stack[len(stack)-3]
			stack = stack[:len(stack)-3]
			if g, ok := coll.(Gettable); ok {
				ok, v := g.Get(key)
				if ok {
					stack = append(stack, v)
				} else {
					stack = append(stack, def)
				}
			} else {
				stack = append(stack, def)
			}

		case irAssoc:
			val := stack[len(stack)-1]
			key := stack[len(stack)-2]
			coll := stack[len(stack)-3]
			stack = stack[:len(stack)-3]
			// Fast path: transient vector in-place mutation
			if tv, ok := coll.(*TransientVector); ok {
				stack = append(stack, tv.AssocInPlace(key, val))
			} else if a, ok := coll.(Associative); ok {
				stack = append(stack, a.Assoc(key, val))
			} else {
				return nil
			}

		case irNth:
			idxObj := stack[len(stack)-1]
			coll := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			idx, iok := idxObj.(Int)
			if !iok {
				return nil
			}
			switch c := coll.(type) {
			case *ArrayVector:
				if idx.I >= 0 && idx.I < len(c.arr) {
					stack = append(stack, c.arr[idx.I])
				} else {
					return nil
				}
			case *TransientVector:
				stack = append(stack, c.Nth(idx.I))
			case Indexed:
				stack = append(stack, c.Nth(idx.I))
			default:
				return nil
			}

		case irConj:
			val := stack[len(stack)-1]
			coll := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch c := coll.(type) {
			case Conjable:
				stack = append(stack, c.Conj(val))
			default:
				return nil
			}

		case irSqrt:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch av := a.(type) {
			case Double:
				stack = append(stack, Double{D: math.Sqrt(av.D)})
			case Int:
				stack = append(stack, Double{D: math.Sqrt(float64(av.I))})
			default:
				return nil
			}

		case irCallSlot:
			slotIdx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			nargs := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			fnObj := slots[slotIdx]
			var args []Object
			if nargs > 0 {
				args = make([]Object, nargs)
				for i := nargs - 1; i >= 0; i-- {
					args[i] = stack[len(stack)-1]
					stack = stack[:len(stack)-1]
				}
			}
			// Try compiled fn dispatch first
			if fn, ok := fnObj.(*Fn); ok {
				if fnProg := irCompileFn(fn); fnProg != nil {
					if result := irExec(fnProg, args); result != nil {
						stack = append(stack, result)
						continue
					}
				}
			}
			// Fallback to normal Fn.Call
			prevExpr := RT.currentExpr
			RT.currentExpr = &CallExpr{}
			switch fn := fnObj.(type) {
			case Callable:
				stack = append(stack, fn.Call(args))
			default:
				RT.currentExpr = prevExpr
				return nil
			}
			RT.currentExpr = prevExpr

		case irCallSelf:
			nargs := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			args := make([]Object, nargs)
			for i := nargs - 1; i >= 0; i-- {
				args[i] = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
			result := irExec(prog, args)
			if result == nil {
				return nil
			}
			stack = append(stack, result)

		case irFirst:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch v := a.(type) {
			case *ArrayVector:
				if len(v.arr) > 0 {
					stack = append(stack, v.arr[0])
				} else {
					stack = append(stack, NIL)
				}
			case Seqable:
				s := v.Seq()
				if s.IsEmpty() {
					stack = append(stack, NIL)
				} else {
					stack = append(stack, s.First())
				}
			default:
				return nil
			}

		case irBuildVec:
			n := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			arr := make([]Object, n)
			for i := n - 1; i >= 0; i-- {
				arr[i] = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, &ArrayVector{arr: arr})

		case irStr2:
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, String{S: a.ToString(false) + b.ToString(false)})

		case irCount:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch v := a.(type) {
			case Counted:
				stack = append(stack, Int{I: v.Count()})
			case String:
				stack = append(stack, Int{I: len(v.S)})
			default:
				return nil
			}

		case irToTransient:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if av, ok := a.(*ArrayVector); ok {
				stack = append(stack, ToTransient(av))
			} else {
				return nil
			}

		case irAssocBang:
			val := stack[len(stack)-1]
			key := stack[len(stack)-2]
			coll := stack[len(stack)-3]
			stack = stack[:len(stack)-3]
			if tv, ok := coll.(*TransientVector); ok {
				stack = append(stack, tv.AssocInPlace(key, val))
			} else {
				return nil
			}

		case irToPersistent:
			a := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if tv, ok := a.(*TransientVector); ok {
				stack = append(stack, tv.ToPersistent())
			} else {
				return nil
			}
		case irFallback:
			return nil

		default:
			return nil
		}
	}
	if len(stack) > 0 {
		return stack[len(stack)-1]
	}
	return NIL
}
