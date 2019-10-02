package registry

import (
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util/atomic"
	"github.com/FactomProject/factomd/worker"
	"sync"
)


// indexed by  goId
type storeMgr struct {
	Mutex       sync.Mutex
	Index map[string]*locals
}

// handle to invoke async/deferred behavior
type callback struct {
	call worker.Thread
	args []interface{}
}

// thread local vars
type locals struct {
	Logger log.Log
	OnStop callback
}

var StoreManager = storeMgr{Index: make(map[string]*locals) }

func getLocal(gid string) *locals {
	StoreManager.Mutex.Lock()
	defer StoreManager.Mutex.Unlock()
	l, ok := StoreManager.Index[gid]
	if ! ok {
		panic("unregistered thread")
	}
	return l
}

func Locals() *locals{
	return getLocal(atomic.Goid())
}

//var stop chan interface{}

// signal all threads to start
var start = make(chan interface{})

func Logger() log.Log {
	return getLocal(atomic.Goid()).Logger
}

// kick off all threads
func Start() {
	close(start)
}

// register a callback for global stop
func OnStop(t worker.Thread, args ...interface{}) {
	l := getLocal(atomic.Goid())
	l.OnStop = callback{t, args}
}

// Run a thread w/ coordinated start and callback hooks
func Run(t worker.Thread, args ...interface{}) {

	// set thread local data

	// run the thread
	go func() {
		gid := atomic.Goid()

		l := &locals{
			Logger: log.ThreadLogger,
		}

		func() { // add locals to registry
			StoreManager.Mutex.Lock()
			defer StoreManager.Mutex.Unlock()
			StoreManager.Index[gid] = l
		}()

		<- start // synchronize thread starts

		t(args...) // run the Thread

		// run onStop Behavior
		if l.OnStop.call != nil {
			l.OnStop.call(l.OnStop.args...)
		}
	}()
}