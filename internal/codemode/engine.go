// Package codemode runs model-written JavaScript in QuickJS compiled to
// WebAssembly (quickjs-wasi) under wazero, with Pi's codemode semantics: the
// script is the body of an async function whose only capabilities are the
// injected tools and globals (tools.<name>(args), ALL_TOOLS, text, image,
// exit, console.*, store/load). Nested tool calls never reach the model; only
// the script's output and return value do.
//
// The guest side is Pi's own prelude (vendor/prelude.js, MIT); this package is
// the host: it instantiates a fresh VM per execution, relays bridge calls,
// runs tool calls concurrently and settles their promises, detects scripts
// waiting on nothing that can resume them, and enforces timeout, cancellation,
// memory and stack limits. See docs/implementation/plans/mcp-codemode-plan.md.
package codemode

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed vendor/quickjs.wasm
var quickjsWasm []byte

//go:embed vendor/prelude.js
var preludeSource string

const (
	// DefaultTimeout is Pi's default execution deadline.
	DefaultTimeout = 300 * time.Second
	// maxStackSize is quickjs-wasi's MAX_STACK_SIZE: deep recursion throws a
	// catchable RangeError instead of trapping.
	maxStackSize = 512 * 1024
	// memoryLimitPages caps a VM's linear memory (64 KiB pages: 256 MiB).
	memoryLimitPages = 4096
)

// Error kinds, as in Pi's CodemodeError.
const (
	ErrorScript  = "script"
	ErrorTimeout = "timeout"
	ErrorAborted = "aborted"
	ErrorSandbox = "sandbox"
)

