package atomic

import "sync/atomic"


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