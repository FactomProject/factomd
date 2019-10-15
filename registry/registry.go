package registry

import (
	"fmt"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/worker"
	"runtime"
	"sync"
)

// Index of all top-level threads
type process struct {
	Mutex     sync.Mutex
	ID        int
	Parent    int
	Index     []*worker.Thread
	initDone  bool
	initWait  sync.WaitGroup // init code
	runWait   sync.WaitGroup // completes when all threads are running
	doneWait  sync.WaitGroup // completes when all threads are complete
	exitWait  sync.WaitGroup // used to trigger exit logic
	exitWatch sync.WaitGroup // completes when all exit functions have completed
}

// type used to to provide initializer function and
// a hook to invoke top level threads to begin execution
type processRegistry struct {
	Mutex sync.Mutex
	Index []*process
}

// top level process list
var globalRegistry = &processRegistry{}

// trigger exit calls
func (p *process) Exit() {
	defer func() { recover() }() // don't panic if exitWait is already Done
	p.exitWait.Done()
}

// add a new thread to the global registry
func (p *process) addThread(args ...interface{}) *worker.Thread {
	if p.initDone {
		panic("sub-threads must only spawn during initialization")
	}

	p.initWait.Add(1)
	p.runWait.Add(1)
	p.doneWait.Add(1)
	p.exitWatch.Add(1)

	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	threadId := len(p.Index)

	w := worker.New()
	w.ID = threadId
	w.Register = p
	p.Index = append(p.Index, w)
	return w
}

// Bind thread run-level callbacks to wait groups
func (p *process) bindCallbacks(w *worker.Thread, initHandler worker.Handle) {
	go func() {
		// initHandler binds all other callbacks
		// and can spawn child threads
		initHandler(w)
		p.initWait.Done()
	}()

	go func() {
		// runs actual thread logic - will likely be a pub/sub handler
		// that binds to the subscription manager
		p.initWait.Wait()
		p.runWait.Done()
		w.Call(worker.RUN)
		w.Call(worker.COMPLETE)
		p.doneWait.Done()
	}()

	go func() {
		// cleanup on exit
		p.exitWait.Wait()
		w.Call(worker.EXIT)
		p.exitWatch.Done()
	}()

}

// Start a new root thread w/ coordinated start/stop callback hooks
func (p *process) Register(initFunction worker.Handle) {
	_, file, line, _ := runtime.Caller(1)
	caller := fmt.Sprintf("%s:%v", file[worker.Prefix:], line)
	r := p.addThread()
	r.Caller = caller
	r.Parent = r.ID // root threads are their own parent
	p.bindCallbacks(r, initFunction)
}

// Start a child process and register callbacks
func (p *process) Thread(w *worker.Thread, initFunction worker.Handle) {
	t := p.addThread()
	t.Parent = w.ID // child threads have a parent
	t.PID = p.ID    // set process ID
	p.bindCallbacks(t, initFunction)
}

// fork a new process with it's own lifecycle
func (p *process) Process(w *worker.Thread, initFunction worker.Handle) {
	f := new()
	f.Parent = p.ID // keep relation to parent process
	// break parent relation
	f.Register(initFunction)

	// cause this process to execute as part of the run lifecycle of the parent thread
	w.Run(f.Run)
}

// interface to avoid exposing registry internals
type Process interface {
	Register(worker worker.Handle)
	Run()
	Exit()
	WaitForRunning()
}

// create a new root process
func New() Process {
	return new()
}

// top level call to begin a new process definition
// a process has many sub-threads (goroutines)

func new() *process {
	globalRegistry.Mutex.Lock()
	defer globalRegistry.Mutex.Unlock()
	// bind to global interrupt handler
	p := &process{}
	p.ID = len(globalRegistry.Index)
	p.Parent = p.ID // root processes are their own parent
	globalRegistry.Index = append(globalRegistry.Index, p)
	fnode.AddInterruptHandler(p.Exit) // trigger exit behavior in the case of SIGINT
	p.exitWait.Add(1)
	return p
}

func (p *process) WaitForRunning() {
	p.runWait.Wait()
}

// execute all threads
func (p *process) Run() {
	p.initWait.Wait()
	p.initDone = true
	p.runWait.Wait()
	p.doneWait.Wait()
	p.Exit()
	p.exitWatch.Wait()
}

func GetRegistry() *processRegistry {
	return globalRegistry
}

func Graph() (out string) {

	out = out + "\n\n"
	var colors []string = []string{"95cde5", "b01700", "db8e3c", "ffe35f"}

	// NOTE: we don't deal w/ relations between processes
	// though the Fork() function does provide for that
	// currently this feature is unused
	for _, p := range globalRegistry.Index {
		for _, t := range p.Index {
			if t.IsRoot() {
				continue
			}
			out = out + fmt.Sprintf("%v -> %v\n", t.Parent, t.ID)
		}
	}

	for _, p := range globalRegistry.Index {
		for i, t := range p.Index {
			out = out + fmt.Sprintf("%d {color:#%v, shape:dot, label:%v}\n", t.ID, colors[i%len(colors)], t.Caller)
		}
	}

	return out
}
