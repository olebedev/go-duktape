package duktape

/*
# include "duktape.h"
static void go_duk_eval_string(duk_context *ctx, const char *str) {
  return duk_eval_string(ctx, str);
}
extern duk_ret_t goFuncCall(duk_context *ctx);
*/
import "C"
import "errors"
import "fmt"
import "log"
import "regexp"
import "time"
import "unsafe"

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

const goFuncCallName = "__goFuncCall__"

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
	fmt.Printf("goFuncCall with context: %#v\n", c)
	if c.GetTop() == 0 {
		// unexpected call, without function name's hash
		panic("Go function call without arguments is not supported")
		return C.DUK_RET_UNSUPPORTED_ERROR
	}
	if !c.GetType(0).IsString() {
		// unexpected type of function name's hash
		panic("Wrong type of hashName argument")
		return C.DUK_RET_EVAL_ERROR
	}
	hash := c.GetString(0)
	if fn, ok := goFuncMap[hash]; ok {
		return C.duk_ret_t(fn(c))
	}
	log.Printf("hash is: %s", hash)
	panic("Unimplemented")
	return C.DUK_RET_UNIMPLEMENTED_ERROR
}

func hashFor(funcName string) string {
	return fmt.Sprintf("__%s_%d__", funcName, time.Now().Nanosecond())
}

var goFuncMap = map[string]func(*Context) int{}
var reFuncName = regexp.MustCompile("^[a-z][a-z0-9]*([A-Z][a-z0-9]*)*$")

func (d *Context) PushGoFunc(name string, fn func(*Context) int) error {
	if !reFuncName.MatchString(name) {
		return errors.New("Malformed function name '" + name + "'")
	}
	hashed := hashFor(name)
	goFuncMap[hashed] = fn

	d.EvalString(fmt.Sprintf(`
      function %s (){
        %s.apply(this, ['%s'].concat(Array.prototype.slice.apply(arguments)));
      };
  `, name, goFuncCallName, hashed))
	d.Pop()
	return nil
}

func (d *Context) defineGoFuncCall() {
	d.PushGlobalObject()
	d.PushCFunction((*[0]byte)(C.goFuncCall), int(C.DUK_VARARGS))
	d.PutPropString(-2, goFuncCallName)
	d.Pop()
}

func (d *Context) GetTop() int {
	return int(C.duk_get_top(d.duk_context))
}

func (d *Context) PushCFunction(fn *[0]byte, nargs int) {
	C.duk_push_c_function(
		d.duk_context,
		fn,
		C.duk_idx_t(nargs),
	)
}

func (d *Context) EvalString(script string) {
	str := C.CString(script)
	defer C.free(unsafe.Pointer(str))
	C.go_duk_eval_string(d.duk_context, str)
}

func (d *Context) Pop() {
	C.duk_pop(d.duk_context)
}

func (d *Context) GetType(i int) Type {
	return Type(C.duk_get_type(d.duk_context, C.duk_idx_t(i)))
}

func (d *Context) GetString(i int) string {
	if d.GetType(i).IsString() {
		if s := C.duk_get_string(d.duk_context, C.duk_idx_t(i)); s != nil {
			return C.GoString(s)
		}
	}
	return ""
}

func (d *Context) PushGlobalObject() {
	C.duk_push_global_object(d.duk_context)
}

func (d *Context) GetNumber(i int) float64 {
	return float64(C.duk_get_number(d.duk_context, C.duk_idx_t(i)))
}

func (d *Context) PushNumber(i float64) {
	C.duk_push_number(d.duk_context, C.duk_double_t(i))
}

func (d *Context) PutPropString(i int, prop string) {
	str := C.CString(prop)
	defer C.free(unsafe.Pointer(str))
	C.duk_put_prop_string(d.duk_context, C.duk_idx_t(i), str)
}

func (d *Context) DestroyHeap() {
	C.duk_destroy_heap(d.duk_context)
}
