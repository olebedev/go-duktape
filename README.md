# Duktape bindings for Go(Golang)

[![wercker status](https://app.wercker.com/status/3a5bb2e639a4b4efaf4c8bf7cab7442d/s "wercker status")](https://app.wercker.com/project/bykey/3a5bb2e639a4b4efaf4c8bf7cab7442d)
[![Travis status](https://travis-ci.org/olebedev/go-duktape.svg?branch=v3)](https://travis-ci.org/olebedev/go-duktape)
[![Appveyor status](https://ci.appveyor.com/api/projects/status/github/olebedev/go-duktape?branch=v3&svg=true)](https://ci.appveyor.com/project/olebedev/go-duktape/branch/v3)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/olebedev/go-duktape?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

[Duktape](http://duktape.org/index.html) is a thin, embeddable javascript engine.
Most of the [api](http://duktape.org/api.html) is implemented.
The exceptions are listed [here](https://github.com/olebedev/go-duktape/blob/master/api.go#L1566).

### Usage

The package is fully go-getable, no need to install any external C libraries.  
So, just type `go get gopkg.in/olebedev/go-duktape.v3` to install.


```go
package main

import "fmt"
import "gopkg.in/olebedev/go-duktape.v3"

func main() {
  ctx := duktape.New()
  ctx.PevalString(`2 + 3`)
  result := ctx.GetNumber(-1)
  ctx.Pop()
  fmt.Println("result is:", result)
  // To prevent memory leaks, don't forget to clean up after
  // yourself when you're done using a context.
  ctx.DestroyHeap()
}
```

### Go specific notes

Bindings between Go and Javascript contexts are not fully functional.
However, binding a Go function to the Javascript context is available:
```go
package main

import "fmt"
import "gopkg.in/olebedev/go-duktape.v3"

func main() {
  ctx := duktape.New()
  ctx.PushGlobalGoFunction("log", func(c *duktape.Context) int {
    fmt.Println(c.SafeToString(-1))
    return 0
  })
  ctx.PevalString(`log('Go lang Go!')`)
}
```
then run it.
```bash
$ go run *.go
Go lang Go!
$
```

### Timers

There is a method to inject timers to the global scope:
```go
package main

import "fmt"
import "gopkg.in/olebedev/go-duktape.v3"

func main() {
  ctx := duktape.New()

  // Let's inject `setTimeout`, `setInterval`, `clearTimeout`,
  // `clearInterval` into global scope.
  ctx.PushTimers()

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
then run it
```bash
$ go run *.go
first step
second step
$
```
### Native constructor
```
package main

import (
  "github.com/olebedev/go-duktape"
  "fmt"
  "time"
)

func printName (c *duktape.Context) int {
  c.PushThis()
  c.GetPropString(-1, "name")
  fmt.Printf( "My name is: %v\n", c.SafeToString(-1) )
  return 0
}

func myObjectConstructor (c *duktape.Context) int {

  if c.IsConstructorCall() == false {
    return duktape.ErrRetType
  }

  c.PushThis()
  c.Dup( 0 )
  c.PutPropString(-2, "name")
  return 0
}

func main() {

  start := time.Now()

  ctx := duktape.New()

  /*
  function MyObject(name) {
      // When called as new MyObject(), "this" will be bound to the default
      // instance object here.

      this.name = name;

      // Return undefined, which causes the default instance to be used.
  }

  // For an ECMAScript function an automatic MyObject.prototype value will be
  // set.  That object will also have a .constructor property pointing back to
  // Myobject.  You can add instance methods to MyObject.prototype.

  MyObject.prototype.printName = function () {
      print('My name is: ' + this.name);
  };

  var obj = new MyObject('test object');
  obj.printName();  // My name is: test object
  */

  ctx.PushGoFunction(myObjectConstructor)
  ctx.PushObject()
  ctx.PushGoFunction(printName)
  ctx.PutPropString(-2, "printName")
  ctx.PutPropString(-2, "prototype")
  ctx.PutGlobalString("MyObject")

  /*
  make this
  */
  ctx.EvalString( "var a = new MyObject('test name one');" )
  ctx.EvalString( "a.printName();" )


  /*
  or make this
  */
  ctx.EvalString( "new MyObject('test name two');" )
  ctx.GetPropString(-1, "printName")
  ctx.Dup(-2)
  ctx.CallMethod(0)
  ctx.Pop()
  ctx.Pop()
  ctx.GetGlobalString("MyObject")


  ctx.DestroyHeap()
  elapsed := time.Since(start)
  fmt.Printf("total time: %s\n", elapsed)
}
```

Also you can `FlushTimers()`.

### Command line tool

Install `go get gopkg.in/olebedev/go-duktape.v3/...`.  
Execute file.js: `$GOPATH/bin/go-duk file.js`.

### Benchmarks
| prog        | time  |
| ------------|-------|
|[otto](https://github.com/robertkrimen/otto)|200.13s|
|[anko](https://github.com/mattn/anko)|231.19s|
|[agora](https://github.com/PuerkitoBio/agora/)|149.33s|
|[GopherLua](https://github.com/yuin/gopher-lua/)|8.39s|
|**go-duktape**|**9.80s**|

More details are [here](https://github.com/olebedev/go-duktape/wiki/Benchmarks).

### Status

The package is not fully tested, so be careful.


### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.
__Convention:__ fork the repository and make changes on your fork in a feature branch.
