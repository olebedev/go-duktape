package duktape

import "reflect"
import "testing"

func TestEvalString(t *testing.T) {
	ctx := NewContext()
	ctx.EvalString(`"Golang love Duktape!"`)
	expect(t, Type(ctx.GetType(-1)).IsString(), true)
	expect(t, ctx.GetString(-1), "Golang love Duktape!")
	ctx.DestroyHeap()
}

func TestGoFuncCallWOArgs(t *testing.T) {
	defer func() {
		r := recover()
		expect(t, r, "Go function call without arguments is not supported")
	}()
	ctx := NewContext()
	ctx.EvalString(goFuncCallName + `();`)
	ctx.DestroyHeap()
}

func TestGoFuncCallWWrongArg(t *testing.T) {
	defer func() {
		r := recover()
		expect(t, r, "Wrong type of function's key argument")
	}()
	ctx := NewContext()
	ctx.EvalString(goFuncCallName + `(0); // first argument must be a string`)
	ctx.DestroyHeap()
}

func TestGoFuncCallWWrongFuncName(t *testing.T) {
	defer func() {
		r := recover()
		expect(t, r, "Unimplemented")
	}()
	ctx := NewContext()
	ctx.EvalString(goFuncCallName + `('noFunctionName'); // this func is not defined`)
	ctx.DestroyHeap()
}

func TestGofuncCall(t *testing.T) {
	var check bool
	ctx := NewContext()
	ctx.PushGoFunc("test", func(c *Context) int {
		check = !check
		return 0
	})
	expect(t, len(goFuncMap), 1)
	for k, _ := range goFuncMap {
		ctx.EvalString(goFuncCallName + `('` + k + `');`)
		expect(t, check, true)
		ctx.EvalString(goFuncCallName + `('` + k + `');`)
		expect(t, check, false)
		break
	}
	ctx.DestroyHeap()
}

func expectError(t *testing.T, ctx *Context, code int, errMsg string) {
	if !ctx.IsError(-1) {
		t.Errorf("Expected Error type, got %v", ctx.GetType(-1))
	}

	if got := ctx.GetErrorCode(-1); code != code {
		t.Errorf("Expected error %#v, got %#v", code, got)
	}

	if msg := ctx.SafeToString(-1); msg != errMsg {
		t.Errorf("Expected message %q, got %q", errMsg, msg)
	}
}

func TestErrorObj(t *testing.T) {
	ctx := NewContext()
	defer ctx.DestroyHeap()
	ctx.PushErrorObject(ErrType, "Got an error thingy: ", 5)
	expectError(t, ctx, ErrType, "TypeError: Got an error thingy: 5")

	ctx = NewContext()
	defer ctx.DestroyHeap()
	ctx.PushErrorObjectf(ErrURI, "Got an error thingy: %x", 0xdeadbeef)
	expectError(t, ctx, ErrURI, "URIError: Got an error thingy: deadbeef")
}

func pushTestFunc(d *Context) {
	d.PushCFunction((*[0]byte)(testFuncPtr), (-1))
}

// from duktape examples
func TestTestFunc(t *testing.T) {
	ctx := NewContext()
	ctx.PushGlobalObject()
	pushTestFunc(ctx)
	ctx.PutPropString(-2, "adder")
	ctx.Pop()
	ctx.EvalString(`adder(2, 3);`)
	res := ctx.GetNumber(-1)
	ctx.Pop()
	expect(t, res, float64(5))
}

func TestMyAddTwo(t *testing.T) {
	ctx := NewContext()
	ctx.PushGoFunc("adder", goTestfunc)
	ctx.EvalString(`print("2 + 3 =", adder(2,3))`)
	ctx.Pop()
	ctx.EvalString(`adder(2,3)`)
	result := ctx.GetNumber(-1)
	expect(t, result, float64(5))
	ctx.DestroyHeap()
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
