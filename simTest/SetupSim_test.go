package simtest

import (
	"testing"

	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
)

/*
* Super simplified test to reporduce dependent holidng
*
* Steps:
* 1. run this test
* 2. collect logs
* 3. check the 'simtest' logs to see what messages were stuck
* 4. compare traces w/ old logs
 */

// simplest test to vet that simulation works
func TestSetupSim(t *testing.T) {

	params := map[string]string{"--debuglog": ""}
	state0 := SetupSim("L", params, 8, 0, 0, t)

	WaitMinutes(state0, 1)
	assert.True(t, true)
	ShutDownEverything(t)

	for _, ml := range state0.Hold.Messages() {
		for _, m := range ml {
			state0.LogMessage("simTest", "stuck", m)
		}
	}

}
