package worker

import (
	"fmt"
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"runtime"
	"strings"
)

// callback handle
type Handle func(r *Thread)

// interface to catch SIGINT
type InterruptHandler func(func())

// create new thread
func New() *Thread {
	w := &Thread{}
	w.Log = log.New(w)
	return w
}

// thread ID
func (r *Thread) GetID() int {
	return r.ID
}

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
	Thread(*Thread, Handle)  // RegistryCallback for sub-threads
	Process(*Thread, Handle) // callback to fork a new process
}

// worker process with structured callbacks
// parent relation helps trace worker dependencies
type Thread struct {
	common.Name                // support hierarchical naming
	log.ICaller                // interface to for some fields used by logger
	Log         interfaces.Log // threaded logger
	Register    IRegister      // callbacks to register threads
	PID         int            // process ID that this thread belongs to
	ID          int            // thread id
	ParentID    int            // parent thread
	Caller      string         // runtime location where thread starts
	onRun       func()         // execute during 'run' state
	onComplete  func()         // execute after all run functions complete
	onExit      func()         // executes during SIGINT or after shutdown of run state
}

// indicates a specific thread callback
type callback int

// list of all thread callbacks
const (
	RUN = iota + 1
	COMPLETE
	EXIT
)

// KLUDGE: duplicated from atomic
var Prefix int // index to trim the file paths to just the interesting parts

func init() {
	_, fn, _, _ := runtime.Caller(0)
	end := strings.Index(fn, "factomd/") + len("factomd")
	s := fn[0:end]
	_ = s
	Prefix = end
}

// convenience wrapper starts a closure in a sub-thread
// useful for situations where only Run callback is needed
// can be thought of as 'leaves' of the thread runtime dependency graph
func (r *Thread) Run(runFunc func()) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%v", file[Prefix:], line)

	r.Spawn(func(w *Thread) {
		w.Caller = caller
		w.OnRun(runFunc)
	})
}

// Spawn a child thread and register callbacks
// this is useful to bind functions to Init/Run/Stop callbacks
func (r *Thread) Spawn(initFunction Handle) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%v", file[Prefix:], line)

	r.Register.Thread(r, func(w *Thread) {
		w.Caller = caller
		initFunction(w)
	})
}

// Fork process with it's own thread lifecycle
// NOTE: it's required to run the process
func (r *Thread) Fork(initFunction Handle) {
	r.Register.Process(r, initFunction)
}

// Invoke specific callbacks synchronously
func (r *Thread) Call(c callback) {
	switch c {
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

// Add Run Callback
func (r *Thread) OnRun(f func()) *Thread {
	r.onRun = f
	return r
}

// Add Complete Callback
func (r *Thread) OnComplete(f func()) *Thread {
	r.onComplete = f
	return r
}

// Add Exit Callback
func (r *Thread) OnExit(f func()) *Thread {
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
