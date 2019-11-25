package log_test

import (
	"github.com/FactomProject/factomd/registry"
	"testing"

	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/worker"
	"github.com/stretchr/testify/assert"
)

func TestLogPrintf(t *testing.T) {
	assert.NotPanics(t, func() {
		log.LogPrintf("testing", "unittest %v", "FOO")
	})
}
func TestRegisterThread(t *testing.T) {

	p := registry.New()

	threadFactory := func(w *worker.Thread) {
		assert.NotPanics(t, func() {
			w.Log.LogPrintf("testing", "%v", "foo")
		})
		p.Exit()
	}
	// create a process with 3 root nodes
	p.Register(threadFactory)
	p.Run()
}
