package duktape

/*
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

const goFuncCallName = "__goFuncCall__"
const (
	DUK_TYPE_NONE Type = iota
	DUK_TYPE_UNDEFINED
	DUK_TYPE_NULL
	DUK_TYPE_BOOLEAN
	DUK_TYPE_NUMBER
	DUK_TYPE_STRING
	DUK_TYPE_OBJECT
	DUK_TYPE_BUFFER
	DUK_TYPE_POINTER
)

type Type int

func (t Type) IsNone() bool      { return t == DUK_TYPE_NONE }
func (t Type) IsUndefined() bool { return t == DUK_TYPE_UNDEFINED }
func (t Type) IsNull() bool      { return t == DUK_TYPE_NULL }
func (t Type) IsBool() bool      { return t == DUK_TYPE_BOOLEAN }
func (t Type) IsNumber() bool    { return t == DUK_TYPE_NUMBER }
func (t Type) IsString() bool    { return t == DUK_TYPE_STRING }
func (t Type) IsObject() bool    { return t == DUK_TYPE_OBJECT }
func (t Type) IsBuffer() bool    { return t == DUK_TYPE_BUFFER }
func (t Type) IsPointer() bool   { return t == DUK_TYPE_POINTER }

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
