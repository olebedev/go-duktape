package duktape

import (
	"fmt"
	"time"
)

// DefineTimers defines `setTimeout`, `clearTimeout`, `setInterval`,
// `clearInterval` into global context.
func (d *Context) DefineTimers() {
	d.PushGlobalStash()
	d.PushObject()
	d.PutPropString(-2, "timers") // stash -> [ timers:{} ]
	d.Pop2()

	d.PushGlobalGoFunction("setTimeout", setTimeout)
}

func setTimeout(c *Context) int {
	id := c.pushTimer(0)
	timeout := c.ToNumber(1)
	go func(id float64) {
		<-time.After(time.Duration(timeout) * time.Millisecond)
		c.putTimer(id)
		// TODO: check if timer still exists
		// if c.GetType(-1).IsLightFunc() {
		// }
		fmt.Println("timer time is", c.GetType(-1))
		c.Pcall(0 /* nargs */)
		c.dropTimer(id)
	}(id)
	c.PushNumber(id)
	return 1
}

func (d *Context) pushTimer(index int) float64 {
	id := d.timerIndex.get()

	d.PushGlobalStash()
	d.GetPropString(-1, "timers")
	d.PushNumber(id)
	d.Dup(index) // cbk index
	d.PutProp(-3)
	d.Pop2()

	return id
}

func (d *Context) dropTimer(id float64) {
	d.PushGlobalStash()
	d.GetPropString(-1, "timers")
	d.PushNumber(id)
	d.DelProp(-2)

	d.PushContextDump()
	d.Pop()

}

func (d *Context) putTimer(id float64) {
	d.PushGlobalStash()           // -> [ timers: { <id>: { func: true } } ]
	d.GetPropString(-1, "timers") // -> [ timers: { <id>: { func: true } } }, { <id>: { func: true } ]
	d.PushNumber(id)
	d.GetProp(-1)
}