// Tool is a function scripts call as tools.<name>(args). Arguments and results
// make a JSON round trip; an error rejects the script's promise with its
// message.
type Tool struct {
	Name        string
	Description string
	Execute     func(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

// Global is a top-level function (or namespace.function) scripts call
// directly; with Spread it receives all call arguments as a JSON array.
type Global struct {
	Name    string
	Spread  bool
	Execute func(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

// Options for one execution.
type Options struct {
	Tools            []Tool
	Globals          []Global
	Store            map[string]json.RawMessage // values store()d by earlier runs
	Timeout          time.Duration              // 0: DefaultTimeout
	MemoryLimitBytes int                        // QuickJS heap limit; 0: none
}

// OutputItem is one text or image item a script produced.
type OutputItem struct {
	Type     string `json:"type"` // "text" or "image"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"` // base64
	MimeType string `json:"mimeType,omitempty"`
}

// Call records one tool call made by the script.
type Call struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // ok, error, cancelled
	Duration time.Duration `json:"durationMs"`
}

// Error describes a failed execution.
type Error struct {
	Kind    string `json:"kind"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

func (e *Error) Error() string { return e.Kind + ": " + e.Message }

// StoreWrites are the store changes of a successful run.
type StoreWrites struct {
	Set    map[string]json.RawMessage `json:"set"`
	Delete []string                   `json:"delete"`
}

// Result of an execution. Value is the script's return value as JSON (nil
// for undefined).
type Result struct {
	OK          bool
	Value       json.RawMessage
	Error       *Error
	Output      []OutputItem
	Calls       []Call
	StoreWrites StoreWrites
}

// Engine compiles the QuickJS module once and runs executions concurrently,
// each in its own module instance.
type Engine struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

var (
	defaultOnce   sync.Once
	defaultEngine *Engine
	defaultErr    error
)

// Default returns the process-wide engine, compiling QuickJS on first use.
func Default() (*Engine, error) {
	defaultOnce.Do(func() { defaultEngine, defaultErr = NewEngine(context.Background()) })
	return defaultEngine, defaultErr
}

// NewEngine compiles the QuickJS module.
func NewEngine(ctx context.Context) (*Engine, error) {
	cfg := wazero.NewRuntimeConfig().WithCloseOnContextDone(true).WithMemoryLimitPages(memoryLimitPages)
	r := wazero.NewRuntimeWithConfig(ctx, cfg)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("codemode: wasi: %w", err)
	}
	if _, err := hostModule(r).Instantiate(ctx); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("codemode: host module: %w", err)
	}
	compiled, err := r.CompileModule(ctx, quickjsWasm)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("codemode: compile quickjs: %w", err)
	}
	return &Engine{runtime: r, compiled: compiled}, nil
}

// Close releases the runtime.
func (e *Engine) Close(ctx context.Context) error { return e.runtime.Close(ctx) }

type ctxKey struct{}

// vm is one execution: a QuickJS instance and the host state around it.
type vm struct {
	mod     api.Module
	mem     api.Memory
	fn      map[string]api.Function
	tools   map[string]Tool
	globals map[string]Global

	undefined, trueV, falseV uint32

	output []OutputItem
	calls  []Call
	events chan callResult

	apiHandles apiHandles
	timeout    time.Duration
	done       bool
	result     Result
	pendingIDs map[int]int // call id -> index into calls (tools only)
	running    sync.WaitGroup
	callCtx    context.Context
	interrupt  bool
}

type callResult struct {
	id      int
	ok      bool
	payload *string
	tool    bool
	elapsed time.Duration
}

func hostModule(r wazero.Runtime) wazero.HostModuleBuilder {
	b := r.NewHostModuleBuilder("env")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, namePtr, nameLen, thisPtr, argc, argvPtr uint32) uint32 {
		v, _ := ctx.Value(ctxKey{}).(*vm)
		if v == nil {
			return 0
		}
		return v.hostCall(ctx, namePtr, nameLen, argc, argvPtr)
	}).Export("host_call")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context) uint32 {
		v, _ := ctx.Value(ctxKey{}).(*vm)
		if v != nil && (v.interrupt || ctx.Err() != nil) {
			return 1
		}
		return 0
	}).Export("host_interrupt")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, promise, reason, handled uint32) {
		if v, _ := ctx.Value(ctxKey{}).(*vm); v != nil {
			v.free(promise)
			v.free(reason)
		}
	}).Export("host_promise_rejection")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context, base, name uint32) uint32 { return 0 }).Export("host_module_normalize")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context, name, outLen uint32) uint32 { return 0 }).Export("host_module_load")
	b.NewFunctionBuilder().WithFunc(func(ctx context.Context, hi, lo uint32) uint32 { return 0 }).Export("host_get_timezone_offset")
	return b
}

// Execute runs a script (the body of an async function). It never returns
// an error for script failures; those are Result.Error. Cancelling ctx aborts
// the run.
func (e *Engine) Execute(ctx context.Context, code string, opts Options) Result {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	v := &vm{tools: map[string]Tool{}, globals: map[string]Global{}, events: make(chan callResult, 64), pendingIDs: map[int]int{}, timeout: timeout}
	callCtx, cancelCalls := context.WithCancel(runCtx)
	v.callCtx = callCtx
	defer func() {
		cancelCalls()
		v.running.Wait()
	}()
	for _, t := range opts.Tools {
		v.tools[t.Name] = t
	}
	for _, g := range opts.Globals {
		v.globals[g.Name] = g
	}
	wasmCtx := context.WithValue(runCtx, ctxKey{}, v)
	finish := func(err *Error) Result {
		v.interrupt = true
		res := Result{OK: err == nil, Error: err, Output: v.output, Calls: v.markCancelled()}
		if err == nil {
			res = v.result
			res.Output, res.Calls = v.output, v.markCancelled()
		}
		if res.StoreWrites.Set == nil {
			res.StoreWrites.Set = map[string]json.RawMessage{}
		}
		if res.StoreWrites.Delete == nil {
			res.StoreWrites.Delete = []string{}
		}
		return res
	}
	modCfg := wazero.NewModuleConfig().WithName("").WithStartFunctions().
		WithStdout(io.Discard).WithStderr(io.Discard).WithRandSource(rand.Reader).WithSysWalltime().WithSysNanotime()
	mod, err := e.runtime.InstantiateModule(wasmCtx, e.compiled, modCfg)
	if err != nil {
		return finish(v.contextError(runCtx, ctx, fmt.Sprintf("Failed to start QuickJS: %v", err)))
	}
	defer mod.Close(context.Background())
	v.mod, v.mem = mod, mod.Memory()
	v.fn = map[string]api.Function{}
	if err := v.setup(wasmCtx, opts); err != nil {
		return finish(v.contextError(runCtx, ctx, err.Error()))
	}
	// The prefix shares the script's first line so line numbers match.
	fnHandle, exc, err := v.eval(wasmCtx, "(async (tools, console) => {"+code+"\n})", "codemode.js")
	if err != nil {
		return finish(v.contextError(runCtx, ctx, err.Error()))
	}
	if exc != nil {
		return finish(exc)
	}
	// failed is a run that stopped with err: the script's own result once it
	// reported one (done interrupts the script), else err classified.
	failed := func(err error) Result {
		if v.done {
			return finish(v.result.Error)
		}
		return finish(v.contextError(runCtx, ctx, err.Error()))
	}
	handles := v.apiHandles
	if _, err := v.callFn(wasmCtx, handles.run, handles.obj, fnHandle); err != nil {
		return failed(err)
	}
	v.free(fnHandle)
	if err := v.drain(wasmCtx); err != nil {
		return failed(err)
	}
	for !v.done {
		select {
		case r := <-v.events:
			v.recordCall(r)
			if err := v.settle(wasmCtx, r); err != nil {
				return failed(err)
			}
		case <-runCtx.Done():
			return finish(v.contextError(runCtx, ctx, ""))
		}
	}
	return finish(v.result.Error)
}

// contextError classifies a failure: timeout, abort (caller cancelled) or a
// sandbox failure.
func (v *vm) contextError(runCtx, parent context.Context, detail string) *Error {
	switch {
	case parent.Err() != nil:
		return &Error{Kind: ErrorAborted, Message: "Execution aborted"}
	case errors.Is(runCtx.Err(), context.DeadlineExceeded):
		return &Error{Kind: ErrorTimeout, Message: fmt.Sprintf("Execution timed out after %d ms", v.timeout.Milliseconds())}
	}
	if detail == "" {
		detail = "QuickJS stopped unexpectedly"
	}
	return &Error{Kind: ErrorSandbox, Message: detail}
}

func (v *vm) markCancelled() []Call {
	out := append([]Call(nil), v.calls...)
	return out
}

func (v *vm) recordCall(r callResult) {
	if !r.tool {
		return
	}
	if i, ok := v.pendingIDs[r.id]; ok {
		v.calls[i].Duration = r.elapsed
		v.calls[i].Status = "error"
		if r.ok {
			v.calls[i].Status = "ok"
		}
		delete(v.pendingIDs, r.id)
	}
}

// ---- low-level QuickJS helpers -------------------------------------------

func (v *vm) call(ctx context.Context, name string, args ...uint64) (uint64, error) {
	f := v.fn[name]
	if f == nil {
		f = v.mod.ExportedFunction(name)
		if f == nil {
			return 0, fmt.Errorf("quickjs export %s missing", name)
		}
		v.fn[name] = f
	}
	res, err := f.Call(ctx, args...)
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 0, nil
	}
	return res[0], nil
}

func (v *vm) mustCall(ctx context.Context, name string, args ...uint64) uint32 {
	r, err := v.call(ctx, name, args...)
	if err != nil {
		panic(err)
	}
	return uint32(r)
}

func (v *vm) free(h uint32) {
	if h != 0 {
		_, _ = v.call(context.Background(), "qjs_free_value", uint64(h))
	}
}

// writeBytes copies data into guest memory (NUL-terminated); caller frees.
func (v *vm) writeBytes(ctx context.Context, data []byte) (uint32, error) {
	ptr64, err := v.call(ctx, "wasm_malloc", uint64(len(data)+1))
	ptr := uint32(ptr64)
	if err != nil || ptr == 0 {
		return 0, fmt.Errorf("wasm_malloc failed")
	}
	if !v.mem.Write(ptr, append(data, 0)) {
		return 0, fmt.Errorf("guest memory write out of range")
	}
	return ptr, nil
}

func (v *vm) newString(ctx context.Context, s string) (uint32, error) {
	ptr, err := v.writeBytes(ctx, []byte(s))
	if err != nil {
		return 0, err
	}
	defer v.call(ctx, "wasm_free", uint64(ptr))
	h, err := v.call(ctx, "qjs_new_string", uint64(ptr), uint64(len(s)))
	return uint32(h), err
}

func (v *vm) toString(ctx context.Context, h uint32) (string, error) {
	lenPtr64, err := v.call(ctx, "wasm_malloc", 4)
	lenPtr := uint32(lenPtr64)
	if err != nil || lenPtr == 0 {
		return "", fmt.Errorf("wasm_malloc failed")
	}
	defer v.call(ctx, "wasm_free", uint64(lenPtr))
	cstr64, err := v.call(ctx, "qjs_get_string_len", uint64(h), uint64(lenPtr))
	cstr := uint32(cstr64)
	if err != nil {
		return "", err
	}
	if cstr == 0 {
		return "", nil
	}
	defer v.call(ctx, "qjs_free_cstring", uint64(cstr))
	n, _ := v.mem.ReadUint32Le(lenPtr)
	b, ok := v.mem.Read(cstr, n)
	if !ok {
		return "", fmt.Errorf("guest memory read out of range")
	}
	return string(b), nil
}

func (v *vm) isUndefined(ctx context.Context, h uint32) bool {
	r, _ := v.call(ctx, "qjs_is_undefined", uint64(h))
	return r != 0
}

func (v *vm) truthy(ctx context.Context, h uint32) bool {
	r, _ := v.call(ctx, "qjs_get_bool", uint64(h))
	return r != 0
}

func (v *vm) getProp(ctx context.Context, obj uint32, name string) (uint32, error) {
	ptr, err := v.writeBytes(ctx, []byte(name))
	if err != nil {
		return 0, err
	}
	defer v.call(ctx, "wasm_free", uint64(ptr))
	h, err := v.call(ctx, "qjs_get_prop_string", uint64(obj), uint64(ptr))
	return uint32(h), err
}

// exception returns the pending exception as a script Error.
func (v *vm) exception(ctx context.Context) *Error {
	exc := v.mustCall(ctx, "qjs_get_exception")
	defer v.free(exc)
	field := func(name string) string {
		h, err := v.getProp(ctx, exc, name)
		if err != nil {
			return ""
		}
		defer v.free(h)
		if v.isUndefined(ctx, h) {
			return ""
		}
		s, _ := v.toString(ctx, h)
		return s
	}
	name, message, stack := field("name"), field("message"), strings.TrimRight(field("stack"), "\n")
	if name == "" && message == "" {
		s, _ := v.toString(ctx, exc)
		return &Error{Kind: ErrorScript, Message: s}
	}
	head := name
	if message != "" {
		head = name + ": " + message
	}
	if stack != "" {
		stack = head + "\n" + stack
	} else {
		stack = head
	}
	return &Error{Kind: ErrorScript, Name: name, Message: message, Stack: stack}
}

// eval evaluates code; a JS exception is returned as a script Error.
func (v *vm) eval(ctx context.Context, code, filename string) (uint32, *Error, error) {
	codePtr, err := v.writeBytes(ctx, []byte(code))
	if err != nil {
		return 0, nil, err
	}
	defer v.call(ctx, "wasm_free", uint64(codePtr))
	namePtr, err := v.writeBytes(ctx, []byte(filename))
	if err != nil {
		return 0, nil, err
	}
	defer v.call(ctx, "wasm_free", uint64(namePtr))
	h64, err := v.call(ctx, "qjs_eval", uint64(codePtr), uint64(len(code)), uint64(namePtr), 0)
	if err != nil {
		return 0, nil, err
	}
	h := uint32(h64)
	if r, _ := v.call(ctx, "qjs_is_exception", uint64(h)); r != 0 {
		v.free(h)
		return 0, v.exception(ctx), nil
	}
	return h, nil, nil
}

// callFn calls fn(this, args...); a JS exception is a Go error (prelude
// functions never throw for script failures).
func (v *vm) callFn(ctx context.Context, fn, this uint32, args ...uint32) (uint32, error) {
	argv := uint32(0)
	if len(args) > 0 {
		p, err := v.call(ctx, "wasm_malloc", uint64(4*len(args)))
		if err != nil || p == 0 {
			return 0, fmt.Errorf("wasm_malloc failed")
		}
		argv = uint32(p)
		defer v.call(ctx, "wasm_free", uint64(argv))
		for i, a := range args {
			v.mem.WriteUint32Le(argv+uint32(4*i), a)
		}
	}
	h64, err := v.call(ctx, "qjs_call", uint64(fn), uint64(this), uint64(len(args)), uint64(argv))
	if err != nil {
		return 0, err
	}
	h := uint32(h64)
	if r, _ := v.call(ctx, "qjs_is_exception", uint64(h)); r != 0 {
		v.free(h)
		e := v.exception(ctx)
		return 0, fmt.Errorf("prelude error: %s", e.Stack)
	}
	return h, nil
}

// drain runs queued jobs, then lets the prelude fail a script that waits on
// nothing that can resume it.
func (v *vm) drain(ctx context.Context) error {
	for {
		pending, err := v.call(ctx, "qjs_is_job_pending")
		if err != nil {
			return err
		}
		if pending == 0 {
			break
		}
		r, err := v.call(ctx, "qjs_execute_pending_job")
		if err != nil {
			return err
		}
		if int32(r) < 0 {
			e := v.exception(ctx)
			return fmt.Errorf("job execution error: %s", e.Stack)
		}
	}
	h, err := v.callFn(ctx, v.apiHandles.stalled, v.apiHandles.obj)
	v.free(h)
	return err
}

func (v *vm) settle(ctx context.Context, r callResult) error {
	id64, err := v.call(ctx, "qjs_new_number", math.Float64bits(float64(r.id)))
	if err != nil {
		return err
	}
	id := uint32(id64)
	defer v.free(id)
	okV := v.falseV
	if r.ok {
		okV = v.trueV
	}
	payload := v.undefined
	if r.payload != nil {
		p, err := v.newString(ctx, *r.payload)
		if err != nil {
			return err
		}
		defer v.free(p)
		payload = p
	}
	h, err := v.callFn(ctx, v.apiHandles.settle, v.apiHandles.obj, id, okV, payload)
	if err != nil {
		return err
	}
	v.free(h)
	return v.drain(ctx)
}

type apiHandles struct{ obj, run, settle, stalled uint32 }

// setup initialises QuickJS, limits and the prelude API.
func (v *vm) setup(ctx context.Context, opts Options) error {
	if _, err := v.call(ctx, "_initialize"); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	if r, err := v.call(ctx, "qjs_init"); err != nil || r != 0 {
		return fmt.Errorf("failed to initialize QuickJS runtime")
	}
	if opts.MemoryLimitBytes > 0 {
		if _, err := v.call(ctx, "qjs_set_memory_limit", uint64(opts.MemoryLimitBytes)); err != nil {
			return err
		}
	}
	if _, err := v.call(ctx, "qjs_set_max_stack_size", maxStackSize); err != nil {
		return err
	}
	if _, err := v.call(ctx, "qjs_set_interrupt_handler", 1); err != nil {
		return err
	}
	v.undefined = v.mustCall(ctx, "qjs_get_undefined")
	v.trueV = v.mustCall(ctx, "qjs_get_true")
	v.falseV = v.mustCall(ctx, "qjs_get_false")

	type toolInfo struct {
		Name        string `json:"name"`
		JSName      string `json:"jsName"`
		Description string `json:"description"`
	}
	type globalInfo struct {
		Name   string `json:"name"`
		Spread bool   `json:"spread"`
	}
	tools := make([]toolInfo, 0, len(opts.Tools))
	for _, t := range opts.Tools {
		tools = append(tools, toolInfo{Name: t.Name, JSName: Identifier(t.Name), Description: t.Description})
	}
	globals := make([]globalInfo, 0, len(opts.Globals))
	for _, g := range opts.Globals {
		globals = append(globals, globalInfo{Name: g.Name, Spread: g.Spread})
	}
	store := map[string]string{}
	for k, raw := range opts.Store {
		store[k] = string(raw)
	}
	toolsJSON, _ := json.Marshal(tools)
	globalsJSON, _ := json.Marshal(globals)
	storeJSON, _ := json.Marshal(store)

	namePtr, err := v.writeBytes(ctx, []byte("bridge"))
	if err != nil {
		return err
	}
	bridge64, err := v.call(ctx, "qjs_new_host_function", uint64(namePtr), 6, 0)
	v.call(ctx, "wasm_free", uint64(namePtr))
	if err != nil {
		return err
	}
	bridge := uint32(bridge64)
	defer v.free(bridge)
	preludeFn, exc, err := v.eval(ctx, preludeSource, "codemode-prelude.js")
	if err != nil {
		return err
	}
	if exc != nil {
		return fmt.Errorf("prelude: %s", exc.Stack)
	}
	defer v.free(preludeFn)
	var strs []uint32
	for _, s := range []string{string(toolsJSON), string(globalsJSON), string(storeJSON)} {
		h, err := v.newString(ctx, s)
		if err != nil {
			return err
		}
		strs = append(strs, h)
		defer v.free(h)
	}
	obj, err := v.callFn(ctx, preludeFn, v.undefined, bridge, strs[0], strs[1], strs[2])
	if err != nil {
		return err
	}
	v.apiHandles.obj = obj
	for name, dst := range map[string]*uint32{"run": &v.apiHandles.run, "settle": &v.apiHandles.settle, "stalled": &v.apiHandles.stalled} {
		h, err := v.getProp(ctx, obj, name)
		if err != nil {
			return err
		}
		*dst = h
	}
	return nil
}

// hostCall implements Pi's bridge(kind, a, b, c), called with primitives.
func (v *vm) hostCall(ctx context.Context, namePtr, nameLen, argc, argvPtr uint32) uint32 {
	args := make([]uint32, 4)
	for i := uint32(0); i < argc && i < 4; i++ {
		args[i], _ = v.mem.ReadUint32Le(argvPtr + 4*i)
	}
	arg := func(i int) (string, bool) {
		if args[i] == 0 || v.isUndefined(ctx, args[i]) {
			return "", false
		}
		s, _ := v.toString(ctx, args[i])
		return s, true
	}
	kind, _ := arg(0)
	switch kind {
	case "call", "global":
		idText, _ := arg(1)
		name, _ := arg(2)
		argsJSON, hasArgs := arg(3)
		var id int
		fmt.Sscan(idText, &id)
		v.startCall(kind == "call", id, name, argsJSON, hasArgs)
	case "output":
		typ, _ := arg(1)
		b, _ := arg(2)
		c, _ := arg(3)
		if typ == "image" {
			v.output = append(v.output, OutputItem{Type: "image", Data: b, MimeType: c})
		} else {
			v.output = append(v.output, OutputItem{Type: "text", Text: b})
		}
	case "done":
		ok := args[1] != 0 && v.truthy(ctx, args[1])
		// The result is final: the host ends the script (Pi), so a script
		// that catches the output-limit error cannot keep running.
		v.done, v.interrupt = true, true
		if ok {
			value, has := arg(2)
			writes, _ := arg(3)
			if has {
				v.result.Value = json.RawMessage(value)
			}
			v.result.OK = true
			v.result.StoreWrites = parseStoreWrites(writes)
		} else {
			payload, _ := arg(2)
			var e Error
			if json.Unmarshal([]byte(payload), &e) != nil {
				e.Message = payload
			}
			e.Kind = ErrorScript
			v.result.Error = &e
		}
	}
	return uint32(v.mustCall(ctx, "qjs_dup_value", uint64(v.undefined)))
}

func parseStoreWrites(raw string) StoreWrites {
	w := StoreWrites{Set: map[string]json.RawMessage{}, Delete: []string{}}
	var entries [][]*string
	if json.Unmarshal([]byte(raw), &entries) != nil {
		return w
	}
	for _, e := range entries {
		if len(e) == 0 || e[0] == nil {
			continue
		}
		if len(e) < 2 || e[1] == nil {
			w.Delete = append(w.Delete, *e[0])
		} else {
			w.Set[*e[0]] = json.RawMessage(*e[1])
		}
	}
	return w
}

// startCall runs a tool or global call concurrently; its result is settled
// from the event loop.
func (v *vm) startCall(isTool bool, id int, name, argsJSON string, hasArgs bool) {
	var exec func(context.Context, json.RawMessage) (json.RawMessage, error)
	if isTool {
		v.pendingIDs[id] = len(v.calls)
		v.calls = append(v.calls, Call{Name: name, Status: "cancelled"})
		if t, ok := v.tools[name]; ok {
			exec = t.Execute
		}
	} else if g, ok := v.globals[name]; ok {
		exec = g.Execute
	}
	var args json.RawMessage
	if hasArgs {
		args = json.RawMessage(argsJSON)
	}
	v.running.Add(1)
	go func() {
		defer v.running.Done()
		start := time.Now()
		r := callResult{id: id, tool: isTool}
		if exec == nil {
			what := "global"
			if isTool {
				what = "tool"
			}
			msg := fmt.Sprintf("Unknown %s %q", what, name)
			r.payload = &msg
		} else if out, err := exec(v.callCtx, args); err != nil {
			msg := err.Error()
			r.payload = &msg
		} else {
			r.ok = true
			if out != nil {
				s := string(out)
				r.payload = &s
			}
		}
		r.elapsed = time.Since(start)
		select {
		case v.events <- r:
		case <-v.callCtx.Done():
		}
	}()
}

// Identifier ports Pi's toCodemodeIdentifier: characters not valid in a
// JavaScript identifier become _.
func Identifier(name string) string {
	var b strings.Builder
	for _, r := range name {
		valid := r == '_' || r == '$' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (b.Len() > 0 && r >= '0' && r <= '9')
		if valid {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}
