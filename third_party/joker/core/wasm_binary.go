package core

// wasm_binary.go — low-level WASM binary format helpers.
// Produces raw WASM module bytes from sections.

import (
	"encoding/binary"
	"math"
)

// wasmModule builds a minimal WASM module binary.
type wasmModule struct {
	buf []byte
}

func newWasmModule() *wasmModule {
	m := &wasmModule{}
	m.buf = append(m.buf, 0x00, 0x61, 0x73, 0x6d) // magic: \0asm
	m.buf = append(m.buf, 0x01, 0x00, 0x00, 0x00) // version 1
	return m
}

// addTypeSection adds a single func type: (i64, i64, ...) -> (i64)
func (m *wasmModule) addTypeSection(numParams int) {
	var body []byte
	body = append(body, 0x01)          // 1 type entry
	body = append(body, 0x60)          // functype
	body = appendULEB(body, numParams) // param count
	for i := 0; i < numParams; i++ {
		body = append(body, 0x7e) // i64
	}
	body = append(body, 0x01, 0x7e) // 1 result: i64
	m.addSection(0x01, body)
}

// addFuncSection declares 1 function with type index 0.
func (m *wasmModule) addFuncSection() {
	m.addSection(0x03, []byte{0x01, 0x00})
}

// addExportSection exports function 0 as "exec".
func (m *wasmModule) addExportSection() {
	name := []byte("exec")
	var body []byte
	body = append(body, 0x01) // 1 export
	body = appendULEB(body, len(name))
	body = append(body, name...)
	body = append(body, 0x00, 0x00) // funcidx 0
	m.addSection(0x07, body)
}

// addCodeSection adds 1 function body.
func (m *wasmModule) addCodeSection(funcBody []byte) {
	var inner []byte
	inner = append(inner, 0x01) // 1 function
	inner = appendULEB(inner, len(funcBody))
	inner = append(inner, funcBody...)
	m.addSection(0x0a, inner)
}

func (m *wasmModule) addSection(id byte, body []byte) {
	m.buf = append(m.buf, id)
	m.buf = appendULEB(m.buf, len(body))
	m.buf = append(m.buf, body...)
}

func (m *wasmModule) bytes() []byte { return m.buf }

// --- LEB128 encoding ---

func appendULEB(buf []byte, v int) []byte {
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		buf = append(buf, b)
		if v == 0 {
			break
		}
	}
	return buf
}

func appendSLEB(buf []byte, v int64) []byte {
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if (v == 0 && b&0x40 == 0) || (v == -1 && b&0x40 != 0) {
			buf = append(buf, b)
			break
		}
		buf = append(buf, b|0x80)
	}
	return buf
}

func appendF64(buf []byte, v float64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], math.Float64bits(v))
	return append(buf, b[:]...)
}
