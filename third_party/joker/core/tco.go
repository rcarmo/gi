package core

// tco.go — generic tail-call optimization via trampoline.
//
// When a function body's tail expression is a call to the same function,
// it returns a TailCall marker instead of actually recursing. Fn.Call
// detects this and loops with the new args, eliminating stack growth.
//
// This benefits any self-recursive function where the self-call is in
// tail position (e.g. accumulators, state machines, list traversals).
// It does NOT help tree-recursive patterns like naive fib where the
// recursive calls are not in tail position.

// TailCall is a marker returned by evalTailCall when a self-call in
// tail position is detected. It is NOT a valid Joker Object — it is
// only used internally between evalLoop and Fn.Call.
type TailCall struct {
	fn   *Fn
	args []Object
}

// Object interface stubs — TailCall should never escape to user code.
func (tc *TailCall) ToString(escape bool) string   { return "#<tail-call>" }
func (tc *TailCall) Equals(other interface{}) bool { return false }
func (tc *TailCall) GetInfo() *ObjectInfo          { return nil }
func (tc *TailCall) WithInfo(*ObjectInfo) Object   { return tc }
func (tc *TailCall) GetType() *Type                { return TYPE.Fn }
func (tc *TailCall) Hash() uint32                  { return 0 }

// activeFn tracks the currently executing Fn for TCO detection.
// This is stored on the Runtime (single-threaded evaluator).
var activeFn *Fn

// evalBodyTCO evaluates a body and, for the last expression, checks
// if it's a self-call in tail position. If so, returns a *TailCall
// instead of actually calling.
func evalBodyTCO(body []Expr, env *LocalEnv, self *Fn) Object {
	if len(body) == 0 {
		return NIL
	}
	// Evaluate all but the last expression normally
	for i := 0; i < len(body)-1; i++ {
		Eval(body[i], env)
	}
	// For the last expression, check for tail self-call
	last := body[len(body)-1]
	return evalTailExpr(last, env, self)
}

// evalLoopTCO is like evalLoop but with TCO awareness.
func evalLoopTCO(body []Expr, env *LocalEnv, self *Fn) Object {
	var res Object = NIL
loop:
	for _, expr := range body {
		res = Eval(expr, env)
	}
	switch res := res.(type) {
	default:
		return res
	case RecurBindings:
		env.bindings = res
		goto loop
	}
}

// evalTailExpr evaluates an expression in tail position with self-call detection.
func evalTailExpr(expr Expr, env *LocalEnv, self *Fn) Object {
	switch e := expr.(type) {
	case *IfExpr:
		if ToBool(Eval(e.cond, env)) {
			return evalTailExpr(e.positive, env, self)
		}
		return evalTailExpr(e.negative, env, self)

	case *CallExpr:
		// Check if this is a self-call
		callable := Eval(e.callable, env)
		if fn, ok := callable.(*Fn); ok && fn == self {
			// This is a tail self-call — return TailCall marker
			args := evalSeq(e.args, env)
			return &TailCall{fn: fn, args: args}
		}
		// Not a self-call — evaluate normally
		switch c := callable.(type) {
		case Callable:
			args := evalSeq(e.args, env)
			return c.Call(args)
		default:
			panic(RT.NewErrorWithPos(callable.ToString(false)+" is not a Fn", e.callable.Pos()))
		}

	case *DoExpr:
		return evalBodyTCO(e.body, env, self)

	case *LetExpr:
		childEnv := LocalEnv{bindings: make([]Object, 0, len(e.names)), parent: env}
		if env != nil {
			childEnv.frame = env.frame + 1
		}
		for _, bindingExpr := range e.values {
			childEnv.bindings = append(childEnv.bindings, Eval(bindingExpr, &childEnv))
		}
		return evalBodyTCO(e.body, &childEnv, self)

	default:
		// Not a recognized tail form — evaluate normally
		return Eval(expr, env)
	}
}
