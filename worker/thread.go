package worker

import (
	"fmt"
)

/*
Defines an interface that we can use to register
coordinated behavior for starting/stopping various parts of factomd
*/
type Handle func(r *Thread, args ...interface{})

// worker process with structured callbacks
// parent relation helps trace worker dependencies
type Thread struct {
	Index      int
	Parent     int
	onRun      func()
	onComplete func()
	onExit     func()
}

// indicates a specific thread callback
type callback int

// list of all thread callbacks
const (
	RUN = iota + 1
	COMPLETE
	EXIT
)

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
func (r *Thread) Run(f func()) *Thread {
	r.onRun = f
	return r
}

// Add Complete Callback
func (r *Thread) Complete(f func()) *Thread {
	r.onComplete = f
	return r
}

// Add Exit Callback
func (r *Thread) Exit(f func()) *Thread {
	r.onExit = f
	return r
}
