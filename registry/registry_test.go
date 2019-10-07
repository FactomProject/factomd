package registry_test

import (
	"fmt"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func subThreadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread, args ...interface{}) {
		t.Logf("initializing %s", name)

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
	reg := registry.New()
	reg(threadFactory(t, "foo"))
	reg(threadFactory(t, "bar"))
	reg(threadFactory(t, "baz"))
	reg.Run()
}
