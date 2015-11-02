package duktape

/*
#cgo linux LDFLAGS: -lm
#cgo CFLAGS: -DDUK_OPT_DEEP_C_STACK

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

var reFuncName = regexp.MustCompile("^[a-z_][a-z0-9_]*([A-Z_][a-z0-9_]*)*$")

const (
	goFunctionPtrProp = "\xff" + "goFunctionPtrProp"
	goContextPtrProp  = "\xff" + "goContextPtrProp"
)

type Context struct {
	*context
}

// transmute replaces the value from Context with the value of pointer
func (c *Context) transmute(p unsafe.Pointer) {
	*c = *(*Context)(p)
}

// this is a pojo containing only the values of the Context
type context struct {
	sync.Mutex
	duk_context *C.duk_context
	fnIndex     *functionIndex
	timerIndex  *timerIndex
}

// New returns plain initialized duktape context object
// See: http://duktape.org/api.html#duk_create_heap_default
func New() *Context {
	return &Context{
		&context{
			duk_context: C.duk_create_heap(nil, nil, nil, nil, nil),
			fnIndex:     newFunctionIndex(),
			timerIndex:  &timerIndex{},
		},
	}
}

func contextFromPointer(ctx *C.duk_context) *Context {
	return &Context{&context{duk_context: ctx}}
}

// PushGlobalGoFunction push the given function into duktape global object
// Returns non-negative index (relative to stack bottom) of the pushed function
// also returns error if the function name is invalid
func (d *Context) PushGlobalGoFunction(name string, fn func(*Context) int) (int, error) {
	if !reFuncName.MatchString(name) {
		return -1, errors.New("Malformed function name '" + name + "'")
	}

	d.PushGlobalObject()
	idx := d.PushGoFunction(fn)
	d.PutPropString(-2, name)
	d.Pop()

	return idx, nil
}

// PushGoFunction push the given function into duktape stack, returns non-negative
// index (relative to stack bottom) of the pushed function
func (d *Context) PushGoFunction(fn func(*Context) int) int {
	funPtr := d.fnIndex.Add(fn)
	ctxPtr := unsafe.Pointer(d)

	idx := d.PushCFunction((*[0]byte)(C.goFunctionCall), C.DUK_VARARGS)
	d.PushCFunction((*[0]byte)(C.goFinalizeCall), 1)
	d.PushPointer(funPtr)
	d.PutPropString(-2, goFunctionPtrProp)
	d.PushPointer(ctxPtr)
	d.PutPropString(-2, goContextPtrProp)
	d.SetFinalizer(-2)

	d.PushPointer(funPtr)
	d.PutPropString(-2, goFunctionPtrProp)
	d.PushPointer(ctxPtr)
	d.PutPropString(-2, goContextPtrProp)

	return idx
}

//export goFunctionCall
func goFunctionCall(cCtx *C.duk_context) C.duk_ret_t {
	d := contextFromPointer(cCtx)

	funPtr, ctxPtr := d.getFunctionPtrs()
	d.transmute(ctxPtr)

	result := d.fnIndex.Get(funPtr)(d)

	return C.duk_ret_t(result)
}

//export goFinalizeCall
func goFinalizeCall(cCtx *C.duk_context) {
	d := contextFromPointer(cCtx)

	funPtr, ctxPtr := d.getFunctionPtrs()
	d.transmute(ctxPtr)

	d.fnIndex.Delete(funPtr)
}

func (d *Context) getFunctionPtrs() (funPtr, ctxPtr unsafe.Pointer) {
	d.PushCurrentFunction()
	d.GetPropString(-1, goFunctionPtrProp)
	funPtr = d.GetPointer(-1)

	d.Pop()

	d.GetPropString(-1, goContextPtrProp)
	ctxPtr = d.GetPointer(-1)

	d.Pop2()
	return
}

// Destroy destroy all the references to the functions and freed the pointers
func (d *Context) Destroy() {
	d.fnIndex.Destroy()
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

type timerIndex struct {
	c float64
	sync.Mutex
}

func (t *timerIndex) get() float64 {
	t.Lock()
	defer t.Unlock()
	t.c++
	return t.c
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

func (i *functionIndex) Destroy() {
	i.Lock()
	defer i.Unlock()

	for ptr, _ := range i.functions {
		delete(i.functions, ptr)
		C.free(ptr)
	}
}
