package registry_test

import (
	"fmt"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
	"github.com/stretchr/testify/assert"
	"testing"
)

// add a sub thread that uses all callbacks
// and also spawn's a child that uses only 'Run' callback
func subThreadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread) {
		t.Logf("initializing %s", name)

		w.Run(func() {
			// add a thread with no initialization behavior
			t.Logf("running %v %s", w.ID, fmt.Sprintf("%s/%s", name, "qux"))
		})

		w.OnRun(func() {
			t.Logf("running %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		}).OnComplete(func() {
			t.Logf("complete %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		}).OnExit(func() {
			t.Logf("exit %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

// spawn threads that in turn spawn children during initialization
func threadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread) {
		t.Logf("initializing %s", name)

		// add sub-thread
		sub := fmt.Sprintf("%v/%v", name, "sub")
		w.Spawn(subThreadFactory(t, sub))

		// add sub-process - entire thread lifecycle lives inside the 'running' lifecycle of parent thread
		subProc := fmt.Sprintf("%v/%v", name, "subproc")
		w.Fork(subThreadFactory(t, subProc))

		w.OnRun(func() {
			t.Logf("running %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
			assert.Panics(t, func() {
				subSub := fmt.Sprintf("%v/%v", name, "sub")
				w.Spawn(subThreadFactory(t, subSub))
			}, "should fail when trying to spawn outside of init phase")
		}).OnComplete(func() {
			t.Logf("complete %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		}).OnExit(func() {
			t.Logf("exit %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

func TestRegisterThread(t *testing.T) {
	// create a process with 3 root nodes
	p := registry.New()
	p.Register(threadFactory(t, "foo"))
	p.Register(threadFactory(t, "bar"))
	p.Run()
	t.Log(registry.Graph())
}
