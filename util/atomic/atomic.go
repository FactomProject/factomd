package atomic

import (
"sync/atomic"
"sync"
"runtime"
"fmt"
"strings"
"time"
)

type AtomicBool int32

func (a *AtomicBool) Store(x bool) {
	var v int = 0
	if x {
		v = 1
	}
	atomic.StoreInt32((*int32)(a), int32(v))
}

func (a *AtomicBool) Load() (v bool) {
	if atomic.LoadInt32((*int32)(a)) != 0 {
		v = true
	}
	return v
}

type AtomicUint8 uint32

func (a *AtomicUint8) Store(x uint8) {
	atomic.StoreUint32((*uint32)(a), uint32(x))
}

func (a *AtomicUint8) Load() uint8 {
	return uint8(atomic.LoadUint32((*uint32)(a)))
}

type AtomicUint32 uint32

func (a *AtomicUint32) Store(x uint32) {
	atomic.StoreUint32((*uint32)(a), x)
}

func (a *AtomicUint32) Load() uint32 {
	return uint32(atomic.LoadUint32((*uint32)(a)))
}

type AtomicInt int64

func (a *AtomicInt) Store(x int) {
	atomic.StoreInt64((*int64)(a), int64(x))
}

func (a *AtomicInt) Load() int {
	return int(atomic.LoadInt64((*int64)(a)))
}

type AtomicInt64 int64

func (a *AtomicInt64) Store(x int64) {
	atomic.StoreInt64((*int64)(a), x)
}

func (a *AtomicInt64) Load() int64 {
	return atomic.LoadInt64((*int64)(a))
}

type AtomicString struct {
	s  string
	mu sync.Mutex
}

func (a AtomicString) Store(x string) {
	a.mu.Lock()
	a.s = x
	a.mu.Unlock()
}

func (a AtomicString) Load() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.s
}

/*
func main() {

	var t AtomicBool

	fmt.Printf("%v, %v\n", t, t.Load())
	t.Store(true)
	fmt.Printf("%v, %v\n", t, t.Load())
	t.Store(false)
	fmt.Printf("%v, %v\n", t, t.Load())
}
*/

// Hacky debugging stuff... probably not a great home for it

func Goid() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	s := string(buf[:n])
	idField := s[:strings.Index(s,"[")]
	return idField
}

func WhereAmIString(msg string, depth int) string {
	_, fn, line, _ := runtime.Caller(depth)
	return fmt.Sprintf("%10v [%s] %s:%d", Goid(), msg, fn, line)
}

func WhereAmI(msg string, depth int) {
	fmt.Println("\n"+WhereAmIString(msg, depth+1))
}

type DebugMutex struct {
	name AtomicString
	mu   sync.Mutex
	lock AtomicBool
}

 var flag bool = false

func (c *DebugMutex) Lock() {
	if (c.lock.Load()) {
		WhereAmI("Already Locked:"+c.name.Load(), 2)
		if flag {
			// Make a timer to whine if I am starving!
			done := make(chan struct{})
			go func() {
				for {
					for i := 0; i < 30; i++ {
						select {
						case <-done:
							return
						default:
							time.Sleep(100 * time.Millisecond)
						}
					}
					WhereAmI(c.name.Load()+" Lock starving!\n", 2)
				}
			}()
			defer func() { done <- struct{}{} }() // End the timer when I get the lock
		}
	}
	// It is possible to loose the lock after the check and before here and starve anyway
	c.mu.Lock()
	c.lock.Store(true)
	// set the name of the lock the first time it is acquired
	if (c.name.Load() == "") {
		c.name.Store(WhereAmIString("DebugMutex ", 1))
	}
}
func (c *DebugMutex) Unlock() {
	if (c.lock.Load()==false) {
		WhereAmI("Already Unlocked:"+c.name.Load(), 2)
		panic("Double Unlock")
	}
	c.lock.Store(false)
	c.mu.Unlock()
}


func main() {
	fmt.Println("Begin Main")

	var l  DebugMutex

	go func() {
		//		fmt.Println("Start 1")
		for i := 0; i < 20; i++ {
			l.Lock()
			fmt.Printf("[%d]", i)
			l.Unlock()
			time.Sleep(2 * time.Second)
		}
	}()

	//	fmt.Println("Start main")
	for i := 0; i < 20; i++ {
		fmt.Printf("<%d>", i)
		if (i == 5) {
			l.Lock()
		}
		if (i == 15) {
			l.Unlock()
		}
		time.Sleep(1 * time.Second)

	}
}

/*
func (c *DebugMutex) Lock() {
	b := atomic.CompareAndSwapInt32(&c.lock, 0, 1)
	if (!b) {
		WhereAmI("Already Locked:"+c.name.Load(), 2)
		// Make a timer to whine if I am starving!
		done := make(chan struct{})
		go func() {
			for {
				for i := 0; i < 30; i++ {
					select {
					case <-done:
						return
					default:
						time.Sleep(100 * time.Millisecond)
						//						fmt.Printf("+")
					}
				}
				WhereAmI("Lock Starving:"+c.name.Load(), 2)
			}
		}()
		defer func() { done <- struct{}{} }() // End the timer when I get the lock
		for {
			b := atomic.CompareAndSwapInt32(&c.lock, 0, 1)
			if (b) {
				//				fmt.Printf("!")
				break
			}
			//			fmt.Printf(".")
			time.Sleep(100 * time.Millisecond)
		} // sit and spin }
	}
	c.mu.Lock()
	// set the name of the lock the first time it is acquired
	if (c.name.Load() == "") {
		c.name.Store(WhereAmIString("DebugMutex ", 1))
	}
}

func (c *DebugMutex) Unlock() {
    c.mu.Unlock()
	b := atomic.CompareAndSwapInt32(&c.lock, 1, 0)
	if (!b) {
		WhereAmI("Already Unlocked:"+c.name.Load(), 2)
		panic("Double Unlock")
	}
}



func main() {
	fmt.Println("Begin Main")

	var l  DebugMutex

	go func() {
		//		fmt.Println("Start 1")
		for i := 0; i < 20; i++ {
			l.Lock()
			fmt.Printf("[%d]", i)
			l.Unlock()
			time.Sleep(2 * time.Second)
		}
	}()

	//	fmt.Println("Start main")
	for i := 0; i < 20; i++ {
		fmt.Printf("<%d>", i)
		if (i == 5) {
			l.Lock()
		}
		if (i == 15) {
			l.Unlock()
		}
		time.Sleep(1 * time.Second)

	}
}

*/