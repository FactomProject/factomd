package registry

import (
	"github.com/FactomProject/factomd/worker"
	"sync"
)

// Index of all top-level threads
type globalRegistry struct {
	Mutex sync.Mutex
	Index []*worker.Registry
}

// singleton thread registry
var threadMgr = &globalRegistry{}

// wait group for init code
// in this section thread specific copies of data are constructed
var initWait sync.WaitGroup

// wait group that completes when all threads are running
var runWait sync.WaitGroup

// wait group that completes when all threads are complete
var doneWait sync.WaitGroup

// wait group that is used to trigger exit logic
var exitWait sync.WaitGroup

// wait group that completes when all exit functions have completed
var exitWatch sync.WaitGroup

// trigger exit calls
func Exit() {
	exitWait.Done()
}

// add a new thread to the global registry
func addThread(args ...interface{}) *worker.Registry {
	initWait.Add(1)
	runWait.Add(1)
	doneWait.Add(1)
	exitWatch.Add(1)

	w := &worker.Registry{}
	threadMgr.Mutex.Lock()
	defer threadMgr.Mutex.Unlock()
	w.Index = len(threadMgr.Index)
	threadMgr.Index = append(threadMgr.Index, w)
	return w
}

// Bind thread run-level callbacks to wait groups
func bindCallbacks(r *worker.Registry, initFunction worker.Thread, args ...interface{}) {
	go func() {
		initFunction(r, args...)
		initWait.Done()
	}()

	go func() {
		initWait.Wait()
		runWait.Done()
		r.Call(worker.RUN)
		r.Call(worker.COMPLETE)
		doneWait.Done()
	}()

	go func() {
		exitWait.Wait()
		r.Call(worker.EXIT)
		exitWatch.Done()
	}()

}

// Start a a new root thread w/ coordinated start and callback hooks
func initializer(initFunction worker.Thread, args ...interface{}) {
	r := addThread()
	bindCallbacks(r, initFunction, args...)
}

// Start a child process
func Spawn(r *worker.Registry, initFunction worker.Thread, args ...interface{}) {
	t := addThread()
	t.Parent = r.Index
	bindCallbacks(t, initFunction, args...)
}

// type used to to provide initializer function and
// a hook to invoke top level threads to begin execution
type process func(worker worker.Thread, args ...interface{})

// top level call to begin a new process definition
// a process has many sub-threads (goroutines)
func New() process {
	exitWait.Add(1)
	return initializer
}

// execute all threads
func (process) Run() {
	initWait.Wait()
	runWait.Wait()
	doneWait.Wait()
	Exit()
	exitWatch.Wait()
}
