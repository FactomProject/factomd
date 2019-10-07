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
	return func(w *worker.Thread, args ...interface{}) {
		t.Logf("initializing %s", name)

		w.Run(func() {
			// add a thread with no initialization behavior
			t.Logf("running %v %s", w.Index, fmt.Sprintf("%s/%s", name, "qux"))
		})

		w.OnRun(func() {
			t.Logf("running %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).OnComplete(func() {
			t.Logf("complete %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).OnExit(func() {
			t.Logf("exit %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

// spawn threads that in turn spawn children during initialization
func threadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread, args ...interface{}) {
		t.Logf("initializing %s", name)

		w.Spawn(subThreadFactory(t, fmt.Sprintf("%v/%v", name, "sub")))

		w.OnRun(func() {
			t.Logf("running %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
			assert.Panics(t, func() {
				w.Spawn(subThreadFactory(t, fmt.Sprintf("%v/%v", name, "sub")))
			}, "should fail when trying to spawn outside of init phase")
		}).OnComplete(func() {
			t.Logf("complete %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).OnExit(func() {
			t.Logf("exit %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

func TestRegisterThread(t *testing.T) {
	// create a process with 3 root nodes
	reg := registry.New()
	reg(threadFactory(t, "foo"))
	reg(threadFactory(t, "bar"))
	reg(threadFactory(t, "baz"))
	reg.Run()
	t.Log(registry.Graph())
}
