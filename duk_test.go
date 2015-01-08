package duk

import "reflect"
import "testing"

func TestEvalString(t *testing.T) {
	ctx := NewContext()
	ctx.EvalString(`"Duktape from Golang!"`)
	expect(t, ctx.GetType(-1).IsString(), true)
	expect(t, ctx.GetString(-1), "Duktape from Golang!")
	ctx.DestroyHeap()
}

func TestGoFunc(t *testing.T) {
	ctx := NewContext()
	ctx.PushGlobalObject()

	// cfn := C.duk_c_function(unsafe.Pointer(&GO_TO_C_FUNC))
	// C.duk_push_c_function(ctx.duk_context, cfn, C.duk_idx_t(2))
	//
	// ctx.PushGoFunc(func(c *Context) int {
	// 	return 0 // returns EcmaScript undefined
	// }, -1)
	// ctx.PutPropString(-2, "adder")
	ctx.pushAddFunction()
	ctx.PutPropString(-2, "adder")
	ctx.Pop()

	ctx.EvalString(`""+adder(2, 3)`)
	expect(t, ctx.GetType(-1).IsString(), true)
	expect(t, ctx.GetString(-1), "5")
	ctx.DestroyHeap()
}

// 	ctx.Pop()
// 	ctx.EvalString(`"Hello Oleg!"`)
// 	fmt.Printf("GetString: %s\n", ctx.GetString(-1))
// 	ctx.Pop()
//
// 	// непонятка с определением функции в Go
// 	ctx.Pop() // pop global
// 	ctx.EvalString(`print(adder.toString();`)
// 	ctx.EvalString(`adder(1,2);`)
// 	// ctx.EvalString(`print('2+3=' + adder(2, 3));`)
// }

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
