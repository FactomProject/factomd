package registry

import (
	"fmt"
	"sync"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/modules/worker"
)

// Index of all top-level threads
type process struct {
	Mutex      sync.Mutex
	ID         int
	Parent     int
	Index      []*worker.Thread
	initDone   bool
	initWait   sync.WaitGroup // init code + publishers created
	readyWait  sync.WaitGroup // subscribers setup
	doneWait   sync.WaitGroup // completes when all threads are complete
	exitWait   sync.WaitGroup // used to trigger exit logic
	exitSignal sync.WaitGroup // completes when all exit functions have completed
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
	defer func() { recover() }() // don't panic if exitSignal is already Done
	p.exitSignal.Done()
}

// add a new thread to the global registry
func (p *process) addThread(parent *common.Name, name string) *worker.Thread {
	if p.initDone {
		panic("sub-threads must only spawn during initialization")
	}

	p.initWait.Add(1)
	p.readyWait.Add(1)
	p.doneWait.Add(1)
	p.exitWait.Add(1)

	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	threadId := len(p.Index)

	w := worker.New(parent, name)
	w.ID = threadId
	w.Register = p
	p.Index = append(p.Index, w)
	return w
}

// Bind thread run-level callbacks to wait groups
func (p *process) bindCallbacks(w *worker.Thread, initHandler worker.Handle) {
	// initHandler binds all other callbacks
	// and can spawn child threads
	// publishers are instantiated here
	go func() {
		{ // Init
			initHandler(w)
			p.initWait.Done()
			p.initWait.Wait()
		}
		{ // Ready
			w.Call(worker.READY)
			p.readyWait.Done()
			p.readyWait.Wait()
		}
		{ // Running
			w.Call(worker.RUN)
			p.doneWait.Done()
		}
	}()

	go func() { // exit
		p.exitSignal.Wait()
		w.Call(worker.EXIT)
		p.doneWait.Wait()
		w.Call(worker.COMPLETE)
		p.exitWait.Done()
	}()

}

// Start a new root thread w/ coordinated start/stop callback hooks
func (p *process) Register(initFunction worker.Handle) {
	caller := worker.PrettyCallerString()
	r := p.addThread(common.NilName, "rootThread")
	r.Caller = caller
	r.ParentID = r.ID // root threads are their own parent
	r.PID = p.ID
	p.bindCallbacks(r, initFunction)
}

// Start a child process and register callbacks
func (p *process) Thread(w *worker.Thread, name string, initFunction worker.Handle) {
	t := p.addThread(&w.Name, name)
	t.ParentID = w.ID // child threads have a parent
	t.PID = p.ID      // set process ID
	p.bindCallbacks(t, initFunction)
}

// fork a new process with it's own lifecycle
func (p *process) Process(w *worker.Thread, name string, initFunction worker.Handle) {
	f := newProcess()
	f.Parent = p.ID // keep relation to parent process
	// break parent relation
	f.Register(initFunction)

	// cause this process to execute as part of the run lifecycle of the parent thread
	w.Run(name, f.Run)
}

// interface to avoid exposing registry internals
type Process interface {
	Register(worker worker.Handle)
	Run()
	Exit()
	WaitForRunning()
	WaitForExit()
}

// create a new root process
func New() Process {
	return newProcess()
}

// top level call to begin a new process definition
// a process has many sub-threads (goroutines)

func newProcess() *process {
	globalRegistry.Mutex.Lock()
	defer globalRegistry.Mutex.Unlock()
	// bind to global interrupt handler
	p := &process{}
	p.ID = len(globalRegistry.Index)
	p.Parent = p.ID // root processes are their own parent
	globalRegistry.Index = append(globalRegistry.Index, p)
	worker.AddInterruptHandler(p.Exit) // trigger exit behavior in the case of SIGINT
	p.exitSignal.Add(1)
	return p
}

func (p *process) WaitForRunning() {
	p.readyWait.Wait()
}

func (p *process) WaitForExit() {
	p.exitWait.Wait()
}

// execute all threads
func (p *process) Run() {
	p.initWait.Wait()
	p.initDone = true
	p.readyWait.Wait()
	p.doneWait.Wait()
	p.exitSignal.Wait()
}

func Graph() (out string) {

	out = out + "\n\n"
	var colors = []string{"95cde5", "b01700", "db8e3c", "ffe35f"}

	// FIXME: this function needs to be updated to support graphing multiple processes
	for _, p := range globalRegistry.Index {
		for _, t := range p.Index {
			if t.IsRoot() {
				continue
			}
			out = out + fmt.Sprintf("%v.%v -> %v.%v\n", t.PID, t.ParentID, t.PID, t.ID)
		}
	}

	for _, p := range globalRegistry.Index {
		for i, t := range p.Index {
			out = out + fmt.Sprintf("%v.%v {color:#%v, shape:dot, label:%v}\n", t.PID, t.ID, colors[i%len(colors)], t.GetName())
		}
	}
	out = out + "Paste the network info above into http://arborjs.org/halfviz to visualize the network\n"
	return out
}
