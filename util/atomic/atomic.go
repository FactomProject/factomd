package atomic

import (
	"fmt"
	"os"
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

func (a *AtomicInt) Add(x int) {
	atomic.AddInt64((*int64)(a), int64(x))
}

func (a *AtomicInt) Load() int {
	return int(atomic.LoadInt64((*int64)(a)))
}

type AtomicInt64 int64

func (a *AtomicInt64) Store(x int64) {
	atomic.StoreInt64((*int64)(a), x)
}

func (a *AtomicInt64) Add(x int64) {
	atomic.AddInt64((*int64)(a), x)
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

var prefix int // index to trim the file paths to just the interesting parts
func init() {
	_, fn, _, _ := runtime.Caller(0)
	end := strings.Index(fn, "factomd/") + len("factomd")
	s := fn[0:end]
	_ = s
	prefix = end
}

func Goid() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	s := string(buf[:n])
	idField := s[:strings.Index(s, "[")-1]
	return idField
}

func GetTestHomeDir() string {
	_, fn, line, _ := runtime.Caller(2)
	fn = fn[prefix:]
	testString := strings.ReplaceAll(fn, string(os.PathSeparator), "_")
	return fmt.Sprintf("HOME%s_%d", strings.ReplaceAll(testString, ".go", ""), line)
}

func WhereAmIString(depth int) string {
	_, fn, line, _ := runtime.Caller(depth + 1)
	fn = fn[prefix:]
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
	waiting  AtomicInt     // Count of routines waiting on this lock
	owner    AtomicString  // owner of the lock at the moment
	done     chan struct{} // Channel to signal success to starvation detector
}

const yeaOfLittleFaith = true // true means mutex lock instead of CAS lock
const enableStarvationDetection = true
const enableOwnerTracking = true && enableStarvationDetection // no point in tracking owners if not detecting starvation

func (c *DebugMutex) timeStarvation(whereAmI string) {
	var owner string
	for {
		for i := 0; i < 1000; i++ {
			select {
			case <-c.done:
				return
			default:
				if enableOwnerTracking && owner == "" {
					owner = c.owner.Load()
				}
				time.Sleep(3 * time.Millisecond)
			}
		}
		if enableOwnerTracking {
			fmt.Printf("%s:Lock starving waiting for [%s] at %s\n", c.name.Load(), owner, whereAmI)
		} else {
			fmt.Printf("%s:Lock starving at %s\n", c.name.Load(), whereAmI)
		}
	}
}

func (c *DebugMutex) lockCAS() {
	b := atomic.CompareAndSwapInt32(&c.lock, 0, 1) // set lock to 1 iff it is 0
	if !b {
		if enableStarvationDetection {
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
			runtime.Gosched() // sit and spin
			//time.Sleep(3 * time.Millisecond) // sit and spin
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
	if enableStarvationDetection && c.lockBool.Load() {
		// Make a timer to whine if I am starving!
		if c.done == nil {
			c.done = make(chan struct{}, 1)
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

// Pick your poison ...
func (c *DebugMutex) Lock() {
	c.waiting.Add(1) // Add me to the waiting count
	// actually do the locking()
	if yeaOfLittleFaith {
		c.lockMutex()
	} else {
		c.lockCAS()
	}
	c.waiting.Add(-1) // I am no longer waiting

	if c.name.Load() == "" {
		c.name.Store("DebugMutex " + WhereAmIString(1))
	}
	if enableOwnerTracking {
		c.owner.Store(WhereAmIString(1))
	}
	//time.Sleep(20 * time.Millisecond) // Hog the lock -- debug -- clay
}

func (c *DebugMutex) Unlock() {
	if yeaOfLittleFaith {
		c.unlockMutex()
	} else {
		c.unlockCAS()
	}
	if enableOwnerTracking && c.waiting.Load() == 0 {
		c.owner.Store("unlocked") // only set unlocked if no one is waiting
	}
}
