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

// from duktape examples
func TestTestFunc(t *testing.T) {
	ctx := NewContext()
	ctx.PushGlobalObject()
	ctx.pushTestFunc()
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
