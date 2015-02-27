// +build !windows
package duktape

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func testCallError() {
	ctx := NewContext()
	defer ctx.DestroyHeap()
	ctx.Error(ErrType, "Nope: ", 500)
}

func testCallErrorf() {
	ctx := NewContext()
	defer ctx.DestroyHeap()
	ctx.Error(ErrType, "Nope: %d", 0xdeadbeef)
}

func waitForAbort(t *testing.T, sig <-chan os.Signal, expectedSig os.Signal) {
	const wait = 500 * time.Millisecond
	select {
	case s := <-sig:
		if s != expectedSig {
			t.Error("Failed to receive %v, but did receive %v", expectedSig, s)
			return
		}
		t.Logf("Received expected signal: %v == %v", expectedSig, s)
	case <-time.After(wait):
		t.Errorf("Failed to receive %v after %v", expectedSig, wait)
	}
}

// TestErrorAbort tests for whether SIGABRT is received when Context.Error[f]
// is called.
func TestErrorAbort(t *testing.T) {
	abrt := os.Signal(syscall.SIGABRT)
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, abrt)
	defer signal.Stop(sig)

	go testCallError()
	waitForAbort(t, sig, abrt)

	go testCallErrorf()
	waitForAbort(t, sig, abrt)
}
