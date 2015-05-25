package duktape

import "testing"

func TestPevalString(t *testing.T) {
	ctx := Default()
	err := ctx.PevalString("var = 'foo';")
	expect(t, err.(*Error).Type, "SyntaxError")
	ctx.DestroyHeap()
}

func TestPevalFile(t *testing.T) {
	ctx := Default()
	err := ctx.PevalFile("foo.js")
	expect(t, err.(*Error).Message, "no sourcecode")
	ctx.DestroyHeap()
}

func TestPcompileString(t *testing.T) {
	ctx := Default()
	err := ctx.PcompileString(CompileFunction, "foo")
	expect(t, err.(*Error).Type, "SyntaxError")
	expect(t, err.(*Error).LineNumber, 1)
	ctx.DestroyHeap()
}
