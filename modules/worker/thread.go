package worker

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/FactomProject/factomd/common"
)

// callback handle
type Handle func(r *Thread)

// interface to catch SIGINT
type InterruptHandler func(func())

// create new thread
func New(parent *common.Name, name string) *Thread {
	w := &Thread{}
	w.Name.NameInit(parent, name, reflect.TypeOf(w).String())
	// REVIEW: Layered Logger is the newest log implementation
	//w.logging = logging.NewLayerLogger(log.GlobalLogger, map[string]string{"fnode": w.GetPath()})
	return w
}

// thread ID
func (r *Thread) GetID() int {
	return r.ID
}

// loggin someday .AddNameField("thread", Formatter("%s"), "unknown_thread")

// returns caller caller
// which is a string containing source file and line where thread is spawned
func (r *Thread) GetCaller() string {
	return r.Caller
}

// register w/ global SIGINT handler
func (*Thread) RegisterInterruptHandler(handler func()) {
	AddInterruptHandler(handler)
}

type IRegister interface {
	Thread(*Thread, string, Handle)  // RegistryCallback for sub-threads
	Process(*Thread, string, Handle) // callback to fork a new process
}

// worker process with structured callbacks
// parent relation helps trace worker dependencies
type Thread struct {
	common.Name // support hierarchical naming
	//log.ICaller                // interface to for some fields used by logger
	//logging     *logging.LayerLogger

	Register   IRegister // callbacks to register threads
	PID        int       // process ID that this thread belongs to
	ID         int       // thread id
	ParentID   int       // parent thread
	Caller     string    // runtime location where thread starts
	onReady    func()    // execute just after init - this is where subscriptions should happen
	onRun      func()    // execute during 'run' state
	onComplete func()    // execute after all run functions complete
	onExit     func()    // executes during SIGINT or after shutdown of run state
}

// indicates a specific thread callback
type callback int

// list of all thread callbacks
const (
	READY = iota + 1
	RUN
	COMPLETE
	EXIT
)

// KLUDGE: duplicated from atomic
var prefix int // index to trim the file paths to just the interesting parts

func init() {
	_, fn, _, _ := runtime.Caller(0)
	if pos := strings.Index(fn, "factomd/"); pos > -1 { // found
		prefix = pos + len("factomd")
	} else {
		prefix = 0
	}
}

// PrettyCallerString returns a file:line string of the parent of parent function
func PrettyCallerString() string {
	_, file, line, ok := runtime.Caller(2) // parent of parent
	if !ok {
		file = "unknown"
		line = -1
	}

	if strings.Contains(file, "factomd/") {
		file = file[prefix:]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// convenience wrapper starts a closure in a sub-thread
// useful for situations where only Run callback is needed
// can be thought of as 'leaves' of the thread runtime dependency graph
func (r *Thread) Run(name string, runFunc func()) {
	caller := PrettyCallerString()
	r.Spawn(name, func(w *Thread) {
		w.Caller = caller
		w.OnRun(runFunc)
	})
}

// Spawn a child thread and register callbacks
// this is useful to bind functions to Init/Run/Stop callbacks
func (r *Thread) Spawn(name string, initFunction Handle) {
	caller := PrettyCallerString()
	r.Register.Thread(r, name, func(w *Thread) {
		w.Caller = caller
		initFunction(w)
	})
}

// Fork process with it's own thread lifecycle
// NOTE: it's required to run the process
func (r *Thread) Fork(name string, initFunction Handle) {
	r.Register.Process(r, name, initFunction)
}

// Invoke specific callbacks synchronously
func (r *Thread) Call(c callback) {
	switch c {
	case READY:
		if r.onReady != nil {
			r.onReady()
		}
	case RUN:
		if r.onRun != nil {
			r.onRun()
		}
	case COMPLETE:
		if r.onComplete != nil {
			r.onComplete()
		}
	case EXIT:
		if r.onExit != nil {
			r.onExit()
		}
	default:
		panic(fmt.Sprintf("unknown callback %v", c))
	}
}

func assertNotBound(f func()) {
	if f != nil {
		panic("already bound")
	}
}

// Add Ready Callback to add subscribers
func (r *Thread) OnReady(f func()) *Thread {
	assertNotBound(r.onReady)
	r.onReady = f
	return r
}

// Add Run Callback
func (r *Thread) OnRun(f func()) *Thread {
	assertNotBound(r.onRun)
	r.onRun = f
	return r
}

// Add Complete Callback
func (r *Thread) OnComplete(f func()) *Thread {
	assertNotBound(r.onComplete)
	r.onComplete = f
	return r
}

// Add Exit Callback
func (r *Thread) OnExit(f func()) *Thread {
	assertNotBound(r.onExit)
	r.onExit = f
	return r
}

// root level threads are their own parent
func (r *Thread) IsRoot() bool {
	return r.ID == r.ParentID
}

// use thread name as label
func (r *Thread) Label() string {
	return r.GetName()
}
