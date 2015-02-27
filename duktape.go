package duktape

/*
#cgo linux LDFLAGS: -lm

# include "duktape.h"
extern duk_ret_t goFuncCall(duk_context *ctx);
extern duk_ret_t testFunc(duk_context *ctx);
*/
import "C"
import "errors"
import "fmt"
import "regexp"
import "time"
import "unsafe"

const (
	CompileEval uint = 1 << iota
	CompileFunction
	CompileStrict
	CompileSafe
	CompileNoResult
	CompileNoSource
	CompileStrlen
)

const goFuncCallName = "__goFuncCall__"
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

type Context struct {
	duk_context unsafe.Pointer
}

// Returns initialized duktape context object
func NewContext() *Context {
	ctx := &Context{
		duk_context: C.duk_create_heap(nil, nil, nil, nil, nil),
	}
	ctx.defineGoFuncCall()
	return ctx
}

//export goFuncCall
func goFuncCall(ctx unsafe.Pointer) C.duk_ret_t {
	c := &Context{ctx}
	if c.GetTop() == 0 {
		// unexpected call, without function name's hash
		panic("Go function call without arguments is not supported")
		return C.DUK_RET_UNSUPPORTED_ERROR
	}
	if !Type(c.GetType(0)).IsString() {
		// unexpected type of function name's hash
		panic("Wrong type of function's key argument")
		return C.DUK_RET_EVAL_ERROR
	}
	key := c.GetString(0)
	if fn, ok := goFuncMap[key]; ok {
		r := fn(c)
		return C.duk_ret_t(r)
	}
	panic("Unimplemented")
	return C.DUK_RET_UNIMPLEMENTED_ERROR
}

func getKeyFor(funcName string) string {
	c := 0
	key := fmt.Sprintf("__%s_%d%d__", funcName, time.Now().Nanosecond(), c)
	for {
		if _, ok := goFuncMap[key]; ok {
			c++
			key = fmt.Sprintf("__%s_%d%d__", funcName, time.Now().Nanosecond(), c)
			continue
		}
		break
	}
	return key
}

var goFuncMap = map[string]func(*Context) int{}
var reFuncName = regexp.MustCompile("^[a-z_][a-z0-9_]*([A-Z_][a-z0-9_]*)*$")

func (d *Context) PushGoFunc(name string, fn func(*Context) int) error {
	if !reFuncName.MatchString(name) {
		return errors.New("Malformed function name '" + name + "'")
	}
	key := getKeyFor(name)
	goFuncMap[key] = fn

	// TODO: apply dot notation names
	d.EvalString(fmt.Sprintf(`
      function %s (){
        return %s.apply(this, ['%s'].concat(Array.prototype.slice.apply(arguments)));
      };
  `, name, goFuncCallName, key))
	d.Pop()
	return nil
}

func (d *Context) defineGoFuncCall() {
	d.PushGlobalObject()
	d.PushCFunction((*[0]byte)(C.goFuncCall), int(C.DUK_VARARGS))
	d.PutPropString(-2, goFuncCallName)
	d.Pop()
}

/**
 * only for tests
 */
func goTestfunc(ctx *Context) int {
	top := ctx.GetTop()
	a := ctx.GetNumber(top - 2)
	b := ctx.GetNumber(top - 1)
	ctx.PushNumber(a + b)
	return 1
}

//export testFunc
func testFunc(ctx unsafe.Pointer) C.duk_ret_t {
	return C.duk_ret_t(goTestfunc(&Context{ctx}))
}

func (d *Context) pushTestFunc() {
	d.PushCFunction((*[0]byte)(C.testFunc), (-1))
}
