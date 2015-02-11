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


func TestEvalFunc(t *testing.T) {
	ctx := NewContext()
	ctx.PevalString(`(function (x) { return x + x; })`)
	expect(t, ctx.IsCallable(-1), true)
	expect(t, Type(ctx.GetType(-1)).IsObject(), true)
	ctx.PushInt(5)
	ctx.Pcall(1)
	expect(t, ctx.GetInt(-1), 10)
	ctx.DestroyHeap()
}

func TestEvalWith(t *testing.T) {
	ctx := NewContext()

	obj := MethodSuite{
		"hi": func(d *Context) int {
			x := d.GetInt(-2)
			d.PushString("hi! " + string(48 + x))
			return 1
		},
	}
	err := ctx.EvalWith("(function(o) { return o.hi(1, 2, 3) })", obj)
	expect(t, err, nil)

	actual := ctx.GetString(-1)

	expect(t, actual, "hi! 2")

	ctx.DestroyHeap()
}


// from duktape examples


func TestMyAddTwo(t *testing.T) {
	obj := MethodSuite{
		"add": func(d *Context) int {
			top := d.GetTop()
			a := d.GetNumber(top - 2)
			b := d.GetNumber(top - 1)
			d.PushNumber(a + b)
			return 1
		},
	}

	ctx := NewContext()
	ctx.PushGlobalObject()

	// Hmm... a property value can outlive an object. look out!
	ctx.EvalWith("(function(o) { return o.add })", obj)

	ctx.PutPropString(-2, "adder")

	ctx.PevalString(`adder(2, 3);`)
	res := ctx.GetNumber(-1)
	ctx.Pop()
	expect(t, res, float64(5))
	ctx.DestroyHeap()
}


func TestGoClosure(t *testing.T) {
	sharedState := 0
	obj := MethodSuite{
		"inc": func(d *Context) int {
			sharedState++
			d.PushInt(sharedState)
			return 1
		},
		"dec": func(d *Context) int {
			sharedState--
			d.PushInt(sharedState)
			return 1
		},
	}

	ctx := NewContext()

	ctx.EvalWith(`
            (function(o) {
                 o.inc();
                 o.inc();
                 o.dec();
                 o.inc();
                 return o.inc();
             })`, obj)
	res := ctx.GetNumber(-1)
	expect(t, res, float64(3))
	ctx.DestroyHeap()
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
