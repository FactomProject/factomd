package registry_test

import (
	"github.com/FactomProject/factomd/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

// syntactic sugar for register locals and callbacks
var local =  registry.Locals
var onStop = registry.OnStop
var runThread = registry.Run
var startAllThreads = registry.Start

func registerStopThreadFunc(args ...interface{}) {
	t := args[0].(*testing.T)

	t.Logf("testing %v", args[1])
	local().Logger.LogPrintf("testing", "formatter: %v", args[1])

	wait := args[2].(chan interface{})
	close(wait) // let test resume
}

func registerStartThreadFunc(args ...interface{}) {
	t := args[0].(*testing.T)
	name := args[1].(string)

	// add a hook for stopping
	onStop(registerStopThreadFunc, t, "bar", args[2])

	t.Logf("testing %v", name)
	local().Logger.LogPrintf("testing", "formatter: %v", name)
}

func TestRegisterThread(t *testing.T){
	wait := make(chan interface{})
	runThread(registerStartThreadFunc, t, "foo", wait)
	startAllThreads()
	<- wait // wait for execution

	// cannot access locals outside of a registered thread
	assert.Panics(t, func() {
		local()
	}, "Cannot access locals outside of a registered thread")
}