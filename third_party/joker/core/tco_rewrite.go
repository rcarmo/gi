package core

// tco_rewrite.go — parse-time rewriting of tail-self-calls to recur.
//
// When a named fn (from letfn or named fn) has a tail-position call
// to itself, rewrite the fn body as a loop/recur. This eliminates
// the runtime trampoline overhead entirely.
//
// Before: (fn self [x] (if (= x 0) 1 (self (dec x))))
// After:  (fn self [x] (loop [x x] (if (= x 0) 1 (recur (dec x)))))
//
// The rewrite wraps the body in a LoopExpr with the fn args as bindings,
// and replaces tail self-calls with RecurExpr.

// rewriteTailCallsToRecur checks if a FnExpr with a self-binding
// has tail-position self-calls, and if so, rewrites them to recur.
func rewriteTailCallsToRecur(fnExpr *FnExpr, selfBinding *Binding) {
	if selfBinding == nil || fnExpr.self.name == nil {
		return
	}
	for i := range fnExpr.arities {
		arity := &fnExpr.arities[i]
		if len(arity.body) == 0 {
			continue
		}
		lastExpr := arity.body[len(arity.body)-1]
		has := hasTailSelfCall(lastExpr, selfBinding)
		if has {
			newBody := make([]Expr, len(arity.body))
			copy(newBody, arity.body)
			newBody[len(newBody)-1] = rewriteTailExpr(newBody[len(newBody)-1], selfBinding)
			arity.body = newBody
			fnExpr.tailRewritten = true
		}
	}
}

// hasTailSelfCall checks if an expression in tail position calls selfBinding.
func hasTailSelfCall(expr Expr, self *Binding) bool {
	switch e := expr.(type) {
	case *CallExpr:
		if bind, ok := e.callable.(*BindingExpr); ok {
			// The self-call may be through the letfn binding or the fn's own self binding.
			// Match by name since they may have different frame/index.
			if bind.binding.name.name != nil && self.name.name != nil &&
				*bind.binding.name.name == *self.name.name {
				return true
			}
		}
		return false
	case *IfExpr:
		return hasTailSelfCall(e.positive, self) || hasTailSelfCall(e.negative, self)
	case *DoExpr:
		if len(e.body) == 0 {
			return false
		}
		return hasTailSelfCall(e.body[len(e.body)-1], self)
	case *LetExpr:
		if len(e.body) == 0 {
			return false
		}
		return hasTailSelfCall(e.body[len(e.body)-1], self)
	default:
		return false
	}
}

// rewriteTailExpr replaces tail-position self-calls with RecurExpr.
func rewriteTailExpr(expr Expr, self *Binding) Expr {
	switch e := expr.(type) {
	case *CallExpr:
		if bind, ok := e.callable.(*BindingExpr); ok {
			if bind.binding.name.name != nil && self.name.name != nil &&
				*bind.binding.name.name == *self.name.name {
				return &RecurExpr{
					Position: e.Position,
					args:     e.args,
				}
			}
		}
		return e
	case *IfExpr:
		return &IfExpr{
			Position: e.Position,
			cond:     e.cond,
			positive: rewriteTailExpr(e.positive, self),
			negative: rewriteTailExpr(e.negative, self),
		}
	case *DoExpr:
		if len(e.body) == 0 {
			return e
		}
		newBody := make([]Expr, len(e.body))
		copy(newBody, e.body)
		newBody[len(newBody)-1] = rewriteTailExpr(newBody[len(newBody)-1], self)
		return &DoExpr{
			Position: e.Position,
			body:     newBody,
		}
	case *LetExpr:
		if len(e.body) == 0 {
			return e
		}
		newBody := make([]Expr, len(e.body))
		copy(newBody, e.body)
		newBody[len(newBody)-1] = rewriteTailExpr(newBody[len(newBody)-1], self)
		return &LetExpr{
			Position: e.Position,
			names:    e.names,
			values:   e.values,
			body:     newBody,
		}
	default:
		return expr
	}
}
