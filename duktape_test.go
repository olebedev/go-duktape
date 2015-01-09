package duktape

import "log"
import "reflect"
import "testing"

func TestEvalString(t *testing.T) {
	ctx := NewContext()
	ctx.EvalString(`"Golang love Duktape!"`)
	expect(t, ctx.GetType(-1).IsString(), true)
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
		expect(t, r, "Wrong type of hashName argument")
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
	log.Println("start go func call test")
	var check bool
	ctx := NewContext()
	ctx.PushGoFunc("test", func(c *Context) int {
		check = !check
		return 0
	})
	expect(t, len(goFuncMap), 1)
	log.Printf("goFuncMap is: %#v", goFuncMap)
	for k, _ := range goFuncMap {
		ctx.EvalString(goFuncCallName + `('` + k + `');`)
		expect(t, check, true)
		ctx.EvalString(goFuncCallName + `('` + k + `');`)
		expect(t, check, false)
		break
	}
	ctx.DestroyHeap()
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
