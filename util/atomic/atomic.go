package atomic

import (
	"sync/atomic"
	"sync"
)


type AtomicBool int32

func (a *AtomicBool) StoreBool(x bool) {
	var v int = 0
	if x {
		v = 1
	}
	atomic.StoreInt32((*int32)(a), int32(v))
}

func (a *AtomicBool) LoadBool() (v bool) {
	if atomic.LoadInt32((*int32)(a)) != 0 {
		v = true
	}
	return v
}

type AtomicUint8 uint32

func (a *AtomicUint8) StoreUint8(x uint8) {
	atomic.StoreUint32((*uint32)(a), uint32(x))
}

func (a *AtomicUint8) LoadUint8()  uint8 {
	return uint8(atomic.LoadUint32((*uint32)(a)))
}

type AtomicUint32 uint32

func (a *AtomicUint32) StoreUint32(x uint32) {
	atomic.StoreUint32((*uint32)(a), x)
}

func (a *AtomicUint32) LoadUint32()  uint32 {
	return uint32(atomic.LoadUint32((*uint32)(a)))
}


type AtomicString struct {
	s string
	mu sync.Mutex
}

func (a AtomicString) StoreString(x string) {
	a.mu.Lock()
	a.s = x
	a.mu.Unlock()
}

func (a AtomicString) LoadString()  string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.s
}

/*
func main() {

	var t AtomicBool

	fmt.Printf("%v, %v\n", t, t.LoadBool())
	t.StoreBool(true)
	fmt.Printf("%v, %v\n", t, t.LoadBool())
	t.StoreBool(false)
	fmt.Printf("%v, %v\n", t, t.LoadBool())
}
*/