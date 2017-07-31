// Example of serialization duktape function to a byte slice; deserialization from byte slice to duktape
// function and eval in another duktape context.
//
// Usecase: get byte slice from duktape function (to possibly later save it to the file, send by network and etc.)
// Please, keep in mind restrictions for a deserialized duktape function from bytecode:
// https://github.com/svaarala/duktape/blob/master/doc/bytecode.rst
//
//Excerpts:
//	When to use bytecode dump/load
//		There are two main motivations for using bytecode dump/load:
//			-	Performance
//			-	Obfuscation

package duktape_test

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	"gopkg.in/olebedev/go-duktape.v2"
)

func ExampleContext() {
	// Parenthesis is necessary.
	// DeserializeAndRunInNewContext assuming that script in jsfunc doesn't have arguments and return string.
	jsfunc := "(function dump_from() { return 'It\\'s alive!'; })"

	bytecode, err := getSerializedFunc(jsfunc)
	if err != nil {
		log.Fatalf("Can't serialize '%s' to bytecode, err: %q.", jsfunc, err)
	}

	retval, err := deserializeAndRunInNewContext(bytecode)
	if err != nil {
		log.Fatalf("Can't deserialize and run '%s' from bytecode, err: %q.", jsfunc, err)
	}

	fmt.Print(retval)
	// Output:
	// It's alive!
}

func getSerializedFunc(script string) ([]byte, error) {
	ctx := duktape.New()

	ctx.EvalLstring(script, len(script))

	// Transmute function into serializable duktape bytecode
	ctx.DumpFunction()

	// void *duk_get_buffer(duk_context *ctx, duk_idx_t idx, duk_size_t *out_size);
	// typedef size_t duk_size_t;
	// duk_size_t equivalent in Go is uint, however slice len is of a type int
	// and I don't think that the limit of max(int) can be reached so for the sake of simplicity
	// lets assume that bytecode size is always can be contained in int
	var sz int
	rawmem := ctx.GetBuffer(-1, &sz)
	// Check for null is necessary because duk_get_buffer can return NULL.
	if uintptr(rawmem) == uintptr(0) {
		return nil, errors.New("Can't interpret bytecode dump as a valid, non-empty buffer.")
	}
	rawmemslice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(rawmem), Len: sz, Cap: sz}))

	// Creating another slice for return is necessary because rawmemslice pointing at memory that belongs to
	// current context. That memory will be freed during execution of DestroyHeap().
	retval := make([]byte, sz)
	copy(retval, rawmemslice)

	// To prevent memory leaks, don't forget to clean up after
	// yourself when you're done using a context.
	ctx.DestroyHeap()

	return retval, nil
}

func deserializeAndRunInNewContext(bc []byte) (string, error) {
	ctx := duktape.New()

	//creating buffer on the context stack
	rawmem := ctx.PushBuffer(len(bc), false)
	if uintptr(rawmem) == uintptr(0) {
		return "", errors.New("Can't push buffer to the context stack.")
	}

	//copying bytecode into the created buffer
	rawmemslice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(rawmem), Len: len(bc), Cap: len(bc)}))
	copy(rawmemslice, bc)

	// Transmute duktape bytecode into duktape function
	ctx.LoadFunction()

	// Call the function on top of a stack, example function doesn't have arguments.
	ctx.Call(0)

	//example function return value is string
	retval := ctx.GetString(-1)
	log.Printf("Return value is: %s", retval)

	// To prevent memory leaks, don't forget to clean up after
	// yourself when you're done using a context.
	ctx.DestroyHeap()

	return retval, nil
}
