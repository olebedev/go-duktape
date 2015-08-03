# Duktape bindings for Go(Golang) [![wercker status](https://app.wercker.com/status/1ce7671d7223880e967bf8a81b96341d/s/master "wercker status")](https://app.wercker.com/project/bykey/1ce7671d7223880e967bf8a81b96341d)
[Duktape](http://duktape.org/index.html) is a thin, embeddable javascript engine.
Most of the [api](http://duktape.org/api.html) is implemented.
The exceptions are listed [here](https://github.com/olebedev/go-duktape/blob/master/api.go#L1679).

### Usage
```go
package main

import "fmt"
import "github.com/olebedev/go-duktape"

func main() {
  ctx := duktape.New()
  ctx.EvalString(`2 + 3`)
  result := ctx.GetNumber(-1)
  ctx.Pop()
  fmt.Println("result is:", result)
}
```

### Go specific notes

Bindings between Go and Javascript contexts are not fully functional.
However, binding a Go function to the Javascript context is available:
```go
package main

import "fmt"
import "github.com/olebedev/go-duktape"

func main() {
  ctx := duktape.New()
  ctx.PushGlobalGoFunction("log", func(c *duktape.Context) int {
    fmt.Println(c.SafeToString(-1))
    return 0
  })
  ctx.EvalString(`log('Go lang Go!')`)
}
```
than run it.
```bash
$ go run *.go
Go lang Go!
$
```

### Timers

There is a method to inject to the global scope implemented in Go timers:
```go
package main

import "fmt"
import "github.com/olebedev/go-duktape"

func main() {
  ctx := duktape.New()

  // Let's inject `setTimeout`, `setInterval`, `clearTimeout`,
  // `clearInterval` into global scope.
  ctx.DefineTimers()

  ch := make(chan string)
  ctx.PushGlobalGoFunction("second", func(_ *Context) int {
    ch <- "second step"
    return 0
  })
  ctx.PevalString(`
    setTimeout(second, 0);
    print('first step');
  `)
  fmt.Println(<-ch)
}
```
than run it
```bash
$ go run *.go
first step
second step
$
```


### Status

The package is not fully tested, so be careful.


### Contribution

Pull requests are welcome!  
__Convention:__ fork the repository and make changes on your fork in a feature branch.
