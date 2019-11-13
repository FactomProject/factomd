package registry_test

import (
	"context"
	"fmt"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
	"testing"
)

// add a sub thread that uses all callbacks
// and also spawn's a child that uses only 'Run' callback
func threadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread) {
		t.Logf("initializing %s", name)
		ctx, cancel := context.WithCancel(context.Background())

		// Launch another sub-thread
		w.Run(func() {
			name := fmt.Sprintf("%s/%s", name, "bar")
			// this thread that only uses OnRun group
			t.Logf("running %v %s", w.ID, name)
			select {
			case <- ctx.Done():
				t.Logf("context.Done() %v %s", w.ID, name)
			}
		})

		// bind functions to thread lifecycle
		w.OnReady(func() {
			t.Logf("Ready %v %s", w.ID, name)
		}).OnRun(func() {
			t.Logf("running %v %s", w.ID, name)
			select {
			case <- ctx.Done():
				t.Logf("context.Done() %v %s", w.ID, name)
			}
		}).OnExit(func() {
			t.Logf("exit %v %s", w.ID, name)
			cancel()
		}).OnComplete(func() {
			t.Logf("complete %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

func TestRegisterThread(t *testing.T) {
	// create a process with 1 root process
	p := registry.New()
	p.Register(threadFactory(t, "foo"))
	go func(){
		p.WaitForRunning()
		go p.Exit() // normally invoked via SIGINT
	}()
	go p.Run()
	p.WaitForExit()
	t.Log(registry.Graph())
}
