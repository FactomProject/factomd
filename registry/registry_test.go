package registry_test

import (
	"fmt"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
	"testing"
)

func subThreadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread, args ...interface{}) {
		t.Logf("initializing %s", name)

		w.Run(func() {
			t.Logf("running %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).Complete(func() {
			t.Logf("complete %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).Exit(func() {
			t.Logf("exit %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

func threadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread, args ...interface{}) {
		t.Logf("initializing %s", name)

		registry.Spawn(w, subThreadFactory(t, fmt.Sprintf("%v/%v", name, "sub")))

		w.Run(func() {
			t.Logf("running %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).Complete(func() {
			t.Logf("complete %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		}).Exit(func() {
			t.Logf("exit %v %s", w.Index, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

func TestRegisterThread(t *testing.T) {
	reg := registry.New()
	reg(threadFactory(t, "foo"))
	reg(threadFactory(t, "bar"))
	reg(threadFactory(t, "baz"))
	reg.Run()
}
