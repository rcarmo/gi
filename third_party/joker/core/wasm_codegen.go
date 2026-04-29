package core

// wasm_codegen.go — translates IR bytecode to WASM function body.

func irToWasm(prog *IRProgram) []byte {
	if !isWasmEligible(prog) {
		return nil
	}
	body := compileWasmBody(prog)
	if body == nil {
		return nil
	}
	m := newWasmModule()
	m.addTypeSection(prog.numSlots)
	m.addFuncSection()
	m.addExportSection()
	m.addCodeSection(body)
	return m.bytes()
}

func isWasmEligible(prog *IRProgram) bool {
	code := prog.code
	pc := 0
	for pc < len(code) {
		op := code[pc]
		pc++
		switch op {
		case irLiteral, irLoadSlot, irStoreSlot:
			pc += 2
		case irAdd, irSub, irMul, irDiv, irRem, irInc, irDec,
			irLt, irEq, irIsZero, irReturn, irSqrt:
			// ok
		case irJumpIfNot, irJump:
			pc += 2
		case irRecur:
			pc += 4
			tgt := int(code[pc-2])<<8 | int(code[pc-1])
			if tgt != 0 {
				pc += 2
				return false
			}
		default:
			return false
		}
	}
	return true
}

// compileWasmBody generates WASM instructions.
//
// Layout:
//
//	block $exit (result i64)     ;; depth from inside if: +2
//	  loop $loop (void)          ;; depth from inside if: +1
//	    ;; body
//	    ;; irReturn → br $exit (depth = nesting + 1)
//	    ;; irRecur  → set locals, br $loop (depth = nesting)
//	  end
//	  i64.const 0  ;; unreachable
//	end
//
// For if/else: both branches end with `br` (stack-polymorphic),
// so `if void` works and no values need to flow through the if block.
func compileWasmBody(prog *IRProgram) []byte {
	var o []byte
	o = append(o, 0x00) // 0 local decls

	o = append(o, 0x02, 0x7e) // block $exit -> i64
	o = append(o, 0x03, 0x40) // loop $loop -> void

	code := prog.code
	pc := 0
	depth := 0 // extra nesting from if blocks

	for pc < len(code) {
		op := code[pc]
		pc++

		switch op {
		case irLiteral:
			idx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			v, ok := prog.constants[idx].(Int)
			if !ok {
				return nil
			}
			o = append(o, 0x42)
			o = appendSLEB(o, int64(v.I))

		case irLoadSlot:
			idx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			o = append(o, 0x20)
			o = appendULEB(o, idx)

		case irStoreSlot:
			idx := int(code[pc])<<8 | int(code[pc+1])
			pc += 2
			o = append(o, 0x21)
			o = appendULEB(o, idx)

		case irAdd:
			o = append(o, 0x7c)
		case irSub:
			o = append(o, 0x7d)
		case irMul:
			o = append(o, 0x7e)
		case irRem:
			o = append(o, 0x81) // i64.rem_s
		case irInc:
			o = append(o, 0x42, 0x01, 0x7c)
		case irDec:
			o = append(o, 0x42, 0x01, 0x7d)
		case irLt:
			o = append(o, 0x53, 0xad)
		case irEq:
			o = append(o, 0x51, 0xad)
		case irIsZero:
			o = append(o, 0x50, 0xad)

		case irJumpIfNot:
			pc += 2
			o = append(o, 0xa7)       // i32.wrap_i64
			o = append(o, 0x04, 0x40) // if void
			depth++

		case irJump:
			pc += 2
			o = append(o, 0x05) // else

		case irReturn:
			// Value on stack → br to $exit (block i64)
			// Depth: depth (ifs) + 1 (loop) + 0 ($exit is the block)
			// br N targets the Nth enclosing label from current position.
			// Labels: if_0..if_{depth-1}, loop, block
			// $exit = depth + 1
			o = append(o, 0x0c)
			o = appendULEB(o, depth+1)
			// If we're inside an if and no explicit else follows,
			// emit else so the false branch code has somewhere to go.
			if depth > 0 && pc < len(code) && code[pc] != irJump {
				o = append(o, 0x05) // else
			}

		case irRecur:
			nargs := int(code[pc])<<8 | int(code[pc+1])
			pc += 4
			for i := nargs - 1; i >= 0; i-- {
				o = append(o, 0x21)
				o = appendULEB(o, i)
			}
			// br to $loop
			// $loop = depth (ifs are between us and loop)
			o = append(o, 0x0c)
			o = appendULEB(o, depth)

		default:
			return nil
		}
	}

	// Close any open if blocks
	for depth > 0 {
		o = append(o, 0x0b)
		depth--
	}

	o = append(o, 0x0b)       // end loop
	o = append(o, 0x42, 0x00) // unreachable i64
	o = append(o, 0x0b)       // end block
	o = append(o, 0x0b)       // end func
	return o
}
