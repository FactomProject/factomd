package longtests

import (
	"fmt"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFastBootSaveAndRestore(t *testing.T) {
	var saveRate = 4
	var state0 *state.State
	var fastBootFile string

	startSim := func(nodes string) {
		state0 = SetupSim(
			nodes,
			map[string]string{"--debuglog": ".", "--fastsaverate": fmt.Sprintf("%v", saveRate) },
			saveRate*3+1,
			0,
			0,
			t,
		)
	}

	stopSim := func(t *testing.T) {
		WaitForAllNodes(state0)
		ShutDownEverything(t)
		state0 = nil
	}

	t.Run("run sim to create Fastboot", func(t *testing.T) {
		startSim("L")
		WaitForBlock(state0, saveRate*2+2)
		fastBootFile = state.NetworkIDToFilename(state0.Network, state0.FastBootLocation)
		assert.FileExists(t, fastBootFile)
		stopSim(t)
	})
	t.Run("start with Fastboot", func(t *testing.T) {
		// FIXME
		// there is a bug where messages are stuck in holding after booting w/ fastboot
		// as consequence WaitForAllNodes never returns
		startSim("LF")
		panic("NeverGetsHere")
		WaitBlocks(state0, 1)
		stopSim(t)
		assert.Nil(t, os.Remove(fastBootFile))
	})
}
