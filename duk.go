package duk

/*
Нужно сделать так:
Объявить в go функцию, которая может делать замыкание на уже реальную функцию
func(*Context) int. Эскпортировать её. В Си сделать враппер, который вызывает эту
go функцию О_о
*/

/*
# include "duktape.h"
duk_c_function go_duk_c_function();
static void go_duk_eval_string(duk_context *ctx, const char *str) {
  return duk_eval_string(ctx, str);
}
*/
import "C"
import "unsafe"

//export toDukCFunction
func toDukCFunction(f func(*Context) int) *unsafe.Pointer {
	fn := func(pnt unsafe.Pointer) C.duk_ret_t {
		return C.duk_ret_t(f(&Context{duk_context: pnt}))
	}
	return (*unsafe.Pointer)(unsafe.Pointer(&fn))
}

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

func (d *Context) PushGoFunc(fn func(*Context) int, nargs int) {
	// TODO
	// wrap := func(pnt unsafe.Pointer) C.duk_ret_t {
	// 	return C.duk_ret_t(fn(&Context{duk_context: pnt}))
	// }
}

func (d *Context) pushAddFunction() {
	C.duk_push_c_function(
		d.duk_context,
		C.go_duk_c_function(),
		C.duk_idx_t(2),
	)
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

// Returns initialized duktape context object
func NewContext() *Context {
	return &Context{
		duk_context: C.duk_create_heap(nil, nil, nil, nil, nil),
	}
}
