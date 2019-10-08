package worker

import (
	"fmt"
)

// callback handle
type Handle func(r *Thread, args ...interface{})

// interface to registry
type RegistryHandler func(r *Thread, initFunction Handle, args ...interface{})

// interface to catch SIGINT
type InterruptHandler func(func())

// worker process with structured callbacks
// parent relation helps trace worker dependencies
type Thread struct {
	RegisterThread           RegistryHandler  // RegistryCallback for sub-threads
	RegisterProcess          RegistryHandler  // callback to fork a new process
	RegisterInterruptHandler InterruptHandler // register w/ global SIGINT handler
	PID                      int              // process ID that this thread belongs to
	ID                       int              // thread id
	Parent                   int              // parent thread
	onRun                    func()           // execute during 'run' state
	onComplete               func()           // execute after all run functions complete
	onExit                   func()           // executes during SIGINT or after shutdown of run state
}

// indicates a specific thread callback
type callback int

// list of all thread callbacks
const (
	RUN = iota + 1
	COMPLETE
	EXIT
)

// convenience wrapper starts a closure in a sub-thread
// useful for situations where only Run callback is needed
// can be thought of as 'leaves' of the thread runtime dependency graph
func (r *Thread) Run(runFunc func()) {
	r.Spawn(func(w *Thread, args ...interface{}) {
		w.OnRun(runFunc)
	})
}

// Spawn a child thread and register callbacks
// this is useful to bind functions to Init/Run/Stop callbacks
func (r *Thread) Spawn(initFunction Handle, args ...interface{}) {
	r.RegisterThread(r, initFunction, args...)
}

// Fork process with it's own thread lifecycle
// NOTE: it's required to run the process
func (r *Thread) Fork(initFunction Handle, args ...interface{}) {
	r.RegisterProcess(r, initFunction, args...)
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
	return r.ID == r.Parent
}
