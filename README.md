# Duktape bindings for Go(Golang)
[Duktape](http://duktape.org/index.html) is thin embedded javascript engine. Mostly all apis are implemented(see [api](http://duktape.org/api.html)). Except several functions, see the list [here](https://github.com/olebedev/go-duktape/blob/master/api.go#L1294).

### Usage
```go
package main

import "fmt"
import "github.com/olebedev/go-duktape"

func main() {
  ctx := duktape.NewContext()
  ctx.EvalString(`2 + 3`)
  result = ctx.GetNumber(-1)
  ctx.Pop()
  fmt.Println("result is:", result)
}
```

### Go specific notes
There is very important that is need to be done. This is binding between Go and Javascript contexts. Now you can define written in Go function into Javascript context quite simple. Example usage:
```go
package main

import "fmt"
import "github.com/olebedev/go-duktape"

func main() {
  ctx := duktape.NewContext()
  ctx.PushGofunc("log", func(ctx *duktape.Context) int {
    fmt.Println("Go lang Go!")
  })
  ctx.EvalString(`log()`)
}
```
than run it.
```bash
$ go run
$ Go lang Go!
```

### Status
Not fully tested, be careful. Basically, there is nothing to be break. Additional functionality tested.

### Contribution
PR's are welcome!
