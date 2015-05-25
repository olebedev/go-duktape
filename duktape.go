package duktape

/*
#cgo linux LDFLAGS: -lm

# include "duktape.h"
extern duk_ret_t goFuncCall(duk_context *ctx);
extern void goFinalizeCall(duk_context *ctx);

*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"
	"unsafe"
)

const (
	CompileEval uint = 1 << iota
	CompileFunction
	CompileStrict
	CompileSafe
	CompileNoResult
	CompileNoSource
	CompileStrlen
)

const (
	TypeNone Type = iota
	TypeUndefined
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeObject
	TypeBuffer
	TypePointer
	TypeLightFunc
)

const (
	TypeMaskNone uint = 1 << iota
	TypeMaskUndefined
	TypeMaskNull
	TypeMaskBoolean
	TypeMaskNumber
	TypeMaskString
	TypeMaskObject
	TypeMaskBuffer
	TypeMaskPointer
	TypeMaskLightFunc
)

const (
	EnumIncludeNonenumerable uint = 1 << iota
	EnumIncludeInternal
	EnumOwnPropertiesOnly
	EnumArrayIndicesOnly
	EnumSortArrayIndices
	NoProxyBehavior
)

const (
	ErrNone int = 0

	// Internal to Duktape
	ErrUnimplemented int = 50 + iota
	ErrUnsupported
	ErrInternal
	ErrAlloc
	ErrAssertion
	ErrAPI
	ErrUncaughtError
)

const (
	// Common prototypes
	ErrError int = 100 + iota
	ErrEval
	ErrRange
	ErrReference
	ErrSyntax
	ErrType
	ErrURI
)

const (
	// Returned error values
	ErrRetUnimplemented int = -(ErrUnimplemented + iota)
	ErrRetUnsupported
	ErrRetInternal
	ErrRetAlloc
	ErrRetAssertion
	ErrRetAPI
	ErrRetUncaughtError
)

const (
	ErrRetError int = -(ErrError + iota)
	ErrRetEval
	ErrRetRange
	ErrRetReference
	ErrRetSyntax
	ErrRetType
	ErrRetURI
)

const (
	ExecSuccess = iota
	ExecError
)

const (
	LogTrace int = iota
	LogDebug
	LogInfo
	LogWarn
	LogError
	LogFatal
)

const goFuncCallName = "__goFuncCall__"
const goFinalizeCallName = "__goFinalizeCall__"
const goCtxName = "__goCtx__"
const goFunctionHandler = `
    function(){
	    return %s.apply(this, ['%s'].concat(Array.prototype.slice.apply(arguments)));
    };
`

const goFinalizeHandler = `
    function(){
	    %s('%s');
    };
`

type Type int

func (t Type) IsNone() bool      { return t == TypeNone }
func (t Type) IsUndefined() bool { return t == TypeUndefined }
func (t Type) IsNull() bool      { return t == TypeNull }
func (t Type) IsBool() bool      { return t == TypeBoolean }
func (t Type) IsNumber() bool    { return t == TypeNumber }
func (t Type) IsString() bool    { return t == TypeString }
func (t Type) IsObject() bool    { return t == TypeObject }
func (t Type) IsBuffer() bool    { return t == TypeBuffer }
func (t Type) IsPointer() bool   { return t == TypePointer }
func (t Type) IsLightFunc() bool { return t == TypeLightFunc }

func (t Type) String() string {
	switch t {
	case TypeNone:
		return "None"
	case TypeUndefined:
		return "Undefined"
	case TypeNull:
		return "Null"
	case TypeBoolean:
		return "Boolean"
	case TypeNumber:
		return "Number"
	case TypeString:
		return "String"
	case TypeObject:
		return "Object"
	case TypeBuffer:
		return "Buffer"
	case TypePointer:
		return "Pointer"
	case TypeLightFunc:
		return "LightFunc"
	default:
		return "Unknown"
	}
}

type Context struct {
	duk_context unsafe.Pointer
	fn          map[string]func(*Context) int
	mu          sync.Mutex
}

func New() *Context {
	panic("Unimplemented")
}

// Default returns plain initialized duktape context object
// See: http://duktape.org/api.html#duk_create_heap_default
func Default() *Context {
	ctx := &Context{
		duk_context: C.duk_create_heap(nil, nil, nil, nil, nil),
		fn:          make(map[string]func(*Context) int),
	}
	ctx.defineGoFuncCall()
	ctx.defineGoFinalizeCall()
	ctx.pushGoCtx()
	return ctx
}

// DEPRICATED
func NewContext() *Context {
	log.Println(`
		duktape.NewContext() is depricated, please use 
		duktape.New() or duktape.Default() instead
	`)
	return Default()
}

//export goFuncCall
func goFuncCall(ctx unsafe.Pointer) C.duk_ret_t {
	d := &Context{duk_context: ctx}
	d.pullGoCtx()

	// d.PushContextDump()
	// log.Printf("goCall context: %s", d.GetString(-1))
	// d.Pop()

	if d.GetTop() == 0 {
		panic("Go function call without arguments is not supported")
	}
	if !d.GetType(0).IsString() {
		panic("Wrong type of function's key argument")
	}
	name := d.GetString(0)
	if fn, ok := d.fn[name]; ok {
		r := fn(d)
		return C.duk_ret_t(r)
	}
	panic("Unimplemented")
}

//export goFinalizeCall
func goFinalizeCall(ctx unsafe.Pointer) {
	d := &Context{duk_context: ctx}
	d.pullGoCtx()

	if d.GetTop() == 0 {
		panic("Go function call without arguments is not supported")
	}
	if !d.GetType(0).IsString() {
		panic("Wrong type of function's key argument")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	name := d.GetString(0)
	if _, ok := d.fn[name]; ok {
		delete(d.fn, name)
	}
}

var reFuncName = regexp.MustCompile("^[a-z_][a-z0-9_]*([A-Z_][a-z0-9_]*)*$")

// PushGlobalGoFunction push the given function into duktape global object
func (d *Context) PushGlobalGoFunction(name string, fn func(*Context) int) error {
	if !reFuncName.MatchString(name) {
		return errors.New("Malformed function name '" + name + "'")
	}

	d.PushGlobalObject()
	if err := d.pushGoFunction(name, fn); err != nil {
		return err
	}

	d.PutPropString(-2, name)
	d.Pop()

	return nil
}

// PushGoFunction push the given function into duktape stack
func (d *Context) PushGoFunction(fn func(*Context) int) (string, error) {
	name := fmt.Sprintf("anon_%d", time.Now().Nanosecond())
	return name, d.pushGoFunction(name, fn)
}

func (d *Context) pushGoFunction(name string, fn func(*Context) int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.fn[name] = fn

	d.CompileString(CompileFunction, fmt.Sprintf(
		goFunctionHandler, goFuncCallName, name,
	))

	d.CompileString(CompileFunction, fmt.Sprintf(
		goFinalizeHandler, goFinalizeCallName, name,
	))

	d.SetFinalizer(-2)
	return nil
}

// PopGoFunc cleans given function from duktape javascript context
func (d *Context) PopGoFunc(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.fn[name]; ok {
		d.EvalString(fmt.Sprintf(`%s = undefined;`, name))
		delete(d.fn, name)
	}
}

func (d *Context) defineGoFuncCall() {
	d.PushGlobalObject()
	d.PushCFunction((*[0]byte)(C.goFuncCall), int(C.DUK_VARARGS))
	d.PutPropString(-2, goFuncCallName)
	d.Pop()
}

func (d *Context) defineGoFinalizeCall() {
	d.PushGlobalObject()
	d.PushCFunction((*[0]byte)(C.goFinalizeCall), int(C.DUK_VARARGS))
	d.PutPropString(-2, goFinalizeCallName)
	d.Pop()
}

func (d *Context) pushGoCtx() {
	d.PushGlobalObject()
	d.PushPointer(unsafe.Pointer(d))
	d.PutPropString(-2, goCtxName)
	d.Pop()
}

func (d *Context) pullGoCtx() {
	d.PushGlobalObject()
	d.GetPropString(-1, goCtxName)
	ctx := (*Context)(d.GetPointer(-1))
	d.fn = ctx.fn
	d.mu = ctx.mu
	d.Pop2()
}

type Error struct {
	Type       string
	Message    string
	FileName   string
	LineNumber int
	Stack      string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}
