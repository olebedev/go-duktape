package duktape

func (s *DuktypeSuite) TestSetTimeOut(c *C) {
	ch := make(chan struct{})
	s.ctx.PushGlobalGoFunction("test", func(c *duktape.Context) int {
		go func() {
			ch <- struct{}{}
		}()
		return 0
	})
	s.ctx.PevalString(`setTimeout(test, 1);`)
	<-ch
}
