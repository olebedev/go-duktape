package duktape

/*
#cgo linux LDFLAGS: -lm

# include "duktape.h"
extern duk_ret_t goCall(duk_context *ctx);
*/
import "C"
import "errors"
import "unsafe"

const goFuncProp = "goFuncData"
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
		// TODO: "A caller SHOULD implement a fatal error handler in most applications."
		duk_context: C.duk_create_heap(nil, nil, nil, nil, nil),
	}
	return ctx
}

//export goCall
func goCall(ctx unsafe.Pointer) C.duk_ret_t {
	d := &Context{ctx}

	/*
	d.PushContextDump()
	log.Printf("goCall context: %s", d.GetString(-1))
	d.Pop()
        */

	d.PushCurrentFunction()
	d.GetPropString(-1, goFuncProp)
	if ! Type(d.GetType(-1)).IsPointer() {
		d.Pop2()
		return C.duk_ret_t(C.DUK_RET_TYPE_ERROR)
	}
	fd := (*GoFuncData)(d.GetPointer(-1))
	d.Pop2()

	return C.duk_ret_t(fd.f(d))
}

type GoFunc func (d *Context) int
type GoFuncData struct {
	f GoFunc
}

// Push goCall with its "goFuncData" property set to fd
func (d *Context) pushGoFunc(fd *GoFuncData) {
	d.PushCFunction((*[0]byte)(C.goCall), C.DUK_VARARGS)
	d.PushPointer(unsafe.Pointer(fd))
	d.PutPropString(-2, goFuncProp)
}


type MethodSuite map[string]GoFunc

func (d *Context) EvalWith(source string, suite MethodSuite) error {
	if err := d.PevalString(source); err != 0 {
		return errors.New(d.SafeToString(-1))
	}

	d.PushObject()

	// Make sure we keep references to all the GoFuncData
	suiteData := make(map[string]*GoFuncData)
	for prop, f := range suite {
		suiteData[prop] = &GoFuncData{f}
	}

	for prop, fd := range suiteData {
		d.pushGoFunc(fd)
		d.PutPropString(-2, prop)
	}

	if err := d.Pcall(1); err != 0 {
		return errors.New(d.SafeToString(-1))
	}

	return nil
}
