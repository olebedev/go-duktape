package duktape

func mustNotNil(c *Context) {
	if c.duk_context == nil {
		panic("[duktape] Context does not exists!\nYou cannot call any contexts methods after `DestroyHeap()` was called.")
	}
}
