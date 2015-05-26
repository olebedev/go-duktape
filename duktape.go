package duktape

/*
#cgo linux LDFLAGS: -lm

# include "duktape.h"
extern duk_ret_t goFunctionCall(duk_context *ctx);
extern void goFinalizeCall(duk_context *ctx);

*/
import "C"
import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"unsafe"
)

var (
	reFuncName = regexp.MustCompile("^[a-z_][a-z0-9_]*([A-Z_][a-z0-9_]*)*$")
	fnIndex    = newFunctionIndex()
)

const goFunctionPtrProp = "_go_function_ptr"

type Context struct {
	duk_context unsafe.Pointer
}

// New returns plain initialized duktape context object
// See: http://duktape.org/api.html#duk_create_heap_default
func New() *Context {
	return &Context{
		duk_context: C.duk_create_heap(nil, nil, nil, nil, nil),
	}
}

// PushGlobalGoFunction push the given function into duktape global object
func (d *Context) PushGlobalGoFunction(name string, fn func(*Context) int) error {
	if !reFuncName.MatchString(name) {
		return errors.New("Malformed function name '" + name + "'")
	}

	d.PushGlobalObject()
	if err := d.PushGoFunction(fn); err != nil {
		return err
	}

	d.PutPropString(-2, name)
	d.Pop()

	return nil
}

// PushGoFunction push the given function into duktape stack
func (d *Context) PushGoFunction(fn func(*Context) int) error {
	ptr := fnIndex.Add(fn)

	d.PushCFunction((*[0]byte)(C.goFunctionCall), C.DUK_VARARGS)
	d.PushCFunction((*[0]byte)(C.goFinalizeCall), 1)
	d.PushPointer(ptr)
	d.PutPropString(-2, goFunctionPtrProp)
	d.SetFinalizer(-2)

	d.PushPointer(ptr)
	d.PutPropString(-2, goFunctionPtrProp)

	return nil
}

//export goFunctionCall
func goFunctionCall(ctx unsafe.Pointer) C.duk_ret_t {
	d := &Context{duk_context: ctx}
	ptr := d.getGoFunctionPtr()

	return C.duk_ret_t(fnIndex.Get(ptr)(d))
}

//export goFinalizeCall
func goFinalizeCall(ctx unsafe.Pointer) {
	d := &Context{duk_context: ctx}
	ptr := d.getGoFunctionPtr()

	fnIndex.Delete(ptr)
}

func (d *Context) getGoFunctionPtr() unsafe.Pointer {
	d.PushCurrentFunction()
	d.GetPropString(-1, goFunctionPtrProp)
	defer func() {
		d.Pop2()
	}()

	if !d.IsPointer(-1) {
		return nil
	}

	return d.GetPointer(-1)
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

type functionIndex struct {
	functions map[unsafe.Pointer]func(*Context) int
	sync.Mutex
}

func newFunctionIndex() *functionIndex {
	return &functionIndex{
		functions: make(map[unsafe.Pointer]func(*Context) int, 0),
	}
}

func (i *functionIndex) Add(fn func(*Context) int) unsafe.Pointer {
	ptr := C.malloc(1)

	i.Lock()
	defer i.Unlock()
	i.functions[ptr] = fn

	return ptr
}

func (i *functionIndex) Get(ptr unsafe.Pointer) func(*Context) int {
	i.Lock()
	defer i.Unlock()

	return i.functions[ptr]
}

func (i *functionIndex) Delete(ptr unsafe.Pointer) {
	i.Lock()
	defer i.Unlock()

	delete(i.functions, ptr)
	C.free(ptr)
}
