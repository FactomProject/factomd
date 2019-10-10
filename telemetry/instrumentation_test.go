package telemetry_test

import (
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/worker"
	"testing"
)

func TestRegistration(t *testing.T) {
	// pass
}

// spawn threads that in turn spawn children during initialization
func threadFactory(t *testing.T, name string) worker.Handle {
	return func(w *worker.Thread, args ...interface{}) {
		t.Logf("initializing %s", name)

		w.OnRun(func() {
			t.Logf("running %v %s", w.ID, name)
		}).OnComplete(func() {
			t.Logf("complete %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		}).OnExit(func() {
			t.Logf("exit %v %s", w.ID, name)
			//time.Sleep(50*time.Millisecond)
		})
	}
}

func TestSimulation(t *testing.T) {
	// Just load simulator
	testHelper.SetupSim("L", map[string]string{}, 10, 0, 0, t)
}

func TestRegisterThread(t *testing.T) {
	// create a process with 3 root nodes
	p := registry.New()
	p.Register(threadFactory(t, "bar"))
	p.Run()
}
