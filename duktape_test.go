package duktape

import (
	"reflect"
	"testing"
)

func TestEvalString(t *testing.T) {
	ctx := Default()
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
	ctx := Default()
	ctx.EvalString(goFuncCallName + `();`)
	ctx.DestroyHeap()
}

func TestGoFuncCallWWrongArg(t *testing.T) {
	defer func() {
		r := recover()
		expect(t, r, "Wrong type of function's key argument")
	}()
	ctx := Default()
	ctx.EvalString(goFuncCallName + `(0); // first argument must be a string`)
	ctx.DestroyHeap()
}

func TestGoFuncCallWWrongFuncName(t *testing.T) {
	defer func() {
		r := recover()
		expect(t, r, "Unimplemented")
	}()
	ctx := Default()
	ctx.EvalString(goFuncCallName + `('noFunctionName'); // this func is not defined`)
	ctx.DestroyHeap()
}

func TestPushGlobalGoFunction_Call(t *testing.T) {
	var check bool
	ctx := Default()
	ctx.PushGlobalGoFunction("test", func(c *Context) int {
		check = !check
		return 0
	})
	expect(t, len(ctx.fn), 1)
	for k, _ := range ctx.fn {
		ctx.EvalString(goFuncCallName + `('` + k + `');`)
		expect(t, check, true)
		ctx.EvalString(goFuncCallName + `('` + k + `');`)
		expect(t, check, false)
		break
	}

	ctx.DestroyHeap()
}

func TestPushGlobalGoFunction_Finalize(t *testing.T) {
	ctx := Default()
	ctx.PushGlobalGoFunction("test", func(c *Context) int {
		return 0
	})

	expect(t, len(ctx.fn), 1)
	ctx.EvalString("test = undefined")
	ctx.Gc(0)
	expect(t, len(ctx.fn), 0)

	ctx.DestroyHeap()
}

func TestPushGoFunction_Call(t *testing.T) {
	var check bool
	ctx := Default()
	name, err := ctx.PushGoFunction(func(c *Context) int {
		check = !check
		return 0
	})

	expect(t, err, nil)
	expect(t, len(ctx.fn), 1)

	ctx.EvalString(goFuncCallName + `('` + name + `');`)
	expect(t, check, true)
	ctx.EvalString(goFuncCallName + `('` + name + `');`)
	expect(t, check, false)

	ctx.DestroyHeap()
}

func TestErrorObj(t *testing.T) {
	ctx := Default()
	defer ctx.DestroyHeap()
	ctx.PushErrorObject(ErrType, "Got an error thingy: ", 5)
	expectError(t, ctx, ErrType, "TypeError: Got an error thingy: 5")

	ctx = Default()
	defer ctx.DestroyHeap()
	ctx.PushErrorObjectf(ErrURI, "Got an error thingy: %x", 0xdeadbeef)
	expectError(t, ctx, ErrURI, "URIError: Got an error thingy: deadbeef")
}

func goTestfunc(ctx *Context) int {
	top := ctx.GetTop()
	a := ctx.GetNumber(top - 2)
	b := ctx.GetNumber(top - 1)
	ctx.PushNumber(a + b)
	return 1
}

func TestMyAddTwo(t *testing.T) {
	ctx := Default()
	ctx.PushGlobalGoFunction("adder", goTestfunc)
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
