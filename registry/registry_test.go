package registry_test

import (
	"fmt"
	"github.com/FactomProject/factomd/registry"
	"github.com/FactomProject/factomd/worker"
	"testing"
)

func registerStartThreadFunc(w *worker.Registry, args ...interface{}) {

	t := args[0].(*testing.T)
	name := fmt.Sprintf("%s", args[1])
	wait0 := args[2].(chan interface{})
	wait1 := args[3].(chan interface{})

	t.Logf("initializing %s", name)

	log := w.Log // get a logger handle
	log.LogPrintf("testing logger %s", name)

	w.Run(func() {
		t.Logf("running %s", name)
	}).Complete(func() {
		t.Logf("complete %s", name)
		close(wait0)
	}).Exit(func() {
		t.Logf("exit %s", name)
		close(wait1)
	})
}

func TestRegisterThread(t *testing.T) {
	wait0 := make(chan interface{})
	wait1 := make(chan interface{})
	registry.Init(registerStartThreadFunc, t, "foo", wait0, wait1)
	<- wait0 // wait for execution
	registry.Exit()
	<- wait1 // wait for execution
}