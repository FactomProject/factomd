package registry

import (
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/worker"
	"sync"
)

// Index of all top-level threads
type globalRegistry struct {
	Mutex    sync.Mutex
	Index    []*worker.Thread
	initDone bool
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
func addThread(args ...interface{}) *worker.Thread {
	if threadMgr.initDone {
		panic("sub-threads must only spawn during initialization")
	}

	initWait.Add(1)
	runWait.Add(1)
	doneWait.Add(1)
	exitWatch.Add(1)

	w := &worker.Thread{
		RegisterNew: spawn, // inject spawn callback
	}

	threadMgr.Mutex.Lock()
	defer threadMgr.Mutex.Unlock()
	w.Index = len(threadMgr.Index)
	threadMgr.Index = append(threadMgr.Index, w)
	return w
}

// Bind thread run-level callbacks to wait groups
func bindCallbacks(r *worker.Thread, initHandler worker.Handle, args ...interface{}) {
	go func() {
		// initHandler binds all other callbacks
		// and can spawn child threads
		initHandler(r, args...)
		initWait.Done()
	}()

	go func() {
		// runs actual thread logic - will likely be a pub/sub handler
		// that binds to the subscription manager
		initWait.Wait()
		runWait.Done()
		r.Call(worker.RUN)
		r.Call(worker.COMPLETE)
		doneWait.Done()
	}()

	go func() {
		// cleanup on exit
		exitWait.Wait()
		r.Call(worker.EXIT)
		exitWatch.Done()
	}()

}

// Start a new root thread w/ coordinated start/stop callback hooks
func initializer(initFunction worker.Handle, args ...interface{}) {
	r := addThread()
	r.Parent = r.Index // root threads are their own parent
	bindCallbacks(r, initFunction, args...)
}

// Start a child process
func spawn(r *worker.Thread, initFunction worker.Handle, args ...interface{}) {
	t := addThread()
	t.Parent = r.Index // child threads have a different parent
	bindCallbacks(t, initFunction, args...)
}

// type used to to provide initializer function and
// a hook to invoke top level threads to begin execution
type process func(worker worker.Handle, args ...interface{})

// top level call to begin a new process definition
// a process has many sub-threads (goroutines)
func New() process {
	// bind to global interrupt handler
	fnode.AddInterruptHandler(Exit)
	exitWait.Add(1)
	return initializer
}

// execute all threads
func (process) Run() {
	initWait.Wait()
	threadMgr.initDone = true
	runWait.Wait()
	doneWait.Wait()
	fnode.SendSigInt()
	exitWatch.Wait()
}

func (process) WaitForRunning() {
	runWait.Wait()
}
