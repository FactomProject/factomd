package worker

import (
	"fmt"
)

/*
Defines an interface that we can use to register
coordinated behavior for starting/stopping various parts of factomd
*/
type Handle func(r *Thread, args ...interface{})

type RegistryHandler func(r *Thread, initFunction Handle, args ...interface{})

// worker process with structured callbacks
// parent relation helps trace worker dependencies
type Thread struct {
	RegisterNew RegistryHandler // RegistryCallback for sub-threads
	Index       int
	Parent      int
	onRun       func()
	onComplete  func()
	onExit      func()
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
func (r *Thread) Run(runFunc func()) {
	r.Spawn(func(w *Thread, args ...interface{}) {
		w.OnRun(runFunc)
	})
}

func (r *Thread) Init(initFunc func()) {
	r.Spawn(func(w *Thread, args ...interface{}) {
		initFunc()
	})
}

// Spawn a child thread
func (r *Thread) Spawn(initFunction Handle, args ...interface{}) {
	r.RegisterNew(r, initFunction, args...)
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
