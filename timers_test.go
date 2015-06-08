package duktape

import (
	"time"

	. "gopkg.in/check.v1"
)

func (s *DuktapeSuite) TestSetTimeOut(c *C) {
	ch := make(chan struct{})
	s.ctx.DefineTimers()
	s.ctx.PushGlobalGoFunction("test", func(_ *Context) int {
		ch <- struct{}{}
		return 0
	})
	s.ctx.PevalString(`setTimeout(test, 0);`)
	<-ch
	c.Succeed()
}

func (s *DuktapeSuite) TestClearTimeOut(c *C) {
	ch := make(chan struct{}, 1) // buffered channel
	s.ctx.DefineTimers()
	s.ctx.PushGlobalGoFunction("test", func(_ *Context) int {
		ch <- struct{}{}
		return 0
	})
	s.ctx.PevalString(`
		var id = setTimeout(test, 0);
		clearTimeout(id);
	`)
	<-time.After(2 * time.Millisecond)
	select {
	case <-ch:
		c.Fail()
	default:
		c.Succeed()
	}
}

func (s *DuktapeSuite) TestSetInterval(c *C) {
	ch := make(chan struct{}, 5)
	s.ctx.DefineTimers()
	s.ctx.PushGlobalGoFunction("test", func(_ *Context) int {
		ch <- struct{}{}
		return 0
	})

	s.ctx.PevalString(`var id = setInterval(test, 0);`)

	<-ch
	<-ch
	<-ch
	s.ctx.PevalString(`clearInterval(id);`)

	<-time.After(4 * time.Millisecond)
	select {
	case <-ch:
		c.Fail()
	default:
		c.Succeed()
	}
}
