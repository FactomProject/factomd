package atomic

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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

func (a *AtomicString) Store(x string) {
	a.mu.Lock()
	a.s = x
	a.mu.Unlock()
}

func (a *AtomicString) Load() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.s
}

// Hacky debugging stuff... probably not a great home for it

func Goid() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	s := string(buf[:n])
	idField := s[:strings.Index(s, "[")]
	return idField
}

func WhereAmIString(depth int) string {
	_, fn, line, _ := runtime.Caller(depth + 1)
	return fmt.Sprintf("%v-%s:%d", Goid(), fn, line)
}

func WhereAmIMsg(msg string) {
	fmt.Printf("\n[%v] %v\n", msg, WhereAmIString(1))
}

func WhereAmI() {
	fmt.Printf("\n%v\n", WhereAmIString(1))
}

func WhereAmI2(msg string, depth int) {
	fmt.Printf("\n[%v] %v\n", msg, WhereAmIString(depth+1))
}

type DebugMutex struct {
	name     AtomicString  // Name of this mutex
	lock     int32         // lock for debug lock functionality
	mu       sync.Mutex    // lock for not trusting the debug lock functionality or for traditional locking
	lockBool AtomicBool    // lock for detecting starvation when not trusting the debug lock functionality
	owner    AtomicString  // owner of the lock at the moment
	done     chan struct{} // Channel to signal success to starvation detector
}

var yeaOfLittleFaith AtomicBool = AtomicBool(0) // true means mutex lock instead of CAS lock
var enableStarvationDetection AtomicBool = AtomicBool(1)
var enableAlreadyLockedDetection AtomicBool = AtomicBool(0)
var enableOwnerTracking AtomicBool = AtomicBool(1 & enableStarvationDetection) // no point in tracking owners if not detecting starvation

func (c *DebugMutex) timeStarvation(whereAmI string) {
	for {
		for i := 0; i < 1000; i++ {
			select {
			case <-c.done:
				return
			default:
				time.Sleep(3 * time.Millisecond)
			}
		}
		if enableOwnerTracking.Load() {
			fmt.Printf("%s:Lock starving waiting for [%s] at %s\n", c.name.Load(), c.owner.Load(), whereAmI)
		} else {
			fmt.Printf("%s:Lock starving at %s\n", c.name.Load(), whereAmI)
		}
	}
}

func (c *DebugMutex) lockCAS() {
	b := atomic.CompareAndSwapInt32(&c.lock, 0, 1) // set lock to 1 iff it is 0
	if !b {
		if enableStarvationDetection.Load() {
			// Make a timer to whine if I am starving!
			if c.done == nil {
				c.done = make(chan struct{})
			}
			go c.timeStarvation(WhereAmIString(2))
			defer func() { c.done <- struct{}{} }() // End the timer when I get the lock
		} // end of starvation detection code
		for { // blocking waiting on lock loop
			b := atomic.CompareAndSwapInt32(&c.lock, 0, 1) // set lock to 1 iff it is 0
			if b {
				break // Yea! we got the lock
			}
			time.Sleep(3 * time.Millisecond) // sit and spin
		}
	}
}

func (c *DebugMutex) unlockCAS() {
	c.lockBool.Store(false)
	b := atomic.CompareAndSwapInt32(&c.lock, 1, 0) // set lock to 0 iff it is 1
	if !b {
		WhereAmI2(c.name.Load()+":Already Unlocked", 2)
		panic("Double Unlock")
	}
}

// try and detect bad behaviors using a traditional lock
func (c *DebugMutex) lockMutex() {
	if enableStarvationDetection.Load() && c.lockBool.Load() {
		// Make a timer to whine if I am starving!
		if c.done == nil {
			c.done = make(chan struct{})
		}
		go c.timeStarvation(WhereAmIString(2))
		defer func() { c.done <- struct{}{} }() // End the timer when I get the lock
	} // end of starvation detection code
	// It is possible to loose the lock after the check and before here and starve anyway undetected but ...
	c.mu.Lock()
	c.lockBool.Store(true)
}

func (c *DebugMutex) unlockMutex() {
	if c.lockBool.Load() == false {
		WhereAmI2(c.name.Load()+":Already Unlocked", 2)
		panic("Double Unlock")
	}
	c.lockBool.Store(false)
	c.mu.Unlock()
}

// Pick your posion ...
func (c *DebugMutex) Lock() {
	if c.lockBool.Load() {
		if enableAlreadyLockedDetection.Load() {
			WhereAmI2(c.name.Load()+":Already Locked", 2)
		}
	}
	// actually do the locking()
	if yeaOfLittleFaith.Load() {
		c.lockMutex()
	} else {
		c.lockCAS()
	}

	if c.name.Load() == "" {
		c.name.Store("DebugMutex " + WhereAmIString(1))
	}
	if enableOwnerTracking.Load() {
		c.owner.Store(WhereAmIString(1))
	}
	//time.Sleep(20 * time.Millisecond) // Hog the lock -- debug -- clay
}

func (c *DebugMutex) Unlock() {
	if yeaOfLittleFaith.Load() {
		c.unlockMutex()
	} else {
		c.unlockCAS()
	}
	if enableOwnerTracking.Load() {
		c.owner.Store("unlocked")
	}
}

/*
func main() {
	fmt.Println("Begin Main")

	var l DebugMutex

	// try all eight flavors
	for i := 0; i < 4; i++ {
		enableStarvationDetection.Store(i&(1<<0) != 0)
		yeaOfLittleFaith.Store(i&(1<<1) != 0)

		timeUnit := 1000 * time.Millisecond

		go func() {
			//		fmt.Println("Start 1")
			fmt.Println("Timer started")
			for i := 0; i < 10; i++ {
				l.Lock()
				fmt.Printf("[%d]", i)
				l.Unlock()
				time.Sleep(4 * timeUnit) // must be > 3 second to see starvation
			}
			fmt.Println("Timer done")
		}()

		fmt.Printf("Start test faith1 = %v,  starvationDetection = %v\n", yeaOfLittleFaith, enableStarvationDetection)
		for i := 0; i < 20; i++ {
			fmt.Printf("<%d>", i)
			if i == 5 {
				l.Lock()
			}
			if i == 10 {
				l.Unlock()
			}
			time.Sleep((40 * 2) / 20 * timeUnit)
		}
		time.Sleep(3 * time.Second) // make sure the test finishes
		fmt.Println("\nMain loop done")
	}
}
*/
