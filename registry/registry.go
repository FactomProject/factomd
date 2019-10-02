package registry

import (
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/worker"
	"sync"
)

type globalRegistry struct {
	// TODO: record parent
	Mutex       sync.Mutex
	Index       []*worker.Registry
}


var threadMgr = &globalRegistry{}

func Exit() {
	var wait sync.WaitGroup

	for _, r := range threadMgr.Index {
		wait.Add(1)
		go func(){
			r.Call(worker.EXIT)
			wait.Done()
		}()
	}
	wait.Wait()
}

var initWait sync.WaitGroup

// Run a thread w/ coordinated start and callback hooks
func Init(initFunction worker.Thread, args ...interface{}) {

	r := &worker.Registry{}

	initWait.Add(1)
	go func() {
		r.Log = log.ThreadLogger // inject global logging
		threadMgr.Index = append(threadMgr.Index, r)
		initFunction(r, args...)
		initWait.Done()
	}()

	go func() {
		initWait.Wait()
		r.Call(worker.RUN)
		r.Call(worker.COMPLETE)
	}()
}