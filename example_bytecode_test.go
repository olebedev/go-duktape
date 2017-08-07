package duktape

import (
	"fmt"
	"reflect"
	"unsafe"
	"log"
)

func ExampleContext_LoadFunction() {
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

	// Parenthesis is necessary.
	js := "(function dump_from() { return 'It\\'s alive!'; })"

	ctxSerialize := New()

	// Compile js to duktape function and put it on the context stack
	ctxSerialize.EvalLstring(js, len(js))

	// Transmute duktape function on the top of the context stack into serializable duktape bytecode
	ctxSerialize.DumpFunction()

	// void *duk_get_buffer(duk_context *ctx, duk_idx_t idx, duk_size_t *out_size);
	// typedef size_t duk_size_t;
	// duk_size_t equivalent in Go is uint, however slice len is of a type int
	// and I don't think that the limit of max(int) can be reached so for the sake of simplicity
	// lets assume that bytecode size is always can be contained in int
	rawmem, bufsize := ctxSerialize.GetBuffer(-1)
	// Check for null is necessary because duk_get_buffer can return NULL.
	if uintptr(rawmem) == uintptr(0) {
		log.Fatalf("Can't interpret bytecode dump as a valid, non-empty buffer.")
	}
	rawmemslice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(rawmem), Len: int(bufsize), Cap: int(bufsize)}))

	// Creating another slice for bytecode is necessary because rawmemslice pointing at memory that belongs to
	// current context. That memory will be freed during execution of DestroyHeap().
	bytecode := make([]byte, bufsize)
	copy(bytecode, rawmemslice)

	// To prevent memory leaks, don't forget to clean up after
	// yourself when you're done using a context.
	ctxSerialize.DestroyHeap()

	ctxDeserialize := New()

	//creating buffer on the context stack
	rawmem = ctxDeserialize.PushBuffer(len(bytecode), false)
	if uintptr(rawmem) == uintptr(0) {
		log.Fatalf("Can't push buffer to the context stack.")
	}

	//copying bytecode into the created buffer
	rawmemslice = *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(rawmem), Len: len(bytecode), Cap: len(bytecode)}))
	copy(rawmemslice, bytecode)

	// Transmute duktape bytecode into duktape function
	ctxDeserialize.LoadFunction()

	// Call the function on top of a stack, example function doesn't have arguments.
	ctxDeserialize.Call(0)

	//example function return value is string
	retval := ctxDeserialize.GetString(-1)

	// To prevent memory leaks, don't forget to clean up after
	// yourself when you're done using a context.
	ctxDeserialize.DestroyHeap()

	fmt.Println(retval)

	// Output:
	// It's alive!
}
